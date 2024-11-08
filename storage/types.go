package storage

import (
	"math/big"
)

// SWAP
type SwapPrice struct {
	Tick      string `json:"tick"`
	LastPrice string `json:"last_price"`
}

type SwapInfoSummary struct {
	Id          int64    `json:"id"`
	Tick        string   `json:"tick"`
	Tick0       string   `json:"tick0"`
	Tick1       string   `json:"tick1"`
	OpenPrice   float64  `json:"open_price"`
	ClosePrice  float64  `json:"close_price"`
	LowestAsk   float64  `json:"lowest_ask"`
	HighestBid  float64  `json:"highest_bid"`
	Liquidity   *big.Int `json:"liquidity"`
	BaseVolume  *big.Int `json:"base_volume"`
	QuoteVolume *big.Int `json:"quote_volume"`
	LastDate    string   `json:"last_date"`
	DogeUsdt    float64  `json:"doge_usdt"`
}

type SwapInfoSummaryTvl struct {
	DogeUsdt   float64  `json:"doge_usdt"`
	Liquidity  *big.Int `json:"liquidity"`
	BaseVolume *big.Int `json:"base_volume"`
	LastDate   string   `json:"last_date"`
	ClosePrice float64  `json:"close_price"`
	OpenPrice  float64  `json:"open_price"`
	HighestBid float64  `json:"highest_bid"`
	LowestAsk  float64  `json:"lowest_ask"`
}
