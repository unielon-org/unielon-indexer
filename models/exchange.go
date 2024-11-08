package models

type ExchangeInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `json:"order_id"`
	Op            string    `json:"op"`
	ExId          string    `json:"ex_id"`
	Tick0         string    `json:"tick0"`
	Tick1         string    `json:"tick1"`
	Amt0          *Number   `json:"amt0"`
	Amt1          *Number   `json:"amt1"`
	FeeAddress    string    `json:"fee_address"`
	FeeTxHash     string    `json:"fee_tx_hash"`
	TxHash        string    `json:"tx_hash"`
	BlockNumber   int64     `json:"block_number"`
	BlockHash     string    `json:"block_hash"`
	HolderAddress string    `json:"holder_address"`
	ErrInfo       string    `json:"err_info"`
	OrderStatus   int64     `json:"order_status"`
	CreateDate    LocalTime `json:"create_date"`
	UpdateDate    LocalTime `json:"update_date"`
}

func (ExchangeInfo) TableName() string {
	return "exchange_info"
}

type ExchangeCollect struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	ExId            string    `json:"ex_id"`
	Tick0           string    `json:"tick0"`
	Tick1           string    `json:"tick1"`
	Amt0            *Number   `json:"amt0"`
	Amt1            *Number   `json:"amt1"`
	Amt0Finish      *Number   `json:"amt0_finish"`
	Amt1Finish      *Number   `json:"amt1_finish"`
	HolderAddress   string    `json:"holder_address"`
	ReservesAddress string    `json:"reserves_address"`
	CreateDate      LocalTime `json:"create_date"`
	UpdateDate      LocalTime `json:"update_date"`
}

func (ExchangeCollect) TableName() string {
	return "exchange_collect"
}

type ExchangeRevert struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Op          string    `json:"op"`
	Tick        string    `json:"tick"`
	ExId        string    `json:"ex_id"`
	Amt0        *Number   `json:"amt0"`
	Amt1        *Number   `json:"amt1"`
	BlockNumber int64     `json:"block_number"`
	TxHash      string    `json:"tx_hash"`
	UpdateDate  LocalTime `json:"update_date"`
	CreateDate  LocalTime `json:"create_date"`
}

func (ExchangeRevert) TableName() string {
	return "exchange_revert"
}

type ExchangeSummary struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Tick         string    `json:"tick"`
	Tick0        string    `json:"tick0"`
	Tick1        string    `json:"tick1"`
	OpenPrice    float64   `json:"open_price"`
	ClosePrice   float64   `json:"close_price"`
	LowestAsk    float64   `json:"lowest_ask"`
	HighestBid   float64   `json:"highest_bid"`
	BaseVolume   *Number   `json:"base_volume"`
	QuoteVolume  *Number   `json:"quote_volume"`
	LastDate     string    `json:"last_date"`
	DateInterval string    `json:"date_interval"`
	UpdateDate   LocalTime `json:"update_date"`
	CreateDate   LocalTime `json:"create_date"`
}

func (ExchangeSummary) TableName() string {
	return "exchange_summary"
}
