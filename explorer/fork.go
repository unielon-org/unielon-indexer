package explorer

import (
	"errors"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
)

func (e *Explorer) forkBack() error {

	height := e.fromBlock

	blockHash, err := e.node.GetBlockHash(e.fromBlock)
	if err != nil {
		return err
	}

	block, err := e.node.GetBlockVerboseBool(blockHash)
	if err != nil {
		return err
	}

	localHash, _ := e.dbc.FindBlockByHeight(height - 1)
	if localHash == "" {
		e.dbc.UpdateBlock(height-1, block.PreviousHash)
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
			localHash, _ = e.dbc.FindBlockByHeight(height)
			if localHash == "" {
				return errors.New("localHash is nil")
			}
		}

		err := e.fork(height)
		if err != nil {
			return err
		}

		e.fromBlock = height
		log.Warn("forkBack End", "height", height)
	}

	return nil
}

func (e *Explorer) fork(height int64) error {

	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = e.dbc.UpdateCardinalsInfoFork(tx, height)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("swapFork UpdateCardinalsInfoFork error: %v", err)
	}

	err = e.dbc.UpdateSwapInfoFork(tx, height)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("swapFork UpdateSwapInfoFork error: %v", err)
	}

	err = e.dbc.UpdateWDogeInfoFork(tx, height)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("swapFork UpdateWDogeInfoFork error: %v", err)
	}

	err = e.dbc.UpdateNftInfoFork(tx, height)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("fork UpdateNftInfoFork error: %v", err)
	}

	swapReverts, err := e.dbc.FindRevertByNumber(height)
	if err != nil {
		return fmt.Errorf("swapFork FindSwapRevertByNumber error: %v", err)
	}

	for _, revert := range swapReverts {
		if revert.FromAddress == "" {
			err = e.dbc.Burn(tx, revert.Tick, revert.ToAddress, revert.Amt, true, height)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("swapFork Burn error: %v", err)
			}
		}

		if revert.ToAddress == "" {
			err = e.dbc.Mint(tx, revert.Tick, revert.FromAddress, revert.Amt, true, height)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("swapFork Mint error: %v", err)
			}
		}

		err = e.dbc.Transfer(tx, revert.Tick, revert.ToAddress, revert.FromAddress, revert.Amt, true, height)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("swapFork Transfer error: %v", err)
		}
	}

	err = e.dbc.DelRevert(tx, height)
	if err != nil {
		tx.Rollback()
		return err
	}

	nftReverts, err := e.dbc.FindNftRevertByNumber(height)
	if err != nil {
		return fmt.Errorf("swapFork FindSwapRevertByNumber error: %v", err)
	}

	for _, revert := range nftReverts {
		if revert.FromAddress == "" {
			err = e.dbc.BurnNft(tx, revert.Tick, revert.ToAddress, revert.TickId)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("swapFork Burn error: %v", err)
			}
		}

		if revert.ToAddress == "" {
			err = e.dbc.MintNft(tx, revert.Tick, revert.FromAddress, revert.TickId, revert.Prompt, revert.Image, revert.DeployHash, true, height)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("swapFork Mint error: %v", err)
			}
		}

		err = e.dbc.TransferNft(tx, revert.Tick, revert.ToAddress, revert.FromAddress, revert.TickId, true, height)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("swapFork Transfer error: %v", err)
		}
	}

	err = e.dbc.DelNftRevert(tx, height)
	if err != nil {
		tx.Rollback()
		return err
	}

	stakeReverts, err := e.dbc.FindStakeRevertByNumber(height)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("swapFork FindSwapRevertByNumber error: %v", err)
	}

	for _, revert := range stakeReverts {
		if revert.FromAddress == "" {
			err = e.dbc.StakeUnStake(tx, revert.Tick, revert.ToAddress, revert.Amt, true, height)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("swapFork Burn error: %v", err)
			}
		}

		if revert.ToAddress == "" {
			err = e.dbc.StakeStake(tx, revert.Tick, revert.FromAddress, revert.Amt, true, height)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("swapFork Mint error: %v", err)
			}
		}
	}

	err = e.dbc.DelStakeRevert(tx, height)
	if err != nil {
		tx.Rollback()
		return err
	}

	stakeRewardReverts, err := e.dbc.FindRewardStakeRevertByNumber(height)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("swapFork FindSwapRevertByNumber error: %v", err)
	}

	for _, revert := range stakeRewardReverts {
		err := e.dbc.StakeUpdatePoolFork(tx, revert.Tick, revert.FromAddress, revert.ToAddress, revert.Amt, height)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = e.dbc.DelStakeRewardRevert(tx, height)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = e.dbc.DelStakeRewardInfo(tx, height)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}
