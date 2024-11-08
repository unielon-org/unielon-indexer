package models

import "math/big"

// stake model
type StakeInfo struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	OrderId          string         `json:"order_id"`
	Op               string         `json:"op"`
	Tick             string         `json:"tick"`
	Amt              *Number        `json:"amt"`
	FeeTxHash        string         `json:"fee_tx_hash"`
	TxHash           string         `json:"tx_hash"`
	BlockHash        string         `json:"block_hash"`
	BlockNumber      int64          `json:"block_number"`
	FeeAddress       string         `json:"fee_address"`
	HolderAddress    string         `json:"holder_address"`
	ErrInfo          *string        `json:"err_info"`
	StakeRewardInfos []*StakeRevert `gorm:"-" json:"stake_reward_infos"`
	OrderStatus      int64          `json:"order_status"`
	UpdateDate       LocalTime      `json:"update_date"`
	CreateDate       LocalTime      `json:"create_date"`
}

func (StakeInfo) TableName() string {
	return "stake_info"
}

type StakeCollect struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	Tick            string    `json:"tick"`
	Amt             *Number   `json:"amt"`
	Reward          *Number   `json:"reward"`
	ReservesAddress string    `json:"reserves_address"`
	Holders         int64     `json:"holders"`
	UpdateDate      LocalTime `json:"update_date"`
	CreateDate      LocalTime `json:"create_date"`
}

func (StakeCollect) TableName() string {
	return "stake_collect"
}

type StakeCollectAddress struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	Tick           string    `json:"tick"`
	Amt            *Number   `json:"amt"`
	Reward         *Number   `json:"reward"`
	ReceivedReward *Number   `json:"received_reward"`
	HolderAddress  string    `json:"holder_address"`
	CardiAmt       *big.Int  `gorm:"-" json:"cardi_amt"`
	UpdateDate     LocalTime `json:"update_date"`
	CreateDate     LocalTime `json:"create_date"`
}

func (StakeCollectAddress) TableName() string {
	return "stake_collect_address"
}

type StakeCollectReward struct {
	Tick       string    `json:"tick"`
	RewardTick string    `json:"reward_tick"`
	Reward     *Number   `json:"reward"`
	UpdateDate LocalTime `json:"update_date"`
	CreateDate LocalTime `json:"create_date"`
}

func (StakeCollectReward) TableName() string {
	return "stake_collect_reward"
}

type StakeRevert struct {
	Tick        string  `json:"tick"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amt         *Number `json:"amt"`
	TxHash      string  `json:"tx_hash"`
	BlockNumber int64   `json:"block_number"`
}

func (StakeRevert) TableName() string {
	return "stake_revert"
}

type StakeRewardInfo struct {
	ID          uint      `gorm:"primarykey"`
	OrderId     string    `json:"order_id"`
	Tick        string    `json:"tick"`
	Amt         *Number   `json:"amt"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	BlockNumber int64     `json:"block_number"`
	UpdateDate  LocalTime `json:"update_date"`
	CreateDate  LocalTime `json:"create_date"`
}

func (StakeRewardInfo) TableName() string {
	return "stake_reward_info"
}

type StakeRewardRevert struct {
	Tick        string  `json:"tick"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amt         *Number `json:"amt"`
	TxHash      string  `json:"tx_hash"`
	BlockNumber int64   `json:"block_number"`
}

func (StakeRewardRevert) TableName() string {
	return "stake_reward_revert"
}

type HolderReward struct {
	Tick            string   `json:"tick"`
	TotalRewardPool *big.Int `json:"total_reward_pool"`
	Reward          *big.Int `json:"reward"`
	ReceivedReward  *big.Int `json:"received_reward"`
}
