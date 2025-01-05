package models

import "math/big"

type Drc20Info struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `json:"order_id"`
	P             string    `json:"p"`
	Op            string    `json:"op"`
	Tick          string    `json:"tick"`
	Amt           *Number   `json:"amt"`
	Max           *Number   `gorm:"column:max_" json:"max"`
	Lim           *Number   `gorm:"column:lim_" json:"lim"`
	Dec           uint      `gorm:"column:dec_" json:"dec"`
	Burn          string    `gorm:"column:burn_" json:"burn"`
	Func          string    `gorm:"column:func_" json:"func"`
	Repeat        int64     `gorm:"column:repeat_mint" json:"repeat"`
	HolderAddress string    `json:"holder_address"`
	ToAddress     string    `json:"to_address"`
	FeeAddress    string    `json:"fee_address"`
	FeeTxHash     string    `json:"fee_tx_hash"`
	TxHash        string    `json:"tx_hash"`
	BlockNumber   int64     `json:"block_number"`
	BlockHash     string    `json:"block_hash"`
	ErrInfo       string    `json:"err_info"`
	OrderStatus   int64     `json:"order_status"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `gorm:"column:create_date" json:"create_date"`
}

func (Drc20Info) TableName() string {
	return "drc20_info"
}

type Drc20Collect struct {
	Tick          string    `json:"tick"`
	AmtSum        *Number   `json:"amt_sum"`
	Max           *Number   `gorm:"column:max_" json:"max"`
	RealSum       *Number   `json:"real_sum"`
	Lim           *Number   `gorm:"column:lim_" json:"lim"`
	Dec           uint      `gorm:"column:dec_" json:"dec"`
	Burn          string    `gorm:"column:burn_" json:"burn"`
	Func          string    `gorm:"column:func_" json:"func"`
	HolderAddress string    `json:"holder_address"`
	TxHash        string    `json:"tx_hash"`
	Transactions  uint64    `json:"transactions"`
	Logo          *string   `json:"logo"`
	Introduction  *string   `json:"introduction"`
	WhitePaper    *string   `json:"white_paper"`
	Official      *string   `json:"official"`
	Telegram      *string   `json:"telegram"`
	Discorad      *string   `json:"discorad"`
	Twitter       *string   `json:"twitter"`
	Facebook      *string   `json:"facebook"`
	Github        *string   `json:"github"`
	IsCheck       uint64    `json:"is_check"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
}

func (Drc20Collect) TableName() string {
	return "drc20_collect"
}

type Drc20CollectAddress struct {
	Tick          string    `json:"tick"`
	AmtSum        *Number   `json:"amt_sum"`
	LockAmt       *Number   `json:"lock_amt"`
	Max           *Number   `gorm:"column:max_" json:"max"`
	Lim           *Number   `gorm:"column:lim_" json:"lim"`
	Dec           uint      `gorm:"column:dec_" json:"dec"`
	Burn          string    `gorm:"column:burn_" json:"burn"`
	Func          string    `gorm:"column:func_" json:"func"`
	HolderAddress string    `json:"holder_address"`
	Transactions  uint64    `json:"transactions"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
}

func (Drc20CollectAddress) TableName() string {
	return "drc20_collect_address"
}

type Drc20Revert struct {
	ID          uint      `gorm:"primarykey"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Tick        string    `json:"tick"`
	Amt         *Number   `json:"amt"`
	TxHash      string    `json:"tx_hash"`
	BlockNumber int64     `json:"block_number"`
	UpdateDate  LocalTime `json:"update_date"`
	CreateDate  LocalTime `json:"create_date"`
}

func (Drc20Revert) TableName() string {
	return "drc20_revert"
}

type Drc20CollectAll struct {
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

type Drc20CollectAllCache struct {
	Results     []*Drc20CollectAll
	Total       int64
	CacheNumber int64
}

type Drc20CollectRouter struct {
	Tick         string     `json:"tick"`
	MaxAmt       string     `json:"max_amt" gorm:"column:max_amt"`
	MintAmt      string     `json:"mint_amt" gorm:"column:mint_amt"`
	Dec          uint8      `json:"dec" gorm:"column:dec_"`
	Lim          string     `json:"lim" gorm:"column:lim_"`
	Holders      uint64     `json:"holders"`
	Transactions uint64     `json:"transactions"`
	DeployTime   LocalTime  `json:"deploy_time"`
	LastMintTime *LocalTime `json:"last_mint_time"`
	DeployBy     string     `json:"deploy_by"`
	Inscription  string     `json:"inscription"`
	Logo         *string    `json:"logo"`
	Introduction *string    `json:"introduction"`
	WhitePaper   *string    `json:"white_paper"`
	Official     *string    `json:"official"`
	Telegram     *string    `json:"telegram"`
	Discorad     *string    `json:"discorad"`
	Twitter      *string    `json:"twitter"`
	Facebook     *string    `json:"facebook"`
	Github       *string    `json:"github"`
	IsCheck      uint64     `json:"is_check"`
}

type Drc20CollectCache struct {
	Results     []*Drc20CollectRouter
	Total       int64
	CacheNumber int64
}
