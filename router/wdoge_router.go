package router

import (
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

func (r *Router) WDogeInfo(c *gin.Context) {
	params := &struct {
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
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

	swapInfos, total, err := r.dbc.FindWDogeInfo(params.Op, params.HolderAddress)
	if err != nil {
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result.Data = swapInfos
	result.Total = total

	c.JSON(http.StatusOK, result)
}
