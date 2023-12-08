package explorer

import (
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/go-dogecoin/log"
	"github.com/google/uuid"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
)

const (
	MINI_LIQUIDITY = 20
)

func (e *Explorer) swapDecode(tx *btcjson.TxRawResult, pushedData []byte, number int64) (*utils.SwapInfo, error) {

	param := &utils.SwapParams{}
	err := json.Unmarshal(pushedData, param)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	swap, err := utils.ConvetSwap(param)
	if err != nil {
		return nil, fmt.Errorf("ConvetSwap err: %s", err.Error())
	}

	swap.OrderId = uuid.New().String()
	swap.SwapTxHash = tx.Hash
	swap.SwapBlockHash = tx.BlockHash
	swap.SwapBlockNumber = number
	swap.HolderAddress = tx.Vout[0].ScriptPubKey.Addresses[0]

	txhash0, _ := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	txRawResult0, err := e.node.GetRawTransactionVerboseBool(txhash0)
	if err != nil {
		return nil, chainNetworkErr
	}

	swap.FeeAddress = txRawResult0.Vout[tx.Vin[0].Vout].ScriptPubKey.Addresses[0]

	txhash1, _ := chainhash.NewHashFromStr(txRawResult0.Vin[0].Txid)
	txRawResult1, err := e.node.GetRawTransactionVerboseBool(txhash1)
	if err != nil {
		return nil, chainNetworkErr
	}

	if swap.HolderAddress != txRawResult1.Vout[txRawResult0.Vin[0].Vout].ScriptPubKey.Addresses[0] {
		return nil, fmt.Errorf("The address is not the same as the previous transaction")
	}

	swap1, err := e.dbc.FindSwapInfoBySwapTxHash(swap.SwapTxHash)
	if err != nil {
		return nil, fmt.Errorf("FindSwapInfoBySwapTxHash err: %s", err.Error())
	}

	if swap1 != nil {
		if swap1.SwapBlockNumber != 0 {
			return nil, fmt.Errorf("swap already exist or err %s", swap1.SwapTxHash)
		}
		swap.OrderId = swap1.OrderId
		return swap, nil
	} else {
		if err := e.dbc.InstallSwapInfo(swap); err != nil {
			return nil, fmt.Errorf("InstallSwapInfo err: %s", err.Error())
		}
	}

	return swap, nil
}

func (e Explorer) swapCreateOrAdd(swap *utils.SwapInfo) error {

	log.Info("swapCreateOrAdd", "op", swap.Op, "tick0", swap.Tick0, "tick1", swap.Tick1, "amt0", swap.Amt0, "amt1", swap.Amt1, "amt0Min", swap.Amt0Min, "amt1Min", swap.Amt1Min, "FeeAddress", swap.FeeAddress, "HolderAddress", swap.HolderAddress)
	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, swap.Amt0Min, swap.Amt1Min)

	info, err := e.dbc.FindSwapLiquidity(swap.Tick0, swap.Tick1)
	if err != nil {
		return fmt.Errorf("swapCreate FindSwapLiquidity error: %v", err)
	}

	if info == nil {
		reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(swap.Tick0+swap.Tick1), &chaincfg.MainNetParams)
		swap.Tick = swap.Tick0 + "-SWAP-" + swap.Tick1
		liquidityBase := new(big.Int).Sqrt(new(big.Int).Mul(swap.Amt0, swap.Amt1))
		if liquidityBase.Cmp(big.NewInt(MINI_LIQUIDITY)) > 0 {
			liquidityBase = new(big.Int).Sub(liquidityBase, big.NewInt(MINI_LIQUIDITY))
		}

		swap.Amt0Out = swap.Amt0
		swap.Amt1Out = swap.Amt1

		err = e.dbc.SwapCreate(swap, reservesAddress.String(), liquidityBase)

		if err != nil {
			return fmt.Errorf("swapCreate SwapCreate error: %v", err)
		}

	} else if info.LiquidityTotal.Cmp(big.NewInt(0)) == 0 {

		reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(swap.Tick0+swap.Tick1), &chaincfg.MainNetParams)
		swap.Tick = swap.Tick0 + "-SWAP-" + swap.Tick1
		liquidityBase := new(big.Int).Sqrt(new(big.Int).Mul(swap.Amt0, swap.Amt1))
		if liquidityBase.Cmp(big.NewInt(MINI_LIQUIDITY)) > 0 {
			liquidityBase = new(big.Int).Sub(liquidityBase, big.NewInt(MINI_LIQUIDITY))
		}

		swap.Amt0Out = swap.Amt0
		swap.Amt1Out = swap.Amt1

		err = e.dbc.SwapAdd(swap, reservesAddress.String(), swap.Amt0, swap.Amt1, liquidityBase)

		if err != nil {
			return fmt.Errorf("swapCreate SwapCreate error: %v", err)
		}

	} else {
		swap.Tick = info.Tick
		reservesAddress, _ := btcutil.NewAddressScriptHash([]byte(swap.Tick0+swap.Tick1), &chaincfg.MainNetParams)

		amt0Out := big.NewInt(0)
		amt1Out := big.NewInt(0)

		amountBOptimal := big.NewInt(0).Mul(swap.Amt0, info.Amt1)
		amountBOptimal = big.NewInt(0).Div(amountBOptimal, info.Amt0)
		if amountBOptimal.Cmp(swap.Amt1Min) >= 0 {
			amt0Out = swap.Amt0
			amt1Out = amountBOptimal
		} else {
			amountAOptimal := big.NewInt(0).Mul(swap.Amt1, info.Amt0)
			amountAOptimal = big.NewInt(0).Div(amountAOptimal, info.Amt1)
			if amountAOptimal.Cmp(swap.Amt0Min) >= 0 {
				amt0Out = amountAOptimal
				amt1Out = swap.Amt1
			} else {
				log.Error("The amount of tokens exceeds the balance")
				return nil
			}
		}

		liquidity0 := new(big.Int).Mul(amt0Out, info.LiquidityTotal)
		liquidity0 = new(big.Int).Div(liquidity0, info.Amt0)

		liquidity1 := new(big.Int).Mul(amt1Out, info.LiquidityTotal)
		liquidity1 = new(big.Int).Div(liquidity1, info.Amt1)

		liquidity := liquidity0
		if liquidity0.Cmp(liquidity1) > 0 {
			liquidity = liquidity1
		}

		err = e.dbc.SwapAdd(swap, reservesAddress.String(), amt0Out, amt1Out, liquidity)

		if err != nil {
			return fmt.Errorf("swapCreate SwapAdd error: %v", err)
		}
	}

	return nil
}

func (e Explorer) swapRemove(swap *utils.SwapInfo) error {

	swap.Tick0, swap.Tick1, _, _, _, _ = utils.SortTokens(swap.Tick0, swap.Tick1, nil, nil, nil, nil)

	info, err := e.dbc.FindSwapLiquidity(swap.Tick0, swap.Tick1)
	if err != nil {
		return fmt.Errorf("swapRemove FindSwapLiquidity error: %v", err)
	}
	swap.Tick = info.Tick

	amt0Out := new(big.Int).Mul(swap.Liquidity, info.Amt0)
	amt0Out = new(big.Int).Div(amt0Out, info.LiquidityTotal)

	amt1Out := new(big.Int).Mul(swap.Liquidity, info.Amt1)
	amt1Out = new(big.Int).Div(amt1Out, info.LiquidityTotal)

	if info.Amt0.Cmp(amt0Out) < 0 || info.Amt1.Cmp(amt1Out) < 0 {
		return fmt.Errorf("swapRemove FindSwapLiquidity error: %v", err)
	}

	err = e.dbc.SwapRemove(swap, info.ReservesAddress, amt0Out, amt1Out)
	if err != nil {
		return fmt.Errorf("swapRemove SwapRemove error: %v", err)
	}

	return nil
}

// swapNow
func (e Explorer) swapNow(swap *utils.SwapInfo) error {

	tick0, tick1, _, _, _, _ := utils.SortTokens(swap.Tick0, swap.Tick1, nil, nil, nil, nil)

	info, err := e.dbc.FindSwapLiquidity(tick0, tick1)
	if err != nil {
		return fmt.Errorf("swapNow FindSwapLiquidity error: %v", err)
	}

	swap.Tick = info.Tick

	amtMap := make(map[string]*big.Int)
	amtMap[info.Tick0] = info.Amt0
	amtMap[info.Tick1] = info.Amt1

	amtfee0 := new(big.Int).Div(swap.Amt0, big.NewInt(100))
	amt1 := new(big.Int).Mul(amtfee0, big.NewInt(2))

	amtin := new(big.Int).Add(amtfee0, amt1)
	amtin = new(big.Int).Sub(swap.Amt0, amtin)

	amtout := new(big.Int).Mul(amtin, amtMap[swap.Tick1])
	amtout = new(big.Int).Div(amtout, new(big.Int).Add(amtMap[swap.Tick0], amtin))

	amtfee1 := new(big.Int).Mul(amtfee0, amtMap[swap.Tick1])
	amtfee1 = new(big.Int).Div(amtfee1, new(big.Int).Add(amtMap[swap.Tick0], amtfee0))

	err = e.dbc.SwapNow(swap, info.ReservesAddress, swap.Amt0, amtout, amtfee0, amtfee1)
	if err != nil {
		return fmt.Errorf("swapNow SwapNow error: %v", err)
	}

	return nil
}
