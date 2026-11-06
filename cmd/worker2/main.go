package main

import (
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
)

func main() {
	app := fx.New(jobs.Module, database.Module)

	app.Run()
}
