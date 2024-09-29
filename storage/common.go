package storage

import (
	"errors"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
	"math/big"
	"time"
)

func (e *DBClient) ScheduledTasks(height int64) error {

	s := time.Now()

	//if height < 5260645 {
	//	err = e.StakeUpdatePoolScheduled(tx, height)
	//	if err != nil {
	//		tx.Rollback()
	//		return err
	//	}
	//}

	tx := e.DB.Begin()
	err := e.BoxDeployScheduled(tx, height)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	log.Info("explorer", "StakeUpdatePool time", time.Now().Sub(s).String())
	return nil
}

func (e *DBClient) TransferDrc20(tx *gorm.DB, tick, from, to string, amt *big.Int, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Transfer", "start", "tick", tick, "from", from, "to", to, "amt", amt.String(), "fork", fork)

	if amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("transfer amt < 0")
	}

	if from == to {
		return fmt.Errorf("transfer from and to addresses are the same")
	}

	addFrom := &models.Drc20CollectAddress{}
	err := tx.Where("tick = ? and holder_address = ?", tick, from).First(addFrom).Error
	if err != nil {
		return fmt.Errorf("transfer err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if amt.Cmp(addFrom.AmtSum.Int()) > 0 {
		return fmt.Errorf("insufficient balance : %s tick: %s from : %s  balance : %s  transfer : %s", amt.String(), tick, from, amt.String(), addFrom.AmtSum.String())
	}

	addTo := &models.Drc20CollectAddress{}
	err = tx.Where("tick = ? and holder_address = ?", tick, to).First(addTo).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mint err: %s tick: %s to : %s", err.Error(), tick, to)
		}

		addTo.AmtSum = (*models.Number)(big.NewInt(0))
		addTo.Tick = tick
		addTo.HolderAddress = to
		err := tx.Create(addTo).Error
		if err != nil {
			return fmt.Errorf("mint err: %s tick: %s to : %s", err.Error(), tick, to)
		}
	}

	count1 := addFrom.AmtSum.Int()
	count2 := addTo.AmtSum.Int()

	sub := big.NewInt(0).Sub(count1, amt)
	add := big.NewInt(0).Add(count2, amt)

	err = tx.Model(addFrom).Where("tick = ? and holder_address = ?", tick, from).Update("amt_sum", sub.String()).Error
	if err != nil {
		return fmt.Errorf("transfer err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	err = tx.Model(addTo).Where("tick = ? and holder_address = ?", tick, to).Update("amt_sum", add.String()).Error
	if err != nil {
		return fmt.Errorf("transfer err: %s tick: %s to : %s", err.Error(), tick, to)
	}

	if !fork {
		revert := &models.Drc20Revert{
			FromAddress: from,
			ToAddress:   to,
			Tick:        tick,
			Amt:         (*models.Number)(amt),
			BlockNumber: height,
		}
		err = tx.Create(revert).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) MintDrc20(tx *gorm.DB, tick, holderAddress string, amt *big.Int, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Mint", "start", "tick", tick, "holderAddress", holderAddress, "amt", amt.String())

	drc20c := &models.Drc20Collect{}
	err := tx.Where("tick = ?", tick).First(drc20c).Error
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoByTick err: %s tick: %s", err.Error(), tick)
	}

	drc20ca := &models.Drc20CollectAddress{}

	err = tx.Where("tick = ? and holder_address = ?", tick, holderAddress).First(drc20ca).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mint FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
		}

		drc20ca.AmtSum = (*models.Number)(big.NewInt(0))
		drc20ca.Tick = tick
		drc20ca.HolderAddress = holderAddress
		err := tx.Create(drc20ca).Error
		if err != nil {
			return fmt.Errorf("mint CreateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
		}
	}

	count := drc20c.AmtSum.Int()
	count1 := drc20ca.AmtSum.Int()

	sum := big.NewInt(0).Add(count, amt)
	sum1 := big.NewInt(0).Add(count1, amt)

	trans := drc20c.Transactions + 1
	if fork {
		trans = drc20c.Transactions - 1
		if trans < 0 {
			trans = 0
		}
	}

	err = tx.Model(drc20c).Where("tick = ?", tick).Updates(map[string]interface{}{"amt_sum": sum.String(), "transactions": trans}).Error
	if err != nil {
		return fmt.Errorf("mint UpdateDrc20InfoMint err: %s tick: %s", err.Error(), tick)
	}

	err = tx.Model(drc20ca).Where("tick = ? and holder_address = ?", tick, holderAddress).Update("amt_sum", sum1.String()).Error
	if err != nil {
		return fmt.Errorf("mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
	}

	if !fork {
		revert := &models.Drc20Revert{
			ToAddress:   holderAddress,
			Tick:        tick,
			Amt:         (*models.Number)(amt),
			BlockNumber: height,
		}
		err = tx.Create(revert).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) BurnDrc20(tx *gorm.DB, tick, holderAddress string, amt *big.Int, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Mint", "start", "tick", tick, "holderAddress", holderAddress, "amt", amt.String())

	drc20c := &models.Drc20Collect{}
	err := tx.Where("tick = ?", tick).First(drc20c).Error
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoByTick err: %s tick: %s", err.Error(), tick)
	}

	drc20ca := &models.Drc20CollectAddress{}

	err = tx.Where("tick = ? and holder_address = ?", tick, holderAddress).First(drc20ca).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mint FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
		}

		drc20ca.AmtSum = (*models.Number)(big.NewInt(0))
		drc20ca.Tick = tick
		drc20ca.HolderAddress = holderAddress
		err := tx.Create(drc20ca).Error
		if err != nil {
			return fmt.Errorf("mint CreateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
		}
	}

	count := drc20c.AmtSum.Int()
	count1 := drc20ca.AmtSum.Int()

	if count.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count < amount")
	}

	if count1.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count1 < amount")
	}

	sum := big.NewInt(0).Sub(count, amt)
	sum1 := big.NewInt(0).Sub(count1, amt)

	trans := drc20c.Transactions + 1
	if fork {
		trans = drc20c.Transactions - 1
		if trans < 0 {
			trans = 0
		}
	}

	err = tx.Model(drc20c).Where("tick = ?", tick).Updates(map[string]interface{}{"amt_sum": sum.String(), "transactions": trans}).Error
	if err != nil {
		return fmt.Errorf("mint UpdateDrc20InfoMint err: %s tick: %s", err.Error(), tick)
	}

	err = tx.Model(drc20ca).Where("tick = ? and holder_address = ?", tick, holderAddress).Update("amt_sum", sum1.String()).Error
	if err != nil {
		return fmt.Errorf("mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
	}

	if !fork {
		revert := &models.Drc20Revert{
			FromAddress: holderAddress,
			Tick:        tick,
			Amt:         (*models.Number)(amt),
			BlockNumber: height,
		}
		err = tx.Create(revert).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) TransferNft(tx *gorm.DB, tick, from, to string, tickId int64, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "TransferNft", "start", "tick", tick, "from", from, "to", to, "tickId", tickId, "fork", fork)

	err := tx.Model(&models.NftCollect{}).
		Where("tick = ?", tick).
		Updates(map[string]interface{}{
			"transactions": gorm.Expr("transactions + 1"),
			"tick_sum":     gorm.Expr("tick_sum + 1"),
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.NftCollectAddress{}).
		Where("tick = ? AND tick_id = ? AND holder_address = ?", tick, tickId, from).
		Update("holder_address", to).Error
	if err != nil {
		return err
	}

	if !fork {
		nr := &models.NftRevert{
			Tick:        tick,
			TickId:      tickId,
			FromAddress: from,
			ToAddress:   to,
			BlockNumber: height,
		}

		err = tx.Create(nr).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) MintNft(tx *gorm.DB, tick, holderAddress string, seed int64, prompt, image, imagePath, txHash string, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "MintNft", "start", "tick", tick, "holderAddress", holderAddress, "txHash", txHash)

	err := tx.Model(&models.NftCollect{}).
		Where("tick = ?", tick).
		Updates(map[string]interface{}{
			"transactions": gorm.Expr("transactions + 1"),
			"tick_sum":     gorm.Expr("tick_sum + 1"),
		}).Error
	if err != nil {
		return err
	}

	nc := &models.NftCollect{}
	err = tx.Where("tick = ?", tick).First(nc).Error
	if err != nil {
		return err
	}

	nac := &models.NftCollectAddress{
		Tick:          tick,
		TickId:        nc.TickSum,
		Prompt:        prompt,
		Image:         image,
		ImagePath:     imagePath,
		HolderAddress: holderAddress,
		DeployHash:    txHash,
	}

	err = tx.Create(nac).Error
	if err != nil {
		return err
	}

	if !fork {
		nr := &models.NftRevert{
			Tick:        tick,
			TickId:      nc.TickSum,
			FromAddress: "",
			ToAddress:   holderAddress,
			BlockNumber: height,
		}
		err = tx.Create(nr).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *DBClient) BurnNft(tx *gorm.DB, tick, holderAddress string, tickId int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "BurnNft", "start", "tick", tick, "holderAddress", holderAddress, "tickId", tickId)

	err := tx.Model(&models.NftCollect{}).
		Where("tick = ?", tick).
		Updates(map[string]interface{}{
			"transactions": gorm.Expr("transactions + 1"),
			"tick_sum":     gorm.Expr("tick_sum - 1"),
		}).Error
	if err != nil {
		return err
	}

	// Delete from nft_collect_address
	err = tx.Where("tick = ? AND tick_id = ? AND holder_address = ?", tick, tickId, holderAddress).
		Delete(&models.NftCollectAddress{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) TransferFile(tx *gorm.DB, from, to string, fileId string, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "TransferFile", "start", "from", from, "to", to, "fileId", fileId, "fork", fork)

	err := tx.Model(&models.FileCollectAddress{}).Where("file_id = ? AND holder_address = ?", fileId, from).Update("holder_address", to).Error
	if err != nil {
		return err
	}

	if !fork {
		revert := &models.FileRevert{
			FromAddress: from,
			ToAddress:   to,
			FileId:      fileId,
			BlockNumber: height,
		}
		err = tx.Create(revert).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) BurnFile(tx *gorm.DB, holderAddress string, fileId string) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "BurnFile", "start", "holderAddress", holderAddress, "fileId", fileId)

	// Delete from nft_collect_address
	err := tx.Where("file_id = ? AND holder_address = ?", fileId, holderAddress).
		Delete(&models.FileCollectAddress{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeStakeV1(tx *gorm.DB, tick, holderAddress string, amt *big.Int, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "stake", "start", "tick", tick, "holderAddress", holderAddress, "amt", amt.String())

	stakec := &models.StakeCollect{}

	err := tx.Where("tick = ?", tick).First(stakec).Error
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectByTick err: %s tick: %s", err.Error(), tick)
	}

	amt0 := big.NewInt(0).Add(stakec.Amt.Int(), amt)
	err = tx.Model(stakec).Where("tick = ?", tick).Update("amt", amt0.String()).Error
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollect err: %s tick: %s", err.Error(), tick)
	}

	stakeca := &models.StakeCollectAddress{}
	err = tx.Where("tick = ? and holder_address = ?", tick, holderAddress).First(stakeca).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stakeca.Amt = (*models.Number)(amt)
			stakeca.Tick = tick
			stakeca.HolderAddress = holderAddress
			stakeca.Reward = (*models.Number)(big.NewInt(0))
			err := tx.Create(stakeca).Error
			if err != nil {
				return fmt.Errorf("StakeStake CreateStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
			}
		} else {
			return fmt.Errorf("StakeStake FindStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
		}
	} else {
		amt1 := big.NewInt(0).Add(stakeca.Amt.Int(), amt)
		err = tx.Model(stakeca).Where("tick = ? and holder_address = ?", tick, holderAddress).Update("amt", amt1.String()).Error
		if err != nil {
			return fmt.Errorf("StakeStake UpdateStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
		}
	}

	if !fork {
		sr := &models.StakeRevert{
			Tick:        tick,
			ToAddress:   holderAddress,
			Amt:         (*models.Number)(amt),
			BlockNumber: height,
		}

		err = tx.Create(sr).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) StakeUnStakeV1(tx *gorm.DB, tick, holderAddress string, amt *big.Int, height int64, fork bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "unstake", "start", "tick", tick, "holderAddress", holderAddress, "amt", amt.String())

	stakec := &models.StakeCollect{}
	err := tx.Where("tick = ?", tick).First(stakec).Error
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectByTick err: %s tick: %s", err.Error(), tick)
	}

	amt0 := big.NewInt(0).Sub(stakec.Amt.Int(), amt)
	if amt0.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("StakeStake amt0 < 0 err: %s tick: %s", err.Error(), tick)
	}

	err = tx.Model(stakec).Where("tick = ?", tick).Update("amt", amt0.String()).Error
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollect err: %s tick: %s", err.Error(), tick)
	}

	stakeca := &models.StakeCollectAddress{}
	err = tx.Where("tick = ? and holder_address = ?", tick, holderAddress).First(stakeca).Error
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
	}

	amt1 := big.NewInt(0).Sub(stakeca.Amt.Int(), amt)
	if amt1.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("StakeStake amt1 < 0 err: %s tick: %s", err, tick)
	}

	err = tx.Model(stakeca).Where("tick = ? and holder_address = ?", tick, holderAddress).Update("amt", amt1.String()).Error
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
	}

	if !fork {
		sr := &models.StakeRevert{
			Tick:        tick,
			FromAddress: holderAddress,
			Amt:         (*models.Number)(amt),
			BlockNumber: height,
		}
		err = tx.Create(sr).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *DBClient) StakeRewardV1(tx *gorm.DB, tick, holderAddress string, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	stakeca := &models.StakeCollectAddress{}
	err := tx.Where("tick = ? and holder_address = ?", tick, holderAddress).First(stakeca).Error
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
	}

	err = tx.Model(stakeca).Where("tick = ? and holder_address = ?", tick, holderAddress).Update("received_reward", stakeca.Reward).Error
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, holderAddress)
	}

	reward := big.NewInt(0).Sub(stakeca.Reward.Int(), stakeca.ReceivedReward.Int())

	rsr := &models.StakeRewardRevert{
		Tick:        tick,
		FromAddress: holderAddress,
		Amt:         (*models.Number)(reward),
		BlockNumber: height,
	}

	err = tx.Create(rsr).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) StakeGetRewardV1(tx *gorm.DB, holderAddress, tick string) ([]*models.HolderReward, error) {

	poolResults := make([]*models.Drc20CollectAddress, 0)
	err := tx.Where("holder_address = ? and amt_sum != '0'", stakePoolAddress).Find(&poolResults).Error
	if err != nil {
		return nil, err
	}

	stakeAddressCollect := &models.StakeCollectAddress{}
	err = tx.Where("tick = ? and holder_address = ?", tick, holderAddress).First(stakeAddressCollect).Error
	if err != nil {
		return nil, err
	}

	rewards := make([]*models.HolderReward, 0)
	reward := big.NewInt(0).Sub(stakeAddressCollect.Reward.Int(), stakeAddressCollect.ReceivedReward.Int())

	if tick == "UNIX-SWAP-WDOGE(WRAPPED-DOGE)" {
		rewards = append(rewards, &models.HolderReward{
			Tick:   "WDOGE(WRAPPED-DOGE)",
			Reward: reward,
		})
		return rewards, nil
	}

	unixPool := &models.Drc20CollectAddress{}
	err = tx.Where("tick = 'UNIX' and holder_address = ? and amt_sum != '0'", stakePoolAddress).Find(&unixPool).Error
	if err != nil {
		return nil, err
	}

	for _, ar := range poolResults {
		if ar.Tick == "UNIX" || ar.Tick == "WDOGE(WRAPPED-DOGE)" {
			continue
		}

		amt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.AmtSum.Int(), reward), unixPool.AmtSum.Int())
		receivedAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.AmtSum.Int(), stakeAddressCollect.ReceivedReward.Int()), unixPool.AmtSum.Int())
		TotalAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.AmtSum.Int(), stakeAddressCollect.Reward.Int()), unixPool.AmtSum.Int())
		TotalAmt = big.NewInt(0).Add(TotalAmt, amt)

		rewards = append(rewards, &models.HolderReward{
			Tick:            ar.Tick,
			Reward:          amt,
			ReceivedReward:  receivedAmt,
			TotalRewardPool: TotalAmt,
		})
	}

	rewards = append(rewards, &models.HolderReward{
		Tick:            "UNIX",
		Reward:          reward,
		ReceivedReward:  stakeAddressCollect.ReceivedReward.Int(),
		TotalRewardPool: big.NewInt(0).Add(stakeAddressCollect.Reward.Int(), reward),
	})

	return rewards, nil
}

func (e *DBClient) BoxDeployScheduled(tx *gorm.DB, height int64) error {

	bcs := make([]*models.BoxCollect, 0)
	err := tx.Where("liqblock = ? and is_del = 0", height).Find(&bcs).Error
	if err != nil {
		return err
	}

	for _, bc := range bcs {
		if bc.LiqAmtFinish.Int().Cmp(big.NewInt(0)) == 1 {
			err = e.BoxFinish(tx, bc, height)
			if err != nil {
				return err
			}
		} else {
			err = e.BoxRefund(tx, bc, height)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
