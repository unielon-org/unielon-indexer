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

type NftRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewNftRouter(db *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *NftRouter {
	return &NftRouter{
		dbc:    db,
		node:   node,
		verify: verify,
	}
}

func (r *NftRouter) Order(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
		BlockNumber   int64  `json:"block_number"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{Limit: 10, OffSet: 0}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 400
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	filter := &models.NftInfo{
		OrderId:       params.OrderId,
		Op:            params.Op,
		HolderAddress: params.HolderAddress,
		BlockNumber:   params.BlockNumber,
	}

	infos := make([]*models.NftInfo, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.NftInfo{}).Where(filter).Count(&total).Limit(params.Limit).Offset(params.OffSet).Find(&infos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)

}

func (r *NftRouter) Collect(c *gin.Context) {
	params := &struct {
		Tick   string `json:"tick"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	type NftCollectQuery struct {
		Tick         string           `json:"tick"`
		TickSum      int64            `json:"tick_sum"`
		Total        int64            `json:"total"`
		Model        string           `json:"model"`
		Prompt       string           `json:"prompt"`
		ImagePath    string           `json:"image_path"`
		DeployHash   string           `json:"deploy_hash"`
		Transactions int64            `json:"transactions"`
		Holders      int64            `json:"holders"`
		Introduction *string          `json:"introduction"`
		DeployTime   models.LocalTime `gorm:"create_date" json:"deploy_time"`
		IsCheck      uint64           `json:"is_check"`
	}

	var results []models.NftCollect
	var totalCount int64

	subQuery := r.dbc.DB.Table("nft_collect").
		Select(`tick,
				tick_sum,
				total,
				prompt,
				image_path,
				transactions,
				(SELECT COUNT(holder_address) FROM nft_collect_address WHERE nft_collect_address.tick = nft_collect.tick) AS holders,
				create_date,
				deploy_hash,
				introduction,
				is_check`).
		Order("create_date DESC").
		Count(&totalCount).
		Limit(params.Limit).
		Offset(params.OffSet)

	if params.Tick != "" {
		subQuery = subQuery.Where("tick = ?", params.Tick)
	}

	// Query with pagination
	err := subQuery.
		Scan(&results).Error

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
	result.Data = results
	result.Total = totalCount

	c.JSON(http.StatusOK, result)
}

func (r *NftRouter) CollectAddress(c *gin.Context) {
	params := &struct {
		Tick          string `json:"tick"`
		TickId        int64  `json:"tick_id"`
		HolderAddress string `json:"holder_address"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.NftCollectAddress{
		Tick:          params.Tick,
		TickId:        params.TickId,
		HolderAddress: params.HolderAddress,
	}

	var results []models.NftCollectAddress
	var totalCount int64

	// Count total records
	err := r.dbc.DB.Model(&models.NftCollectAddress{}).Count(&totalCount).Where(filter).Limit(params.Limit).Offset(params.OffSet).Find(&results).Error
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
	result.Data = results
	result.Total = totalCount
	c.JSON(http.StatusOK, result)

}
