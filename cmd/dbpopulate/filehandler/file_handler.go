package filehandler

import (
	"b3challenge/internal/domain/entity"
	"context"
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	minRecordLength     = 10
	minHourLength       = 6
	tickerColumnIndex   = 1
	priceColumnIndex    = 3
	quantityColumnIndex = 4
	hourColumnIndex     = 5
	dateColumnIndex     = 8
)

func FindTXTFiles(pathDir string) ([]string, error) {
	entries, err := os.ReadDir(pathDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading directory %s", pathDir)
	}
	var list []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			path := filepath.Join(pathDir, e.Name())
			list = append(list, path)
		}
	}

	return list, nil
}

func ParseFileToTrades(ctx context.Context, filePath string, out chan<- entity.Trade) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.Wrap(err, "cannot open file")
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	if _, err := reader.Read(); err != nil {
		return errors.Wrap(err, "reading header")
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping file parsing")

			return errors.Wrap(ctx.Err(), "context cancelled")

		default:
			rec, err := reader.Read()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				slog.Error("CSV read error: ", slog.Any("error", err))

				continue
			}
			trade, err := parseTradeToEntity(rec)
			if err != nil {
				slog.Error("parsing trade: ", slog.Any("error", err))

				continue
			}
			out <- *trade
		}
	}
}

func parseTradeToEntity(rows []string) (*entity.Trade, error) {
	if len(rows) < minRecordLength {
		return nil, errors.New("invalid record length")
	}

	ticker := rows[tickerColumnIndex]
	rawPrice := strings.ReplaceAll(rows[priceColumnIndex], ",", ".")
	rawQty := rows[quantityColumnIndex]
	rawHour := rows[hourColumnIndex]
	rawDate := rows[dateColumnIndex]

	price, err := decimal.NewFromString(rawPrice)
	if err != nil {
		return nil, errors.Wrap(err, "parsing price")
	}
	qty, err := strconv.ParseInt(rawQty, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "parsing quantity")
	}

	hourPart := rawHour
	if len(rawHour) >= minHourLength {
		hourPart = rawHour[:minHourLength]
	}

	date, err := time.Parse(time.DateOnly, rawDate)
	if err != nil {
		return nil, errors.Wrap(err, "parsing date")
	}

	return entity.NewTrade(ticker, hourPart, date, price, int32(qty)), nil
}
