-- name: CreateTrades :copyfrom
INSERT INTO trades (hour, date, ticker, price, quantity)
VALUES ($1, $2, $3, $4, $5);

-- name: ListTradeInfoByTickerAndDate :many
SELECT
    date,
    price,
    quantity
FROM trades
WHERE ticker = @ticker
  AND (@trade_date::date IS NULL OR date >= @trade_date::date);