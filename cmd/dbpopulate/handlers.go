package main

import (
	"b3challenge/internal/domain/entity"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"time"
)

func parseTradeToEntity(r []string) (*entity.Trade, error) {
	if len(r) < 10 {
		return nil, errors.New("invalid record length")
	}

	// fields
	ticker := r[1]
	rawPrice := strings.ReplaceAll(r[3], ",", ".")
	rawQty := r[4]
	rawHour := r[5]
	rawDate := r[8]

	price, err := decimal.NewFromString(rawPrice)
	if err != nil {
		return nil, errors.Wrap(err, "parsing price")
	}
	qty, err := strconv.Atoi(rawQty)
	if err != nil {
		return nil, errors.Wrap(err, "parsing quantity")
	}

	hourPart := rawHour
	if len(rawHour) >= 6 {
		hourPart = rawHour[:6]
	}

	date, err := time.Parse(time.DateOnly, rawDate)
	if err != nil {
		return nil, errors.Wrap(err, "parsing date")
	}

	return entity.NewTrade(ticker, hourPart, date, price, qty), nil
}
