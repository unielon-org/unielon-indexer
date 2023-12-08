package explorer

import (
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

func (e Explorer) wdogeDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*utils.WDogeInfo, error) {

	param := &utils.WDogeParams{}
	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	wdoge, err := utils.ConvertWDoge(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertWDoge err: %s", err.Error())
	}

	if len(tx.Vout) < 1 {
		return nil, fmt.Errorf("op error, vout length is not 0")
	}

	if wdoge.Op == "deposit" {
		if len(tx.Vout) != 3 {
			return nil, fmt.Errorf("op error, vout length is not 3")
		}

		fee := big.NewInt(0)
		fee.Mul(wdoge.Amt, big.NewInt(3))
		fee.Div(fee, big.NewInt(1000))
		if fee.Cmp(big.NewInt(50000000)) == -1 {
			fee = big.NewInt(50000000)
		}

		if utils.Float64ToBigInt(tx.Vout[1].Value*100000000).Cmp(wdoge.Amt) < 0 {
			return nil, fmt.Errorf("The amount of tokens is incorrect %f %s", tx.Vout[1].Value, utils.Float64ToBigInt(tx.Vout[1].Value*100000000).String())
		}

		if tx.Vout[1].ScriptPubKey.Addresses[0] != wdogeCoolAddress {
			return nil, fmt.Errorf("The address is incorrect")
		}

		if utils.Float64ToBigInt(tx.Vout[2].Value*100000000).Cmp(fee) < 0 {
			return nil, fmt.Errorf("The amount of tokens is incorrect fee %f", tx.Vout[2].Value)
		}

		if tx.Vout[2].ScriptPubKey.Addresses[0] != wdogeFeeAddress {
			return nil, fmt.Errorf("The address is incorrect")
		}
	} else if wdoge.Op == "withdraw" {

	}

	wdoge.OrderId = uuid.New().String()
	wdoge.WDogeTxHash = tx.Hash
	wdoge.WDogeBlockHash = tx.BlockHash
	wdoge.WDogeBlockNumber = number
	wdoge.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, chainNetworkErr
	}

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, chainNetworkErr
	}

	if wdoge.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("The address is not the same as the previous transaction")
	}

	wdoge1, err := e.dbc.FindWDogeInfoByTxHash(wdoge.WDogeTxHash)
	if err != nil {
		return nil, fmt.Errorf("FindWDogeInfoByTxHash err: %s", err.Error())
	}

	if wdoge1 != nil {
		if wdoge1.WDogeBlockNumber != 0 {
			return nil, fmt.Errorf("wdoge already exist or err %s", wdoge1.WDogeTxHash)
		}
		wdoge.OrderId = wdoge1.OrderId
		return wdoge, nil
	} else {
		if err := e.dbc.InstallWDogeInfo(wdoge); err != nil {
			return nil, fmt.Errorf("InstallWDogeInfo err: %s", err.Error())
		}
	}

	return wdoge, nil
}

func (e Explorer) dogeDeposit(wdoge *utils.WDogeInfo) error {
	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = e.dbc.Mint(tx, wdoge.Tick, wdoge.HolderAddress, wdoge.Amt, false, wdoge.WDogeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	exec := "update wdoge_info set wdoge_block_hash = ?, wdoge_block_number = ? where wdoge_tx_hash = ?"
	_, err = tx.Exec(exec, wdoge.WDogeBlockHash, wdoge.WDogeBlockNumber, wdoge.WDogeTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (e Explorer) dogeWithdraw(wdoge *utils.WDogeInfo) error {

	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = e.dbc.Burn(tx, wdoge.Tick, wdoge.HolderAddress, wdoge.Amt, false, wdoge.WDogeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	exec := "update wdoge_info set wdoge_block_hash = ?, wdoge_block_number = ? where wdoge_tx_hash = ?"
	_, err = tx.Exec(exec, wdoge.WDogeBlockHash, wdoge.WDogeBlockNumber, wdoge.WDogeTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
