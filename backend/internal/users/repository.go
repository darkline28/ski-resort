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
	query := `SELECT id, name, email, phone, role, password_hash
		FROM users
		WHERE email = $1
	`

	var user User

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Role,
		&user.PasswordHash,
	)
	if err != nil {
		return User{}, err
	}

	return user, nil

}

func (r *PostgresRepository) GetByID(ctx context.Context, id int) (User, error) {
	query := `SELECT id, name, email, phone, role FROM users WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
