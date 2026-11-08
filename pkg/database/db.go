package database

import (
	"context"
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
	defer cancel()

	dbPool, err := pgxpool.New(ctx, params.DatabaseURL)
	if err != nil {
		return nil, err
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				return dbPool.Ping(ctx)
			},
			OnStop: func(ctx context.Context) error {
				dbPool.Close()
				return nil
			},
		})

	return dbPool, nil
}
