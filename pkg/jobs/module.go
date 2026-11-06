package jobs

import (
	"context"
	"fmt"
	"sort"
	"time"

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
	fx.Provide(NewJobClient),
	fx.Provide(newWorkersRegistry, NewSortWorker, fx.Private),
	fx.Invoke(StartClient),
	fx.Invoke(addWorkersToRegistry),
)

func StartClient(lc fx.Lifecycle, client jobClient) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return client.client.Start(ctx)
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

// -------------------- this is where the workers get registered -----------------
func newWorkersRegistry() *river.Workers {
	return river.NewWorkers()
}

type AddWorkersParams struct {
	fx.In

	Registry *river.Workers

	// just one worker type. There can be many. The point is, registry only ever sees
	// the interface
	SortWorker river.Worker[SortArgs]
}

func addWorkersToRegistry(params AddWorkersParams) error {
	registry := params.Registry

	err := river.AddWorkerSafely(registry, params.SortWorker)
	return err
}

// ----------------------------- a worker's args -----------------------------
type SortArgs struct {
	// Strings is a slice of strings to sort.
	Strings []string `json:"strings"`
}

func (SortArgs) Kind() string { return "sort" }

// ----------------------------- the actual worker ----------------------------
type SortWorker struct {
	// An embedded WorkerDefaults sets up default methods to fulfill the rest of
	// the Worker interface:
	river.WorkerDefaults[SortArgs]
}

func (w *SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
	sort.Strings(job.Args.Strings)
	time.Sleep(10 * time.Second)
	fmt.Printf("Sorted strings: %+v\n", job.Args.Strings)
	return nil
}

func NewSortWorker() river.Worker[SortArgs] {
	return &SortWorker{}
}
