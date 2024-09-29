package models

// WDOGE
type WDogeInfo struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	OrderId             string    `json:"order_id"`
	Op                  string    `json:"op"`
	Tick                string    `json:"tick"`
	Amt                 *Number   `json:"amt"`
	HolderAddress       string    `json:"holder_address"`
	FeeAddress          string    `json:"fee_address"`
	FeeTxHash           string    `json:"fee_tx_hash"`
	TxHash              string    `json:"tx_hash"`
	BlockNumber         int64     `json:"block_number"`
	BlockHash           string    `json:"block_hash"`
	WithdrawTxHash      string    `json:"withdraw_tx_hash"`
	WithdrawTxIndex     uint32    `json:"withdraw_tx_index"`
	WithdrawTxRaw       string    `json:"withdraw_tx_raw"`
	WithdrawBlockNumber int64     `json:"withdraw_block_number"`
	WithdrawBlockHash   string    `json:"withdraw_block_hash"`
	ErrInfo             string    `json:"err_info"`
	OrderStatus         int64     `json:"order_status"`
	UpdateDate          LocalTime `json:"update_date"`
	CreateDate          LocalTime `json:"create_date"`
}

func (WDogeInfo) TableName() string {
	return "wdoge_info"
}
