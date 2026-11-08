package main

import (
	"context"
	"errors"

	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/workers"
)

type workerCmd struct{}

func (w workerCmd) Run() error {
	app := fx.New(workers.Module, jobs.Module, database.Module,
		fx.Invoke(jobs.StartClient),
	)

	return runApp(app, func() error {
		<-app.Wait()
		return nil
	})
}

func runApp(app *fx.App, callbackFn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()
	if err := app.Start(ctx); err != nil {
		return err
	}

	callbackErr := callbackFn()

	stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	defer cancel()
	return errors.Join(callbackErr, app.Stop(stopCtx))
}
