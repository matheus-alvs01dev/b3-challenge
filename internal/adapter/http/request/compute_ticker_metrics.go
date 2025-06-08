package request

import (
	"github.com/pkg/errors"
	"time"
)

var (
	ErrTickerIsRequired = errors.New("invalid ticker")
	ErrInvalidTradeDate = errors.New("invalid trade date, must be in format YYYY-MM-DD")
)

type ComputeTickerMetricsRequest struct {
	Ticker     string     `query:"ticker"`
	TradeDate  *string    `query:"trade_date"`
	ParsedDate *time.Time `query:"-"`
}

func (r *ComputeTickerMetricsRequest) Validate() error {
	if r.Ticker == "" {
		return ErrTickerIsRequired
	}

	if r.ParsedDate == nil {
		parsed, err := time.Parse("2006-01-02", *r.TradeDate)
		if err != nil {
			return errors.Wrap(ErrInvalidTradeDate, err.Error())
		}

		r.ParsedDate = &parsed
	}

	return nil
}
