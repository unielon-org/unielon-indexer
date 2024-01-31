package storage

import (
	"math/big"
)

type FindDrc20AllByAddressResult struct {
	Tick        string   `json:"tick"`
	Amt         *big.Int `json:"amt"`
	Inscription string   `json:"inscription"`
}

type FindDrc20AllResult struct {
	Tick        string   `json:"tick"`
	MintAmt     *big.Int `json:"mint_amt"`
	MaxAmt      *big.Int `json:"max_amt"`
	Dec         uint8    `json:"dec"`
	Lim         *big.Int `json:"lim"`
	Holders     uint64   `json:"holders"`
	DeployTime  string   `json:"deploy_time"`
	DeployBy    string   `json:"deploy_by"`
	Inscription string   `json:"inscription"`
}

type FindDrc20HoldersResult struct {
	Address string   `json:"address"`
	Amt     *big.Int `json:"amt"`
}

type OrderResult struct {
	OrderId          string   `json:"order_id"`
	P                string   `json:"p"`
	Op               string   `json:"op"`
	Tick             string   `json:"tick"`
	Amt              *big.Int `json:"amt"`
	Max              *big.Int `json:"max"`
	Lim              *big.Int `json:"lim"`
	Dec              int64    `json:"dec"`
	Burn             string   `json:"burn"`
	Func             string   `json:"func"`
	Repeat           int64    `json:"repeat"`
	Drc20TxHash      string   `json:"drc20_tx_hash"`
	FeeTxHash        string   `json:"fee_tx_hash"`
	Inscription      string   `json:"inscription"`
	Drc20Inscription string   `json:"drc20_inscription"`
	BlockHash        string   `json:"block_hash"`
	BlockNumber      int64    `json:"block_number"`
	ReceiveAddress   string   `json:"receive_address"`
	ToAddress        string   `json:"to_address"`
	FeeAddress       string   `json:"fee_address"`
	OrderStatus      int64    `json:"order_status"`
	CreateDate       string   `json:"create_date"`
}

type SwapPrice struct {
	Tick      string  `json:"tick"`
	LastPrice float64 `json:"last_price"`
}

// NFT model
type NftCollect struct {
	Tick          string  `json:"tick"`
	TickSum       int64   `json:"tick_sum"`
	Total         int64   `json:"total"`
	Prompt        string  `json:"prompt"`
	Image         string  `json:"image"`
	HolderAddress string  `json:"holder_address"`
	DeployHash    string  `json:"deploy_hash"`
	Transactions  int64   `json:"transactions"`
	Holders       int64   `json:"holders"`
	Introduction  *string `json:"introduction"`
	WhitePaper    *string `json:"white_paper"`
	Official      *string `json:"official"`
	Telegram      *string `json:"telegram"`
	Discorad      *string `json:"discorad"`
	Twitter       *string `json:"twitter"`
	Facebook      *string `json:"facebook"`
	Github        *string `json:"github"`
	IsCheck       uint64  `json:"is_check"`
	UpdateDate    int64   `json:"update_date"`
	CreateDate    int64   `json:"create_date"`
}

type NftCollectAddress struct {
	Tick          string `json:"tick"`
	TickId        int64  `json:"tick_id"`
	Image         string `json:"image"`
	HolderAddress string `json:"holder_address"`
	Transactions  int64  `json:"transactions"`
	UpdateDate    int64  `json:"update_date"`
	CreateDate    int64  `json:"create_date"`
}
