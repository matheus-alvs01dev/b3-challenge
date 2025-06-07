package ctrl

import (
	"b3challenge/internal/domain/entity"
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
)

type TradesUC interface {
	SearchTrades(ctx context.Context, filters interface{}) (*entity.Trade, error) // TODO: DEFINE DTOs
}
type TradesCtrl struct {
	uc TradesUC
}

func NewTradesCtrl(uc TradesUC) *TradesCtrl {
	return &TradesCtrl{
		uc: uc,
	}
}

func (h *TradesCtrl) SearchTrades(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not implemented")
}
