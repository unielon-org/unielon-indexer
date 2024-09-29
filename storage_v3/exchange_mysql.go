package storage_v3

import (
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"strconv"
)

func (c *MysqlClient) FindExchangeInfo(orderId, op, exId, tick, tick0, tick1, holder_address string, limit, offset int64) ([]*models.ExchangeInfo, int64, error) {

	query := "SELECT  order_id, op, ex_id, tick0, tick1, amt0, amt1, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, order_status, update_date, create_date   FROM exchange_info  "

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

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*models.ExchangeInfo{}
	for rows.Next() {
		ex := &models.ExchangeInfo{}
		var amt0, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &ex.FeeTxHash, &ex.TxHash, &ex.BlockHash, &ex.BlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Amt0, _ = utils.ConvetStringToNumber(amt0)
		ex.Amt1, _ = utils.ConvetStringToNumber(amt1)
		exs = append(exs, ex)
	}

	var total int64
	err = c.MysqlDB.QueryRow("SELECT COUNT(order_id) FROM exchange_info "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *MysqlClient) FindExchangeInfoByTick(op, tick, holder_address string, limit, offset int64) ([]*models.ExchangeInfo, int64, error) {

	query := "SELECT  order_id, op, ex_id, tick0, tick1, amt0, amt1, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, order_status, update_date, create_date FROM exchange_info where op = ? and holder_address = ? and ( tick0 = ? or tick1 = ?) "
	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ? "

	rows, err := c.MysqlDB.Query(query+order+lim, op, holder_address, tick, tick, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*models.ExchangeInfo{}
	for rows.Next() {
		ex := &models.ExchangeInfo{}
		var amt0, amt1 string
		err := rows.Scan(&ex.OrderId, &ex.Op, &ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &ex.FeeTxHash, &ex.TxHash, &ex.BlockHash, &ex.BlockNumber, &ex.FeeAddress, &ex.HolderAddress, &ex.OrderStatus, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Amt0, _ = utils.ConvetStringToNumber(amt0)
		ex.Amt1, _ = utils.ConvetStringToNumber(amt1)
		exs = append(exs, ex)
	}

	var total int64
	err = c.MysqlDB.QueryRow("SELECT COUNT(order_id) FROM exchange_info where  op = ? and holder_address = ?  and ( tick0 = ? or tick1 = ?) ", op, holder_address, tick, tick).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *MysqlClient) FindExchangeCollect(exId, tick0, tick1, holderAddress string, notDone, limit, offset int64) ([]*models.ExchangeCollect, int64, error) {
	query := "SELECT  ex_id, tick0, tick1, amt0, amt1, amt0_finish, amt1_finish, holder_address, reserves_address,update_date, create_date   FROM exchange_collect  "

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

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exs := []*models.ExchangeCollect{}
	for rows.Next() {
		ex := &models.ExchangeCollect{}
		var amt0, amt1, amt0_finish, amt1_finish string
		err := rows.Scan(&ex.ExId, &ex.Tick0, &ex.Tick1, &amt0, &amt1, &amt0_finish, &amt1_finish, &ex.HolderAddress, &ex.ReservesAddress, &ex.UpdateDate, &ex.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ex.Amt0, _ = utils.ConvetStringToNumber(amt0)
		ex.Amt1, _ = utils.ConvetStringToNumber(amt1)
		ex.Amt0Finish, _ = utils.ConvetStringToNumber(amt0_finish)
		ex.Amt1Finish, _ = utils.ConvetStringToNumber(amt1_finish)
		exs = append(exs, ex)
	}

	var total int64
	err = c.MysqlDB.QueryRow("SELECT COUNT(ex_id) FROM exchange_collect "+where, whereAges...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return exs, total, nil
}

func (c *MysqlClient) FindExchangeSummary() (*FindSummaryResult, error) {
	query := `SELECT
				    COUNT(id) AS total_records,  
				    CAST(
				        COALESCE(SUM(
				            CASE
				                WHEN tick0 = 'WDOGE(WRAPPED-DOGE)' THEN amt0_finish
				                WHEN tick1 = 'WDOGE(WRAPPED-DOGE)' THEN amt1_finish
				                ELSE 0
				            END
				        ), 0) AS DECIMAL(32,0)
				    ) AS total_doge_amt_last
				FROM
				    exchange_collect;`

	rows, err := c.MysqlDB.Query(query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	result := &FindSummaryResult{}
	if rows.Next() {
		err := rows.Scan(&result.Exchange, &result.Value24h)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (c *MysqlClient) FindExchangeSummaryAll() ([]*FindSummaryAllResult, error) {
	query := `
		SELECT
    d20i.tick,
    COALESCE(e.close_price * d20i.amt_sum, 0) AS total_doge_amt,
    COALESCE(e.close_price,0) AS closePrice,
    COALESCE(e.quote_volume, 0) AS totalQuoteVolume, 
   COALESCE(((e.close_price - e.open_price) / e.open_price) * 100, 0) AS priceChange,
    (SELECT COUNT(holder_address) FROM drc20_collect_address WHERE drc20_collect_address.tick = e.tick0) AS receive_address_count,
    COALESCE(e.lowest_ask, 0) AS footPrice,
    d20i.logo,
    d20i.is_check
FROM
    drc20_collect d20i
LEFT JOIN (
    SELECT
        es.tick0,
        es.close_price,
        es.lowest_ask,
        es.quote_volume,
        es.open_price,
        es.id
    FROM
        exchange_summary es
    INNER JOIN (
        SELECT
            tick0,
            MAX(id) AS max_id
        FROM
            exchange_summary
        WHERE
            date_interval = '1d'
            AND (tick0 = 'WDOGE(WRAPPED-DOGE)' OR tick1 = 'WDOGE(WRAPPED-DOGE)')
        GROUP BY
            tick0
    ) es_max ON es.id = es_max.max_id
) e ON e.tick0 = d20i.tick
WHERE
    LENGTH(d20i.tick) < 9
ORDER BY
    totalQuoteVolume DESC, receive_address_count DESC
`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*FindSummaryAllResult, 0)
	for rows.Next() {
		var marketcap *string
		result := &FindSummaryAllResult{}
		err := rows.Scan(&result.Tick, &marketcap, &result.LastPrice, &result.QuoteVolume, &result.PriceChangePercent24H, &result.Holders, &result.FootPrice, &result.Logo, &result.IsCheck)
		if err != nil {
			return nil, err
		}

		f, err := strconv.ParseFloat(*marketcap, 64)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return nil, err
		}

		result.TotalMaxAmt = strconv.FormatInt(int64(f), 10)
		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
		}

		results = append(results, result)
	}
	return results, nil
}

func (c *MysqlClient) FindExchangeSummaryByTick(tick string) (*FindSummaryAllResult, error) {
	query := `
			SELECT
      			d20i.tick,
			    COALESCE(e.close_price * d20i.amt_sum, 0) AS total_doge_amt,
			    COALESCE(e.close_price,0) AS closePrice,
			    COALESCE(e.highest_bid, 0) AS highestBid,
			    COALESCE(e.quote_volume, 0) AS totalQuoteVolume, 
			    COALESCE(((e.close_price - e.open_price) / e.open_price) * 100, 0) AS priceChange,
			    (SELECT COUNT(holder_address) FROM drc20_collect_address WHERE drc20_collect_address.tick = e.tick0) AS receive_address_count,
			    COALESCE(e.lowest_ask, 0) AS footPrice,
			    d20i.max_,
			    d20i.logo,
			    d20i.is_check
			FROM
			    drc20_collect d20i
			LEFT JOIN (
			    SELECT
			        e.tick0,
			        e.close_price,
			        e.highest_bid,
			        e.lowest_ask,
			        e.quote_volume,
			        e.open_price,
			        e.id
			    FROM
			        exchange_summary e
			    INNER JOIN (
			        SELECT
			            tick0,
			            MAX(id) AS max_id
			        FROM
			            exchange_summary
			        WHERE
			            date_interval = '1d'
			            AND (tick0 = 'WDOGE(WRAPPED-DOGE)' OR tick1 = 'WDOGE(WRAPPED-DOGE)')
			        GROUP BY
			            tick0
			    ) es_max ON e.id = es_max.max_id
			) e ON e.tick0 = d20i.tick
			WHERE
			    d20i.tick = ? `

	rows, err := c.MysqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var marketcap *string

		result := &FindSummaryAllResult{}
		err := rows.Scan(&result.Tick, &marketcap, &result.LastPrice, &result.HighestPrice24H, &result.QuoteVolume, &result.PriceChangePercent24H, &result.Holders, &result.FootPrice, &result.MaxAmt, &result.Logo, &result.IsCheck)
		if err != nil {
			return nil, err
		}

		result.LowestPrice24H = result.FootPrice

		f, err := strconv.ParseFloat(*marketcap, 64)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return nil, err
		}

		result.TotalMaxAmt = strconv.FormatInt(int64(f), 10)
		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
		}

		return result, nil

	}
	return nil, nil
}

func (c *MysqlClient) FindExchangeSummaryK(tick0, tick1, dateInterval string) ([]*utils.ExchangeInfoSummary, error) {
	query := `SELECT tick0, tick1, open_price, close_price, lowest_ask, highest_bid, base_volume, quote_volume, last_date FROM exchange_summary WHERE tick0  = ? and tick1 = ? and date_interval = ? ORDER BY last_date DESC LIMIT 1500`
	rows, err := c.MysqlDB.Query(query, tick0, tick1, dateInterval)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*utils.ExchangeInfoSummary, 0)
	for rows.Next() {
		result := &utils.ExchangeInfoSummary{}
		var baseVolume, quoteVolume string
		err := rows.Scan(&result.Tick0, &result.Tick1, &result.OpenPrice, &result.ClosePrice, &result.LowestAsk, &result.HighestBid, &baseVolume, &quoteVolume, &result.LastDate)
		if err != nil {
			return nil, err
		}

		result.BaseVolume, _ = utils.ConvetStr(baseVolume)
		result.QuoteVolume, _ = utils.ConvetStr(quoteVolume)
		results = append(results, result)
	}

	return results, nil
}
