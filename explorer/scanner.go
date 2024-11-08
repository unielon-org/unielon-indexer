package explorer

import (
	"context"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/google/uuid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/config"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/verifys"
	"math/big"
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
	CHAIN_NETWORK_ERR = errors.New("chain network error")
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
			return fmt.Errorf("scan ScheduledTasks err: %s", err.Error())
		}

		for _, tx := range block.Tx {

			txhash, _ := chainhash.NewHashFromStr(tx)
			txv, err := e.node.GetRawTransactionVerboseBool(txhash)
			if err != nil {
				return fmt.Errorf("scan GetRawtxvBool err: %s", err.Error())
			}

			decode, pushedData, err := e.reDecode(txv.Vin[0])
			if err != nil {
				log.Trace("scanning", "verifyReDecode", err, "txhash", txv.Txid)
				continue
			}

			switch decode.P {
			case "drc-20":
				drc20, err := e.drc20Decode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "drc20Decode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeDrc20(drc20)
				if err != nil {
					e.dbc.DB.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("err_info", err.Error())
					continue
				}

			case "pair-v1":

				swaps, err := e.swapRouterDecode(txv, e.currentHeight)
				if err != nil {
					log.Error("scanning", "swapRouterDecode", err, "txhash", tx)
					continue
				}

				err = e.executePairV1(swaps)
				if err != nil {
					e.dbc.DB.Model(&models.SwapInfo{}).Where("tx_hash = ?", tx).Update("err_info", err.Error())
					continue
				}

			case "wdoge":
				wdoge, err := e.wdogeDecode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "wdogeDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeWdoge(wdoge)
				if err != nil {
					e.dbc.DB.Model(&models.WDogeInfo{}).Where("tx_hash = ?", wdoge.TxHash).Update("err_info", err.Error())
					continue
				}

			case "nft/ai":
				nft, err := e.nftDecode(txv, e.currentHeight)
				if err != nil {
					log.Error("scanning", "nftDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeNft(nft)
				if err != nil {
					e.dbc.DB.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("err_info", err.Error())
					continue
				}

			case "file":
				file, err := e.fileDecode(txv, e.currentHeight)
				if err != nil {
					log.Error("scanning", "nftDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeFile(file)
				if err != nil {
					e.dbc.DB.Model(&models.FileInfo{}).Where("tx_hash = ?", file.TxHash).Update("err_info", err.Error())
					continue
				}

			case "stake-v1":
				stake, err := e.stakeDecode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "stakeDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeStakeV1(stake)
				if err != nil {
					e.dbc.DB.Model(&models.StakeInfo{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
					continue
				}

			case "stake-v2":
				stake, err := e.stakeV2Decode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "stakeDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeStakeV2(stake)
				if err != nil {
					e.dbc.DB.Model(&models.StakeV2Info{}).Where("tx_hash = ?", stake.TxHash).Update("err_info", err.Error())
					continue
				}

			case "order-v1":
				ex, err := e.exchangeDecode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "exchangeDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeOrderV1(ex)
				if err != nil {
					e.dbc.DB.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
					continue
				}

			case "order-v2":
				ex, err := e.fileExchangeDecode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "fileExchangeDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeOrderV2(ex)
				if err != nil {
					e.dbc.DB.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("err_info", err.Error())
					continue
				}

			case "box-v1":
				box, err := e.boxDecode(txv, pushedData, e.currentHeight)
				if err != nil {
					log.Error("scanning", "boxDecode", err, "txhash", txv.Txid)
					continue
				}

				err = e.executeBoxV1(box)
				if err != nil {
					e.dbc.DB.Model(&models.BoxInfo{}).Where("tx_hash = ?", box.TxHash).Update("err_info", err.Error())
					continue
				}

			default:
				log.Error("scanning", "op", "not found", "txhash", txv.Txid)
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

func (e *Explorer) executeDrc20(drc20 *models.Drc20Info) error {

	err := e.verify.VerifyDrc20(drc20)
	if err != nil {
		return fmt.Errorf("VerifyDrc20 err: %s", err.Error())
	}

	if drc20.Op == "deploy" {
		err = e.drc20Deploy(drc20)
		if err != nil {
			return fmt.Errorf("drc20Deploy err: %s", err.Error())
		}
	}

	if drc20.Op == "mint" {
		err = e.drc20Mint(drc20)
		if err != nil {
			return fmt.Errorf("drc20Mint err: %s", err.Error())
		}
	}

	if drc20.Op == "transfer" {
		err = e.drc20Transfer(drc20)
		if err != nil {
			return fmt.Errorf("drc20Transfer err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeWdoge(wdoge *models.WDogeInfo) error {

	err := e.verify.VerifyWDoge(wdoge)
	if err != nil {
		return fmt.Errorf("VerifyWDoge err: %s", err.Error())
	}

	if wdoge.Op == "deposit" {
		if err = e.wdogeDeposit(wdoge); err != nil {
			return fmt.Errorf("wdogeDeposit err: %s", err.Error())
		}
	}

	if wdoge.Op == "withdraw" {
		if err = e.wdogeWithdraw(wdoge); err != nil {
			return fmt.Errorf("wdogeWithdraw err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeNft(nft *models.NftInfo) error {

	err := e.verify.VerifyNFT(nft)
	if err != nil {
		return fmt.Errorf("VerifyNFT err: %s", err.Error())
	}

	if nft.Op == "deploy" {
		err = e.nftDeploy(nft)
		if err != nil {
			return fmt.Errorf("nftDeploy err: %s", err.Error())
		}
	}

	if nft.Op == "mint" {
		err = e.nftMint(nft)
		if err != nil {
			return fmt.Errorf("nftMint err: %s", err.Error())
		}
	}

	if nft.Op == "transfer" {
		err = e.nftTransfer(nft)
		if err != nil {
			return fmt.Errorf("nftTransfer err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeFile(file *models.FileInfo) error {

	err := e.verify.VerifyFile(file)
	if err != nil {
		return fmt.Errorf("VerifyFile err: %s", err.Error())
	}

	if file.Op == "deploy" {
		err = e.fileDeploy(file)
		if err != nil {
			return fmt.Errorf("fileDeploy err: %s", err.Error())
		}
	}

	if file.Op == "transfer" {
		err = e.fileTransfer(file)
		if err != nil {
			return fmt.Errorf("fileTransfer err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeStakeV1(stake *models.StakeInfo) error {

	err := e.verify.VerifyStake(stake)
	if err != nil {
		return fmt.Errorf("VerifyStake err: %s", err.Error())
	}

	if stake.Op == "stake" {
		err = e.stakeStake(stake)
		if err != nil {
			return fmt.Errorf("stakeStake err: %s", err.Error())
		}
	}

	if stake.Op == "unstake" {
		err = e.stakeUnStake(stake)
		if err != nil {
			return fmt.Errorf("stakeUnStake err: %s", err.Error())
		}
	}

	if stake.Op == "getallreward" {
		err = e.stakeGetAllReward(stake)
		if err != nil {
			return fmt.Errorf("stakeGetAllReward err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeStakeV2(stake *models.StakeV2Info) error {

	err := e.verify.VerifyStakeV2(stake)
	if err != nil {
		return fmt.Errorf("VerifyStakeV2 err: %s", err.Error())
	}

	if stake.Op == "deploy" {
		err = e.stakeV2Deploy(stake)
		if err != nil {
			return fmt.Errorf("stakeV2Create err: %s", err.Error())
		}
	}

	if stake.Op == "stake" {
		err = e.stakeV2Stake(stake)
		if err != nil {
			return fmt.Errorf("stakeV2Stake err: %s", err.Error())
		}
	}

	if stake.Op == "unstake" {
		err = e.stakeV2UnStake(stake)
		if err != nil {
			return fmt.Errorf("stakeV2UnStake err: %s", err.Error())
		}
	}

	if stake.Op == "getreward" {
		err = e.stakeV2GetReward(stake)
		if err != nil {
			return fmt.Errorf("stakeV2GetReward err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeOrderV1(ex *models.ExchangeInfo) error {

	err := e.verify.VerifyExchange(ex)
	if err != nil {
		return fmt.Errorf("VerifyExchange err: %s", err.Error())
	}

	if ex.Op == "create" {
		err = e.exchangeCreate(ex)
		if err != nil {
			return fmt.Errorf("exchangeCreate err: %s", err.Error())
		}
	}

	if ex.Op == "trade" {
		err = e.exchangeTrade(ex)
		if err != nil {
			return fmt.Errorf("exchangeTrade err: %s", err.Error())
		}
	}

	if ex.Op == "cancel" {
		err = e.exchangeCancel(ex)
		if err != nil {
			return fmt.Errorf("exchangeCancel err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeOrderV2(ex *models.FileExchangeInfo) error {
	err := e.verify.VerifyFileExchange(ex)
	if err != nil {
		return fmt.Errorf("VerifyFileExchange err: %s", err.Error())
	}

	if ex.Op == "create" {
		err = e.fileExchangeCreate(ex)
		if err != nil {
			return fmt.Errorf("fileExchangeCreate err: %s", err.Error())
		}
	}

	if ex.Op == "trade" {
		err = e.fileExchangeTrade(ex)
		if err != nil {
			return fmt.Errorf("fileExchangeTrade err: %s", err.Error())
		}
	}

	if ex.Op == "cancel" {
		err = e.fileExchangeCancel(ex)
		if err != nil {
			return fmt.Errorf("fileExchangeCancel err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executeBoxV1(box *models.BoxInfo) error {
	err := e.verify.VerifyBox(box)
	if err != nil {
		return fmt.Errorf("VerifyBox err: %s", err.Error())
	}

	if box.Op == "deploy" {
		err = e.boxDeploy(box)
		if err != nil {
			return fmt.Errorf("boxDeploy err: %s", err.Error())
		}
	}

	if box.Op == "mint" {
		err = e.boxMint(box)
		if err != nil {
			return fmt.Errorf("boxMint err: %s", err.Error())
		}
	}

	return nil
}

func (e *Explorer) executePairV1(swaps []*models.SwapInfo) error {

	dbtx := e.dbc.DB.Begin()

	dogeDepositAmt := big.NewInt(0)
	dogeWithdrawAmt := big.NewInt(0)

	for _, swap := range swaps {
		if swap.Doge == 1 {

			if swap.Op == "create" {
				if swap.Tick0 == "WDOGE(WRAPPED-DOGE)" {
					dogeDepositAmt.Add(dogeDepositAmt, swap.Amt0.Int())
				}

				if swap.Tick1 == "WDOGE(WRAPPED-DOGE)" {
					dogeDepositAmt.Add(dogeDepositAmt, swap.Amt1.Int())
				}
			}

			if swap.Op == "add" {
				if swap.Tick0 == "WDOGE(WRAPPED-DOGE)" {
					dogeDepositAmt.Add(dogeDepositAmt, swap.Amt0.Int())
				}

				if swap.Tick1 == "WDOGE(WRAPPED-DOGE)" {
					dogeDepositAmt.Add(dogeDepositAmt, swap.Amt1.Int())
				}
			}

			if swap.Op == "swap" {
				if swap.Tick0 == "WDOGE(WRAPPED-DOGE)" {
					dogeDepositAmt.Add(dogeDepositAmt, swap.Amt0.Int())
				}
			}

		}
	}

	if dogeDepositAmt.Cmp(big.NewInt(0)) > 0 {
		wdoge := &models.WDogeInfo{}
		wdoge.OrderId = uuid.New().String()
		wdoge.Op = "deposit-swap"
		wdoge.Tick = "WDOGE(WRAPPED-DOGE)"
		wdoge.Amt = (*models.Number)(dogeDepositAmt)
		wdoge.HolderAddress = swaps[0].HolderAddress
		wdoge.TxHash = swaps[0].TxHash
		wdoge.BlockHash = swaps[0].BlockHash
		wdoge.BlockNumber = swaps[0].BlockNumber
		err := e.wdogeDepositSwap(dbtx, wdoge)
		if err != nil {
			dbtx.Rollback()
			return fmt.Errorf("wdogeDepositSwap err: %s", err.Error())
		}
	}

	for _, swap := range swaps {

		err := e.verify.VerifySwap(dbtx, swap)
		if err != nil {
			dbtx.Rollback()
			return fmt.Errorf("VerifySwap err: %s", err.Error())
		}

		if swap.Op == "create" {
			err = e.swapCreate(dbtx, swap)
			if err != nil {
				dbtx.Rollback()
				return fmt.Errorf("swapCreate err: %s", err.Error())
			}
		}

		if swap.Op == "add" {
			err = e.swapAdd(dbtx, swap)
			if err != nil {
				dbtx.Rollback()
				return fmt.Errorf("swapAdd err: %s", err.Error())
			}
		}

		if swap.Op == "remove" {
			err = e.swapRemove(dbtx, swap)
			if err != nil {
				dbtx.Rollback()
				return fmt.Errorf("swapRemove err: %s", err.Error())
			}

			if swap.Doge == 1 {
				if swap.Tick0 == "WDOGE(WRAPPED-DOGE)" {
					dogeWithdrawAmt.Add(dogeWithdrawAmt, swap.Amt0Out.Int())
				}
				if swap.Tick1 == "WDOGE(WRAPPED-DOGE)" {
					dogeWithdrawAmt.Add(dogeWithdrawAmt, swap.Amt1Out.Int())
				}
			}
		}

		if swap.Op == "swap" {
			if err = e.swapExec(dbtx, swap); err != nil {
				dbtx.Rollback()
				return fmt.Errorf("swapExec err: %s", err.Error())
			}

			if swap.Doge == 1 {
				if swap.Tick1 == "WDOGE(WRAPPED-DOGE)" {
					dogeWithdrawAmt.Add(dogeWithdrawAmt, swap.Amt1Out.Int())
				}
			}
		}
	}

	if dogeWithdrawAmt.Cmp(big.NewInt(0)) > 0 {
		wdoge := &models.WDogeInfo{}
		wdoge.OrderId = uuid.New().String()
		wdoge.Op = "withdraw-swap"
		wdoge.Tick = "WDOGE(WRAPPED-DOGE)"
		wdoge.Amt = (*models.Number)(dogeWithdrawAmt)
		wdoge.HolderAddress = swaps[0].HolderAddress
		wdoge.TxHash = swaps[0].TxHash
		wdoge.BlockHash = swaps[0].BlockHash
		wdoge.BlockNumber = swaps[0].BlockNumber
		err := e.wdogeWithdrawSwap(dbtx, wdoge)
		if err != nil {
			dbtx.Rollback()
			return fmt.Errorf("wdogeWithdrawSwap err: %s", err.Error())
		}
	}

	if err := dbtx.Commit().Error; err != nil {
		dbtx.Rollback()
		return fmt.Errorf("swapCommit err: %s", err.Error())
	}

	return nil
}
