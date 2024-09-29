package explorer

import (
	"context"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/go-dogecoin/log"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/config"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/verifys"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	startInterval    = 3 * time.Second
	wdogeFeeAddress  = "D86Dc4n49LZDiXvB41ds2XaDAP1BFjP1qy"
	wdogeCoolAddress = "DKMyk8cfSTGfnCVXfmo8gXta9F6gziu7Z5"
	nftFeeAddress    = "DBFQmJ5oGCgtnDVxUU7xEraztpEyqJHdxz"
)

var (
	chainNetworkErr = errors.New("chain network error")
)

type Explorer struct {
	config        *config.Config
	node          *rpcclient.Client
	dbc           *storage.DBClient
	ipfs          *shell.Shell
	verify        *verifys.Verifys
	currentHeight int64

	ctx context.Context
	wg  *sync.WaitGroup
}

func NewExplorer(ctx context.Context, wg *sync.WaitGroup, rpcClient *rpcclient.Client, dbc *storage.DBClient, ipfs *shell.Shell, currentHeight int64) *Explorer {
	exp := &Explorer{
		node:          rpcClient,
		dbc:           dbc,
		ipfs:          ipfs,
		verify:        verifys.NewVerifys(dbc),
		currentHeight: currentHeight,
		ctx:           ctx,
		wg:            wg,
	}
	return exp
}

func (e *Explorer) Start() {

	defer e.wg.Done()
	if e.currentHeight == 0 {
		maxHeight := e.currentHeight
		err := e.dbc.DB.Model(&models.Block{}).Select("max(block_number)").Scan(&maxHeight).Error
		if err != nil {
			e.currentHeight = 0
		} else {
			e.currentHeight = maxHeight + 1
		}
	}

	startTicker := time.NewTicker(startInterval)
out:
	for {
		select {
		case <-startTicker.C:
			if err := e.scan(); err != nil {
				log.Error("explorer", "Start", err.Error())
			}
		case <-e.ctx.Done():
			log.Warn("explorer", "Stop", "Done")
			break out
		}
	}
}

func (e *Explorer) scan() error {

	blockCount, err := e.node.GetBlockCount()
	if err != nil {
		return fmt.Errorf("scan GetBlockCount err: %s", err.Error())
	}

	temp := int64(0)
	if blockCount-e.currentHeight > 100 {
		temp = 100
	} else {
		temp = blockCount - e.currentHeight
	}

	blockCount = e.currentHeight + temp

	for ; e.currentHeight < blockCount; e.currentHeight++ {
		err := e.forkBack()
		if err != nil {
			return fmt.Errorf("scan forkBack err: %s", err.Error())
		}

		blockHash, err := e.node.GetBlockHash(e.currentHeight)
		if err != nil {
			return fmt.Errorf("scan GetBlockHash err: %s", err.Error())
		}

		block, err := e.node.GetBlockVerboseBool(blockHash)
		if err != nil {
			return fmt.Errorf("scan GetBlockVerboseBool err: %s", err.Error())
		}

		log.Info("explorer", "scanning start ", e.currentHeight, "txs", len(block.Tx))

		err = e.dbc.ScheduledTasks(e.currentHeight)
		if err != nil {
			return err
		}

		for _, tx := range block.Tx {

			txhash, _ := chainhash.NewHashFromStr(tx)
			transactionVerbose, err := e.node.GetRawTransactionVerboseBool(txhash)
			if err != nil {
				log.Error("scanning", "GetRawTransactionVerboseBool", err, "txhash", txhash)
				return err
			}

			if len(transactionVerbose.Vin) > 1 {
				temp := 0
				for _, in := range transactionVerbose.Vin {
					decode, _, err := e.reDecode(in)
					if err == nil && decode.P == "pair-v1" {
						log.Trace("scanning", "verifyReDecode", err, "txhash", txhash)
						temp++
					}
				}

				if temp == len(transactionVerbose.Vin) {
					dbtx := e.dbc.DB.Begin()
				swapf:
					for _, in := range transactionVerbose.Vin {
						_, pushedData, _ := e.reDecode(in)

						swap, err := e.swapDecode(dbtx, transactionVerbose.Txid, pushedData, in, transactionVerbose.Vout, blockHash.String(), e.currentHeight)
						if err != nil {
							log.Trace("scanning", "swapDecode", err, "txhash", transactionVerbose.Txid)
							break swapf
						}

						err = e.execSwap(dbtx, transactionVerbose.Txid, swap)
						if err != nil {
							dbtx.Rollback()
							dbtx.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Update("err_info", err.Error())
							break swapf
						}
					}

					err := dbtx.Commit().Error
					if err != nil {
						log.Error("scanning", "execSwap", err, "txhash", transactionVerbose.Txid)
						continue
					}
					continue
				}
			}

			decode, pushedData, err := e.reDecode(transactionVerbose.Vin[0])
			if err != nil {
				log.Trace("scanning", "verifyReDecode", err, "txhash", transactionVerbose.Txid)
				continue
			}

			switch decode.P {
			case "drc-20":
				drc20, err := e.drc20Decode(transactionVerbose, pushedData, e.currentHeight)
				if err != nil {
					log.Trace("scanning", "drc20Decode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyDrc20(drc20)
				if err != nil {
					log.Error("scanning", "VerifyDrc20", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("err_info", err.Error())
					continue
				}

				if drc20.Op == "deploy" {
					err = e.drc20Deploy(drc20)
					if err != nil {
						log.Error("scanning", "drc20Deploy", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if drc20.Op == "mint" {
					err = e.drc20Mint(drc20)
					if err != nil {
						log.Error("scanning", "drc20Mint", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if drc20.Op == "transfer" {
					err = e.drc20Transfer(drc20)
					if err != nil {
						log.Error("scanning", "drc20Transfer", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			case "pair-v1":
				swap, err := e.swapDecode(e.dbc.DB, transactionVerbose.Txid, pushedData, transactionVerbose.Vin[0], transactionVerbose.Vout, blockHash.String(), e.currentHeight)
				if err != nil {
					log.Trace("scanning", "swapDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				dbtx := e.dbc.DB.Begin()
				err = e.execSwap(dbtx, transactionVerbose.Txid, swap)
				if err != nil {
					dbtx.Rollback()
					e.dbc.DB.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Update("err_info", err.Error())
					continue
				}

				err = dbtx.Commit().Error
				if err != nil {
					log.Error("scanning", "execSwap", err, "txhash", transactionVerbose.Txid)
					continue
				}

			case "wdoge":
				wdoge, err := e.wdogeDecode(transactionVerbose, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "wdogeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyWDoge(wdoge)
				if err != nil {
					log.Error("scanning", "VerifyWDoge", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.WDogeInfo{}).Where("tx_hash = ?", wdoge.TxHash).Update("err_info", err.Error())
					continue
				}

				if wdoge.Op == "deposit" {
					if err = e.dogeDeposit(wdoge); err != nil {
						log.Error("scanning", "dogeDeposit", err.Error(), "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.WDogeInfo{}).Where("tx_hash = ?", wdoge.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if wdoge.Op == "withdraw" {
					if err = e.dogeWithdraw(wdoge); err != nil {
						log.Error("scanning", "dogeWithdraw", err.Error(), "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.WDogeInfo{}).Where("tx_hash = ?", wdoge.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			case "nft/ai":
				nft, err := e.nftDecode(transactionVerbose, e.currentHeight)
				if err != nil {
					log.Error("scanning", "nftDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyNFT(nft)
				if err != nil {
					log.Error("scanning", "VerifyNft", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("err_info", err.Error())
					continue
				}

				if nft.Op == "deploy" {
					err = e.nftDeploy(nft)
					if err != nil {
						log.Error("scanning", "nftDeployOrMintOrTransfer", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if nft.Op == "mint" {
					err = e.nftMint(nft)
					if err != nil {
						log.Error("scanning", "nftDeployOrMintOrTransfer", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if nft.Op == "transfer" {
					err = e.nftTransfer(nft)
					if err != nil {
						log.Error("scanning", "nftDeployOrMintOrTransfer", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			case "file":
				file, err := e.fileDecode(transactionVerbose, e.currentHeight)
				if err != nil {
					log.Error("scanning", "nftDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyFile(file)
				if err != nil {
					log.Error("scanning", "VerifyFile", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.FileInfo{}).Where("tx_hash = ?", file.TxHash).Update("err_info", err.Error())
					continue
				}

				if file.Op == "deploy" {
					err = e.fileDeploy(file)
					if err != nil {
						log.Error("scanning", "fileDeploy", err, "hash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.FileInfo{}).Where("tx_hash = ?", file.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if file.Op == "transfer" {
					err = e.fileTransfer(file)
					if err != nil {
						log.Error("scanning", "fileTransfer", err, "hash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.FileInfo{}).Where("tx_hash = ?", file.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			case "stake-v1":
				stake, err := e.stakeDecode(transactionVerbose, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "stakeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyStake(stake)
				if err != nil {
					log.Error("scanning", "VerifyStake", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
					continue
				}

				if stake.Op == "stake" {
					err = e.stakeStake(stake)
					if err != nil {
						log.Error("scanning", "stakeStake", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if stake.Op == "unstake" {
					err = e.stakeUnStake(stake)
					if err != nil {
						log.Error("scanning", "stakeUnStake", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if stake.Op == "getallreward" {
					err = e.stakeGetAllReward(stake)
					if err != nil {
						log.Error("scanning", "stakeGetAllReward", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			//case "stake-v2":
			//	stake, err := e.stakeV2Decode(transactionVerbose, pushedData, e.currentHeight)
			//	if err != nil {
			//		log.Error("scanning", "stakeDecode", err, "txhash", transactionVerbose.Txid)
			//		continue
			//	}
			//
			//	err = e.verify.VerifyStakeV2(stake)
			//	if err != nil {
			//		log.Error("scanning", "VerifyStake", err, "txhash", transactionVerbose.Txid)
			//		e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
			//		continue
			//	}
			//
			//	if stake.Op == "create" {
			//		err = e.stakeV2Create(stake)
			//		if err != nil {
			//			log.Error("scanning", "stakeV2Create", err, "txhash", transactionVerbose.Txid)
			//			e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
			//			continue
			//		}
			//	}
			//
			//	if stake.Op == "cancel" {
			//		err = e.stakeV2Cancel(stake)
			//		if err != nil {
			//			log.Error("scanning", "stakeV2Cancel", err, "txhash", transactionVerbose.Txid)
			//			e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
			//			continue
			//		}
			//	}
			//
			//	if stake.Op == "stake" {
			//		err = e.stakeV2Stake(stake)
			//		if err != nil {
			//			log.Error("scanning", "stakeV2Stake", err, "txhash", transactionVerbose.Txid)
			//			e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
			//			continue
			//		}
			//	}
			//
			//	if stake.Op == "unstake" {
			//		err = e.stakeV2UnStake(stake)
			//		if err != nil {
			//			log.Error("scanning", "stakeUnStake", err, "txhash", transactionVerbose.Txid)
			//			e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
			//			continue
			//		}
			//	}
			//
			//	if stake.Op == "getreward" {
			//		err = e.stakeV2GetReward(stake)
			//		if err != nil {
			//			log.Error("scanning", "stakeGetAllReward", err, "txhash", transactionVerbose.Txid)
			//			e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
			//			continue
			//		}
			//	}

			case "order-v1":
				ex, err := e.exchangeDecode(transactionVerbose, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "exchangeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyExchange(ex)
				if err != nil {
					log.Error("scanning", "VerifyExchange", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
					continue
				}

				if ex.Op == "create" {
					err = e.exchangeCreate(ex)
					if err != nil {
						log.Error("scanning", "exchangeCreate", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if ex.Op == "trade" {
					err = e.exchangeTrade(ex)
					if err != nil {
						log.Error("scanning", "exchangeTrade", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if ex.Op == "cancel" {
					err = e.exchangeCancel(ex)
					if err != nil {
						log.Error("scanning", "exchangeCancel", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			case "order-v2":
				ex, err := e.fileExchangeDecode(transactionVerbose, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "fileExchangeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyFileExchange(ex)
				if err != nil {
					log.Error("scanning", "VerifyFileExchange", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
					continue
				}

				if ex.Op == "create" {
					err = e.fileExchangeCreate(ex)
					if err != nil {
						log.Error("scanning", "fileExchangeCreate", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if ex.Op == "trade" {
					err = e.fileExchangeTrade(ex)
					if err != nil {
						log.Error("scanning", "fileExchangeTrade", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if ex.Op == "cancel" {
					err = e.fileExchangeCancel(ex)
					if err != nil {
						log.Error("scanning", "fileExchangeCancel", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}
			case "box-v1":
				ex, err := e.boxDecode(transactionVerbose, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "boxDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyBox(ex)
				if err != nil {
					log.Error("scanning", "VerifyBox", err, "txhash", transactionVerbose.Txid)
					e.dbc.DB.Model(&models.BoxInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
					continue
				}

				if ex.Op == "deploy" {
					err = e.boxDeploy(ex)
					if err != nil {
						log.Error("scanning", "boxDeploy", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.BoxInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

				if ex.Op == "mint" {
					err = e.boxMint(ex)
					if err != nil {
						log.Error("scanning", "boxTrade", err, "txhash", transactionVerbose.Txid)
						e.dbc.DB.Model(&models.BoxInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
						continue
					}
				}

			default:
				log.Error("scanning", "op", "not found", "txhash", transactionVerbose.Txid)
			}
		}

		block1 := &models.Block{
			BlockHash:   blockHash.String(),
			BlockNumber: e.currentHeight,
		}

		err = e.dbc.DB.Save(block1).Error
		if err != nil {
			return fmt.Errorf("scan SetBlockHash err: %s", err.Error())
		}

		log.Info("explorer", "scanning end ", e.currentHeight)
	}
	return nil
}

func (e *Explorer) execSwap(db *gorm.DB, txHash string, swap *models.SwapInfo) error {

	err := e.verify.VerifySwap(swap)
	if err != nil {
		log.Error("scanning", "VerifySwap", err, "txhash", txHash)
		return err
	}

	if swap.Op == "create" {
		err = e.swapCreate(db, swap)
		if err != nil {
			log.Error("scanning", "swapCreateOrAdd", err, "txhash", txHash)
			return err
		}
	}

	if swap.Op == "add" {
		err = e.swapAdd(db, swap)
		if err != nil {
			log.Error("scanning", "swapAdd", err, "txhash", txHash)
			db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Update("err_info", err.Error())
			return err
		}
	}

	if swap.Op == "remove" {
		err = e.swapRemove(db, swap)
		if err != nil {
			log.Error("scanning", "swapRemove", err, "txhash", txHash)
			db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Update("err_info", err.Error())
			return err
		}
	}

	if swap.Op == "swap" {
		if err = e.swapExec(db, swap); err != nil {
			log.Error("scanning", "swapNow", err, "txhash", txHash)
			db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Update("err_info", err.Error())
			return err
		}
	}

	return nil
}
