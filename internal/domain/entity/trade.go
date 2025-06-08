package entity

import "time"

type Trade struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Ticker    string
	Hour      string
	Date      time.Time
	Price     float64
	Quantity  int
}

func NewTrade(ticker, time string, date time.Time, price float64, quantity int) *Trade {
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
	Price    float64
	Date     time.Time
	Quantity int
}
