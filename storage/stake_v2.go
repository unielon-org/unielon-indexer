package storage

import (
	"errors"
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
	"math/big"
)

func (e *DBClient) StakeV2Deploy(tx *gorm.DB, stake *models.StakeV2Info, reservesAddress string) error {

	err := e.TransferDrc20(tx, stake.Tick1, stake.HolderAddress, reservesAddress, stake.Reward.Int(), stake.TxHash, stake.BlockNumber, false)
	if err != nil {
		return err
	}

	stakec := models.StakeV2Collect{
		StakeId:         stake.StakeId,
		Tick0:           stake.Tick0,
		Tick1:           stake.Tick1,
		Reward:          stake.Reward,
		EachReward:      stake.EachReward,
		LastRewardBlock: stake.BlockNumber,
		ReservesAddress: reservesAddress,
	}

	err = tx.Create(&stakec).Error
	if err != nil {
		return err
	}

	revert := &models.StakeV2Revert{
		Op:          "deploy",
		StakeId:     stake.StakeId,
		Tick:        stake.Tick0,
		BlockNumber: stake.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeV2Stake(tx *gorm.DB, stake *models.StakeV2Info) error {

	stakec := &models.StakeV2Collect{}
	err := tx.Where("stake_id = ?", stake.StakeId).First(stakec).Error
	if err != nil {
		return err
	}

	stakec, err = e.StakeV2UpdatePool(tx, stakec, stake.BlockNumber)
	if err != nil {
		return err
	}

	stakea := &models.StakeV2CollectAddress{}
	err = tx.Where("stake_id = ? AND holder_address = ?", stake.StakeId, stake.HolderAddress).First(stakea).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stakea = &models.StakeV2CollectAddress{
				StakeId:       stake.StakeId,
				Tick:          stakec.Tick1,
				Amt:           models.NewNumber(0),
				RewardDebt:    models.NewNumber(0),
				PendingReward: models.NewNumber(0),
				HolderAddress: stake.HolderAddress,
			}
			err = tx.Create(stakea).Error
			if err != nil {
				return err
			}

			// revert
			revert := &models.StakeV2Revert{
				Op:            "stake-create",
				StakeId:       stake.StakeId,
				HolderAddress: stake.HolderAddress,
				BlockNumber:   stake.BlockNumber,
			}

			err = tx.Create(revert).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = e.TransferDrc20(tx, stakec.Tick1, stake.HolderAddress, stakec.ReservesAddress, stake.Amt.Int(), stake.TxHash, stake.BlockNumber, false)
	if err != nil {
		return err
	}

	revert := &models.StakeV2Revert{
		Op:            "stake",
		StakeId:       stake.StakeId,
		Amt:           stakea.Amt,
		RewardDebt:    stakea.RewardDebt,
		PendingReward: stakea.PendingReward,
		HolderAddress: stake.HolderAddress,
		BlockNumber:   stake.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	if big.NewInt(0).Cmp(stakea.Amt.Int()) < 0 {
		pending := big.NewInt(0).Div(big.NewInt(0).Mul(stakea.Amt.Int(), stakec.AccRewardPerShare.Int()), big.NewInt(1e8))
		pending = big.NewInt(0).Sub(pending, stakea.RewardDebt.Int())
		if big.NewInt(0).Cmp(pending) < 0 {
			stakea.PendingReward = (*models.Number)(big.NewInt(0).Add(stakea.PendingReward.Int(), pending))
		}
	}

	stakea.Amt = (*models.Number)(big.NewInt(0).Add(stakea.Amt.Int(), stake.Amt.Int()))
	stakea.RewardDebt = (*models.Number)(big.NewInt(0).Div(big.NewInt(0).Mul(stakea.Amt.Int(), stakec.AccRewardPerShare.Int()), big.NewInt(1e8)))

	err = tx.Model(stakea).Updates(map[string]interface{}{
		"amt":            stakea.Amt,
		"reward_debt":    stakea.RewardDebt,
		"pending_reward": stakea.PendingReward,
	}).Error
	if err != nil {
		return err
	}

	stakec.TotalStaked = (*models.Number)(big.NewInt(0).Add(stakec.TotalStaked.Int(), stake.Amt.Int()))
	err = tx.Model(stakec).Update("total_staked", stakec.TotalStaked).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeV2UnStake(tx *gorm.DB, stake *models.StakeV2Info) error {

	stakec := &models.StakeV2Collect{}
	err := tx.Where("stake_id = ?", stake.StakeId).First(stakec).Error
	if err != nil {
		return err
	}

	stakec, err = e.StakeV2UpdatePool(tx, stakec, stake.BlockNumber)
	if err != nil {
		return err
	}

	stakea := &models.StakeV2CollectAddress{}
	err = tx.Where("stake_id = ? AND holder_address = ?", stake.StakeId, stake.HolderAddress).First(stakea).Error
	if err != nil {
		return err
	}

	revert := &models.StakeV2Revert{
		Op:            "unstake",
		StakeId:       stake.StakeId,
		Tick:          stake.Tick1,
		Amt:           stakea.Amt,
		RewardDebt:    stakea.RewardDebt,
		PendingReward: stakea.PendingReward,
		HolderAddress: stake.HolderAddress,
		BlockNumber:   stake.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	pending := big.NewInt(0).Div(big.NewInt(0).Mul(stakea.Amt.Int(), stakec.AccRewardPerShare.Int()), big.NewInt(1e8))
	pending = big.NewInt(0).Sub(pending, stakea.RewardDebt.Int())
	if big.NewInt(0).Cmp(pending) < 0 {
		stakea.PendingReward = (*models.Number)(big.NewInt(0).Add(stakea.PendingReward.Int(), pending))
	}

	stakea.RewardDebt = (*models.Number)(big.NewInt(0).Div(big.NewInt(0).Mul(stakea.Amt.Int(), stakec.AccRewardPerShare.Int()), big.NewInt(1e8)))
	stakea.Amt = (*models.Number)(big.NewInt(0).Sub(stakea.Amt.Int(), stake.Amt.Int()))
	err = tx.Model(stakea).Updates(map[string]interface{}{
		"amt":            stakea.Amt,
		"reward_debt":    stakea.RewardDebt,
		"pending_reward": stakea.PendingReward,
	}).Error
	if err != nil {
		return err
	}

	stakec.TotalStaked = (*models.Number)(big.NewInt(0).Sub(stakec.TotalStaked.Int(), stake.Amt.Int()))
	err = tx.Model(stakec).Update("total_staked", stakec.TotalStaked).Error
	if err != nil {
		return err
	}

	err = e.TransferDrc20(tx, stakea.Tick, stakec.ReservesAddress, stake.HolderAddress, stakea.Amt.Int(), stake.TxHash, stake.BlockNumber, false)
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeV2GetReward(tx *gorm.DB, stake *models.StakeV2Info) error {

	stakec := &models.StakeV2Collect{}
	err := tx.Where("stake_id = ?", stake.StakeId).First(stakec).Error
	if err != nil {
		return err
	}

	stakec, err = e.StakeV2UpdatePool(tx, stakec, stake.BlockNumber)
	if err != nil {
		return err
	}

	stakea := &models.StakeV2CollectAddress{}
	err = tx.Where("stake_id = ? AND holder_address = ?", stake.StakeId, stake.HolderAddress).First(stakea).Error
	if err != nil {
		return err
	}

	revert := &models.StakeV2Revert{
		Op:            "getreward",
		StakeId:       stake.StakeId,
		RewardDebt:    stakea.RewardDebt,
		PendingReward: stakea.PendingReward,
		HolderAddress: stake.HolderAddress,
		BlockNumber:   stake.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	pending := big.NewInt(0).Div(big.NewInt(0).Mul(stakea.Amt.Int(), stakec.AccRewardPerShare.Int()), big.NewInt(1e8))
	pending = big.NewInt(0).Sub(pending, stakea.RewardDebt.Int())
	if big.NewInt(0).Cmp(pending) < 0 {
		stakea.PendingReward = (*models.Number)(big.NewInt(0).Add(stakea.PendingReward.Int(), pending))
	}

	rewardsToPay := stakea.PendingReward.Int()
	if big.NewInt(0).Cmp(rewardsToPay) >= 0 {
		return errors.New("no rewards to claim")
	}

	stakea.PendingReward = (*models.Number)(big.NewInt(0))
	stakea.RewardDebt = (*models.Number)(big.NewInt(0).Div(big.NewInt(0).Mul(stakea.Amt.Int(), stakec.AccRewardPerShare.Int()), big.NewInt(1e8)))
	err = tx.Model(stakea).Updates(map[string]interface{}{
		"pending_reward": stakea.PendingReward,
		"reward_debt":    stakea.RewardDebt,
	}).Error
	if err != nil {
		return err
	}

	err = e.TransferDrc20(tx, stakec.Tick0, stakec.ReservesAddress, stake.HolderAddress, rewardsToPay, stake.TxHash, stake.BlockNumber, false)
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeV2UpdatePool(tx *gorm.DB, stakec *models.StakeV2Collect, height int64) (*models.StakeV2Collect, error) {

	if stakec.LastRewardBlock >= height {
		return stakec, nil
	}

	revert := &models.StakeV2Revert{
		Op:                "stake-pool",
		StakeId:           stakec.StakeId,
		Amt:               stakec.TotalStaked,
		AccRewardPerShare: stakec.AccRewardPerShare,
		LastRewardBlock:   stakec.LastRewardBlock,
		BlockNumber:       height,
	}

	err := tx.Create(revert).Error
	if err != nil {
		return stakec, err
	}

	if stakec.TotalStaked.Int().Cmp(big.NewInt(0)) == 0 {
		stakec.LastRewardBlock = height
		err = tx.Model(stakec).Updates(map[string]interface{}{
			"last_reward_block": stakec.LastRewardBlock,
		}).Error
		if err != nil {
			return stakec, err
		}
		return stakec, nil
	}

	blocksPassed := height - stakec.LastRewardBlock
	reward := big.NewInt(0).Mul(big.NewInt(blocksPassed), stakec.EachReward.Int())
	accRewardPerShare := big.NewInt(0).Div(big.NewInt(0).Mul(reward, big.NewInt(1e8)), stakec.TotalStaked.Int())

	stakec.AccRewardPerShare = (*models.Number)(big.NewInt(0).Add(stakec.AccRewardPerShare.Int(), accRewardPerShare))
	stakec.LastRewardBlock = height

	err = tx.Model(stakec).Updates(map[string]interface{}{
		"acc_reward_per_share": stakec.AccRewardPerShare,
		"last_reward_block":    stakec.LastRewardBlock,
	}).Error
	if err != nil {
		return stakec, err
	}

	return stakec, nil

}
