package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/dogecoinw/doged/btcec/ecdsa"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/unielon-org/unielon-indexer/models"
	"math"
	"math/big"
	"time"
)

var (
	MAX_NUMBER, _ = big.NewInt(0).SetString("99999999999999999999999999999999999999999", 10)
)

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

func ConvetStringToNumber(number string) (*models.Number, error) {
	if number != "" {
		num, is_ok := new(models.Number).SetString(number, 10)
		if !is_ok {
			return num, fmt.Errorf("number error")
		}
		return num, nil
	}
	return new(models.Number), nil
}

func SortTokens(Tick0 string, Tick1 string, Amt0, Amt1, Amt0Min, Amt1Min *models.Number) (string, string, *models.Number, *models.Number, *models.Number, *models.Number) {
	if Tick0 > Tick1 {
		return Tick1, Tick0, Amt1, Amt0, Amt1Min, Amt0Min
	}
	return Tick0, Tick1, Amt0, Amt1, Amt0Min, Amt1Min
}

func TimeCom(dateInterval string) (time.Time, error) {
	startDate := time.Now()
	if dateInterval == "10m" {
		hourStart := startDate.Truncate(time.Hour)
		if startDate.Minute() < 10 {
			startDate = hourStart.Add(-50 * time.Minute)
		} else if startDate.Minute() < 20 {
			startDate = hourStart.Add(-40 * time.Minute)
		} else if startDate.Minute() < 30 {
			startDate = hourStart.Add(-30 * time.Minute)
		} else if startDate.Minute() < 40 {
			startDate = hourStart.Add(-20 * time.Minute)
		} else if startDate.Minute() < 50 {
			startDate = hourStart.Add(-10 * time.Minute)
		} else {
			startDate = hourStart.Add(0 * time.Minute)
		}
	}

	if dateInterval == "30m" {
		hourStart := startDate.Truncate(time.Hour)
		if startDate.Minute() < 30 {
			startDate = hourStart.Add(-30 * time.Minute)
		} else {
			startDate = hourStart.Add(30 * time.Minute)
		}
	}

	if dateInterval == "1h" {
		startDate = startDate.Truncate(time.Hour)
	}

	if dateInterval == "2h" {
		hourStart := startDate.Truncate(time.Hour)
		if startDate.Hour() < 2 {
			startDate = hourStart.Add(-22 * time.Hour)
		} else if startDate.Hour() < 4 {
			startDate = hourStart.Add(-20 * time.Hour)
		} else if startDate.Hour() < 6 {
			startDate = hourStart.Add(-18 * time.Hour)
		} else if startDate.Hour() < 8 {
			startDate = hourStart.Add(-16 * time.Hour)
		} else if startDate.Hour() < 10 {
			startDate = hourStart.Add(-14 * time.Hour)
		} else if startDate.Hour() < 12 {
			startDate = hourStart.Add(-12 * time.Hour)
		} else if startDate.Hour() < 14 {
			startDate = hourStart.Add(-10 * time.Hour)
		} else if startDate.Hour() < 16 {
			startDate = hourStart.Add(-8 * time.Hour)
		} else if startDate.Hour() < 18 {
			startDate = hourStart.Add(-6 * time.Hour)
		} else if startDate.Hour() < 20 {
			startDate = hourStart.Add(-4 * time.Hour)
		} else if startDate.Hour() < 22 {
			startDate = hourStart.Add(-2 * time.Hour)
		} else {
			startDate = hourStart.Add(0 * time.Hour)
		}
	}

	if dateInterval == "6h" {
		hourStart := startDate.Truncate(time.Hour)
		if startDate.Hour() < 6 {
			startDate = hourStart.Add(-18 * time.Hour)
		} else if startDate.Hour() < 12 {
			startDate = hourStart.Add(-12 * time.Hour)
		} else if startDate.Hour() < 18 {
			startDate = hourStart.Add(-6 * time.Hour)
		} else {
			startDate = hourStart.Add(0 * time.Hour)
		}
	}

	if dateInterval == "12h" {
		hourStart := startDate.Truncate(time.Hour)
		if startDate.Hour() < 12 {
			startDate = hourStart.Add(-12 * time.Hour)
		} else {
			startDate = hourStart.Add(0 * time.Hour)
		}
	}

	if dateInterval == "1d" {
		startDate = startDate.Truncate(time.Hour)
		startDate = startDate.Add(-24 * time.Hour)
	}

	if dateInterval == "1w" {
		startDate = startDate.Truncate(time.Hour)
		startDate = startDate.Add(-7 * 24 * time.Hour)
	}

	return startDate, nil
}

func Float64ToBigInt(input float64) *big.Int {

	rounded := math.Ceil(input)

	if rounded < math.MinInt64 || rounded > math.MaxInt64 {
		return big.NewInt(0)
	}

	result := int64(rounded)
	return big.NewInt(result)
}

func Base64ToPng(base64Str string) ([]byte, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		panic(err)
	}
	return imageBytes, nil
}

func PngToBase64(pngBytes []byte) string {
	base64String := base64.StdEncoding.EncodeToString(pngBytes)
	return base64String
}

func Decode26Base(s string) *big.Int {
	result := big.NewInt(0)
	twentySix := big.NewInt(26)

	for i := 0; i < len(s); i++ {

		charValue := big.NewInt(int64(s[i] - 'A' + 1))
		result.Mul(result, twentySix)
		result.Add(result, charValue)
	}
	return result.Sub(result, big.NewInt(1))
}

func WriteVarInt(w *bytes.Buffer, n int64) {
	if n < 0xfd {
		w.WriteByte(byte(n))
	} else if n <= 0xffff {
		w.WriteByte(0xfd)
		w.WriteByte(byte(n))
		w.WriteByte(byte(n >> 8))
	} else if n <= 0xffffffff {
		w.WriteByte(0xfe)
		w.Write([]byte{
			byte(n),
			byte(n >> 8),
			byte(n >> 16),
			byte(n >> 24),
		})
	} else {
		w.WriteByte(0xff)
		w.Write([]byte{
			byte(n),
			byte(n >> 8),
			byte(n >> 16),
			byte(n >> 24),
			byte(n >> 32),
			byte(n >> 40),
			byte(n >> 48),
			byte(n >> 56),
		})
	}
}

func GetAddressFromSig(msg string, sig string) (string, error) {

	message := msg
	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.Write([]byte("\x18Bitcoin Signed Message:\n"))
	WriteVarInt(&buf, int64(len(message)))
	buf.Write([]byte(message))

	// Compute the hash of the preprocessed message
	hash := sha256.Sum256(buf.Bytes())
	hash = sha256.Sum256(hash[:])

	publicKey, _, err := ecdsa.RecoverCompact(sigBytes, hash[:])
	if err != nil {
		return "", err
	}

	inadress, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(publicKey.SerializeCompressed()), &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	return inadress.String(), nil
}
