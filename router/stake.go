package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
)

type StakeRouter struct {
	dbc    *storage.DBClient
	node   *rpcclient.Client
	verify *verifys.Verifys
}

func NewStakeRouter(db *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *StakeRouter {
	return &StakeRouter{
		dbc:    db,
		node:   node,
		verify: verify,
	}
}

func (r *StakeRouter) Order(c *gin.Context) {
	type params struct {
		Tick          string `json:"tick"`
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
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.StakeInfo{
		Tick:          p.Tick,
		HolderAddress: p.HolderAddress,
		BlockNumber:   p.BlockNumber,
	}

	stakeInfos := make([]*models.StakeInfo, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.StakeInfo{}).Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&stakeInfos).Error
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
	result.Total = total
	result.Data = stakeInfos
	c.JSON(http.StatusOK, result)

}

func (r *StakeRouter) Collect(c *gin.Context) {

	type params struct {
		Tick   string `json:"tick"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
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

	stakecs := make([]*models.StakeCollect, 0)
	total := int64(0)

	filter := &models.StakeInfo{
		Tick: p.Tick,
	}

	err := r.dbc.DB.Model(&models.StakeCollect{}).Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&stakecs).Error
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
	result.Data = stakecs
	result.Total = total
	c.JSON(http.StatusOK, result)

}

func (r *StakeRouter) Reward(c *gin.Context) {
	params := &struct {
		HolderAddress string `json:"holder_address"`
		Tick          string `json:"tick"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	tx := r.dbc.DB.Begin()
	staker, err := r.dbc.StakeGetRewardV1(tx, params.HolderAddress, params.Tick)
	if err != nil {
		tx.Rollback()
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	tx.Commit()
	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = staker
	c.JSON(http.StatusOK, result)

}

func (r *StakeRouter) CollectAddress(c *gin.Context) {
	type params struct {
		Tick          string `json:"tick"`
		HolderAddress string `json:"holder_address"`
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

	filter := &models.StakeCollectAddress{
		Tick:          p.Tick,
		HolderAddress: p.HolderAddress,
	}

	stakecs := make([]*models.StakeCollectAddress, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.StakeCollectAddress{}).Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&stakecs).Error
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
	result.Data = stakecs
	result.Total = total
	c.JSON(http.StatusOK, result)

}

func (r *StakeRouter) Total(c *gin.Context) {

	type CollectInfo struct {
		Tick    string `json:"tick"`
		Amt     int64  `json:"amt"`
		Reward  int64  `json:"reward"`
		Holders int    `json:"holders"`
	}

	var results []CollectInfo
	var total int64

	r.dbc.DB.Table("stake_collect AS ci").
		Select("ci.tick, ci.amt, ci.reward, COUNT(di.tick) AS count").
		Joins("LEFT JOIN stake_collect_address AS di ON ci.tick = di.tick").
		Group("ci.tick, ci.amt, ci.reward").
		Count(&total).
		Scan(&results)

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = total

	c.JSON(http.StatusOK, result)
}
