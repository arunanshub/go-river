package workers

import "go.uber.org/fx"

var Module = fx.Module("job.workers",
	sortArgsModule,
)
