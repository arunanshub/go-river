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
	fx.Provide(NewJobClient, NewWorkersRegistry),
	// fx.Invoke(StartClient),
	// fx.Invoke(addWorkersToRegistry),
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

	Pool *pgxpool.Pool
	// responsible for having the workers registered
	Registry *river.Workers
}

// This returns my client that is not generic
func NewJobClient(params ClientParams) (jobClient, error) {
	client, err := river.NewClient(riverpgxv5.New(params.Pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			string(Slow): {MaxWorkers: 100},
		},
		// so this is where the client gets to know about the workers
		// but how do you run just one worker per process?
		Workers: params.Registry,
	})
	if err != nil {
		return jobClient{}, fmt.Errorf("failed to create job client: %w", err)
	}

	return jobClient{client}, err
}

type jobClient struct {
	client *river.Client[pgx.Tx]
}

// A function that adds a worker to the registry. Workers are contributed to the
// "workers" value group as these closures since each river.Worker[T] is a
// distinct type.
type WorkerRegistration func(*river.Workers) error

// Returned by worker constructors to contribute the worker to the "workers"
// value group without needing fx.Annotate.
type WorkerResult struct {
	fx.Out

	Registration WorkerRegistration `group:"workers"`
}

// Wraps worker as a WorkerResult. T is inferred from worker's Work method, so
// callers never restate the args type.
func AsWorker[T river.JobArgs](worker river.Worker[T]) WorkerResult {
	return WorkerResult{Registration: func(registry *river.Workers) error {
		return river.AddWorkerSafely(registry, worker)
	}}
}

type registryParams struct {
	fx.In

	Registrations []WorkerRegistration `group:"workers"`
}

// Depending on the "workers" group forces every worker in it to be constructed
// and registered before the client that depends on this registry is built.
func NewWorkersRegistry(params registryParams) (*river.Workers, error) {
	registry := river.NewWorkers()
	for _, register := range params.Registrations {
		if err := register(registry); err != nil {
			return nil, err
		}
	}
	return registry, nil
}
