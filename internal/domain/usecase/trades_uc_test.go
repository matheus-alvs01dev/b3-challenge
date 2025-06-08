package usecase

import (
	"b3challenge/internal/domain/entity"
	"context"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestNewTradeUC(t *testing.T) {
	repoMock := NewMockTradesRepository(gomock.NewController(t))
	uc := &TradesUC{repo: repoMock}
	got := NewTradesUC(repoMock)
	assert.Equal(t, got, uc)
}

func TestTradeUC_CreateTrades(t *testing.T) {
	expectedTrades := []entity.Trade{
		{Ticker: "AAPL", Price: decimal.NewFromFloat(200.00), Quantity: 100},
		{Ticker: "GOOGL", Price: decimal.NewFromFloat(2800.00), Quantity: 50},
	}

	tests := []struct {
		name     string
		repo     TradesRepository
		wantCode int
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "successful creation",

			repo: func() TradesRepository {
				ctrl := gomock.NewController(t)
				repo := NewMockTradesRepository(ctrl)
				repo.EXPECT().CreateTrades(gomock.Any(), expectedTrades).Return(int64(2), nil)
				return repo
			}(),
			wantCode: 2,
			wantErr:  assert.NoError,
		},
		{
			name: "error case",
			repo: func() TradesRepository {
				ctrl := gomock.NewController(t)
				repo := NewMockTradesRepository(ctrl)
				repo.EXPECT().CreateTrades(gomock.Any(), expectedTrades).Return(int64(0), assert.AnError)
				return repo
			}(),
			wantCode: 0,
			wantErr:  assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &TradesUC{
				repo: tt.repo,
			}
			got, err := uc.CreateTrades(context.Background(), expectedTrades)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.wantCode, got)
		})
	}
}

func TestTradeUC_ComputeTickerMetrics(t *testing.T) {
	expectedTicker := "AAPL"
	expectedMaxRangeValue := decimal.NewFromFloat(150.00)
	expectedMaxDailyValue := 500
	today := time.Now()

	tests := []struct {
		name     string
		repo     TradesRepository
		wantMax  decimal.Decimal
		wantCode int
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "successful metrics computation",
			repo: func() TradesRepository {
				ctrl := gomock.NewController(t)
				repo := NewMockTradesRepository(ctrl)
				repo.EXPECT().ListTradeInfoByTickerAndDate(gomock.Any(), expectedTicker, gomock.Any()).Return(
					[]entity.TradeInfo{
						{Price: decimal.NewFromFloat(100.00), Quantity: 200, Date: today},
						{Price: decimal.NewFromFloat(150.00), Quantity: 300, Date: today},
					}, nil,
				)
				return repo
			}(),
			wantMax:  expectedMaxRangeValue,
			wantCode: expectedMaxDailyValue,
			wantErr:  assert.NoError,
		},
		{
			name: "error case",
			repo: func() TradesRepository {
				ctrl := gomock.NewController(t)
				repo := NewMockTradesRepository(ctrl)
				repo.EXPECT().ListTradeInfoByTickerAndDate(gomock.Any(), expectedTicker, gomock.Any()).Return(nil, assert.AnError)
				return repo
			}(),
			wantMax:  decimal.Decimal{},
			wantCode: 0,
			wantErr:  assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &TradesUC{
				repo: tt.repo,
			}
			gotMax, gotCode, err := uc.ComputeTickerMetrics(context.Background(), expectedTicker, &today)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.wantMax, gotMax)
			assert.Equal(t, tt.wantCode, gotCode)
		})
	}
}
