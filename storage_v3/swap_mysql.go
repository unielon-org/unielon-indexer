package storage_v3

import (
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	MINI_LIQUIDITY = 1000
)

func (c *MysqlClient) FindSwapInfoById(OrderId string) (*models.SwapInfo, error) {
	query := "SELECT  order_id, op, tick0, tick1, amt0, amt1, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address,  update_date, create_date   FROM swap_info where order_id = ?"
	rows, err := c.MysqlDB.Query(query, OrderId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		swap := &models.SwapInfo{}
		var amt0, amt1 string
		err := rows.Scan(&swap.OrderId, &swap.Op, &swap.Tick0, &swap.Tick1, &amt0, &amt1, &swap.FeeTxHash, &swap.TxHash, &swap.BlockHash, &swap.BlockNumber, &swap.FeeAddress, &swap.HolderAddress, &swap.UpdateDate, &swap.CreateDate)
		if err != nil {
			return nil, err
		}

		swap.Amt0, _ = utils.ConvetStringToNumber(amt0)
		swap.Amt1, _ = utils.ConvetStringToNumber(amt1)

		return swap, err
	}
	return nil, nil
}

func (c *MysqlClient) FindSwapInfoVolumeByTick(tick0, tick1 string) (float64, int64, error) {

	swapPrice, _, err := c.FindSwapPriceAll()
	if err != nil {
		return 0, 0, err
	}

	prices := make(map[string]float64)
	for _, price := range swapPrice {
		prices[price.Tick] = price.LastPrice
	}

	query := `
			SELECT COALESCE(CAST(SUM(CASE
			                             WHEN tick0 = 'WDOGE(WRAPPED-DOGE)' THEN amt0
			                             WHEN tick1 = 'WDOGE(WRAPPED-DOGE)' THEN amt1_out
			                             ELSE 0 END) AS DECIMAL(32, 0)), 0) AS total_amount,
			       COALESCE(CAST(SUM(amt0) AS DECIMAL(32, 0)) , 0)  AS total_amount_in,
			       COALESCE(CAST(SUM(amt1_out) AS DECIMAL(32, 0)) , 0)  AS total_amount_out
			FROM swap_info
			WHERE op = 'swap'
			  and ((tick0 = ? and tick1 = ?) or (tick0 = ? and tick1 = ?))
			  and block_number > 0
			  and block_hash != ''`

	rows, err := c.MysqlDB.Query(query, tick0, tick1, tick1, tick0)
	if err != nil {
		return 0, 0, err
	}

	defer rows.Close()

	var volume float64
	if rows.Next() {
		var volumeDoge float64
		var volumeIn float64
		var volumeOut float64
		err := rows.Scan(&volumeDoge, &volumeIn, &volumeOut)
		if err != nil {
			return 0, 0, err
		}

		if tick0 == "WDOGE(WRAPPED-DOGE)" || tick1 == "WDOGE(WRAPPED-DOGE)" {
			volume = volumeDoge
		} else {

			valdoge := 0.0
			if _, ok := prices[tick0]; ok {
				valdoge = prices[tick0] * volumeIn
			} else {
				valdoge = prices[tick1] * volumeOut
			}

			volume = valdoge
		}

	}

	query1 := " SELECT count(order_id) FROM swap_info WHERE op = 'swap' and ((tick0 =? and tick1 = ?) or (tick0 =? and tick1 = ?)) "
	rows1, err := c.MysqlDB.Query(query1, tick0, tick1, tick1, tick0)
	if err != nil {
		return 0, 0, err
	}

	defer rows1.Close()
	if rows1.Next() {
		var count int64
		err := rows1.Scan(&count)
		if err != nil {
			return 0, 0, err
		}
		return volume, count, err
	}

	return 0, 0, nil
}

func (c *MysqlClient) FindSwapInfoVolumeAll() (map[string]float64, error) {

	swapPrice, _, err := c.FindSwapPriceAll()
	if err != nil {
		return nil, err
	}

	prices := make(map[string]float64)
	for _, price := range swapPrice {
		prices[price.Tick] = price.LastPrice
	}

	query := `
			SELECT tick0,
			       tick1,
			       COALESCE(CAST(SUM(CASE  WHEN tick0 = 'WDOGE(WRAPPED-DOGE)' THEN amt0    WHEN tick1 = 'WDOGE(WRAPPED-DOGE)' THEN amt1_out ELSE 0 END) AS DECIMAL(32, 0)) , 0)  AS total_amount,
			       COALESCE(CAST(SUM(amt0) AS DECIMAL(32, 0)) , 0)  AS total_amount_in,
			       COALESCE(CAST(SUM(amt1_out) AS DECIMAL(32, 0)) , 0)  AS total_amount_out
			FROM swap_info
			WHERE op = 'swap' AND update_date >= NOW() - INTERVAL 24 HOUR and block_number > 0  and block_hash != ''
			GROUP BY tick0, tick1;`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	volumeMap := make(map[string]float64)

	for rows.Next() {
		var tick0, tick1 string
		var volumeDoge float64
		var volumeIn float64
		var volumeOut float64
		err := rows.Scan(&tick0, &tick1, &volumeDoge, &volumeIn, &volumeOut)
		if err != nil {
			return nil, err
		}

		if tick0 == "WDOGE(WRAPPED-DOGE)" || tick1 == "WDOGE(WRAPPED-DOGE)" {
			tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)
			if _, ok := volumeMap[tick0+"-SWAP-"+tick1]; ok {
				volumeMap[tick0+"-SWAP-"+tick1] += volumeDoge
			} else {
				volumeMap[tick0+"-SWAP-"+tick1] = volumeDoge
			}
		} else {

			valdoge := 0.0
			if _, ok := prices[tick0]; ok {
				valdoge = prices[tick0] * volumeIn
			} else {
				valdoge = prices[tick1] * volumeOut
			}

			tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)
			if _, ok := volumeMap[tick0+"-SWAP-"+tick1]; ok {
				volumeMap[tick0+"-SWAP-"+tick1] += valdoge
			} else {
				volumeMap[tick0+"-SWAP-"+tick1] = valdoge
			}
		}
	}

	return volumeMap, nil
}

func (c *MysqlClient) FindSwapInfo(orderId, op, tick, tick0, tick1, holder_address string, limit, offset int64) ([]*SwapInfo, int64, error) {
	query := "SELECT  order_id, op, tick0, tick1, amt0, amt1, amt0_min, amt1_min, amt0_out, amt1_out, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, order_status, update_date, create_date   FROM swap_info  "

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
		where += "  (tick0 = ? or tick1 = ?) "
		whereAges = append(whereAges, tick, tick)
	}

	if tick0 != "" {
		if where != "where" {
			where += " and "
		}
		where += "  ((tick0 = ? and tick1 = ?) or (tick1 = ? and tick0 = ?)) "
		whereAges = append(whereAges, tick0, tick1, tick0, tick1)
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
	swaps := make([]*SwapInfo, 0)
	for rows.Next() {
		swap := &SwapInfo{}
		var amt0, amt1, amt0min, amt1min, amt0out, amt1out string
		err := rows.Scan(&swap.OrderId, &swap.Op, &swap.Tick0, &swap.Tick1, &amt0, &amt1, &amt0min, &amt1min, &amt0out, &amt1out, &swap.FeeTxHash, &swap.SwapTxHash, &swap.SwapBlockHash, &swap.SwapBlockNumber, &swap.FeeAddress, &swap.HolderAddress, &swap.OrderStatus, &swap.UpdateDate, &swap.CreateDate)
		if err != nil {
			return nil, 0, err
		}

		swap.Amt0, _ = utils.ConvetStr(amt0)
		swap.Amt1, _ = utils.ConvetStr(amt1)
		swap.Amt0Min, _ = utils.ConvetStr(amt0min)
		swap.Amt1Min, _ = utils.ConvetStr(amt1min)
		swap.Amt0Out, _ = utils.ConvetStr(amt0out)
		swap.Amt1Out, _ = utils.ConvetStr(amt1out)
		swaps = append(swaps, swap)
	}

	query1 := "SELECT count(order_id)  FROM swap_info "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
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

func (c *MysqlClient) FindSwapLiquidityAll() ([]*models.SwapLiquidity, int64, error) {
	query := "SELECT tick, tick0, tick1, amt0, amt1, liquidity_total, close_price from swap_liquidity where liquidity_total != '0'"
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*models.SwapLiquidity, 0)
	for rows.Next() {
		liquidity := &models.SwapLiquidity{}
		var amt0, amt1, liquidity_total string
		err := rows.Scan(&liquidity.Tick, &liquidity.Tick0, &liquidity.Tick1, &amt0, &amt1, &liquidity_total, &liquidity.ClosePrice)
		if err != nil {
			return nil, 0, err
		}
		liquidity.Amt0, _ = utils.ConvetStringToNumber(amt0)
		liquidity.Amt1, _ = utils.ConvetStringToNumber(amt1)
		liquidity.LiquidityTotal, _ = utils.ConvetStringToNumber(liquidity_total)
		list = append(list, liquidity)
	}

	query1 := "SELECT count(tick)  FROM swap_liquidity  where liquidity_total != '0'"
	rows1, err := c.MysqlDB.Query(query1)
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

func (c *MysqlClient) FindSwapLiquidity(tick0 string, tick1 string) (*models.SwapLiquidity, error) {
	query := "SELECT tick, tick0, tick1, amt0, amt1, holder_address, liquidity_total, reserves_address, close_price from swap_liquidity where tick0 = ? and tick1 = ? "
	tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)
	rows, err := c.MysqlDB.Query(query, tick0, tick1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquidity := &models.SwapLiquidity{}
	if rows.Next() {
		var amt0, amt1, liquidity_total string
		err := rows.Scan(&liquidity.Tick, &liquidity.Tick0, &liquidity.Tick1, &amt0, &amt1, &liquidity.HolderAddress, &liquidity_total, &liquidity.ReservesAddress, &liquidity.ClosePrice)
		if err != nil {
			return nil, err
		}
		liquidity.Amt0, _ = utils.ConvetStringToNumber(amt0)
		liquidity.Amt1, _ = utils.ConvetStringToNumber(amt1)
		liquidity.LiquidityTotal, _ = utils.ConvetStringToNumber(liquidity_total)
		return liquidity, nil
	}
	return nil, nil
}

func (c *MysqlClient) FindSwapLiquidityWeb(tick0 string, tick1 string) (*models.SwapLiquidity, error) {
	query := "SELECT tick, tick0, tick1, amt0, amt1, holder_address, liquidity_total, reserves_address,close_price from swap_liquidity where tick0 = ? and tick1 = ? and liquidity_total != '0'"
	tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)
	rows, err := c.MysqlDB.Query(query, tick0, tick1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquidity := &models.SwapLiquidity{}
	if rows.Next() {
		var amt0, amt1, liquidity_total string
		err := rows.Scan(&liquidity.Tick, &liquidity.Tick0, &liquidity.Tick1, &amt0, &amt1, &liquidity.HolderAddress, &liquidity_total, &liquidity.ReservesAddress, &liquidity.ClosePrice)
		if err != nil {
			return nil, err
		}
		liquidity.Amt0, _ = utils.ConvetStringToNumber(amt0)
		liquidity.Amt1, _ = utils.ConvetStringToNumber(amt1)
		liquidity.LiquidityTotal, _ = utils.ConvetStringToNumber(liquidity_total)
		return liquidity, nil
	}
	return nil, nil
}

func (c *MysqlClient) FindSwapLiquidityLP(tick string) ([]*models.SwapLiquidityLP, error) {
	query := "SELECT amt_sum, holder_address from drc20_collect_address where tick = ? "
	rows, err := c.MysqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquiditys := make([]*models.SwapLiquidityLP, 0)
	for rows.Next() {
		liquidity := &models.SwapLiquidityLP{}
		var amt_sum string
		err := rows.Scan(&amt_sum, &liquidity.HolderAddress)
		if err != nil {
			return nil, err
		}
		liquidity.Liquidity, _ = utils.ConvetStr(amt_sum)
		liquiditys = append(liquiditys, liquidity)
	}
	return liquiditys, nil
}

func (c *MysqlClient) FindSwapLiquidityByHolder(holder_address string, tick0, tick1 string) ([]*models.SwapLiquidity, error) {

	query := "SELECT tick, amt_sum from drc20_collect_address where holder_address = ? and LENGTH(tick) >= 10 and tick != 'WDOGE(WRAPPED-DOGE)' and amt_sum != '0'"
	queryf := []any{holder_address}

	if tick0 != "" {
		query = "SELECT tick, amt_sum from drc20_collect_address where holder_address = ? and LENGTH(tick) >= 10 and tick = ? and tick != 'WDOGE(WRAPPED-DOGE)' and amt_sum != '0'"
		queryf = []any{holder_address, tick0 + "-SWAP-" + tick1}
	}

	rows, err := c.MysqlDB.Query(query, queryf...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	liquiditys := make([]*models.SwapLiquidity, 0)

	for rows.Next() {
		liquidity := &models.SwapLiquidity{}
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

		amt_sum_big, _ := utils.ConvetStringToNumber(amt_sum)
		liquidity.Amt0 = (*models.Number)(new(big.Int).Div(new(big.Int).Mul(swapLiquidity.Amt0.Int(), amt_sum_big.Int()), swapLiquidity.LiquidityTotal.Int()))
		liquidity.Amt1 = (*models.Number)(new(big.Int).Div(new(big.Int).Mul(swapLiquidity.Amt1.Int(), amt_sum_big.Int()), swapLiquidity.LiquidityTotal.Int()))
		liquidity.Tick0 = tick0
		liquidity.Tick1 = tick1
		liquidity.LiquidityTotal = amt_sum_big
		liquiditys = append(liquiditys, liquidity)
	}

	return liquiditys, nil
}

func (c *MysqlClient) FindSwapPriceAll() ([]*SwapPrice, int64, error) {

	liquidityAll, _, err := c.FindSwapLiquidityAll()
	if err != nil {
		return nil, 0, err
	}

	liquidityMap := make(map[string]*big.Float)
	noWdoge := make([]*models.SwapLiquidity, 0)
	for _, v := range liquidityAll {
		if v.Tick0 != "WDOGE(WRAPPED-DOGE)" && v.Tick1 != "WDOGE(WRAPPED-DOGE)" {
			noWdoge = append(noWdoge, v)
			continue
		}

		if v.Tick0 == "WDOGE(WRAPPED-DOGE)" {
			liquidityMap[v.Tick1] = new(big.Float).Quo(new(big.Float).SetInt(v.Amt0.Int()), new(big.Float).SetInt(v.Amt1.Int()))
		} else {
			liquidityMap[v.Tick0] = new(big.Float).Quo(new(big.Float).SetInt(v.Amt1.Int()), new(big.Float).SetInt(v.Amt0.Int()))
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

	liquidityMapNoDoge := make(map[string]*big.Float)
	for _, v := range noWdoge {
		price0, ok0 := liquidityMap[v.Tick0]
		price1, ok1 := liquidityMap[v.Tick1]
		if ok0 && ok1 {
			continue
		}

		if ok0 {

			price := new(big.Float).Quo(new(big.Float).SetInt(v.Amt0.Int()), new(big.Float).SetInt(v.Amt1.Int()))
			price = new(big.Float).Mul(price, price0)
			liquidityMapNoDoge[v.Tick1] = price

		} else if ok1 {

			price := new(big.Float).Quo(new(big.Float).SetInt(v.Amt1.Int()), new(big.Float).SetInt(v.Amt0.Int()))
			price = new(big.Float).Mul(price, price1)
			liquidityMapNoDoge[v.Tick0] = price
		}
	}

	for k, v := range liquidityMapNoDoge {
		f, _ := v.Float64()
		liquiditys = append(liquiditys, &SwapPrice{
			Tick:      k,
			LastPrice: f,
		})
		len++
	}

	return liquiditys, len, nil
}

func (c *MysqlClient) FindSWAPTempOrder(HolderAddress string) (int64, error) {
	query := "SELECT count(holder_address) FROM swap_info where fee_tx_hash = '' and holder_address = ? and create_date > NOW() - INTERVAL 10 MINUTE "
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

func (c *MysqlClient) FindSwapSummaryAll() ([]*utils.SummaryAllResult, error) {
	query := `
SELECT
    d20i.tick,
    d20i.amt_sum,
    d20i.max_,
    COALESCE(e.close_price * d20i.amt_sum, 0) AS total_doge_amt,
    COALESCE(e.close_price,0) AS closePrice,
    COALESCE(e.base_volume, 0) AS totalBaseVolume, -- 使用聚合得到的总和
   COALESCE(((e.close_price - e.open_price) / e.open_price) * 100, 0) AS priceChange,
    (SELECT COUNT(holder_address) FROM drc20_collect_address WHERE drc20_collect_address.tick = e.tick) AS receive_address_count,
    COALESCE(e.lowest_ask, 0) AS footPrice,
    d20i.logo,
    d20i.is_check
FROM
    drc20_collect d20i
LEFT JOIN (
    SELECT
        es.tick,
        es.close_price,
        es.lowest_ask,
        es.base_volume,
        es.open_price,
        es.id
    FROM
        swap_summary es
    INNER JOIN (
        SELECT
            tick,
            MAX(id) AS max_id
        FROM
            swap_summary
        WHERE
            date_interval = '1d'
        GROUP BY
            tick
    ) es_max ON es.id = es_max.max_id
) e ON e.tick = d20i.tick
WHERE
    LENGTH(d20i.tick) < 9
ORDER BY
    totalBaseVolume DESC, receive_address_count DESC

`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*utils.SummaryAllResult, 0)
	for rows.Next() {
		var marketcap *string
		result := &utils.SummaryAllResult{}
		err := rows.Scan(&result.Tick, &result.AmtSum, &result.MaxAmt, &marketcap, &result.LastPrice, &result.BaseVolume, &result.PriceChangePercent24H, &result.Holders, &result.FootPrice, &result.Logo, &result.IsCheck)
		if err != nil {
			return nil, err
		}

		f, err := strconv.ParseFloat(*marketcap, 64)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return nil, err
		}

		result.MarketCap = strconv.FormatInt(int64(f), 10)
		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
		}

		results = append(results, result)
	}
	return results, nil
}

func (c *MysqlClient) FindSwapSummaryByTick(tick string) (*utils.SummaryAllResult, error) {
	query := `
SELECT
    d20i.tick,
    COALESCE(e.close_price * d20i.amt_sum, 0) AS total_doge_amt,
    COALESCE(e.close_price,0) AS closePrice,
    COALESCE(e.base_volume, 0) AS totalBaseVolume, 
    COALESCE(((e.close_price - e.open_price) / e.open_price) * 100, 0) AS priceChange,
    (SELECT COUNT(holder_address) FROM drc20_collect_address WHERE drc20_collect_address.tick = e.tick) AS receive_address_count,
    COALESCE(e.lowest_ask, 0) AS footPrice,
    d20i.max_,
     d20i.amt_sum,
    d20i.logo,
    d20i.is_check,
    COALESCE(l.liquidity, 0) AS liquidity
FROM
    drc20_collect d20i
LEFT JOIN (
    SELECT
        es.tick,
  		es.close_price,
        es.lowest_ask,
        es.base_volume,
        es.open_price,
        es.id
    FROM
        swap_summary es
    INNER JOIN (
        SELECT
            tick,
            MAX(id) AS max_id
        FROM
            swap_summary
        WHERE
            date_interval = '1d'
        GROUP BY
            tick
    ) es_max ON es.id = es_max.max_id
) e ON e.tick = d20i.tick
LEFT JOIN (
    SELECT
        ? AS tick,
        SUM(liquidity) AS liquidity
    FROM
        swap_summary_liquidity
    WHERE
        (tick0 = ? OR tick1 = ?)
    AND  last_date = ?
) l ON l.tick = d20i.tick
WHERE
    d20i.tick = ?;
`
	const layout = "2006-01-02 15:04:05"
	startDate := time.Now()
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	rows, err := c.MysqlDB.Query(query, tick, tick, tick, startDate.Format(layout), tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var marketcap *string
		var Liquidity string

		result := &utils.SummaryAllResult{}
		err := rows.Scan(&result.Tick, &marketcap, &result.LastPrice, &result.BaseVolume, &result.PriceChangePercent24H, &result.Holders, &result.FootPrice, &result.AmtSum, &result.MaxAmt, &result.Logo, &result.IsCheck, &Liquidity)
		if err != nil {
			return nil, err
		}

		f, err := strconv.ParseFloat(*marketcap, 64)
		if err != nil {
			return nil, err
		}

		result.MarketCap = strconv.FormatInt(int64(f), 10)
		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
		}

		f1, err := strconv.ParseFloat(Liquidity, 64)
		if err != nil {
			return nil, err
		}

		result.Liquidity = f1

		return result, nil

	}
	return nil, nil
}

func (c *MysqlClient) FindSwapPairByTick(tick string) ([]*utils.SwapPairSummary, error) {
	query := `
SELECT
    es.tick,
    es.tick0,
    es.tick1,
    COALESCE(((es.close_price - es.open_price) / es.open_price) * 100, 0) AS priceChange,
    es.liquidity * 2,
    es.base_volume,
    es.doge_usdt,
    sl.amt0,
    sl.amt1
FROM
    swap_summary_liquidity es
INNER JOIN (
    SELECT
        tick,
        MAX(id) AS max_id
    FROM
        swap_summary_liquidity
    WHERE
        tick0 = ? OR tick1 = ?
    GROUP BY
        tick
) es_max ON es.id = es_max.max_id
LEFT JOIN swap_liquidity sl ON es.tick = sl.tick
ORDER BY es.liquidity DESC;
`
	rows, err := c.MysqlDB.Query(query, tick, tick)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*utils.SwapPairSummary, 0)
	for rows.Next() {
		var Liquidity string
		result := &utils.SwapPairSummary{}
		err := rows.Scan(&result.Tick, &result.Tick0, &result.Tick1, &result.PriceChangePercent24H, &Liquidity, &result.BaseVolume, &result.DogeUsdt, &result.Amt0, &result.Amt1)
		if err != nil {
			return nil, err
		}

		f1, err := strconv.ParseFloat(Liquidity, 64)
		if err != nil {
			return nil, err
		}

		result.Liquidity = f1
		results = append(results, result)

	}
	return results, nil
}

func (c *MysqlClient) FindSwapPairAll() ([]*utils.SwapPairSummary, error) {
	query := `
SELECT
    es.tick,
    es.tick0,
    es.tick1,
    COALESCE(((es.close_price - es.open_price) / es.open_price) * 100, 0) AS priceChange,
    es.liquidity * 2,
    es.base_volume,
    es.doge_usdt,
 COALESCE(sl.amt0, 0) AS amt0,
COALESCE(sl.amt1, 0) AS amt1
FROM
    swap_summary_liquidity es
INNER JOIN (
    SELECT
        tick,
        MAX(id) AS max_id
    FROM
        swap_summary_liquidity
    GROUP BY
        tick
) es_max ON es.id = es_max.max_id
LEFT JOIN swap_liquidity sl ON es.tick = sl.tick
ORDER BY es.liquidity DESC;
`
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*utils.SwapPairSummary, 0)
	for rows.Next() {
		var Liquidity string
		result := &utils.SwapPairSummary{}
		err := rows.Scan(&result.Tick, &result.Tick0, &result.Tick1, &result.PriceChangePercent24H, &Liquidity, &result.BaseVolume, &result.DogeUsdt, &result.Amt0, &result.Amt1)
		if err != nil {
			return nil, err
		}

		f1, err := strconv.ParseFloat(Liquidity, 64)
		if err != nil {
			return nil, err
		}

		result.Liquidity = f1
		results = append(results, result)

	}
	return results, nil
}
