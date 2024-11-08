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

type CrossRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewCrossRouter(dbc *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *CrossRouter {
	return &CrossRouter{
		dbc:    dbc,
		node:   node,
		verify: verify,
	}
}

// CrossOrder
func (r *CrossRouter) Order(c *gin.Context) {
	type params struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
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

	filter := &models.CrossInfo{
		OrderId: p.OrderId,
		Op:      p.Op,
		Tick:    p.Tick0,
	}

	infos := make([]*models.CrossInfo, 0)
	total := int64(0)
	err := r.dbc.DB.Where(filter).Order("id desc").Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&infos).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *CrossRouter) Collect(c *gin.Context) {
	type params struct {
		OrderId string `json:"order_id"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	info := &models.CrossInfo{
		OrderId: p.OrderId,
	}

	err := r.dbc.DB.Where(info).First(info).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Data = info
	c.JSON(http.StatusOK, result)
}
