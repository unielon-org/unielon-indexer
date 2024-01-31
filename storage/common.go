package storage

import (
	"database/sql"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
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

	err = c.UpdateAddressBalanceMintOrBurn(tx, tick, sum, sum1, from, fork)
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

	err = c.UpdateAddressBalanceMintOrBurn(tx, tick, sum, sum1, from, fork)
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

func (c *DBClient) BurnNft(tx *sql.Tx, tick, from string, tickId int64, fork bool, height int64) error {
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

	if !fork {
		err = c.InstallNftRevert(tx, tick, from, "", tickId, height, "", "", "")
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}
