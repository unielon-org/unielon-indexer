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

type BoxRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewBoxRouter(db *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *BoxRouter {
	return &BoxRouter{
		dbc:    db,
		node:   node,
		verify: verify,
	}
}

func (r *BoxRouter) Order(c *gin.Context) {

	type params struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
		HolderAddress string `json:"holder_address"`
		TxHash        string `json:"tx_hash"`
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

	filter := &models.BoxInfo{
		OrderId:       p.OrderId,
		Op:            p.Op,
		Tick0:         p.Tick0,
		Tick1:         p.Tick1,
		HolderAddress: p.HolderAddress,
		TxHash:        p.TxHash,
		BlockNumber:   p.BlockNumber,
	}

	infos := make([]*models.BoxInfo, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.BoxInfo{}).Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&infos).Error
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
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *BoxRouter) Collect(c *gin.Context) {

	type params struct {
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
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

	filter := &models.BoxCollect{
		Tick0:         p.Tick0,
		Tick1:         p.Tick1,
		HolderAddress: p.HolderAddress,
		IsDel:         0,
	}

	excs := make([]*models.BoxCollect, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.BoxCollect{}).Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Order("update_date desc").Find(&excs).Error
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
	result.Data = excs
	result.Total = total
	c.JSON(http.StatusOK, result)
}
