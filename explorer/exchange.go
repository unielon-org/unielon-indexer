package explorer

import (
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (e *Explorer) exchangeDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*utils.ExchangeInfo, error) {

	// 解析数据
	param := &utils.ExchangeParams{}
	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	ex, err := utils.ConvertExChange(param)
	if err != nil {
		return nil, fmt.Errorf("exchange err: %s", err.Error())
	}

	ex.OrderId = uuid.New().String()
	ex.FeeTxHash = tx.Vin[0].Txid
	ex.ExchangeTxHash = tx.Hash
	ex.ExchangeBlockHash = tx.BlockHash
	ex.ExchangeBlockNumber = number
	ex.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
	if ex.Op == "create" {
		ex.ExId = tx.Hash
	}

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, chainNetworkErr
	}

	ex.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, chainNetworkErr
	}

	if ex.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("The address is not the same as the previous transaction")
	}

	ex1, err := e.dbc.FindExchangeInfoByTxHash(ex.ExchangeTxHash)
	if err != nil {
		return nil, fmt.Errorf("FindExchangeInfoByTxHash err: %s", err.Error())
	}

	if ex1 != nil {
		if ex1.ExchangeBlockNumber != 0 {
			return nil, fmt.Errorf("ex already exist or err %s", ex.ExchangeTxHash)
		}
		ex.OrderId = ex1.OrderId
		return ex, nil
	} else {
		if err := e.dbc.InstallExchangeInfo(ex); err != nil {
			return nil, fmt.Errorf("InstallExchangeInfo err: %s", err.Error())
		}
	}

	return ex, nil
}

func (e *Explorer) exchangeCreate(ex *utils.ExchangeInfo) error {

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(ex.ExId), &chaincfg.MainNetParams)
	err := e.dbc.ExchangeCreate(ex, reservesAddress.String())
	if err != nil {
		return err
	}

	return nil
}

func (e *Explorer) exchangeTrade(ex *utils.ExchangeInfo) error {
	err := e.dbc.ExchangeTrade(ex)
	if err != nil {
		return err
	}
	return nil
}

func (e *Explorer) exchangeCancel(ex *utils.ExchangeInfo) error {

	err := e.dbc.ExchangeCancel(ex)
	if err != nil {
		return nil
	}
	return nil
}
