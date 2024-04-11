package explorer

import (
	"context"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/unielon-org/unielon-indexer/config"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/verifys"
	"sync"
	"time"
)

const (
	startInterval = 3 * time.Second

	feeAddress       = "D92uJjQ9eHUcv2GjJUgp6m58V8wYvGV2g9"
	wdogeFeeAddress  = "D86Dc4n49LZDiXvB41ds2XaDAP1BFjP1qy"
	wdogeCoolAddress = "DKMyk8cfSTGfnCVXfmo8gXta9F6gziu7Z5"
	nftFeeAddress    = "DBFQmJ5oGCgtnDVxUU7xEraztpEyqJHdxz"
	stakePoolAddress = "DS8eFcobjXp6oL8YoXoVazDQ32bcDdWwui"
)

var (
	chainNetworkErr = errors.New("Chain network error")
)

type Explorer struct {
	config    *config.Config
	node      *rpcclient.Client
	dbc       *storage.DBClient
	verify    *verifys.Verifys
	fromBlock int64

	ctx context.Context
	wg  *sync.WaitGroup
}

func NewExplorer(ctx context.Context, wg *sync.WaitGroup, rpcClient *rpcclient.Client, dbc *storage.DBClient, fromBlock int64) *Explorer {
	exp := &Explorer{
		node:      rpcClient,
		dbc:       dbc,
		verify:    verifys.NewVerifys(dbc),
		fromBlock: fromBlock,
		ctx:       ctx,
		wg:        wg,
	}
	return exp
}

func (e *Explorer) Start() {

	defer e.wg.Done()

	if e.fromBlock == 0 {
		forkBlockHash, err := e.dbc.LastBlock()
		if err != nil {
			e.fromBlock = 0
		} else {
			e.fromBlock = forkBlockHash + 1
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

	if blockCount-e.fromBlock > 10 {
		blockCount = e.fromBlock + 10
	}

	for ; e.fromBlock < blockCount; e.fromBlock++ {
		err := e.forkBack()
		if err != nil {
			return fmt.Errorf("scan forkBack err: %s", err.Error())
		}

		log.Info("explorer", "scanning start ", e.fromBlock)
		blockHash, err := e.node.GetBlockHash(e.fromBlock)
		if err != nil {
			return fmt.Errorf("scan GetBlockHash err: %s", err.Error())
		}

		block, err := e.node.GetBlockVerboseBool(blockHash)
		if err != nil {
			return fmt.Errorf("scan GetBlockVerboseBool err: %s", err.Error())
		}

		err = e.dbc.ScheduledTasks(e.fromBlock)
		if err != nil {
			return err
		}

		// del
		dbtx, err := e.dbc.SqlDB.Begin()
		err = e.dbc.DelStakeRewardRevert2(dbtx, e.fromBlock-1000)
		if err != nil {
			dbtx.Rollback()
			return err
		}

		err = e.dbc.DelStakeRevert2(dbtx, e.fromBlock-1000)
		if err != nil {
			dbtx.Rollback()
			return err
		}

		err = dbtx.Commit()
		if err != nil {
			dbtx.Rollback()
			return err
		}

		for _, tx := range block.Tx {

			txhash, _ := chainhash.NewHashFromStr(tx)
			transactionVerbose, err := e.node.GetRawTransactionVerboseBool(txhash)
			if err != nil {
				log.Error("scanning", "GetRawTransactionVerboseBool", err, "txhash", transactionVerbose.Txid)
				return err
			}

			decode, pushedData, err := e.reDecode(transactionVerbose)
			if err != nil {
				log.Trace("scanning", "verifyReDecode", err, "txhash", transactionVerbose.Txid)
				continue
			}

			if decode.P == "drc-20" {
				card, err := e.drc20Decode(transactionVerbose, pushedData, e.fromBlock)
				if err != nil {
					log.Error("scanning", "drc20Decode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyDrc20(card)
				if err != nil {
					log.Error("scanning", "VerifyDrc20", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateCardinalsInfoNewErrInfo(card.OrderId, err.Error())
					continue
				}

				err = e.deployOrMintOrTransfer(card)
				if err != nil {
					log.Error("scanning", "deployOrMintOrTransfer", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateCardinalsInfoNewErrInfo(card.OrderId, err.Error())
					continue
				}
			} else if decode.P == "pair-v1" {
				swap, err := e.swapDecode(transactionVerbose, pushedData, e.fromBlock)
				if err != nil {
					log.Error("scanning", "swapDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifySwap(swap)
				if err != nil {
					log.Error("scanning", "VerifySwap", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateSwapInfoErr(swap.OrderId, err.Error())
					continue
				}

				if swap.Op == "create" || swap.Op == "add" {
					err = e.swapCreateOrAdd(swap)
					if err != nil {
						log.Error("scanning", "swapCreateOrAdd", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateSwapInfoErr(swap.OrderId, err.Error())
						continue
					}
				}

				if swap.Op == "remove" {
					err = e.swapRemove(swap)
					if err != nil {
						log.Error("scanning", "swapRemove", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateSwapInfoErr(swap.OrderId, err.Error())
						continue
					}
				}

				if swap.Op == "swap" {
					if err = e.swapNow(swap); err != nil {
						log.Error("scanning", "swapNow", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateSwapInfoErr(swap.OrderId, err.Error())
						continue
					}
				}

			} else if decode.P == "wdoge" {
				wdoge, err := e.wdogeDecode(transactionVerbose, pushedData, e.fromBlock)
				if err != nil {
					log.Error("scanning", "wdogeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyWDoge(wdoge)
				if err != nil {
					log.Error("scanning", "VerifyWDoge", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateWDogeInfoErr(wdoge.OrderId, err.Error())
					continue
				}

				if wdoge.Op == "deposit" {
					if err = e.dogeDeposit(wdoge); err != nil {
						log.Error("scanning", "dogeDeposit", err.Error(), "txhash", transactionVerbose.Txid)
						e.dbc.UpdateWDogeInfoErr(wdoge.OrderId, err.Error())
						continue
					}
				}

				if wdoge.Op == "withdraw" {
					if err = e.dogeWithdraw(wdoge); err != nil {
						log.Error("scanning", "dogeWithdraw", err.Error(), "txhash", transactionVerbose.Txid)
						e.dbc.UpdateWDogeInfoErr(wdoge.OrderId, err.Error())
						continue
					}
				}
			} else if decode.P == "nft/ai" {

				nft, err := e.nftDecode(transactionVerbose, e.fromBlock)
				if err != nil {
					log.Error("scanning", "nftDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyNft(nft)
				if err != nil {
					log.Error("scanning", "VerifyNft", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateNftInfoErr(nft.OrderId, err.Error())
					continue
				}

				err = e.nftDeployOrMintOrTransfer(nft, e.fromBlock)
				if err != nil {
					log.Error("scanning", "nftDeployOrMintOrTransfer", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateNftInfoErr(nft.OrderId, err.Error())
					continue
				}

			} else if decode.P == "stake-v1" {
				stake, err := e.stakeDecode(transactionVerbose, pushedData, e.fromBlock)
				if err != nil {
					log.Error("scanning", "stakeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyStake(stake)
				if err != nil {
					log.Error("scanning", "VerifyStake", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateStakeInfoErr(stake.OrderId, err.Error())
					continue
				}

				if stake.Op == "stake" {
					err = e.stakeStake(stake)
					if err != nil {
						log.Error("scanning", "stakeStake", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateStakeInfoErr(stake.OrderId, err.Error())
						continue
					}
				}

				if stake.Op == "unstake" {
					err = e.stakeUnStake(stake)
					if err != nil {
						log.Error("scanning", "stakeUnStake", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateStakeInfoErr(stake.OrderId, err.Error())
						continue
					}
				}

				if stake.Op == "getallreward" {
					err = e.stakeGetAllReward(stake)
					if err != nil {
						log.Error("scanning", "stakeGetAllReward", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateStakeInfoErr(stake.OrderId, err.Error())
						continue
					}
				}
			} else if decode.P == "order-v1" {

				ex, err := e.exchangeDecode(transactionVerbose, pushedData, e.fromBlock)
				if err != nil {
					log.Error("scanning", "exchangeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyExchange(ex)
				if err != nil {
					log.Error("scanning", "VerifyExchange", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateExchangeInfoErr(ex.OrderId, err.Error())
					continue
				}

				if ex.Op == "create" {
					err = e.exchangeCreate(ex)
					if err != nil {
						log.Error("scanning", "exchangeCreate", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateExchangeInfoErr(ex.OrderId, err.Error())
						continue
					}
				}

				if ex.Op == "trade" {
					err = e.exchangeTrade(ex)
					if err != nil {
						log.Error("scanning", "exchangeTrade", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateExchangeInfoErr(ex.OrderId, err.Error())
						continue
					}
				}

				if ex.Op == "cancel" {
					err = e.exchangeCancel(ex)
					if err != nil {
						log.Error("scanning", "exchangeCancel", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateExchangeInfoErr(ex.OrderId, err.Error())
						continue
					}
				}
			} else if decode.P == "box-v1" {
				ex, err := e.boxDecode(transactionVerbose, pushedData, e.fromBlock)
				if err != nil {
					log.Error("scanning", "exchangeDecode", err, "txhash", transactionVerbose.Txid)
					continue
				}

				err = e.verify.VerifyBox(ex)
				if err != nil {
					log.Error("scanning", "VerifyExchange", err, "txhash", transactionVerbose.Txid)
					e.dbc.UpdateBoxInfoErr(ex.OrderId, err.Error())
					continue
				}

				if ex.Op == "deploy" {
					err = e.boxDeploy(ex)
					if err != nil {
						log.Error("scanning", "exchangeCreate", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateBoxInfoErr(ex.OrderId, err.Error())
						continue
					}
				}

				if ex.Op == "mint" {
					err = e.boxMint(ex)
					if err != nil {
						log.Error("scanning", "exchangeTrade", err, "txhash", transactionVerbose.Txid)
						e.dbc.UpdateBoxInfoErr(ex.OrderId, err.Error())
						continue
					}
				}
			} else {
				log.Error("scanning", "op", "not found", "txhash", transactionVerbose.Txid)
			}
		}

		err = e.dbc.UpdateBlock(e.fromBlock, blockHash.String())
		if err != nil {
			return fmt.Errorf("scan SetBlockHash err: %s", err.Error())
		}

		log.Info("explorer", "scanning end ", e.fromBlock)
	}
	return nil
}
