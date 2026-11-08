package main

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/workers"
)

type clientCmd struct{}

func (c clientCmd) Run(common sharedFlags) error {
	var client *jobs.JobClient[workers.SortArgs]
	app := fx.New(
		workers.Module, jobs.Module, database.Module,
		fx.Supply(database.DatabasePoolConfig{DatabaseURL: common.DatabaseURL}),
		fx.Populate(&client),
	)

	// one-shot: insert once the app (and the pool) has started, then stop
	return runApp(app, func() error {
		return insertJobs(client)
	})
}

func insertJobs(c *jobs.JobClient[workers.SortArgs]) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := make([]workers.SortArgs, 230)
	for i := range args {
		args[i] = workers.SortArgs{
			Strings:  []string{"fefe", "fofo", strconv.Itoa(i)},
			WaitTime: 5 * time.Second,
		}
	}

	rows, err := c.InsertMany(ctx, args)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "inserted jobs", slog.Int("count", len(rows)))
	return nil
}
