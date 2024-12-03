package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
)

func (e Explorer) crossDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.CrossInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Txid).First(&models.CrossInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("cross already exist or err %s", tx.Txid)
	}

	param := &models.CrossInscription{}
	err = json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	cross, err := utils.ConvertCross(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertCross err: %s", err.Error())
	}

	cross.OrderId = uuid.New().String()
	cross.FeeTxHash = tx.Vin[0].Txid
	cross.TxHash = tx.Hash
	cross.BlockHash = tx.BlockHash
	cross.BlockNumber = number
	cross.OrderStatus = 1
	cross.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, CHAIN_NETWORK_ERR
	}

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, CHAIN_NETWORK_ERR
	}

	if cross.Op == "mint" {

		txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
		txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
		if err != nil {
			return nil, fmt.Errorf("GetRawTransactionVerboseBool err: %s", err.Error())
		}

		cross.HolderAddress = txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0]

	} else {
		if cross.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
			return nil, fmt.Errorf("the address is not the same as the previous transaction")
		}

	}

	err = e.dbc.DB.Create(cross).Error
	if err != nil {
		return nil, fmt.Errorf("err: %s", err.Error())
	}

	return cross, nil
}

func (e Explorer) crossDeploy(cross *models.CrossInfo) error {

	tx := e.dbc.DB.Begin()
	err := e.dbc.CrossDeploy(tx, cross)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新 status
	err = tx.Model(&models.CrossInfo{}).Where("tx_hash = ?", cross.TxHash).Update("order_status", 0).Error
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

func (e Explorer) crossMint(cross *models.CrossInfo) error {

	tx := e.dbc.DB.Begin()
	err := e.dbc.CrossMint(tx, cross)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新 status
	err = tx.Model(&models.CrossInfo{}).Where("tx_hash = ?", cross.TxHash).Update("order_status", 0).Error
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

func (e Explorer) crossBurn(cross *models.CrossInfo) error {

	tx := e.dbc.DB.Begin()
	err := e.dbc.CrossBurn(tx, cross)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新 status
	err = tx.Model(&models.CrossInfo{}).Where("tx_hash = ?", cross.TxHash).Update("order_status", 0).Error
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
