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

type WdogeRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewWdogeRouter(db *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *WdogeRouter {
	return &WdogeRouter{
		dbc:    db,
		node:   node,
		verify: verify,
	}
}

func (r *WdogeRouter) Order(c *gin.Context) {
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
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.WDogeInfo{
		OrderId:       params.OrderId,
		Op:            params.Op,
		HolderAddress: params.HolderAddress,
		BlockNumber:   params.BlockNumber,
	}

	infos := make([]*models.WDogeInfo, 0)
	total := int64(0)

	err := r.dbc.DB.Model(&models.WDogeInfo{}).Where(filter).Count(&total).Limit(params.Limit).Offset(params.OffSet).Find(&infos).Error
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
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)

}
