package explorer

import (
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (e Explorer) stakeDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*utils.StakeInfo, error) {
	param := &utils.StakeParams{}
	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	stake, err := utils.ConvertStake(param)
	if err != nil {
		return nil, fmt.Errorf("ConvertWDoge err: %s", err.Error())
	}

	if len(tx.Vout) < 1 {
		return nil, fmt.Errorf("op error, vout length is not 0")
	}

	stake.OrderId = uuid.New().String()
	stake.FeeTxHash = tx.Vin[0].Txid
	stake.StakeTxHash = tx.Txid
	stake.StakeBlockHash = tx.BlockHash
	stake.StakeBlockNumber = number

	stake.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, chainNetworkErr
	}

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, chainNetworkErr
	}

	if stake.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("The address is not the same as the previous transaction")
	}

	stake1, err := e.dbc.FindStakeInfoByTxHash(stake.StakeTxHash)
	if err != nil {
		return nil, fmt.Errorf("FindWDogeInfoByTxHash err: %s", err.Error())
	}

	if stake1 != nil {
		if stake1.StakeBlockNumber != 0 {
			return nil, fmt.Errorf("stake already exist or err %s", stake1.StakeTxHash)
		}
		stake.OrderId = stake1.OrderId
		return stake, nil
	} else {
		if err := e.dbc.InstallStakeInfo(stake); err != nil {
			return nil, fmt.Errorf("InstallWDogeInfo err: %s", err.Error())
		}
	}

	return stake, nil
}

func (e Explorer) stakeStake(stake *utils.StakeInfo) error {
	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(stake.Tick+"--STAKE"), &chaincfg.MainNetParams)

	err = e.dbc.Transfer(tx, stake.Tick, stake.HolderAddress, reservesAddress.String(), stake.Amt, false, stake.StakeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = e.dbc.StakeStake(tx, stake.Tick, stake.HolderAddress, stake.Amt, false, stake.StakeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	exec := "update stake_info set stake_block_hash = ?, stake_block_number = ?, order_status = 0 where stake_tx_hash = ?"
	_, err = tx.Exec(exec, stake.StakeBlockHash, stake.StakeBlockNumber, stake.StakeTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (e Explorer) stakeUnStake(stake *utils.StakeInfo) error {

	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(stake.Tick+"--STAKE"), &chaincfg.MainNetParams)

	err = e.dbc.Transfer(tx, stake.Tick, reservesAddress.String(), stake.HolderAddress, stake.Amt, false, stake.StakeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = e.dbc.StakeUnStake(tx, stake.Tick, stake.HolderAddress, stake.Amt, false, stake.StakeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	exec := "update stake_info set stake_block_hash = ?, stake_block_number = ?, order_status = 0 where stake_tx_hash = ?"
	_, err = tx.Exec(exec, stake.StakeBlockHash, stake.StakeBlockNumber, stake.StakeTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (e Explorer) stakeGetAllReward(stake *utils.StakeInfo) error {

	tx, err := e.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	rewards, err := e.dbc.StakeGetReward(stake.HolderAddress, stake.Tick)
	if err != nil {
		return err
	}

	for _, reward := range rewards {
		err = e.dbc.Transfer(tx, reward.Tick, stakePoolAddress, stake.HolderAddress, reward.Reward, false, stake.StakeBlockNumber)
		if err != nil {
			tx.Rollback()
			return err
		}
		err := e.dbc.InstallStakeRewardInfo(tx, stake.OrderId, reward.Tick, stakePoolAddress, stake.HolderAddress, reward.Reward, stake.StakeBlockNumber)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = e.dbc.StakeReward(tx, stake.HolderAddress, stake.Tick, stake.StakeBlockNumber)
	if err != nil {
		return err
	}

	exec := "update stake_info set stake_block_hash = ?, stake_block_number = ?, order_status = 0 where stake_tx_hash = ?"
	_, err = tx.Exec(exec, stake.StakeBlockHash, stake.StakeBlockNumber, stake.StakeTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
