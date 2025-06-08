package ctrl

import (
	"b3challenge/internal/adapter/http/request"
	"b3challenge/internal/adapter/http/response"
	"context"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

//go:generate mockgen -source=trades_ctrl.go -destination=trades_ctrl_mock.go -package=ctrl TradesUC
type TradesUC interface {
	ComputeTickerMetrics(ctx context.Context, ticker string, date *time.Time) (decimal.Decimal, int, error)
}
type TradesCtrl struct {
	uc TradesUC
}

func NewTradesCtrl(uc TradesUC) *TradesCtrl {
	return &TradesCtrl{
		uc: uc,
	}
}

func (h *TradesCtrl) ComputeTickerMetrics(c echo.Context) error {
	var req request.ComputeTickerMetricsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := req.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	maxRangeValue, maxDailyValue, err := h.uc.ComputeTickerMetrics(c.Request().Context(), req.Ticker, req.ParsedDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error: "+err.Error())
	}

	res := response.NewComputeTickerMetricsResponse(req.Ticker, maxRangeValue, maxDailyValue)

	return c.JSON(http.StatusOK, res)
}
