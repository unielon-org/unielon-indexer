package storage_v3

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

func (e *MysqlClient) InstallDrc20Revert(tx *sql.Tx, tick, from, to string, amt *big.Int, height int64) error {
	exec := "INSERT INTO drc20_revert (tick, from_address, to_address, amt, block_number) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, tick, from, to, amt.String(), height)
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlClient) UpdateAddressBalanceMint(tx *sql.Tx, tick string, sum1, sum2 *big.Int, address string, sub bool) error {

	update1 := "UPDATE drc20_collect SET amt_sum=?, transactions = transactions + 1 WHERE tick = ?"
	if sub {
		update1 = "UPDATE drc20_collect SET amt_sum=?, transactions = transactions - 1 WHERE tick = ? and transactions > 0"
	}
	_, err := tx.Exec(update1, sum1.String(), tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	update2 := "INSERT INTO drc20_collect_address (tick, holder_address, amt_sum) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE amt_sum = ?"
	_, err = tx.Exec(update2, tick, address, sum2.String(), sum2.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (c *MysqlClient) UpdateAddressBalanceTran(tx *sql.Tx, tick string, sum1 *big.Int, address1 string, sum2 *big.Int, address2 string, sub bool) error {

	update1 := "UPDATE drc20_collect SET transactions = transactions + 1 WHERE tick = ?"
	if sub {
		update1 = "UPDATE drc20_collect SET transactions = transactions - 1 WHERE tick = ? and transactions > 0"
	}
	_, err := tx.Exec(update1, tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	update2 := "INSERT INTO drc20_collect_address (tick, holder_address, amt_sum) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE amt_sum = ?"
	_, err = tx.Exec(update2, tick, address1, sum1.String(), sum1.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	update3 := "INSERT INTO drc20_collect_address (tick, holder_address, amt_sum) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE amt_sum = ?"
	_, err = tx.Exec(update3, tick, address2, sum2.String(), sum2.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (c *MysqlClient) UpdateConvertAddress(addressA, addressD string) error {
	//query := "UPDATE address_info_new SET receive_address=? WHERE receive_address = ?"
	//_, err := c.MysqlDB.Exec(query, addressD, addressA)
	//if err != nil {
	//	return err
	//}
	//
	//query2 := "UPDATE cardinals_info_new SET receive_address=? WHERE receive_address = ?"
	//_, err = c.MysqlDB.Exec(query2, addressD, addressA)
	//if err != nil {
	//	return err
	//}
	//
	//query3 := "UPDATE drc20_address_info SET receive_address=? WHERE receive_address = ?"
	//_, err = c.MysqlDB.Exec(query3, addressD, addressA)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (c *MysqlClient) FindSwapDrc20InfoByTick(tx *sql.Tx, tick string) (*big.Int, *big.Int, *big.Int, error) {
	query := "SELECT amt_sum, max_, lim_ FROM drc20_collect WHERE tick = ?"
	rows, err := tx.Query(query, tick)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	if rows.Next() {

		var sum, max, lim string
		err := rows.Scan(&sum, &max, &lim)

		if err != nil {
			return nil, nil, nil, err
		}

		sum_big, _ := utils.ConvetStr(sum)
		max_big, _ := utils.ConvetStr(max)
		lim_big, _ := utils.ConvetStr(lim)
		return sum_big, max_big, lim_big, nil
	}

	return nil, nil, nil, errors.New("not found")
}

func (c *MysqlClient) FindDrc20InfoByTick(tick string) (*string, error) {
	query := "SELECT holder_address FROM drc20_collect WHERE tick = ?"
	rows, err := c.MysqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var receive_address *string
		err := rows.Scan(&receive_address)
		if err != nil {
			return nil, err
		}
		return receive_address, nil
	}

	return nil, errors.New("not found")
}

func (c *MysqlClient) FindDrc20AddressInfoByTick(tick string, address string) (*big.Int, error) {
	query := "SELECT amt_sum  FROM drc20_collect_address WHERE tick = ? and holder_address = ?"
	rows, err := c.MysqlDB.Query(query, tick, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {

		var sum string
		err := rows.Scan(&sum)

		if err != nil {
			return nil, err
		}

		is_ok := false
		sum_big := new(big.Int)
		if sum != "" {
			sum_big, is_ok = new(big.Int).SetString(sum, 10)
			if !is_ok {
				return nil, fmt.Errorf("max error")
			}
		}

		return sum_big, nil
	}

	return nil, ErrNotFound
}

func (c *MysqlClient) FindSwapDrc20AddressInfoByTick(tx *sql.Tx, tick string, address string) (*big.Int, error) {
	query := "SELECT amt_sum  FROM drc20_collect_address WHERE tick = ? and holder_address = ?"
	rows, err := tx.Query(query, tick, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {

		var sum string
		err := rows.Scan(&sum)

		if err != nil {
			return nil, err
		}

		is_ok := false
		sum_big := new(big.Int)
		if sum != "" {
			sum_big, is_ok = new(big.Int).SetString(sum, 10)
			if !is_ok {
				return nil, fmt.Errorf("max error")
			}
		}

		return sum_big, nil
	}

	return nil, ErrNotFound
}

func (c *MysqlClient) FindOrderByDrc20Hash(drc20Hash string) (*OrderResult, error) {
	query := "SELECT order_id, p, op, tick, amt, max_, lim_, repeat_mint,  tx_hash, block_hash, holder_address, create_date, to_address FROM drc20_info where tx_hash = ?"
	rows, err := c.MysqlDB.Query(query, drc20Hash)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		card := &OrderResult{}
		var max, amt, lim string

		err := rows.Scan(&card.OrderId, &card.P, &card.Op, &card.Tick, &amt, &max, &lim, &card.Repeat, &card.Drc20TxHash, &card.BlockHash, &card.ReceiveAddress, &card.CreateDate, &card.ToAddress)
		for err != nil {
			return nil, err
		}

		is_ok := false
		max_big := new(big.Int)
		if max != "" {
			max_big, is_ok = new(big.Int).SetString(max, 10)
			if !is_ok {
				return nil, fmt.Errorf("max error")
			}
		}
		card.Max = max_big

		amt_big := new(big.Int)
		if amt != "" {
			amt_big, is_ok = new(big.Int).SetString(amt, 10)
			if !is_ok {
				return nil, fmt.Errorf("amt error")
			}
		}
		card.Amt = amt_big

		lim_big := new(big.Int)
		if lim != "" {
			lim_big, is_ok = new(big.Int).SetString(lim, 10)
			if !is_ok {
				return nil, fmt.Errorf("lim error")
			}
		}
		card.Lim = lim_big
		return card, nil
	}
	return nil, nil
}

func (c *MysqlClient) FindDrc20() ([]*models.Drc20CollectAll, int64, error) {
	query := `
			SELECT di.tick                        AS ticker,
			       di.amt_sum,
			       di.max_,
			       di.lim_,
			       di.transactions,
			       UNIX_TIMESTAMP(di.create_date) AS DeployTime,
			       di.tx_hash
			FROM drc20_collect AS di
			GROUP BY di.tick`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*models.Drc20CollectAll
	for rows.Next() {

		result := &models.Drc20CollectAll{}
		var max, amt, lim string
		err := rows.Scan(&result.Tick, &amt, &max, &lim, &result.Transactions, &result.DeployTime, &result.Inscription)
		if err != nil {
			return nil, 0, err
		}

		is_ok := false
		max_big := new(big.Int)
		if max != "" {
			max_big, is_ok = new(big.Int).SetString(max, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("max error")
			}
		}
		result.MaxAmt = max_big

		amt_big := new(big.Int)
		if amt != "" {
			amt_big, is_ok = new(big.Int).SetString(amt, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("amt error")
			}
		}
		result.MintAmt = amt_big

		Lim, err := utils.ConvetStr(lim)
		if err != nil {
			return nil, 0, err
		}
		result.Lim = Lim
		results = append(results, result)
	}
	return results, 0, nil
}

func (c *MysqlClient) FindDrc20All() ([]*Drc20CollectAll, int64, error) {
	query := "SELECT di.tick AS ticker, di.amt_sum, di.max_, di.lim_, di.transactions, COUNT( ci.tick = di.tick ) AS Holders, di.create_date  AS DeployTime, di.tx_hash, di.logo, di.introduction, di.is_check FROM drc20_collect_address AS ci RIGHT JOIN drc20_collect AS di ON ci.tick = di.tick  GROUP BY di.tick ORDER BY DeployTime DESC "
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*Drc20CollectAll
	for rows.Next() {

		result := &Drc20CollectAll{}
		var max, amt, lim string
		err := rows.Scan(&result.Tick, &amt, &max, &lim, &result.Transactions, &result.Holders, &result.DeployTime, &result.Inscription, &result.Logo, &result.Introduction, &result.IsCheck)
		if err != nil {
			return nil, 0, err
		}

		is_ok := false
		max_big := new(big.Int)
		if max != "" {
			max_big, is_ok = new(big.Int).SetString(max, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("max error")
			}
		}
		result.MaxAmt = max_big

		amt_big := new(big.Int)
		if amt != "" {
			amt_big, is_ok = new(big.Int).SetString(amt, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("amt error")
			}
		}
		result.MintAmt = amt_big

		Lim, err := utils.ConvetStr(lim)
		if err != nil {
			return nil, 0, err
		}
		result.Lim = Lim

		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
			result.Introduction = &de
			result.WhitePaper = &de
			result.Official = &de
			result.Telegram = &de
			result.Discorad = &de
			result.Twitter = &de
			result.Facebook = &de
			result.Github = &de
		}

		results = append(results, result)
	}

	query1 := "SELECT COUNT(tick) AS UniqueTicks FROM drc20_info "

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

func (c *MysqlClient) FindDrc20TickAddress(address string) ([]string, error) {
	query := "SELECT tick FROM drc20_info where holder_address = ?"
	rows, err := c.MysqlDB.Query(query, address)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var results []string
	for rows.Next() {
		tick := ""
		err := rows.Scan(&tick)
		if err != nil {
			return nil, err
		}
		results = append(results, tick)
	}

	return results, nil
}

func (c *MysqlClient) FindDrc20ByTick(tick string) (*Drc20CollectAll, error) {
	query := "SELECT     di.tick AS ticker,     di.amt_sum,     di.max_ AS max_,     di.transactions AS Transactions,     di.update_date AS LastMintTime,     COUNT(CASE WHEN ci.tick = di.tick THEN 1 ELSE NULL END) AS Holders,     di.create_date AS DeployTime,     di.lim_ AS lim_,     di.dec_ AS dec_,     di.holder_address, di.tx_hash AS drc20_tx_hash_i0, di.logo, di.introduction, di.white_paper, di.official, di.telegram, di.discorad, di.twitter, di.facebook, di.github, di.is_check   FROM     drc20_collect_address AS ci     RIGHT JOIN drc20_collect AS di ON ci.tick = di.tick WHERE     di.tick = ? GROUP BY di.tick"
	rows, err := c.MysqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {

		result := &Drc20CollectAll{}
		var max, amt, lim *string
		err := rows.Scan(&result.Tick, &amt, &max, &result.Transactions, &result.LastMintTime, &result.Holders, &result.DeployTime, &lim, &result.Dec, &result.DeployBy, &result.Inscription, &result.Logo, &result.Introduction, &result.WhitePaper, &result.Official, &result.Telegram, &result.Discorad, &result.Twitter, &result.Facebook, &result.Github, &result.IsCheck)
		if err != nil {
			return nil, err
		}

		is_ok := false
		max_big := new(big.Int)
		if max != nil {
			max_big, is_ok = new(big.Int).SetString(*max, 10)
			if !is_ok {
				return nil, fmt.Errorf("max error")
			}
		}
		result.MaxAmt = max_big

		amt_big := new(big.Int)
		if amt != nil {
			amt_big, is_ok = new(big.Int).SetString(*amt, 10)
			if !is_ok {
				return nil, fmt.Errorf("amt error")
			}
		}
		result.MintAmt = amt_big

		lim_big := new(big.Int)
		if lim != nil {
			lim_big, is_ok = new(big.Int).SetString(*lim, 10)
			if !is_ok {
				return nil, fmt.Errorf("lim error")
			}
		}
		result.Lim = lim_big

		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
			result.Introduction = &de
			result.WhitePaper = &de
			result.Official = &de
			result.Telegram = &de
			result.Discorad = &de
			result.Twitter = &de
			result.Facebook = &de
			result.Github = &de
		}
		return result, nil
	}
	return nil, nil
}

func (c *MysqlClient) FindDrc20HoldersByTick(tick string, limit, offset int64) ([]*FindDrc20HoldersResult, int64, error) {
	query := "SELECT amt_sum, holder_address FROM drc20_collect_address WHERE tick = ? ORDER BY CAST(amt_sum AS DECIMAL(64, 0)) DESC LIMIT ? OFFSET ? ;"
	rows, err := c.MysqlDB.Query(query, tick, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*FindDrc20HoldersResult
	var amt string
	for rows.Next() {

		result := &FindDrc20HoldersResult{}
		err := rows.Scan(&amt, &result.Address)
		if err != nil {
			return nil, 0, err
		}

		Amt, err := utils.ConvetStr(amt)
		if err != nil {
			return nil, 0, err
		}
		result.Amt = Amt
		results = append(results, result)
	}

	query1 := "SELECT count(holder_address) FROM drc20_collect_address WHERE tick = ?"
	rows1, err := c.MysqlDB.Query(query1, tick)
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

func (c *MysqlClient) FindDrc20AllByAddress(receive_address string, limit, offset int64) ([]*FindDrc20AllByAddressResult, int64, error) {
	query := "SELECT tick, amt_sum FROM drc20_collect_address where holder_address = ? and amt_sum != '0' LIMIT ? OFFSET ?;"
	rows, err := c.MysqlDB.Query(query, receive_address, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*FindDrc20AllByAddressResult

	for rows.Next() {

		result := &FindDrc20AllByAddressResult{}
		var amt string

		err := rows.Scan(&result.Tick, &amt)
		if err != nil {
			return nil, 0, err
		}

		amt_big := new(big.Int)
		is_ok := false
		if amt != "" {
			amt_big, is_ok = new(big.Int).SetString(amt, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("lim error")
			}
		}
		result.Amt = amt_big

		results = append(results, result)
	}

	query1 := "SELECT count(tick) FROM drc20_collect_address where holder_address = ? and amt_sum != '0' "
	rows1, err := c.MysqlDB.Query(query1, receive_address)
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

func (c *MysqlClient) FindDrc20ByAddressPopular(receive_address string) ([]*FindDrc20AllByAddressResult, int64, error) {
	query := `SELECT
    t.tick,
    COALESCE(d.amt_sum, 0) AS amt_sum
FROM (
    SELECT 'UNIX' AS tick
    UNION SELECT 'CARDI'
    UNION SELECT 'DIS'
    UNION SELECT 'RONA'
    UNION SELECT 'CZZ'
    UNION SELECT 'ETHF'
	UNION SELECT 'WOW'
) t
LEFT JOIN drc20_collect_address d ON t.tick = d.tick AND d.holder_address = ?;`
	rows, err := c.MysqlDB.Query(query, receive_address)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*FindDrc20AllByAddressResult
	for rows.Next() {
		result := &FindDrc20AllByAddressResult{}
		var amt string

		err := rows.Scan(&result.Tick, &amt)
		if err != nil {
			return nil, 0, err
		}

		amt_big := new(big.Int)
		is_ok := false
		if amt != "" {
			amt_big, is_ok = new(big.Int).SetString(amt, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("lim error")
			}
		}
		result.Amt = amt_big
		results = append(results, result)
	}

	return results, 0, nil
}

func (c *MysqlClient) FindDrc20AllByAddressTick(receive_address, tick string) (*FindDrc20AllByAddressResult, error) {
	query := "SELECT tick, amt_sum FROM drc20_collect_address where holder_address = ? and amt_sum != '0' and tick = ?"
	rows, err := c.MysqlDB.Query(query, receive_address, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {

		result := &FindDrc20AllByAddressResult{}
		var amt string

		err := rows.Scan(&result.Tick, &amt)
		if err != nil {
			return nil, err
		}

		amt_big := new(big.Int)
		is_ok := false
		if amt != "" {
			amt_big, is_ok = new(big.Int).SetString(amt, 10)
			if !is_ok {
				return nil, fmt.Errorf("lim error")
			}
		}
		result.Amt = amt_big
		return result, nil
	}

	return nil, nil
}

func (c *MysqlClient) FindOrders(receiveAddress, op, tick string, limit, offset int64) ([]*models.Drc20Info, int64, error) {
	query := "SELECT order_id, p, op, tick, max_, lim_, amt, fee_address, holder_address, fee_tx_hash, tx_hash, block_number, block_hash, repeat_mint, create_date, order_status, to_address  FROM drc20_info  "

	where := "where"
	whereAges := []any{}

	if receiveAddress != "" {
		where += " receive_address = ? "
		whereAges = append(whereAges, receiveAddress)
	}

	if op != "" {
		if where != "where" {
			where += " and "
		}
		where += " op = ? "
		whereAges = append(whereAges, op)
	}

	if tick != "" {
		if where != "where" {
			where += " and "
		}
		where += " tick = ? "
		whereAges = append(whereAges, tick)
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
	var cards []*models.Drc20Info
	for rows.Next() {
		card := &models.Drc20Info{}
		var max *string
		var lim *string
		var amt *string

		err := rows.Scan(&card.OrderId, &card.P, &card.Op, &card.Tick, &max, &lim, &amt, &card.FeeAddress, &card.HolderAddress, &card.FeeTxHash, &card.TxHash, &card.BlockNumber, &card.BlockHash, &card.Repeat, &card.CreateDate, &card.OrderStatus, &card.ToAddress)
		if err != nil {
			return nil, 0, err
		}

		card.Max, err = utils.ConvetStringToNumber(*max)
		if err != nil {
			return nil, 0, err
		}

		card.Amt, err = utils.ConvetStringToNumber(*amt)
		if err != nil {
			return nil, 0, err
		}

		card.Lim, err = utils.ConvetStringToNumber(*lim)
		if err != nil {
			return nil, 0, err
		}

		cards = append(cards, card)
	}

	query1 := "SELECT count(order_id)  FROM drc20_info "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return cards, total, nil
}

func (c *MysqlClient) FindOrderByAddress(receiveAddress string, limit, offset int64) ([]*models.Drc20Info, int64, error) {
	query := "SELECT order_id, p, op, tick, max_, lim_, amt, fee_address, holder_address, fee_tx_hash,  tx_hash, block_hash, repeat_mint, create_date, order_status, to_address  FROM drc20_info where holder_address = ? or to_address = ?  order by update_date desc LIMIT ? OFFSET ?"

	rows, err := c.MysqlDB.Query(query, receiveAddress, receiveAddress, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var cards []*models.Drc20Info
	for rows.Next() {
		card := &models.Drc20Info{}
		var max *string
		var lim *string
		var amt *string

		err := rows.Scan(&card.OrderId, &card.P, &card.Op, &card.Tick, &max, &lim, &amt, &card.FeeAddress, &card.HolderAddress, &card.FeeTxHash, &card.TxHash, &card.BlockHash, &card.Repeat, &card.CreateDate, &card.OrderStatus, &card.ToAddress)
		if err != nil {
			return nil, 0, err
		}

		card.Max, err = utils.ConvetStringToNumber(*max)
		if err != nil {
			return nil, 0, err
		}

		card.Amt, err = utils.ConvetStringToNumber(*amt)
		if err != nil {
			return nil, 0, err
		}

		card.Lim, err = utils.ConvetStringToNumber(*lim)
		if err != nil {
			return nil, 0, err
		}

		cards = append(cards, card)
	}

	query1 := "SELECT count(order_id)  FROM drc20_info where holder_address = ? "

	rows1, err := c.MysqlDB.Query(query1, receiveAddress)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return cards, total, nil
}

func (c *MysqlClient) FindOrdersindex(receiveAddress, tick, hash string, number int64, limit, offset int64) ([]*OrderResult, int64, error) {
	query := `
SELECT ci.order_id,
       ci.p,
       ci.op,
       ci.tick,
       ci.max_,
       ci.lim_,
       ci.amt,
       ci.fee_address,
       ci.holder_address,
       ci.tx_hash,
       ci.fee_tx_hash,
       ci.block_hash,
       ci.block_number,
       ci.repeat_mint,
    ci.create_date,
       ci.order_status,
       ci.to_address,
       di.tx_hash
FROM drc20_info ci left join drc20_collect di on ci.tick = di.tick 
`
	where := "where"
	whereAges := []any{}

	if receiveAddress != "" {
		where += " (ci.holder_address = ? or ci.to_address = ?) "
		whereAges = append(whereAges, receiveAddress)
		whereAges = append(whereAges, receiveAddress)
	}

	if tick != "" {
		if where != "where" {
			where += " and "
		}
		where += "  ci.tick = ? "
		whereAges = append(whereAges, tick)
	}

	if hash != "" {
		if where != "where" {
			where += " and "
		}
		where += "  ci.tx_hash = ? "
		whereAges = append(whereAges, hash)
	}

	if number != 0 {
		if where != "where" {
			where += " and "
		}
		where += "  ci.block_number = ? "
		whereAges = append(whereAges, number)
	}

	order := " order by ci.update_date desc "
	lim := " LIMIT ? OFFSET ?"
	whereAgesLim := append(whereAges, limit)
	whereAgesLim = append(whereAgesLim, offset)

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAgesLim...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var cards []*OrderResult
	for rows.Next() {
		card := &OrderResult{}
		var max *string
		var lim *string
		var amt *string

		err := rows.Scan(&card.OrderId, &card.P, &card.Op, &card.Tick, &max, &lim, &amt, &card.FeeAddress, &card.ReceiveAddress, &card.Drc20TxHash, &card.FeeTxHash, &card.BlockHash, &card.BlockNumber, &card.Repeat, &card.CreateDate, &card.OrderStatus, &card.ToAddress, &card.Drc20Inscription)
		if err != nil {
			return nil, 0, err
		}

		card.Max, _ = utils.ConvetStr(*max)
		card.Amt, _ = utils.ConvetStr(*amt)
		card.Lim, _ = utils.ConvetStr(*lim)
		card.Inscription = card.Drc20TxHash + "i0"
		card.Drc20Inscription = card.Drc20Inscription + "i0"

		cards = append(cards, card)
	}

	query1 := "SELECT count(order_id)  FROM drc20_info ci "

	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return cards, total, nil
}

func (c *MysqlClient) FindOrderBytick(receiveAddress, tick string, limit, offset int64) ([]*OrderResult, int64, error) {
	query := "SELECT order_id, p, op, tick, max_, lim_, amt, fee_address,holder_address,  fee_tx_hash,  tx_hash, block_hash, repeat_mint, create_date, order_status, to_address  FROM drc20_info where holder_address = ? and tick = ? order by create_date desc LIMIT ? OFFSET ?;"
	rows, err := c.MysqlDB.Query(query, receiveAddress, tick, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var cards []*OrderResult
	for rows.Next() {
		card := &OrderResult{}
		var max *string
		var lim *string
		var amt *string

		err := rows.Scan(&card.OrderId, &card.P, &card.Op, &card.Tick, &max, &lim, &amt, &card.FeeAddress, &card.ReceiveAddress, &card.FeeTxHash, &card.Drc20TxHash, &card.BlockHash, &card.Repeat, &card.CreateDate, &card.OrderStatus, &card.ToAddress)
		if err != nil {
			return nil, 0, err
		}

		is_ok := false
		max_big := new(big.Int)
		if max != nil {
			max_big, is_ok = new(big.Int).SetString(*max, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("max error")
			}
		}
		card.Max = max_big

		amt_big := new(big.Int)
		if amt != nil {
			amt_big, is_ok = new(big.Int).SetString(*amt, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("amt error")
			}
		}
		card.Amt = amt_big

		lim_big := new(big.Int)
		if lim != nil {
			lim_big, is_ok = new(big.Int).SetString(*lim, 10)
			if !is_ok {
				return nil, 0, fmt.Errorf("lim error")
			}
		}
		card.Lim = lim_big

		cards = append(cards, card)
	}

	query1 := "SELECT count(order_id)  FROM drc20_info where holder_address = ? and tick = ? "

	rows1, err := c.MysqlDB.Query(query1, receiveAddress, tick)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return cards, total, nil
}

func (c *MysqlClient) FindOrderById(order_id string) (*OrderResult, error) {
	query := "SELECT order_id, p, op, tick, max_, lim_, amt, fee_address,holder_address,   fee_tx_hash,  tx_hash, block_hash, repeat_mint,  create_date, order_status, to_address  FROM drc20_info where order_id = ?"
	rows, err := c.MysqlDB.Query(query, order_id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		card := &OrderResult{}
		var max *string
		var lim *string
		var amt *string

		err := rows.Scan(&card.OrderId, &card.P, &card.Op, &card.Tick, &max, &lim, &amt, &card.FeeAddress, &card.ReceiveAddress, &card.FeeTxHash, &card.Drc20TxHash, &card.BlockHash, &card.Repeat, &card.CreateDate, &card.OrderStatus, &card.ToAddress)
		if err != nil {
			return nil, err
		}

		is_ok := false
		max_big := new(big.Int)
		if max != nil {
			max_big, is_ok = new(big.Int).SetString(*max, 10)
			if !is_ok {
				return nil, fmt.Errorf("max error")
			}
		}
		card.Max = max_big

		amt_big := new(big.Int)
		if amt != nil {
			amt_big, is_ok = new(big.Int).SetString(*amt, 10)
			if !is_ok {
				return nil, fmt.Errorf("amt error")
			}
		}
		card.Amt = amt_big

		lim_big := new(big.Int)
		if lim != nil {
			lim_big, is_ok = new(big.Int).SetString(*lim, 10)
			if !is_ok {
				return nil, fmt.Errorf("lim error")
			}
		}
		card.Lim = lim_big

		return card, nil
	}

	return nil, nil
}
