package models

// box
type BoxInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `json:"order_id"`
	Op            string    `json:"op"`
	Tick0         string    `json:"tick0"`
	Tick1         string    `json:"tick1"`
	Max           *Number   `gorm:"column:max_" json:"max"`
	Amt0          *Number   `json:"amt0"`
	LiqAmt        *Number   `gorm:"column:liqamt" json:"liqamt"`
	LiqBlock      int64     `gorm:"column:liqblock" json:"liqblock"`
	Amt1          *Number   `json:"amt1"`
	FeeAddress    string    `json:"fee_address"`
	FeeTxHash     string    `json:"fee_tx_hash"`
	TxHash        string    `json:"tx_hash"`
	BlockNumber   int64     `json:"block_number"`
	BlockHash     string    `json:"block_hash"`
	HolderAddress string    `json:"holder_address"`
	OrderStatus   int64     `json:"order_status"`
	ErrInfo       string    `json:"err_info"`
	CreateDate    LocalTime `json:"create_date"`
	UpdateDate    LocalTime `json:"update_date"`
}

func (BoxInfo) TableName() string {
	return "box_info"
}

type BoxCollect struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	Tick0           string    `json:"tick0"`
	Tick1           string    `json:"tick1"`
	Max             *Number   `gorm:"column:max_" json:"max"`
	Amt0            *Number   `json:"amt0"`
	LiqAmt          *Number   `gorm:"column:liqamt" json:"liqamt"`
	LiqBlock        int64     `gorm:"column:liqblock" json:"liqblock"`
	Amt1            *Number   `json:"amt1"`
	Amt0Finish      *Number   `gorm:"column:amt0_finish" json:"amt0_finish"`
	LiqAmtFinish    *Number   `gorm:"column:liqamt_finish" json:"liqamt_finish"`
	HolderAddress   string    `json:"holder_address"`
	ReservesAddress string    `json:"reserves_address"`
	IsDel           int64     `json:"is_del"`
	CreateDate      LocalTime `json:"create_date"`
	UpdateDate      LocalTime `json:"update_date"`
}

func (BoxCollect) TableName() string {
	return "box_collect"
}

type BoxCollectAddress struct {
	Id            int64     `json:"id"`
	Tick          string    `json:"tick"`
	HolderAddress string    `json:"holder_address"`
	Amt           *Number   `json:"amt"`
	BlockNumber   int64     `json:"block_number"`
	CreateDate    LocalTime `json:"create_date"`
}

func (BoxCollectAddress) TableName() string {
	return "box_collect_address"
}

type BoxRevert struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Op            string    `json:"op"`
	Tick0         string    `json:"tick0"`
	Tick1         string    `json:"tick1"`
	Max           *Number   `gorm:"column:max_" json:"max"`
	HolderAddress string    `json:"holder_address"`
	TxHash        string    `json:"tx_hash"`
	BlockNumber   int64     `json:"block_number"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
}

func (BoxRevert) TableName() string {
	return "box_revert"
}
