package workers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/riverqueue/river"
)

type SortArgs struct {
	// Strings is a slice of strings to sort.
	Strings []string `json:"strings"`

	// added new arg
	WaitTime time.Duration
}

func (SortArgs) Kind() string { return "sort" }

// route sort jobs to the queue the worker listens on, regardless of which
// client inserts them
func (SortArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: string(jobs.Slow)}
}

type SortWorker struct {
	// An embedded WorkerDefaults sets up default methods to fulfill the rest of
	// the Worker interface:
	river.WorkerDefaults[SortArgs]
}

func (w *SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
	sort.Strings(job.Args.Strings)
	time.Sleep(job.Args.WaitTime)
	fmt.Printf("Sorted strings: %+v\n", job.Args.Strings)
	return nil
}

var _ river.Worker[SortArgs] = (*SortWorker)(nil)

// func NewSortWorker(registry *river.Workers) (river.Worker[SortArgs], error) {
// 	worker := &SortWorker{}
// 	err := river.AddWorkerSafely(registry, worker)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return worker, nil
// }

func NewSortWorker() jobs.WorkerResult {
	return jobs.AsWorker(&SortWorker{})
}
