package jobs

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/fx"
)

type QueueType string

const (
	Slow QueueType = "slow"
	Fast QueueType = "fast"
)

var Module = fx.Module("jobs",
	fx.Provide(NewJobClient, NewWorkersRegistry, NewInsertOnlyClient, NewQueueConfig),
)

func StartClient(lc fx.Lifecycle, client jobClient) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// river runs for the lifetime of ctx, but fx cancels OnStart's ctx once
			// startup finishes (and it carries the 15s start timeout), so detach it
			return client.client.Start(context.WithoutCancel(ctx))
		},
		OnStop: func(ctx context.Context) error {
			return client.client.Stop(ctx)
		},
	})
}

type ClientParams struct {
	fx.In

	Pool        *pgxpool.Pool
	QueueConfig map[string]river.QueueConfig
	// responsible for having the workers registered
	Registry *river.Workers
}

type jobClient struct {
	client *river.Client[pgx.Tx]
}

// This returns my client that is not generic.
func NewJobClient(params ClientParams) (jobClient, error) {
	client, err := river.NewClient(riverpgxv5.New(params.Pool), &river.Config{
		Queues: params.QueueConfig,
		// so this is where the client gets to know about the workers
		// but how do you run just one worker per process?
		Workers: params.Registry,
	})
	if err != nil {
		return jobClient{}, fmt.Errorf("failed to create job client: %w", err)
	}

	return jobClient{client}, err
}
