package models

type Block struct {
	BlockNumber int64  `gorm:"primarykey" json:"block_number"`
	BlockHash   string `json:"block_hash"`
}

func (Block) TableName() string {
	return "block"
}
