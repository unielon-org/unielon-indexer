package utils

import (
	"github.com/unielon-org/unielon-indexer/models"
	"strings"
)

func ConvetCard(inscription *models.Drc20Inscription) (*models.Drc20Info, error) {

	card := &models.Drc20Info{
		P:    inscription.P,
		Op:   inscription.Op,
		Tick: strings.ToUpper(inscription.Tick),
		Dec:  inscription.Dec,
		Burn: inscription.Burn,
		Func: inscription.Func,
	}

	if inscription.Dec == 0 {
		card.Dec = 8
	}

	amt, err := ConvetStringToNumber(inscription.Amt)
	if err != nil {
		return nil, err
	}
	card.Amt = amt

	max, err := ConvetStringToNumber(inscription.Max)
	if err != nil {
		return nil, err
	}
	card.Max = max

	lim, err := ConvetStringToNumber(inscription.Lim)
	if err != nil {
		return nil, err
	}
	card.Lim = lim

	return card, nil
}

func ConvetSwap(inscription *models.SwapInscription) (*models.SwapInfo, error) {
	swap := &models.SwapInfo{
		Op:    inscription.Op,
		Tick0: strings.ToUpper(inscription.Tick0),
		Tick1: strings.ToUpper(inscription.Tick1),
	}

	var err error
	swap.Amt0, err = ConvetStringToNumber(inscription.Amt0)
	if err != nil {
		return nil, err
	}

	swap.Amt1, err = ConvetStringToNumber(inscription.Amt1)
	if err != nil {
		return nil, err
	}

	swap.Liquidity = models.NewNumber(0)
	swap.Amt0Min = models.NewNumber(0)
	swap.Amt1Min = models.NewNumber(0)

	if swap.Op == "swap" {
		swap.Amt1Min, err = ConvetStringToNumber(inscription.Amt1Min)
		if err != nil {
			return nil, err
		}
	}

	if swap.Op == "create" || swap.Op == "add" {
		swap.Amt0Min, err = ConvetStringToNumber(inscription.Amt0Min)
		if err != nil {
			return nil, err
		}
		swap.Amt1Min, err = ConvetStringToNumber(inscription.Amt1Min)
		if err != nil {
			return nil, err
		}
	}

	if swap.Op == "remove" {
		swap.Liquidity, err = ConvetStringToNumber(inscription.Liquidity)
		if err != nil {
			return nil, err
		}
	}

	if swap.Op == "create" || swap.Op == "add" || swap.Op == "remove" {
		swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)
	}

	return swap, nil
}

func ConvertWDoge(inscription *models.WDogeInscription) (*models.WDogeInfo, error) {
	swap := &models.WDogeInfo{
		Op:            inscription.Op,
		Tick:          strings.ToUpper(inscription.Tick),
		HolderAddress: inscription.HolderAddress,
	}

	var err error
	swap.Amt, err = ConvetStringToNumber(inscription.Amt)
	if err != nil {
		return nil, err
	}

	return swap, nil
}

func ConvertNft(inscription *models.NftInscription) (*models.NftInfo, error) {
	nft := &models.NftInfo{
		Op:     inscription.Op,
		Tick:   strings.ToUpper(inscription.Tick),
		TickId: inscription.TickId,
		Total:  inscription.Total,
		Model:  inscription.Model,
		Prompt: inscription.Prompt,
		Image:  inscription.Image,
		Seed:   inscription.Seed,
	}

	if inscription.Op != "transfer" {
		var err error
		nft.ImageData, err = Base64ToPng(nft.Image)
		if err != nil {
			return nil, err
		}
	}

	return nft, nil
}

func ConvertFile(inscription *models.FileInscription) (*models.FileInfo, error) {
	file := &models.FileInfo{
		Op:     inscription.Op,
		FileId: inscription.FileId,
	}

	if inscription.Op != "transfer" {
		var err error
		file.FileData, err = Base64ToPng(string(inscription.File))
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

func ConvertStake(inscription *models.StakeInscription) (*models.StakeInfo, error) {
	stake := &models.StakeInfo{
		Op:   inscription.Op,
		Tick: strings.ToUpper(inscription.Tick),
	}

	var err error
	stake.Amt, err = ConvetStringToNumber(inscription.Amt)
	if err != nil {
		return nil, err
	}

	return stake, nil
}

func ConvertStakeV2(inscription *models.StakeV2Inscription) (*models.StakeV2Info, error) {
	stake := &models.StakeV2Info{
		Op:      inscription.Op,
		Tick0:   strings.ToUpper(inscription.Tick0),
		Tick1:   strings.ToUpper(inscription.Tick1),
		StakeId: inscription.StakeId,
	}

	var err error
	stake.Amt, err = ConvetStringToNumber(inscription.Amt)
	if err != nil {
		return nil, err
	}

	stake.Reward, err = ConvetStringToNumber(inscription.Reward)
	if err != nil {
		return nil, err
	}

	stake.EachReward, err = ConvetStringToNumber(inscription.EachReward)
	if err != nil {
		return nil, err
	}

	return stake, nil
}

func ConvertExChange(inscription *models.ExchangeInscription) (*models.ExchangeInfo, error) {
	ex := &models.ExchangeInfo{
		Op:    inscription.Op,
		Tick0: strings.ToUpper(inscription.Tick0),
		Tick1: strings.ToUpper(inscription.Tick1),
		ExId:  inscription.ExId,
	}

	var err error
	ex.Amt0, err = ConvetStringToNumber(inscription.Amt0)
	if err != nil {
		return nil, err
	}

	ex.Amt1, err = ConvetStringToNumber(inscription.Amt1)
	if err != nil {
		return nil, err
	}

	return ex, nil

}

func ConvertFileExchange(inscription *models.FileExchangeInscription) (*models.FileExchangeInfo, error) {
	ex := &models.FileExchangeInfo{
		Op:     inscription.Op,
		FileId: inscription.FileId,
		ExId:   inscription.ExId,
		Tick:   strings.ToUpper(inscription.Tick),
	}

	var err error
	ex.Amt, err = ConvetStringToNumber(inscription.Amt)
	if err != nil {
		return nil, err
	}

	return ex, nil
}

func ConvertBox(inscription *models.BoxInscription) (*models.BoxInfo, error) {
	box := &models.BoxInfo{
		Op:    inscription.Op,
		Tick0: strings.ToUpper(inscription.Tick0),
		Tick1: strings.ToUpper(inscription.Tick1),
	}

	var err error
	box.Amt0, err = ConvetStringToNumber(inscription.Amt0)
	if err != nil {
		return nil, err
	}

	box.LiqAmt, err = ConvetStringToNumber(inscription.LiqAmt)
	if err != nil {
		return nil, err
	}

	box.Max, err = ConvetStringToNumber(inscription.Max)
	if err != nil {
		return nil, err
	}

	box.Amt1, err = ConvetStringToNumber(inscription.Amt1)
	if err != nil {
		return nil, err
	}
	return box, nil

}
