package db

import (
	"b3challenge/internal/adapter/db/migrations"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

type Client struct {
	db *pgxpool.Pool
}

func NewClient(dsn string) (*Client, error) {
	dbPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	goose.SetBaseFS(migrations.Embed)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return nil, errors.Wrap(err, "failed to set dialect")
	}

	db := stdlib.OpenDBFromPool(dbPool)
	if err := goose.Up(db, "."); err != nil {
		return nil, errors.Wrap(err, "run migrations")
	}
	if err := db.Close(); err != nil {
		return nil, errors.Wrap(err, "db close")
	}

	return &Client{db: dbPool}, nil
}

func (c *Client) Close() {
	c.db.Close()
}

func (c *Client) DB() *pgxpool.Pool {
	return c.db
}
