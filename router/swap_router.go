package router

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"net/http"
)

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

	swapLiquidity, err := r.dbc.FindSwapLiquidityWeb(params.Tick0, params.Tick1)
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
		result.Data = data
		c.JSON(http.StatusOK, result)
		return
	}
	data := make(map[string]interface{})
	data["tick0"] = params.Tick0
	data["tick1"] = params.Tick1
	data["reserve0"] = swapLiquidity.Amt0.String()
	data["reserve1"] = swapLiquidity.Amt1.String()
	data["liquidity_total"] = swapLiquidity.LiquidityTotal.String()
	result.Data = data

	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapGetReservesAll(c *gin.Context) {

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapLiquidity, total, err := r.dbc.FindSwapLiquidityAll()
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

	swapLiquiditys, err := r.dbc.FindSwapLiquidityByHolder(params.HolderAddress, params.Tick0, params.Tick1)
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
		datas = append(datas, data)
	}
	result.Data = datas
	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapInfo(c *gin.Context) {
	type params struct {
		Op            string `json:"op"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		Limit         int64  `json:"limit"`
		OffSet        int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
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

	swapInfos, total, err := r.dbc.FindSwapInfo(p.Op, p.Tick0, p.Tick1, p.HolderAddress, p.Limit, p.OffSet)
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

	swapInfos, total, err := r.dbc.FindSwapPriceAll()
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
