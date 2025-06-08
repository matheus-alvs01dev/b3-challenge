package response

import "github.com/shopspring/decimal"

type ComputeTickerMetricsResponse struct {
	Ticker         string  `json:"ticker"`
	MaxRangeValue  float64 `json:"max_range_value"`
	MaxDailyVolume int     `json:"max_daily_volume"`
}

func NewComputeTickerMetricsResponse(
	ticker string,
	maxRangeValue decimal.Decimal,
	maxDailyVolume int,
) ComputeTickerMetricsResponse {
	return ComputeTickerMetricsResponse{
		Ticker:         ticker,
		MaxRangeValue:  maxRangeValue.InexactFloat64(),
		MaxDailyVolume: maxDailyVolume,
	}
}
