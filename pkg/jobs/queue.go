package jobs

import (
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

type Queue struct {
	name     string
	defaults river.QueueConfig
}

// Listed as unexported variables here so that no one can modify them. The only
// way to consume queues is to build them with [SlowQueue], [FastQueue] etc.
var (
	slowQueue = Queue{"slow_queue", river.QueueConfig{MaxWorkers: 30}}
	fastQueue = Queue{"fast_queue", river.QueueConfig{MaxWorkers: 200}}
)

func SlowQueue() Queue {
	return slowQueue
}

func FastQueue() Queue {
	return fastQueue
}

type QueueConfigParams struct {
	fx.In

	Queues []Queue `group:"queues"`
}

func NewQueueConfig(params QueueConfigParams) map[string]river.QueueConfig {
	queues := make(map[string]river.QueueConfig, len(params.Queues))
	for _, queue := range params.Queues {
		queues[queue.name] = queue.defaults
	}
	return queues
}
