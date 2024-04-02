package storage

import (
	"database/sql"
	"fmt"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

func (c *DBClient) InstallStakeAddressInfo(info *utils.AddressInfo) error {
	exec := "INSERT INTO stake_address_info (order_id, address, public_key, private_key, receive_address, fee_address) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := c.SqlDB.Exec(exec, info.OrderId, info.Address, info.PubKey, info.PrveWif.String(), info.ReceiveAddress, info.FeeAddress)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *DBClient) InstallStakeInfo(stake *utils.StakeInfo) error {
	query := "INSERT INTO stake_info (order_id, op, tick, amt, fee_address, fee_tx_hash, holder_address, stake_tx_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := c.SqlDB.Exec(query, stake.OrderId, stake.Op, stake.Tick, stake.Amt.String(), stake.FeeAddress, stake.FeeTxHash, stake.HolderAddress, stake.StakeTxHash)
	return err
}

func (c *DBClient) InstallStakeRevert(tx *sql.Tx, tick, from, to string, amt *big.Int, height int64) error {
	exec := "INSERT INTO stake_revert (tick, from_address, to_address, amt, block_number) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, tick, from, to, amt.String(), height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) InstallRewardStakeRevert(tx *sql.Tx, tick, from, to string, reward *big.Int, height int64) error {
	exec := "INSERT INTO stake_reward_revert (tick, from_address, to_address, reward, block_number) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, tick, from, to, reward.String(), height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) InstallStakeRewardInfo(tx *sql.Tx, order_id, tick, from, to string, amt *big.Int, height int64) error {
	exec := "INSERT INTO stake_reward_info (order_id, tick, from_address, to_address, amt, block_number) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, order_id, tick, from, to, amt.String(), height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateStakeInfo(stake *utils.StakeInfo) error {
	query := "update stake_info set fee_tx_hash = ?, fee_tx_index = ?, fee_block_hash = ?, fee_block_number = ?, stake_tx_hash = ?, stake_tx_raw = ? where order_id = ?"
	_, err := c.SqlDB.Exec(query, stake.FeeTxHash, stake.FeeTxIndex, stake.FeeBlockHash, stake.FeeBlockNumber, stake.StakeTxHash, stake.StakeTxRaw, stake.OrderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateStakeInfoErr(orderId, errInfo string) error {
	query := "update stake_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.SqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateStakeInfoFork(tx *sql.Tx, height int64) error {
	query := "update stake_info set stake_block_number = 0, stake_block_hash = '', order_status = 0 where stake_block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelStakeRevert(tx *sql.Tx, height int64) error {
	query := "delete from stake_revert where block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelStakeRevert2(tx *sql.Tx, height int64) error {
	query := "delete from stake_revert where block_number < ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelStakeRewardRevert(tx *sql.Tx, height int64) error {
	query := "delete from stake_reward_revert where block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelStakeRewardRevert2(tx *sql.Tx, height int64) error {
	query := "delete from stake_reward_revert where block_number < ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelStakeRewardInfo(tx *sql.Tx, height int64) error {
	query := "delete from stake_reward_info where block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindStakeRevertByNumber(height int64) ([]*utils.StakeRevert, error) {
	query := "SELECT tick, from_address, to_address, amt FROM stake_revert where block_number > ? order by id desc "
	rows, err := c.SqlDB.Query(query, height)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	reverts := make([]*utils.StakeRevert, 0)
	for rows.Next() {
		revert := &utils.StakeRevert{}
		var amt string
		err := rows.Scan(&revert.Tick, &revert.FromAddress, &revert.ToAddress, &amt)
		if err != nil {
			return nil, err
		}

		revert.Amt, _ = utils.ConvetStr(amt)
		reverts = append(reverts, revert)
	}

	return reverts, nil
}

func (c *DBClient) FindRewardStakeRevertByNumber(height int64) ([]*utils.StakeRevert, error) {
	query := "SELECT tick, from_address, to_address, reward FROM stake_reward_revert where block_number > ? order by id desc "
	rows, err := c.SqlDB.Query(query, height)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	reverts := make([]*utils.StakeRevert, 0)
	for rows.Next() {
		revert := &utils.StakeRevert{}
		var reward string
		err := rows.Scan(&revert.Tick, &revert.FromAddress, &revert.ToAddress, &reward)
		if err != nil {
			return nil, err
		}

		revert.Amt, _ = utils.ConvetStr(reward)
		reverts = append(reverts, revert)
	}

	return reverts, nil
}

func (c *DBClient) FindStakeAddressInfo(orderId string) (*utils.AddressInfo, error) {
	query := "SELECT address, public_key, private_key, receive_address, order_id from stake_address_info where order_id = ?"
	rows, err := c.SqlDB.Query(query, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addressInfo := &utils.AddressInfo{}
	if rows.Next() {
		var private_key string
		err := rows.Scan(&addressInfo.Address, &addressInfo.PubKey, &private_key, &addressInfo.ReceiveAddress, &addressInfo.OrderId)
		if err != nil {
			return nil, err
		}
		private_key_wif, err := btcutil.DecodeWIF(private_key)
		if err != nil {
			return nil, err
		}
		addressInfo.PrveWif = private_key_wif
		return addressInfo, nil
	}
	return nil, nil
}

func (c *DBClient) FindStakeInfoByTxHash(txHash string) (*utils.StakeInfo, error) {
	query := "SELECT order_id, op, tick, amt, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, stake_tx_hash, stake_tx_raw, stake_block_hash, stake_block_number, fee_address, holder_address, update_date, create_date FROM stake_info where stake_tx_hash = ?"
	rows, err := c.SqlDB.Query(query, txHash)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		stake := &utils.StakeInfo{}
		var amt string
		err := rows.Scan(&stake.OrderId, &stake.Op, &stake.Tick, &amt, &stake.FeeTxHash, &stake.FeeTxIndex, &stake.FeeBlockHash, &stake.FeeBlockNumber, &stake.StakeTxHash, &stake.StakeTxRaw, &stake.StakeBlockHash, &stake.StakeBlockNumber, &stake.FeeAddress, &stake.HolderAddress, &stake.UpdateDate, &stake.CreateDate)
		if err != nil {
			return nil, err
		}
		stake.Amt, _ = utils.ConvetStr(amt)
		return stake, nil
	}
	return nil, nil
}

func (c *DBClient) FindStakeInfo(orderId, op, tick, holder_address string, limit, offset int64) ([]*utils.StakeInfo, int64, error) {
	query := "SELECT order_id, op, tick, amt, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, stake_tx_hash, stake_block_hash, stake_block_number, fee_address, holder_address, order_status,  update_date, create_date  FROM stake_info  "

	where := "where"
	whereAges := []any{}
	if orderId != "" {
		where += " order_id = ? "
		whereAges = append(whereAges, orderId)
	}

	if op != "" {
		if where != "where" {
			where += " and "
		}
		where += "  op = ? "
		whereAges = append(whereAges, op)
	}

	if tick != "" {
		if where != "where" {
			where += " and "
		}
		where += "  tick = ? "
		whereAges = append(whereAges, tick)
	}

	if holder_address != "" {
		if where != "where" {
			where += " and "
		}
		where += "  holder_address = ? "
		whereAges = append(whereAges, holder_address)
	}

	if where == "where" {
		where = ""
	}

	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ? "

	whereAges1 := append(whereAges, limit)
	whereAges1 = append(whereAges1, offset)

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	stakes := make([]*utils.StakeInfo, 0)
	for rows.Next() {
		stake := &utils.StakeInfo{}
		var amt string
		err := rows.Scan(&stake.OrderId, &stake.Op, &stake.Tick, &amt, &stake.FeeTxHash, &stake.FeeTxIndex, &stake.FeeBlockHash, &stake.FeeBlockNumber, &stake.StakeTxHash, &stake.StakeBlockHash, &stake.StakeBlockNumber, &stake.FeeAddress, &stake.HolderAddress, &stake.OrderStatus, &stake.UpdateDate, &stake.CreateDate)
		if err != nil {
			return nil, 0, err
		}

		stake.Amt, _ = utils.ConvetStr(amt)
		stakes = append(stakes, stake)
	}

	query1 := "SELECT count(order_id)  FROM stake_info "
	rows1, err := c.SqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return stakes, total, nil
}

func (c *DBClient) FindStakeInfoByFee(feeAddress string) (*utils.StakeInfo, error) {
	query := "SELECT order_id, op, tick, amt, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, stake_tx_hash, stake_block_hash, stake_block_number, fee_address, holder_address, order_status, update_date, create_date  FROM stake_info  where fee_address = ? and fee_tx_hash = ''"
	rows, err := c.SqlDB.Query(query, feeAddress)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		nft := &utils.StakeInfo{}
		var amt string
		err := rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &amt, &nft.FeeTxHash, &nft.FeeTxIndex, &nft.FeeBlockHash, &nft.FeeBlockNumber, &nft.StakeTxHash, &nft.StakeBlockHash, &nft.StakeBlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)

		nft.Amt, _ = utils.ConvetStr(amt)
		if err != nil {
			return nil, err
		}
		return nft, nil
	}
	return nil, nil
}

func (c *DBClient) FindStakeAll() ([]*utils.StakeCollect, int64, error) {
	query := `SELECT ci.tick,
				   ci.amt,
				   ci.reward,
				   COUNT(di.tick)
			FROM stake_collect AS ci
					 LEFT JOIN stake_collect_address AS di ON ci.tick = di.tick
			GROUP BY ci.tick, ci.amt, ci.reward`

	rows, err := c.SqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*utils.StakeCollect
	for rows.Next() {
		result := &utils.StakeCollect{}
		var amt string
		var reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &result.Holders)
		if err != nil {
			return nil, 0, err
		}

		result.Amt, _ = utils.ConvetStr(amt)
		result.Reward, _ = utils.ConvetStr(reward)

		results = append(results, result)
	}

	query1 := "SELECT COUNT(tick) FROM stake_collect "

	rows1, err := c.SqlDB.Query(query1)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}
	return results, total, nil
}

func (c *DBClient) FindStakeCollectByTick(tx *sql.Tx, tick string) (*utils.StakeCollect, error) {
	query := `SELECT tick, amt, reward FROM stake_collect WHERE tick = ?`
	rows, err := tx.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		stake := &utils.StakeCollect{}
		var amt string
		var reward string
		err := rows.Scan(&stake.Tick, &amt, &reward)
		if err != nil {
			return nil, err
		}

		stake.Amt, _ = utils.ConvetStr(amt)
		stake.Reward, _ = utils.ConvetStr(reward)
		return stake, nil
	}

	return nil, ErrNotFound
}

func (c *DBClient) FindStakeCollect() ([]*utils.StakeCollect, error) {
	query := `SELECT tick, amt, reward FROM stake_collect`
	rows, err := c.SqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*utils.StakeCollect
	for rows.Next() {
		stake := &utils.StakeCollect{}
		var amt string
		var reward string
		err := rows.Scan(&stake.Tick, &amt, &reward)
		if err != nil {
			return nil, err
		}

		stake.Amt, _ = utils.ConvetStr(amt)
		stake.Reward, _ = utils.ConvetStr(reward)
		results = append(results, stake)
	}

	return results, nil
}

func (c *DBClient) FindStakeByAddressTick(holder_address, tick string, limit, offset int64) ([]*utils.StakeCollectAddress, int64, error) {
	query := " SELECT tick, amt, reward, received_reward, holder_address, update_date, create_date FROM stake_collect_address "

	where := "where"
	whereAges := []any{}
	if holder_address != "" {
		where += " holder_address = ? "
		whereAges = append(whereAges, holder_address)
	}

	if tick != "" {
		if where != "where" {
			where += " and "
		}
		where += "  tick = ? "
		whereAges = append(whereAges, tick)
	}

	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ? "

	whereAgesLim := append(whereAges, limit)
	whereAgesLim = append(whereAgesLim, offset)

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAgesLim...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	stakeas := make([]*utils.StakeCollectAddress, 0)
	for rows.Next() {
		result := &utils.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, 0, err
		}

		result.Amt, _ = utils.ConvetStr(amt)
		result.Reward, _ = utils.ConvetStr(reward)
		result.ReceivedReward, _ = utils.ConvetStr(received_reward)
		stakeas = append(stakeas, result)
	}

	query1 := " SELECT count(id)  FROM stake_collect_address "
	rows1, err := c.SqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return stakeas, total, nil
}

func (c *DBClient) FindStakeCollectAddressAll() ([]*utils.StakeCollectAddress, error) {
	query := `
			SELECT sca.tick,
				   sca.amt,
				   sca.reward,
				   sca.received_reward,
				   sca.holder_address,
				   COALESCE(d20ai.amt_sum, 0) AS amt_sum, 
				   sca.update_date,
				   sca.create_date
			FROM stake_collect_address sca
			LEFT JOIN drc20_address_info d20ai
			ON sca.holder_address = d20ai.receive_address AND d20ai.tick = 'CARDI';
			`

	rows, err := c.SqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	stakeas := make([]*utils.StakeCollectAddress, 0)
	for rows.Next() {
		result := &utils.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		var cardiAmt string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &cardiAmt, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, err
		}
		result.Amt, _ = utils.ConvetStr(amt)
		result.Reward, _ = utils.ConvetStr(reward)
		result.ReceivedReward, _ = utils.ConvetStr(received_reward)
		result.CardiAmt, _ = utils.ConvetStr(cardiAmt)
		stakeas = append(stakeas, result)
	}

	return stakeas, nil
}

func (c *DBClient) FindStakeCollectAddressByTickAndHolder(tx *sql.Tx, holder_address, tick string) (*utils.StakeCollectAddress, error) {
	query := " SELECT tick, amt, reward, received_reward, holder_address, update_date, create_date FROM stake_collect_address WHERE holder_address = ? and tick = ?"

	rows, err := tx.Query(query, holder_address, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		result := &utils.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, err
		}
		result.Amt, _ = utils.ConvetStr(amt)
		result.Reward, _ = utils.ConvetStr(reward)
		result.ReceivedReward, _ = utils.ConvetStr(received_reward)
		return result, nil
	}

	return nil, nil
}

func (c *DBClient) FindStakeCollectAddressByTick(holder_address, tick string) (*utils.StakeCollectAddress, error) {
	query := " SELECT tick, amt, reward, received_reward, holder_address, update_date, create_date FROM stake_collect_address WHERE holder_address = ? and tick = ?"

	rows, err := c.SqlDB.Query(query, holder_address, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		result := &utils.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, err
		}
		result.Amt, _ = utils.ConvetStr(amt)
		result.Reward, _ = utils.ConvetStr(reward)
		result.ReceivedReward, _ = utils.ConvetStr(received_reward)
		return result, nil
	}

	return nil, nil
}

func (c *DBClient) FindStakeCollectReward() ([]*utils.StakeCollectReward, error) {
	query := `SELECT tick, reward FROM stake_collect_reward `
	rows, err := c.SqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*utils.StakeCollectReward
	for rows.Next() {
		result := &utils.StakeCollectReward{}
		var reward string
		err := rows.Scan(&result.Tick, &reward)
		if err != nil {
			return nil, err
		}

		result.Reward, _ = utils.ConvetStr(reward)
		results = append(results, result)
	}

	return results, nil
}

func (c *DBClient) FindStakeRewardInfo(orderId string) ([]*utils.StakeRevert, error) {
	query := `SELECT tick, to_address, amt FROM stake_reward_info where order_id = ? `
	rows, err := c.SqlDB.Query(query, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*utils.StakeRevert
	for rows.Next() {
		result := &utils.StakeRevert{}
		var amt string
		err := rows.Scan(&result.Tick, &result.ToAddress, &amt)
		if err != nil {
			return nil, err
		}

		result.Amt, _ = utils.ConvetStr(amt)
		results = append(results, result)
	}

	return results, nil
}
