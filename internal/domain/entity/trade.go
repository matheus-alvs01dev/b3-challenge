package entity

import (
	"github.com/shopspring/decimal"
	"time"
)

type Trade struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Ticker    string
	Hour      string
	Date      time.Time
	Price     decimal.Decimal
	Quantity  int
}

func NewTrade(ticker, time string, date time.Time, price decimal.Decimal, quantity int) *Trade {
	return &Trade{
		Ticker:   ticker,
		Hour:     time,
		Date:     date,
		Price:    price,
		Quantity: quantity,
	}
}

type TradeInfo struct {
	Ticker   string
	Price    decimal.Decimal
	Date     time.Time
	Quantity int
}
