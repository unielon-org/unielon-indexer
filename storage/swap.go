package storage

import (
	"database/sql"
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"strings"
)

func (c *DBClient) SwapCreate(swap *utils.SwapInfo, reservesAddress string, liquidityTotal *big.Int) error {
	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	query := "INSERT INTO swap_liquidity (tick, tick0, tick1, holder_address, reserves_address, liquidity_total) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(query, swap.Tick, swap.Tick0, swap.Tick1, swap.HolderAddress, reservesAddress, liquidityTotal.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick0, swap.HolderAddress, reservesAddress, swap.Amt0, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick1, swap.HolderAddress, reservesAddress, swap.Amt1, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	execQ := "INSERT INTO drc20_info (tick, `max_`, lim_, receive_address, drc20_tx_hash) VALUES (?, ?, ?, ?, ?)"
	_, err = tx.Exec(execQ, swap.Tick, "99999999999999999999999999999999999999999", "0", reservesAddress, swap.SwapTxHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Mint(tx, swap.Tick, swap.HolderAddress, liquidityTotal, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.UpdateLiquidity(tx, swap.Tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = "update swap_info set amt0_out = ?, amt1_out = ?, liquidity = ?, swap_block_hash = ?, swap_block_number = ?, order_status = 0  where swap_tx_hash = ?"
	_, err = tx.Exec(query, swap.Amt0Out.String(), swap.Amt1Out.String(), liquidityTotal.String(), swap.SwapBlockHash, swap.SwapBlockNumber, swap.SwapTxHash)
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

func (c *DBClient) SwapAdd(swap *utils.SwapInfo, reservesAddress string, amt0, amt1, liquidity *big.Int) error {
	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = c.Transfer(tx, swap.Tick0, swap.HolderAddress, reservesAddress, amt0, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick1, swap.HolderAddress, reservesAddress, amt1, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Mint(tx, swap.Tick, swap.HolderAddress, liquidity, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.UpdateLiquidity(tx, swap.Tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	query := "update swap_info set amt0_out = ?, amt1_out = ?, liquidity = ?, swap_block_hash = ?, swap_block_number = ?, order_status = 0  where swap_tx_hash = ?"
	_, err = tx.Exec(query, amt0.String(), amt1.String(), liquidity.String(), swap.SwapBlockHash, swap.SwapBlockNumber, swap.SwapTxHash)
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

// Remove
func (c *DBClient) SwapRemove(swap *utils.SwapInfo, reservesAddress string, amt0Out, amt1Out *big.Int) error {
	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = c.Transfer(tx, swap.Tick0, reservesAddress, swap.HolderAddress, amt0Out, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick1, reservesAddress, swap.HolderAddress, amt1Out, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Burn(tx, swap.Tick, swap.HolderAddress, swap.Liquidity, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.UpdateLiquidity(tx, swap.Tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	query := "update swap_info set amt0_out = ?, amt1_out = ?, swap_block_hash = ?, swap_block_number = ?, order_status = 0  where swap_tx_hash = ?"
	_, err = tx.Exec(query, amt0Out.String(), amt1Out.String(), swap.SwapBlockHash, swap.SwapBlockNumber, swap.SwapTxHash)
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

// SwapNow
func (c *DBClient) SwapNow(swap *utils.SwapInfo, reservesAddress string, amtin, amtout, amtoutFeeCommunity, amtoutFeeCommunityOut *big.Int) error {

	tx, err := c.SqlDB.Begin()
	if err != nil {
		return err
	}

	err = c.Transfer(tx, swap.Tick0, swap.HolderAddress, reservesAddress, amtin, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick1, reservesAddress, swap.HolderAddress, amtout, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick0, reservesAddress, "DMmdAkMPXb9H1JfRYmkvpyby5EFgkVKmmQ", amtoutFeeCommunity, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.Transfer(tx, swap.Tick1, reservesAddress, "DMmdAkMPXb9H1JfRYmkvpyby5EFgkVKmmQ", amtoutFeeCommunityOut, false, swap.SwapBlockNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = c.UpdateLiquidity(tx, swap.Tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	exec := "update swap_info set amt1_out= ?, swap_block_hash = ?, swap_block_number = ?, order_status = 0  where swap_tx_hash = ?"
	_, err = tx.Exec(exec, amtout.String(), swap.SwapBlockHash, swap.SwapBlockNumber, swap.SwapTxHash)
	if err != nil {
		log.Error("swapCreate", "tx.Exec", err, "Tick0", swap.Tick0, "Tick1", swap.Tick1)
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

func (c *DBClient) UpdateLiquidity(tx *sql.Tx, tick string) error {
	exec := `UPDATE swap_liquidity
				SET amt0 = (
					SELECT b.amt_sum
					FROM drc20_address_info b
					WHERE 
						swap_liquidity.tick0 = b.tick AND 
						swap_liquidity.reserves_address = b.receive_address
				)
				WHERE EXISTS (
					SELECT 1
					FROM drc20_address_info b
					WHERE 
						swap_liquidity.tick0 = b.tick AND 
						swap_liquidity.reserves_address = b.receive_address AND 
						swap_liquidity.tick = ?
				);`
	_, err := tx.Exec(exec, tick)
	if err != nil {
		return err
	}

	exec = `UPDATE swap_liquidity
				SET amt1 = (
					SELECT b.amt_sum
					FROM drc20_address_info b
					WHERE 
						swap_liquidity.tick1 = b.tick AND 
						swap_liquidity.reserves_address = b.receive_address
				)
				WHERE EXISTS (
					SELECT 1
					FROM drc20_address_info b
					WHERE 
						swap_liquidity.tick1 = b.tick AND 
						swap_liquidity.reserves_address = b.receive_address AND 
						swap_liquidity.tick = ?
				);`
	_, err = tx.Exec(exec, tick)
	if err != nil {
		return err
	}

	exec = `UPDATE swap_liquidity
				SET liquidity_total = (
					SELECT b.amt_sum
					FROM drc20_info b
					WHERE swap_liquidity.tick = b.tick
				)
				WHERE tick = ?`
	_, err = tx.Exec(exec, tick)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) InstallSwapInfo(swap *utils.SwapInfo) error {
	query := "INSERT INTO swap_info (order_id, op, tick0, tick1, amt0, amt1, amt0_min, amt1_min, liquidity, fee_address, holder_address, swap_tx_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?)"
	_, err := c.SqlDB.Exec(query, swap.OrderId, swap.Op, swap.Tick0, swap.Tick1, swap.Amt0.String(), swap.Amt1.String(), swap.Amt0Min.String(), swap.Amt1Min.String(), swap.Liquidity.String(), swap.FeeAddress, swap.HolderAddress, swap.SwapTxHash)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *DBClient) UpdateSwapInfoErr(orderId, errInfo string) error {
	query := "update swap_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.SqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateSwapInfoFork(tx *sql.Tx, height int64) error {
	query := "update swap_info set swap_block_number = 0, swap_block_hash = '', order_status = 0 where swap_block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindSwapInfoBySwapTxHash(swapTxHash string) (*utils.SwapInfo, error) {
	query := "SELECT  order_id, op, tick0, tick1, amt0, amt1, liquidity, swap_tx_hash, swap_block_hash, swap_block_number, fee_address, holder_address, order_status, update_date, create_date  FROM swap_info where swap_tx_hash = ?"
	rows, err := c.SqlDB.Query(query, swapTxHash)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		swap := &utils.SwapInfo{}
		var amt0, amt1, liquidity string
		err := rows.Scan(&swap.OrderId, &swap.Op, &swap.Tick0, &swap.Tick1, &amt0, &amt1, &liquidity, &swap.SwapTxHash, &swap.SwapBlockHash, &swap.SwapBlockNumber, &swap.FeeAddress, &swap.HolderAddress, &swap.OrderStatus, &swap.UpdateDate, &swap.CreateDate)
		if err != nil {
			return nil, err
		}

		swap.Amt0, _ = utils.ConvetStr(amt0)
		swap.Amt1, _ = utils.ConvetStr(amt1)
		swap.Liquidity, _ = utils.ConvetStr(liquidity)

		return swap, err
	}
	return nil, nil
}

func (c *DBClient) FindSwapInfo(op, tick0, tick1, holder_address string, limit, offset int64) ([]*utils.SwapInfo, int64, error) {
	query := "SELECT  order_id, op, tick0, tick1, amt0, amt1, swap_tx_hash, swap_block_hash, swap_block_number, fee_address, holder_address, order_status,  update_date, create_date FROM swap_info  "

	where := "where"
	whereAges := []any{}
	if op != "" {
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

	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ?"
	whereAges = append(whereAges, limit)
	whereAges = append(whereAges, offset)

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	swaps := make([]*utils.SwapInfo, 0)
	for rows.Next() {
		swap := &utils.SwapInfo{}
		var amt0, amt1 string
		err := rows.Scan(&swap.OrderId, &swap.Op, &swap.Tick0, &swap.Tick1, &amt0, &amt1, &swap.SwapTxHash, &swap.SwapBlockHash, &swap.SwapBlockNumber, &swap.FeeAddress, &swap.HolderAddress, &swap.OrderStatus, &swap.UpdateDate, &swap.CreateDate)
		if err != nil {
			return nil, 0, err
		}

		swap.Amt0, _ = utils.ConvetStr(amt0)
		swap.Amt1, _ = utils.ConvetStr(amt1)
		swaps = append(swaps, swap)
	}

	query1 := "SELECT count(order_id)  FROM swap_info "
	rows1, err := c.SqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return swaps, total, nil
}

func (c *DBClient) FindSwapLiquidityAll() ([]*utils.SwapLiquidity, int64, error) {
	query := "SELECT tick, tick0, tick1, amt0, amt1, liquidity_total from swap_liquidity where liquidity_total != '0'"
	rows, err := c.SqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*utils.SwapLiquidity, 0)
	for rows.Next() {
		liquidity := &utils.SwapLiquidity{}
		var amt0, amt1, liquidity_total string
		err := rows.Scan(&liquidity.Tick, &liquidity.Tick0, &liquidity.Tick1, &amt0, &amt1, &liquidity_total)
		if err != nil {
			return nil, 0, err
		}
		liquidity.Amt0, _ = utils.ConvetStr(amt0)
		liquidity.Amt1, _ = utils.ConvetStr(amt1)
		liquidity.LiquidityTotal, _ = utils.ConvetStr(liquidity_total)
		list = append(list, liquidity)
	}

	query1 := "SELECT count(tick)  FROM swap_liquidity  where liquidity_total != '0'"
	rows1, err := c.SqlDB.Query(query1)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return list, total, nil
}

func (c *DBClient) FindSwapLiquidity(tick0 string, tick1 string) (*utils.SwapLiquidity, error) {
	query := "SELECT tick, tick0, tick1, amt0, amt1, holder_address, liquidity_total, reserves_address from swap_liquidity where tick0 = ? and tick1 = ? "
	tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)
	rows, err := c.SqlDB.Query(query, tick0, tick1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquidity := &utils.SwapLiquidity{}
	if rows.Next() {
		var amt0, amt1, liquidity_total string
		err := rows.Scan(&liquidity.Tick, &liquidity.Tick0, &liquidity.Tick1, &amt0, &amt1, &liquidity.HolderAddress, &liquidity_total, &liquidity.ReservesAddress)
		if err != nil {
			return nil, err
		}
		liquidity.Amt0, _ = utils.ConvetStr(amt0)
		liquidity.Amt1, _ = utils.ConvetStr(amt1)
		liquidity.LiquidityTotal, _ = utils.ConvetStr(liquidity_total)
		return liquidity, nil
	}
	return nil, nil
}

func (c *DBClient) FindSwapLiquidityWeb(tick0 string, tick1 string) (*utils.SwapLiquidity, error) {
	query := "SELECT tick, tick0, tick1, amt0, amt1, holder_address, liquidity_total, reserves_address from swap_liquidity where tick0 = ? and tick1 = ? and liquidity_total != '0'"
	tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)
	rows, err := c.SqlDB.Query(query, tick0, tick1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquidity := &utils.SwapLiquidity{}
	if rows.Next() {
		var amt0, amt1, liquidity_total string
		err := rows.Scan(&liquidity.Tick, &liquidity.Tick0, &liquidity.Tick1, &amt0, &amt1, &liquidity.HolderAddress, &liquidity_total, &liquidity.ReservesAddress)
		if err != nil {
			return nil, err
		}
		liquidity.Amt0, _ = utils.ConvetStr(amt0)
		liquidity.Amt1, _ = utils.ConvetStr(amt1)
		liquidity.LiquidityTotal, _ = utils.ConvetStr(liquidity_total)
		return liquidity, nil
	}
	return nil, nil
}

func (c *DBClient) FindSwapLiquidityByHolder(holder_address string, tick0, tick1 string) ([]*utils.SwapLiquidity, error) {

	query := "SELECT tick, amt_sum from drc20_address_info where receive_address = ? and LENGTH(tick) >= 10 and tick != 'WDOGE(WRAPPED-DOGE)' and amt_sum != '0'"
	queryf := []any{holder_address}

	if tick0 != "" {
		query = "SELECT tick, amt_sum from drc20_address_info where receive_address = ? and LENGTH(tick) >= 10 and tick = ? and tick != 'WDOGE(WRAPPED-DOGE)' and amt_sum != '0'"
		queryf = []any{holder_address, tick0 + "-SWAP-" + tick1}
	}

	rows, err := c.SqlDB.Query(query, queryf...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquiditys := make([]*utils.SwapLiquidity, 0)

	for rows.Next() {
		liquidity := &utils.SwapLiquidity{}
		var tick, amt_sum string
		err := rows.Scan(&tick, &amt_sum)
		if err != nil {
			return nil, err
		}

		var tick0, tick1 string

		tick0, tick1 = strings.Split(tick, "-SWAP-")[0], strings.Split(tick, "-SWAP-")[1]

		swapLiquidity, err := c.FindSwapLiquidity(tick0, tick1)
		if err != nil {
			return nil, err
		}

		amt_sum_big, _ := utils.ConvetStr(amt_sum)
		liquidity.Amt0 = new(big.Int).Div(new(big.Int).Mul(swapLiquidity.Amt0, amt_sum_big), swapLiquidity.LiquidityTotal)
		liquidity.Amt1 = new(big.Int).Div(new(big.Int).Mul(swapLiquidity.Amt1, amt_sum_big), swapLiquidity.LiquidityTotal)
		liquidity.Tick0 = tick0
		liquidity.Tick1 = tick1
		liquidity.LiquidityTotal = amt_sum_big
		liquiditys = append(liquiditys, liquidity)
	}

	return liquiditys, nil
}

func (c *DBClient) FindSwapPriceAll() ([]*SwapPrice, int64, error) {

	liquidityAll, _, err := c.FindSwapLiquidityAll()
	if err != nil {
		return nil, 0, err
	}

	liquidityMap := make(map[string]*big.Float)
	for _, v := range liquidityAll {
		if v.Tick0 != "WDOGE(WRAPPED-DOGE)" && v.Tick1 != "WDOGE(WRAPPED-DOGE)" {
			continue
		}

		if v.Tick0 == "WDOGE(WRAPPED-DOGE)" {
			liquidityMap[v.Tick1] = new(big.Float).Quo(new(big.Float).SetInt(v.Amt0), new(big.Float).SetInt(v.Amt1))
		} else {
			liquidityMap[v.Tick0] = new(big.Float).Quo(new(big.Float).SetInt(v.Amt1), new(big.Float).SetInt(v.Amt0))
		}
	}

	liquiditys := make([]*SwapPrice, 0)
	len := int64(0)
	for k, v := range liquidityMap {
		f, _ := v.Float64()
		liquiditys = append(liquiditys, &SwapPrice{
			Tick:      k,
			LastPrice: f,
		})
		len++
	}
	return liquiditys, len, nil
}
