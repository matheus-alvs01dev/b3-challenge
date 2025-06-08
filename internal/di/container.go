package di

import (
	"b3challenge/internal/adapter/db"
	"b3challenge/internal/api/ctrl"
	"b3challenge/internal/domain/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	database         *pgxpool.Pool
	tradesRepository *db.TradeRepository
	tradesUC         *usecase.TradesUC
}

func NewContainer(database *pgxpool.Pool) *Container {
	tradesRepository := db.NewTradeRepository(database)
	tradesUC := usecase.NewTradesUC(tradesRepository)

	return &Container{
		database:         database,
		tradesRepository: tradesRepository,
		tradesUC:         tradesUC,
	}
}

func (c *Container) NewTradesHandler() *ctrl.TradesCtrl {
	return ctrl.NewTradesCtrl(c.tradesUC)
}

func (c *Container) GetTradesUC() *usecase.TradesUC {
	return c.tradesUC
}
