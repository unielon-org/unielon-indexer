package router_v3

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
	"strings"
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

	exc, total, err := r.mysql.FindExchangeCollect(p.ExId, p.Tick0, p.Tick1, p.HolderAddress, p.NotDone, p.Limit, p.OffSet)

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

	exInfos, total, err := r.mysql.FindExchangeInfo(p.OrderId, p.Op, p.ExId, p.Tick, p.Tick0, p.Tick1, p.HolderAddress, p.Limit, p.OffSet)
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

	exInfos, total, err := r.mysql.FindExchangeInfoByTick(p.Op, p.Tick, p.HolderAddress, p.Limit, p.OffSet)
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

func (r *Router) ExchangeSummary(c *gin.Context) {

	result := &utils.HttpResult{}
	summary, err := r.mysql.FindExchangeSummary()
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	result.Code = 200
	result.Msg = "success"
	result.Data = summary
	c.JSON(http.StatusOK, result)
}

func (r *Router) ExchangeSummaryAll(c *gin.Context) {

	result := &utils.HttpResult{}
	summaryall, err := r.mysql.FindExchangeSummaryAll()
	if err != nil {
		return
	}
	result.Code = 200
	result.Msg = "success"
	result.Data = summaryall

	c.JSON(http.StatusOK, result)
}

func (r *Router) ExchangeSummaryByTick(c *gin.Context) {

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

	summary, err := r.mysql.FindExchangeSummaryByTick(p.Tick)
	if err != nil {
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = summary

	c.JSON(http.StatusOK, result)
}

func (r *Router) ExchangeSummaryK(c *gin.Context) {
	type params struct {
		Tick0        string `json:"tick0"`
		Tick1        string `json:"tick1"`
		DateInterval string `json:"interval"`
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
	p.Tick0, p.Tick1, _, _, _, _ = utils.SortTokens(p.Tick0, p.Tick1, nil, nil, nil, nil)

	results, err := r.mysql.FindExchangeSummaryK(p.Tick0, p.Tick1, p.DateInterval)
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
