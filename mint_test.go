package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/txscript"
	"github.com/dogecoinw/doged/wire"
	"log"
	"testing"
)

const (
	baseAmount = 100000
)

// test
// WIF: QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ
// addr:  9vQLNwYnR1BEpRiCXjyyPnA79hWcoDQREK
func TestCreateAddress(t *testing.T) {

	// craete a new private key
	//key, _ := btcec.NewPrivateKey()
	//wif, _ := btcutil.NewWIF(key, &chaincfg.MainNetParams, true)

	wifStr1 := "QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ"
	wif, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		log.Fatal(err)
		return
	}

	data := make(map[string]interface{})
	data["p"] = "drc-20"
	data["op"] = "transfer"
	data["tick"] = "WDOGE(WRAPPED-DOGE)"
	data["amt"] = "100000000"
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	builder := txscript.NewScriptBuilder()
	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(wif.PrivKey.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
		return
	}

	addr, _ := btcutil.NewAddressScriptHash(redeemScript, &chaincfg.MainNetParams)
	fmt.Println("addr: ", addr.String())
}

func TestSend(t *testing.T) {

	utxoHash, err := chainhash.NewHashFromStr("52491cf5bafff1b1098d997a93429f818239e764084007e4fbef8b290dde051e")
	if err != nil {
		log.Fatal(err)
		return
	}
	utxoIndex := uint32(0)

	wifStr1 := "QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ"
	wif, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		log.Fatal(err)
		return
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	data := make(map[string]interface{})
	data["p"] = "drc-20"
	data["op"] = "transfer"
	data["tick"] = "WDOGE(WRAPPED-DOGE)"
	data["amt"] = "100000000"
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	builder := txscript.NewScriptBuilder()
	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(wif.PrivKey.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
		return
	}

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, utxoIndex)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the output to tx,
	decodedAddr, err := btcutil.DecodeAddress("DFUQLPRz7Fc9v37s3XZUwtMgcLBiXKVgPR", &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("generate decode addr ")
		log.Fatal(err)
		return
	}

	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		log.Fatal(err)
		return
	}

	// adding the destination address and the amount to the transaction
	redeemTxOut := wire.NewTxOut(baseAmount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	sig, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif.PrivKey)
	if err != nil {
		log.Fatal(err)
		return
	}

	// add the redeem script to the transaction
	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_10).AddOp(txscript.OP_FALSE).AddData(sig)
	signature.AddData(redeemScript)
	signatureScript, err := signature.Script()
	if err != nil {
		return
	}

	redeemTx.TxIn[0].SignatureScript = signatureScript
	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())
	fmt.Println("signed tx: ", hexSignedTx)

}
