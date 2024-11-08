package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type SwapRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewSwapRouter(db *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *SwapRouter {
	return &SwapRouter{
		dbc:    db,
		node:   node,
		verify: verify,
	}
}

func (r *SwapRouter) Order(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		Tick          string `json:"tick"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		BlockNumber   int64  `json:"block_number"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	filter := &models.SwapInfo{
		OrderId:       params.OrderId,
		Op:            params.Op,
		Tick:          params.Tick,
		Tick0:         params.Tick0,
		Tick1:         params.Tick1,
		HolderAddress: params.HolderAddress,
		BlockNumber:   params.BlockNumber,
	}

	infos := make([]*models.SwapInfo, 0)
	total := int64(0)
	err := r.dbc.DB.Where(filter).Limit(params.Limit).Offset(params.OffSet).Find(&infos).Limit(-1).Offset(-1).Count(&total).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = infos
	result.Total = total

	c.JSON(http.StatusOK, result)

}

func (r *SwapRouter) SwapLiquidity(c *gin.Context) {
	params := &struct {
		Tick0  string `json:"tick0"`
		Tick1  string `json:"tick1"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}{
		Limit:  -1,
		OffSet: -1,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	filter := &models.SwapLiquidity{
		Tick0: params.Tick0,
		Tick1: params.Tick1,
	}

	infos := make([]*models.SwapLiquidity, 0)
	total := int64(0)
	err := r.dbc.DB.Where(filter).Limit(params.Limit).Offset(params.OffSet).Find(&infos).Limit(-1).Offset(-1).Count(&total).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = infos
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *SwapRouter) SwapLiquidityHolder(c *gin.Context) {
	params := &struct {
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	params.Tick0, params.Tick1, _, _, _, _ = utils.SortTokens(params.Tick0, params.Tick1, nil, nil, nil, nil)

	type QueryResult struct {
		Liquidity      *models.Number `gorm:"column:amt_sum"`
		LiquidityTotal *models.Number `gorm:"column:liquidity_total"`
		Reserve0       *models.Number `gorm:"column:amt0"`
		Reserve1       *models.Number `gorm:"column:amt1"`
		Price          *big.Float     `gorm:"-" json:"price"`
		Tick0          string         `json:"tick0"`
		Tick1          string         `json:"tick1"`
		Tick           string         `json:"tick"`
	}

	var results []QueryResult
	var dbModels []*QueryResult
	var total int64

	tick := params.Tick0 + "-SWAP-" + params.Tick1
	subQuery := r.dbc.DB.Table("drc20_collect_address dca").
		Select("dca.tick, dca.amt_sum, sl.liquidity_total, sl.amt0, sl.amt1, sl.tick0, sl.tick1").
		Joins("left join swap_liquidity sl on sl.tick = dca.tick")

	if params.HolderAddress != "" {
		subQuery = subQuery.Where("dca.holder_address = ?", params.HolderAddress)
	}

	if tick != "-SWAP-" {
		subQuery = subQuery.Where("dca.tick = ?", tick)
	}

	err := subQuery.Where("length(dca.tick) > ? AND dca.amt_sum != ? AND dca.tick != ?", 9, "0", "WDOGE(WRAPPED-DOGE)").
		Count(&total).Limit(params.Limit).Offset(params.OffSet).Scan(&results).Error

	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	for _, res := range results {
		dbModels = append(dbModels, &QueryResult{
			Tick:           res.Tick0 + "-SWAP-" + res.Tick1,
			Tick0:          res.Tick0,
			Tick1:          res.Tick1,
			Liquidity:      res.Liquidity,
			LiquidityTotal: res.LiquidityTotal,
			Price:          new(big.Float).Quo(new(big.Float).SetInt(res.Reserve0.Int()), new(big.Float).SetInt(res.Reserve1.Int())),
			Reserve0:       (*models.Number)(new(big.Int).Div(new(big.Int).Mul(res.Reserve0.Int(), res.Liquidity.Int()), res.LiquidityTotal.Int())),
			Reserve1:       (*models.Number)(new(big.Int).Div(new(big.Int).Mul(res.Reserve1.Int(), res.Liquidity.Int()), res.LiquidityTotal.Int())),
		})

	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = dbModels
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *SwapRouter) TickByAddress(c *gin.Context) {
	params := &struct {
		HolderAddress string `json:"holder_address"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var ticks []string
	err := r.dbc.DB.Table("drc20_collect_address").Where("holder_address = ?", params.HolderAddress).Pluck("tick", &ticks).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = ticks
	result.Total = int64(len(ticks))

	c.JSON(http.StatusOK, result)
}

func (r *SwapRouter) SwapPrice(c *gin.Context) {

	result := &utils.HttpResult{}

	swapInfos, total, err := r.dbc.FindSwapPriceAll()
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Code = 200
	result.Msg = "success"
	result.Data = swapInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *SwapRouter) SwapK(c *gin.Context) {
	type params struct {
		Tick         string `json:"tick"`
		DateInterval string `json:"date_interval"`
		Limit        int    `json:"limit"`
		Offset       int    `json:"offset"`
	}

	p := &params{
		DateInterval: "1d",
		Limit:        1500,
		Offset:       0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	p.DateInterval = strings.ToLower(p.DateInterval)

	results := make([]*models.SwapSummary, 0)
	total := int64(0)
	err := r.dbc.DB.Model(&models.SwapSummary{}).Where("tick = ? and date_interval = ?", p.Tick, p.DateInterval).
		Count(&total).
		Limit(p.Limit).Offset(p.Offset).
		Find(&results).Error

	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	c.JSON(http.StatusOK, result)
}

func (r *SwapRouter) SwapTvl(c *gin.Context) {
	type params struct {
		Tick   string `json:"tick"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}

	p := &params{
		Limit:  -1,
		OffSet: -1,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	ticks := strings.Split(p.Tick, "-SWAP-")

	if len(ticks) == 0 {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = "tick is invalid"
		c.JSON(http.StatusOK, result)
		return
	}

	tick0 := ticks[0]
	tick1 := ticks[0]

	if len(ticks) == 2 {
		tick1 = ticks[1]
	}

	type SwapInfoSummaryResult struct {
		Liquidity  string  `gorm:"column:liquidity" json:"liquidity"`
		BaseVolume string  `gorm:"column:base_volume" json:"base_volume"`
		DogeUsdt   float64 `gorm:"column:doge_usdt" json:"doge_usdt"`
		LastDate   string  `gorm:"column:last_date" json:"last_date"`
	}

	var results []SwapInfoSummaryResult

	total := int64(0)
	subQuery := r.dbc.DB.Table("swap_summary_liquidity").Select("SUM(liquidity * 2) AS liquidity, SUM(base_volume) AS base_volume, doge_usdt, MAX(last_date) AS last_date")
	if tick0 != tick1 {
		subQuery.Where("tick0 = ? AND tick1 = ?", tick0, tick1).
			Group("last_date, doge_usdt")
	} else {
		subQuery.Where("tick0 = ? OR tick1 = ?", tick0, tick1).
			Group("last_date, doge_usdt")
	}

	err := subQuery.Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&results).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Total = total
	result.Data = results
	c.JSON(http.StatusOK, result)
}

func (r *SwapRouter) SwapSummaryTvlTotal(c *gin.Context) {
	type params struct {
		Limit  int `json:"limit"`
		OffSet int `json:"offset"`
	}

	p := &params{
		Limit:  -1,
		OffSet: -1,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	type SwapInfoSummaryResult struct {
		Liquidity  string  `gorm:"column:liquidity" json:"liquidity"`
		BaseVolume string  `gorm:"column:base_volume" json:"base_volume"`
		DogeUsdt   float64 `gorm:"column:doge_usdt" json:"doge_usdt"`
		LastDate   string  `gorm:"column:last_date" json:"last_date"`
	}

	var results []SwapInfoSummaryResult
	var totalCount int64

	// Perform the query using GORM
	err := r.dbc.DB.Table("(SELECT SUM(liquidity) * 2 AS TotalLiquidity, MAX(doge_usdt) AS MaxDogeUsdt, last_date FROM swap_summary_liquidity GROUP BY last_date) A").
		Select("A.TotalLiquidity as liquidity, B.TotalBaseVolume as base_volume, A.MaxDogeUsdt as doge_usdt, A.last_date").
		Joins("JOIN (SELECT SUM(base_volume) AS TotalBaseVolume, last_date FROM swap_summary WHERE date_interval='1d' GROUP BY last_date) B ON A.last_date = B.last_date").
		Count(&totalCount).Limit(p.Limit).Offset(p.OffSet).Scan(&results).Error

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
	result.Data = results
	c.JSON(http.StatusOK, result)
}

type T struct {
}

func (r *SwapRouter) SwapSummary(c *gin.Context) {
	type params struct {
		Tick   string `json:"tick"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}

	p := &params{
		Limit:  -1,
		OffSet: -1,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	type Drc20InfoResult struct {
		Tick       string  `json:"tick"`
		MaxAmt     string  `grom:"max_amt" json:"max_amt"`
		AmtSum     string  `grom:"amt_sum" json:"amt_sum"`
		LastPrice  float64 `gorm:"last_price" json:"last_price"`
		OpenPrice  float64 `gorm:"open_price" json:"open_price"`
		BaseVolume int64   `json:"base_volume"`
		Holders    int     `gorm:"holders" json:"holders"`
		FootPrice  float64 `json:"foot_price"`
		LastDate   string  `json:"last_date"`
		Logo       string  `json:"logo"`
		IsCheck    int     `json:"is_check"`
	}

	const layout = "2006-01-02 15:04:05"
	startDate := time.Now()
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	subQuery0 := r.dbc.DB.Table("swap_summary").
		Select("tick, MAX(id) AS max_id").
		Where("date_interval = '1d' ").
		Group("tick")

	subQuery := r.dbc.DB.Table("swap_summary es").
		Select("es.tick, es.close_price, es.lowest_ask, es.base_volume, es.open_price, es.id, es.last_date").
		Joins("INNER JOIN (?) es_max ON es.id = es_max.max_id", subQuery0)

	var results []Drc20InfoResult
	var totalCount int64
	mainQuery := r.dbc.DB.Table("drc20_collect d20i").
		Select(`
        d20i.tick,
        d20i.amt_sum,
        d20i.max_ as max_amt,
		COALESCE(e.open_price, 0) AS open_price,
        COALESCE(e.close_price, 0) AS last_price,
        COALESCE(CASE WHEN e.last_date != ? THEN 0 ELSE e.base_volume END, 0) AS base_volume,
        (SELECT COUNT(holder_address) FROM drc20_collect_address WHERE drc20_collect_address.tick = d20i.tick) AS holders,
        COALESCE(e.lowest_ask, 0) AS foot_price,
		e.last_date,
        d20i.logo,
        d20i.is_check`, startDate.Format(layout)).
		Joins("LEFT JOIN (?) e ON e.tick = d20i.tick", subQuery).
		Where("LENGTH(d20i.tick) < 9").
		Order("base_volume DESC, holders DESC")

	// Apply pagination
	mainQuery = mainQuery.Limit(p.Limit).Offset(p.OffSet)

	if p.Tick != "" {
		mainQuery = mainQuery.Where("d20i.tick = ?", p.Tick)
	}

	err := mainQuery.Count(&totalCount).Scan(&results).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = totalCount
	c.JSON(http.StatusOK, result)

}

func (r *SwapRouter) SwapPair(c *gin.Context) {
	type params struct {
		Tick   string `json:"tick"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}

	p := &params{
		Limit:  10,
		OffSet: -1,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	type SwapLiquiditySummaryResult struct {
		Tick        string
		Tick0       string
		Tick1       string
		PriceChange float64 `gorm:"column:price_change"`
		Liquidity   float64
		BaseVolume  float64
		DogeUsdt    float64
		Amt0        float64
		Amt1        float64
	}

	// Initialize the slice for the results
	var results []SwapLiquiditySummaryResult

	subQuery := r.dbc.DB.Table("swap_summary_liquidity").Select("tick, MAX(id) AS max_id")

	if p.Tick != "" {
		subQuery = subQuery.Where("tick0 = ? OR tick1 = ?", p.Tick, p.Tick)
	}

	subQuery = subQuery.Group("tick")

	mainQuery := r.dbc.DB.Table("swap_summary_liquidity es").
		Select(`
        es.tick,
        es.tick0,
        es.tick1,
        COALESCE(((es.close_price - es.open_price) / es.open_price) * 100, 0) AS priceChange,
        es.liquidity * 2 AS liquidity,
        es.base_volume,
        es.doge_usdt,
        sl.amt0,
        sl.amt1`).
		Joins("INNER JOIN (?) es_max ON es.id = es_max.max_id", subQuery).
		Joins("LEFT JOIN swap_liquidity sl ON es.tick = sl.tick").
		Order("es.liquidity DESC")

	mainQuery = mainQuery.Limit(p.Limit).Offset(p.OffSet)

	var totalCount int64
	mainQuery.Count(&totalCount).Scan(&results)

	result := &utils.HttpResult{
		Code:  200,
		Msg:   "success",
		Data:  results,
		Total: totalCount,
	}

	c.JSON(http.StatusOK, result)

}
