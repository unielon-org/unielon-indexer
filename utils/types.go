package utils

import (
	"github.com/dogecoinw/go-dogecoin/rlp"
	"github.com/unielon-org/unielon-indexer/models"
	"io"
	"math/big"
)

// Config
type HttpConfig struct {
	Switch bool   `json:"switch"`
	Server string `json:"server"`
}

type LevelDBConfig struct {
	Path string `json:"path"`
}

type SqliteConfig struct {
	Switch   bool   `json:"switch"`
	Database string `json:"database"`
}

type MysqlConfig struct {
	Switch   bool   `json:"switch"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	UserName string `json:"user_name"`
	PassWord string `json:"pass_word"`
	Database string `json:"database"`
}

type ChainConfig struct {
	ChainName string `json:"chain_name"`
	Rpc       string `json:"rpc"`
	UserName  string `json:"user_name"`
	PassWord  string `json:"pass_word"`
}

type ExplorerConfig struct {
	Switch       bool  `json:"switch"`
	FromBlock    int64 `json:"from_block"`
	InitMintData bool  `json:"init_mint_data"`
	InitForkData bool  `json:"init_fork_data"`
}

type HttpResult struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
}

type OrderAddressCache struct {
	Orders      []*models.Drc20Info
	Total       int64
	CacheNumber int64
}

type extOrderAddressCache struct {
	Orders      []*models.Drc20Info
	Total       int64
	CacheNumber int64
}

func (r *OrderAddressCache) DecodeRLP(s *rlp.Stream) error {
	var et extOrderAddressCache
	if err := s.Decode(&et); err != nil {
		return err
	}
	return nil
}

func (r *OrderAddressCache) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, extOrderAddressCache{
		Orders:      r.Orders,
		Total:       r.Total,
		CacheNumber: r.CacheNumber,
	})
}

// summary
type SwapPairSummary struct {
	Tick                  string  `json:"tick"`
	Tick0                 string  `json:"tick0"`
	Tick1                 string  `json:"tick1"`
	Amt0                  string  `json:"amt0"`
	Amt1                  string  `json:"amt1"`
	PriceChangePercent24H float64 `json:"price_change_percent_24h"`
	Liquidity             float64 `json:"liquidity"`
	BaseVolume            float64 `json:"base_volume"`
	DogeUsdt              float64 `json:"doge_usdt"`
}

type SummaryAllResult struct {
	Tick                  string  `json:"tick"`
	MaxAmt                string  `json:"max_amt"`
	AmtSum                string  `json:"amt_sum"`
	MarketCap             string  `json:"market_cap"`
	LastPrice             float64 `json:"last_price"`
	BaseVolume            float64 `json:"base_volume"`
	PriceChangePercent24H float64 `json:"price_change_percent_24h"`
	Holders               *uint64 `json:"holders"`
	FootPrice             float64 `json:"foot_price"`
	Logo                  *string `json:"logo"`
	IsCheck               uint64  `json:"is_check"`
	Liquidity             float64 `json:"liquidity"`
}

type ExchangeInfoSummary struct {
	Id          int64    `json:"id"`
	Tick        string   `json:"tick"`
	Tick0       string   `json:"tick0"`
	Tick1       string   `json:"tick1"`
	OpenPrice   float64  `json:"open_price"`
	ClosePrice  float64  `json:"close_price"`
	LowestAsk   float64  `json:"lowest_ask"`
	HighestBid  float64  `json:"highest_bid"`
	BaseVolume  *big.Int `json:"base_volume"`
	QuoteVolume *big.Int `json:"quote_volume"`
	LastDate    string   `json:"last_date"`
}

type ExchangeDrc20Collect struct {
	Id         int64    `json:"id"`
	Tick       string   `json:"tick"`
	OldPrice   float64  `json:"old_price"`
	LastPrice  float64  `json:"last_price"`
	Amt        *big.Int `json:"amt"`
	Holders    int64    `json:"holders"`
	UpdateDate int64    `json:"update_date"`
	CreateDate int64    `json:"create_date"`
}

type Infos struct {
	P             string   `json:"p"`
	Op            string   `json:"op"`
	Tick          string   `json:"tick"`
	Tick0         string   `json:"tick0"`
	Tick1         string   `json:"tick1"`
	Amt           *big.Int `json:"amt"`
	Amt0          *big.Int `json:"amt0"`
	Amt1          *big.Int `json:"amt1"`
	Liquidity     *big.Int `json:"liquidity"`
	HolderAddress string   `json:"holder_address"`
	ToAddress     string   `json:"to_address"`
	BlockNumber   int64    `json:"block_number"`
	BlockHash     string   `json:"block_hash"`
	TxHash        string   `json:"tx_hash"`
	OrderStatus   int64    `json:"order_status"`
	UpdateDate    int64    `json:"update_date"`
}
