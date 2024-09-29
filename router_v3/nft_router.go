package router_v3

import (
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

func (r *Router) FindNftAll(c *gin.Context) {

	cards, total, err := r.mysql.FindNftAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) FindNftByTick(c *gin.Context) {

	params := &struct {
		Tick string `json:"tick"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	nfts, err := r.mysql.FindNftCollectAllByTick(params.Tick)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = nfts
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindNftByTickAndId(c *gin.Context) {

	params := &struct {
		Tick   string `json:"tick"`
		TickId int64  `json:"tick_id"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	nfts, err := r.mysql.FindNftCollectAllByTickAndId(params.Tick, params.TickId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = nfts
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindNftHolders(c *gin.Context) {
	type params struct {
		Tick   string `json:"tick"`
		Limit  int64  `json:"limit"`
		OffSet int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if p.Limit > 50 {
		p.Limit = 50
	}

	cards, total, err := r.mysql.FindNftHoldersByTick(p.Tick, p.Limit, p.OffSet)
	if err != nil {
		log.Error("Router", "FindNftHoldersByTick", fmt.Sprintf("mysql.FindNftHoldersByTick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) NftInfoById(c *gin.Context) {

	params := &struct {
		OrderId string `json:"order_id"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	swapInfo, err := r.mysql.FindNftInfoById(params.OrderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result.Data = swapInfo
	c.JSON(http.StatusOK, result)
}

func (r *Router) NftInfo(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
		Limit         int64  `json:"limit"`
		OffSet        int64  `json:"offset"`
	}{Limit: 10, OffSet: 0}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	wdogeInfos, total, err := r.mysql.FindNftInfo(params.OrderId, params.Op, params.HolderAddress, params.Limit, params.OffSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result.Data = wdogeInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) FindNftByAddress(c *gin.Context) {
	type params struct {
		Tick          string `json:"tick"`
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
		c.JSON(http.StatusOK, result)
		return
	}

	nacs, total, err := r.mysql.FindNftByAddressTick(p.HolderAddress, p.Tick, p.Limit, p.OffSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = nacs
	result.Total = total
	c.JSON(http.StatusOK, result)
}
