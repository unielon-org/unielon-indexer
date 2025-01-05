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
	"math/big"
	"strings"
)

func (e *Explorer) drc20Decode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.Drc20Info, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Hash).First(&models.Drc20Info{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("drc20 already exist or err %s", tx.Hash)
	}

	param := &models.Drc20Inscription{}
	err = json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	card, err := utils.ConvetCard(param)
	if err != nil {
		return nil, fmt.Errorf("ConvetCard err: %s", err.Error())
	}

	card.OrderId = uuid.New().String()
	card.FeeTxHash = tx.Vin[0].Txid

	card.TxHash = tx.Hash
	card.BlockHash = tx.BlockHash
	card.BlockNumber = number
	card.Repeat = 1
	card.OrderStatus = 1

	if card.Op == "deploy" {
		card.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
		if tx.Vout[0].Value != 0.001 {
			return nil, fmt.Errorf("The amount of tokens exceeds the 0.0001")
		}
	}

	if card.Op == "mint" {

		card.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
		card.Repeat = int64(tx.Vout[0].Value / 0.001)
		if card.Repeat > 30 {
			card.Repeat = 30
		}

		if tx.Vout[0].Value != 0.001*float64(card.Repeat) {
			return nil, fmt.Errorf("The amount of tokens exceeds the 0.0001")
		}

	}

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerboseBool err: %s", err.Error())
	}

	if card.Op == "transfer" {

		txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
		txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
		if err != nil {
			return nil, fmt.Errorf("GetRawTransactionVerboseBool err: %s", err.Error())
		}

		card.HolderAddress = txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0]
		card.ToAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
		if len(tx.Vout) > 2 {
			for i := 1; i < len(tx.Vout)-1; i++ {
				card.ToAddress += ("," + tx.Vout[i].ScriptPubKey.Addresses[0])
			}
		}
	}

	card.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	for _, v := range strings.Split(card.ToAddress, ",") {
		if card.HolderAddress == v {
			return nil, errors.New("The address is not the same as the previous transaction")
		}
	}

	err = e.dbc.DB.Save(card).Error
	if err != nil {
		return nil, fmt.Errorf("Save err: %s", err.Error())
	}

	return card, nil
}

func (e Explorer) drc20Deploy(drc20 *models.Drc20Info) error {
	tx := e.dbc.DB.Begin()
	drc20c := &models.Drc20Collect{
		Tick:          drc20.Tick,
		Max:           drc20.Max,
		Lim:           drc20.Lim,
		Dec:           drc20.Dec,
		Burn:          drc20.Burn,
		Func:          drc20.Func,
		HolderAddress: drc20.HolderAddress,
		TxHash:        drc20.TxHash,
	}

	err := tx.Create(drc20c).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Save err: %s", err.Error())
	}

	err = tx.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Update err: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Commit err: %s", err.Error())
	}

	return nil
}

func (e *Explorer) drc20Mint(drc20 *models.Drc20Info) error {
	tx := e.dbc.DB.Begin()

	amount := big.NewInt(0).Mul(drc20.Amt.Int(), big.NewInt(drc20.Repeat))
	err := e.dbc.MintDrc20(tx, drc20.Tick, drc20.HolderAddress, amount, drc20.TxHash, drc20.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Update err: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Commit err: %s", err.Error())
	}

	return nil
}

func (e *Explorer) drc20Transfer(drc20 *models.Drc20Info) error {

	tx := e.dbc.DB.Begin()
	err := e.dbc.TransferDrc20(tx, drc20.Tick, drc20.HolderAddress, drc20.ToAddress, drc20.Amt.Int(), drc20.TxHash, drc20.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.Drc20Info{}).Where("tx_hash = ?", drc20.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Update err: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Commit err: %s", err.Error())
	}

	return nil
}
