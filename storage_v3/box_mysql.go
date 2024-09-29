package storage_v3

import (
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (c *MysqlClient) FindBoxInfo(orderId, op, tick0, tick1, holder_address string, limit, offset int64) ([]*models.BoxInfo, int64, error) {

	query := "SELECT  order_id, op, tick0, tick1, max_, amt0, liqamt, liqblock, amt1, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, order_status,update_date, create_date FROM box_info  "

	where := "where"
	whereAges := []any{}

	if orderId != "" {
		whereAges = append(whereAges, orderId)
		where += " order_id = ? "
	}

	if op != "" {
		if where != "where" {
			where += " and "
		}
		where += "  op = ? "
		whereAges = append(whereAges, op)
	}

	if tick0 != "" {
		if where != "where" {
			where += " and "
		}
		where += "  tick0 = ? "
		whereAges = append(whereAges, tick0)
	}

	if tick1 != "" {
		if where != "where" {
			where += " and "
		}
		where += "  tick1 = ? "
		whereAges = append(whereAges, tick1)
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

	exs := []*models.BoxInfo{}
	for rows.Next() {
		ex := &models.BoxInfo{}
		var max, amt0, liqamt, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &ex.FeeTxHash, &ex.TxHash, &ex.BlockHash, &ex.BlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Max, _ = utils.ConvetStringToNumber(max)
		ex.Amt0, _ = utils.ConvetStringToNumber(amt0)
		ex.LiqAmt, _ = utils.ConvetStringToNumber(liqamt)
		ex.Amt1, _ = utils.ConvetStringToNumber(amt1)
		exs = append(exs, ex)
	}

	var total int64
	err = c.MysqlDB.QueryRow("SELECT COUNT(order_id) FROM box_info "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *MysqlClient) FindBoxCollect(tick0, tick1, holder_address string, limit, offset int64) ([]*models.BoxCollect, int64, error) {
	query := "SELECT  tick0, tick1, max_, amt0, liqamt, liqblock, amt1, amt0_finish, liqamt_finish, holder_address, reserves_address,update_date, create_date  FROM box_collect  "
	where := "where"
	whereAges := []any{}

	if tick0 != "" {
		where += " tick0 = ? "
		whereAges = append(whereAges, tick0)
	}

	if tick1 != "" {
		if where != "where" {
			where += " and "
		}
		where += "  tick1 = ? "
		whereAges = append(whereAges, tick1)
	}

	if holder_address != "" {
		if where != "where" {
			where += " and "
		}
		where += "  holder_address = ? "
		whereAges = append(whereAges, holder_address)
	}

	if where != "where" {
		where += "and "
	}

	where += "  is_del = 0 "

	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ? "

	whereAges1 := append(whereAges, limit)
	whereAges1 = append(whereAges1, offset)

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*models.BoxCollect{}
	for rows.Next() {
		ex := &models.BoxCollect{}
		var max, liqamt, amt0, amt1, amt0_finish, liqamt_finish string
		err := rows.Scan(&ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &amt0_finish, &liqamt_finish, &ex.HolderAddress, &ex.ReservesAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Max, _ = utils.ConvetStringToNumber(max)
		ex.LiqAmt, _ = utils.ConvetStringToNumber(liqamt)
		ex.Amt0, _ = utils.ConvetStringToNumber(amt0)
		ex.Amt1, _ = utils.ConvetStringToNumber(amt1)
		ex.Amt0Finish, _ = utils.ConvetStringToNumber(amt0_finish)
		ex.LiqAmtFinish, _ = utils.ConvetStringToNumber(liqamt_finish)
		exs = append(exs, ex)
	}

	var total int64
	err = c.MysqlDB.QueryRow("SELECT COUNT(id) FROM box_collect "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}
