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

func (e Explorer) stakeV2Decode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*models.StakeV2Info, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Txid).First(&models.StakeV2Info{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("stake already exist or err %s", tx.Txid)
	}

	param := &models.StakeV2Inscription{}
	err = json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	stake, err := utils.ConvertStakeV2(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertWDoge err: %s", err.Error())
	}

	if len(tx.Vout) < 1 {
		return nil, fmt.Errorf("op error, vout length is not 0")
	}

	stake.OrderId = uuid.New().String()
	stake.FeeTxHash = tx.Vin[0].Txid
	stake.TxHash = tx.Txid
	stake.BlockHash = tx.BlockHash
	stake.BlockNumber = number
	stake.OrderStatus = 1

	if stake.Op == "create" {
		stake.StakeId = tx.Txid
	}

	stake.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

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

	if stake.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("The address is not the same as the previous transaction")
	}

	err = e.dbc.DB.Save(stake).Error
	if err != nil {
		return nil, fmt.Errorf("SaveStakeV2 err: %s", err.Error())
	}

	return stake, nil
}

func (e Explorer) stakeV2Create(stake *models.StakeV2Info) error {
	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(stake.StakeId+"--STAKE-V2"), &chaincfg.MainNetParams)

	tx := e.dbc.DB.Begin()
	err := e.dbc.StakeV2Create(tx, stake, reservesAddress.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.StakeV2Info{}).Where("tx_hash = ?", stake.TxHash).Update("order_status", 0).Error
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

func (e Explorer) stakeV2Cancel(stake *models.StakeV2Info) error {
	return nil
}

func (e Explorer) stakeV2Stake(stake *models.StakeV2Info) error {

	return nil
}

func (e Explorer) stakeV2UnStake(stake *models.StakeV2Info) error {
	return nil
}

func (e Explorer) stakeV2GetReward(stake *models.StakeV2Info) error {
	return nil
}
