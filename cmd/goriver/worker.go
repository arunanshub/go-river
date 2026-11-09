package main

import (
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/users"
	"github.com/arunanshub/go-river/pkg/workers"
)

type workerCmd struct{}

func (w workerCmd) Run(common sharedFlags) error {
	app := fx.New(workers.Module, jobs.Module, database.Module, users.Module,
		fx.Supply(database.DatabasePoolConfig{DatabaseURL: common.DatabaseURL}),
		fx.Invoke(jobs.StartClient),
	)

	return runApp(app, func() error {
		<-app.Wait()
		return nil
	})
}
