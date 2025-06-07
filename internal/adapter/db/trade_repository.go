package db

import (
	"b3challenge/internal/domain/entity"
	"database/sql"
)

type TradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{
		db: db,
	}
}

func (r *TradeRepository) GetTrades() ([]entity.Trade, error) {
	panic("implement me")
}
