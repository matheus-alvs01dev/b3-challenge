package main

import (
	"b3challenge/config"
	"b3challenge/internal/adapter/db"
	"b3challenge/internal/di"
	"b3challenge/internal/domain/entity"
	"b3challenge/internal/domain/usecase"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	diContainer, ctx, cancel := setupComponents()
	defer cancel()
	defer diContainer.DB().Close()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	parserWorkers := config.GetParserWorkersCount()
	batchSize := config.GetBatchSize()
	dbWorkers := config.GetDBWorkersCount()

	tradesCh := make(chan entity.Trade, parserWorkers)
	jobCh := make(chan string)
	dbCh := make(chan []entity.Trade, dbWorkers)

	start := time.Now()
	files, err := findTXTFiles("b3Data")
	if err != nil {
		slog.Error("Error finding TXT files: ", slog.Any("err", err))
	}

	slog.Info("Found TXT files: ", slog.Any("files", files))

	var dbWg sync.WaitGroup
	startDBWorkers(ctx, dbCh, diContainer.GetTradesUC(), dbWorkers, &dbWg)

	var parserWg sync.WaitGroup
	startParserWorkers(ctx, parserWorkers, jobCh, tradesCh, &parserWg)

	go func() {
		for _, f := range files {
			jobCh <- f
		}
		close(jobCh)

		slog.Info("All jobs dispatched. Waiting for parsers to finish...")
		parserWg.Wait()

		slog.Info("All parsing workers finished. Closing trades channel.")
		close(tradesCh)
	}()

	processAndBatchTrades(ctx, tradesCh, dbCh, batchSize)

	slog.Info("Main process finished, waiting for DB workers to commit last batches...")
	dbWg.Wait()

	duration := time.Since(start)
	slog.Info("Application finished.", slog.Duration(" Total processing time", duration))
}

func processAndBatchTrades(
	ctx context.Context,
	tradesIn <-chan entity.Trade,
	dbOut chan<- []entity.Trade,
	batchSize int,
) {
	batch := make([]entity.Trade, 0, batchSize)
	defer close(dbOut)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, flushing final batch and stopping.")
			if len(batch) > 0 {
				dbOut <- batch
			}

			return

		case tr, ok := <-tradesIn:
			if !ok {
				slog.Info("Trades channel closed, flushing final batch.")
				if len(batch) > 0 {
					dbOut <- batch
				}

				return
			}
			batch = append(batch, tr)
			if len(batch) >= batchSize {
				slog.Info("Batch size reached. ", slog.Int("size", len(batch)))
				dbOut <- batch
				batch = make([]entity.Trade, 0, batchSize)
			}
		}
	}
}

func startParserWorkers(
	ctx context.Context,
	numWorkers int,
	jobCh <-chan string,
	tradesCh chan<- entity.Trade,
	wg *sync.WaitGroup,
) {
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobCh {
				select {
				case <-ctx.Done():
					slog.Info("Parser worker shutting down due to context cancellation.")

					return

				default:
					if err := parseFileToTrades(ctx, file, tradesCh); err != nil {
						slog.Error("Error parsing file %s: %v", file, err)

						continue
					}
				}
			}
		}()
	}
}

func setupComponents() (*di.Container, context.Context, context.CancelFunc) {
	err := config.LoadConfig()
	if err != nil {
		slog.Error("Error loading env configs: %v", slog.Any("err", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbClient, err := db.NewClient(config.GetDatabaseDSN())
	if err != nil {
		slog.Error("Error initializing database client", slog.Any("err", err))
	}

	diContainer := di.NewContainer(dbClient.DB())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("Received shutdown signal, cancelling context...")
		cancel()
	}()

	return diContainer, ctx, cancel
}

func startDBWorkers(
	ctx context.Context,
	dbCh chan []entity.Trade,
	uc *usecase.TradesUC,
	workerCount int,
	wg *sync.WaitGroup,
) {
	for i := range workerCount {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range dbCh {
				select {
				case <-ctx.Done():
					slog.Info(
						"DB worker shutting down due to context cancellation.",
						slog.Any("worker_id", id),
					)

					return

				default:
					if len(batch) == 0 {
						continue
					}

					count, err := uc.CreateTrades(ctx, batch)
					if err != nil {
						slog.Error("DB worker error", slog.Int("worker_id", id), slog.Any("err", err))

						continue
					}
					slog.Info(
						"DB worker wrote batch",
						slog.Int("worker_id", id),
						slog.Int("trades_written", count),
					)
				}
			}
		}(i)
	}
}
