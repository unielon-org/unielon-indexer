package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"strings"
)

func (e *Explorer) drc20Decode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*utils.Cardinals, error) {

	param := &utils.NewParams{}

	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	card, err := utils.ConvetCard(param)
	if err != nil {
		return nil, fmt.Errorf("ConvetCard err: %s", err.Error())
	}

	card.OrderId = uuid.New().String()
	card.Drc20TxHash = tx.Hash
	card.BlockHash = tx.BlockHash
	card.BlockNumber = number
	card.Repeat = 1

	if card.Op == "deploy" {

		if len(tx.Vout) != 2 {
			return nil, errors.New("mint op error, vout length is not 2")
		}

		card.ReceiveAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

		if tx.Vout[1].ScriptPubKey.Addresses[0] != e.feeAddress {
			return nil, fmt.Errorf("the address is incorrect")
		}

		if tx.Vout[0].Value != 0.001 {
			return nil, fmt.Errorf("the amount of tokens exceeds the 0.0001")
		}

		if tx.Vout[1].Value < 100 {
			return nil, fmt.Errorf("the balance is insufficient")
		}
	}

	if card.Op == "mint" {

		if len(tx.Vout) != 2 {
			return nil, errors.New("mint op error, vout length is not 2")
		}

		card.ReceiveAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
		card.Repeat = int64(tx.Vout[0].Value / 0.001)
		if card.Repeat > 30 {
			card.Repeat = 30
		}

		if tx.Vout[0].Value != 0.001*float64(card.Repeat) {
			return nil, fmt.Errorf("the amount of tokens exceeds the 0.001")
		}

		if tx.Vout[1].Value < float64(card.Repeat)*0.5 {
			return nil, fmt.Errorf("The balance is insufficient")
		}

		if tx.Vout[1].ScriptPubKey.Addresses[0] != e.feeAddress {
			return nil, fmt.Errorf("The address is incorrect")
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

		card.ReceiveAddress = txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0]
		card.ToAddress = tx.Vout[0].ScriptPubKey.Addresses[0]
		if len(tx.Vout) > 2 {
			for i := 1; i < len(tx.Vout)-1; i++ {
				card.ToAddress += ("," + tx.Vout[i].ScriptPubKey.Addresses[0])
			}
		}

	}

	card.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	for _, v := range strings.Split(card.ToAddress, ",") {
		if card.ReceiveAddress == v {
			return nil, errors.New("cardinals op error")
		}
	}

	cardinals, err := e.dbc.FindCardinalsInfoNewByDrc20Hash(card.Drc20TxHash)
	if err != nil {
		return nil, fmt.Errorf("FindCardinalsInfoNewByDrc20Hash err: %s", err.Error())
	}

	if cardinals != nil {
		if cardinals.BlockNumber != 0 {
			return nil, fmt.Errorf("cardinals already exist or err %s", card.Drc20TxHash)
		}
		card.OrderId = cardinals.OrderId
		return card, nil
	} else {
		if err := e.dbc.InstallCardinalsInfo(card); err != nil {
			return nil, fmt.Errorf("InstallCardinalsInfoTransferNew err: %v", err)
		}
	}
	return card, nil
}

func (e *Explorer) deployOrMintOrTransfer(card *utils.Cardinals) error {

	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return fmt.Errorf("fork Begin err: %s order_id: %s", err.Error(), card.OrderId)
	}

	if card.Op == "deploy" {
		log.Info("deploy", "tick", card.Tick, "max", card.Max, "lim", card.Lim, "tick", card.Tick, "drc20_tx_hash", card.Drc20TxHash)
		err := e.dbc.InstallDrc20(tx, card.Max, card.Lim, card.Tick, card.ReceiveAddress, card.Drc20TxHash)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("fork deploy InstallDrc20 err: %s order_id: %s", err.Error(), card.OrderId)
		}
	}

	if card.Op == "mint" {
		log.Info("mint", "tick", card.Tick, "amt", card.Amt, "repeat", card.Repeat, "tick", card.Tick, "drc20_tx_hash", card.Drc20TxHash)
		amount := big.NewInt(0).Mul(card.Amt, big.NewInt(card.Repeat))
		if err := e.dbc.Mint(tx, card.Tick, card.ReceiveAddress, amount, false, card.BlockNumber); err != nil {
			tx.Rollback()
			return fmt.Errorf("fork mint Mint err: %s order_id: %s", err.Error(), card.OrderId)
		}

	}

	if card.Op == "transfer" {
		for _, v := range strings.Split(card.ToAddress, ",") {
			err = e.dbc.Transfer(tx, card.Tick, card.ReceiveAddress, v, card.Amt, false, card.BlockNumber)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("explorer Transfer err: %s order_id: %s", err.Error(), card.OrderId)
			}
		}
	}

	err = e.dbc.UpdateCardinalsBlockNumber(tx, card)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("fork UpdateCardinalsBlockNumber err: %s order_id: %s", err.Error(), card.OrderId)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fork Commit err: %s order_id: %s", err.Error(), card.OrderId)
	}

	return nil
}
