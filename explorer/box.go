package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
)

func (e *Explorer) boxDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.BoxInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Hash).First(&models.BoxInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("box already exist or err %s", tx.Hash)
	}

	// 解析数据
	param := &models.BoxInscription{}
	err = json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	box, err := utils.ConvertBox(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertBox err: %s", err.Error())
	}

	box.OrderId = uuid.New().String()
	box.FeeTxHash = tx.Vin[0].Txid
	box.TxHash = tx.Hash
	box.BlockHash = tx.BlockHash
	box.BlockNumber = number
	box.OrderStatus = 1

	if len(tx.Vout) < 1 {
		return nil, fmt.Errorf("vout length is not enough")
	}

	box.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

	txHashIn, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txHashIn)
	if err != nil {
		return nil, CHAIN_NETWORK_ERR
	}

	box.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	txHashIn1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txHashIn1)
	if err != nil {
		return nil, CHAIN_NETWORK_ERR
	}

	if box.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("the address is not the same as the previous transaction")
	}

	err = e.dbc.DB.Save(box).Error
	if err != nil {
		return nil, err
	}

	return box, nil
}

func (e *Explorer) boxDeploy(box *models.BoxInfo) error {

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(box.Tick0+"--BOX"), &chaincfg.MainNetParams)

	tx := e.dbc.DB.Begin()
	err := e.dbc.BoxDeploy(tx, box, reservesAddress.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.BoxInfo{}).Where("tx_hash = ?", box.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (e *Explorer) boxMint(box *models.BoxInfo) error {

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(box.Tick0+"--BOX"), &chaincfg.MainNetParams)
	tx := e.dbc.DB.Begin()
	err := e.dbc.BoxMint(tx, box, reservesAddress.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.BoxInfo{}).Where("tx_hash = ?", box.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
