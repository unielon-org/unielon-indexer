package verifys

import (
	"errors"
	"fmt"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"strings"
)

type Verifys struct {
	dbc *storage.DBClient
}

func NewVerifys(dbc *storage.DBClient) *Verifys {
	return &Verifys{
		dbc: dbc,
	}
}

func (v *Verifys) VerifyDrc20(card *utils.Cardinals) error {
	switch card.Op {
	case "deploy":
		return v.verifyDeploy(card)
	case "mint":
		return v.verifyMint(card)
	case "transfer":
		return v.verifyTransfer(card)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) verifyDeploy(card *utils.Cardinals) error {

	if len(card.Tick) < 2 || len(card.Tick) > 8 {
		return fmt.Errorf("the token symbol must be 2 or 8 letters")
	}

	if card.Max.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if card.Lim.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if card.Max.Cmp(big.NewInt(0).Sub(big.NewInt(0).Exp(big.NewInt(16), big.NewInt(64), nil), big.NewInt(1))) > 0 {
		return fmt.Errorf("the maximum value cannot be greater 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")
	}

	if card.Lim.Cmp(big.NewInt(0).Sub(big.NewInt(0).Exp(big.NewInt(16), big.NewInt(64), nil), big.NewInt(1))) > 0 {
		return fmt.Errorf("the limit value cannot be greater 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")
	}

	if card.Max.Cmp(card.Lim) < 0 {
		return fmt.Errorf("the maximum value is less than the limit value")
	}

	_, _, _, err := v.dbc.FindDrc20InfoSumByTick(card.Tick)
	if err == nil {
		return fmt.Errorf("has been deployed contracts")
	}

	return nil
}

func (v *Verifys) verifyMint(card *utils.Cardinals) error {

	if len(card.Tick) < 2 || len(card.Tick) > 8 {
		return fmt.Errorf("the token symbol must be 2 or 8 letters")
	}

	sum, max, lim, err := v.dbc.FindDrc20InfoSumByTick(card.Tick)

	if err != nil {
		return fmt.Errorf("the contract does not exist")
	}

	if card.Amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if card.Amt.Cmp(lim) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the limit")
	}

	if sum != nil {
		amount := big.NewInt(0).Mul(card.Amt, big.NewInt(card.Repeat))
		Amt := new(big.Int).Add(sum, amount)
		if Amt.Cmp(max) > 0 {
			return fmt.Errorf("the amount of tokens exceeds the maximum")
		}
	}

	return nil
}

func (v *Verifys) verifyTransfer(card *utils.Cardinals) error {

	if card.Amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	tranCount := len(strings.Split(card.ToAddress, ","))

	sum, err := v.dbc.FindDrc20AddressInfoByTick(card.Tick, card.ReceiveAddress)
	if err != nil || sum == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	CAmt := big.NewInt(0).Mul(card.Amt, big.NewInt(int64(tranCount)))
	if CAmt.Cmp(sum) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}
	return nil
}

func (v *Verifys) VerifySwap(swap *utils.SwapInfo) error {
	switch swap.Op {
	case "create":
		return v.verifySwapCreate(swap)
	case "add":
		return v.verifySwapAdd(swap)
	case "remove":
		return v.verifySwapRemove(swap)
	case "swap":
		return v.verifySwapNow(swap)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) verifySwapCreate(swap *utils.SwapInfo) error {

	if swap.Amt0.Cmp(big.NewInt(0)) < 1 || swap.Amt1.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if swap.Tick0 == swap.Tick1 {
		return fmt.Errorf("the token symbol must be different")
	}

	tick0, tick1, amt0, amt1, _, _ := utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, nil, nil)

	info, err := v.dbc.FindSwapLiquidity(tick0, tick1)
	if err != nil {
		return fmt.Errorf("The contract does not exist err %s", err.Error())
	}

	if info != nil {
		return fmt.Errorf("Has been deployed pool")
	}

	sum0, err := v.dbc.FindDrc20AddressInfoByTick(tick0, swap.HolderAddress)
	if err != nil || sum0 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	sum1, err := v.dbc.FindDrc20AddressInfoByTick(tick1, swap.HolderAddress)
	if err != nil || sum1 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if amt0.Cmp(sum0) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	if amt1.Cmp(sum1) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	return nil
}

func (v *Verifys) verifySwapAdd(swap *utils.SwapInfo) error {

	tick0, tick1, amt0, amt1, amt0Min, amt1Min := utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	if swap.Amt0.Cmp(big.NewInt(0)) < 1 || swap.Amt1.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if swap.Tick0 == swap.Tick1 {
		return fmt.Errorf("the token symbol must be different")
	}
	info, err := v.dbc.FindSwapLiquidity(tick0, tick1)
	if err != nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if info == nil {
		return fmt.Errorf("the contract does not exist")
	}

	sum0, err := v.dbc.FindDrc20AddressInfoByTick(tick0, swap.HolderAddress)
	if err != nil || sum0 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	sum1, err := v.dbc.FindDrc20AddressInfoByTick(tick1, swap.HolderAddress)
	if err != nil || sum1 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if info.LiquidityTotal.Cmp(big.NewInt(0)) == 0 {

		if amt0.Cmp(sum0) > 0 {
			return fmt.Errorf("the amount of tokens exceeds the balance")
		}

		if amt1.Cmp(sum1) > 0 {
			return fmt.Errorf("the amount of tokens exceeds the balance")
		}

	} else {
		amountBOptimal := big.NewInt(0).Mul(amt0, info.Amt1)

		if amountBOptimal.Cmp(big.NewInt(0)) < 1 {
			return fmt.Errorf("the amount of tokens exceeds the 0")
		}

		amountBOptimal = big.NewInt(0).Div(amountBOptimal, info.Amt0)

		if amountBOptimal.Cmp(amt1Min) < 0 {
			amountAOptimal := big.NewInt(0).Mul(amt1, info.Amt0)
			if amountAOptimal.Cmp(big.NewInt(0)) < 1 {
				return fmt.Errorf("the amount of tokens exceeds the 0")
			}
			amountAOptimal = big.NewInt(0).Div(amountAOptimal, info.Amt1)

			if amountAOptimal.Cmp(amt0Min) < 0 {
				return fmt.Errorf("the amount of tokens exceeds the min")
			} else {
				if amountAOptimal.Cmp(sum0) > 0 {
					return fmt.Errorf("the amount of tokens exceeds the balance")
				}

				if amt1.Cmp(sum1) > 0 {
					return fmt.Errorf("the amount of tokens exceeds the balance")
				}
			}
		} else {
			if amt0.Cmp(sum0) > 0 {
				return fmt.Errorf("the amount of tokens exceeds the balance")
			}

			if amountBOptimal.Cmp(sum1) > 0 {
				return fmt.Errorf("the amount of tokens exceeds the max")
			}
		}
	}

	return nil
}

func (v *Verifys) verifySwapNow(swap *utils.SwapInfo) error {

	if swap.Amt0.Cmp(big.NewInt(0)) < 1 || swap.Amt1.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if swap.Tick0 == swap.Tick1 {
		return fmt.Errorf("the token symbol must be different")
	}
	tick0, tick1, _, _, _, _ := utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	info, err := v.dbc.FindSwapLiquidity(tick0, tick1)
	if info == nil {
		return fmt.Errorf("the contract does not exist")
	}

	amtMap := make(map[string]*big.Int)
	amtMap[info.Tick0] = info.Amt0
	amtMap[info.Tick1] = info.Amt1

	amt0t := new(big.Int).Div(swap.Amt0, big.NewInt(100))
	amt1t := new(big.Int).Mul(amt0t, big.NewInt(2))

	amtin := new(big.Int).Add(amt0t, amt1t)
	amtin = new(big.Int).Sub(swap.Amt0, amtin)

	amtout := new(big.Int).Mul(amtin, amtMap[swap.Tick1])
	amtout = new(big.Int).Div(amtout, new(big.Int).Add(amtMap[swap.Tick0], amtin))

	if amtout.Cmp(swap.Amt1Min) < 0 {
		return fmt.Errorf("the minimum output less than the limit.")
	}

	sum0, err := v.dbc.FindDrc20AddressInfoByTick(swap.Tick0, swap.HolderAddress)
	if err != nil || sum0 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	sum1, err := v.dbc.FindDrc20AddressInfoByTick(swap.Tick1, info.ReservesAddress)
	if err != nil || sum1 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if swap.Amt0.Cmp(sum0) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	if amtout.Cmp(sum1) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	return nil
}

func (v *Verifys) verifySwapRemove(swap *utils.SwapInfo) error {

	if swap.Liquidity.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if swap.Tick0 == swap.Tick1 {
		return fmt.Errorf("the token symbol must be different")
	}
	tick0, tick1, _, _, _, _ := utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	info, err := v.dbc.FindSwapLiquidity(tick0, tick1)
	if info == nil {
		return fmt.Errorf("the contract does not exist")
	}

	if swap.Liquidity.Cmp(info.LiquidityTotal) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	tick := tick0 + "-SWAP-" + tick1

	sum0, err := v.dbc.FindDrc20AddressInfoByTick(tick, swap.HolderAddress)
	if err != nil || sum0 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if swap.Liquidity.Cmp(sum0) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	return nil
}

func (v *Verifys) VerifyWDoge(wdoge *utils.WDogeInfo) error {
	switch wdoge.Op {
	case "deposit":
		return v.verifyWDogeDeposit(wdoge)
	case "withdraw":
		return v.verifyWDogeWithdraw(wdoge)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) verifyWDogeDeposit(wdoge *utils.WDogeInfo) error {
	if wdoge.Amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}
	return nil
}

func (v *Verifys) verifyWDogeWithdraw(wdoge *utils.WDogeInfo) error {
	if wdoge.Amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if wdoge.Amt.Cmp(big.NewInt(100000000)) < 0 {
		return fmt.Errorf("the amount of tokens exceeds the 1")
	}

	sum0, err := v.dbc.FindDrc20AddressInfoByTick("WDOGE(WRAPPED-DOGE)", wdoge.HolderAddress)
	if err != nil || sum0 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if wdoge.Amt.Cmp(sum0) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}
	return nil
}

func (v *Verifys) VerifyNft(nft *utils.NFTInfo) error {
	switch nft.Op {
	case "deploy":
		return v.verifyNftDeploy(nft)
	case "mint":
		return v.verifyNftMint(nft)
	case "transfer":
		return v.verifyNftTransfer(nft)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) verifyNftDeploy(nft *utils.NFTInfo) error {

	if len(nft.Tick) < 2 || len(nft.Tick) > 32 {
		return fmt.Errorf("the token symbol must be 2 or 32 letters")
	}

	if nft.Total < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	_, err := v.dbc.FindNftCollectByTick(nft.Tick)
	if err == nil {
		return fmt.Errorf("has been deployed contracts")
	}

	sum, err := v.dbc.FindDrc20AddressInfoByTick("CARDI", nft.HolderAddress)
	if err != nil || sum == nil {
		return errors.New("Deploying AI/NFT requires holding 8400 CARDI for deployment. Please note that holding is only for identity verification and will not affect your assets.")
	}

	cardiTotal := big.NewInt(840000000000)
	if sum.Cmp(cardiTotal) < 0 {
		return fmt.Errorf("Deploying AI/NFT requires holding 8400 CARDI for deployment. Please note that holding is only for identity verification and will not affect your assets.")
	}

	return nil
}

func (v *Verifys) verifyNftMint(nft *utils.NFTInfo) error {

	if len(nft.Tick) < 2 || len(nft.Tick) > 32 {
		return fmt.Errorf("the token symbol must be 2 or 32 letters")
	}

	nftc, err := v.dbc.FindNftCollectByTick(nft.Tick)
	if err != nil {
		return fmt.Errorf("the contract does not exist")
	}

	if nftc.Total <= nftc.TickSum {
		return fmt.Errorf("the amount of tokens exceeds the maximum")
	}

	return nil
}

func (v *Verifys) verifyNftTransfer(nft *utils.NFTInfo) error {

	if nft.TickId < 0 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	tx, _ := v.dbc.SqlDB.Begin()
	_, err := v.dbc.FindNftCollectAddressByTickAndId(tx, nft.Tick, nft.HolderAddress, nft.TickId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}
	tx.Commit()
	return nil
}

func (v *Verifys) VerifyStake(stake *utils.StakeInfo) error {
	switch stake.Op {
	case "stake":
		return v.verifyStakeStake(stake)
	case "unstake":
		return v.verifyStakeUnstake(stake)
	case "getallreward":
		return v.verifyStakeGetallreward(stake)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) verifyStakeStake(si *utils.StakeInfo) error {

	if si.Amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	sum0, err := v.dbc.FindDrc20AddressInfoByTick(si.Tick, si.HolderAddress)
	if err != nil || sum0 == nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if si.Amt.Cmp(sum0) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	return nil
}

func (v *Verifys) verifyStakeUnstake(si *utils.StakeInfo) error {

	if si.Amt.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	tx, err := v.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	stake, err := v.dbc.FindStakeCollectAddressByTickAndHolder(tx, si.HolderAddress, si.Tick)
	if err != nil {
		return fmt.Errorf("FindWDogeInfoByTxHash err: %s", err.Error())
	}

	if stake == nil {
		return fmt.Errorf("stake does not exist")
	}

	if si.Amt.Cmp(stake.Amt) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the balance")
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (v *Verifys) verifyStakeGetallreward(si *utils.StakeInfo) error {

	tx, err := v.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	stake, err := v.dbc.FindStakeCollectAddressByTickAndHolder(tx, si.HolderAddress, si.Tick)
	if err != nil {
		return fmt.Errorf("FindWDogeInfoByTxHash err: %s", err.Error())
	}

	if stake == nil {
		return fmt.Errorf("stake does not exist")
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (v *Verifys) VerifyExchange(ex *utils.ExchangeInfo) error {

	switch ex.Op {
	case "create":
		return v.VerifyExchangeCreate(ex)
	case "trade":
		return v.VerifyExchangeTrade(ex)
	case "cancel":
		return v.VerifyExchangeCancel(ex)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) VerifyExchangeCreate(ex *utils.ExchangeInfo) error {
	_, _, _, err := v.dbc.FindDrc20InfoSumByTick(ex.Tick0)
	if err != nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	_, _, _, err = v.dbc.FindDrc20InfoSumByTick(ex.Tick1)
	if err != nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if ex.Amt0.Cmp(big.NewInt(0)) < 1 || ex.Amt1.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	return nil
}

func (v *Verifys) VerifyExchangeTrade(ex *utils.ExchangeInfo) error {
	if ex.Amt1.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	_, err := v.dbc.FindExchangeCollectByExId(ex.ExId)
	if err != nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}
	return nil
}

func (v *Verifys) VerifyExchangeCancel(ex *utils.ExchangeInfo) error {
	if ex.Amt0.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	exc, err := v.dbc.FindExchangeCollectByExId(ex.ExId)
	if err != nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	if exc.HolderAddress == ex.HolderAddress {
		return fmt.Errorf("the same address cannot be traded")
	}

	return nil
}

func (v *Verifys) VerifyBox(box *utils.BoxInfo) error {

	switch box.Op {
	case "deploy":
		return v.VerifyBoxDeploy(box)
	case "mint":
		return v.VerifyBoxMint(box)
	default:
		return fmt.Errorf("do not support the type of tokens")
	}
}

func (v *Verifys) VerifyBoxDeploy(box *utils.BoxInfo) error {

	if len(box.Tick0) < 2 || len(box.Tick0) > 8 {
		return fmt.Errorf("the token symbol must be 2 or 8 letters")
	}

	_, _, _, err := v.dbc.FindDrc20InfoSumByTick(box.Tick0)
	if err == nil {
		return fmt.Errorf("has been deployed contracts")
	}

	if box.Max.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if box.LiqAmt.Cmp(big.NewInt(0)) < 1 && box.LiqBlock < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	if box.LiqAmt.Cmp(big.NewInt(0)) > 0 && box.LiqBlock > 0 {
		return fmt.Errorf("two cannot exist at the same time")
	}

	return nil
}

func (v *Verifys) VerifyBoxMint(box *utils.BoxInfo) error {

	tx, err := v.dbc.SqlDB.Begin()
	if err != nil {
		return err
	}

	if box.Amt1.Cmp(big.NewInt(0)) < 1 {
		return fmt.Errorf("the amount of tokens exceeds the 0")
	}

	boxc, err := v.dbc.FindBoxCollectByTick(tx, box.Tick0)
	if err != nil {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
	}

	liqa := big.NewInt(0).Add(boxc.LiqAmtFinish, box.Amt1)
	if boxc.LiqAmt.Cmp(big.NewInt(0)) > 0 && liqa.Cmp(boxc.LiqAmt) > 0 {
		return fmt.Errorf("the amount of tokens exceeds the maximum")
	}

	return nil
}
