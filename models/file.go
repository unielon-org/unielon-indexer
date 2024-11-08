package models

type FileInfo struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `gorm:"column:order_id" json:"order_id"`
	FileId        string    `gorm:"column:file_id" json:"file_id"`
	Op            string    `gorm:"column:op" json:"op"`
	File          string    `gorm:"column:file" json:"file"`
	FilePath      string    `gorm:"column:file_path" json:"file_path"`
	FileData      []byte    `gorm:"-" json:"file_data"`
	FileLength    int       `gorm:"column:file_length" json:"file_length"`
	FileType      string    `gorm:"column:file_type" json:"file_type"`
	HolderAddress string    `gorm:"column:holder_address" json:"holder_address"`
	ToAddress     string    `gorm:"column:to_address" json:"to_address"`
	FeeAddress    string    `gorm:"column:fee_address" json:"fee_address"`
	FeeTxHash     string    `gorm:"column:fee_tx_hash" json:"fee_tx_hash"`
	TxHash        string    `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber   int64     `gorm:"column:block_number" json:"block_number"`
	BlockHash     string    `gorm:"column:block_hash" json:"block_hash"`
	ErrInfo       string    `gorm:"column:err_info" json:"err_info"`
	OrderStatus   int64     `gorm:"column:order_status" json:"order_status"`
	UpdateDate    LocalTime `gorm:"column:update_date" json:"update_date"`
	CreateDate    LocalTime `gorm:"column:create_date" json:"create_date"`
}

func (FileInfo) TableName() string {
	return "file_info"
}

type FileCollectAddress struct {
	ID            uint      `gorm:"primarykey"`
	FileId        string    `gorm:"column:file_id" json:"file_id"`
	File          string    `gorm:"column:file" json:"file"`
	FilePath      string    `gorm:"column:file_path" json:"file_path"`
	FileLength    int       `gorm:"column:file_length" json:"file_length"`
	FileType      string    `gorm:"column:file_type" json:"file_type"`
	HolderAddress string    `gorm:"column:holder_address" json:"holder_address"`
	UpdateDate    LocalTime `gorm:"column:update_date" json:"update_date"`
	CreateDate    LocalTime `gorm:"column:create_date" json:"create_date"`
}

func (FileCollectAddress) TableName() string {
	return "file_collect_address"
}

type FileRevert struct {
	ID          uint      `gorm:"primarykey"`
	FromAddress string    `gorm:"column:from_address" json:"from_address"`
	ToAddress   string    `gorm:"column:to_address" json:"to_address"`
	FileId      string    `gorm:"column:file_id" json:"file_id"`
	BlockNumber int64     `gorm:"column:block_number" json:"block_number"`
	TxHash      string    `gorm:"column:tx_hash" json:"tx_hash"`
	UpdateDate  LocalTime `gorm:"column:update_date" json:"update_date"`
	CreateDate  LocalTime `gorm:"column:create_date" json:"create_date"`
}

func (FileRevert) TableName() string {
	return "file_revert"
}

type FileMeta struct {
	ID            uint      `gorm:"primarykey" json:"id,omitempty"`
	MetaId        string    `gorm:"uniqueIndex" json:"meta_id"`
	Description   string    `json:"description"`
	DiscordLink   string    `json:"discord_link"`
	Icon          string    `json:"icon"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	TwitterLink   string    `json:"twitter_link"`
	WebsiteLink   string    `json:"website_link"`
	HolderAddress string    `json:"holder_address"`
	IsCheck       int       `json:"is_check"`
	UpdateDate    LocalTime `gorm:"column:update_date" json:"update_date"`
	CreateDate    LocalTime `gorm:"column:create_date" json:"create_date"`
}

func (FileMeta) TableName() string {
	return "file_meta"
}

type FileMetaInscription struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	MetaId string `json:"meta_id"`
	FileId string `json:"file_id"`
	Name   string `json:"name"`
}

func (FileMetaInscription) TableName() string {
	return "file_meta_inscription"
}

type FileMetaAttribute struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	MetaId    string `json:"meta_id"`
	FileId    string `json:"file_id"`
	Name      string `json:"name"`
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

func (FileMetaAttribute) TableName() string {
	return "file_meta_attribute"
}
