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
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	diContainer, ctx, cancel := setupComponents()
	defer cancel()
	defer diContainer.DB().Close()

	parserWorkers := config.GetParserWorkersCount()
	batchSize := config.GetBatchSize()
	dbWorkers := config.GetDBWorkersCount()

	tradesCh := make(chan entity.Trade, parserWorkers)
	jobCh := make(chan string)
	dbCh := make(chan []entity.Trade, dbWorkers)

	start := time.Now()
	files, err := findTXTFiles("b3Data")
	if err != nil {
		log.Fatalf("Error finding TXT files: %v", err)
	}

	var dbWg sync.WaitGroup
	startDBWorkers(ctx, dbCh, diContainer.GetTradesUC(), dbWorkers, &dbWg)

	var parserWg sync.WaitGroup
	startParserWorkers(ctx, parserWorkers, jobCh, tradesCh, &parserWg)

	go func() {
		for _, f := range files {
			jobCh <- f
		}
		close(jobCh)

		log.Info("All jobs dispatched. Waiting for parsers to finish...")
		parserWg.Wait()

		log.Info("All parsing workers finished. Closing trades channel.")
		close(tradesCh)
	}()

	processAndBatchTrades(ctx, tradesCh, dbCh, batchSize)

	log.Info("Main process finished, waiting for DB workers to commit last batches...")
	dbWg.Wait()

	duration := time.Since(start)
	log.Infof("Application finished. Total processing time: %s", duration)
}

func processAndBatchTrades(ctx context.Context, tradesIn <-chan entity.Trade, dbOut chan<- []entity.Trade, batchSize int) {
	batch := make([]entity.Trade, 0, batchSize)
	defer close(dbOut)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Context cancelled, flushing final batch and stopping.")
			if len(batch) > 0 {
				dbOut <- batch
			}
			return

		case tr, ok := <-tradesIn:
			if !ok {
				log.Info("Trades channel closed, flushing final batch.")
				if len(batch) > 0 {
					dbOut <- batch
				}
				return
			}
			batch = append(batch, tr)
			if len(batch) >= batchSize {
				log.Infof("Batch size reached, sending %d trades", len(batch))
				dbOut <- batch
				batch = make([]entity.Trade, 0, batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				log.Infof("Ticker triggered, sending incomplete batch of %d trades", len(batch))
				dbOut <- batch
				batch = make([]entity.Trade, 0, batchSize)
			}
		}
	}
}

func startParserWorkers(ctx context.Context, numWorkers int, jobCh <-chan string, tradesCh chan<- entity.Trade, wg *sync.WaitGroup) {
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobCh {
				select {
				case <-ctx.Done():
					log.Infof("Parser worker shutting down due to context cancellation.")
					return
				default:
					if err := parseFileToTrades(ctx, file, tradesCh); err != nil {
						log.Errorf("Error parsing file %s: %v", file, err)
						continue
					}
				}
			}
		}()
	}
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

func parseFileToTrades(ctx context.Context, filePath string, out chan<- entity.Trade) error {
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
			log.Info("Context cancelled, stopping file parsing")
			return ctx.Err()

		default:
			rec, err := reader.Read()
			if err == io.EOF {
				return nil
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
	}
}

func setupComponents() (*di.Container, context.Context, context.CancelFunc) {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading env configs: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbClient, err := db.NewClient(config.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Error initializing database client: %v", err)
	}

	diContainer := di.NewContainer(dbClient.DB())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("Received shutdown signal, cancelling context...")
		cancel()
	}()

	return diContainer, ctx, cancel
}

func startDBWorkers(ctx context.Context, dbCh chan []entity.Trade, uc *usecase.TradesUC, workerCount int, wg *sync.WaitGroup) {
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range dbCh {
				select {
				case <-ctx.Done():
					log.Infof("DB worker %d shutting down due to context cancellation.", id)
					return

				default:
					if len(batch) == 0 {
						continue
					}
					count, err := uc.CreateTrades(ctx, batch)
					if err != nil {
						log.Errorf("DB worker %d error: %v", id, err)
						continue
					}
					log.Infof("DB worker %d wrote %d trades", id, count)
				}
			}
		}(i)
	}
}
