package router

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

func (r *Router) ExchangeCollect(c *gin.Context) {

	type params struct {
		ExId          string `json:"exid"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		NotDone       int64  `json:"not_done"`
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

	exc, total, err := r.dbc.FindExchangeCollect(p.ExId, p.Tick0, p.Tick1, p.HolderAddress, p.NotDone, p.Limit, p.OffSet)

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

func (r *Router) ExchangeInfo(c *gin.Context) {

	type params struct {
		OrderId       string `json:"order_id"`
		ExId          string `json:"exid"`
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

	exInfos, total, err := r.dbc.FindExchangeInfo(p.OrderId, p.Op, p.ExId, p.Tick, p.Tick0, p.Tick1, p.HolderAddress, p.Limit, p.OffSet)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Code = 200
	result.Msg = "success"
	result.Data = exInfos
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) ExchangeInfoByTick(c *gin.Context) {

	type params struct {
		Op            string `json:"op"`
		Tick          string `json:"tick"`
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

	exInfos, total, err := r.dbc.FindExchangeInfoByTick(p.Op, p.Tick, p.HolderAddress, p.Limit, p.OffSet)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Code = 200
	result.Msg = "success"
	result.Data = exInfos
	result.Total = total
	c.JSON(http.StatusOK, result)
}
