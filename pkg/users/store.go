package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("users", fx.Provide(NewStore))

var ErrNotFound = errors.New("user not found")

type User struct {
	ID    int64
	Email string
}

type Store struct {
	pool *pgxpool.Pool
}

// stands in for a migration until the project has migration tooling.
const schema = `CREATE TABLE IF NOT EXISTS users (
	id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	email text NOT NULL
)`

func NewStore(lc fx.Lifecycle, pool *pgxpool.Pool) *Store {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			_, err := pool.Exec(ctx, schema)
			return err
		},
	})
	return &Store{pool: pool}
}

func (s *Store) CreateTx(ctx context.Context, tx pgx.Tx, email string) (User, error) {
	user := User{Email: email}
	err := tx.QueryRow(ctx, `INSERT INTO users (email) VALUES ($1) RETURNING id`, email).Scan(&user.ID)
	return user, err
}

func (s *Store) DeleteTx(ctx context.Context, tx pgx.Tx, id int64) error {
	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id int64) (User, error) {
	user := User{ID: id}
	err := s.pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, id).Scan(&user.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}
