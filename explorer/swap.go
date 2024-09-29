package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
)

func (e *Explorer) swapDecode(db *gorm.DB, txHash string, pushedData []byte, vin btcjson.Vin, vout []btcjson.Vout, blockHash string, blockHeight int64) (*models.SwapInfo, error) {

	param := &models.SwapInscription{}
	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json Unmarshal err: %s", err.Error())
	}

	swap, err := utils.ConvetSwap(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertSwap err: %s", err.Error())
	}

	test := &models.SwapInfo{}
	err = e.dbc.DB.Where("tx_hash = ?", txHash).First(test).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("swap already exist or err %s", txHash)
	}

	swap.OrderId = uuid.New().String()
	swap.FeeTxHash = vin.Txid
	swap.TxHash = txHash
	swap.BlockHash = blockHash
	swap.BlockNumber = blockHeight
	swap.HolderAddress = vout[0].ScriptPubKey.Addresses[0]
	swap.OrderStatus = 1

	txhash0, _ := chainhash.NewHashFromStr(vin.Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, chainNetworkErr
	}

	swap.FeeAddress = txRawResult0.Vout[vin.Vout].ScriptPubKey.Addresses[0]

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, chainNetworkErr
	}

	if swap.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("the address is not the same as the previous transaction")
	}

	err = db.Create(swap).Error
	if err != nil {
		return nil, fmt.Errorf("swap create err: %s", err.Error())
	}

	return swap, nil
}

func (e *Explorer) swapCreate(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "create", "tx_hash", swap.TxHash)
	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	err := e.dbc.SwapCreate(db, swap)
	if err != nil {
		return fmt.Errorf("swapCreate Create err: %s", err.Error())
	}

	update := map[string]interface{}{"order_status": 0, "amt0_out": swap.Amt0.String(), "amt1_out": swap.Amt1.String(), "liquidity": swap.Liquidity.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapCreate Update err: %s", err.Error())
	}

	return nil
}

func (e *Explorer) swapAdd(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "add", "tx_hash", swap.TxHash)
	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	err := e.dbc.SwapAdd(db, swap)
	if err != nil {
		return fmt.Errorf("swapCreate Create err: %s", err.Error())
	}

	update := map[string]interface{}{"order_status": 0, "amt0_out": swap.Amt0.String(), "amt1_out": swap.Amt1.String(), "liquidity": swap.Liquidity.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapCreate Update err: %s", err.Error())
	}

	return nil
}

func (e Explorer) swapRemove(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "remove", "tx_hash", swap.TxHash)

	swap.Tick0, swap.Tick1, _, _, _, _ = utils.SortTokens(swap.Tick0, swap.Tick1, nil, nil, nil, nil)

	err := e.dbc.SwapRemove(db, swap)
	if err != nil {
		return fmt.Errorf("swapRemove SwapRemove error: %v", err)
	}

	update := map[string]interface{}{"order_status": 0, "amt0_out": swap.Amt0.String(), "amt1_out": swap.Amt1.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapRemove Update err: %s", err.Error())
	}

	return nil
}

// swapNow
func (e Explorer) swapExec(db *gorm.DB, swap *models.SwapInfo) error {

	log.Info("explorer", "p", "swap", "op", "exec", "tx_hash", swap.TxHash)

	err := e.dbc.SwapExec(db, swap)
	if err != nil {
		return fmt.Errorf("swapExec error: %v", err)
	}

	update := map[string]interface{}{"order_status": 0, "amt1_out": swap.Amt1.String()}
	err = db.Model(&models.SwapInfo{}).Where("tx_hash = ?", swap.TxHash).Updates(update).Error
	if err != nil {
		return fmt.Errorf("swapExec Update err: %s", err.Error())
	}

	return nil
}
