package usecase

import (
	"b3challenge/internal/domain/entity"
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"time"
)

//go:generate mockgen -source=trades_uc.go -destination=trades_uc_mock.go -package=usecase TradesRepository
type TradesRepository interface {
	CreateTrades(ctx context.Context, trades []entity.Trade) (int64, error)
	ListTradeInfoByTickerAndDate(ctx context.Context, ticker string, date *time.Time) ([]entity.TradeInfo, error)
}

type TradesUC struct {
	repo TradesRepository
}

func NewTradesUC(repo TradesRepository) *TradesUC {
	return &TradesUC{
		repo: repo,
	}
}

func (tr *TradesUC) CreateTrades(ctx context.Context, trades []entity.Trade) (int, error) {
	affected, err := tr.repo.CreateTrades(ctx, trades)
	if err != nil {
		return 0, errors.Wrap(err, "repo create")
	}

	return int(affected), nil
}

func (tr *TradesUC) ComputeTickerMetrics(
	ctx context.Context,
	ticker string,
	date *time.Time,
) (decimal.Decimal, int, error) {
	trades, err := tr.repo.ListTradeInfoByTickerAndDate(ctx, ticker, date)
	if err != nil {
		return decimal.Decimal{}, 0, errors.Wrap(err, "repo list")
	}

	maxRangeValue := calcMaxRangeValue(trades)
	maxDailyValue := calcMaxDailyValue(trades)

	return maxRangeValue, maxDailyValue, nil
}

func calcMaxRangeValue(trades []entity.TradeInfo) decimal.Decimal {
	var max decimal.Decimal
	for _, trade := range trades {
		if trade.Price.GreaterThan(max) {
			max = trade.Price
		}
	}

	return max
}

func calcMaxDailyValue(trades []entity.TradeInfo) int {
	dailyTotal := make(map[string]int)
	for _, trade := range trades {
		date := trade.Date.Format("2006-01-02")
		dailyTotal[date] += trade.Quantity
	}

	var max int
	for _, total := range dailyTotal {
		if total > max {
			max = total
		}
	}

	return max
}
