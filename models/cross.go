package models

// CrossInfo Cross
type CrossInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `json:"order_id"`
	Op            string    `json:"op"`
	Tick          string    `json:"tick"`
	Amt           *Number   `json:"amt"`
	Chain         string    `json:"chain"`
	AdminAddress  string    `json:"admin_address"`
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
	CreateDate    LocalTime `json:"create_date"`
}

func (CrossInfo) TableName() string {
	return "cross_info"
}

type CrossCollect struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Tick          string    `json:"tick"`
	AdminAddress  string    `json:"admin_address"`
	HolderAddress string    `json:"holder_address"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
}

func (CrossCollect) TableName() string {
	return "cross_collect"
}

type CrossRevert struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Op          string    `json:"op"`
	Tick        string    `json:"tick"`
	BlockNumber int64     `json:"block_number"`
	UpdateDate  LocalTime `json:"update_date"`
	CreateDate  LocalTime `json:"create_date"`
}

func (CrossRevert) TableName() string {
	return "cross_revert"
}

type CrossBotInfo struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	Amt             *Number   `json:"amt"`
	FromChain       string    `json:"from_chain"`
	FromToken       string    `json:"from_token"`
	FromAddress     string    `json:"from_address"`
	FromTxHash      string    `json:"from_tx_hash"`
	FromBlockNumber int64     `json:"from_block_number"`
	FromBlockHash   string    `json:"from_block_hash"`
	ToChain         string    `json:"to_chain"`
	ToToken         string    `json:"to_token"`
	ToAddress       string    `json:"to_address"`
	ToTxHash        string    `json:"to_tx_hash"`
	ToTxIndex       int64     `json:"to_tx_index"`
	ToBlockNumber   int64     `json:"to_block_number"`
	ToBlockHash     string    `json:"to_block_hash"`
	ErrInfo         string    `json:"err_info"`
	OrderStatus     int64     `json:"order_status"`
	UpdateDate      LocalTime `json:"update_date"`
	CreateDate      LocalTime `json:"create_date"`
}

func (CrossBotInfo) TableName() string {
	return "cross_bot_info"
}
