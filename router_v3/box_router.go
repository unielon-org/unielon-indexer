package router_v3

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

func (r *Router) BoxCollect(c *gin.Context) {

	type params struct {
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
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

	exc, total, err := r.mysql.FindBoxCollect(p.Tick0, p.Tick1, p.HolderAddress, p.Limit, p.OffSet)

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
	result.Data = exc
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) BoxInfo(c *gin.Context) {

	type params struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		Tick0         string `json:"tick0"`
		Tick1         string `json:"tick1"`
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

	exInfos, total, err := r.mysql.FindBoxInfo(p.OrderId, p.Op, p.Tick0, p.Tick1, p.HolderAddress, p.Limit, p.OffSet)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Code = 200
	result.Msg = "success"
	result.Data = exInfos
	result.Total = total
	c.JSON(http.StatusOK, result)
}
