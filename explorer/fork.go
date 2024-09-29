package explorer

import (
	"errors"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
	"math/big"
)

func (e *Explorer) forkBack() error {

	height := e.currentHeight

	blockHash, err := e.node.GetBlockHash(e.currentHeight)
	if err != nil {
		return err
	}

	block, err := e.node.GetBlockVerboseBool(blockHash)
	if err != nil {
		return err
	}

	localHash := ""
	err = e.dbc.DB.Model(&models.Block{}).Where("block_number = ?", height-1).Select("block_hash").First(&localHash).Error
	if err != nil {
		block0 := &models.Block{
			BlockNumber: height - 1,
			BlockHash:   block.PreviousHash,
		}
		err = e.dbc.DB.Create(block0).Error
		if err != nil {
			return err
		}
		return errors.New("localHash is nil")
	}

	if localHash != block.PreviousHash {
		log.Warn("forkBack Begin", "height", height)
		for blockHash.String() != localHash {
			height--
			blockHash, err = e.node.GetBlockHash(height)
			if err != nil {
				return fmt.Errorf("GetBlockHash error: %v", err)
			}
			localHash := ""
			err = e.dbc.DB.Model(&models.Block{}).Where("block_number = ?", height-1).Select("block_hash").First(&localHash).Error
			if localHash == "" {
				return errors.New("localHash is nil")
			}
		}

		tx := e.dbc.DB.Begin()
		err := e.fork(tx, height)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit().Error
		if err != nil {
			return err
		}

		e.currentHeight = height
		log.Warn("forkBack End", "height", height)
	}

	return nil
}

func (e *Explorer) fork(tx *gorm.DB, height int64) error {

	err := e.delInfo(tx, height)
	if err != nil {
		return err
	}

	// drc20
	var drc20Reverts []*models.Drc20Revert
	err = tx.Model(&models.Drc20Revert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&drc20Reverts).Error

	if err != nil {
		return fmt.Errorf("FindDrc20Revert error: %v", err)
	}

	for _, revert := range drc20Reverts {
		if revert.ToAddress != "" && revert.FromAddress == "" {
			err = e.dbc.BurnDrc20(tx, revert.Tick, revert.ToAddress, revert.Amt.Int(), height, true)
			if err != nil {
				return fmt.Errorf("swapFork Burn error: %v", err)
			}
		} else if revert.FromAddress != "" && revert.ToAddress == "" {
			err = e.dbc.MintDrc20(tx, revert.Tick, revert.FromAddress, revert.Amt.Int(), height, true)
			if err != nil {
				return fmt.Errorf("swapFork Mint error: %v", err)
			}
		} else {
			err = e.dbc.TransferDrc20(tx, revert.Tick, revert.ToAddress, revert.FromAddress, revert.Amt.Int(), height, true)
			if err != nil {
				return fmt.Errorf("swapFork Transfer error: %v", err)
			}
		}
	}

	err = tx.Where("block_number > ?", height).Delete(&models.Drc20Revert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteDrc20Revert error: %v", err)
	}

	// nft
	var nftReverts []*models.NftRevert
	err = tx.Model(&models.NftRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&nftReverts).Error

	if err != nil {
		return fmt.Errorf("FindNftRevert error: %v", err)

	}

	for _, revert := range nftReverts {
		if revert.ToAddress != "" && revert.FromAddress == "" {
			err = e.dbc.BurnNft(tx, revert.Tick, revert.ToAddress, revert.TickId)
			if err != nil {
				return fmt.Errorf("nftFork Burn error: %v", err)
			}
		} else {
			err = e.dbc.TransferNft(tx, revert.Tick, revert.ToAddress, revert.FromAddress, revert.TickId, height, true)
			if err != nil {
				return fmt.Errorf("nftFork Transfer error: %v", err)
			}
		}
	}

	err = tx.Where("block_number > ?", height).Delete(&models.NftRevert{}).Error
	if err != nil {
		return err
	}

	// Exchange
	var exchangeReverts []*models.ExchangeRevert
	err = tx.Model(&models.ExchangeRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&exchangeReverts).Error

	if err != nil {
		return fmt.Errorf("exchangeFork error: %v", err)
	}

	for _, revert := range exchangeReverts {
		if revert.Op == "create" {
			err = tx.Where("ex_id = ?", revert.ExId).Delete(&models.ExchangeCollect{}).Error
			if err != nil {
				return fmt.Errorf("delete exchange_collect error: %v", err)
			}
		}

		if revert.Op == "trade" {

			ec := &models.ExchangeCollect{}
			err = tx.Where("ex_id = ?", revert.ExId).First(ec).Error
			if err != nil {
				return fmt.Errorf("select exchange_collect error: %v", err)
			}

			amt0 := ec.Amt0Finish.Int()
			amt1 := ec.Amt1Finish.Int()

			amt0_0 := big.NewInt(0).Sub(amt0, revert.Amt0.Int())
			amt1_1 := big.NewInt(0).Sub(amt1, revert.Amt1.Int())

			err = tx.Model(&models.ExchangeCollect{}).
				Where("ex_id = ?", revert.ExId).
				Updates(map[string]interface{}{
					"amt0_finish": amt0_0.String(),
					"amt1_finish": amt1_1.String(),
				}).Error

			if err != nil {
				return fmt.Errorf("update exchange_collect error: %v", err)
			}
		}

		if revert.Op == "cancel" {

			ec := &models.ExchangeCollect{}
			err = tx.Where("ex_id = ?", revert.ExId).First(ec).Error
			if err != nil {
				return fmt.Errorf("select exchange_collect error: %v", err)
			}

			amt0 := ec.Amt0Finish.Int()
			amt0_0 := big.NewInt(0).Add(amt0, revert.Amt0.Int())

			err = tx.Model(&models.ExchangeCollect{}).Where("ex_id = ?", revert.ExId).Update("amt0_finish", amt0_0.String()).Error
			if err != nil {

				return fmt.Errorf("update exchange_collect error: %v", err)
			}
		}
	}

	err = tx.Where("block_number > ?", height).Delete(&models.ExchangeRevert{}).Error
	if err != nil {

		return err
	}

	// stake
	var stakeReverts []*models.StakeRevert
	err = tx.Model(&models.StakeRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&stakeReverts).Error

	if err != nil {

		return fmt.Errorf("FindStakeRevert error: %v", err)
	}

	for _, revert := range stakeReverts {
		if revert.FromAddress == "" && revert.ToAddress != "" {
			err = e.dbc.StakeUnStakeV1(tx, revert.Tick, revert.ToAddress, revert.Amt.Int(), height, true)
			if err != nil {

				return fmt.Errorf("swapFork Burn error: %v", err)
			}
		}

		if revert.ToAddress == "" && revert.FromAddress != "" {
			err = e.dbc.StakeStakeV1(tx, revert.Tick, revert.FromAddress, revert.Amt.Int(), height, true)
			if err != nil {

				return fmt.Errorf("swapFork Mint error: %v", err)
			}
		}
	}

	err = tx.Where("block_number > ?", height).Delete(&models.StakeRevert{}).Error
	if err != nil {

		return fmt.Errorf("DeleteStakeRevert error: %v", err)
	}

	stakeRewardReverts := []*models.StakeRewardRevert{}
	err = tx.Model(&models.StakeRewardRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&stakeRewardReverts).Error

	if err != nil {

		return fmt.Errorf("FindStakeRewardRevert error: %v", err)
	}

	for _, revert := range stakeRewardReverts {

		if revert.ToAddress == "collect" {
			stakeCollect := &models.StakeCollect{}
			err = tx.Model(&models.StakeCollect{}).Where("tick = ?", revert.Tick).First(stakeCollect).Error
			if err != nil {

				return fmt.Errorf("FindStakeCollect error: %v", err)
			}

			reward := big.NewInt(0).Sub(stakeCollect.Reward.Int(), revert.Amt.Int())

			err = tx.Model(&models.StakeCollect{}).Where("tick = ?", revert.Tick).Update("reward", reward.String()).Error
			if err != nil {

				return fmt.Errorf("update stake_collect error: %v", err)
			}

			return nil
		}

		if revert.FromAddress != "" {
			stakeAddressCollect := &models.StakeCollectAddress{}
			err = tx.Where("tick = ? AND holder_address = ?", revert.Tick, revert.FromAddress).
				First(stakeAddressCollect).Error

			if err != nil {

				return fmt.Errorf("FindStakeCollectAddress error: %v", err)
			}

			reward := big.NewInt(0).Sub(stakeAddressCollect.ReceivedReward.Int(), revert.Amt.Int())

			err = tx.Model(&models.StakeCollectAddress{}).
				Where("tick = ? AND holder_address = ?", revert.Tick, revert.FromAddress).
				Update("received_reward", reward.String()).Error
			if err != nil {

				return err
			}

			return nil
		}

		stakeAddressCollect := &models.StakeCollectAddress{}
		err = tx.Where("tick = ? AND holder_address = ?", revert.Tick, revert.ToAddress).
			First(stakeAddressCollect).Error

		if err != nil {

			return fmt.Errorf("FindStakeCollectAddress error: %v", err)
		}

		reward := big.NewInt(0).Sub(stakeAddressCollect.Reward.Int(), revert.Amt.Int())

		err = tx.Model(&models.StakeCollectAddress{}).
			Where("tick = ? AND holder_address = ?", revert.Tick, revert.ToAddress).
			Update("received_reward", reward.String()).Error
		if err != nil {

			return err
		}
	}

	// box
	err = tx.Exec("update box_collect a, drc20_collect_address b set a.liqamt_finish = b.amt_sum where a.tick1 = b.tick and a.reserves_address = b.holder_address").Error
	if err != nil {
		return fmt.Errorf("update box_collect error: %v", err)
	}

	err = tx.Model(&models.BoxCollect{}).
		Where("liqblock > ?", height).
		Update("is_del", 0).Error
	if err != nil {

		return fmt.Errorf("update box_collect error: %v", err)
	}

	// file
	var fileReverts []*models.FileRevert
	err = tx.Model(&models.FileRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&fileReverts).Error

	if err != nil {

		return fmt.Errorf("FileRevert error: %v", err)

	}

	for _, revert := range fileReverts {
		if revert.ToAddress != "" && revert.FromAddress == "" {
			err = e.dbc.BurnFile(tx, revert.ToAddress, revert.FileId)
			if err != nil {

				return fmt.Errorf("fileFork Burn error: %v", err)
			}
		} else {
			err = e.dbc.TransferFile(tx, revert.ToAddress, revert.FromAddress, revert.FileId, height, true)
			if err != nil {

				return fmt.Errorf("fileFork Transfer error: %v", err)
			}
		}
	}

	err = tx.Where("block_number > ?", height).Delete(&models.FileRevert{}).Error
	if err != nil {

		return err
	}

	// FileExchange
	var fileExchangeReverts []*models.FileExchangeRevert
	err = tx.Model(&models.FileExchangeRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&fileExchangeReverts).Error

	if err != nil {
		return fmt.Errorf("exchangeFork error: %v", err)
	}

	for _, revert := range fileExchangeReverts {
		if revert.Op == "create" {
			err = tx.Where("ex_id = ?", revert.ExId).Delete(&models.FileExchangeCollect{}).Error
			if err != nil {
				return fmt.Errorf("delete error: %v", err)
			}
		}

		if revert.Op == "trade" {

			ec := &models.FileExchangeCollect{}
			err = tx.Where("ex_id = ?", revert.ExId).First(ec).Error
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}

			err = tx.Model(&models.ExchangeCollect{}).
				Where("ex_id = ?", revert.ExId).
				Updates(map[string]interface{}{
					"amt_finish": models.NewNumber(0),
				}).Error

			if err != nil {
				return fmt.Errorf("update exchange_collect error: %v", err)
			}
		}

		if revert.Op == "cancel" {

			ec := &models.ExchangeCollect{}
			err = tx.Where("ex_id = ?", revert.ExId).First(ec).Error
			if err != nil {
				return fmt.Errorf("select exchange_collect error: %v", err)
			}

			err = tx.Model(&models.ExchangeCollect{}).Where("ex_id = ?", revert.ExId).Update("amt_finish", models.NewNumber(0)).Error
			if err != nil {
				return fmt.Errorf("update exchange_collect error: %v", err)
			}
		}
	}

	err = tx.Where("block_number > ?", height).Delete(&models.FileExchangeRevert{}).Error
	if err != nil {
		return err
	}

	return nil

}

func (e *Explorer) delInfo(tx *gorm.DB, height int64) error {
	err := tx.Where("block_number > ?", height).Delete(&models.Drc20Info{}).Error
	if err != nil {
		return fmt.Errorf("DeleteDrc20Info error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.SwapInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteDrc20Collect error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.WDogeInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteDrc20Revert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.NftInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteNftInfo error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.BoxInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteBoxInfo error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.ExchangeInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteExchangeInfo error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.StakeInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteStakeInfo error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.StakeRewardInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteStakeRewardInfo error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.BoxAddress{}).Error
	if err != nil {
		return fmt.Errorf("DeleteBoxAddress error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.FileInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteFileInfo error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.FileExchangeInfo{}).Error
	if err != nil {
		return fmt.Errorf("DeleteFileExchangeInfo error: %v", err)
	}

	return nil

}
