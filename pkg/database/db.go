package database

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("database", fx.Provide(NewDatabasePool))

type DatabasePoolConfig struct {
	DatabaseURL string
}

func NewDatabasePool(lc fx.Lifecycle, params DatabasePoolConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				cancel()
				dbPool.Close()
				return nil
			},
			OnStart: func(ctx context.Context) error {
				return dbPool.Ping(ctx)
			},
		})

	return dbPool, err
}
