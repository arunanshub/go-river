package workers

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/riverqueue/river"
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/jobs"
)

type SortArgs struct {
	// Strings is a slice of strings to sort.
	Strings []string `json:"strings"`

	// added new arg
	WaitTime time.Duration `json:"wait_time"`
}

func (SortArgs) Kind() string { return "sort" }

type SortWorker struct {
	// An embedded WorkerDefaults sets up default methods to fulfill the rest of
	// the Worker interface:
	river.WorkerDefaults[SortArgs]
}

func (w SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
	sort.Strings(job.Args.Strings)
	time.Sleep(job.Args.WaitTime)
	slog.InfoContext(ctx, "sorted", slog.Any("args", job.Args))
	return nil
}

var _ river.Worker[SortArgs] = (*SortWorker)(nil)

type NewSortWorkerParams struct {
	fx.In
}

func NewSortWorker(params NewSortWorkerParams) (river.Worker[SortArgs], error) {
	return SortWorker{}, nil
}

var sortArgsModule = jobs.AsWorkerModule(jobs.FastQueue(), NewSortWorker)
