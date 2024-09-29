package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
)

type InfoRouter struct {
	dbc    *storage.DBClient
	node   *rpcclient.Client
	ipfs   *shell.Shell
	level  *storage.LevelDB
	verify *verifys.Verifys
}

func NewInfoRouter(db *storage.DBClient, node *rpcclient.Client, level *storage.LevelDB, ipfs *shell.Shell, verify *verifys.Verifys) *InfoRouter {
	return &InfoRouter{
		dbc:    db,
		node:   node,
		ipfs:   ipfs,
		level:  level,
		verify: verify,
	}
}

func (r *InfoRouter) LastNumber(c *gin.Context) {
	maxHeight := 0
	err := r.dbc.DB.Model(&models.Block{}).Select("max(block_number)").Scan(&maxHeight).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusOK, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = maxHeight
	c.JSON(http.StatusOK, result)
}
