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

		e.drc20Fork(height)
		e.swapFork(height)
		e.wdogeFork(height)

		log.Warn("forkBack End", "height", height)
	}
	e.fromBlock = height
	return nil
}
