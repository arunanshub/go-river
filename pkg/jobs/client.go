package jobs

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
)

type JobClient[A river.JobArgs] struct {
	// placeholder, plus it reminds me of PhantomData
	_ A

	client *river.Client[pgx.Tx]

	insertOpts river.InsertOpts
}

func (c *JobClient[A]) Insert(ctx context.Context, args A) (*rivertype.JobInsertResult, error) {
	return c.client.Insert(ctx, args, &c.insertOpts)
}

func (c *JobClient[A]) InsertTx(ctx context.Context, tx pgx.Tx, args A) (*rivertype.JobInsertResult, error) {
	return c.client.InsertTx(ctx, tx, args, &c.insertOpts)
}

// insertOnlyJobClient is not to be used. This only exists  Prefer using [JobClient].
type insertOnlyJobClient struct {
	client *river.Client[pgx.Tx]
}

func NewInsertOnlyClient(pool *pgxpool.Pool) (insertOnlyJobClient, error) {
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		return insertOnlyJobClient{}, err
	}
	return insertOnlyJobClient{client}, nil
}
