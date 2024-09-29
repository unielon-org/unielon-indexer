package utils

import (
	"encoding/json"
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"strings"
)

func FormatDrc20ToJson(card *models.Drc20Info) []byte {
	data := make(map[string]interface{})
	card.Tick = strings.ToUpper(card.Tick)
	data["p"] = card.P
	data["op"] = card.Op

	switch card.Op {
	case "deploy":
		data["tick"] = card.Tick
		data["max"] = card.Max.String()
		if card.Lim.String() != "0" {
			data["lim"] = card.Lim.String()
		}
		if card.Dec != 0 {
			data["dec"] = card.Dec
		}
		if card.Burn != "" {
			data["burn"] = card.Burn
		}
		if card.Func != "" {
			data["func"] = card.Func
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "mint":
		data["tick"] = card.Tick
		data["amt"] = card.Amt.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "transfer":
		data["tick"] = card.Tick
		data["amt"] = card.Amt.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatSwapToJson(swap *models.SwapInfo) []byte {
	swap.Tick0 = strings.ToUpper(swap.Tick0)
	swap.Tick1 = strings.ToUpper(swap.Tick1)
	data := make(map[string]interface{})
	data["p"] = "pair-v1"
	switch swap.Op {
	case "create":
		data["op"] = swap.Op
		data["tick0"] = swap.Tick0
		data["tick1"] = swap.Tick1
		data["amt0"] = swap.Amt0.String()
		data["amt1"] = swap.Amt1.String()
		data["amt0_min"] = swap.Amt0Min.String()
		data["amt1_min"] = swap.Amt1Min.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "add":
		data["op"] = swap.Op
		data["tick0"] = swap.Tick0
		data["tick1"] = swap.Tick1
		data["amt0"] = swap.Amt0.String()
		data["amt1"] = swap.Amt1.String()
		data["amt0_min"] = swap.Amt0Min.String()
		data["amt1_min"] = swap.Amt1Min.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "remove":
		data["op"] = swap.Op
		data["tick0"] = swap.Tick0
		data["tick1"] = swap.Tick1
		data["liquidity"] = swap.Liquidity.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "swap":
		data["op"] = swap.Op
		data["tick0"] = swap.Tick0
		data["tick1"] = swap.Tick1
		data["amt0"] = swap.Amt0.String()
		data["amt1"] = swap.Amt1.String()
		data["amt1_min"] = swap.Amt1Min.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatWDogeToJson(wdoge *models.WDogeInfo) []byte {
	wdoge.Tick = "WDOGE(WRAPPED-DOGE)"
	data := make(map[string]interface{})
	data["p"] = "wdoge"
	switch wdoge.Op {
	case "deposit":
		data["op"] = wdoge.Op
		data["tick"] = wdoge.Tick
		data["amt"] = wdoge.Amt.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "withdraw":
		data["op"] = wdoge.Op
		data["tick"] = wdoge.Tick
		data["amt"] = wdoge.Amt.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatNftToJson(nft *models.NftInfo) []byte {
	data := make(map[string]interface{})
	data["p"] = "nft/ai"
	data["op"] = nft.Op
	switch nft.Op {
	case "deploy":
		data["tick"] = nft.Tick
		data["total"] = nft.Total
		data["model"] = nft.Model
		data["prompt"] = nft.Prompt
		data["seed"] = nft.Seed
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "mint":
		data["tick"] = nft.Tick
		data["prompt"] = nft.Prompt
		data["seed"] = nft.Seed
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "transfer":
		data["tick"] = nft.Tick
		data["tick_id"] = nft.TickId
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatStakeToJson(stake *models.StakeInfo) []byte {
	data := make(map[string]interface{})
	data["p"] = "stake-v1"
	data["op"] = stake.Op
	switch stake.Op {
	case "stake":
		data["tick"] = stake.Tick
		data["amt"] = stake.Amt.String()
	case "unstake":
		data["tick"] = stake.Tick
		data["amt"] = stake.Amt.String()
	case "getallreward":
		data["tick"] = stake.Tick
	default:
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON encoding failed:", err)
	}
	return jsonData
}

func FormatExChangeToJson(ex *models.ExchangeInfo) []byte {

	ex.Tick0 = strings.ToUpper(ex.Tick0)
	ex.Tick1 = strings.ToUpper(ex.Tick1)

	data := make(map[string]interface{})
	data["p"] = "order-v1"
	data["op"] = ex.Op

	switch ex.Op {
	case "create":
		data["tick0"] = ex.Tick0
		data["tick1"] = ex.Tick1
		data["amt0"] = ex.Amt0.String()
		data["amt1"] = ex.Amt1.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "trade":
		data["exid"] = ex.ExId
		data["amt1"] = ex.Amt1.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "cancel":
		data["exid"] = ex.ExId
		data["amt0"] = ex.Amt0.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatBoxToJson(ex *models.BoxInfo) []byte {

	ex.Tick0 = strings.ToUpper(ex.Tick0)
	ex.Tick1 = strings.ToUpper(ex.Tick1)

	data := make(map[string]interface{})
	data["p"] = "box-v1"
	data["op"] = ex.Op
	switch ex.Op {
	case "deploy":
		data["tick0"] = ex.Tick0
		data["tick1"] = ex.Tick1
		data["max"] = ex.Max.String()
		data["amt0"] = ex.Amt0.String()
		data["liqamt"] = ex.LiqAmt.String()
		data["liqblock"] = ex.LiqBlock
		data["amt1"] = ex.Amt1.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "mint":
		data["tick0"] = ex.Tick0
		data["amt1"] = ex.Amt1.String()
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatFileToJson(file *models.FileInfo) []byte {
	data := make(map[string]interface{})
	data["p"] = "file"
	data["op"] = file.Op
	switch file.Op {
	case "deploy":
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "transfer":
		data["file_id"] = file.FileId
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}

func FormatFileExchangeToJson(fe *models.FileExchangeInfo) []byte {
	data := make(map[string]interface{})
	data["p"] = "order-v2"
	data["op"] = fe.Op
	switch fe.Op {
	case "create":
		data["file_id"] = fe.FileId
		data["tick"] = fe.FileId
		data["amt"] = fe.FileId
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "trade":
		data["file_id"] = fe.FileId
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	case "cancel":
		data["file_id"] = fe.FileId
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON encoding failed:", err)
		}
		return jsonData
	default:
		return nil
	}
}
