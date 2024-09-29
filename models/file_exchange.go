package models

type FileExchangeInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `gorm:"column:order_id" json:"order_id"`
	Op            string    `gorm:"column:op" json:"op"`
	ExId          string    `gorm:"column:ex_id" json:"ex_id"`
	FileId        string    `gorm:"column:file_id" json:"file_id"`
	Tick          string    `gorm:"column:tick" json:"tick"`
	Amt           *Number   `gorm:"column:amt" json:"amt"`
	HolderAddress string    `gorm:"column:holder_address" json:"holder_address"`
	FeeAddress    string    `gorm:"column:fee_address" json:"fee_address"`
	FeeTxHash     string    `gorm:"column:fee_tx_hash" json:"fee_tx_hash"`
	TxHash        string    `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber   int64     `gorm:"column:block_number" json:"block_number"`
	BlockHash     string    `gorm:"column:block_hash" json:"block_hash"`
	ErrInfo       string    `gorm:"column:err_info" json:"err_info"`
	OrderStatus   int64     `gorm:"column:order_status" json:"order_status"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`

	// add
	IsNft    int64  `gorm:"->" json:"is_nft"`
	FileName string `gorm:"->" json:"file_name"`
	MetaName string `gorm:"->" json:"meta_name"`
	FilePath string `gorm:"->" json:"file_path"`
}

func (FileExchangeInfo) TableName() string {
	return "file_exchange_info"
}

type FileExchangeCollect struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	ExId            string    `gorm:"column:ex_id" json:"exid"`
	FileId          string    `gorm:"column:file_id" json:"file_id"`
	Tick            string    `gorm:"column:tick" json:"tick"`
	Amt             *Number   `gorm:"column:amt" json:"amt"`
	AmtFinish       *Number   `gorm:"column:amt_finish" json:"amt_finish"`
	HolderAddress   string    `json:"holder_address"`
	ReservesAddress string    `json:"reserves_address"`
	IsNft           int64     `json:"is_nft"`
	UpdateDate      LocalTime `json:"update_date"`
	CreateDate      LocalTime `json:"create_date"`
}

func (FileExchangeCollect) TableName() string {
	return "file_exchange_collect"
}

type FileExchangeRevert struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Op          string    `json:"op"`
	ExId        string    `gorm:"column:ex_id" json:"exid"`
	FileId      string    `gorm:"column:file_id" json:"file_id"`
	Tick        string    `gorm:"column:tick" json:"tick"`
	Amt         *Number   `gorm:"column:amt" json:"amt"`
	AmtFinish   *Number   `gorm:"column:amt_finish" json:"amt_finish"`
	BlockNumber int64     `json:"block_number"`
	IsNft       int64     `json:"is_nft"`
	UpdateDate  LocalTime `json:"update_date"`
	CreateDate  LocalTime `json:"create_date"`
}

func (FileExchangeRevert) TableName() string {
	return "file_exchange_revert"
}

type FileExchangeSummary struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	MetaName     string    `json:"meta_name"`
	LowestAsk    float64   `json:"lowest_ask"`
	HighestBid   float64   `json:"highest_bid"`
	BaseVolume   *Number   `json:"base_volume"`
	LastDate     string    `json:"last_date"`
	DateInterval string    `json:"date_interval"`
	DogeUsdt     float64   `json:"doge_usdt"`
	UpdateDate   LocalTime `json:"update_date"`
	CreateDate   LocalTime `json:"create_date"`
}

func (FileExchangeSummary) TableName() string {
	return "file_exchange_summary"
}
