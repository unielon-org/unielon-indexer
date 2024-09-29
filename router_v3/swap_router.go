package router_v3

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"net/http"
	"strings"
)

func (r *Router) SwapGetReservesAll(c *gin.Context) {

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapLiquidity, total, err := r.mysql.FindSwapLiquidityAll()
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = swapLiquidity
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapGetReserves(c *gin.Context) {

	params := &struct {
		Tick0 string `json:"tick0"`
		Tick1 string `json:"tick1"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	params.Tick0, params.Tick1, _, _, _, _ = utils.SortTokens(params.Tick0, params.Tick1, nil, nil, nil, nil)

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapLiquidity, err := r.mysql.FindSwapLiquidityWeb(params.Tick0, params.Tick1)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	if swapLiquidity == nil {
		data := make(map[string]interface{})
		data["tick0"] = params.Tick0
		data["tick1"] = params.Tick1
		data["reserve0"] = new(big.Int).String()
		data["reserve1"] = new(big.Int).String()
		data["liquidity"] = new(big.Int).String()
		data["close_price"] = new(big.Int).String()
		data["volume"] = new(big.Int).String()
		data["total"] = new(big.Int).String()
		result.Data = data
		c.JSON(http.StatusOK, result)
		return
	}

	volume, total, err := r.mysql.FindSwapInfoVolumeByTick(params.Tick0, params.Tick1)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, "error")
		return
	}

	data := make(map[string]interface{})
	data["tick0"] = params.Tick0
	data["tick1"] = params.Tick1
	data["reserve0"] = swapLiquidity.Amt0.String()
	data["reserve1"] = swapLiquidity.Amt1.String()
	data["liquidity_total"] = swapLiquidity.LiquidityTotal.String()
	data["close_price"] = swapLiquidity.ClosePrice
	data["volume"] = volume
	data["total"] = total

	result.Data = data
	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapGetLiquidity(c *gin.Context) {

	params := &struct {
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	params.Tick0, params.Tick1, _, _, _, _ = utils.SortTokens(params.Tick0, params.Tick1, nil, nil, nil, nil)

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapLiquiditys, err := r.mysql.FindSwapLiquidityByHolder(params.HolderAddress, params.Tick0, params.Tick1)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	datas := make([]map[string]interface{}, 0, 0)
	for _, v := range swapLiquiditys {
		data := make(map[string]interface{})
		data["tick0"] = v.Tick0
		data["tick1"] = v.Tick1
		data["reserve0"] = v.Amt0.String()
		data["reserve1"] = v.Amt1.String()
		data["liquidity"] = v.LiquidityTotal.String()
		price := new(big.Float).Quo(new(big.Float).SetInt(v.Amt0.Int()), new(big.Float).SetInt(v.Amt1.Int()))
		data["price"] = price
		datas = append(datas, data)
	}
	result.Data = datas
	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapInfoById(c *gin.Context) {

	params := &struct {
		OrderId string `json:"order_id"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapInfo, err := r.mysql.FindSwapInfoById(params.OrderId)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = swapInfo

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapInfo(c *gin.Context) {
	type params struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		Tick          string `json:"tick"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		Limit         int64  `json:"limit"`
		OffSet        int64  `json:"offset"`
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

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapInfos, total, err := r.mysql.FindSwapInfo(p.OrderId, p.Op, p.Tick, p.Tick0, p.Tick1, p.HolderAddress, p.Limit, p.OffSet)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = swapInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapPrice(c *gin.Context) {

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapInfos, total, err := r.mysql.FindSwapPriceAll()
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = swapInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapSummaryTvlAll(c *gin.Context) {

	results, err := r.mysql.FindCMCSummaryTVLAll()
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

func (r *Router) SwapSummaryTvl(c *gin.Context) {

	type params struct {
		Tick string `json:"tick"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	ticks := strings.Split(p.Tick, "-SWAP-")

	if len(ticks) == 0 {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = " "
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	tick0 := ticks[0]
	tick1 := ticks[0]

	if len(ticks) == 2 {
		tick1 = ticks[1]
	}

	results, err := r.mysql.FindCMCSummaryTVL(tick0, tick1)
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

func (r *Router) SwapSummaryKNew(c *gin.Context) {
	type params struct {
		Tick         string `json:"tick"`
		DateInterval string `json:"date_interval"`
	}

	p := &params{
		DateInterval: "1d",
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	p.DateInterval = strings.ToLower(p.DateInterval)

	resultnew, err := r.mysql.FindCMCSummaryK2(p.Tick, p.DateInterval)
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
	result.Data = resultnew
	c.JSON(http.StatusOK, result)

}

func (r *Router) SwapSummaryAll(c *gin.Context) {

	result := &utils.HttpResult{}
	summaryall, err := r.mysql.FindSwapSummaryAll()
	if err != nil {
		return
	}
	result.Code = 200
	result.Msg = "success"
	result.Data = summaryall

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapSummaryByTick(c *gin.Context) {

	type params struct {
		Tick string `json:"tick"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	summary, err := r.mysql.FindSwapSummaryByTick(p.Tick)
	if err != nil {
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = summary

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapPairByTick(c *gin.Context) {

	type params struct {
		Tick string `json:"tick"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	summary, err := r.mysql.FindSwapPairByTick(p.Tick)
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
	result.Data = summary

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapPairAll(c *gin.Context) {

	summary, err := r.mysql.FindSwapPairAll()
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
	result.Data = summary

	c.JSON(http.StatusOK, result)
}
