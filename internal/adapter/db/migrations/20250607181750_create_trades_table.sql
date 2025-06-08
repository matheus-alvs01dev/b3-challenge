-- +goose Up
-- +goose StatementBegin
CREATE TABLE trades
(
    id         SERIAL PRIMARY KEY,
    hour       TEXT           NOT NULL,
    date       DATE           NOT NULL,
    ticker     TEXT           NOT NULL,
    price      DECIMAL(18, 2) NOT NULL,
    quantity   INTEGER        NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_trades_ticker ON trades (ticker);
CREATE INDEX idx_trades_date ON trades (date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE trades;
-- +goose StatementEnd
