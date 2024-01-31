package explorer

import (
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (e *Explorer) nftDecode(tx *btcjson.TxRawResult, number int64) (*utils.NFTInfo, error) {

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

	nft.NftTxHash = tx.Hash
	nft.NftBlockHash = tx.BlockHash
	nft.NftBlockNumber = number

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
	nft.FeeAddressAll = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	nfts, err := e.dbc.FindNftInfoByTxHash(nft.NftTxHash)
	if err != nil {
		return nil, fmt.Errorf("FindNftInfoByTxHash err: %s", err.Error())
	}

	if nfts != nil {
		if nfts.NftBlockNumber != 0 {
			return nil, fmt.Errorf("nft already exist or err %s", nft.NftTxHash)
		}
		nft.OrderId = nfts.OrderId
		return nft, nil
	} else {
		if err := e.dbc.InstallNftInfo(nft); err != nil {
			return nil, fmt.Errorf("InstallNftInfo err: %v", err)
		}
	}
	return nft, nil
}

func (e *Explorer) nftDeployOrMintOrTransfer(nft *utils.NFTInfo, height int64) error {

	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return fmt.Errorf("fork Begin err: %s order_id: %s", err.Error(), nft.OrderId)
	}

	if nft.Op == "deploy" {
		err := e.dbc.InstallNftCollect(tx, nft.Tick, nft.Total, nft.Model, nft.Prompt, nft.Image, nft.HolderAddress, nft.NftTxHash)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("deploy InstallNftCollect err: %s order_id: %s", err.Error(), nft.OrderId)
		}
	}

	if nft.Op == "mint" {
		if err := e.dbc.MintNft(tx, nft.Tick, nft.HolderAddress, nft.TickId, nft.Prompt, nft.Image, nft.NftTxHash, false, height); err != nil {
			tx.Rollback()
			return fmt.Errorf("mint MintNft err: %s order_id: %s", err.Error(), nft.OrderId)
		}
	}

	if nft.Op == "transfer" {
		err = e.dbc.TransferNft(tx, nft.Tick, nft.HolderAddress, nft.ToAddress, nft.TickId, false, height)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("transfer TransferNft err: %s order_id: %s", err.Error(), nft.OrderId)
		}
	}

	query := "update nft_info set nft_block_hash = ?, nft_block_number = ?, order_status = 0  where nft_tx_hash = ?"
	_, err = tx.Exec(query, nft.NftBlockHash, nft.NftBlockNumber, nft.NftTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("nftDeployOrMintOrTransfer commit err: %s order_id: %s", err.Error(), nft.OrderId)
	}

	return nil
}
