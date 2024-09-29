package explorer

import (
	"bytes"
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

func (e *Explorer) nftDecode(tx *btcjson.TxRawResult, number int64) (*models.NftInfo, error) {

	err := e.dbc.DB.Where("tx_hash = ?", tx.Hash).First(&models.NftInfo{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("nft already exist or err %s", tx.Hash)
	}

	param, err := e.reDecodeNft(tx)
	if err != nil {
		return nil, fmt.Errorf("reDecodeNft err: %s", err.Error())
	}

	nft, err := utils.ConvertNft(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertNft err: %s", err.Error())
	}

	nft.OrderId = uuid.New().String()
	nft.FeeTxHash = tx.Vin[0].Txid

	nft.TxHash = tx.Hash
	nft.BlockHash = tx.BlockHash
	nft.BlockNumber = number

	if nft.Op == "deploy" {

		if len(tx.Vout) != 2 {
			return nil, errors.New("deploy op error, vout length is not 2")
		}

		nft.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

		if tx.Vout[0].Value != 0.001 {
			return nil, fmt.Errorf("The amount of tokens exceeds the 0.0001")
		}

		if tx.Vout[1].Value < 1000 {
			return nil, fmt.Errorf("The balance is insufficient")
		}

		if tx.Vout[1].ScriptPubKey.Addresses[0] != nftFeeAddress {
			return nil, fmt.Errorf("The address is incorrect")
		}
	}

	if nft.Op == "mint" {

		if len(tx.Vout) != 2 {
			return nil, errors.New("mint op error, vout length is not 2")
		}

		nft.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

		if tx.Vout[0].Value != 0.001 {
			return nil, fmt.Errorf("The amount of tokens exceeds the 0.0001")
		}

		if tx.Vout[1].Value < 10 {
			return nil, fmt.Errorf("The balance is insufficient")
		}

		if tx.Vout[1].ScriptPubKey.Addresses[0] != nftFeeAddress {
			return nil, fmt.Errorf("The address is incorrect")
		}
	}

	txHash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txHash0)
	if err != nil {
		return nil, fmt.Errorf("GetRawTransactionVerboseBool err: %s", err.Error())
	}

	if nft.Op == "transfer" {

		txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
		txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
		if err != nil {
			return nil, fmt.Errorf("GetRawTransactionVerboseBool err: %s", err.Error())
		}

		nft.HolderAddress = txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0]
		nft.ToAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

		if nft.HolderAddress == nft.ToAddress {
			return nil, errors.New("The address is the same")
		}
	}

	nft.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	reader := bytes.NewReader(nft.ImageData)
	hash, _ := e.ipfs.Add(reader)
	nft.ImagePath = "https://ipfs.unielon.com/ipfs/" + hash

	err = e.dbc.DB.Create(nft).Error
	if err != nil {
		return nil, fmt.Errorf("InstallNftInfo err: %v", err)
	}

	return nft, nil
}

func (e *Explorer) nftDeploy(nft *models.NftInfo) error {
	log.Info("explorer", "p", "nft/ai", "op", "deploy", "tx_hash", nft.TxHash)

	tx := e.dbc.DB.Begin()
	err := e.dbc.NftDeploy(tx, nft)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("nftDeploy commit err: %s order_id: %s", err.Error(), nft.OrderId)
	}

	return nil
}

func (e *Explorer) nftMint(nft *models.NftInfo) error {

	log.Info("explorer", "p", "nft/ai", "op", "mint", "tx_hash", nft.TxHash)
	tx := e.dbc.DB.Begin()

	err := e.dbc.NftMint(tx, nft)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("nftMint commit err: %s order_id: %s", err.Error(), nft.OrderId)
	}

	return nil
}

func (e *Explorer) nftTransfer(nft *models.NftInfo) error {

	log.Info("explorer", "p", "nft/ai", "op", "transfer", "tx_hash", nft.TxHash)

	tx := e.dbc.DB.Begin()
	err := e.dbc.NftTransfer(tx, nft)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.NftInfo{}).Where("tx_hash = ?", nft.TxHash).Update("order_status", 0).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("nftTransfer commit err: %s order_id: %s", err.Error(), nft.OrderId)
	}

	return nil
}
