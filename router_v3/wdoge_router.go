package router_v3

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

// WDogeInfoById
func (r *Router) WDogeInfoById(c *gin.Context) {

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

	swapInfo, err := r.mysql.FindWDogeInfoById(params.OrderId)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = swapInfo

	c.JSON(http.StatusOK, result)
}

func (r *Router) WDogeInfo(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
		Limit         int64  `json:"limit"`
		OffSet        int64  `json:"offset"`
	}{Limit: 10, OffSet: 0}

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

	wdogeInfos, total, err := r.mysql.FindWDogeInfo(params.OrderId, params.Op, params.HolderAddress, params.Limit, params.OffSet)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = wdogeInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}
