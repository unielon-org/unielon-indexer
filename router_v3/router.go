package router_v3

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/doged/wire"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/storage_v3"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
	"strings"
)

type Router struct {
	mysql *storage_v3.MysqlClient
	dbc   *storage.DBClient
	node  *rpcclient.Client
	level *storage.LevelDB
	ipfs  *shell.Shell

	verify *verifys.Verifys
}

func NewRouter(mysql *storage_v3.MysqlClient, dbc *storage.DBClient, level *storage.LevelDB, node *rpcclient.Client, ipfs *shell.Shell) *Router {
	return &Router{
		mysql:  mysql,
		node:   node,
		ipfs:   ipfs,
		level:  level,
		dbc:    dbc,
		verify: verifys.NewVerifys(dbc),
	}
}

func (r *Router) LastNumber(c *gin.Context) {

	maxHeight := int64(0)
	err := r.dbc.DB.Model(&models.Block{}).Select("max(block_number)").First(&maxHeight).Error
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
	result.Data = maxHeight
	c.JSON(http.StatusOK, result)
}

func (r *Router) TxBroadcast(c *gin.Context) {
	type params struct {
		TxHex string `json:"tx_hex"`
	}

	p := &params{}
	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	bytesData, err := hex.DecodeString(p.TxHex)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	msgTx := new(wire.MsgTx)
	err = msgTx.Deserialize(bytes.NewReader(bytesData))
	if err != nil {
		fmt.Println(err)
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	txhash, err := r.node.SendRawTransaction(msgTx, true)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	data := make(map[string]interface{})
	data["tx_hash"] = txhash.String()
	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = data
	c.JSON(http.StatusOK, result)
}

func (r *Router) SwapSummaryK(c *gin.Context) {
	type params struct {
		Tick         string `json:"tick"`
		DateInterval string `json:"date_interval"`
	}

	p := &params{
		DateInterval: "1d",
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	p.DateInterval = strings.ToLower(p.DateInterval)

	resultall := make([]*storage_v3.SwapInfoSummary, 0)
	resultnew, err := r.mysql.FindCMCSummaryKNew(p.Tick, p.DateInterval)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}
	if resultnew != nil {
		resultall = append(resultall, resultnew)
	}

	results, err := r.mysql.FindCMCSummaryK(p.Tick, p.DateInterval)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	if results != nil {
		resultall = append(resultall, results...)
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = resultall
	c.JSON(http.StatusOK, result)

}
