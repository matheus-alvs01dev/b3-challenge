package usecase

import (
	"b3challenge/internal/domain/entity"
	"context"
)

type TradesRepository interface{}

type TradesUC struct {
	repo TradesRepository
}

func NewTradesUC(repo TradesRepository) *TradesUC {
	return &TradesUC{
		repo: repo,
	}
}

func (tr *TradesUC) SearchTrades(ctx context.Context, filters interface{}) (*entity.Trade, error) {
	panic("implement me")
}
