package models

import "math/big"

type StakeV2Info struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OrderId       string    `json:"order_id"`
	Op            string    `json:"op"`
	StakeId       string    `json:"stake_id"`
	Tick0         string    `json:"tick0"`
	Tick1         string    `json:"tick1"`
	Reward        *Number   `json:"reward"`
	EachReward    *Number   `json:"each_reward"`
	Amt           *Number   `json:"amt"`
	FeeTxHash     string    `json:"fee_tx_hash"`
	TxHash        string    `json:"tx_hash"`
	BlockHash     string    `json:"block_hash"`
	BlockNumber   int64     `json:"block_number"`
	FeeAddress    string    `json:"fee_address"`
	HolderAddress string    `json:"holder_address"`
	ErrInfo       *string   `json:"err_info"`
	OrderStatus   int64     `json:"order_status"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
}

func (StakeV2Info) TableName() string {
	return "stake_v2_info"
}

type StakeV2Collect struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	StakeId         string    `json:"stake_id"`
	Tick0           string    `json:"tick0"`
	Tick1           string    `json:"tick1"`
	Amt             *Number   `json:"amt"`
	Reward          *Number   `json:"reward"`
	EachReward      *Number   `json:"each_reward"`
	ReservesAddress string    `json:"reserves_address"`
	UpdateDate      LocalTime `json:"update_date"`
	CreateDate      LocalTime `json:"create_date"`
}

func (StakeV2Collect) TableName() string {
	return "stake_v2_collect"
}

type StakeCollectAddress struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	StakeId        string    `json:"stake_id"`
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

type StakeV2Revert struct {
	ID          uint    `gorm:"primarykey" json:"id"`
	Op          string  `json:"op"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Tick        string  `json:"tick"`
	Amt         *Number `json:"amt"`
	BlockNumber int64   `json:"block_number"`
}

func (StakeV2Revert) TableName() string {
	return "stake_revert"
}

type StakeV2RewardInfo struct {
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

func (StakeV2RewardInfo) TableName() string {
	return "stake_reward_info"
}

type StakeV2RewardRevert struct {
	Tick        string  `json:"tick"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amt         *Number `json:"amt"`
	BlockNumber int64   `json:"block_number"`
}

func (StakeV2RewardRevert) TableName() string {
	return "stake_reward_revert"
}

type HolderV2Reward struct {
	Tick            string   `json:"tick"`
	TotalRewardPool *big.Int `json:"total_reward_pool"`
	Reward          *big.Int `json:"reward"`
	ReceivedReward  *big.Int `json:"received_reward"`
}
