package ctrl

import (
	"b3challenge/internal/adapter/http/request"
	"b3challenge/internal/adapter/http/response"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewTradesCtrl(t *testing.T) {
	ucMock := NewMockTradesUC(gomock.NewController(t))
	expected := NewTradesCtrl(ucMock)
	ctrl := NewTradesCtrl(ucMock)
	assert.Equal(t, ctrl, expected)
}

func TestTradesCtrl_ComputeTickerMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name        string
		reqBody     request.ComputeTickerMetricsRequest
		uc          TradesUC
		wantErr     assert.ErrorAssertionFunc
		expectedRes *response.ComputeTickerMetricsResponse
	}{
		{
			name: "successful request",
			reqBody: request.ComputeTickerMetricsRequest{
				Ticker:    "AAPL",
				TradeDate: pointer.To("2025-06-08"),
			},
			uc: func() TradesUC {
				uc := NewMockTradesUC(ctrl)
				time, err := time.Parse(time.DateOnly, "2025-06-08")
				assert.NoError(t, err)
				uc.EXPECT().ComputeTickerMetrics(gomock.Any(), "AAPL", &time).Return(
					decimal.NewFromFloat(150.00), 100, nil,
				).Times(1)
				return uc
			}(),
			wantErr: assert.NoError,
			expectedRes: &response.ComputeTickerMetricsResponse{
				Ticker:         "AAPL",
				MaxRangeValue:  150.00,
				MaxDailyVolume: 100,
			},
		},
		{
			name: "invalid request - missing ticker",
			reqBody: request.ComputeTickerMetricsRequest{
				Ticker:    "",
				TradeDate: pointer.To("2025-06-08"),
			},
			uc:      NewMockTradesUC(ctrl),
			wantErr: assert.Error,
		},
		{
			name: "invalid request - invalid date format",
			reqBody: request.ComputeTickerMetricsRequest{
				Ticker:    "AAPL",
				TradeDate: pointer.To("invalid-date"),
			},
			wantErr: assert.Error,
		},
		{
			name: "internal server error",
			reqBody: request.ComputeTickerMetricsRequest{
				Ticker:    "AAPL",
				TradeDate: pointer.To("2025-06-08"),
			},
			uc: func() TradesUC {
				uc := NewMockTradesUC(ctrl)
				uc.EXPECT().ComputeTickerMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					decimal.Decimal{}, 0, assert.AnError,
				)
				return uc
			}(),
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			body, err := json.Marshal(tt.reqBody)
			assert.NoError(t, err)
			req := httptest.NewRequest(http.MethodGet, "/ticker-metrics", strings.NewReader(string(body)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewTradesCtrl(tt.uc)
			if !tt.wantErr(t, h.ComputeTickerMetrics(c)) {
				return
			}

			if tt.expectedRes != nil {
				res, err := json.Marshal(tt.expectedRes)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.JSONEq(t, string(res), rec.Body.String())
			}
		})
	}
}
