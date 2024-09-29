package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
)

var (
	cacheDrc20CollectAll *models.Drc20CollectCache
)

type Drc20Router struct {
	dbc   *storage.DBClient
	node  *rpcclient.Client
	ipfs  *shell.Shell
	level *storage.LevelDB

	verify *verifys.Verifys
}

func NewDrc20Router(db *storage.DBClient, node *rpcclient.Client, level *storage.LevelDB, ipfs *shell.Shell, verify *verifys.Verifys) *Drc20Router {
	return &Drc20Router{
		dbc:    db,
		node:   node,
		level:  level,
		ipfs:   ipfs,
		verify: verify,
	}
}

func (r *Drc20Router) Order(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
		ToAddress     string `json:"to_address"`
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
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.Drc20Info{
		OrderId:       params.OrderId,
		Op:            params.Op,
		HolderAddress: params.HolderAddress,
		ToAddress:     params.ToAddress,
		BlockNumber:   params.BlockNumber,
	}

	infos := make([]*models.Drc20Info, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.Drc20Info{}).Where(filter).Count(&total).Limit(params.Limit).Offset(params.OffSet).Find(&infos).Error
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

func (r *Drc20Router) CollectAddress(c *gin.Context) {
	params := &struct {
		Tick          string `json:"tick"`
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
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.Drc20CollectAddress{
		HolderAddress: params.HolderAddress,
		Tick:          params.Tick,
	}

	var results []*models.Drc20CollectAddress
	var total int64
	err := r.dbc.DB.Where(filter).Where("amt_sum != '0'").Limit(params.Limit).Offset(params.OffSet).Find(&results).Limit(-1).Offset(-1).Count(&total).Error
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
	result.Data = results
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Drc20Router) Collect(c *gin.Context) {
	params := &struct {
		Tick string `json:"tick"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	maxHeight := 0
	err := r.dbc.DB.Model(&models.Block{}).Select("max(block_number)").Scan(&maxHeight).Error

	if params.Tick == "" {
		if cacheDrc20CollectAll != nil && cacheDrc20CollectAll.CacheNumber == int64(maxHeight) {
			result := &utils.HttpResult{}
			result.Code = 200
			result.Msg = "success"
			result.Data = cacheDrc20CollectAll.Results
			result.Total = cacheDrc20CollectAll.Total
			c.JSON(http.StatusOK, result)
			return
		}
	}

	results := make([]*models.Drc20CollectRouter, 0)
	subQuery := r.dbc.DB.Table("drc20_collect_address AS ci").
		Select(`di.tick, di.amt_sum as mint_amt, di.max_ as max_amt, di.lim_, di.transactions, di.holder_address as deploy_by,
	        di.update_date AS last_mint_time, COUNT(ci.tick = di.tick) AS holders,
			di.create_date AS deploy_time, di.tx_hash as inscription, di.logo, di.introduction, di.is_check`).
		Joins("RIGHT JOIN drc20_collect AS di ON ci.tick = di.tick")

	if params.Tick != "" {
		subQuery = subQuery.Where("di.tick = ?", params.Tick)
	}

	err = subQuery.Group(`di.tick, di.amt_sum, di.max_, di.lim_, di.transactions, di.create_date, di.tx_hash, di.logo,
	       di.introduction, di.is_check, di.holder_address, di.update_date`).
		Order("deploy_time DESC").
		Scan(&results).Error

	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	total := int64(0)
	if params.Tick != "" {
		total = int64(len(results))
	} else {
		r.dbc.DB.Model(&models.Drc20Collect{}).Count(&total)
		cacheDrc20CollectAll = &models.Drc20CollectCache{
			Results:     results,
			Total:       total,
			CacheNumber: int64(maxHeight),
		}
	}

	for _, result := range results {
		if result.IsCheck == 0 {
			de := ""
			result.Logo = &de
			result.Introduction = &de
			result.WhitePaper = &de
			result.Official = &de
			result.Telegram = &de
			result.Discorad = &de
			result.Twitter = &de
			result.Facebook = &de
			result.Github = &de
		}
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = total

	c.JSON(http.StatusOK, result)
}
