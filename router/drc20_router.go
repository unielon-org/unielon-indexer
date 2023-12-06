package router

import (
	"fmt"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

func (r *Router) FindDrc20All(c *gin.Context) {

	params := &utils.Drc20Params{
		Limit:  1500,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	cards, total, err := r.dbc.FindDrc20All(params)
	if err != nil {
		log.Error("Router", "FindDrc20All", fmt.Sprintf("mysql.FindDrc20All is err:%s", err.Error()))
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

func (r *Router) FindDrc20ByTick(c *gin.Context) {

	params := &utils.Drc20Params{}
	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	card, err := r.dbc.FindDrc20ByTick(params.Tick)
	if err != nil {
		log.Error("Router", "FindDrc20ByTick", fmt.Sprintf("mysql.FindDrc20ByTick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = card
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20Holders(c *gin.Context) {
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

	cards, total, err := r.dbc.FindDrc20HoldersByTick(p.Tick, p.Limit, p.OffSet)
	if err != nil {
		log.Error("Router", "FindDrc20Holders", fmt.Sprintf("mysql.FindDrc20HoldersByTick is err:%s", err.Error()))
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

func (r *Router) FindDrc20ByAddress(c *gin.Context) {
	type params struct {
		ReceiveAddress string `json:"receive_address"`
		Limit          int64  `json:"limit"`
		OffSet         int64  `json:"offset"`
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

	cards, total, err := r.dbc.FindDrc20AllByAddress(p.ReceiveAddress, p.Limit, p.OffSet)
	if err != nil {
		log.Error("Router", "FindDrc20sByAddress", fmt.Sprintf("mysql.FindDrc20AllByAddress is err:%s", err.Error()))
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

func (r *Router) FindDrc20ByAddressTick(c *gin.Context) {
	type params struct {
		Tick           string `json:"tick"`
		ReceiveAddress string `json:"receive_address"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	card, err := r.dbc.FindDrc20AllByAddressTick(p.ReceiveAddress, p.Tick)
	if err != nil {
		log.Error("Router", "FindDrc20sByAddress", fmt.Sprintf("mysql.FindDrc20AllByAddressTick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = card
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrders(c *gin.Context) {
	type params struct {
		Address string `json:"address"`
		Tick    string `json:"tick"`
		Hash    string `json:"hash"`
		Number  int64  `json:"number"`
		Limit   int64  `json:"limit"`
		OffSet  int64  `json:"offset"`
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

	orders, total, err := r.dbc.FindOrders(p.Address, p.Tick, p.Hash, p.Number, p.Limit, p.OffSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = orders
	result.Total = total
	c.JSON(http.StatusOK, result)
}
