package storage_v3

import (
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (c *MysqlClient) FindWDogeInfoById(OrderId string) (*models.WDogeInfo, error) {
	query := "SELECT  order_id, op, tick, amt, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, update_date, create_date , order_status  FROM wdoge_info where order_id = ?"
	rows, err := c.MysqlDB.Query(query, OrderId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		wdoge := &models.WDogeInfo{}
		var amt string
		err := rows.Scan(&wdoge.OrderId, &wdoge.Op, &wdoge.Tick, &amt, &wdoge.FeeTxHash, &wdoge.TxHash, &wdoge.BlockHash, &wdoge.BlockNumber, &wdoge.FeeAddress, &wdoge.HolderAddress, &wdoge.UpdateDate, &wdoge.CreateDate, &wdoge.OrderStatus)
		if err != nil {
			return nil, err
		}

		wdoge.Amt, _ = utils.ConvetStringToNumber(amt)

		return wdoge, err
	}
	return nil, nil
}

func (c *MysqlClient) FindWDogeInfo(orderId, op, holder_address string, limit, offset int64) ([]*models.WDogeInfo, int64, error) {
	query := "SELECT  order_id, op, tick, amt, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, withdraw_tx_hash, withdraw_tx_index, withdraw_block_hash, withdraw_block_number,update_date, create_date, order_status  FROM wdoge_info  "

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

	order := " order by create_date desc "
	lim := " LIMIT ? OFFSET ?"
	whereAges1 := append(whereAges, limit)
	whereAges1 = append(whereAges1, offset)

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	wdoges := make([]*models.WDogeInfo, 0)
	for rows.Next() {
		wdoge := &models.WDogeInfo{}
		var amt string
		rows.Scan(&wdoge.OrderId, &wdoge.Op, &wdoge.Tick, &amt, &wdoge.FeeTxHash, &wdoge.TxHash, &wdoge.BlockHash, &wdoge.BlockNumber, &wdoge.FeeAddress, &wdoge.HolderAddress, &wdoge.WithdrawTxHash, &wdoge.WithdrawTxIndex, &wdoge.WithdrawBlockHash, &wdoge.WithdrawBlockNumber, &wdoge.UpdateDate, &wdoge.CreateDate, &wdoge.OrderStatus)

		if err != nil {
			return nil, 0, err
		}

		wdoge.Amt, _ = utils.ConvetStringToNumber(amt)
		wdoges = append(wdoges, wdoge)
	}

	query1 := "SELECT count(order_id)  FROM wdoge_info "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return wdoges, total, nil
}

func (c *MysqlClient) UpdateWDogeInfoErr(orderId, errInfo string) error {
	query := "update wdoge_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.MysqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlClient) FindWDOGETempOrder(HolderAddress string) (int64, error) {
	query := "SELECT count(holder_address) FROM wdoge_info where fee_tx_hash = '' and holder_address = ? and create_date > NOW() - INTERVAL 10 MINUTE "
	rows, err := c.MysqlDB.Query(query, HolderAddress)
	if err != nil {
		return 0, err
	}

	defer rows.Close()
	count := int64(0)
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
