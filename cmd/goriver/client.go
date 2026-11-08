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

func (c clientCmd) Run() error {
	app := fx.New(
		workers.Module, jobs.Module, database.Module,
		fx.Invoke(insertJob),
	)

	return runApp(app, func() error {
		<-app.Wait()
		return nil
	})
}

func insertJob(c *jobs.JobClient[workers.SortArgs], sd fx.Shutdowner) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for i := range 230 {
		row, err := c.Insert(ctx, workers.SortArgs{
			Strings:  []string{"fefe", "fofo", strconv.Itoa(i)},
			WaitTime: 5 * time.Second,
		})
		if err != nil {
			return err
		}

		slog.InfoContext(ctx, "row", slog.Any("row", row))
	}

	// one-shot: exit once the job is inserted instead of blocking in Run
	return sd.Shutdown()
}
