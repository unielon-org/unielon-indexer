package utils

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

func ConvetCard(params *NewParams) (*Cardinals, error) {
	card := &Cardinals{
		P:              params.P,
		Op:             params.Op,
		Tick:           strings.ToUpper(params.Tick),
		Dec:            params.Dec,
		Burn:           params.Burn,
		Func:           params.Func,
		ReceiveAddress: params.ReceiveAddress,
		ToAddress:      params.ToAddress,
	}

	if params.Dec == 0 {
		card.Dec = 8
	}

	amt, err := ConvetStr(params.Amt)
	if err != nil {
		return nil, err
	}
	card.Amt = amt

	max, err := ConvetStr(params.Max)
	if err != nil {
		return nil, err
	}
	card.Max = max

	lim, err := ConvetStr(params.Lim)
	if err != nil {
		return nil, err
	}
	card.Lim = lim
	card.Repeat = params.Repeat

	return card, nil
}

func ConvetSwap(params *SwapParams) (*SwapInfo, error) {
	swap := &SwapInfo{
		Op:            params.Op,
		Tick0:         strings.ToUpper(params.Tick0),
		Tick1:         strings.ToUpper(params.Tick1),
		HolderAddress: params.HolderAddress,
	}

	var err error
	swap.Amt0, err = ConvetStr(params.Amt0)
	if err != nil {
		return nil, err
	}

	swap.Amt1, err = ConvetStr(params.Amt1)
	if err != nil {
		return nil, err
	}

	swap.Liquidity = big.NewInt(0)
	swap.Amt0Min = big.NewInt(0)
	swap.Amt1Min = big.NewInt(0)

	if swap.Op == "swap" {
		swap.Amt1Min, err = ConvetStr(params.Amt1Min)
		if err != nil {
			return nil, err
		}
	}

	if swap.Op == "create" || swap.Op == "add" {
		swap.Amt0Min, err = ConvetStr(params.Amt0Min)
		if err != nil {
			return nil, err
		}
		swap.Amt1Min, err = ConvetStr(params.Amt1Min)
		if err != nil {
			return nil, err
		}
	}

	if swap.Op == "remove" {
		swap.Liquidity, err = ConvetStr(params.Liquidity)
		if err != nil {
			return nil, err
		}
	}

	if swap.Op == "create" || swap.Op == "add" || swap.Op == "remove" {
		swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)
	}

	return swap, nil
}

func ConvertWDoge(params *WDogeParams) (*WDogeInfo, error) {
	swap := &WDogeInfo{
		Op:            params.Op,
		Tick:          strings.ToUpper(params.Tick),
		HolderAddress: params.HolderAddress,
	}

	amt_big, err := ConvetStr(params.Amt)
	if err != nil {
		return nil, err
	}
	swap.Amt = amt_big

	return swap, nil
}

func SortTokens(Tick0 string, Tick1 string, Amt0, Amt1, Amt0Min, Amt1Min *big.Int) (string, string, *big.Int, *big.Int, *big.Int, *big.Int) {
	if Tick0 > Tick1 {
		return Tick1, Tick0, Amt1, Amt0, Amt1Min, Amt0Min
	}
	return Tick0, Tick1, Amt0, Amt1, Amt0Min, Amt1Min
}

func ConvetStr(number string) (*big.Int, error) {
	if number != "" {
		max_big, is_ok := new(big.Int).SetString(number, 10)
		if !is_ok {
			return big.NewInt(0), fmt.Errorf("number error")
		}
		return max_big, nil
	}

	return big.NewInt(0), nil
}

func Float64ToBigInt(input float64) *big.Int {

	rounded := math.Ceil(input)

	if rounded < math.MinInt64 || rounded > math.MaxInt64 {
		return big.NewInt(0)
	}

	result := int64(rounded)
	return big.NewInt(result)
}
