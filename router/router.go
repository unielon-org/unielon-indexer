package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/verifys"
)

type Router struct {
	dbc  *storage.DBClient
	node *rpcclient.Client

	verify *verifys.Verifys
}
