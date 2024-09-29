package storage

import (
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
)

func (e *DBClient) StakeStake(tx *gorm.DB, stake *models.StakeInfo, reservesAddress string) error {

	err := e.TransferDrc20(tx, stake.Tick, stake.HolderAddress, reservesAddress, stake.Amt.Int(), stake.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.StakeStakeV1(tx, stake.Tick, stake.HolderAddress, stake.Amt.Int(), stake.BlockNumber, false)
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeUnStake(tx *gorm.DB, stake *models.StakeInfo, reservesAddress string) error {

	err := e.TransferDrc20(tx, stake.Tick, reservesAddress, stake.HolderAddress, stake.Amt.Int(), stake.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = e.StakeUnStakeV1(tx, stake.Tick, stake.HolderAddress, stake.Amt.Int(), stake.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (e *DBClient) StakeGetReward(tx *gorm.DB, stake *models.StakeInfo) error {

	rewards, err := e.StakeGetRewardV1(tx, stake.HolderAddress, stake.Tick)
	if err != nil {
		return err
	}

	for _, reward := range rewards {
		err = e.TransferDrc20(tx, reward.Tick, stakePoolAddress, stake.HolderAddress, reward.Reward, stake.BlockNumber, false)
		if err != nil {
			return err
		}

		sri := &models.StakeRewardInfo{
			OrderId:     stake.OrderId,
			Tick:        reward.Tick,
			Amt:         (*models.Number)(reward.Reward),
			FromAddress: stakePoolAddress,
			ToAddress:   stake.HolderAddress,
			BlockNumber: stake.BlockNumber,
		}

		err = tx.Create(sri).Error
		if err != nil {
			return err
		}
	}

	err = e.StakeRewardV1(tx, stake.Tick, stake.HolderAddress, stake.BlockNumber)
	if err != nil {
		return err
	}

	return nil

}
