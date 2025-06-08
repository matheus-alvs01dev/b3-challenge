package main

import (
	"b3challenge/config"
	"b3challenge/internal/adapter/db"
	"b3challenge/internal/di"
	"b3challenge/internal/domain/entity"
	"b3challenge/internal/domain/usecase"
	"context"
	"encoding/csv"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading env configs: %v", err)
	}

	workerCount := config.GetWorkerCount()
	batchSize := config.GetBatchSize()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbClient, err := db.NewClient(config.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}

	diContainer := di.NewContainer(dbClient.DB())
	uc := diContainer.GetTradesUC()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("Received shutdown signal, cancelling context...")
		cancel()
	}()

	filesList, err := findTXTFiles("b3Data")
	if err != nil {
		log.Fatalf("Error finding TXT files: %v", err)
	}

	tradesCh := make(chan entity.Trade, workerCount)
	jobCh := make(chan string)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobCh {
				if err := parseFileToTrades(file, tradesCh); err != nil {
					log.Error("Error parsing file %s: %v", file, err)
					continue
				}
			}
		}()
	}

	go func() {
		for _, p := range filesList {
			jobCh <- p
		}
		close(jobCh)
	}()

	log.Info("Waiting for workers to finish...")

	go func() {
		wg.Wait()
		close(tradesCh)
	}()

	batch := make([]entity.Trade, 0, batchSize)
	for {
		select {
		case <-ctx.Done():
			return
		case tr, ok := <-tradesCh:
			if !ok {
				if len(batch) > 0 {
					handleBatch(ctx, batch, uc)
				}
				return
			}
			batch = append(batch, tr)

			if len(batch) == batchSize {
				cur := batch
				go handleBatch(ctx, cur, uc)
				batch = make([]entity.Trade, 0, batchSize)
			}
		}
	}
}

func handleBatch(ctx context.Context, trades []entity.Trade, uc *usecase.TradesUC) error {
	count, err := uc.CreateTrades(ctx, trades)
	if err != nil {
		log.Errorf("Error inserting batch of trades: %v", err)
		return err
	}
	log.Infof("Inserted batch of %d trades", count)
	return nil
}

func findTXTFiles(pathDir string) ([]string, error) {
	entries, err := os.ReadDir(pathDir)
	if err != nil {
		return nil, err
	}
	var list []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			path := filepath.Join(pathDir, e.Name())
			log.Infof("Found TXT file: %s", path)
			list = append(list, path)
		}
	}
	return list, nil
}

func parseFileToTrades(filePath string, out chan<- entity.Trade) error {
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
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("CSV read error: ", err)
			continue
		}
		trade, err := parseTradeToEntity(rec)
		if err != nil {
			log.Error("parsing trade: ", err)
			continue
		}
		out <- *trade
	}
	return nil
}

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

	price, err := strconv.ParseFloat(rawPrice, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parsing price")
	}
	qty, err := strconv.Atoi(rawQty)
	if err != nil {
		return nil, errors.Wrap(err, "parsing quantity")
	}
	// parse time HHMMSSXXX => HHMMSS
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
