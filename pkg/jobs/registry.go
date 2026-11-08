package jobs

import (
	"errors"

	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

type RegistryParams struct {
	fx.In

	Registrations []registrableWorkerFunc `group:"workers"`
}

// Depending on the "workers" group forces every worker in it to be constructed
// and registered before the client that depends on this registry is built.
func NewWorkersRegistry(params RegistryParams) (*river.Workers, error) {
	if len(params.Registrations) == 0 {
		return nil, errors.New("no workers were registered. did you provide a worker?")
	}

	registry := river.NewWorkers()

	for _, register := range params.Registrations {
		if err := register(registry); err != nil {
			return nil, err
		}
	}
	return registry, nil
}
