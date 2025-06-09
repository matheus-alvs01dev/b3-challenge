package db

import (
	"b3challenge/internal/adapter/db/sqlc"
	"b3challenge/internal/domain/entity"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type TradeRepository struct {
	db      *pgxpool.Pool
	querier sqlc.Querier
}

func NewTradeRepository(db *pgxpool.Pool) *TradeRepository {
	return &TradeRepository{
		db:      db,
		querier: sqlc.New(db),
	}
}

func (r *TradeRepository) CreateTrades(ctx context.Context, trades []entity.Trade) (int64, error) {
	params := sqlc.NewToCreateTradesParams(trades)

	affected, err := r.querier.CreateTrades(ctx, params)
	if err != nil {
		return 0, errors.Wrap(err, "create")
	}

	return affected, nil
}

func (r *TradeRepository) ListTradeInfoByTickerAndDate(
	ctx context.Context,
	ticker string,
	date *time.Time,
) ([]entity.TradeInfo, error) {
	params := sqlc.NewListTradeInfoByTickerAndDateParams(ticker, date)

	trades, err := r.querier.ListTradeInfoByTickerAndDate(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "list")
	}

	var result []entity.TradeInfo
	for _, trade := range trades {
		result = append(result, trade.ToTradeInfo(ticker))
	}

	return result, nil
}
