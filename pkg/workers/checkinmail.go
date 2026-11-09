package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"go.uber.org/fx"

	"github.com/arunanshub/go-river/pkg/jobs"
	"github.com/arunanshub/go-river/pkg/users"
)

// CheckInDelay is how long after registration the check-in mail goes out.
const CheckInDelay = 7 * 24 * time.Hour

type CheckInMailArgs struct {
	// Only the ID: the worker loads the user, so job rows hold no personal data.
	UserID int64 `json:"user_id"`
}

func (CheckInMailArgs) Kind() string { return "checkin_mail" }

type CheckInMailWorker struct {
	river.WorkerDefaults[CheckInMailArgs]

	users *users.Store
}

func (w CheckInMailWorker) Work(ctx context.Context, job *river.Job[CheckInMailArgs]) error {
	user, err := w.users.Get(ctx, job.Args.UserID)
	if errors.Is(err, users.ErrNotFound) {
		// deleted while this job was running, or after it was fetched; deleting
		// pending jobs can't cover that, so this check is what guarantees it
		return river.JobCancel(fmt.Errorf("user %d no longer exists", job.Args.UserID))
	}
	if err != nil {
		return err
	}

	// stand-in for a real mailer
	slog.InfoContext(ctx, "sent check-in mail", slog.Int64("user_id", user.ID), slog.String("email", user.Email))
	return nil
}

var _ river.Worker[CheckInMailArgs] = (*CheckInMailWorker)(nil)

type NewCheckInMailWorkerParams struct {
	fx.In

	Users *users.Store
}

func NewCheckInMailWorker(params NewCheckInMailWorkerParams) (river.Worker[CheckInMailArgs], error) {
	return CheckInMailWorker{users: params.Users}, nil
}

// ByArgs: one check-in mail per user, however often the insert is retried.
var checkInMailModule = jobs.AsWorkerModule(jobs.SlowQueue(), NewCheckInMailWorker,
	jobs.Unique(river.UniqueOpts{ByArgs: true}),
)
