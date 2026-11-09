package main

import (
	"cmp"
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/database"
	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/users"
	"github.com/arunanshub/go-river/pkg/workers"
)

type userCmd struct {
	Register userRegisterCmd `cmd:"" help:"register a user and schedule their check-in mail"`
	Delete   userDeleteCmd   `cmd:"" help:"delete a user and their pending check-in mail"`
}

type userRegisterCmd struct {
	Email        string        `help:"The user's email address"                              required:""`
	CheckInAfter time.Duration `help:"Override the check-in mail delay (defaults to 7 days)"`
}

func (c userRegisterCmd) Run(common sharedFlags) error {
	delay := cmp.Or(c.CheckInAfter, workers.CheckInDelay)

	return runUserApp(common, func(ctx context.Context, d userDeps) error {
		var user users.User
		// the user and their check-in mail are created together or not at all
		err := pgx.BeginFunc(ctx, d.pool, func(tx pgx.Tx) error {
			var err error
			if user, err = d.users.CreateTx(ctx, tx, c.Email); err != nil {
				return err
			}
			_, err = d.checkIn.InsertTx(ctx, tx, workers.CheckInMailArgs{UserID: user.ID}, jobs.In(delay))
			return err
		})
		if err != nil {
			return err
		}

		slog.InfoContext(ctx, "registered user", slog.Int64("id", user.ID), slog.Duration("check_in_after", delay))
		return nil
	})
}

type userDeleteCmd struct {
	ID int64 `help:"The user's ID" required:""`
}

// deletePendingCheckInTx deletes userID's check-in mail if it hasn't started
// yet, returning how many jobs were deleted.
func deletePendingCheckInTx(
	ctx context.Context,
	c *jobs.JobClient[workers.CheckInMailArgs],
	tx pgx.Tx,
	userID int64,
) (int, error) {
	return c.DeletePendingTx(ctx, tx, "(args->>'user_id')::bigint = @user_id", river.NamedArgs{"user_id": userID})
}

func (c userDeleteCmd) Run(common sharedFlags) error {
	return runUserApp(common, func(ctx context.Context, d userDeps) error {
		var deleted int
		err := pgx.BeginFunc(ctx, d.pool, func(tx pgx.Tx) error {
			if err := d.users.DeleteTx(ctx, tx, c.ID); err != nil {
				return err
			}

			if _, err := deletePendingCheckInTx(ctx, d.checkIn, tx, c.ID); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		slog.InfoContext(ctx, "deleted user", slog.Int64("id", c.ID), slog.Int("pending_jobs_deleted", deleted))
		return nil
	})
}

type userDeps struct {
	pool    *pgxpool.Pool
	users   *users.Store
	checkIn *jobs.JobClient[workers.CheckInMailArgs]
}

func runUserApp(common sharedFlags, fn func(context.Context, userDeps) error) error {
	var d userDeps
	app := fx.New(
		workers.Module, jobs.Module, database.Module, users.Module,
		fx.Supply(database.DatabasePoolConfig{DatabaseURL: common.DatabaseURL}),
		fx.Populate(&d.pool, &d.users, &d.checkIn),
	)

	return runApp(app, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return fn(ctx, d)
	})
}
