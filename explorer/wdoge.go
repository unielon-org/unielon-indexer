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
)

func (e Explorer) wdogeDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.WDogeInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Txid).First(&models.WDogeInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("wdoge already exist or err %s", tx.Txid)
	}

	param := &models.WDogeInscription{}
	err = json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	wdoge, err := utils.ConvertWDoge(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertWDoge err: %s", err.Error())
	}

	wdoge.Tick = "WDOGE(WRAPPED-DOGE)"

	if len(tx.Vout) < 1 {
		return nil, fmt.Errorf("op error, vout length is not 0")
	}

	if wdoge.Op == "deposit" {
		if len(tx.Vout) != 3 {
			return nil, fmt.Errorf("mint op error, vout length is not 3")
		}

		fee := big.NewInt(0)
		fee.Mul(wdoge.Amt.Int(), big.NewInt(3))
		fee.Div(fee, big.NewInt(1000))
		if fee.Cmp(big.NewInt(50000000)) == -1 {
			fee = big.NewInt(50000000)
		}

		if utils.Float64ToBigInt(tx.Vout[1].Value*100000000).Cmp(wdoge.Amt.Int()) < 0 {
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

	wdoge.OrderId = uuid.New().String()
	wdoge.FeeTxHash = tx.Vin[0].Txid
	wdoge.TxHash = tx.Hash
	wdoge.BlockHash = tx.BlockHash
	wdoge.BlockNumber = number
	wdoge.OrderStatus = 1
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
		return nil, fmt.Errorf("the address is not the same as the previous transaction")
	}

	err = e.dbc.DB.Create(wdoge).Error
	if err != nil {
		return nil, fmt.Errorf("InstallWDogeInfo err: %s", err.Error())
	}

	return wdoge, nil
}

func (e Explorer) dogeDeposit(wdoge *models.WDogeInfo) error {

	tx := e.dbc.DB.Begin()

	err := e.dbc.DogeDeposit(tx, wdoge)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.WDogeInfo{}).Where("tx_hash = ?", wdoge.TxHash).Update("order_status", 0).Error
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

func (e Explorer) dogeWithdraw(wdoge *models.WDogeInfo) error {

	tx := e.dbc.DB.Begin()
	err := e.dbc.DogeWithdraw(tx, wdoge)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.WDogeInfo{}).Where("tx_hash = ?", wdoge.TxHash).Update("order_status", 0).Error
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
