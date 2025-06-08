package sqlc

import (
	"b3challenge/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"time"
)

func NewToCreateTradesParams(trades []entity.Trade) []CreateTradesParams {
	var params []CreateTradesParams
	for _, trade := range trades {

		params = append(params, CreateTradesParams{
			Hour: trade.Hour,
			Date: pgtype.Date{
				Time:  trade.Date,
				Valid: true,
			},
			Ticker: trade.Ticker,
			Price: pgtype.Numeric{
				Int:   trade.Price.Coefficient(),
				Exp:   trade.Price.Exponent(),
				Valid: true,
			},
			Quantity: int32(trade.Quantity),
		})
	}

	return params
}

func NewListTradeInfoByTickerAndDateParams(ticker string, date *time.Time) ListTradeInfoByTickerAndDateParams {
	params := ListTradeInfoByTickerAndDateParams{
		Ticker:    ticker,
		TradeDate: pgtype.Date{Valid: false},
	}

	if date != nil {
		params.TradeDate = pgtype.Date{
			Time:  *date,
			Valid: true,
		}
	}

	return params
}

func (tr *ListTradeInfoByTickerAndDateRow) ToTradeInfo(ticker string) entity.TradeInfo {
	return entity.TradeInfo{
		Ticker:   ticker,
		Price:    decimal.NewFromBigInt(tr.Price.Int, tr.Price.Exp),
		Date:     tr.Date.Time,
		Quantity: int(tr.Quantity),
	}
}
