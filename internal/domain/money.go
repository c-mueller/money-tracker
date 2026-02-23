package domain

import (
	"strings"

	"github.com/shopspring/decimal"
)

type Money = decimal.Decimal

func NewMoney(value string) (Money, error) {
	value = strings.Replace(value, ",", ".", 1)
	return decimal.NewFromString(value)
}

func MoneyFromInt(cents int64) Money {
	return decimal.New(cents, -2)
}

func ZeroMoney() Money {
	return decimal.Zero
}
