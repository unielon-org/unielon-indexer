package router_v3

import (
	"encoding/hex"
	"fmt"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/txscript"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage_v3"
	"github.com/unielon-org/unielon-indexer/utils"
	"net/http"
)

var (
	cacheDrc20 *storage_v3.Drc20CollectAllCache
)

func (r *Router) FindDrc20All(c *gin.Context) {

	maxHeight := int64(0)
	err := r.dbc.DB.Model(&models.Block{}).Select("max(block_number)").Scan(&maxHeight).Error

	if cacheDrc20 != nil && cacheDrc20.CacheNumber == maxHeight {
		result := &utils.HttpResult{}
		result.Code = 200
		result.Msg = "success"
		result.Data = cacheDrc20.Results
		result.Total = cacheDrc20.Total
		c.JSON(http.StatusOK, result)
		return
	}

	cards, total, err := r.mysql.FindDrc20All()
	if err != nil {
		log.Error("Router", "FindDrc20All", fmt.Sprintf("mysql.FindDrc20All is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	result.Total = total

	cacheDrc20 = &storage_v3.Drc20CollectAllCache{
		Results:     cards,
		Total:       total,
		CacheNumber: maxHeight,
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20TickAddress(c *gin.Context) {
	type params struct {
		Address string `json:"address"`
	}

	p := &params{}
	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	cards, err := r.mysql.FindDrc20TickAddress(p.Address)
	if err != nil {
		log.Error("Router", "FindDrc20All", fmt.Sprintf("mysql.FindDrc20TickAddress is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20Popular(c *gin.Context) {
	type params struct {
		ReceiveAddress string `json:"receive_address"`
	}

	var p params
	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	cards, _, err := r.mysql.FindDrc20ByAddressPopular(p.ReceiveAddress)
	if err != nil {
		log.Error("Router", "FindDrc20All", fmt.Sprintf("mysql.FindDrc20All is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20ByTick(c *gin.Context) {

	params := struct {
		Tick string `json:"tick"`
	}{}
	if err := c.ShouldBindJSON(&params); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	card, err := r.mysql.FindDrc20ByTick(params.Tick)
	if err != nil {
		log.Error("Router", "FindDrc20ByTick", fmt.Sprintf("mysql.FindDrc20ByTick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = card
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20Holders(c *gin.Context) {
	type params struct {
		Tick   string `json:"tick"`
		Limit  int64  `json:"limit"`
		OffSet int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if p.Limit > 50 {
		p.Limit = 50
	}

	cards, total, err := r.mysql.FindDrc20HoldersByTick(p.Tick, p.Limit, p.OffSet)
	if err != nil {
		log.Error("Router", "FindDrc20Holders", fmt.Sprintf("mysql.FindDrc20HoldersByTick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20sByAddress(c *gin.Context) {
	type params struct {
		ReceiveAddress string `json:"receive_address"`
		Limit          int64  `json:"limit"`
		OffSet         int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	cards, total, err := r.mysql.FindDrc20AllByAddress(p.ReceiveAddress, p.Limit, p.OffSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = cards
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindDrc20sByAddressTick(c *gin.Context) {
	type params struct {
		Tick           string `json:"tick"`
		ReceiveAddress string `json:"receive_address"`
	}

	p := &params{}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	card, err := r.mysql.FindDrc20AllByAddressTick(p.ReceiveAddress, p.Tick)
	if err != nil {
		log.Error("Router", "FindDrc20sByAddress", fmt.Sprintf("mysql.FindDrc20AllByAddressTick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = card
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrders(c *gin.Context) {
	type params struct {
		OrderId        string `json:"order_id"`
		Tick           string `json:"tick"`
		Op             string `json:"op"`
		ReceiveAddress string `json:"receive_address"`
		Limit          int64  `json:"limit"`
		OffSet         int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if p.Limit > 50 {
		p.Limit = 50
	}

	orders, total, err := r.mysql.FindOrders(p.ReceiveAddress, p.Op, p.Tick, p.Limit, p.OffSet)

	if err != nil {
		log.Error("Router", "FindOrders", fmt.Sprintf("mysql.FindOrders is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = orders
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrderByAddress(c *gin.Context) {
	type params struct {
		OrderId        string `json:"order_id"`
		Op             string `json:"op"`
		ReceiveAddress string `json:"receive_address"`
		Tick           string `json:"tick"`
		Limit          int64  `json:"limit"`
		OffSet         int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if p.Limit > 50 {
		p.Limit = 50
	}

	_, err := btcutil.DecodeAddress(p.ReceiveAddress, &chaincfg.MainNetParams)
	if err != nil {
		log.Error("Router", "FindOrders", fmt.Sprintf("btcutil.DecodeAddress is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	maxHeight := int64(0)
	err = r.dbc.DB.Model(&models.Block{}).Select("max(block_number)").Scan(&maxHeight).Error
	if err != nil {
		log.Error("Router", "FindOrders", fmt.Sprintf("redis.GetFromHeight is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	cacheOrderAddress, _ := r.level.GetCacheOrderAddress(p.ReceiveAddress)
	if cacheOrderAddress != nil && cacheOrderAddress.CacheNumber == maxHeight {
		result := &utils.HttpResult{}
		result.Code = 200
		result.Msg = "success"
		result.Data = cacheOrderAddress.Orders
		result.Total = cacheOrderAddress.Total
		c.JSON(http.StatusOK, result)
		return
	}

	orders, total, err := r.mysql.FindOrderByAddress(p.ReceiveAddress, p.Limit, p.OffSet)

	if err != nil {
		log.Error("Router", "FindOrderByAddress", fmt.Sprintf("mysql.FindOrders is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	cacheOrder := &utils.OrderAddressCache{
		Orders:      orders,
		Total:       total,
		CacheNumber: maxHeight,
	}

	r.level.SetCacheOrderAddress(p.ReceiveAddress, cacheOrder)

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = orders
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrdersIndex(c *gin.Context) {
	type params struct {
		Address string `json:"address"`
		Tick    string `json:"tick"`
		Hash    string `json:"hash"`
		Number  int64  `json:"number"`
		Limit   int64  `json:"limit"`
		OffSet  int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if p.Limit > 50 {
		p.Limit = 50
	}

	orders, total, err := r.mysql.FindOrdersindex(p.Address, p.Tick, p.Hash, p.Number, p.Limit, p.OffSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = orders
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrdersHash(c *gin.Context) {
	type params struct {
		Hash string
	}

	p := &params{}
	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	orders, err := r.mysql.FindOrderByDrc20Hash(p.Hash)

	if err != nil {
		log.Error("Router", "FindOrdersHash", fmt.Sprintf("mysql.FindOrderByDrc20Hash is err:%s", err.Error()))
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server err"
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = orders
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrdersByid(c *gin.Context) {
	type params struct {
		OrderId string `json:"order_id"`
	}

	p := &params{}
	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	order, err := r.mysql.FindOrderById(p.OrderId)
	if err != nil {
		log.Error("Router", "FindOrdersByid", fmt.Sprintf("mysql.FindOrderById is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = order
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOrdersByTick(c *gin.Context) {
	type params struct {
		ReceiveAddress string `json:"receive_address"`
		Tick           string `json:"tick"`
		Limit          int64  `json:"limit"`
		OffSet         int64  `json:"offset"`
	}

	p := &params{
		Limit:  50,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if p.Limit > 50 {
		p.Limit = 50
	}

	_, err := btcutil.DecodeAddress(p.ReceiveAddress, &chaincfg.MainNetParams)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}
	orders, total, err := r.mysql.FindOrderBytick(p.ReceiveAddress, p.Tick, p.Limit, p.OffSet)
	if err != nil {
		log.Error("Router", "FindOrdersByTick", fmt.Sprintf("mysql.FindOrderBytick is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = orders
	result.Total = total
	c.JSON(http.StatusOK, result)
}

func (r *Router) FindOgAddressAll(c *gin.Context) {

	adds, err := r.mysql.FindOgAddress()
	if err != nil {
		log.Error("Router", "FindOgAddressAll", fmt.Sprintf("mysql.FindOgAddress is err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = adds
	c.JSON(http.StatusOK, result)

}

func (r *Router) ConvertAddress(c *gin.Context) {
	type params struct {
		PublicKey string `json:"public_key"`
	}

	p := &params{}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	PublicKey, err := hex.DecodeString(p.PublicKey)

	pkHash := btcutil.Hash160(PublicKey)
	script, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(pkHash).Script()
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	inadressBad, err := btcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}
	script2, err := txscript.PayToAddrScript(inadressBad)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	inadressBad2, err := btcutil.NewAddressScriptHash(script2, &chaincfg.MainNetParams)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	inadressd, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(PublicKey), &chaincfg.MainNetParams)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	err = r.mysql.UpdateConvertAddress(inadressBad2.String(), inadressd.String())
	if err != nil {
		c.JSON(500, gin.H{"error": "server error"})
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	c.JSON(http.StatusOK, result)
}
