package storage_v3

import (
	"database/sql"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (c *MysqlClient) FindStakeInfo(orderId, op, tick, holder_address string, limit, offset int64) ([]*models.StakeInfo, int64, error) {
	query := "SELECT order_id, op, tick, amt, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, order_status, update_date, create_date  FROM stake_info  "

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

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	stakes := make([]*models.StakeInfo, 0)
	for rows.Next() {
		stake := &models.StakeInfo{}
		var amt string
		err := rows.Scan(&stake.OrderId, &stake.Op, &stake.Tick, &amt, &stake.FeeTxHash, &stake.TxHash, &stake.BlockHash, &stake.BlockNumber, &stake.FeeAddress, &stake.HolderAddress, &stake.OrderStatus, &stake.UpdateDate, &stake.CreateDate)
		if err != nil {
			return nil, 0, err
		}

		stake.Amt, _ = utils.ConvetStringToNumber(amt)
		stakes = append(stakes, stake)
	}

	query1 := "SELECT count(order_id)  FROM stake_info "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
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

func (c *MysqlClient) FindStakeInfoByFee(feeAddress string) (*models.StakeInfo, error) {
	query := "SELECT order_id, op, tick, amt, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, order_status, update_date, create_date   FROM stake_info  where fee_address = ? and fee_tx_hash = ''"
	rows, err := c.MysqlDB.Query(query, feeAddress)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		nft := &models.StakeInfo{}
		var amt string
		err := rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &amt, &nft.FeeTxHash, &nft.TxHash, &nft.BlockHash, &nft.BlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)

		nft.Amt, _ = utils.ConvetStringToNumber(amt)
		if err != nil {
			return nil, err
		}
		return nft, nil
	}
	return nil, nil
}

func (c *MysqlClient) FindStakeAll() ([]*models.StakeCollect, int64, error) {
	query := `SELECT ci.tick,
				   ci.amt,
				   ci.reward,
				   COUNT(di.tick)
			FROM stake_collect AS ci
					 LEFT JOIN stake_collect_address AS di ON ci.tick = di.tick
			GROUP BY ci.tick, ci.amt, ci.reward`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*models.StakeCollect
	for rows.Next() {
		result := &models.StakeCollect{}
		var amt string
		var reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &result.Holders)
		if err != nil {
			return nil, 0, err
		}

		result.Amt, _ = utils.ConvetStringToNumber(amt)
		result.Reward, _ = utils.ConvetStringToNumber(reward)

		results = append(results, result)
	}

	query1 := "SELECT COUNT(tick) FROM stake_collect "

	rows1, err := c.MysqlDB.Query(query1)
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

func (c *MysqlClient) FindStakeCollectByTick(tx *sql.Tx, tick string) (*models.StakeCollect, error) {
	query := `SELECT tick, amt, reward FROM stake_collect WHERE tick = ?`
	rows, err := tx.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		stake := &models.StakeCollect{}
		var amt string
		var reward string
		err := rows.Scan(&stake.Tick, &amt, &reward)
		if err != nil {
			return nil, err
		}

		stake.Amt, _ = utils.ConvetStringToNumber(amt)
		stake.Reward, _ = utils.ConvetStringToNumber(reward)
		return stake, nil
	}

	return nil, ErrNotFound
}

func (c *MysqlClient) FindStakeCollect() ([]*models.StakeCollect, error) {
	query := `SELECT tick, amt, reward FROM stake_collect`
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.StakeCollect
	for rows.Next() {
		stake := &models.StakeCollect{}
		var amt string
		var reward string
		err := rows.Scan(&stake.Tick, &amt, &reward)
		if err != nil {
			return nil, err
		}

		stake.Amt, _ = utils.ConvetStringToNumber(amt)
		stake.Reward, _ = utils.ConvetStringToNumber(reward)
		results = append(results, stake)
	}

	return results, nil
}

func (c *MysqlClient) FindStakeByAddressTick(holder_address, tick string, limit, offset int64) ([]*models.StakeCollectAddress, int64, error) {
	query := " SELECT tick, amt, reward, received_reward, holder_address,update_date, create_date FROM stake_collect_address "

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

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAgesLim...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	stakeas := make([]*models.StakeCollectAddress, 0)
	for rows.Next() {
		result := &models.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, 0, err
		}

		result.Amt, _ = utils.ConvetStringToNumber(amt)
		result.Reward, _ = utils.ConvetStringToNumber(reward)
		result.ReceivedReward, _ = utils.ConvetStringToNumber(received_reward)
		stakeas = append(stakeas, result)
	}

	query1 := " SELECT count(id)  FROM stake_collect_address "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
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

func (c *MysqlClient) FindStakeCollectAddressAll() ([]*StakeCollectAddress, error) {
	query := `
			SELECT sca.tick,
				   sca.amt,
				   sca.reward,
				   sca.received_reward,
				   sca.holder_address,
				   COALESCE(d20ai.amt_sum, 0) AS amt_sum, 
				   UNIX_TIMESTAMP(sca.update_date),
				   UNIX_TIMESTAMP(sca.create_date)
			FROM stake_collect_address sca
			LEFT JOIN drc20_collect_address d20ai
			ON sca.holder_address = d20ai.holder_address AND d20ai.tick = 'CARDI';
			`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	stakeas := make([]*StakeCollectAddress, 0)
	for rows.Next() {
		result := &StakeCollectAddress{}
		var cardiAmt string
		err := rows.Scan(&result.Tick, &result.Amt, &result.Reward, &result.ReceivedReward, &result.HolderAddress, &cardiAmt, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, err
		}
		result.CardiAmt, _ = utils.ConvetStr(cardiAmt)
		stakeas = append(stakeas, result)
	}

	return stakeas, nil
}

func (c *MysqlClient) FindStakeCollectAddressByTickAndHolder(tx *sql.Tx, holder_address, tick string) (*models.StakeCollectAddress, error) {
	query := " SELECT tick, amt, reward, received_reward, holder_address,update_date, create_date FROM stake_collect_address WHERE holder_address = ? and tick = ?"

	rows, err := tx.Query(query, holder_address, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		result := &models.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, err
		}
		result.Amt, _ = utils.ConvetStringToNumber(amt)
		result.Reward, _ = utils.ConvetStringToNumber(reward)
		result.ReceivedReward, _ = utils.ConvetStringToNumber(received_reward)
		return result, nil
	}

	return nil, nil
}

func (c *MysqlClient) FindStakeCollectAddressByTick(holder_address, tick string) (*models.StakeCollectAddress, error) {
	query := " SELECT tick, amt, reward, received_reward, holder_address,update_date, create_date FROM stake_collect_address WHERE holder_address = ? and tick = ?"

	rows, err := c.MysqlDB.Query(query, holder_address, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		result := &models.StakeCollectAddress{}
		var amt string
		var reward string
		var received_reward string
		err := rows.Scan(&result.Tick, &amt, &reward, &received_reward, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, err
		}
		result.Amt, _ = utils.ConvetStringToNumber(amt)
		result.Reward, _ = utils.ConvetStringToNumber(reward)
		result.ReceivedReward, _ = utils.ConvetStringToNumber(received_reward)
		return result, nil
	}

	return nil, nil
}

func (c *MysqlClient) FindStakeCollectReward() ([]*models.StakeCollectReward, error) {
	query := `SELECT tick, reward FROM stake_collect_reward `
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.StakeCollectReward
	for rows.Next() {
		result := &models.StakeCollectReward{}
		var reward string
		err := rows.Scan(&result.Tick, &reward)
		if err != nil {
			return nil, err
		}

		result.Reward, _ = utils.ConvetStringToNumber(reward)
		results = append(results, result)
	}

	return results, nil
}

func (c *MysqlClient) FindStakeRewardInfo(orderId string) ([]*models.StakeRevert, error) {
	query := `SELECT tick, to_address, amt FROM stake_reward_info where order_id = ? `
	rows, err := c.MysqlDB.Query(query, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.StakeRevert
	for rows.Next() {
		result := &models.StakeRevert{}
		var amt string
		err := rows.Scan(&result.Tick, &result.ToAddress, &amt)
		if err != nil {
			return nil, err
		}

		result.Amt, _ = utils.ConvetStringToNumber(amt)
		results = append(results, result)
	}

	return results, nil
}
