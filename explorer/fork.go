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
		return fmt.Errorf("drc20 revert error: %v", err)
	}

	for _, revert := range drc20Reverts {
		if revert.ToAddress != "" && revert.FromAddress == "" {
			err = e.dbc.BurnDrc20(tx, revert.Tick, revert.ToAddress, revert.Amt.Int(), "", 0, true)
			if err != nil {
				return fmt.Errorf("drc20 fork burn error: %v", err)
			}
		} else if revert.FromAddress != "" && revert.ToAddress == "" {
			err = e.dbc.MintDrc20(tx, revert.Tick, revert.FromAddress, revert.Amt.Int(), "", 0, true)
			if err != nil {
				return fmt.Errorf("drc20 fork mint error: %v", err)
			}
		} else {
			err = e.dbc.TransferDrc20(tx, revert.Tick, revert.ToAddress, revert.FromAddress, revert.Amt.Int(), "", 0, true)
			if err != nil {
				return fmt.Errorf("drc20 fork transfer error: %v", err)
			}
		}
	}

	// nft
	//var nftReverts []*models.NftRevert
	//err = tx.Model(&models.NftRevert{}).
	//	Where("block_number > ?", height).
	//	Order("id desc").
	//	Find(&nftReverts).Error
	//
	//if err != nil {
	//	return fmt.Errorf("FindNftRevert error: %v", err)
	//}
	//
	//for _, revert := range nftReverts {
	//	if revert.ToAddress != "" && revert.FromAddress == "" {
	//		err = e.dbc.BurnNft(tx, revert.Tick, revert.ToAddress, revert.TickId)
	//		if err != nil {
	//			return fmt.Errorf("nftFork Burn error: %v", err)
	//		}
	//	} else {
	//		err = e.dbc.TransferNft(tx, revert.Tick, revert.ToAddress, revert.FromAddress, revert.TickId, height, true)
	//		if err != nil {
	//			return fmt.Errorf("nftFork Transfer error: %v", err)
	//		}
	//	}
	//}
	//
	//err = tx.Where("block_number > ?", height).Delete(&models.NftRevert{}).Error
	//if err != nil {
	//	return err
	//}

	// file
	var fileReverts []*models.FileRevert
	err = tx.Model(&models.FileRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&fileReverts).Error

	if err != nil {
		return fmt.Errorf("file revert error: %v", err)
	}

	for _, revert := range fileReverts {
		if revert.ToAddress != "" && revert.FromAddress == "" {
			err := tx.Where("file_id = ? AND holder_address = ?", revert.FileId, revert.ToAddress).
				Delete(&models.FileCollectAddress{}).Error
			if err != nil {
				return fmt.Errorf("fileFork burn error: %v", err)
			}
		} else {
			err = e.dbc.TransferFile(tx, revert.ToAddress, revert.FromAddress, revert.FileId, "", height, true)
			if err != nil {
				return fmt.Errorf("fileFork Transfer error: %v", err)
			}
		}
	}

	// Exchange
	var exchangeReverts []*models.ExchangeRevert
	err = tx.Model(&models.ExchangeRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&exchangeReverts).Error

	if err != nil {
		return fmt.Errorf("exchange fork error: %v", err)
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
			err = e.dbc.StakeUnStakeV1(tx, revert.Tick, revert.ToAddress, revert.Amt.Int(), "", 0, true)
			if err != nil {
				return fmt.Errorf("stakev1Fork UnStakeV1 error: %v", err)
			}
		}

		if revert.FromAddress != "" && revert.ToAddress == "" {
			err = e.dbc.StakeStakeV1(tx, revert.Tick, revert.FromAddress, revert.Amt.Int(), "", 0, true)
			if err != nil {
				return fmt.Errorf("stakev1Fork StakeV1 error: %v", err)
			}
		}
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

	err = tx.Where("block_number > ?", height).Delete(&models.BoxCollectAddress{}).Error
	if err != nil {
		return fmt.Errorf("DeleteBoxCollectAddress error: %v", err)
	}

	var boxReverts []*models.BoxRevert
	err = tx.Model(&models.BoxRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&boxReverts).Error

	if err != nil {
		return fmt.Errorf("box revert error: %v", err)
	}

	for _, revert := range boxReverts {
		if revert.Op == "deploy" {

			err = tx.Where("tick = ?", revert.Tick0).Delete(&models.Drc20Collect{}).Error
			if err != nil {
				return fmt.Errorf("delete drc20_collect error: %v", err)
			}

			err = tx.Where("tick0 = ?", revert.Tick0).Delete(&models.BoxCollect{}).Error
			if err != nil {
				return fmt.Errorf("delete box_collect error: %v", err)
			}
		}

		if revert.Op == "finish" {
			err = tx.Where("tick0 = ? and tick1 = ?", revert.Tick0, revert.Tick1).Delete(&models.SwapLiquidity{}).Error
			if err != nil {
				return fmt.Errorf("delete swap_info error: %v", err)
			}
		}

		if revert.Op == "refund-drc20" {

			drc20c := &models.Drc20Collect{
				Tick:          revert.Tick0,
				Max:           revert.Max,
				Dec:           8,
				HolderAddress: revert.HolderAddress,
				TxHash:        revert.TxHash,
			}

			err := tx.Create(drc20c).Error
			if err != nil {
				return fmt.Errorf("create drc20_collect error: %v", err)
			}
		}
	}

	err = tx.Model(&models.BoxCollect{}).
		Where("liqblock > ?", height).
		Update("is_del", 0).Error
	if err != nil {
		return fmt.Errorf("update box_collect error: %v", err)
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

	// stake_v2
	var stakeV2Reverts []*models.StakeV2Revert
	err = tx.Model(&models.StakeV2Revert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&stakeV2Reverts).Error
	if err != nil {
		return fmt.Errorf("FindStakeV2Revert error: %v", err)
	}

	for _, revert := range stakeV2Reverts {
		if revert.Op == "deploy" {
			err = tx.Where("stake_id = ?", revert.StakeId).Delete(&models.StakeV2Collect{}).Error
			if err != nil {
				return fmt.Errorf("StakeV2Collect error: %v", err)
			}

			err = tx.Where("tick = ?", revert.Tick).Delete(&models.Drc20Collect{}).Error
			if err != nil {
				return fmt.Errorf("Drc20Collect error: %v", err)
			}
		}

		if revert.Op == "stake-pool" {
			err = tx.Model(&models.StakeV2Collect{}).Where("stake_id = ?", revert.StakeId).Updates(map[string]interface{}{
				"total_staked":         revert.Amt,
				"acc_reward_per_share": revert.AccRewardPerShare,
				"last_block":           revert.LastBlock,
			}).Error
			if err != nil {
				return fmt.Errorf("StakeV2Collect error: %v", err)
			}
		}

		if revert.Op == "stake-create" {
			err = tx.Where("stake_id = ? AND holder_address = ? ", revert.StakeId, revert.HolderAddress).Delete(&models.StakeV2CollectAddress{}).Error
			if err != nil {
				return fmt.Errorf("StakeV2CollectAddress error: %v", err)
			}
		}

		if revert.Op == "stake" {
			err = tx.Model(&models.StakeV2CollectAddress{}).Where("stake_id = ? AND holder_address = ?", revert.StakeId, revert.HolderAddress).Updates(map[string]interface{}{
				"amt":            revert.Amt,
				"reward_debt":    revert.RewardDebt,
				"pending_reward": revert.PendingReward,
			}).Error
			if err != nil {
				return fmt.Errorf("StakeV2CollectAddress error: %v", err)
			}
		}

		if revert.Op == "unstake" {
			err = tx.Model(&models.StakeV2CollectAddress{}).Where("stake_id = ? AND holder_address = ?", revert.StakeId, revert.HolderAddress).Updates(map[string]interface{}{
				"amt":            revert.Amt,
				"reward_debt":    revert.RewardDebt,
				"pending_reward": revert.PendingReward,
			}).Error
			if err != nil {
				return fmt.Errorf("StakeV2CollectAddress error: %v", err)
			}
		}

		if revert.Op == "getreward" {
			err = tx.Model(&models.StakeV2CollectAddress{}).Where("stake_id = ? AND holder_address = ?", revert.StakeId, revert.HolderAddress).Updates(map[string]interface{}{
				"reward_debt":    revert.RewardDebt,
				"pending_reward": revert.PendingReward,
			}).Error
			if err != nil {
				return fmt.Errorf("StakeV2CollectAddress error: %v", err)
			}
		}
	}

	// cross
	var crossReverts []*models.CrossRevert
	err = tx.Model(&models.CrossRevert{}).
		Where("block_number > ?", height).
		Order("id desc").
		Find(&crossReverts).Error
	if err != nil {
		return fmt.Errorf("FindCrossRevert error: %v", err)
	}

	for _, revert := range crossReverts {
		if revert.Op == "deploy" {
			err = tx.Where("tick = ?", revert.Tick).Delete(&models.CrossCollect{}).Error
			if err != nil {
				return fmt.Errorf("CrossCollect error: %v", err)
			}

			err = tx.Where("tick = ?", revert.Tick).Delete(&models.Drc20Collect{}).Error
			if err != nil {
				return fmt.Errorf("Drc20Collect error: %v", err)
			}
		}
	}

	err = e.delRevert(tx, height)
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

	err = tx.Where("block_number > ?", height).Delete(&models.BoxCollectAddress{}).Error
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

	err = tx.Where("block_number > ?", height).Delete(&models.StakeV2Info{}).Error
	if err != nil {
		return fmt.Errorf("StakeV2Info error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.CrossInfo{}).Error
	if err != nil {
		return fmt.Errorf("CrossInfo error: %v", err)
	}

	return nil

}

func (e *Explorer) delRevert(tx *gorm.DB, height int64) error {

	err := tx.Where("block_number > ?", height).Delete(&models.Drc20Revert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteDrc20Revert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.FileRevert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteFileRevert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.ExchangeRevert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteExchangeRevert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.StakeRevert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteStakeRevert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.StakeRewardRevert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteStakeRewardRevert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.FileExchangeRevert{}).Error
	if err != nil {
		return fmt.Errorf("DeleteFileExchangeRevert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.StakeV2Revert{}).Error
	if err != nil {
		return fmt.Errorf("StakeV2Revert error: %v", err)
	}

	err = tx.Where("block_number > ?", height).Delete(&models.CrossRevert{}).Error
	if err != nil {
		return fmt.Errorf("CrossRevert error: %v", err)
	}

	return nil
}
