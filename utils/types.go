package utils

import (
	"github.com/dogecoinw/doged/btcutil"
	"math/big"
)

// Config
type ServerConfig struct {
	Port      string `json:"port"`
	FromBlock int64  `json:"from_block"`
}

type SqliteConfig struct {
	Database string `json:"database"`
}

type ChainConfig struct {
	ChainName string `json:"chain_name"`
	Rpc       string `json:"rpc"`
	UserName  string `json:"user_name"`
	PassWord  string `json:"pass_word"`
}

// RouterResult
type HttpResult struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
}

type BaseParams struct {
	P  string `json:"p"`
	Op string `json:"op"`
}

type NewParams struct {
	P              string `json:"p"`
	Op             string `json:"op"`
	Tick           string `json:"tick"`
	Max            string `json:"max"`
	Amt            string `json:"amt"`
	Lim            string `json:"lim"`
	Dec            int64  `json:"dec"`
	Burn           string `json:"burn"`
	Func           string `json:"func"`
	ReceiveAddress string `json:"receive_address"`
	ToAddress      string `json:"to_address"`
	RateFee        string `json:"rate_fee"`
	Repeat         int64  `json:"repeat"`
}

type Drc20Params struct {
	Tick      string `json:"tick"`
	Limit     uint64 `json:"limit"`
	OffSet    uint64 `json:"offset"`
	Completed uint64 `json:"completed"`
}

type SwapParams struct {
	Op            string `json:"op"`
	Tick0         string `json:"tick0"`
	Tick1         string `json:"tick1"`
	Amt0          string `json:"amt0"`
	Amt1          string `json:"amt1"`
	Amt0Min       string `json:"amt0_min"`
	Amt1Min       string `json:"amt1_min"`
	Liquidity     string `json:"liquidity"`
	Path          string `json:"path"`
	HolderAddress string `json:"holder_address"`
}

type WDogeParams struct {
	Op            string `json:"op"`
	Tick          string `json:"tick"`
	Amt           string `json:"amt"`
	HolderAddress string `json:"holder_address"`
}

type NFTParams struct {
	Op            string `json:"op"`
	Tick          string `json:"tick"`
	TickId        int64  `json:"tick_id"`
	Total         int64  `json:"total"`
	Model         string `json:"model"`
	Prompt        string `json:"prompt"`
	Seed          int64  `json:"seed"`
	Image         string `json:"image"`
	OriginImage   string `json:"originImage"`
	HolderAddress string `json:"holder_address"`
	ToAddress     string `json:"to_address"`
}

type StakeParams struct {
	Op            string `json:"op"`
	Tick          string `json:"tick"`
	Amt           string `json:"amt"`
	HolderAddress string `json:"holder_address"`
}

// model
type AddressInfo struct {
	OrderId        string       `json:"order_id"`
	PrveWif        *btcutil.WIF `json:"prve_wif"`
	PubKey         string       `json:"pub_key"`
	Address        string       `json:"address"`
	ReceiveAddress string       `json:"receive_address"`
	FeeAddress     string       `json:"fee_address"`
}

// drc20
type Cardinals struct {
	OrderId        string   `json:"order_id"`
	P              string   `json:"p"`
	Op             string   `json:"op"`
	Tick           string   `json:"tick"`
	Amt            *big.Int `json:"amt"`
	Max            *big.Int `json:"max"`
	Lim            *big.Int `json:"lim"`
	Dec            int64    `json:"dec"`
	Burn           string   `json:"burn"`
	Func           string   `json:"func"`
	Repeat         int64    `json:"repeat"`
	Drc20TxHash    string   `json:"drc20_tx_hash"`
	FeeTxHash      string   `json:"fee_tx_hash"`
	BlockNumber    int64    `json:"block_number"`
	BlockHash      string   `json:"block_hash"`
	ReceiveAddress string   `json:"receive_address"`
	ToAddress      string   `json:"to_address"`
	FeeAddress     string   `json:"fee_address"`
	OrderStatus    int64    `json:"order_status"`
	ErrInfo        string   `json:"err_info"`
	CreateDate     string   `json:"create_date"`
}

// SWAP
type SwapInfo struct {
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
	Path            []string `json:"path"`
	Liquidity       *big.Int `json:"liquidity"`
	HolderAddress   string   `json:"holder_address"`
	FeeAddress      string   `json:"fee_address"`
	FeeTxHash       string   `json:"fee_tx_hash"`
	SwapTxHash      string   `json:"swap_tx_hash"`
	SwapBlockNumber int64    `json:"swap_block_number"`
	SwapBlockHash   string   `json:"swap_block_hash"`
	OrderStatus     int64    `json:"order_status"`
	UpdateDate      string   `json:"update_date"`
	CreateDate      string   `json:"create_date"`
}

// swap_liquidity
type SwapLiquidity struct {
	Tick            string   `json:"tick"`
	Tick0           string   `json:"tick0"`
	Tick1           string   `json:"tick1"`
	Amt0            *big.Int `json:"amt0"`
	Amt1            *big.Int `json:"amt1"`
	Path            string   `json:"path"`
	LiquidityTotal  *big.Int `json:"liquidity_total"`
	ReservesAddress string   `json:"reserves_address"`
	HolderAddress   string   `json:"holder_address"`
}

// WDOGE
type WDogeInfo struct {
	OrderId          string   `json:"order_id"`
	Op               string   `json:"op"`
	Tick             string   `json:"tick"`
	Amt              *big.Int `json:"amt"`
	HolderAddress    string   `json:"holder_address"`
	FeeAddress       string   `json:"fee_address"`
	FeeTxHash        string   `json:"fee_tx_hash"`
	WDogeTxHash      string   `json:"wdoge_tx_hash"`
	WDogeBlockNumber int64    `json:"wdoge_block_number"`
	WDogeBlockHash   string   `json:"wdoge_block_hash"`
	UpdateDate       string   `json:"update_date"`
	CreateDate       string   `json:"create_date"`
}

// cardinals_revert
type CardinalsRevert struct {
	Tick        string   `json:"tick"`
	FromAddress string   `json:"from_address"`
	ToAddress   string   `json:"to_address"`
	Amt         *big.Int `json:"amt"`
	BlockNumber int64    `json:"block_number"`
}

// NFT
type NFTInfo struct {
	OrderId        string  `json:"order_id"`
	Op             string  `json:"op"`
	Tick           string  `json:"tick"`
	TickId         int64   `json:"tick_id"`
	Total          int64   `json:"total"`
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	Image          string  `json:"image"`
	ImageData      []byte  `json:"image_data"`
	HolderAddress  string  `json:"holder_address"`
	ToAddress      string  `json:"to_address"`
	AdminAddress   string  `json:"admin_address"`
	FeeAddress     string  `json:"fee_address"`
	FeeAddressAll  string  `json:"fee_address_all"`
	FeeTxHash      string  `json:"fee_tx_hash"`
	FeeTxIndex     uint32  `json:"fee_tx_index"`
	FeeBlockNumber int64   `json:"fee_block_number"`
	FeeBlockHash   string  `json:"fee_block_hash"`
	NftTxHash      string  `json:"nft_tx_hash"`
	NftBlockNumber int64   `json:"nft_block_number"`
	NftBlockHash   string  `json:"nft_block_hash"`
	ErrInfo        *string `json:"err_info"`
	OrderStatus    int64   `json:"order_status"`
	UpdateDate     string  `json:"update_date"`
	CreateDate     string  `json:"create_date"`
}

// NftCollect
type NftCollect struct {
	Tick          string `json:"tick"`
	TickSum       int64  `json:"tick_sum"`
	Total         int64  `json:"total"`
	Model         string `json:"model"`
	Prompt        string `json:"prompt"`
	Image         string `json:"image"`
	HolderAddress string `json:"holder_address"`
	DeployHash    string `json:"deploy_hash"`
	CreateDate    string `json:"create_date"`
}

// NftCollectAddress
type NftCollectAddress struct {
	Tick          string `json:"tick"`
	TickId        int64  `json:"tick_id"`
	Prompt        string `json:"prompt"`
	NftPrompt     string `json:"nft_prompt"`
	NftModel      string `json:"nft_model"`
	Image         string `json:"image"`
	DeployHash    string `json:"deploy_hash"`
	HolderAddress string `json:"holder_address"`
	CreateDate    string `json:"create_date"`
}

// NftRevert
type NftRevert struct {
	Tick        string `json:"tick"`
	TickId      int64  `json:"tick_id"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	BlockNumber int64  `json:"block_number"`
	Prompt      string `json:"prompt"`
	Image       string `json:"image"`
	DeployHash  string `json:"deploy_hash"`
}

// stake model
type StakeInfo struct {
	OrderId          string         `json:"order_id"`
	Op               string         `json:"op"`
	Tick             string         `json:"tick"`
	Amt              *big.Int       `json:"amt"`
	FeeTxHash        string         `json:"fee_tx_hash"`
	FeeTxIndex       uint32         `json:"fee_tx_index"`
	FeeTxRaw         *string        `json:"fee_tx_raw"`
	FeeBlockHash     string         `json:"fee_block_hash"`
	FeeBlockNumber   int64          `json:"fee_block_number"`
	StakeTxHash      string         `json:"stake_tx_hash"`
	StakeTxRaw       *string        `json:"stake_tx_raw"`
	StakeBlockHash   string         `json:"stake_block_hash"`
	StakeBlockNumber int64          `json:"stake_block_number"`
	FeeAddress       string         `json:"fee_address"`
	HolderAddress    string         `json:"holder_address"`
	ErrInfo          *string        `json:"err_info"`
	StakeRewardInfos []*StakeRevert `json:"stake_reward_infos"`
	OrderStatus      int64          `json:"order_status"`
	UpdateDate       int64          `json:"update_date"`
	CreateDate       int64          `json:"create_date"`
}

type StakeCollect struct {
	Id              int64    `json:"id"`
	Tick            string   `json:"tick"`
	Amt             *big.Int `json:"amt"`
	Reward          *big.Int `json:"reward"`
	ReservesAddress string   `json:"reserves_address"`
	Holders         int64    `json:"holders"`
	UpdateDate      int64    `json:"update_date"`
	CreateDate      int64    `json:"create_date"`
}

type StakeCollectAddress struct {
	Id             int64    `json:"id"`
	Tick           string   `json:"tick"`
	Amt            *big.Int `json:"amt"`
	Reward         *big.Int `json:"reward"`
	ReceivedReward *big.Int `json:"received_reward"`
	HolderAddress  string   `json:"holder_address"`
	CardiAmt       *big.Int `json:"cardi_amt"`
	UpdateDate     int64    `json:"update_date"`
	CreateDate     int64    `json:"create_date"`
}

type StakeCollectReward struct {
	Tick       string   `json:"tick"`
	RewardTick string   `json:"reward_tick"`
	Reward     *big.Int `json:"reward"`
	UpdateDate int64    `json:"update_date"`
	CreateDate int64    `json:"create_date"`
}

type StakeRevert struct {
	Tick        string   `json:"tick"`
	FromAddress string   `json:"from_address"`
	ToAddress   string   `json:"to_address"`
	Amt         *big.Int `json:"amt"`
	BlockNumber int64    `json:"block_number"`
}

type HolderReward struct {
	Tick            string   `json:"tick"`
	TotalRewardPool *big.Int `json:"total_reward_pool"`
	Reward          *big.Int `json:"reward"`
	ReceivedReward  *big.Int `json:"received_reward"`
}
