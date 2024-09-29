package router_v3

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

func (r *Router) StakeAll(c *gin.Context) {
	cards, total, err := r.mysql.FindStakeAll()
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
	result.Data = cards
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) StakeByTick(c *gin.Context) {

	params := &struct {
		Tick string `json:"tick"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	tx, _ := r.mysql.MysqlDB.Begin()
	nfts, err := r.mysql.FindStakeCollectByTick(tx, params.Tick)
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
	result.Data = nfts
	c.JSON(http.StatusOK, result)
}

func (r *Router) StakeReward(c *gin.Context) {

	params := &struct {
		HolderAddress string `json:"holder_address"`
		Tick          string `json:"tick"`
		BlockNumber   int64  `json:"block_number"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	staker, err := r.mysql.StakeGetRewardRouter(params.HolderAddress, params.Tick)
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
	result.Data = staker
	c.JSON(http.StatusOK, result)
}

func (r *Router) StakeHolders(c *gin.Context) {
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

	stakes, total, err := r.mysql.FindStakeByAddressTick("", p.Tick, p.Limit, p.OffSet)
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
	result.Data = stakes
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) StakeByAddressTick(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, result)
		return
	}

	nacs, total, err := r.mysql.FindStakeByAddressTick(p.HolderAddress, p.Tick, p.Limit, p.OffSet)
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
	result.Data = nacs
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) StakeInfoById(c *gin.Context) {
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

	stakeInfo, _, err := r.mysql.FindStakeInfo(params.OrderId, "", "", "", 1, 0)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	if len(stakeInfo) > 0 {
		if stakeInfo[0].Op == "getallreward" {
			info, err := r.mysql.FindStakeRewardInfo(stakeInfo[0].OrderId)
			if err != nil {
				return
			}
			stakeInfo[0].StakeRewardInfos = info
		}

		result.Data = stakeInfo[0]
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) StakeInfo(c *gin.Context) {
	type params struct {
		OrderId       string `json:"order_id"`
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
	result.Code = 200
	result.Msg = "success"

	stakeInfos, total, err := r.mysql.FindStakeInfo(p.OrderId, p.Op, p.Tick, p.HolderAddress, p.Limit, p.OffSet)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = stakeInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}
