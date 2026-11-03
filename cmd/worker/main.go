package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type SortArgs struct {
	// Strings is a slice of strings to sort.
	Strings []string `json:"strings"`
}

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

func (SortArgs) Kind() string { return "sort" }

func main() {
	workers := river.NewWorkers()

	// AddWorker panics if the worker is already registered or invalid:
	river.AddWorker(workers, &SortWorker{})

	ctx := context.TODO()
	dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		panic(err)
	}

	// Run the client inline. All executed jobs will inherit from ctx:
	if err := riverClient.Start(ctx); err != nil {
		panic(err)
	}

	res, err := riverClient.Insert(ctx, SortArgs{
		Strings: []string{
			"whale", "tiger", "bear",
		},
	}, nil)
	fmt.Println(res)
	if err != nil {
		panic(err)
	}

	// Stop fetching new work and wait for active jobs to finish.
	if err := riverClient.Stop(ctx); err != nil {
		panic(err)
	}
}
