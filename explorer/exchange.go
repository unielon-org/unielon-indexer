package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
)

func (e *Explorer) exchangeDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.ExchangeInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Hash).First(&models.ExchangeInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("exchange already exist or err %s", tx.Hash)
	}

	inscription := &models.ExchangeInscription{}
	err = json.Unmarshal(pushedData, inscription)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	ex, err := utils.ConvertExChange(inscription)
	if err != nil {
		return nil, fmt.Errorf("exchange err: %s", err.Error())
	}

	ex.OrderId = uuid.New().String()
	ex.FeeTxHash = tx.Vin[0].Txid
	ex.TxHash = tx.Hash
	ex.BlockHash = tx.BlockHash
	ex.BlockNumber = number
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

	err = e.dbc.DB.Save(ex).Error
	if err != nil {
		return nil, fmt.Errorf("Save exchange err: %s", err.Error())
	}

	return ex, nil
}

func (e *Explorer) exchangeCreate(ex *models.ExchangeInfo) error {
	log.Info("explorer", "p", "exchange", "op", "create", "tx_hash", ex.TxHash)
	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(ex.ExId), &chaincfg.MainNetParams)

	tx := e.dbc.DB.Begin()
	err := e.dbc.ExchangeCreate(tx, ex, reservesAddress.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("update status err: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("tx.Commit err: %s", err.Error())
	}

	return nil
}

func (e *Explorer) exchangeTrade(ex *models.ExchangeInfo) error {

	log.Info("explorer", "p", "exchange", "op", "trade", "tx_hash", ex.TxHash)
	tx := e.dbc.DB.Begin()
	err := e.dbc.ExchangeTrade(tx, ex)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Updates(map[string]interface{}{"order_status": 0, "tick0": ex.Tick0, "tick1": ex.Tick1, "amt0": ex.Amt1.String()}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("update status err: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("tx.Commit err: %s", err.Error())
	}

	return nil
}

func (e *Explorer) exchangeCancel(ex *models.ExchangeInfo) error {
	log.Info("explorer", "p", "exchange", "op", "cancel", "tx_hash", ex.TxHash)
	tx := e.dbc.DB.Begin()
	err := e.dbc.ExchangeCancel(tx, ex)
	if err != nil {
		tx.Rollback()
		return nil
	}

	err = tx.Model(&models.ExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Updates(map[string]interface{}{"order_status": 0, "tick0": ex.Tick0, "tick1": ex.Tick1}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("update status err: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("tx.Commit err: %s", err.Error())
	}
	return nil
}
