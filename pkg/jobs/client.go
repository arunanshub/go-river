package jobs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
)

// JobClient inserts jobs of a single args type. It can only be obtained through
// [AsWorkerModule], so a JobClient[A] existing implies a worker for A is defined.
type JobClient[A river.JobArgs] struct {
	client *river.Client[pgx.Tx]

	insertOpts river.InsertOpts
}

// InsertOption tunes a single insert. Per-kind policy and the queue come from
// [AsWorkerModule]; an InsertOption can't change the queue.
type InsertOption struct{ apply func(*river.InsertOpts) }

// At makes the job available to work at t.
func At(t time.Time) InsertOption {
	return InsertOption{func(o *river.InsertOpts) { o.ScheduledAt = t }}
}

// In makes the job available to work after d, counted from when In is called.
func In(d time.Duration) InsertOption {
	return At(time.Now().Add(d))
}

func (c *JobClient[A]) Insert(ctx context.Context, args A, opts ...InsertOption) (*rivertype.JobInsertResult, error) {
	return c.client.Insert(ctx, args, c.opts(opts))
}

func (c *JobClient[A]) InsertTx(
	ctx context.Context,
	tx pgx.Tx,
	args A,
	opts ...InsertOption,
) (*rivertype.JobInsertResult, error) {
	return c.client.InsertTx(ctx, tx, args, c.opts(opts))
}

func (c *JobClient[A]) InsertMany(
	ctx context.Context,
	args []A,
	opts ...InsertOption,
) ([]*rivertype.JobInsertResult, error) {
	return c.client.InsertMany(ctx, c.manyParams(args, opts))
}

func (c *JobClient[A]) InsertManyTx(
	ctx context.Context,
	tx pgx.Tx,
	args []A,
	opts ...InsertOption,
) ([]*rivertype.JobInsertResult, error) {
	return c.client.InsertManyTx(ctx, tx, c.manyParams(args, opts))
}

// DeletePendingTx deletes this kind's jobs that haven't started yet and match
// where, atomically with tx. Running jobs are left alone, so workers must still
// handle their subject having gone away. where is raw SQL against river_job:
// pass values through args as @named parameters, never by formatting them in.
func (c *JobClient[A]) DeletePendingTx(
	ctx context.Context,
	tx pgx.Tx,
	where string,
	args river.NamedArgs,
) (int, error) {
	var zero A
	res, err := c.client.JobDeleteManyTx(ctx, tx, river.NewJobDeleteManyParams().
		Kinds(zero.Kind()).
		States(rivertype.JobStateAvailable, rivertype.JobStateRetryable, rivertype.JobStateScheduled).
		Where(where, args))
	if err != nil {
		return 0, err
	}
	return len(res.Jobs), nil
}

// opts layers per-call options over a copy of the per-kind template, then
// reasserts the registered queue.
func (c *JobClient[A]) opts(options []InsertOption) *river.InsertOpts {
	opts := c.insertOpts
	for _, opt := range options {
		opt.apply(&opts)
	}
	opts.Queue = c.insertOpts.Queue
	return &opts
}

func (c *JobClient[A]) manyParams(args []A, options []InsertOption) []river.InsertManyParams {
	opts := c.opts(options)
	params := make([]river.InsertManyParams, len(args))
	for i, a := range args {
		params[i] = river.InsertManyParams{Args: a, InsertOpts: opts}
	}
	return params
}

// insertOnlyJobClient is the untyped river client backing every [JobClient].
// It is unexported so nothing outside this package can insert arbitrary args.
type insertOnlyJobClient struct {
	client *river.Client[pgx.Tx]
}

func newInsertOnlyClient(pool *pgxpool.Pool) (insertOnlyJobClient, error) {
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		return insertOnlyJobClient{}, err
	}
	return insertOnlyJobClient{client}, nil
}
