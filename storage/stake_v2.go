package storage

import (
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
)

func (e *DBClient) StakeV2Create(tx *gorm.DB, stake *models.StakeV2Info, reservesAddress string) error {

	err := e.TransferDrc20(tx, stake.Tick1, stake.HolderAddress, reservesAddress, stake.Reward.Int(), stake.BlockNumber, false)
	if err != nil {
		return err
	}

	stakec := models.StakeV2Collect{
		StakeId:         stake.StakeId,
		Tick0:           stake.Tick0,
		Tick1:           stake.Tick1,
		Reward:          stake.Reward,
		EachReward:      stake.EachReward,
		ReservesAddress: reservesAddress,
	}

	err = tx.Create(&stakec).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeV2Cancel(tx *gorm.DB, stake *models.StakeV2Info) error {
	return nil
}

func (e *DBClient) StakeV2Stake(tx *gorm.DB, stake *models.StakeV2Info) error {
	return nil
}

func (e *DBClient) StakeCV2UnStake(tx *gorm.DB, stake *models.StakeInfo) error {
	return nil
}

func (e *DBClient) StakeV2GetReward(tx *gorm.DB, stake *models.StakeInfo) error {

	return nil

}
