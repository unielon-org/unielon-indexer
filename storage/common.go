package storage

import (
	"database/sql"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

func (c *DBClient) Transfer(tx *sql.Tx, tick, from, to string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "Transfer", "start", "tick", tick, "from", from, "to", to, "amt", amt.String(), "fork", fork)

	count1, err := c.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if amt.Cmp(count1) > 0 {
		return fmt.Errorf("Transfer amt > count: %s tick: %s from : %s  amt : %s  count : %s  ", amt.String(), tick, from, amt.String(), count1.String())
	}

	count2, err := c.FindSwapDrc20AddressInfoByTick(tx, tick, to)
	if err != nil {
		if err != ErrNotFound {
			return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s to : %s", err.Error(), tick, to)
		}
		log.Debug("explorer", "Transfer", fmt.Sprintf("tick: %s to : %s", tick, to))
		count2 = big.NewInt(0)
	}

	sub := big.NewInt(0).Sub(count1, amt)
	add := big.NewInt(0).Add(count2, amt)

	err = c.UpdateAddressBalanceTran(tx, tick, sub, from, add, to, fork)
	if err != nil {
		return fmt.Errorf("Transfer UpdateAddressBalanceTran err: %s tick: %s from : %s to : %s", err.Error(), tick, from, to)
	}

	if !fork {
		err = c.InstallRevert(tx, tick, from, to, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) Mint(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "Mint", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := c.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := c.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		if err != ErrNotFound {
			return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, from)
		}
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	sum := big.NewInt(0).Add(count, amt)
	sum1 := big.NewInt(0).Add(count1, amt)

	err = c.UpdateAddressBalanceMintOrBurn(tx, tick, sum, sum1, from, false)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}
	if !fork {
		err = c.InstallRevert(tx, tick, "", from, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) Burn(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "Burn", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := c.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := c.FindSwapDrc20AddressInfoByTick(tx, tick, from)
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

	err = c.UpdateAddressBalanceMintOrBurn(tx, tick, sum, sum1, from, true)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = c.InstallRevert(tx, tick, from, "", amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) TransferNft(tx *sql.Tx, tick, from, to string, tickId int64, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
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
		err = c.InstallNftRevert(tx, tick, from, to, tickId, height, "", "", "")
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (c *DBClient) MintNft(tx *sql.Tx, tick, from string, tickId int64, prompt, image, txHash string, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
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

	update2 := "INSERT INTO nft_collect_address (tick, tick_id, prompt, image, holder_address, deploy_hash) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(update2, tick, tickSum, prompt, image, from, txHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	if !fork {
		err = c.InstallNftRevert(tx, tick, "", from, tickId, height, prompt, image, txHash)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func (c *DBClient) BurnNft(tx *sql.Tx, tick, from string, tickId int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
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

func (c *DBClient) MintStakeReward(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "MintStake", "start", "tick", tick, "from", from, "amt", amt.String())

	stakec, err := c.FindStakeCollectByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectByTick err: %s tick: %s", err.Error(), tick)
	}

	// StakeCollect
	update := "UPDATE stake_collect SET amt = ? WHERE tick = ?"
	_, err = tx.Exec(update, big.NewInt(0).Add(stakec.Amt, amt).String(), tick)
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollect err: %s tick: %s", err.Error(), tick)
	}

	count, _, _, err := c.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := c.FindSwapDrc20AddressInfoByTick(tx, tick, from)
	if err != nil {
		if err != ErrNotFound {
			return fmt.Errorf("Transfer FindDrc20AddressInfoByTick err: %s tick: %s from : %s", err.Error(), tick, from)
		}
		log.Debug("explorer", "Mint", fmt.Sprintf("tick: %s from : %s", tick, from))
		count1 = big.NewInt(0)
	}

	sum := big.NewInt(0).Add(count, amt)
	sum1 := big.NewInt(0).Add(count1, amt)

	err = c.UpdateAddressBalanceMintOrBurn(tx, tick, sum, sum1, from, false)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = c.InstallRevert(tx, tick, "", from, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) BurnStakeReward(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "Burn", "start", "tick", tick, "from", from, "amt", amt.String())

	count, _, _, err := c.FindSwapDrc20InfoByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("Mint FindDrc20InfoSumByTick err: %s tick: %s", err.Error(), tick)
	}

	count1, err := c.FindSwapDrc20AddressInfoByTick(tx, tick, from)
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

	err = c.UpdateAddressBalanceMintOrBurn(tx, tick, sum, sum1, from, true)
	if err != nil {
		return fmt.Errorf("Mint UpdateAddressBalanceMint err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if !fork {
		err = c.InstallRevert(tx, tick, from, "", amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) StakeStake(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "stake", "start", "tick", tick, "from", from, "amt", amt.String())

	stakec, err := c.FindStakeCollectByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectByTick err: %s tick: %s", err.Error(), tick)
	}

	update := "UPDATE stake_collect SET amt = ? WHERE tick = ?"
	_, err = tx.Exec(update, big.NewInt(0).Add(stakec.Amt, amt).String(), tick)
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollect err: %s tick: %s", err.Error(), tick)
	}

	stakeca, err := c.FindStakeCollectAddressByTickAndHolder(tx, from, tick)
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	if stakeca == nil {
		insert := "INSERT INTO stake_collect_address (tick, holder_address, amt, reward) VALUES (?, ?, ?, ?)"
		_, err = tx.Exec(insert, tick, from, amt.String(), "0")
		if err != nil {
			return err
		}
	} else {
		update := "UPDATE stake_collect_address SET amt = ? WHERE tick = ? AND holder_address = ?"
		_, err = tx.Exec(update, big.NewInt(0).Add(stakeca.Amt, amt).String(), tick, from)
		if err != nil {
			return err
		}
	}

	if !fork {
		err = c.InstallStakeRevert(tx, tick, "", from, amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) StakeUnStake(tx *sql.Tx, tick, from string, amt *big.Int, fork bool, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Info("explorer", "stake", "start", "tick", tick, "from", from, "amt", amt.String())

	stakec, err := c.FindStakeCollectByTick(tx, tick)
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectByTick err: %s tick: %s", err.Error(), tick)
	}

	amt0 := big.NewInt(0).Sub(stakec.Amt, amt)
	if amt0.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("StakeStake amt0 < 0 err: %s tick: %s", err.Error(), tick)
	}

	update := "UPDATE stake_collect SET amt = ? WHERE tick = ?"
	_, err = tx.Exec(update, amt0.String(), tick)
	if err != nil {
		return fmt.Errorf("StakeStake UpdateStakeCollect err: %s tick: %s", err.Error(), tick)
	}

	stakeca, err := c.FindStakeCollectAddressByTickAndHolder(tx, from, tick)
	if err != nil {
		return fmt.Errorf("StakeStake FindStakeCollectAddress err: %s tick: %s from : %s", err.Error(), tick, from)
	}

	amt1 := big.NewInt(0).Sub(stakeca.Amt, amt)
	if amt1.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("StakeStake amt1 < 0 err: %s tick: %s", err.Error(), tick)
	}

	if stakeca == nil {
		return fmt.Errorf("StakeStake stakeca == nil err: %s tick: %s", err.Error(), tick)
	} else {
		update := "UPDATE stake_collect_address SET amt = ? WHERE tick = ? AND holder_address = ?"
		_, err = tx.Exec(update, amt1.String(), tick, from)
		if err != nil {
			return err
		}
	}

	if !fork {
		err = c.InstallStakeRevert(tx, tick, from, "", amt, height)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) StakeReward(tx *sql.Tx, holderAddress, tick string, block_number int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	stakeAddressCollect, err := c.FindStakeCollectAddressByTickAndHolder(tx, holderAddress, tick)
	if err != nil {
		return err
	}

	if stakeAddressCollect == nil {
		return fmt.Errorf("stakeAddressCollects is nil")
	}

	update := "UPDATE stake_collect_address SET received_reward = reward WHERE tick = ? AND holder_address = ?"
	_, err = tx.Exec(update, tick, holderAddress)
	if err != nil {
		return err
	}

	reward := big.NewInt(0).Sub(stakeAddressCollect.Reward, stakeAddressCollect.ReceivedReward)
	err = c.InstallRewardStakeRevert(tx, tick, holderAddress, "", reward, block_number)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) StakeGetReward(holderAddress, tick string) ([]*utils.HolderReward, error) {

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return nil, err
	}

	addressResults, _, err := c.FindDrc20AllByAddress(stakePoolAddress, 2000, 0)
	if err != nil {
		return nil, err
	}

	stakeAddressCollect, err := c.FindStakeCollectAddressByTickAndHolder(tx, holderAddress, tick)
	if err != nil {
		return nil, err
	}

	if stakeAddressCollect == nil {
		return nil, fmt.Errorf("stakeAddressCollect is nil")
	}

	rewards := make([]*utils.HolderReward, 0)
	reward := big.NewInt(0).Sub(stakeAddressCollect.Reward, stakeAddressCollect.ReceivedReward)

	if tick == "UNIX-SWAP-WDOGE(WRAPPED-DOGE)" {
		rewards = append(rewards, &utils.HolderReward{
			Tick:   "WDOGE(WRAPPED-DOGE)",
			Reward: reward,
		})
		return rewards, nil
	}

	unixPool, err := c.FindDrc20AddressInfoByTick("UNIX", stakePoolAddress)
	if err != nil {
		return nil, err
	}

	for _, ar := range addressResults {
		if ar.Tick == "UNIX" || ar.Tick == "WDOGE(WRAPPED-DOGE)" {
			continue
		}

		//
		amt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, reward), unixPool)
		//
		receivedAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, stakeAddressCollect.ReceivedReward), unixPool)
		TotalAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, stakeAddressCollect.Reward), unixPool)
		TotalAmt = big.NewInt(0).Add(TotalAmt, amt)

		rewards = append(rewards, &utils.HolderReward{
			Tick:            ar.Tick,
			Reward:          amt,
			ReceivedReward:  receivedAmt,
			TotalRewardPool: TotalAmt,
		})
	}

	rewards = append(rewards, &utils.HolderReward{
		Tick:            "UNIX",
		Reward:          reward,
		ReceivedReward:  stakeAddressCollect.ReceivedReward,
		TotalRewardPool: big.NewInt(0).Add(stakeAddressCollect.Reward, reward),
	})

	return rewards, nil
}

func (c *DBClient) StakeGetRewardRouter(holderAddress, tick string) ([]*utils.HolderReward, error) {

	addressResults, _, err := c.FindDrc20AllByAddress(stakePoolAddress, 2000, 0)
	if err != nil {
		return nil, err
	}

	//
	stakeAddressCollect, err := c.FindStakeCollectAddressByTick(holderAddress, tick)
	if err != nil {
		return nil, err
	}

	if stakeAddressCollect == nil {
		return nil, fmt.Errorf("stakeAddressCollect is nil")
	}

	rewards := make([]*utils.HolderReward, 0)
	reward := big.NewInt(0).Sub(stakeAddressCollect.Reward, stakeAddressCollect.ReceivedReward)

	// wdoge
	if tick == "UNIX-SWAP-WDOGE(WRAPPED-DOGE)" {
		rewards = append(rewards, &utils.HolderReward{
			Tick:   "WDOGE(WRAPPED-DOGE)",
			Reward: reward,
		})
		return rewards, nil
	}

	unixPool, err := c.FindDrc20AddressInfoByTick("UNIX", stakePoolAddress)
	if err != nil {
		return nil, err
	}

	for _, ar := range addressResults {
		if ar.Tick == "UNIX" || ar.Tick == "WDOGE(WRAPPED-DOGE)" {
			continue
		}

		amt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, reward), unixPool)
		receivedAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, stakeAddressCollect.ReceivedReward), unixPool)
		TotalAmt := big.NewInt(0).Div(big.NewInt(0).Mul(ar.Amt, stakeAddressCollect.Reward), unixPool)
		TotalAmt = big.NewInt(0).Add(TotalAmt, amt)

		rewards = append(rewards, &utils.HolderReward{
			Tick:            ar.Tick,
			Reward:          amt,
			ReceivedReward:  receivedAmt,
			TotalRewardPool: TotalAmt,
		})
	}

	rewards = append(rewards, &utils.HolderReward{
		Tick:            "UNIX",
		Reward:          reward,
		ReceivedReward:  stakeAddressCollect.ReceivedReward,
		TotalRewardPool: big.NewInt(0).Add(stakeAddressCollect.Reward, reward),
	})

	return rewards, nil
}

func (c *DBClient) StakeUpdatePool(height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	stakeRewards, err := c.FindStakeCollectReward()
	if err != nil {
		return err
	}

	rewardMap := make(map[string]*big.Int)
	for _, stakeReward := range stakeRewards {
		rewardMap[stakeReward.Tick] = stakeReward.Reward
	}

	stakeCollects, err := c.FindStakeCollect()
	if err != nil {
		return err
	}

	collectMap := make(map[string]*big.Int)
	for _, stakeCollect := range stakeCollects {
		collectMap[stakeCollect.Tick] = stakeCollect.Amt
	}

	stakeAddressCollects, err := c.FindStakeCollectAddressAll()
	if err != nil {
		return err
	}

	for _, stakeAddressCollect := range stakeAddressCollects {
		amt0 := big.NewInt(0)
		if stakeAddressCollect.CardiAmt.Cmp(big.NewInt(10000000000000)) >= 0 {
			amt0 = big.NewInt(0).Div(big.NewInt(0).Mul(stakeAddressCollect.Amt, big.NewInt(50)), big.NewInt(100))
		} else if stakeAddressCollect.CardiAmt.Cmp(big.NewInt(5000000000000)) >= 0 {
			amt0 = big.NewInt(0).Div(big.NewInt(0).Mul(stakeAddressCollect.Amt, big.NewInt(35)), big.NewInt(100))
		} else if stakeAddressCollect.CardiAmt.Cmp(big.NewInt(1000000000000)) >= 0 {
			amt0 = big.NewInt(0).Div(big.NewInt(0).Mul(stakeAddressCollect.Amt, big.NewInt(25)), big.NewInt(100))
		} else if stakeAddressCollect.CardiAmt.Cmp(big.NewInt(500000000000)) >= 0 {
			amt0 = big.NewInt(0).Div(big.NewInt(0).Mul(stakeAddressCollect.Amt, big.NewInt(15)), big.NewInt(100))
		} else if stakeAddressCollect.CardiAmt.Cmp(big.NewInt(200000000000)) >= 0 {
			amt0 = big.NewInt(0).Div(big.NewInt(0).Mul(stakeAddressCollect.Amt, big.NewInt(10)), big.NewInt(100))
		}

		collectMap[stakeAddressCollect.Tick] = big.NewInt(0).Add(collectMap[stakeAddressCollect.Tick], amt0)
		stakeAddressCollect.Amt = big.NewInt(0).Add(stakeAddressCollect.Amt, amt0)
	}

	rewardSum := make(map[string]*big.Int)
	rewardSum["UNIX"] = big.NewInt(0)
	rewardSum["WDOGE(WRAPPED-DOGE)"] = big.NewInt(0)
	rewardSum["UNIX-SWAP-WDOGE(WRAPPED-DOGE)"] = big.NewInt(0)

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	for i, stakeAddressCollect := range stakeAddressCollects {
		rewardAddress := big.NewInt(0)
		if i == len(stakeAddressCollects)-1 {
			rewardAddress = big.NewInt(0).Sub(rewardMap[stakeAddressCollect.Tick], rewardSum[stakeAddressCollect.Tick])
		} else {
			rewardAddress = big.NewInt(0).Div(big.NewInt(0).Mul(stakeAddressCollect.Amt, rewardMap[stakeAddressCollect.Tick]), collectMap[stakeAddressCollect.Tick])
		}

		rewardSum[stakeAddressCollect.Tick] = big.NewInt(0).Add(rewardSum[stakeAddressCollect.Tick], rewardAddress)
		reward := rewardAddress
		rewardAddress = big.NewInt(0).Add(rewardAddress, stakeAddressCollect.Reward)

		update := "UPDATE stake_collect_address SET reward = ? WHERE tick = ? AND holder_address = ?"
		_, err = tx.Exec(update, rewardAddress.String(), stakeAddressCollect.Tick, stakeAddressCollect.HolderAddress)
		if err != nil {
			tx.Rollback()
			return err
		}

		err := c.InstallRewardStakeRevert(tx, stakeAddressCollect.Tick, "", stakeAddressCollect.HolderAddress, reward, height)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, stakeCollect := range stakeCollects {

		reward := big.NewInt(0).Add(rewardSum[stakeCollect.Tick], stakeCollect.Reward)

		update := "UPDATE stake_collect SET reward = ?, last_block = ? WHERE tick = ?"
		_, err = tx.Exec(update, reward.String(), height, stakeCollect.Tick)
		if err != nil {
			tx.Rollback()
			return err
		}

		err := c.InstallRewardStakeRevert(tx, stakeCollect.Tick, "", "collect", rewardSum[stakeCollect.Tick], height)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (c *DBClient) StakeUpdatePoolFork(tx *sql.Tx, tick, from, to string, amt *big.Int, height int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if to == "collect" {
		stakeCollect, err := c.FindStakeCollectByTick(tx, tick)
		if err != nil {
			return err
		}

		reward := big.NewInt(0).Sub(stakeCollect.Reward, amt)
		update := "UPDATE stake_collect SET reward = ? WHERE tick = ?"
		_, err = tx.Exec(update, reward.String(), tick)
		if err != nil {
			return err
		}
		return nil
	}

	if from != "" {
		stakeAddressCollect, err := c.FindStakeCollectAddressByTickAndHolder(tx, from, tick)
		if err != nil {
			return err
		}

		reward := big.NewInt(0).Sub(stakeAddressCollect.ReceivedReward, amt)
		update := "UPDATE stake_collect_address SET received_reward = ? WHERE tick = ? AND holder_address = ?"
		_, err = tx.Exec(update, reward.String(), tick, from)
		if err != nil {
			return err
		}

	}

	stakeAddressCollect, err := c.FindStakeCollectAddressByTickAndHolder(tx, to, tick)
	if err != nil {
		return err
	}

	reward := big.NewInt(0).Sub(stakeAddressCollect.Reward, amt)
	update := "UPDATE stake_collect_address SET reward = ? WHERE tick = ? AND holder_address = ?"
	_, err = tx.Exec(update, reward.String(), tick, to)
	if err != nil {
		return err
	}

	return nil
}
