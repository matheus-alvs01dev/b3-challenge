package entity

import "time"

type Trade struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Ticker    string
	TradeTime time.Time
	TradeDate time.Time
	Price     float64
	Quantity  int
}

func NewTrade(tradeTime, tradeDate time.Time, price float64, quantity int) *Trade {
	return &Trade{
		TradeTime: tradeTime,
		TradeDate: tradeDate,
		Price:     price,
		Quantity:  quantity,
	}
}
