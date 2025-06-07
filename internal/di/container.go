package di

import (
	"b3challenge/internal/adapter/db"
	"b3challenge/internal/api/ctrl"
	"b3challenge/internal/domain/usecase"
	"database/sql"
)

type Container struct {
	database         *sql.DB
	tradesRepository *db.TradeRepository
	tradesUC         *usecase.TradesUC
}

func NewContainer(database *sql.DB) *Container {
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
