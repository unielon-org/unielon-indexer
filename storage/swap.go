package storage

import (
	"fmt"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
	"math/big"
)

const (
	MINI_LIQUIDITY = 1000
)

func (e *DBClient) SwapCreate(tx *gorm.DB, swap *models.SwapInfo) error {

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(swap.Tick0+swap.Tick1), &chaincfg.MainNetParams)
	swap.Tick = swap.Tick0 + "-SWAP-" + swap.Tick1

	liquidityBase := new(big.Int).Sqrt(new(big.Int).Mul(swap.Amt0.Int(), swap.Amt1.Int()))
	if liquidityBase.Cmp(big.NewInt(MINI_LIQUIDITY)) > 0 {
		liquidityBase = new(big.Int).Sub(liquidityBase, big.NewInt(MINI_LIQUIDITY))
	} else {
		return fmt.Errorf("add liquidity must be greater than MINI_LIQUIDITY firstly")
	}

	swap.Amt0Out = swap.Amt0
	swap.Amt1Out = swap.Amt1
	swap.Liquidity = (*models.Number)(liquidityBase)

	sl := &models.SwapLiquidity{
		Tick:            swap.Tick,
		Tick0:           swap.Tick0,
		Tick1:           swap.Tick1,
		HolderAddress:   swap.HolderAddress,
		ReservesAddress: reservesAddress.String(),
		LiquidityTotal:  (*models.Number)(liquidityBase),
	}

	err := tx.Create(sl).Error
	if err != nil {
		return fmt.Errorf("SwapCreate Create err: %s", err.Error())
	}

	err = e.TransferDrc20(tx, swap.Tick0, swap.HolderAddress, reservesAddress.String(), swap.Amt0.Int(), swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.TransferDrc20(tx, swap.Tick1, swap.HolderAddress, reservesAddress.String(), swap.Amt1.Int(), swap.BlockNumber, false)
	if err != nil {
		return err
	}

	drc20c := &models.Drc20Collect{
		Tick:          swap.Tick,
		Max:           models.NewNumber(-1),
		Lim:           models.NewNumber(0),
		Dec:           8,
		HolderAddress: reservesAddress.String(),
		TxHash:        swap.TxHash,
	}

	err = tx.Create(drc20c).Error

	err = e.MintDrc20(tx, swap.Tick, swap.HolderAddress, liquidityBase, swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.MintDrc20(tx, swap.Tick, reservesAddress.String(), big.NewInt(MINI_LIQUIDITY), swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.UpdateLiquidity(tx, swap.Tick)
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) SwapAdd(tx *gorm.DB, swap *models.SwapInfo) error {

	reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(swap.Tick0+swap.Tick1), &chaincfg.MainNetParams)
	swap.Tick = swap.Tick0 + "-SWAP-" + swap.Tick1

	amt0Out := big.NewInt(0)
	amt1Out := big.NewInt(0)

	swapl := &models.SwapLiquidity{}
	err := tx.Where("tick0 = ? and tick1 = ?", swap.Tick0, swap.Tick1).First(swapl).Error
	if err != nil {
		return fmt.Errorf("SwapAdd Find error: %s", err.Error())
	}

	amountBOptimal := big.NewInt(0).Mul(swap.Amt0.Int(), swapl.Amt1.Int())
	amountBOptimal = big.NewInt(0).Div(amountBOptimal, swapl.Amt0.Int())
	if amountBOptimal.Cmp(swap.Amt1Min.Int()) >= 0 && swap.Amt1.Int().Cmp(amountBOptimal) >= 0 {
		amt0Out = swap.Amt0.Int()
		amt1Out = amountBOptimal
	} else {
		amountAOptimal := big.NewInt(0).Mul(swap.Amt1.Int(), swapl.Amt0.Int())
		amountAOptimal = big.NewInt(0).Div(amountAOptimal, swapl.Amt1.Int())
		if amountAOptimal.Cmp(swap.Amt0Min.Int()) >= 0 && swap.Amt0.Int().Cmp(amountAOptimal) >= 0 {
			amt0Out = amountAOptimal
			amt1Out = swap.Amt1.Int()
		} else {
			return fmt.Errorf("The amount of tokens exceeds the balance")
		}
	}

	liquidity0 := new(big.Int).Mul(amt0Out, swapl.LiquidityTotal.Int())
	liquidity0 = new(big.Int).Div(liquidity0, swapl.Amt0.Int())

	liquidity1 := new(big.Int).Mul(amt1Out, swapl.LiquidityTotal.Int())
	liquidity1 = new(big.Int).Div(liquidity1, swapl.Amt1.Int())

	liquidity := liquidity0
	if liquidity0.Cmp(liquidity1) > 0 {
		liquidity = liquidity1
	}

	err = e.TransferDrc20(tx, swap.Tick0, swap.HolderAddress, reservesAddress.String(), amt0Out, swap.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = e.TransferDrc20(tx, swap.Tick1, swap.HolderAddress, reservesAddress.String(), amt1Out, swap.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = e.MintDrc20(tx, swap.Tick, swap.HolderAddress, liquidity, swap.BlockNumber, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新 amt0
	err = e.UpdateLiquidity(tx, swap.Tick)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (e *DBClient) SwapRemove(tx *gorm.DB, swap *models.SwapInfo) error {

	swapl := &models.SwapLiquidity{}
	err := tx.Where("tick0 = ? and tick1 = ?", swap.Tick0, swap.Tick1).First(swapl).Error
	if err != nil {
		return fmt.Errorf("swapRemove FindSwapLiquidity error: %v", err)
	}

	amt0Out := new(big.Int).Mul(swap.Liquidity.Int(), swapl.Amt0.Int())
	amt0Out = new(big.Int).Div(amt0Out, swapl.LiquidityTotal.Int())

	amt1Out := new(big.Int).Mul(swap.Liquidity.Int(), swapl.Amt1.Int())
	amt1Out = new(big.Int).Div(amt1Out, swapl.LiquidityTotal.Int())

	if swapl.Amt0.Int().Cmp(amt0Out) < 0 || swapl.Amt1.Int().Cmp(amt1Out) < 0 {
		return fmt.Errorf("swapRemove FindSwapLiquidity error: %v", err)
	}

	err = e.TransferDrc20(tx, swap.Tick0, swapl.ReservesAddress, swap.HolderAddress, amt0Out, swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.TransferDrc20(tx, swap.Tick1, swapl.ReservesAddress, swap.HolderAddress, amt1Out, swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.BurnDrc20(tx, swapl.Tick, swap.HolderAddress, swap.Liquidity.Int(), swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.UpdateLiquidity(tx, swapl.Tick)
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) SwapExec(tx *gorm.DB, swap *models.SwapInfo) error {

	tick0, tick1, _, _, _, _ := utils.SortTokens(swap.Tick0, swap.Tick1, nil, nil, nil, nil)

	swapl := &models.SwapLiquidity{}
	err := tx.Where("tick0 = ? and tick1 = ?", tick0, tick1).First(swapl).Error
	if err != nil {
		return fmt.Errorf("swapExec FindSwapLiquidity error: %v", err)
	}

	amtMap := make(map[string]*big.Int)
	amtMap[swapl.Tick0] = swapl.Amt0.Int()
	amtMap[swapl.Tick1] = swapl.Amt1.Int()

	amtfee0 := new(big.Int).Div(swap.Amt0.Int(), big.NewInt(1000))
	amtin := new(big.Int).Mul(amtfee0, big.NewInt(3))
	amtin = new(big.Int).Sub(swap.Amt0.Int(), amtin)

	amtout := new(big.Int).Mul(amtin, amtMap[swap.Tick1])
	amtout = new(big.Int).Div(amtout, new(big.Int).Add(amtMap[swap.Tick0], amtin))

	swap.Amt1Out = (*models.Number)(amtout)

	err = e.TransferDrc20(tx, swap.Tick0, swap.HolderAddress, swapl.ReservesAddress, swap.Amt0.Int(), swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.TransferDrc20(tx, swap.Tick1, swapl.ReservesAddress, swap.HolderAddress, amtout, swap.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.UpdateLiquidity(tx, swapl.Tick)
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) UpdateLiquidity(tx *gorm.DB, tick string) error {

	err := tx.Exec(`UPDATE swap_liquidity
				SET amt0 = (
					SELECT b.amt_sum
					FROM drc20_collect_address b
					WHERE 
						swap_liquidity.tick0 = b.tick AND 
						swap_liquidity.reserves_address = b.holder_address
				)
				WHERE EXISTS (
					SELECT 1
					FROM drc20_collect_address b
					WHERE 
						swap_liquidity.tick0 = b.tick AND 
						swap_liquidity.reserves_address = b.holder_address AND 
						swap_liquidity.tick = ?
				)`, tick).Error
	if err != nil {
		return fmt.Errorf("UpdateLiquidity error: %s", err.Error())
	}

	err = tx.Exec(` UPDATE swap_liquidity
				SET amt1 = (
					SELECT b.amt_sum
					FROM drc20_collect_address b
					WHERE 
						swap_liquidity.tick1 = b.tick AND 
						swap_liquidity.reserves_address = b.holder_address
				)
				WHERE EXISTS (
					SELECT 1
					FROM drc20_collect_address b
					WHERE 
						swap_liquidity.tick1 = b.tick AND 
						swap_liquidity.reserves_address = b.holder_address AND 
						swap_liquidity.tick = ?
				)`, tick).Error
	if err != nil {
		return err
	}

	err = tx.Exec(`UPDATE swap_liquidity
				SET liquidity_total = (
					SELECT b.amt_sum
					FROM drc20_collect b
					WHERE swap_liquidity.tick = b.tick
				)
				WHERE tick = ?`, tick).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) FindSwapPriceAll() ([]*SwapPrice, int64, error) {

	liquidityAll := make([]*models.SwapLiquidity, 0)
	total := int64(0)
	err := c.DB.Where("liquidity_total != '0'").Find(&liquidityAll).Limit(-1).Offset(-1).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	liquidityMap := make(map[string]*big.Float)
	noWdoge := make([]*models.SwapLiquidity, 0)
	for _, v := range liquidityAll {
		if v.Tick0 != "WDOGE(WRAPPED-DOGE)" && v.Tick1 != "WDOGE(WRAPPED-DOGE)" {
			noWdoge = append(noWdoge, v)
			continue
		}

		if v.Tick0 == "WDOGE(WRAPPED-DOGE)" {
			liquidityMap[v.Tick1] = new(big.Float).Quo(new(big.Float).SetInt(v.Amt0.Int()), new(big.Float).SetInt(v.Amt1.Int()))
		} else {
			liquidityMap[v.Tick0] = new(big.Float).Quo(new(big.Float).SetInt(v.Amt1.Int()), new(big.Float).SetInt(v.Amt0.Int()))
		}
	}

	liquiditys := make([]*SwapPrice, 0)
	len := int64(0)
	for k, v := range liquidityMap {
		f, _ := v.Float64()
		liquiditys = append(liquiditys, &SwapPrice{
			Tick:      k,
			LastPrice: f,
		})
		len++
	}

	liquidityMapNoDoge := make(map[string]*big.Float)
	for _, v := range noWdoge {
		price0, ok0 := liquidityMap[v.Tick0]
		price1, ok1 := liquidityMap[v.Tick1]
		if ok0 && ok1 {
			continue
		}

		if ok0 {

			price := new(big.Float).Quo(new(big.Float).SetInt(v.Amt0.Int()), new(big.Float).SetInt(v.Amt1.Int()))
			price = new(big.Float).Mul(price, price0)
			liquidityMapNoDoge[v.Tick1] = price

		} else if ok1 {

			price := new(big.Float).Quo(new(big.Float).SetInt(v.Amt1.Int()), new(big.Float).SetInt(v.Amt0.Int()))
			price = new(big.Float).Mul(price, price1)
			liquidityMapNoDoge[v.Tick0] = price
		}
	}

	for k, v := range liquidityMapNoDoge {
		f, _ := v.Float64()
		liquiditys = append(liquiditys, &SwapPrice{
			Tick:      k,
			LastPrice: f,
		})
		len++
	}

	return liquiditys, len, nil
}
