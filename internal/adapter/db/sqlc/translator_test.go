package sqlc

import (
	"b3challenge/internal/domain/entity"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewToCreateTradesParams_Single(t *testing.T) {
	price := decimal.RequireFromString("1.23")
	trade := entity.Trade{
		Hour:     "131001",
		Date:     time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC),
		Ticker:   "ABC123",
		Price:    price,
		Quantity: 42,
	}

	got := NewToCreateTradesParams([]entity.Trade{trade})
	want := []CreateTradesParams{{
		Hour:   trade.Hour,
		Date:   pgtype.Date{Time: trade.Date, Valid: true},
		Ticker: trade.Ticker,
		Price: pgtype.Numeric{
			Int:   big.NewInt(123),
			Exp:   -2,
			Valid: true,
		},
		Quantity: int32(trade.Quantity),
	}}

	assert.Equal(t, want, got)
}

func TestNewListTradeInfoByTickerAndDateParams(t *testing.T) {
	const ticker = "ABC123"
	date := time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		date *time.Time
		want ListTradeInfoByTickerAndDateParams
	}{
		{
			name: "nil date",
			date: nil,
			want: ListTradeInfoByTickerAndDateParams{
				Ticker:    ticker,
				TradeDate: pgtype.Date{Valid: false},
			},
		},
		{
			name: "with date",
			date: &date,
			want: ListTradeInfoByTickerAndDateParams{
				Ticker: ticker,
				TradeDate: pgtype.Date{
					Time:  date,
					Valid: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewListTradeInfoByTickerAndDateParams(ticker, tt.date)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListTradeInfoByTickerAndDateRow_ToTradeInfo(t *testing.T) {
	const ticker = "XYZ789"
	d := time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC)

	row := &ListTradeInfoByTickerAndDateRow{
		Price: pgtype.Numeric{
			Int:   big.NewInt(123),
			Exp:   -2,
			Valid: true,
		},
		Date: pgtype.Date{
			Time:  d,
			Valid: true,
		},
		Quantity: int32(10),
	}

	got := row.ToTradeInfo(ticker)
	wantPrice := decimal.NewFromBigInt(big.NewInt(123), -2)
	want := entity.TradeInfo{
		Ticker:   ticker,
		Price:    wantPrice,
		Date:     d,
		Quantity: 10,
	}

	assert.Equal(t, want, got)
}
