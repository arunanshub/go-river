package main

import (
	"context"
	"fmt"
	"time"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/workers"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		database.Module,
		jobs.ClientModule,
		fx.Invoke(func(c *river.Client[pgx.Tx], sd fx.Shutdowner) error {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			row, err := c.Insert(ctx, workers.SortArgs{
				Strings:  []string{"fefe", "fofo"},
				WaitTime: 2,
			}, nil)
			if err != nil {
				return err
			}

			fmt.Println(row)

			// one-shot: exit once the job is inserted instead of blocking in Run
			return sd.Shutdown()
		}),
	).Run()
}
