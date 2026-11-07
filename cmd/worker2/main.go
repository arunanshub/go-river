package main

import (
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/workers"
)

func main() {
	app := fx.New(workers.Module, jobs.Module, database.Module,
		fx.Invoke(jobs.StartClient),
	)

	app.Run()
}
