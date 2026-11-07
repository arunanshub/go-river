package workers

import (
	"github.com/riverqueue/river"
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/jobs"
)

var Module = fx.Module("job.workers", fx.Provide(
	AsWorker(NewSortWorker),
))

// Contributes the worker built by ctor to the "workers" value group, which the
// registry consumes and registers.
func AsWorker[T river.JobArgs](ctor func() river.Worker[T]) any {
	return fx.Annotate(
		func() jobs.WorkerRegistration {
			return func(registry *river.Workers) error {
				return river.AddWorkerSafely(registry, ctor())
			}
		},
		fx.ResultTags(`group:"workers"`),
	)
}
