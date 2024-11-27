package storage_v3

import (
	"math/big"
)

type Drc20CollectAll struct {
	Tick         string   `json:"tick"`
	MintAmt      *big.Int `json:"mint_amt"`
	MaxAmt       *big.Int `json:"max_amt"`
	Dec          uint8    `json:"dec"`
	Lim          *big.Int `json:"lim"`
	Holders      uint64   `json:"holders"`
	Transactions uint64   `json:"transactions"`
	DeployTime   string   `json:"deploy_time"`
	LastMintTime string   `json:"last_mint_time"`
	DeployBy     string   `json:"deploy_by"`
	Inscription  string   `json:"inscription"`
	Logo         *string  `json:"logo"`
	Introduction *string  `json:"introduction"`
	WhitePaper   *string  `json:"white_paper"`
	Official     *string  `json:"official"`
	Telegram     *string  `json:"telegram"`
	Discorad     *string  `json:"discorad"`
	Twitter      *string  `json:"twitter"`
	Facebook     *string  `json:"facebook"`
	Github       *string  `json:"github"`
	IsCheck      uint64   `json:"is_check"`
}

type Drc20CollectAllCache struct {
	Results     []*Drc20CollectAll
	Total       int64
	CacheNumber int64
}

type FindDrc20AllByAddressResult struct {
	Tick string   `json:"tick"`
	Amt  *big.Int `json:"amt"`
}

type FindDrc20HoldersResult struct {
	Address string   `json:"address"`
	Amt     *big.Int `json:"amt"`
}

type FindSummaryResult struct {
	Exchange int64   `json:"exchange"`
	Value24h float64 `json:"value_all"`
}

type FindSummaryAllResult struct {
	TradingPairs          string  `json:"trading_pairs"`
	Tick                  string  `json:"tick"`
	MaxAmt                string  `json:"max_amt"`
	TotalMaxAmt           string  `json:"total_max_amt"`
	LastPrice             float64 `json:"last_price"`
	OldPrice              float64 `json:"old_price"`
	LowestAsk             float64 `json:"lowest_ask"`
	HighestBid            float64 `json:"highest_bid"`
	BaseVolume            float64 `json:"base_volume"`
	QuoteVolume           float64 `json:"quote_volume"`
	PriceChangePercent24H float64 `json:"price_change_percent_24h"`
	HighestPrice24H       float64 `json:"highest_price_24h"`
	LowestPrice24H        float64 `json:"lowest_price_24h"`
	LastDate              string  `json:"last_date"`
	Holders               *uint64 `json:"holders"`
	FootPrice             float64 `json:"foot_price"`
	Logo                  *string `json:"logo"`
	IsCheck               uint64  `json:"is_check"`
	Liquidity             float64 `json:"liquidity"`
}

type Kline struct {
	Time  string  `json:"time"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

type OrderResult struct {
	OrderId            string   `json:"order_id"`
	P                  string   `json:"p"`
	Op                 string   `json:"op"`
	Tick               string   `json:"tick"`
	Amt                *big.Int `json:"amt"`
	Max                *big.Int `json:"max"`
	Lim                *big.Int `json:"lim"`
	Dec                int64    `json:"dec"`
	Burn               string   `json:"burn"`
	Func               string   `json:"func"`
	RateFee            *big.Int `json:"rate_fee"`
	Repeat             int64    `json:"repeat"`
	FeeTxHash          string   `json:"fee_tx_hash"`
	Drc20TxHash        string   `json:"drc20_tx_hash"`
	BlockHash          string   `json:"block_hash"`
	BlockNumber        int64    `json:"block_number"`
	BlockConfirmations uint64   `json:"block_confirmations"`
	Drc20Inscription   string   `json:"drc20_inscription"`
	Inscription        string   `json:"inscription"`
	ReceiveAddress     string   `json:"receive_address"`
	ToAddress          string   `json:"to_address"`
	FeeAddress         string   `json:"fee_address"`
	OrderStatus        int64    `json:"order_status"`
	CreateDate         string   `json:"create_date"`
}

type SwapInfo struct {
	ID              uint     `gorm:"primarykey" json:"id"`
	OrderId         string   `json:"order_id"`
	Op              string   `json:"op"`
	Tick            string   `json:"tick"`
	Tick0           string   `json:"tick0"`
	Tick1           string   `json:"tick1"`
	Amt0            *big.Int `json:"amt0"`
	Amt1            *big.Int `json:"amt1"`
	Amt0Min         *big.Int `json:"amt0_min"`
	Amt1Min         *big.Int `json:"amt1_min"`
	Amt0Out         *big.Int `json:"amt0_out"`
	Amt1Out         *big.Int `json:"amt1_out"`
	Liquidity       *big.Int `json:"liquidity"`
	Doge            int      `json:"doge"`
	HolderAddress   string   `json:"holder_address"`
	FeeAddress      string   `json:"fee_address"`
	FeeTxHash       string   `json:"fee_tx_hash"`
	FeeTxIndex      uint32   `json:"fee_tx_index"`
	SwapTxHash      string   `json:"swap_tx_hash"`
	SwapBlockNumber int64    `json:"swap_block_number"`
	SwapBlockHash   string   `json:"swap_block_hash"`
	OrderStatus     int64    `json:"order_status"`
	ErrInfo         string   `json:"err_info"`
	UpdateDate      string   `json:"update_date"`
	CreateDate      string   `json:"create_date"`
}

type SwapPrice struct {
	Tick      string  `json:"tick"`
	LastPrice float64 `json:"last_price"`
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

type SwapInfoSummaryTVLAll struct {
	Liquidity  *big.Int `json:"liquidity"`
	BaseVolume *big.Int `json:"base_volume"`
	DogeUsdt   float64  `json:"doge_usdt"`
	LastDate   string   `json:"last_date"`
}

type StakeCollectAddress struct {
	ID             uint     `gorm:"primarykey" json:"id"`
	StakeId        string   `json:"stake_id"`
	Tick           string   `json:"tick"`
	Amt            string   `json:"amt"`
	Reward         string   `json:"reward"`
	ReceivedReward string   `json:"received_reward"`
	HolderAddress  string   `json:"holder_address"`
	CardiAmt       *big.Int `gorm:"-" json:"cardi_amt"`
	UpdateDate     int64    `json:"update_date"`
	CreateDate     int64    `json:"create_date"`
}
