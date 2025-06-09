package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Trade struct {
	ID        int32     `exhaustruct:"optional"`
	CreatedAt time.Time `exhaustruct:"optional"`
	UpdatedAt time.Time `exhaustruct:"optional"`
	Ticker    string
	Hour      string
	Date      time.Time
	Price     decimal.Decimal
	Quantity  int32
}

func NewTrade(ticker, time string, date time.Time, price decimal.Decimal, quantity int32) *Trade {
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
