package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (repo *UserRepository) Create(ctx context.Context, user model.User) error {
	_, err := repo.pool.Exec(ctx,
		`INSERT INTO users (id, name, created_at) VALUES ($1, $2, $3)`,
		user.ID,
		user.Name,
		user.CreatedAt,
	)
	return err
}

func (repo *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	var user model.User
	err := repo.pool.QueryRow(ctx,
		`SELECT id, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, domain.ErrNotFound
	}
	return user, err
}

func (repo *UserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := repo.pool.Query(ctx, `SELECT id, name, created_at FROM users ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
