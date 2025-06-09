package main

import (
	"b3challenge/cmd/dbpopulate/filehandler"
	"b3challenge/config"
	"b3challenge/internal/adapter/db"
	"b3challenge/internal/di"
	"b3challenge/internal/domain/entity"
	"b3challenge/internal/domain/usecase"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := setupLogger()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer logger.Sync() //nolint:errcheck

	diContainer, ctx, cancel := setupComponents(logger)
	defer cancel()
	defer diContainer.DB().Close()

	parserWorkers := config.GetParserWorkersCount()
	batchSize := config.GetBatchSize()
	dbWorkers := config.GetDBWorkersCount()

	tradesCh := make(chan entity.Trade, parserWorkers)
	jobCh := make(chan string)
	dbCh := make(chan []entity.Trade, dbWorkers)

	start := time.Now()
	files, err := filehandler.FindTXTFiles("b3Data")
	if err != nil {
		logger.Error("Error finding TXT files: ", zap.Error(err))
	}

	logger.Info("Found TXT files: ", zap.Strings("files", files))

	var dbWg sync.WaitGroup
	startDBWorkers(ctx, dbCh, diContainer.GetTradesUC(), dbWorkers, &dbWg, logger)

	var parserWg sync.WaitGroup
	startParserWorkers(ctx, parserWorkers, jobCh, tradesCh, &parserWg, logger)

	go func() {
		for _, f := range files {
			jobCh <- f
		}
		close(jobCh)

		logger.Info("All jobs dispatched. Waiting for parsers to finish...")
		parserWg.Wait()

		logger.Info("All parsing workers finished. Closing trades channel.")
		close(tradesCh)
	}()

	processAndBatchTrades(ctx, tradesCh, dbCh, batchSize, logger)

	logger.Info("Main process finished, waiting for DB workers to commit last batches...")
	dbWg.Wait()

	duration := time.Since(start)
	logger.Info("Application finished.", zap.Duration(" Total processing time", duration))
}

func processAndBatchTrades(
	ctx context.Context,
	tradesIn <-chan entity.Trade,
	dbOut chan<- []entity.Trade,
	batchSize int,
	logger *zap.Logger,
) {
	batch := make([]entity.Trade, 0, batchSize)
	defer close(dbOut)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, flushing final batch and stopping.")
			if len(batch) > 0 {
				dbOut <- batch
			}

			return

		case tr, ok := <-tradesIn:
			if !ok {
				logger.Info("Trades channel closed, flushing final batch.")
				if len(batch) > 0 {
					dbOut <- batch
				}

				return
			}
			batch = append(batch, tr)
			if len(batch) >= batchSize {
				logger.Info("Batch size reached. ", zap.Int("size", len(batch)))
				dbOut <- batch
				batch = make([]entity.Trade, 0, batchSize)
			}
		}
	}
}

func setupLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "zap")
	}

	return logger, nil
}

func setupComponents(logger *zap.Logger) (*di.Container, context.Context, context.CancelFunc) {
	err := config.LoadConfig()
	if err != nil {
		logger.Error("Error loading env configs:", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbClient, err := db.NewClient(config.GetDatabaseDSN())
	if err != nil {
		logger.Error("Error initializing database client", zap.Error(err))
	}

	diContainer := di.NewContainer(dbClient.DB())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("Received shutdown signal, cancelling context...")
		cancel()
	}()

	return diContainer, ctx, cancel
}

func startParserWorkers(
	ctx context.Context,
	numWorkers int,
	jobCh <-chan string,
	tradesCh chan<- entity.Trade,
	wg *sync.WaitGroup,
	logger *zap.Logger,
) {
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobCh {
				select {
				case <-ctx.Done():
					logger.Info("Parser worker shutting down due to context cancellation.")

					return

				default:
					if err := filehandler.ParseFileToTrades(ctx, file, tradesCh, logger); err != nil {
						logger.Error("Error parsing file:", zap.Any("file", file), zap.Error(err))

						continue
					}
				}
			}
		}()
	}
}

func startDBWorkers(
	ctx context.Context,
	dbCh chan []entity.Trade,
	uc *usecase.TradesUC,
	workerCount int,
	wg *sync.WaitGroup,
	logger *zap.Logger,
) {
	for i := range workerCount {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range dbCh {
				select {
				case <-ctx.Done():
					logger.Info(
						"DB worker shutting down due to context cancellation.",
						zap.Any("worker_id", id),
					)

					return

				default:
					if len(batch) == 0 {
						continue
					}

					count, err := uc.CreateTrades(ctx, batch)
					if err != nil {
						logger.Error("DB worker error", zap.Int("worker_id", id), zap.Error(err))

						continue
					}
					logger.Info(
						"DB worker wrote batch",
						zap.Int("worker_id", id),
						zap.Int("trades_written", count),
					)
				}
			}
		}(i)
	}
}
