package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcec"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/txscript"
	"github.com/dogecoinw/doged/wire"
	"log"
	"strconv"
	"strings"
	"testing"
)

const (
	baseAmount       = 100000
	wdogeFeeAddress  = "D86Dc4n49LZDiXvB41ds2XaDAP1BFjP1qy"
	wdogeCoolAddress = "DKMyk8cfSTGfnCVXfmo8gXta9F6gziu7Z5"
)

// DFUQLPRz7Fc9v37s3XZUwtMgcLBiXKVgPR transfer WDOGE(WRAPPED-DOGE) 100000000 to DTZSTXecLmSXpRGSfht4tAMyqra1wsL7xb
// WIF: QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ
// p2kh-add: DJu5mMUKprfnyBhot2fqCsW9sZCsfdfcrZ
func TestCreateAddress(t *testing.T) {

	//craete a new private key
	key, _ := btcec.NewPrivateKey()
	wif, _ := btcutil.NewWIF(key, &chaincfg.MainNetParams, true)
	fmt.Println("WIF: ", wif.String())

	//create a new address
	addr1, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed()), &chaincfg.MainNetParams)
	fmt.Println("p2kh-add: ", addr1.EncodeAddress())

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
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
		return
	}

	addr, _ := btcutil.NewAddressScriptHash(redeemScript, &chaincfg.MainNetParams)
	fmt.Println("addr: ", addr.String())
}

func TestTransfer(t *testing.T) {

	utxoHash, err := chainhash.NewHashFromStr("52491cf5bafff1b1098d997a93429f818239e764084007e4fbef8b290dde051e")
	if err != nil {
		log.Fatal(err)
		return
	}
	utxoIndex := uint32(0)
	to := "DTZSTXecLmSXpRGSfht4tAMyqra1wsL7xb"

	wifStr1 := "QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ"
	wif, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		log.Fatal(err)
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
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
	}

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, utxoIndex)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	toAddress := strings.Split(to, ",")
	for _, add := range toAddress {
		decodedAddr, err := btcutil.DecodeAddress(add, &chaincfg.MainNetParams)
		if err != nil {
			log.Fatal(err)
		}

		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			log.Fatal(err)
		}

		// fee address
		redeemTxOut := wire.NewTxOut(baseAmount, destinationAddrByte)
		redeemTx.AddTxOut(redeemTxOut)
	}

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

func TestMintDeploy(t *testing.T) {

	utxoHash, err := chainhash.NewHashFromStr("52491cf5bafff1b1098d997a93429f818239e764084007e4fbef8b290dde051e")
	if err != nil {
		log.Fatal(err)
		return
	}
	utxoIndex := uint32(0)
	to := "DTZSTXecLmSXpRGSfht4tAMyqra1wsL7xb"

	repeat := int64(30) // Multiples of mint, up to 30

	wifStr1 := "QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ"
	wif, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		log.Fatal(err)
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	//data["p"] = "drc-20"
	//data["op"] = "deploy"
	//data["tick"] = "WOW"
	//data["max"] = "100000000"
	//data["lim"] = "1000000"

	data := make(map[string]interface{})
	data["p"] = "drc-20"
	data["op"] = "mint"
	data["tick"] = "WOW"
	data["amt"] = "100000000"
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	builder := txscript.NewScriptBuilder()

	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(wif.PrivKey.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
	}

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, utxoIndex)

	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the output to tx,
	toDec, err := btcutil.DecodeAddress(to, &chaincfg.MainNetParams)
	if err != nil {
		return
	}

	toAddrByte, err := txscript.PayToAddrScript(toDec)
	if err != nil {
		log.Fatal(err)
	}

	baseFee := int64(50000000)
	if data["op"] == "deploy" {
		repeat = 1
		baseFee = 10000000000
	}

	// to address
	out0 := wire.NewTxOut(baseAmount*repeat, toAddrByte)
	redeemTx.AddTxOut(out0)

	decodedAddrFee, _ := btcutil.DecodeAddress("feeAddress", &chaincfg.MainNetParams)
	destinationAddrByteFee, err := txscript.PayToAddrScript(decodedAddrFee)
	if err != nil {
		log.Fatal(err)
	}

	// fee address
	redeemTxOut := wire.NewTxOut(baseFee*repeat, destinationAddrByteFee)
	redeemTx.AddTxOut(redeemTxOut)

	// signing the tx
	sig1, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif.PrivKey)
	if err != nil {

		log.Fatal(err)
	}

	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_10).AddOp(txscript.OP_FALSE).AddData(sig1)
	signature.AddData(redeemScript)
	signatureScript, err := signature.Script()
	if err != nil {
		log.Fatal(err)
	}

	redeemTx.TxIn[0].SignatureScript = signatureScript

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
	hexSignedTx := hex.EncodeToString(signedTx.Bytes())
	fmt.Println("signed tx: ", hexSignedTx)
}

func TestSwap(t *testing.T) {

	utxoHash, err := chainhash.NewHashFromStr("52491cf5bafff1b1098d997a93429f818239e764084007e4fbef8b290dde051e")
	if err != nil {
		log.Fatal(err)
		return
	}
	utxoIndex := uint32(0)
	swapAddress := "DTZSTXecLmSXpRGSfht4tAMyqra1wsL7xb"

	wifStr1 := "QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ"
	wif, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[string]interface{})
	data["p"] = "pair-v1"

	// create
	//data["op"] = swap.Op
	//data["tick0"] = swap.Tick0
	//data["tick1"] = swap.Tick1
	//data["amt0"] = swap.Amt0.String()
	//data["amt1"] = swap.Amt1.String()
	//data["amt0_min"] = swap.Amt0Min.String()
	//data["amt1_min"] = swap.Amt1Min.String()

	// add
	//data["op"] = swap.Op
	//data["tick0"] = swap.Tick0
	//data["tick1"] = swap.Tick1
	//data["amt0"] = swap.Amt0.String()
	//data["amt1"] = swap.Amt1.String()
	//data["amt0_min"] = swap.Amt0Min.String()
	//data["amt1_min"] = swap.Amt1Min.String()

	// remove
	//data["op"] = swap.Op
	//data["tick0"] = swap.Tick0
	//data["tick1"] = swap.Tick1
	//data["liquidity"] = swap.Liquidity.String()

	// swap
	data["op"] = "swap"
	data["tick0"] = "UNIX"
	data["tick1"] = "WDOGE(WRAPPED-DOGE)"
	data["amt0"] = "100000000"
	data["amt1"] = "644043"
	data["amt1_min"] = "64404"
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	builder := txscript.NewScriptBuilder()

	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(wif.PrivKey.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, utxoIndex)

	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the output to tx,
	decodedAddr, err := btcutil.DecodeAddress(swapAddress, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)
	}

	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		log.Fatal(err)
	}

	// fee address
	redeemTxOut := wire.NewTxOut(baseAmount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	// signing the tx
	sig1, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif.PrivKey)
	if err != nil {
		log.Fatal(err)
	}

	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_10).AddOp(txscript.OP_FALSE).AddData(sig1)
	signature.AddData(redeemScript)
	signatureScript, err := signature.Script()
	if err != nil {
		log.Fatal(err)
	}

	redeemTx.TxIn[0].SignatureScript = signatureScript

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	fmt.Println("signed tx: ", hexSignedTx)
}

func TestWdoge(t *testing.T) {
	utxoHash, err := chainhash.NewHashFromStr("52491cf5bafff1b1098d997a93429f818239e764084007e4fbef8b290dde051e")
	if err != nil {
		log.Fatal(err)
		return
	}
	utxoIndex := uint32(0)
	wdogeAddress := "DTZSTXecLmSXpRGSfht4tAMyqra1wsL7xb"

	wifStr1 := "QRJx7uvj55L3oVRADWJfFjJ31H9Beg75xZ2GcmR8rKFNHA4ZacKJ"
	wif, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		log.Fatal(err)
	}

	amt := int64(100000000)
	data := make(map[string]interface{})
	data["p"] = "wdoge"

	// deposit
	data["op"] = "deposit"
	data["tick"] = "WDOGE(WRAPPED-DOGE)"
	data["amt"] = strconv.FormatInt(amt, 10)
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	// withdraw
	//data["op"] = "withdraw"
	//data["tick"] = wdoge.Tick
	//data["amt"] = wdoge.Amt.String()

	builder := txscript.NewScriptBuilder()

	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(wif.PrivKey.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		log.Fatal(err)
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, utxoIndex)

	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the output to tx,
	decodedAddr, err := btcutil.DecodeAddress(wdogeAddress, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)
	}

	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		log.Fatal(err)
	}

	redeemTxOut := wire.NewTxOut(baseAmount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	decodedAddrCool, _ := btcutil.DecodeAddress(wdogeCoolAddress, &chaincfg.MainNetParams)
	destinationAddrByteCool, _ := txscript.PayToAddrScript(decodedAddrCool)

	// The quantity must be consistent with the inscription
	redeemTxOutCool := wire.NewTxOut(amt, destinationAddrByteCool)
	redeemTx.AddTxOut(redeemTxOutCool)

	// charge handling fee
	wFee, _ := btcutil.DecodeAddress(wdogeFeeAddress, &chaincfg.MainNetParams)
	wByteFee, _ := txscript.PayToAddrScript(wFee)

	fee := int64(0)
	if amt*3/1000 < 50000000 {
		fee = 50000000
	} else {
		fee = amt * 3 / 1000
	}

	redeemTxOutWFee := wire.NewTxOut(fee, wByteFee)
	redeemTx.AddTxOut(redeemTxOutWFee)

	// signing the tx
	sig1, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif.PrivKey)
	if err != nil {
		log.Fatal(err)
	}

	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_10).AddOp(txscript.OP_FALSE).AddData(sig1)
	signature.AddData(redeemScript)
	signatureScript, err := signature.Script()
	if err != nil {
		log.Fatal(err)
	}

	redeemTx.TxIn[0].SignatureScript = signatureScript

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	fmt.Println("signed tx: ", hexSignedTx)
}
