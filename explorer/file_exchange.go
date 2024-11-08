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
	"time"
)

func (e *Explorer) fileExchangeDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.FileExchangeInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Hash).First(&models.FileExchangeInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("file-exchange already exist or err %s", tx.Hash)
	}

	param := &models.FileExchangeInscription{}
	err = json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	ex, err := utils.ConvertFileExchange(param)
	if err != nil {
		return nil, fmt.Errorf("exchange err: %s", err.Error())
	}

	ex.OrderId = uuid.New().String()
	ex.FeeTxHash = tx.Vin[0].Txid
	ex.TxHash = tx.Hash
	ex.BlockHash = tx.BlockHash
	ex.BlockNumber = number
	ex.OrderStatus = 1
	ex.UpdateDate = models.LocalTime(time.Now().Unix())
	ex.CreateDate = models.LocalTime(time.Now().Unix())

	ex.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
	if ex.Op == "create" {
		ex.ExId = tx.Hash
	}

	if ex.Op == "trade" {

		exc := &models.FileExchangeCollect{}
		err := e.dbc.DB.Where("ex_id = ? ", ex.ExId).First(exc).Error
		if err != nil {
			return nil, fmt.Errorf("the contract does not exist err %s", err.Error())
		}

		ex.FileId = exc.FileId
	}

	if ex.Op == "cancel" {
		exc := &models.FileExchangeCollect{}
		err := e.dbc.DB.Where("ex_id = ? ", ex.ExId).First(exc).Error
		if err != nil {
			return nil, fmt.Errorf("the contract does not exist err %s", err.Error())
		}

		ex.FileId = exc.FileId
		ex.Tick = exc.Tick
		ex.Amt = exc.Amt

	}

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, CHAIN_NETWORK_ERR
	}

	ex.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, CHAIN_NETWORK_ERR
	}

	if ex.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("the address is not the same as the previous transaction")
	}

	err = e.dbc.DB.Create(ex).Error
	if err != nil {
		return nil, fmt.Errorf("InstallFileExchangeInfo err: %s", err.Error())
	}

	return ex, nil
}

func (e *Explorer) fileExchangeCreate(ex *models.FileExchangeInfo) error {
	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(ex.ExId), &chaincfg.MainNetParams)
	tx := e.dbc.DB.Begin()

	err := e.dbc.FileExchangeCreate(tx, ex, reservesAddress.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("order_status", 0).Error
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

func (e *Explorer) fileExchangeTrade(ex *models.FileExchangeInfo) error {
	tx := e.dbc.DB.Begin()

	err := e.dbc.FileExchangeTrade(tx, ex)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("order_status", 0).Error
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

func (e *Explorer) fileExchangeCancel(ex *models.FileExchangeInfo) error {
	tx := e.dbc.DB.Begin()
	err := e.dbc.FileExchangeCancel(tx, ex)
	if err != nil {
		tx.Rollback()
		return nil
	}

	err = tx.Model(&models.FileExchangeInfo{}).Where("tx_hash = ?", ex.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return nil
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil
	}
	return nil
}
