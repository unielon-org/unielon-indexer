package explorer

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dogecoinw/doged/btcjson"
	"github.com/dogecoinw/doged/txscript"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (e *Explorer) reDecode(tx *btcjson.TxRawResult) (*utils.BaseParams, []byte, error) {

	in := tx.Vin[0]

	if in.ScriptSig == nil {
		return nil, nil, errors.New("ScriptSig is nil")
	}

	scriptbytes, err := hex.DecodeString(in.ScriptSig.Hex)
	if err != nil {
		return nil, nil, fmt.Errorf("hex.DecodeString err: %s", err.Error())
	}

	pkScript, err := txscript.PushedData(scriptbytes)
	if err != nil {
		return nil, nil, fmt.Errorf("PushedData err: %s", err.Error())
	}

	if len(pkScript) < 3 {
		return nil, nil, errors.New("pkScript length < 3")
	}

	pushedData, err := txscript.PushedData(pkScript[2])
	if err != nil {
		return nil, nil, fmt.Errorf("PushedData err: %s", err.Error())
	}

	if len(pushedData) < 4 {
		return nil, nil, errors.New("len(pushedData) < 4")
	}

	param := &utils.BaseParams{}
	err = json.Unmarshal(pushedData[3], param)
	if err != nil {
		return nil, nil, fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	return param, pushedData[3], nil

}
