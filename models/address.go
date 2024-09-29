package models

import "github.com/dogecoinw/doged/btcutil"

// model
type AddressInfo struct {
	OrderId        string       `json:"order_id"`
	PrveWif        *btcutil.WIF `json:"prve_wif"`
	PubKey         string       `json:"pub_key"`
	Address        string       `json:"address"`
	ReceiveAddress string       `json:"receive_address"`
	FeeAddress     string       `json:"fee_address"`
}

func (AddressInfo) TableName() string {
	return "address_info"
}
