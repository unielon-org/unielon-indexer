package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
	"strings"
)

type ExchangeRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewExchangeRouter(db *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *ExchangeRouter {
	return &ExchangeRouter{
		dbc:    db,
		node:   node,
		verify: verify,
	}
}

func (r *ExchangeRouter) Collect(c *gin.Context) {

	type params struct {
		ExId          string `json:"exid"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		NotDone       bool   `json:"not_done"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}

	p := &params{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	exc := make([]*models.ExchangeCollect, 0)
	total := int64(0)

	filter := &models.ExchangeCollect{
		ExId:          p.ExId,
		Tick0:         p.Tick0,
		Tick1:         p.Tick1,
		HolderAddress: p.HolderAddress,
	}

	subQuery := r.dbc.DB.Model(&models.ExchangeCollect{}).
		Where(filter)

	if p.NotDone {
		subQuery = subQuery.Where("amt0 != amt0_finish")
	}

	err := subQuery.Limit(p.Limit).
		Offset(p.OffSet).
		Find(&exc).
		Count(&total).Error

	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = exc
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *ExchangeRouter) Order(c *gin.Context) {
	type params struct {
		OrderId       string `json:"order_id"`
		ExId          string `json:"exid"`
		Op            string `json:"op"`
		Tick          string `json:"tick"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		BlockNumber   int64  `json:"block_number"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}

	p := &params{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	filter := &models.ExchangeInfo{
		OrderId:       p.OrderId,
		ExId:          p.ExId,
		Op:            p.Op,
		Tick0:         p.Tick0,
		Tick1:         p.Tick1,
		HolderAddress: p.HolderAddress,
		BlockNumber:   p.BlockNumber,
	}

	infos := make([]*models.ExchangeInfo, 0)
	total := int64(0)
	subQuery := r.dbc.DB.Model(&models.ExchangeInfo{}).Where(filter)

	if p.Tick != "" {
		subQuery = subQuery.Where("( tick0 = ? or tick1 = ?)", p.Tick, p.Tick)
	}

	err := subQuery.Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&infos).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusOK, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = infos
	result.Total = total

	c.JSON(http.StatusOK, result)

}

func (r *ExchangeRouter) SummaryTotal(c *gin.Context) {

	type FindSummaryResult struct {
		Exchange int64   `json:"exchange"`
		Value24h float64 `json:"value_all"`
	}

	sr := &FindSummaryResult{}

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

	err := r.dbc.DB.Raw(query).Scan(sr).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = sr

	c.JSON(http.StatusOK, result)

}

func (r *ExchangeRouter) Summary(c *gin.Context) {
	type params struct {
		Tick   string `json:"tick"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}

	p := &params{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	type Drc20InfoResult struct {
		TradingPairs          string  `json:"trading_pairs"`
		Tick                  string  `json:"tick"`
		TotalMaxAmt           string  `json:"total_max_amt"`
		LastPrice             float64 `json:"last_price"`
		OldPrice              float64 `json:"old_price"`
		LowestAsk             float64 `json:"lowest_ask"`
		HighestBid            float64 `json:"highest_bid"`
		BaseVolume            float64 `json:"base_volume"`
		QuoteVolume           float64 `json:"quote_volume"`
		PriceChangePercent24H float64 `json:"price_change_percent_24h"`
		HighestPrice24H       float64 `json:"highest_price_24h"`
		LowestPrice24H        float64 `json:"lowest_price_24h"`
		LastDate              string  `json:"last_date"`
		Holders               *uint64 `json:"holders"`
		FootPrice             float64 `json:"foot_price"`
		Logo                  *string `json:"logo"`
		IsCheck               uint64  `json:"is_check"`
		Liquidity             float64 `json:"liquidity"`
	}

	var results []Drc20InfoResult
	var totalCount int64

	subQuery := r.dbc.DB.Table("exchange_info_summary es").
		Select("es.tick0, es.close_price, es.lowest_ask, es.quote_volume, es.open_price").
		Joins("INNER JOIN (SELECT tick0, MAX(id) AS max_id FROM exchange_info_summary WHERE date_interval = '1d' AND (tick0 = 'WDOGE(WRAPPED-DOGE)' OR tick1 = 'WDOGE(WRAPPED-DOGE)') GROUP BY tick0) es_max ON es.id = es_max.max_id")

	mainQuery := r.dbc.DB.Table("drc20_collect d20i").
		Select(`
        d20i.tick,
        COALESCE(e.close_price * d20i.amt_sum, 0) AS total_doge_amt,
        COALESCE(e.close_price, 0) AS close_price,
        COALESCE(e.quote_volume, 0) AS total_quote_volume,
        COALESCE(((e.close_price - e.open_price) / e.open_price) * 100, 0) AS priceChange,
        (SELECT COUNT(holder_address) FROM drc20_collect_address WHERE drc20_collect_address.tick = e.tick0) AS receive_address_count,
        COALESCE(e.lowest_ask, 0) AS footPrice,
        d20i.logo,
        d20i.is_check`).
		Joins("LEFT JOIN (?) e ON e.tick0 = d20i.tick", subQuery).
		Where("LENGTH(d20i.tick) < 9")

	if p.Tick != "" {
		mainQuery = mainQuery.Where("d20i.tick = ?", p.Tick)
	}

	mainQuery = mainQuery.Order("total_quote_volume DESC, receive_address_count DESC")

	err := mainQuery.Count(&totalCount).Scan(&results).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusOK, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Total = totalCount
	result.Data = results
	c.JSON(http.StatusOK, result)
}

func (r *ExchangeRouter) SummaryK(c *gin.Context) {
	type params struct {
		Tick0        string `json:"tick0"`
		Tick1        string `json:"tick1"`
		DateInterval string `json:"date_interval"`
		Limit        int    `json:"limit"`
		Offset       int    `json:"offset"`
	}

	p := &params{
		DateInterval: "1d",
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	p.DateInterval = strings.ToLower(p.DateInterval)
	p.Tick0, p.Tick1, _, _, _, _ = utils.SortTokens(p.Tick0, p.Tick1, nil, nil, nil, nil)

	var results []models.ExchangeSummary
	var total int64
	query := r.dbc.DB.Model(&results).
		Where("tick0 = ? AND tick1 = ? AND date_interval = ?", p.Tick0, p.Tick1, p.DateInterval).
		Count(&total).
		Order("last_date DESC").
		Limit(p.Limit).Offset(p.Offset)

	if err := query.Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = total
	c.JSON(http.StatusOK, result)

}
