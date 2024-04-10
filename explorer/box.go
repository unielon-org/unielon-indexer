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

func (e *Explorer) boxDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*utils.BoxInfo, error) {

	// 解析数据
	param := &utils.BoxParams{}
	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	box, err := utils.ConvertBox(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertBox err: %s", err.Error())
	}

	box.OrderId = uuid.New().String()
	box.FeeTxHash = tx.Vin[0].Txid
	box.BoxTxHash = tx.Hash
	box.BoxBlockHash = tx.BlockHash
	box.BoxBlockNumber = number
	box.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, chainNetworkErr
	}

	box.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, chainNetworkErr
	}

	if box.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("the address is not the same as the previous transaction")
	}

	ex1, err := e.dbc.FindBoxInfoByTxHash(box.BoxTxHash)
	if err != nil {
		return nil, fmt.Errorf("FindBoxInfoByTxHash err: %s", err.Error())
	}

	if ex1 != nil {
		if ex1.BoxBlockNumber != 0 {
			return nil, fmt.Errorf("ex already exist or err %s", box.BoxTxHash)
		}
		box.OrderId = ex1.OrderId
		return box, nil
	} else {
		if err := e.dbc.InstallBoxInfo(box); err != nil {
			return nil, fmt.Errorf("InstallBoxInfo err: %s", err.Error())
		}
	}

	return box, nil
}

func (e *Explorer) boxDeploy(box *utils.BoxInfo) error {
	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(box.Tick0+"--BOX"), &chaincfg.MainNetParams)
	err := e.dbc.BoxDeploy(box, reservesAddress.String())
	if err != nil {
		return err
	}
	return nil
}

func (e *Explorer) boxMint(box *utils.BoxInfo) error {

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(box.Tick0+"--BOX"), &chaincfg.MainNetParams)
	err := e.dbc.BoxMint(box, reservesAddress.String())
	if err != nil {
		return err
	}
	return nil
}
