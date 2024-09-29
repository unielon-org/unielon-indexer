package router

import (
	"fmt"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
	"strings"
)

type FileExchangeRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client
	ipfs *shell.Shell

	verify *verifys.Verifys
}

func NewFileExchangeRouter(db *storage.DBClient, node *rpcclient.Client, ipfs *shell.Shell, verify *verifys.Verifys) *FileExchangeRouter {
	return &FileExchangeRouter{
		dbc:    db,
		node:   node,
		ipfs:   ipfs,
		verify: verify,
	}
}

func (r *FileExchangeRouter) Order(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		MetaId        string `json:"meta_id"`
		FileId        string `json:"file_id"`
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
		BlockNumber   int64  `json:"block_number"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var nfts []*models.FileExchangeInfo
	var total int64
	subQuery := r.dbc.DB.Table("file_exchange_info fei").
		Select("fei.*, fca.file_path, fmi.name as file_name, fm.name as meta_name").
		Joins("LEFT JOIN file_collect_address fca ON fei.file_id = fca.file_id").
		Joins("LEFT JOIN file_meta_inscription fmi ON fei.file_id = fmi.file_id").
		Joins("LEFT JOIN file_meta fm ON fm.meta_id = fmi.meta_id")

	if params.OrderId != "" {
		subQuery = subQuery.Where("fei.order_id = ?", params.OrderId)
	}

	if params.FileId != "" {
		subQuery = subQuery.Where("fei.file_id = ?", params.FileId)
	}

	if params.MetaId != "" {
		subQuery = subQuery.Where("fmi.meta_id = ?", params.MetaId)
	}

	if params.Op != "" {
		ops := strings.Split(params.Op, ",")
		subQuery = subQuery.Where("fei.op in ?", ops)
	}

	if params.HolderAddress != "" {
		subQuery = subQuery.Where("fei.holder_address = ?", params.HolderAddress)
	}

	if params.BlockNumber != 0 {
		subQuery = subQuery.Where("fei.block_number = ?", params.BlockNumber)
	}

	err := subQuery.Count(&total).
		Limit(params.Limit).
		Offset(params.OffSet).
		Order("fei.create_date DESC").
		Scan(&nfts).Error

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
	result.Data = nfts
	result.Total = total

	c.JSON(http.StatusOK, result)

}

func (r *FileExchangeRouter) Activity(c *gin.Context) {
	params := &struct {
		Op            string `json:"op"`
		MetaId        string `json:"meta_id"`
		FileId        string `json:"file_id"`
		HolderAddress string `json:"holder_address"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	type QueryResult struct {
		Op              string           `gorm:"column:op" json:"op"`
		OrderId         string           `gorm:"column:order_id" json:"order_id"`
		ExId            string           `gorm:"column:ex_id" json:"ex_id"`
		FileId          string           `gorm:"column:file_id" json:"file_id"`
		FilePath        string           `gorm:"column:file_path" json:"file_path"`
		Tick            string           `gorm:"column:tick" json:"tick"`
		Amt             *models.Number   `gorm:"column:amt" json:"amt"`
		HolderAddress   string           `gorm:"column:holder_address" json:"holder_address"`
		ReservesAddress string           `gorm:"column:reserves_address" json:"reserves_address"`
		TxHash          string           `gorm:"column:tx_hash" json:"tx_hash"`
		BlockNumber     int64            `gorm:"column:block_number" json:"block_number"`
		BlockHash       string           `gorm:"column:block_hash" json:"block_hash"`
		CreateDate      models.LocalTime `gorm:"column:create_date" json:"create_date"`
		FileName        string           `gorm:"column:file_name" json:"file_name"`
		MetaName        string           `gorm:"column:meta_name" json:"meta_name"`
	}

	var results []*QueryResult
	subQuery := r.dbc.DB.Table("file_exchange_info fei").
		Select("fei.op, fei.order_id, fei.ex_id, fei.file_id, fei.tick, fei.amt, fei.holder_address, fei.create_date, fei.tx_hash, fei.block_number, fei.block_hash,  fca.file_path, fmi.name as file_name, fm.name as meta_name, fec.reserves_address").
		Joins("LEFT JOIN file_meta_inscription fmi ON fei.file_id = fmi.file_id").
		Joins("LEFT JOIN file_meta fm ON fm.meta_id = fmi.meta_id").
		Joins("LEFT JOIN file_collect_address fca ON fca.file_id = fei.file_id").
		Joins("LEFT JOIN file_exchange_collect fec ON fec.ex_id = fei.ex_id")

	if params.MetaId != "" {
		subQuery.Where("fm.meta_id = ? ", params.MetaId)
	}

	if params.FileId != "" {
		subQuery.Where("fei.file_id = ? ", params.FileId)
	}

	if params.HolderAddress != "" {
		subQuery.Where("fei.holder_address = ?", params.HolderAddress)
	}

	if params.Op != "" {
		ops := strings.Split(params.Op, ",")
		subQuery.Where("fei.op in ?", ops)
	}

	err := subQuery.Limit(params.Limit).Offset(params.OffSet).Order("fei.create_date DESC").
		Scan(&results).Error

	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	c.JSON(http.StatusOK, result)

}

func (r *FileExchangeRouter) Collect(c *gin.Context) {
	params := &struct {
		Address string `json:"holder_address"`
		Limit   int    `json:"limit"`
		OffSet  int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var nfts []*models.FileExchangeCollect
	var total int64
	err := r.dbc.DB.Where("holder_address = ?", params.Address).Limit(params.Limit).Offset(params.OffSet).Find(&nfts).Limit(-1).Offset(-1).Count(&total).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = nfts
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *FileExchangeRouter) SummaryAll(c *gin.Context) {
	type params struct {
		MetaId string `json:"meta_id"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
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

	type FileMetaSummary struct {
		Name        string  `gorm:"column:name" json:"name"`
		MetaId      string  `gorm:"column:meta_id" json:"meta_id"`
		Description string  `gorm:"column:description" json:"description"`
		Icon        string  `gorm:"column:icon" json:"icon"`
		LowestAsk   float64 `gorm:"column:lowest_ask" json:"floor_price"`
		Volume      float64 `gorm:"column:base_volume" json:"volume" `
		DogeUsdt    float64 `gorm:"column:doge_usdt" json:"doge_usdt"`
		Total       int     `gorm:"column:total" json:"listed"`
		Count       int     `gorm:"column:count" json:"supply"`
		HolderCount int     `gorm:"column:holder_count" json:"holders"`
		IsCheck     int     `gorm:"column:is_check" json:"is_check"`
	}

	var results []FileMetaSummary

	subQuery1 := `SELECT COUNT(fec.ex_id) FROM (select ex_id, file_id from file_exchange_collect where amt != amt_finish) fec LEFT JOIN file_meta_inscription fmi ON fec.file_id = fmi.file_id WHERE fmi.meta_id = fm.meta_id`
	subQuery2 := `SELECT COUNT(file_meta_inscription.file_id) FROM file_meta_inscription WHERE file_meta_inscription.meta_id = fm.meta_id`
	subQuery3 := `SELECT COUNT(fca.holder_address) FROM file_meta_inscription left join file_collect_address fca on file_meta_inscription.file_id = fca.file_id where file_meta_inscription.meta_id = fm.meta_id`

	subQuery := r.dbc.DB.Table("file_meta fm").
		Select("fm.name,fm.meta_id, fm.description, fm.icon, fes.lowest_ask, fes.base_volume, fes.doge_usdt, (" + subQuery1 + ") AS total, (" + subQuery2 + ") AS count, (" + subQuery3 + ") AS holder_count, fm.is_check").
		Joins("left join file_exchange_summary fes on fm.meta_id = fes.meta_id")

	//subQuery.Where("fm.is_check = 1")

	if p.MetaId != "" {
		subQuery = subQuery.Where("fm.meta_id = ? ", p.MetaId)
	}

	err := subQuery.Order("base_volume DESC").
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var totalCount int64
	err = r.dbc.DB.Model(&models.FileMeta{}).Count(&totalCount).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = totalCount
	c.JSON(http.StatusOK, result)
}

func (r *FileExchangeRouter) SummaryNftAll(c *gin.Context) {
	type params struct {
		MetaName string `json:"meta_name"`
		Limit    int    `json:"limit"`
		OffSet   int    `json:"offset"`
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

	type NftMetaSummary struct {
		Name        string  `gorm:"column:name" json:"name"`
		Description string  `gorm:"column:description" json:"description"`
		Icon        string  `gorm:"column:icon" json:"icon"`
		LowestAsk   float64 `gorm:"column:lowest_ask" json:"floor_price"`
		Volume      float64 `gorm:"column:base_volume" json:"volume" `
		DogeUsdt    float64 `gorm:"column:doge_usdt" json:"doge_usdt"`
		Total       int     `gorm:"column:total" json:"listed"`
		Count       int     `gorm:"column:count" json:"supply"`
		HolderCount int     `gorm:"column:holder_count" json:"holders"`
	}

	var results []NftMetaSummary

	subQuery1 := `SELECT COUNT(fec.ex_id) FROM file_exchange_collect fec LEFT JOIN file_meta_inscription fmi ON fec.file_id = fmi.file_id WHERE fmi.meta_name = fm.name`
	subQuery2 := `SELECT COUNT(file_meta_inscription.file_id) FROM file_meta_inscription WHERE file_meta_inscription.meta_name = fm.name`
	subQuery3 := `SELECT COUNT(fca.holder_address) FROM file_meta_inscription left join file_collect_address fca on file_meta_inscription.file_id = fca.file_id where file_meta_inscription.meta_name = fm.name`

	subQuery := r.dbc.DB.Table("file_meta fm").
		Select("fm.name, fm.description, fm.icon, fes.lowest_ask, fes.base_volume, fes.doge_usdt, (" + subQuery1 + ") AS total, (" + subQuery2 + ") AS count, (" + subQuery3 + ") AS holder_count").
		Joins("left join file_exchange_summary fes on fm.name = fes.meta_name")

	if p.MetaName != "" {
		subQuery = subQuery.Where("fm.name = ?", p.MetaName)
	}

	err := subQuery.Order("base_volume DESC").
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var totalCount int64
	err = r.dbc.DB.Model(&models.FileMeta{}).Count(&totalCount).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = totalCount
	c.JSON(http.StatusOK, result)
}

func (r *FileExchangeRouter) Inscriptions(c *gin.Context) {

	params := &struct {
		MetaId        string              `json:"meta_id"`
		Attributes    map[string][]string `json:"attributes"`
		Listed        bool                `json:"listed"`
		Tick          string              `json:"tick"`
		PriceOrder    string              `json:"price_order"`
		HolderAddress string              `json:"holder_address"`
		Limit         int                 `json:"limit"`
		OffSet        int                 `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	type QueryResult struct {
		FileID             string         `gorm:"column:file_id" json:"file_id"`
		MetaName           string         `gorm:"column:meta_name" json:"meta_name"`
		FileName           string         `gorm:"column:file_name" json:"file_name"`
		ExID               string         `gorm:"column:ex_id" json:"ex_id"`
		Tick               string         `gorm:"column:tick" json:"tick"`
		Amt                *models.Number `gorm:"column:amt" json:"amt"`
		FilePath           string         `gorm:"column:file_path" json:"file_path"`
		FileHolder         string         `gorm:"column:file_holder" json:"file_holder"`
		FileExchangeHolder string         `gorm:"column:file_exchange_holder" json:"file_exchange_holder"`
	}

	var results []QueryResult
	var total int64

	subQuery := r.dbc.DB.Table("file_meta_attribute").
		Select("file_id, meta_id, name")

	if params.MetaId != "" {
		subQuery.Where("meta_id = ?", params.MetaId)
	}

	temp := true
	orQuery := ""
	for key, value := range params.Attributes {
		if len(value) == 0 {
			continue
		}

		if temp {
			orQuery += fmt.Sprintf(" (trait_type = '%s' AND value IN ('%s'))", key, strings.Join(value, "','"))
			temp = false
			continue
		}

		orQuery += fmt.Sprintf(" OR (trait_type = '%s' AND value IN ('%s'))", key, strings.Join(value, "','"))
	}

	if !temp {
		subQuery.Where(orQuery)
	}

	subQuery.Group("file_id, meta_id, name")

	subQuery1 := r.dbc.DB.Table("(?) as arr", subQuery).
		Select("arr.file_id, fm.name as meta_name, arr.name as file_name, fec.ex_id, fec.tick, fec.amt, fec.file_exchange_holder, fmi.file_path, fmi.holder_address as file_holder").
		Joins("left join (select ex_id, tick, file_id, amt, holder_address as file_exchange_holder from file_exchange_collect where amt != amt_finish)  fec on arr.file_id = fec.file_id").
		Joins("left join file_collect_address fmi on arr.file_id = fmi.file_id").
		Joins("left join file_meta fm on arr.meta_id = fm.meta_id")

	if params.Listed {
		subQuery1 = subQuery1.Where("fec.ex_id IS NOT NULL")
	}

	if params.HolderAddress != "" {
		subQuery1 = subQuery1.Where("fmi.holder_address = ?", params.HolderAddress)
	}

	if params.Tick != "" {
		subQuery1 = subQuery1.Where("fec.tick = ?", params.Tick)
	}

	if params.PriceOrder == "asc" {
		subQuery1 = subQuery1.Order("fec.amt ASC")
	} else {
		subQuery1 = subQuery1.Order("fec.amt DESC")
	}

	err := subQuery1.Count(&total).Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	resultr := &utils.HttpResult{}
	resultr.Code = 200
	resultr.Msg = "success"
	resultr.Data = results
	resultr.Total = total

	c.JSON(http.StatusOK, resultr)
}
