package jobs

import (
	"context"

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

func (c *JobClient[A]) Insert(ctx context.Context, args A) (*rivertype.JobInsertResult, error) {
	return c.client.Insert(ctx, args, &c.insertOpts)
}

func (c *JobClient[A]) InsertTx(ctx context.Context, tx pgx.Tx, args A) (*rivertype.JobInsertResult, error) {
	return c.client.InsertTx(ctx, tx, args, &c.insertOpts)
}

func (c *JobClient[A]) InsertMany(ctx context.Context, args []A) ([]*rivertype.JobInsertResult, error) {
	return c.client.InsertMany(ctx, c.manyParams(args))
}

func (c *JobClient[A]) InsertManyTx(ctx context.Context, tx pgx.Tx, args []A) ([]*rivertype.JobInsertResult, error) {
	return c.client.InsertManyTx(ctx, tx, c.manyParams(args))
}

func (c *JobClient[A]) manyParams(args []A) []river.InsertManyParams {
	params := make([]river.InsertManyParams, len(args))
	for i, a := range args {
		params[i] = river.InsertManyParams{Args: a, InsertOpts: &c.insertOpts}
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
