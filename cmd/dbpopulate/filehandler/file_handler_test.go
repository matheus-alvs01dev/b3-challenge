package filehandler

import (
	"b3challenge/internal/domain/entity"
	"context"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func readFileTest(t *testing.T, file string) []byte {
	b, err := os.ReadFile(file)
	assert.NoError(t, err)

	return b
}

func TestFindTXTFiles(t *testing.T) {
	input := "testdata"
	files, err := FindTXTFiles(input)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "testdata/mock-csv.txt", files[0])
}

func TestParseFileToTrades(t *testing.T) {
	expectedTrades := []entity.Trade{
		{
			Ticker:   "TF583R",
			Price:    decimal.New(10000, -3),
			Quantity: 10000,
			Hour:     "030507",
			Date:     time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			Ticker:   "FRCQ25",
			Price:    decimal.New(5740, -3),
			Quantity: 870,
			Hour:     "090000",
			Date:     time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	out := make(chan entity.Trade, 2)
	gotErr := ParseFileToTrades(context.Background(), "testdata/mock-csv.txt", out)
	assert.NoError(t, gotErr)
	close(out)

	for _, want := range expectedTrades {
		got, ok := <-out
		assert.True(t, ok)
		assert.Equal(t, want.Ticker, got.Ticker)
		assert.Equal(t, want.Price, got.Price)
		assert.Equal(t, want.Quantity, got.Quantity)
		assert.Equal(t, want.Hour, got.Hour)
		assert.Equal(t, want.Date, got.Date)
	}
}
