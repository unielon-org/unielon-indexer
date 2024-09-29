package storage_v3

import (
	"database/sql"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/unielon-org/unielon-indexer/models"
	"math/big"
	"time"
)

func (e *MysqlClient) ScheduledTasks(height int64) error {

	s := time.Now()
	tx, err := e.MysqlDB.Begin()
	if err != nil {
		return err
	}

	//if height < 5260645 {
	//	err = e.StakeUpdatePoolScheduled(tx, height)
	//	if err != nil {
	//		tx.Rollback()
	//		return err
	//	}
	//}

	//err = e.BoxDeployScheduled(tx, height)
	//if err != nil {
	//	tx.Rollback()
	//	return err
	//}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	log.Info("explorer", "StakeUpdatePool time", time.Now().Sub(s).String())
	return nil
}

func (e *MysqlClient) Transfer(tx *sql.Tx, tick, from, to string, amt *big.Int, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Transfer", "start", "tick", tick, "from", from, "to", to, "amt", amt.String(), "fork", fork)

	if amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("Transfer amt < 0")
	}

	if from == to {
		return fmt.Errorf("Transfer from == to")
	}

	count1, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if amt.Cmp(count1) > 0 {
		return fmt.Errorf("Transfer amt > count: %s tick: %s from : %s  amt : %s  count : %s  ", amt.String(), tick, from, amt.String(), count1.String())
	}

	count2, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, to)
	if err != nil {
		if err != ErrNotFound {
			return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s to : %s", err.Error(), tick, to)
		}
		log.Debug("explorer", "Transfer", fmt.Sprintf("tick: %s to : %s", tick, to))
		count2 = big.NewInt(0)
	}

	sub := big.NewInt(0).Sub(count1, amt)
	add := big.NewInt(0).Add(count2, amt)

	err = e.UpdateAddressBalanceTran(tx, tick, sub, from, add, to, fork)
	if err != nil {
		return fmt.Errorf("Transfer UpdateAddressBalanceTran err: %s tick: %s from : %s to : %s", err.Error(), tick, from, to)
	}

	if !fork {
		err = e.InstallDrc20Revert(tx, tick, from, to, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *MysqlClient) Mint(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Mint", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := e.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		if err != ErrNotFound {
			return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, from)
		}
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	sum := big.NewInt(0).Add(count, amt)
	sum1 := big.NewInt(0).Add(count1, amt)

	err = e.UpdateAddressBalanceMint(tx, tick, sum, sum1, from, false)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = e.InstallDrc20Revert(tx, tick, "", from, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *MysqlClient) Burn(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Burn", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := e.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	if count.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count < amount")
	}

	if count1.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count1 < amount")
	}

	sum := big.NewInt(0).Sub(count, amt)
	sum1 := big.NewInt(0).Sub(count1, amt)

	err = e.UpdateAddressBalanceMint(tx, tick, sum, sum1, from, true)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = e.InstallDrc20Revert(tx, tick, from, "", amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *MysqlClient) TransferNft(tx *sql.Tx, tick, from, to string, tickId int64, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "TransferNft", "start", "tick", tick, "from", from, "to", to, "tickId", tickId, "fork", fork)

	update := "UPDATE nft_collect SET transactions = transactions + 1 WHERE tick = ?"
	_, err := tx.Exec(update, tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	update1 := "UPDATE  nft_collect_address SET holder_address = ? WHERE tick = ? AND tick_id = ? AND holder_address = ?"
	_, err = tx.Exec(update1, to, tick, tickId, from)
	if err != nil {
		tx.Rollback()
		return err
	}

	if !fork {
		err = e.InstallNftRevert(tx, tick, from, to, tickId, height, "", "", "", "")
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (e *MysqlClient) MintNft(tx *sql.Tx, tick, from string, tickId int64, prompt, image, imagePath, txHash string, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "MintNft", "start", "tick", tick, "from", from, "tickId", tickId)

	update := "UPDATE nft_collect SET transactions = transactions + 1, tick_sum = tick_sum + 1  WHERE tick = ?"
	_, err := tx.Exec(update, tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	query := "SELECT tick_sum FROM nft_collect WHERE tick = ?"
	row := tx.QueryRow(query, tick)
	var tickSum int64
	err = row.Scan(&tickSum)
	if err != nil {
		return err
	}

	update2 := "INSERT INTO nft_collect_address (tick, tick_id, prompt, image, image_path, holder_address, deploy_hash) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(update2, tick, tickSum, prompt, image, imagePath, from, txHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	if !fork {
		err = e.InstallNftRevert(tx, tick, "", from, tickSum, height, prompt, image, imagePath, txHash)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func (e *MysqlClient) BurnNft(tx *sql.Tx, tick, from string, tickId int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "BurnNft", "start", "tick", tick, "from", from, "tickId", tickId)
	update := "UPDATE nft_collect SET transactions = transactions + 1, tick_sum = tick_sum - 1 WHERE tick = ?"
	_, err := tx.Exec(update, tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	update1 := "DELETE FROM nft_collect_address WHERE tick = ? AND tick_id = ? AND holder_address = ?"
	_, err = tx.Exec(update1, tick, tickId, from)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (e *MysqlClient) BurnBox(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Burn", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := e.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	if count.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count < amount")
	}

	if count1.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count1 < amount")
	}

	sum := big.NewInt(0).Sub(count, amt)
	sum1 := big.NewInt(0).Sub(count1, amt)

	err = e.UpdateAddressBalanceMint(tx, tick, sum, sum1, from, true)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = e.InstallDrc20Revert(tx, tick, from, "", amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *MysqlClient) MintStakeReward(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "MintStake", "start", "tick", tick, "from", from, "amt", amt.String())

	stakec, err := e.FindStakeCollectByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectByTick err: %s tick: %s", err.Error(), tick)
	}

	update := "UPDATE stake_collect SET amt = ? WHERE tick = ?"
	_, err = tx.Exec(update, big.NewInt(0).Add(stakec.Amt.Int(), amt).String(), tick)
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollect err: %s tick: %s", err.Error(), tick)
	}

	count, _, _, err := e.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		if err != ErrNotFound {
			return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, from)
		}
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	sum := big.NewInt(0).Add(count, amt)
	sum1 := big.NewInt(0).Add(count1, amt)

	err = e.UpdateAddressBalanceMint(tx, tick, sum, sum1, from, false)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = e.InstallDrc20Revert(tx, tick, "", from, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *MysqlClient) BurnStakeReward(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	log.Info("explorer", "Burn", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := e.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := e.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	if count.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count < amount")
	}

	if count1.Cmp(amt) == -1 {
		return fmt.Errorf("forkBack count1 < amount")
	}

	sum := big.NewInt(0).Sub(count, amt)
	sum1 := big.NewInt(0).Sub(count1, amt)

	err = e.UpdateAddressBalanceMint(tx, tick, sum, sum1, from, true)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = e.InstallDrc20Revert(tx, tick, from, "", amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *MysqlClient) StakeGetRewardRouter(holderAddress, tick string) ([]*models.HolderReward, error) {

	addressResults, _, err := e.FindDrc20AllByAddress(stakePoolAddress, 2000, 0)
	if err != nil {
		return nil, err
	}

	stakeAddressCollect, err := e.FindStakeCollectAddressByTick(holderAddress, tick)
	if err != nil {
		return nil, err
	}

	if stakeAddressCollect == nil {
		return nil, fmt.Errorf("stakeAddressCollect is nil")
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

	unixPool, err := e.FindDrc20AddressInfoByTick("UNIX", stakePoolAddress)
	if err != nil {
		return nil, err
	}

	for _, ar := range addressResults {
		if ar.Tick == "UNIX" || ar.Tick == "WDOGE(WRAPPED-DOGE)" {
			continue
		}

		amt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, reward), unixPool)
		receivedAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, stakeAddressCollect.ReceivedReward.Int()), unixPool)
		TotalAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, stakeAddressCollect.Reward.Int()), unixPool)
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
