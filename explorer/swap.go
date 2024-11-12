package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
	"math/big"
)

func (e *Explorer) swapRouterDecode(tx *btcjson.TxRawResult, height int64) ([]*models.SwapInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Hash).First(&models.SwapInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("swap already exist or err %s", tx.Hash)
	}

	temp := 0
	swaps := make([]*models.SwapInfo, 0)
	dogeDepositAmt := big.NewInt(0)
	for i, in := range tx.Vin {
		decode, pushedData, err := e.reDecode(in)
		if err == nil && decode.P == "pair-v1" {
			log.Trace("scanning", "verifyReDecode", err, "txhash", tx.Txid)
			temp++
		}

		param := &models.SwapInscription{}
		err = json.Unmarshal(pushedData, param)
		if err != nil {
			return nil, fmt.Errorf("json Unmarshal err: %s", err.Error())
		}

		swap, err := utils.ConvetSwap(param)
		if err != nil {
			return nil, fmt.Errorf("ConvertSwap err: %s", err.Error())
		}

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

		swap.OrderId = uuid.New().String()
		swap.FeeTxHash = in.Txid
		swap.FeeTxIndex = in.Vout
		swap.TxHash = tx.Hash
		swap.TxIndex = i
		swap.BlockHash = tx.BlockHash
		swap.BlockNumber = height
		swap.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
		swap.OrderStatus = 1

		txhash0, _ := chainhash.NewHashFromStr(swap.FeeTxHash)
		txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
		if err != nil {
			return nil, CHAIN_NETWORK_ERR
		}

		swap.FeeAddress = txRawResult0.Vout[swap.FeeTxIndex].ScriptPubKey.Addresses[0]

		txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
		txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
		if err != nil {
			return nil, CHAIN_NETWORK_ERR
		}

		if swap.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
			return nil, fmt.Errorf("the address is not the same as the previous transaction")
		}

		err = e.dbc.DB.Create(swap).Error
		if err != nil {
			return nil, fmt.Errorf("swap create err: %s", err.Error())
		}

		swaps = append(swaps, swap)
	}

	if dogeDepositAmt.Cmp(big.NewInt(0)) > 0 {
		if len(tx.Vout) != 3 {
			return nil, fmt.Errorf("mint op error, vout length is not 3")
		}

		fee := big.NewInt(0)
		fee.Mul(dogeDepositAmt, big.NewInt(3))
		fee.Div(fee, big.NewInt(1000))
		if fee.Cmp(big.NewInt(50000000)) == -1 {
			fee = big.NewInt(50000000)
		}

		if utils.Float64ToBigInt(tx.Vout[1].Value*100000000).Cmp(dogeDepositAmt) < 0 {
			return nil, fmt.Errorf("the amount of tokens is incorrect %f %s", tx.Vout[1].Value, utils.Float64ToBigInt(tx.Vout[1].Value*100000000).String())
		}

		if tx.Vout[1].ScriptPubKey.Addresses[0] != wdogeCoolAddress {
			return nil, fmt.Errorf("the address is incorrect")
		}

		if utils.Float64ToBigInt(tx.Vout[2].Value*100000000).Cmp(fee) < 0 {
			return nil, fmt.Errorf("the amount of tokens is incorrect fee %f", tx.Vout[2].Value)
		}

		if tx.Vout[2].ScriptPubKey.Addresses[0] != wdogeFeeAddress {
			return nil, fmt.Errorf("the address is incorrect")
		}
	}

	if temp != len(tx.Vin) {
		return nil, fmt.Errorf("not a swap transaction")
	}

	return swaps, nil
}

func (e *Explorer) swapCreate(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "create", "tx_hash", swap.TxHash)
	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	err := e.dbc.SwapCreate(db, swap)
	if err != nil {
		return fmt.Errorf("swapCreate Create err: %s", err.Error())
	}

	update := map[string]interface{}{"order_status": 0, "amt0_out": swap.Amt0Out.String(), "amt1_out": swap.Amt1Out.String(), "liquidity": swap.Liquidity.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ? and tx_index = ?", swap.TxHash, swap.TxIndex).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapCreate Update err: %s", err.Error())
	}

	return nil
}

func (e *Explorer) swapAdd(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "add", "tx_hash", swap.TxHash)
	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	err := e.dbc.SwapAdd(db, swap)
	if err != nil {
		return fmt.Errorf("swapAdd Add err: %s", err.Error())
	}

	update := map[string]interface{}{"order_status": 0, "amt0_out": swap.Amt0Out.String(), "amt1_out": swap.Amt1Out.String(), "liquidity": swap.Liquidity.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ? and tx_index = ?", swap.TxHash, swap.TxIndex).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapCreate Update err: %s", err.Error())
	}

	return nil
}

func (e Explorer) swapRemove(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "remove", "tx_hash", swap.TxHash)

	swap.Tick0, swap.Tick1, _, _, _, _ = utils.SortTokens(swap.Tick0, swap.Tick1, nil, nil, nil, nil)

	err := e.dbc.SwapRemove(db, swap)
	if err != nil {
		return fmt.Errorf("swapRemove SwapRemove error: %v", err)
	}

	//amt0_out = ?, amt1_out = ?,
	update := map[string]interface{}{"order_status": 0, "amt0_out": swap.Amt0Out.String(), "amt1_out": swap.Amt1Out.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ? and tx_index = ?", swap.TxHash, swap.TxIndex).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapRemove Update err: %s", err.Error())
	}

	return nil
}

// swapNow
func (e Explorer) swapExec(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "exec", "tx_hash", swap.TxHash)

	err := e.dbc.SwapExec(db, swap)
	if err != nil {
		return fmt.Errorf("swapExec error: %v", err)
	}

	update := map[string]interface{}{"order_status": 0, "amt1_out": swap.Amt1Out.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ? and tx_index = ?", swap.TxHash, swap.TxIndex).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapExec Update err: %s", err.Error())
	}

	return nil
}
