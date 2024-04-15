package storage

import (
	"database/sql"
	"fmt"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

func (c *DBClient) ExchangeCreate(ex *utils.ExchangeInfo, reservesAddress string) error {
	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	query := "INSERT INTO exchange_collect (ex_id, tick0, tick1, amt0, amt1, holder_address, reserves_address) VALUES (?, ?, ?, ?, ?, ? ,?)"
	_, err = tx.Exec(query, ex.ExId, ex.Tick0, ex.Tick1, ex.Amt0.String(), ex.Amt1.String(), ex.HolderAddress, reservesAddress)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, ex.Tick0, ex.HolderAddress, reservesAddress, ex.Amt0, false, ex.ExchangeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	exr := &utils.ExchangeRevert{
		Op:   "create",
		Tick: ex.Tick0,
		Exid: ex.ExId,
	}

	err = c.InstallExchangeRevert(tx, exr)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = "update exchange_info set exchange_block_hash = ?, exchange_block_number = ?, order_status = 0  where exchange_tx_hash = ?"
	_, err = tx.Exec(query, ex.ExchangeBlockHash, ex.ExchangeBlockNumber, ex.ExchangeTxHash)
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

func (c *DBClient) ExchangeTrade(ex *utils.ExchangeInfo) error {

	exc, err := c.FindExchangeCollectByExId(ex.ExId)
	if err != nil {
		return err
	}

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	if exc == nil {
		return fmt.Errorf("exchange_collect not found")
	}

	amt0 := new(big.Int).Sub(exc.Amt0, exc.Amt0Finish)

	amt0Out := new(big.Int).Mul(ex.Amt1, exc.Amt0)
	amt0Out = new(big.Int).Div(amt0Out, exc.Amt1)

	if amt0.Cmp(amt0Out) < 0 {
		amt0Out = amt0
	}

	amt0Finish := new(big.Int).Add(exc.Amt0Finish, amt0Out)
	amt1Finish := new(big.Int).Add(exc.Amt1Finish, ex.Amt1)

	query := "update exchange_collect set amt0_finish = ?, amt1_finish = ? where ex_id = ?"
	_, err = tx.Exec(query, amt0Finish.String(), amt1Finish.String(), ex.ExId)
	if err != nil {
		tx.Rollback()
		return err
	}

	exr := &utils.ExchangeRevert{
		Op:   "trade",
		Tick: exc.Tick0,
		Exid: ex.ExId,
		Amt0: amt0Out,
		Amt1: ex.Amt1,
	}

	err = c.InstallExchangeRevert(tx, exr)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = "update exchange_info set tick0 = ?, tick1 = ?, amt0 = ?, exchange_block_hash = ?, exchange_block_number = ?, order_status = 0  where exchange_tx_hash = ?"
	_, err = tx.Exec(query, exc.Tick0, exc.Tick1, amt0Out.String(), ex.ExchangeBlockHash, ex.ExchangeBlockNumber, ex.ExchangeTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, exc.Tick1, ex.HolderAddress, exc.HolderAddress, ex.Amt1, false, ex.ExchangeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, exc.Tick0, exc.ReservesAddress, ex.HolderAddress, amt0Out, false, ex.ExchangeBlockNumber)
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

func (c *DBClient) ExchangeCancel(ex *utils.ExchangeInfo) error {

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	exc, err := c.FindExchangeCollectByExId(ex.ExId)
	if err != nil {
		return err
	}

	if exc == nil {
		return fmt.Errorf("exchange_collect not found")
	}

	err = c.Transfer(tx, exc.Tick0, exc.ReservesAddress, ex.HolderAddress, ex.Amt0, true, ex.ExchangeBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	amt0Finish := new(big.Int).Add(exc.Amt0Finish, ex.Amt0)

	query := "update exchange_collect set amt0_finish = ? where ex_id = ?"
	_, err = tx.Exec(query, amt0Finish.String(), ex.ExId)
	if err != nil {
		tx.Rollback()
		return err
	}

	exr := &utils.ExchangeRevert{
		Op:   "cancel",
		Tick: exc.Tick0,
		Exid: ex.ExId,
		Amt0: ex.Amt0,
	}

	err = c.InstallExchangeRevert(tx, exr)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = "update exchange_info set  tick0 = ?, tick1 = ?, exchange_block_hash = ?, exchange_block_number = ?, order_status = 0  where exchange_tx_hash = ?"
	_, err = tx.Exec(query, exc.Tick0, exc.Tick1, ex.ExchangeBlockHash, ex.ExchangeBlockNumber, ex.ExchangeTxHash)
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

func (c *DBClient) InstallExchangeAddressInfo(info *utils.AddressInfo) error {
	exec := "INSERT INTO exchange_address_info (order_id, address, public_key, private_key, receive_address, fee_address) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := c.SqlDB.Exec(exec, info.OrderId, info.Address, info.PubKey, info.PrveWif.String(), info.ReceiveAddress, info.FeeAddress)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *DBClient) InstallExchangeInfo(ex *utils.ExchangeInfo) error {
	exec := "INSERT INTO exchange_info (order_id, ex_id, op, tick0, tick1, amt0, amt1, fee_address, holder_address, exchange_tx_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?,?)"
	_, err := c.SqlDB.Exec(exec, ex.OrderId, ex.ExId, ex.Op, ex.Tick0, ex.Tick1, ex.Amt0.String(), ex.Amt1.String(), ex.FeeAddress, ex.HolderAddress, ex.ExchangeTxHash)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *DBClient) InstallExchangeRevert(tx *sql.Tx, ex *utils.ExchangeRevert) error {
	exec := "INSERT INTO exchange_revert (op, tick, exid, amt0, amt1, block_number) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, ex.Op, ex.Tick, ex.Exid, ex.Amt0.String(), ex.Amt1.String(), ex.BlockNumber)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelExchangeRevert(tx *sql.Tx, height int64) error {
	query := "delete from exchange_revert where block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}
func (c *DBClient) UpdateExchangeInfo(ex *utils.ExchangeInfo) error {
	query := "update exchange_info set fee_tx_hash = ?,  fee_tx_index = ?, fee_block_hash = ?, fee_block_number = ?, exchange_tx_hash = ?, exchange_tx_raw = ? where order_id = ?"
	_, err := c.SqlDB.Exec(query, ex.FeeTxHash, ex.FeeTxIndex, ex.FeeBlockHash, ex.FeeBlockNumber, ex.ExchangeTxHash, ex.ExchangeTxRaw, ex.OrderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateExchangeInfoErr(orderId, errInfo string) error {
	query := "update exchange_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.SqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateExchangeInfoFork(tx *sql.Tx, height int64) error {
	query := "update exchange_info set exchange_block_number = 0, exchange_block_hash = '', order_status = 0 where exchange_block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindExchangeRevertByNumber(height int64) ([]*utils.ExchangeRevert, error) {
	query := "SELECT  op, tick, exid, amt0, amt1, block_number  FROM exchange_revert where  block_number > ?  order by id desc "
	rows, err := c.SqlDB.Query(query, height)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	exs := []*utils.ExchangeRevert{}
	for rows.Next() {
		ex := &utils.ExchangeRevert{}
		var amt0, amt1 string
		err := rows.Scan(&ex.Op, &ex.Tick, &ex.Exid, &amt0, &amt1, &ex.BlockNumber)
		if err != nil {
			return nil, err
		}

		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		exs = append(exs, ex)
	}
	return exs, nil
}
func (c *DBClient) FindExchangeInfoByFee(feeAddress string) (*utils.ExchangeInfo, error) {
	query := "SELECT  order_id, ex_id, op, tick0, tick1, amt0, amt1, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, fee_address, holder_address,  update_date, create_date  FROM exchange_info where fee_address = ? and fee_tx_hash = '' order by create_date desc"
	rows, err := c.SqlDB.Query(query, feeAddress)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		ex := &utils.ExchangeInfo{}
		var amt0, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.ExId, &ex.Op, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &ex.FeeTxHash, &ex.FeeTxIndex, &ex.FeeBlockHash, &ex.FeeBlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, err
		}

		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		return ex, err
	}
	return nil, nil
}

func (c *DBClient) FindExchangeInfoByTxHash(txHash string) (*utils.ExchangeInfo, error) {
	query := "SELECT  order_id, op, tick0, tick1, amt0, amt1, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, exchange_tx_hash, exchange_tx_raw, exchange_block_hash, exchange_block_number, fee_address, holder_address, order_status, update_date, create_date   FROM exchange_info where exchange_tx_hash = ?"
	rows, err := c.SqlDB.Query(query, txHash)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		ex := &utils.ExchangeInfo{}
		var amt0, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &ex.FeeTxHash, &ex.FeeTxIndex, &ex.FeeBlockHash, &ex.FeeBlockNumber, &ex.ExchangeTxHash, &ex.ExchangeTxRaw, &ex.ExchangeBlockHash, &ex.ExchangeBlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, err
		}

		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)

		return ex, err
	}
	return nil, nil
}

func (c *DBClient) FindExchangeAddressInfo(orderId string) (*utils.AddressInfo, error) {
	query := "SELECT address, public_key, private_key, receive_address, order_id from exchange_address_info where order_id = ?"
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

func (c *DBClient) FindExchangeInfo(orderId, op, exId, tick, tick0, tick1, holder_address string, limit, offset int64) ([]*utils.ExchangeInfo, int64, error) {

	query := "SELECT  order_id, op, ex_id, tick0, tick1, amt0, amt1, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, exchange_tx_hash, exchange_block_hash, exchange_block_number, fee_address, holder_address, order_status,  update_date, create_date  FROM exchange_info  "

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

	if exId != "" {
		if where != "where" {
			where += " and "
		}
		where += "  ex_id = ? "
		whereAges = append(whereAges, exId)
	}

	if tick != "" {
		if where != "where" {
			where += " and "
		}
		where += "  ( tick0 = ? or tick1 = ? ) "
		whereAges = append(whereAges, tick)
		whereAges = append(whereAges, tick)
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

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*utils.ExchangeInfo{}
	for rows.Next() {
		ex := &utils.ExchangeInfo{}
		var amt0, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &ex.FeeTxHash, &ex.FeeTxIndex, &ex.FeeBlockHash, &ex.FeeBlockNumber, &ex.ExchangeTxHash, &ex.ExchangeBlockHash, &ex.ExchangeBlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		exs = append(exs, ex)
	}

	var total int64
	err = c.SqlDB.QueryRow("SELECT COUNT(order_id) FROM exchange_info "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *DBClient) FindExchangeInfoByTick(op, tick, holder_address string, limit, offset int64) ([]*utils.ExchangeInfo, int64, error) {

	query := "SELECT  order_id, op, ex_id, tick0, tick1, amt0, amt1, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, exchange_tx_hash, exchange_block_hash, exchange_block_number, fee_address, holder_address, order_status,  update_date, create_date FROM exchange_info where op = ? and holder_address = ? and ( tick0 = ? or tick1 = ?) "
	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ? "

	rows, err := c.SqlDB.Query(query+order+lim, op, holder_address, tick, tick, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*utils.ExchangeInfo{}
	for rows.Next() {
		ex := &utils.ExchangeInfo{}
		var amt0, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &ex.FeeTxHash, &ex.FeeTxIndex, &ex.FeeBlockHash, &ex.FeeBlockNumber, &ex.ExchangeTxHash, &ex.ExchangeBlockHash, &ex.ExchangeBlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		exs = append(exs, ex)
	}

	var total int64
	err = c.SqlDB.QueryRow("SELECT COUNT(order_id) FROM exchange_info where  op = ? and holder_address = ?  and ( tick0 = ? or tick1 = ?) ", op, holder_address, tick, tick).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *DBClient) FindExchangeCollect(exId, tick0, tick1, holderAddress string, notDone, limit, offset int64) ([]*utils.ExchangeCollect, int64, error) {
	query := "SELECT  ex_id, tick0, tick1, amt0, amt1, amt0_finish, amt1_finish, holder_address, reserves_address, update_date, create_date  FROM exchange_collect  "

	where := "where"
	whereAges := []any{}

	if exId != "" {
		whereAges = append(whereAges, exId)
		where += " ex_id = ? "
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

	if holderAddress != "" {
		if where != "where" {
			where += " and "
		}
		where += "  holder_address = ? "
		whereAges = append(whereAges, holderAddress)
	}

	if notDone == 1 {
		if where != "where" {
			where += " and "
		}
		where += " amt0 != amt0_finish"
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

	exs := []*utils.ExchangeCollect{}
	for rows.Next() {
		ex := &utils.ExchangeCollect{}
		var amt0, amt1, amt0_finish, amt1_finish string
		err := rows.Scan(&ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &amt0_finish, &amt1_finish, &ex.HolderAddress, &ex.ReservesAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		ex.Amt0Finish, _ = utils.ConvetStr(amt0_finish)
		ex.Amt1Finish, _ = utils.ConvetStr(amt1_finish)
		exs = append(exs, ex)
	}

	var total int64
	err = c.SqlDB.QueryRow("SELECT COUNT(ex_id) FROM exchange_collect "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *DBClient) FindExchangeCollectByExId(exId string) (*utils.ExchangeCollect, error) {

	query := "SELECT  ex_id, tick0, tick1, amt0, amt1, amt0_finish, amt1_finish, holder_address, reserves_address, update_date, create_date  FROM exchange_collect where ex_id = ?"
	rows, err := c.SqlDB.Query(query, exId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		ex := &utils.ExchangeCollect{}
		var amt0, amt1, amt0_finish, amt1_finish string
		err := rows.Scan(&ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &amt0_finish, &amt1_finish, &ex.HolderAddress, &ex.ReservesAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, err
		}
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		ex.Amt0Finish, _ = utils.ConvetStr(amt0_finish)
		ex.Amt1Finish, _ = utils.ConvetStr(amt1_finish)
		return ex, nil
	}

	return nil, nil

}
