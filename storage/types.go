package storage

import (
	"math/big"
)

type FindDrc20AllByAddressResult struct {
	Tick string   `json:"tick"`
	Amt  *big.Int `json:"amt"`
}

type FindDrc20AllResult struct {
	Tick         string   `json:"tick"`
	MintAmt      *big.Int `json:"mint_amt"`
	MaxAmt       *big.Int `json:"max_amt"`
	Dec          uint8    `json:"dec"`
	Lim          *big.Int `json:"lim"`
	Holders      uint64   `json:"holders"`
	Transactions uint64   `json:"transactions"`
	DeployTime   int64    `json:"deploy_time"`
	LastMintTime *int64   `json:"last_mint_time"`
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

type FindDrc20HoldersResult struct {
	Address string   `json:"address"`
	Amt     *big.Int `json:"amt"`
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
	BlockConfirmations uint64   `json:"block_confirmations"`
	ReceiveAddress     string   `json:"receive_address"`
	ToAddress          string   `json:"to_address"`
	FeeAddress         string   `json:"fee_address"`
	OrderStatus        int64    `json:"order_status"`
	CreateDate         string   `json:"create_date"`
}

type SwapPrice struct {
	Tick      string  `json:"tick"`
	LastPrice float64 `json:"last_price"`
}
