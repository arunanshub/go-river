package jobs

import (
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

// option tunes per-job insert options. Opaque, so nobody outside this
// package can write an option that touches the queue.
type option struct{ apply func(*river.InsertOpts) }

func MaxAttempts(n int) option {
	return option{func(o *river.InsertOpts) { o.MaxAttempts = n }}
}

func Priority(p int) option {
	return option{func(o *river.InsertOpts) { o.Priority = p }}
}

func Unique(u river.UniqueOpts) option {
	return option{func(o *river.InsertOpts) { o.UniqueOpts = u }}
}

// A function that adds a worker to the registry. Workers are contributed to the
// "workers" value group as these closures since each [river.Worker] is a
// distinct type.
type registrableWorkerFunc func(*river.Workers) error

// A registrable worker is a type erased worker, that may be pending
// registration with [river.Workers] worker registry.
type registrableWorker struct {
	fx.Out

	Queue                 Queue                 `group:"queues"`
	RegistrableWorkerFunc registrableWorkerFunc `group:"workers"`
}

// workerConstructor is a function that creates a river worker, or fails in the
// process.
type workerConstructor[A river.JobArgs, W river.Worker[A], P any] func(P) (W, error)

// Given a queue and a worker's constructor, AsWorkerModule turns it into an fx module.
func AsWorkerModule[A river.JobArgs, W river.Worker[A], P any](
	queue Queue,
	workerCtor workerConstructor[A, W, P],
	opts ...option,
) fx.Option {
	insertOpts := river.InsertOpts{}
	for _, opt := range opts {
		opt.apply(&insertOpts)
	}
	// make sure our queue wins
	insertOpts.Queue = queue.name

	return fx.Options(
		// step 1: have the worker constructed with whatever params are needed.
		// the outside world need not access the concrete worker type.
		fx.Provide(workerCtor, fx.Private),

		// step 2: consume the worker by registering it with the registry
		// This gets used by whoever wants to register the worker.
		fx.Provide(func(worker W) registrableWorker {
			return registrableWorker{Queue: queue, RegistrableWorkerFunc: func(registry *river.Workers) error {
				return river.AddWorkerSafely(registry, worker)
			}}
		}),

		// step 3: provide a proof that registration has happened.
		// This gets consumed by one who wants to use the insert-only client.
		// This way, the worker is never constructed if it's never needed by anyone.
		fx.Provide(func(c insertOnlyJobClient) *JobClient[A] {
			return &JobClient[A]{client: c.client, insertOpts: insertOpts}
		}),
	)
}
