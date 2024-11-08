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
	ID                uint      `gorm:"primarykey" json:"id"`
	StakeId           string    `json:"stake_id"`
	Tick0             string    `json:"tick0"`
	Tick1             string    `json:"tick1"`
	TotalStaked       *Number   `json:"total_staked"`
	Reward            *Number   `json:"reward"`
	RewardFinish      *Number   `json:"reward_finish"`
	EachReward        *Number   `json:"each_reward"`
	AccRewardPerShare *Number   `json:"acc_reward_per_share"`
	LastRewardBlock   int64     `json:"last_reward_block"`
	ReservesAddress   string    `json:"reserves_address"`
	UpdateDate        LocalTime `json:"update_date"`
	CreateDate        LocalTime `json:"create_date"`
}

func (StakeV2Collect) TableName() string {
	return "stake_v2_collect"
}

type StakeV2CollectAddress struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	StakeId       string    `json:"stake_id"`
	Tick          string    `json:"tick"`
	Amt           *Number   `json:"amt"`
	RewardDebt    *Number   `json:"reward_debt"`
	PendingReward *Number   `json:"pending_reward"`
	HolderAddress string    `json:"holder_address"`
	CardiAmt      *big.Int  `gorm:"-" json:"cardi_amt"`
	UpdateDate    LocalTime `json:"update_date"`
	CreateDate    LocalTime `json:"create_date"`
}

func (StakeV2CollectAddress) TableName() string {
	return "stake_v2_collect_address"
}

type StakeV2Revert struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	Op                string    `json:"op"`
	StakeId           string    `json:"stake_id"`
	Tick              string    `json:"tick"`
	Amt               *Number   `json:"amt"`
	RewardDebt        *Number   `json:"reward_debt"`
	PendingReward     *Number   `json:"pending_reward"`
	AccRewardPerShare *Number   `json:"acc_reward_per_share"`
	LastRewardBlock   int64     `json:"last_reward_block"`
	LastBlock         int64     `json:"last_block"`
	HolderAddress     string    `json:"holder_address"`
	ToAddress         string    `json:"to_address"`
	BlockNumber       int64     `json:"block_number"`
	UpdateDate        LocalTime `json:"update_date"`
	CreateDate        LocalTime `json:"create_date"`
}

func (StakeV2Revert) TableName() string {
	return "stake_v2_revert"
}
