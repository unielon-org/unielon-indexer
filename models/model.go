package models

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"time"
)

type LocalTime int64

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	// Convert LocalTime (int64) to Unix timestamp integer
	timestamp := int64(*t)

	// Convert integer to byte slice
	return []byte(fmt.Sprintf("%d", timestamp)), nil
}

func (t LocalTime) Value() (driver.Value, error) {
	tlt := time.Unix(int64(t), 0)
	if t == 0 {
		return time.Now(), nil
	}
	return tlt, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value.Unix())
		return nil
	}

	if value, ok := v.(int64); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

type Number big.Int

func NewNumber(num int64) *Number {
	return (*Number)(big.NewInt(num))
}

func (n *Number) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", (*big.Int)(n).String())), nil
}

func (n *Number) Value() (driver.Value, error) {
	if n == nil {
		return (big.NewInt(0)).String(), nil
	}
	return (*big.Int)(n).String(), nil
}

func (n *Number) Scan(v interface{}) error {
	if value, ok := v.([]byte); ok {
		num, ok := new(big.Int).SetString(string(value), 10)
		if !ok {
			return fmt.Errorf("number error")
		}
		*n = Number(*num)
		return nil
	}

	if value, ok := v.(string); ok {
		num, ok := new(big.Int).SetString(value, 10)
		if !ok {
			return fmt.Errorf("number error")
		}
		*n = Number(*num)
		return nil
	}

	return fmt.Errorf("can not convert %v to number", v)
}

func (n *Number) New(y *big.Int) *Number {
	return (*Number)(y)
}

func (n *Number) String() string {
	return (*big.Int)(n).String()
}

func (n *Number) SetString(s string, base int) (*Number, bool) {
	bigInt, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, false
	}
	*n = Number(*bigInt)
	return n, true
}

func (n *Number) Cmp(y *Number) int {
	return (*big.Int)(n).Cmp((*big.Int)(y))
}

func (n *Number) Int() *big.Int {
	return (*big.Int)(n)
}

func (n *Number) Int64() int64 {
	return (*big.Int)(n).Int64()
}
