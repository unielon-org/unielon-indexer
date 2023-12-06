package main

import (
	"context"
	"fmt"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/gin-gonic/gin"
	"github.com/unielon-org/unielon-indexer/config"
	"github.com/unielon-org/unielon-indexer/explorer"
	"github.com/unielon-org/unielon-indexer/router"
	"github.com/unielon-org/unielon-indexer/storage"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	feeAddress = "D92uJjQ9eHUcv2GjJUgp6m58V8wYvGV2g9"
)

var (
	cfg config.Config
)

func main() {

	// Load configuration file
	config.LoadConfig(&cfg, "")

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(true)))
	glogger.Verbosity(log.Lvl(cfg.DebugLevel))
	log.Root().SetHandler(glogger)

	// Build a channel for Ctrip to exit
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Listen for SIGINT and SIGTERM signals
	dbClient := storage.NewSqliteClient(cfg.Sqlite)

	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.Chain.Rpc,
		Endpoint:     "ws",
		User:         cfg.Chain.UserName,
		Pass:         cfg.Chain.PassWord,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	// Notice the notification parameter is nil since notifications are
	rpcClient, _ := rpcclient.New(connCfg, nil)

	exp := explorer.NewExplorer(ctx, wg, rpcClient, dbClient, cfg.Server.FromBlock, feeAddress)
	wg.Add(1)
	go exp.Start()

	rt := router.NewRouter(dbClient, rpcClient, feeAddress)

	// Create a new Gin router instance
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	router.POST("/v3/info/lastnumber", rt.LastNumber)

	// DRC20
	router.POST("/v3/drc20/all", rt.FindDrc20All)
	router.POST("/v3/drc20/tick", rt.FindDrc20ByTick)
	router.POST("/v3/drc20/holders", rt.FindDrc20Holders)
	router.POST("/v3/drc20/address", rt.FindDrc20ByAddress)
	router.POST("/v3/drc20/address/tick", rt.FindDrc20ByAddressTick)

	router.POST("/v3/drc20/order", rt.FindOrders)

	// SWAP
	router.POST("/v3/swap/getreserves", rt.SwapGetReserves)
	router.POST("/v3/swap/getreserves/all", rt.SwapGetReservesAll)
	router.POST("/v3/swap/getliquidity", rt.SwapGetLiquidity)
	router.POST("/v3/swap/order", rt.SwapInfo)

	router.POST("/v3/swap/price", rt.SwapPrice)

	// DOGEW
	router.POST("/v3/wdoge/order", rt.WDogeInfo)

	// Start the HTTP server and listen on the port
	go router.Run(cfg.Server.Port)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived an interrupt, stopping services...")
		cancel() // Cancel context, this will cancel all workers
	}()
	wg.Wait()
}
