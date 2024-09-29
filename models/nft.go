package models

// AINFT
type NftInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `json:"order_id"`
	Op            string    `json:"op"`
	Tick          string    `json:"tick"`
	TickId        int64     `json:"tick_id"`
	Total         int64     `json:"total"`
	Model         string    `json:"model"`
	Prompt        string    `json:"prompt"`
	Seed          int64     `json:"seed"`
	Image         string    `gorm:"-" json:"image"`
	ImageData     []byte    `gorm:"-" json:"image_data"`
	ImagePath     string    `json:"image_path"`
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

func (NftInfo) TableName() string {
	return "nft_info"
}

type NftCollect struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Tick          string    `json:"tick"`
	TickSum       int64     `json:"tick_sum"`
	Total         int64     `json:"total"`
	Model         string    `json:"model"`
	Prompt        string    `json:"prompt"`
	Image         string    `json:"image"`
	ImagePath     string    `json:"image_path"`
	HolderAddress string    `json:"holder_address"`
	DeployHash    string    `json:"deploy_hash"`
	Transactions  int64     `json:"transactions"`
	Holders       int64     `gorm:"-" json:"holders"`
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

func (NftCollect) TableName() string {
	return "nft_collect"
}

type NftCollectAddress struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Tick          string    `json:"tick"`
	TickId        int64     `json:"tick_id"`
	Prompt        string    `json:"prompt"`
	NftPrompt     string    `gorm:"-" json:"nft_prompt"`
	NftModel      string    `gorm:"-" json:"nft_model"`
	Image         string    `gorm:"-" json:"image"`
	ImagePath     string    `json:"image_path"`
	DeployHash    string    `json:"deploy_hash"`
	HolderAddress string    `json:"holder_address"`
	Transactions  int64     `json:"transactions"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
	IsCheck       int64     `json:"is_check"`
}

func (NftCollectAddress) TableName() string {
	return "nft_collect_address"
}

type NftRevert struct {
	ID          uint   `gorm:"primarykey"`
	Tick        string `json:"tick"`
	TickId      int64  `json:"tick_id"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	BlockNumber int64  `json:"block_number"`
}

func (NftRevert) TableName() string {
	return "nft_revert"
}
