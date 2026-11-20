package services

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

func (r *PostgresRepository) ListServices(ctx context.Context) ([]Service, error) {
	query := `SELECT id, name, description, duration_minutes, price, is_active
	FROM services
	WHERE is_active = true
	ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]Service, 0)

	for rows.Next() {
		var service Service

		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Description,
			&service.DurationMinutes,
			&service.Price,
			&service.IsActive,
		)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *PostgresRepository) GetServiceByID(ctx context.Context, id int) (Service, error) {
	query := `SELECT id, name, description, duration_minutes, price, is_active
	FROM services
	WHERE id = $1 AND is_active = true`

	row := r.pool.QueryRow(ctx, query, id)

	var service Service

	err := row.Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.DurationMinutes,
		&service.Price,
		&service.IsActive,
	)
	if err != nil {
		return Service{}, err
	}

	return service, nil
}
