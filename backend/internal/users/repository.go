package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, user User) (int, error) {
	query := `INSERT INTO users (name, email, phone, password_hash, role) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int

	row := r.pool.QueryRow(ctx, query, user.Name, user.Email, user.Phone, user.PasswordHash, user.Role)

	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	query := `SELECT id, password_hash FROM users WHERE email = $1`

	var id int
	var passwordHash string

	row := r.pool.QueryRow(ctx, query, email)

	err := row.Scan(&id, &passwordHash)
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, PasswordHash: passwordHash}, nil

}
