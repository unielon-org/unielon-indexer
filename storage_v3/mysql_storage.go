package storage_v3

import (
	"database/sql"
	"errors"
	"github.com/dogecoinw/go-dogecoin/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"strings"
	"sync"
	"time"
)

const (
	NETWORK          = "tcp"
	stakePoolAddress = "DS8eFcobjXp6oL8YoXoVazDQ32bcDdWwui"
)

var (
	ErrNotFound = errors.New("not found")
)

type MysqlClient struct {
	MysqlDB *sql.DB
	lock    *sync.RWMutex
}

func NewSqliteClient(cfg utils.SqliteConfig) *MysqlClient {

	db, err := sql.Open("sqlite3", cfg.Database)
	if err != nil {
		log.Error("NewMysqlClient", "err", err)
		return nil
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	_, err = db.Exec("PRAGMA busy_timeout=3000;")
	_, err = db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		log.Error("NewMysqlClient", "err", err)
		return nil
	}

	lock := new(sync.RWMutex)
	conn := &MysqlClient{
		MysqlDB: db,
		lock:    lock,
	}

	return conn
}

func (conn *MysqlClient) Stop() {
	conn.MysqlDB.Close()
}

func (c *MysqlClient) FindOgAddress() ([]string, error) {
	query := "SELECT receive_address  FROM address_info_og"
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	address := make([]string, 0)
	for rows.Next() {
		var add string
		err := rows.Scan(&add)
		if err != nil {
			return nil, err
		}
		address = append(address, add)
	}
	return address, nil
}

func (c *MysqlClient) FindCMCSummaryK(tick, dateInterval string) ([]*SwapInfoSummary, error) {
	query := `SELECT tick,  open_price, close_price, lowest_ask, highest_bid, base_volume, last_date FROM swap_summary WHERE tick = ? and date_interval = ? ORDER BY update_date DESC LIMIT 1500`
	rows, err := c.MysqlDB.Query(query, tick, dateInterval)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*SwapInfoSummary, 0)
	for rows.Next() {
		result := &SwapInfoSummary{}
		var baseVolume, quoteVolume string
		err := rows.Scan(&result.Tick, &result.Tick0, &result.Tick1, &result.OpenPrice, &result.ClosePrice, &result.LowestAsk, &result.HighestBid, &baseVolume, &quoteVolume, &result.LastDate)
		if err != nil {
			return nil, err
		}

		result.BaseVolume, _ = utils.ConvetStr(baseVolume)
		result.QuoteVolume, _ = utils.ConvetStr(quoteVolume)
		results = append(results, result)
	}

	return results, nil
}

func (c *MysqlClient) FindCMCSummaryK2(tick, dateInterval string) ([]*SwapInfoSummary, error) {
	query := `SELECT tick,  open_price, close_price, lowest_ask, highest_bid, base_volume, last_date, doge_usdt FROM swap_summary WHERE tick = ? and date_interval = ? ORDER BY last_date desc LIMIT 1500`
	rows, err := c.MysqlDB.Query(query, tick, dateInterval)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*SwapInfoSummary, 0)
	for rows.Next() {
		result := &SwapInfoSummary{}
		var baseVolume string
		err := rows.Scan(&result.Tick, &result.OpenPrice, &result.ClosePrice, &result.LowestAsk, &result.HighestBid, &baseVolume, &result.LastDate, &result.DogeUsdt)
		if err != nil {
			return nil, err
		}

		result.BaseVolume, _ = utils.ConvetStr(baseVolume)
		results = append(results, result)
	}

	return results, nil
}

func (c *MysqlClient) FindCMCSummaryTVLAll() ([]*SwapInfoSummaryTVLAll, error) {
	query := `
SELECT
    A.TotalLiquidity,
    B.TotalBaseVolume,
    A.MaxDogeUsdt,
    A.last_date
FROM
    (SELECT
         SUM(liquidity) * 2 AS TotalLiquidity,
         MAX(doge_usdt) AS MaxDogeUsdt,
         last_date
     FROM
         swap_summary_liquidity
     GROUP BY
         last_date) A
JOIN
    (SELECT
         SUM(base_volume) AS TotalBaseVolume,
         last_date
     FROM
         swap_summary
     WHERE
         date_interval='1d'
     GROUP BY
         last_date) B
ON
    A.last_date = B.last_date;
`
	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*SwapInfoSummaryTVLAll, 0)
	for rows.Next() {
		result := &SwapInfoSummaryTVLAll{}
		var Liquidity, BaseVolume string
		err := rows.Scan(&Liquidity, &BaseVolume, &result.DogeUsdt, &result.LastDate)
		if err != nil {
			return nil, err
		}

		result.Liquidity, _ = utils.ConvetStr(Liquidity)
		result.BaseVolume, _ = utils.ConvetStr(BaseVolume)

		results = append(results, result)
	}

	return results, nil
}

func (c *MysqlClient) FindCMCSummaryTVL(tick0, tick1 string) ([]*SwapInfoSummaryTVLAll, error) {
	query := `
			SELECT sum(liquidity * 2),
				   sum(base_volume),
				   doge_usdt,
				   max(last_date)
			FROM swap_summary_liquidity
			WHERE tick0 = ?
			   OR tick1 = ?
			group by last_date, doge_usdt;
`

	if tick0 != tick1 {
		query = `
			SELECT sum(liquidity * 2),
			   sum(base_volume),
			   doge_usdt,
			   max(last_date)
			FROM swap_summary_liquidity
			WHERE tick0 = ?
			   AND tick1 = ?
			group by last_date, doge_usdt;
			`
	}
	rows, err := c.MysqlDB.Query(query, tick0, tick1)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]*SwapInfoSummaryTVLAll, 0)
	for rows.Next() {
		result := &SwapInfoSummaryTVLAll{}
		var Liquidity, BaseVolume string
		err := rows.Scan(&Liquidity, &BaseVolume, &result.DogeUsdt, &result.LastDate)
		if err != nil {
			return nil, err
		}

		result.Liquidity, _ = utils.ConvetStr(Liquidity)
		result.BaseVolume, _ = utils.ConvetStr(BaseVolume)

		results = append(results, result)
	}

	return results, nil
}

func (c *MysqlClient) FindCMCSummaryKNew(tick, dateInterval string) (*SwapInfoSummary, error) {
	const layout = "2006-01-02 15:04:05" // This layout is used to parse the date-time string in Go
	startDate, err := utils.TimeCom(dateInterval)
	if err != nil {
		return nil, err
	}

	tick0, tick1 := strings.Split(tick, "-SWAP-")[0], strings.Split(tick, "-SWAP-")[1]
	tick0, tick1, _, _, _, _ = utils.SortTokens(tick0, tick1, nil, nil, nil, nil)

	query := "select op, tick0, tick1, amt0, amt1, amt1_out,update_date, create_date  FROM swap_info where update_date >= ? and op = 'swap' and block_number > 0  and block_hash != '' and ((tick0 = ? and tick1 = ?) or (tick1 = ? and tick0 = ?) )"
	rows, err := c.MysqlDB.Query(query, startDate.Format(layout), tick0, tick1, tick0, tick1)
	if err != nil {
		log.Error("QuerySwapInfoByDate", "err", err)
		return nil, err
	}

	defer rows.Close()
	swaps := make([]*models.SwapInfo, 0)
	for rows.Next() {
		swap := &models.SwapInfo{}
		var amt0, amt1, amt1out string
		err := rows.Scan(&swap.Op, &swap.Tick0, &swap.Tick1, &amt0, &amt1, &amt1out, &swap.UpdateDate, &swap.CreateDate)
		if err != nil {
			return nil, err
		}

		swap.Amt0, _ = utils.ConvetStringToNumber(amt0)
		swap.Amt1, _ = utils.ConvetStringToNumber(amt1)
		swap.Amt1Out, _ = utils.ConvetStringToNumber(amt1out)
		swaps = append(swaps, swap)
	}

	summs := make(map[string]*SwapInfoSummary)

	for _, swap := range swaps {

		tick0, tick1, amt0, amt1, _, _ := utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1Out, nil, nil)

		summ := &SwapInfoSummary{}
		summ.Tick = tick
		summ.Tick0 = tick0
		summ.Tick1 = tick1
		summ.LastDate = time.Now().Format(layout)

		amt0f := new(big.Float).SetInt(amt0.Int())
		amt1f := new(big.Float).SetInt(amt1.Int())
		price, _ := new(big.Float).Quo(amt0f, amt1f).Float64()
		if summs[summ.Tick] == nil {
			summ.OpenPrice = price
			summ.ClosePrice = price
			summ.LowestAsk = price
			summ.HighestBid = price
			summ.BaseVolume = amt0.Int()
			summ.QuoteVolume = amt1.Int()
			summs[summ.Tick] = summ
		} else {
			summOld := summs[summ.Tick]
			summOld.BaseVolume = new(big.Int).Add(summOld.BaseVolume, amt0.Int())
			summOld.QuoteVolume = new(big.Int).Add(summOld.QuoteVolume, amt1.Int())
			summOld.ClosePrice = price
			if price > summOld.HighestBid {
				summOld.HighestBid = price
			}

			if price < summOld.LowestAsk {
				summOld.LowestAsk = price
			}
			summs[summ.Tick] = summOld
		}
	}

	return summs[tick], err

}
