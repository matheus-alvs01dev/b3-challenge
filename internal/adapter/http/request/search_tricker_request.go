package request

type GetTickerDataRequest struct {
	Ticker       *string `query:"ticker"`
	TradeDateGTE *string `query:"tradeDate[gte]"`
	TradeDateLTE *string `query:"tradeDate[lte]"`
}
