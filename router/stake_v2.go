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

type StakeV2Router struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}

func NewStakeV2Router(dbc *storage.DBClient, node *rpcclient.Client, verify *verifys.Verifys) *StakeV2Router {
	return &StakeV2Router{
		dbc:    dbc,
		node:   node,
		verify: verify,
	}
}

func (s *StakeV2Router) Order(c *gin.Context) {
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

	if err := c.ShouldBindJSON(p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.StakeV2Info{
		OrderId: p.OrderId,
	}

	infos := make([]*models.StakeV2Info, 0)
	total := int64(0)
	err := s.dbc.DB.Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&infos).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)
}

// Collect
func (s *StakeV2Router) Collect(c *gin.Context) {
	type params struct {
		OrderId string `json:"order_id"`
		Limit   int    `json:"limit"`
		OffSet  int    `json:"offset"`
	}

	p := &params{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.StakeV2Collect{}

	infos := make([]*models.StakeV2Collect, 0)
	total := int64(0)
	err := s.dbc.DB.Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&infos).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)

}

// CollectAddress
func (s *StakeV2Router) CollectAddress(c *gin.Context) {
	type params struct {
		OrderId string `json:"order_id"`
		Limit   int    `json:"limit"`
		OffSet  int    `json:"offset"`
	}

	p := &params{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	filter := &models.StakeV2CollectAddress{}

	infos := make([]*models.StakeV2CollectAddress, 0)
	total := int64(0)
	err := s.dbc.DB.Where(filter).Count(&total).Limit(p.Limit).Offset(p.OffSet).Find(&infos).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Data = infos
	result.Total = total
	c.JSON(http.StatusOK, result)
}

// reward
func (s *StakeV2Router) Reward(c *gin.Context) {

	type params struct {
		HolderAddress string `json:"holder_address"`
		StakeId       string `json:"stake_id"`
		BlockNumber   int64  `json:"block_number"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	reward, err := s.dbc.StakeGetRewardV2(p.HolderAddress, p.StakeId, p.BlockNumber)

	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Data = reward
	c.JSON(http.StatusOK, result)
}
