package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

func (c *DBClient) BoxDeploy(box *utils.BoxInfo, reservesAddress string) error {

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = c.InstallDrc20(tx, box.Max, big.NewInt(0), box.Tick0, reservesAddress, box.BoxTxHash)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("BoxDeploy Deploy err: %s order_id: %s", err.Error(), box.OrderId)
	}

	if err := c.Mint(tx, box.Tick0, reservesAddress, box.Max, false, box.BoxBlockNumber); err != nil {
		tx.Rollback()
		return fmt.Errorf("BoxDeploy Mint err: %s order_id: %s", err.Error(), box.OrderId)
	}

	boxCollect := &utils.BoxCollect{
		Tick0:           box.Tick0,
		Tick1:           box.Tick1,
		Max:             box.Max,
		Amt0:            box.Amt0,
		LiqAmt:          box.LiqAmt,
		LiqBlock:        box.LiqBlock,
		Amt1:            box.Amt1,
		HolderAddress:   box.HolderAddress,
		ReservesAddress: reservesAddress,
	}

	err = c.InstallBoxCollect(tx, boxCollect)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("BoxDeploy InstallBoxCollect err: %s order_id: %s", err.Error(), box.OrderId)
	}

	update := "update box_info set box_block_hash = ?, box_block_number = ?, order_status = 0  where box_tx_hash = ?"
	_, err = tx.Exec(update, box.BoxBlockHash, box.BoxBlockNumber, box.BoxTxHash)
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

func (c *DBClient) BoxMint(box *utils.BoxInfo, reservesAddress string) error {

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	boxc, err := c.FindBoxCollectByTick(tx, box.Tick0)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, boxc.Tick1, box.HolderAddress, reservesAddress, box.Amt1, false, box.BoxBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	update := "update box_collect a, drc20_address_info b set a.liqamt_finish = b.amt_sum where a.tick1 = b.tick and a.reserves_address = b.receive_address and a.reserves_address = ? and a.tick1 = ? and a.is_del = 0"
	_, err = tx.Exec(update, reservesAddress, boxc.Tick1)
	if err != nil {
		tx.Rollback()
		return err
	}

	ba := &utils.BoxAddress{
		Tick:          boxc.Tick0,
		HolderAddress: box.HolderAddress,
		Amt:           box.Amt1,
		BlockNumber:   box.BoxBlockNumber,
	}

	err = c.InstallBoxAddress(tx, ba)
	if err != nil {
		tx.Rollback()
		return err
	}

	boxc1, err := c.FindBoxCollectByTick(tx, box.Tick0)
	if err != nil {
		tx.Rollback()
		return err
	}

	if boxc1.LiqAmt.Cmp(big.NewInt(0)) > 0 && boxc1.LiqAmtFinish.Cmp(boxc1.LiqAmt) >= 0 {
		err := c.BoxFinish(tx, boxc1, box.BoxBlockNumber)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	update = "update box_info set box_block_hash = ?, box_block_number = ?, order_status = 0  where box_tx_hash = ?"
	_, err = tx.Exec(update, box.BoxBlockHash, box.BoxBlockNumber, box.BoxTxHash)
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

func (c *DBClient) BoxFinish(tx *sql.Tx, boxc *utils.BoxCollect, height int64) error {
	swap := &utils.SwapInfo{
		Tick0:           boxc.Tick0,
		Tick1:           boxc.Tick1,
		Op:              "create",
		Amt0:            boxc.Amt0,
		Amt1:            boxc.LiqAmtFinish,
		HolderAddress:   boxc.ReservesAddress,
		SwapBlockNumber: height,
	}

	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, _, _ = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, nil, nil)

	reservesAddressSwap, _ := btcutil.NewAddressScriptHash([]byte(swap.Tick0+swap.Tick1), &chaincfg.MainNetParams)
	swap.Tick = swap.Tick0 + "-SWAP-" + swap.Tick1
	liquidityBase := new(big.Int).Sqrt(new(big.Int).Mul(swap.Amt0, swap.Amt1))
	if liquidityBase.Cmp(big.NewInt(MINI_LIQUIDITY)) > 0 {
		liquidityBase = new(big.Int).Sub(liquidityBase, big.NewInt(MINI_LIQUIDITY))
	} else {
		return fmt.Errorf("add liquidity must be greater than MINI_LIQUIDITY firstly")
	}

	swap.Amt0Out = swap.Amt0
	swap.Amt1Out = swap.Amt1

	err := c.SwapCreate(tx, swap, reservesAddressSwap.String(), liquidityBase)

	if err != nil {
		return fmt.Errorf("swapCreate SwapCreate error: %v", err)
	}

	total := big.NewInt(0)
	bas, err := c.FindBoxAddressByTick(tx, boxc.Tick0)
	if err != nil {
		return err
	}

	for _, ba := range bas {
		total = total.Add(total, ba.Amt)
	}

	for _, ba := range bas {
		amt := big.NewInt(0).Div(big.NewInt(0).Mul(ba.Amt, boxc.Amt0), total)
		err = c.Transfer(tx, boxc.Tick0, boxc.ReservesAddress, ba.HolderAddress, amt, false, height)
		if err != nil {
			return err
		}
	}

	update := "update box_collect set amt0_finish = amt0 where tick0 = ? and is_del = 0"
	_, err = tx.Exec(update, boxc.Tick0)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) BoxRefund(tx *sql.Tx, boxc *utils.BoxCollect, height int64) error {

	err := c.Burn(tx, boxc.Tick0, boxc.ReservesAddress, boxc.Max, false, height)
	if err != nil {
		return err
	}

	err = c.DeleteDrc20(tx, boxc.Tick0)
	if err != nil {
		return err
	}

	update := "update box_collect set is_del = 1 where tick0 = ? "
	_, err = tx.Exec(update, boxc.Tick0)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) InstallBoxAddressInfo(info *utils.AddressInfo) error {
	exec := "INSERT INTO box_address_info (order_id, address, public_key, private_key, receive_address, fee_address) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := c.SqlDB.Exec(exec, info.OrderId, info.Address, info.PubKey, info.PrveWif.String(), info.ReceiveAddress, info.FeeAddress)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *DBClient) InstallBoxInfo(ex *utils.BoxInfo) error {
	exec := "INSERT INTO box_info (order_id, op, tick0, tick1, max_, amt0, liqamt, liqblock, amt1, fee_address, holder_address, box_tx_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := c.SqlDB.Exec(exec, ex.OrderId, ex.Op, ex.Tick0, ex.Tick1, ex.Max.String(), ex.Amt0.String(), ex.LiqAmt.String(), ex.LiqBlock, ex.Amt1.String(), ex.FeeAddress, ex.HolderAddress, ex.BoxTxHash)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) InstallBoxCollect(tx *sql.Tx, ex *utils.BoxCollect) error {
	exec := "INSERT INTO box_collect (tick0, tick1, max_, amt0, liqamt, liqblock, amt1, holder_address, reserves_address)  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, ex.Tick0, ex.Tick1, ex.Max.String(), ex.Amt0.String(), ex.LiqAmt.String(), ex.LiqBlock, ex.Amt1.String(), ex.HolderAddress, ex.ReservesAddress)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) InstallBoxAddress(tx *sql.Tx, ba *utils.BoxAddress) error {
	exec := "INSERT INTO box_address (tick, holder_address, amt, block_number) VALUES (?, ?, ?, ?)"
	_, err := tx.Exec(exec, ba.Tick, ba.HolderAddress, ba.Amt.String(), ba.BlockNumber)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelBoxAddressFork(tx *sql.Tx, height int64) error {
	exec := "delete from box_address where block_number > ?"
	_, err := tx.Exec(exec, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateBoxCollectFork(tx *sql.Tx, height int64) error {
	update := "update box_collect a, drc20_address_info b set a.liqamt_finish = b.amt_sum where a.tick1 = b.tick and a.reserves_address = b.receive_address"
	_, err := tx.Exec(update)
	if err != nil {
		tx.Rollback()
		return err
	}

	update = "update box_collect set is_del = 0 where liqblock > ? "
	_, err = tx.Exec(update, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateBoxInfoFork(tx *sql.Tx, height int64) error {
	query := "update box_info set box_block_number = 0, box_block_hash = '', order_status = 0 where box_block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateBoxInfoErr(orderId, errInfo string) error {
	query := "update box_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.SqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindBoxInfoByFee(feeAddress string) (*utils.BoxInfo, error) {
	query := "SELECT  order_id, op, tick0, tick1, max_, amt0, liqamt, liqblock, amt1, fee_address, holder_address,  update_date, create_date  FROM box_info where fee_address = ? and fee_tx_hash = '' order by create_date desc"
	rows, err := c.SqlDB.Query(query, feeAddress)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		ex := &utils.BoxInfo{}
		var max, amt0, liqamt, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &ex.FeeAddress, &ex.HolderAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, err
		}

		ex.Max, _ = utils.ConvetStr(max)
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.LiqAmt, _ = utils.ConvetStr(liqamt)
		ex.Amt1, _ = utils.ConvetStr(amt1)

		return ex, nil
	}
	return nil, nil
}

func (c *DBClient) FindBoxInfoByTxHash(txHash string) (*utils.BoxInfo, error) {
	query := "SELECT  order_id, op, tick0, tick1, max_, amt0, liqamt, liqblock, amt1, box_tx_hash, box_block_hash, box_block_number, fee_address, holder_address, order_status, update_date, create_date  FROM box_info where box_tx_hash = ?"
	rows, err := c.SqlDB.Query(query, txHash)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		ex := &utils.BoxInfo{}
		var max, amt0, liqamt, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &ex.BoxTxHash, &ex.BoxBlockHash, &ex.BoxBlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, err
		}

		ex.Max, _ = utils.ConvetStr(max)
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.LiqAmt, _ = utils.ConvetStr(liqamt)
		ex.Amt1, _ = utils.ConvetStr(amt1)

		return ex, err
	}
	return nil, nil
}

func (c *DBClient) FindBoxAddressInfo(orderId string) (*utils.AddressInfo, error) {
	query := "SELECT address, public_key, private_key, receive_address, order_id from box_address_info where order_id = ?"
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

func (c *DBClient) FindBoxInfo(orderId, op, tick0, tick1, holder_address string, limit, offset int64) ([]*utils.BoxInfo, int64, error) {

	query := "SELECT  order_id, op, tick0, tick1, max_, amt0, liqamt, liqblock, amt1, box_tx_hash, box_block_hash, box_block_number, fee_address, holder_address, order_status, update_date, create_date FROM box_info  "

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

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*utils.BoxInfo{}
	for rows.Next() {
		ex := &utils.BoxInfo{}
		var max, amt0, liqamt, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &ex.BoxTxHash, &ex.BoxBlockHash, &ex.BoxBlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Max, _ = utils.ConvetStr(max)
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.LiqAmt, _ = utils.ConvetStr(liqamt)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		exs = append(exs, ex)
	}

	var total int64
	err = c.SqlDB.QueryRow("SELECT COUNT(order_id) FROM box_info "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *DBClient) FindBoxCollect(tick0, tick1, holder_address string, limit, offset int64) ([]*utils.BoxCollect, int64, error) {
	query := "SELECT  tick0, tick1, max_, amt0, liqamt, liqblock, amt1, amt0_finish, liqamt_finish, holder_address, reserves_address, update_date, create_date  FROM box_collect  "
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

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*utils.BoxCollect{}
	for rows.Next() {
		ex := &utils.BoxCollect{}
		var max, liqamt, amt0, amt1, amt0_finish, liqamt_finish string
		err := rows.Scan(&ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &amt0_finish, &liqamt_finish, &ex.HolderAddress, &ex.ReservesAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Max, _ = utils.ConvetStr(max)
		ex.LiqAmt, _ = utils.ConvetStr(liqamt)
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		ex.Amt0Finish, _ = utils.ConvetStr(amt0_finish)
		ex.LiqAmtFinish, _ = utils.ConvetStr(liqamt_finish)
		exs = append(exs, ex)
	}

	var total int64
	err = c.SqlDB.QueryRow("SELECT COUNT(id) FROM box_collect "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *DBClient) FindBoxCollectByTick(tx *sql.Tx, tick string) (*utils.BoxCollect, error) {
	query := "SELECT tick0, tick1, max_, amt0, liqamt, liqblock, amt1, amt0_finish, liqamt_finish, holder_address, reserves_address FROM box_collect WHERE tick0 = ? and is_del = 0"
	rows, err := tx.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		ex := &utils.BoxCollect{}
		var max, liqamt, amt0, amt1, amt0_finish, liqamt_finish string
		err := rows.Scan(&ex.Tick0, &ex.Tick1, &max, &amt0, &liqamt, &ex.LiqBlock, &amt1, &amt0_finish, &liqamt_finish, &ex.HolderAddress, &ex.ReservesAddress)
		if err != nil {
			return nil, err
		}

		ex.Max, _ = utils.ConvetStr(max)
		ex.LiqAmt, _ = utils.ConvetStr(liqamt)
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		ex.Amt0Finish, _ = utils.ConvetStr(amt0_finish)
		ex.LiqAmtFinish, _ = utils.ConvetStr(liqamt_finish)
		return ex, nil
	}

	return nil, errors.New("not found")
}

func (c *DBClient) FindBoxCollectByExId(exId string) (*utils.BoxCollect, error) {

	query := "SELECT  tick0, tick1, amt0, amt1, amt0_finish, liqamt_finish, holder_address, reserves_address, update_date, create_date  FROM box_collect where ex_id = ? and is_del = 0"
	rows, err := c.SqlDB.Query(query, exId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ex := &utils.BoxCollect{}
	if rows.Next() {
		var amt0, amt1, amt0_finish, liqamt_finish string
		err := rows.Scan(&ex.Tick0, &ex.Tick1, &amt0, &amt1, &amt0_finish, &liqamt_finish, &ex.HolderAddress)
		if err != nil {
			return nil, err
		}
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		ex.Amt0Finish, _ = utils.ConvetStr(amt0_finish)
		ex.LiqAmtFinish, _ = utils.ConvetStr(liqamt_finish)
	}

	return ex, nil

}

func (c *DBClient) FindBoxCollectByHeight(height int64) ([]*utils.BoxCollect, error) {

	query := "SELECT  tick0, tick1, amt0, amt1, max_, amt0_finish, liqamt_finish, holder_address, reserves_address, update_date, create_date FROM box_collect where liqblock = ? and is_del = 0"
	rows, err := c.SqlDB.Query(query, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exs := []*utils.BoxCollect{}
	for rows.Next() {
		ex := &utils.BoxCollect{}
		var amt0, amt1, max, amt0_finish, liqamt_finish string
		err := rows.Scan(&ex.Tick0, &ex.Tick1, &amt0, &amt1, &max, &amt0_finish, &liqamt_finish, &ex.HolderAddress, &ex.ReservesAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, err
		}
		ex.Amt0, _ = utils.ConvetStr(amt0)
		ex.Amt1, _ = utils.ConvetStr(amt1)
		ex.Max, _ = utils.ConvetStr(max)
		ex.Amt0Finish, _ = utils.ConvetStr(amt0_finish)
		ex.LiqAmtFinish, _ = utils.ConvetStr(liqamt_finish)
		exs = append(exs, ex)
	}

	return exs, nil
}

func (c *DBClient) FindBoxAddressByTick(tx *sql.Tx, tick string) ([]*utils.BoxAddress, error) {

	query := "SELECT  tick, holder_address, amt, block_number, create_date  FROM box_address where tick = ?"
	rows, err := tx.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exs := []*utils.BoxAddress{}
	for rows.Next() {
		ex := &utils.BoxAddress{}
		var amt string
		err := rows.Scan(&ex.Tick, &ex.HolderAddress, &amt, &ex.BlockNumber, &ex.CreateDate)
		if err != nil {
			return nil, err
		}
		ex.Amt, _ = utils.ConvetStr(amt)
		exs = append(exs, ex)
	}

	return exs, nil
}
