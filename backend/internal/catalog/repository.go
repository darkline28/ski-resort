package catalog

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListEquipment(ctx context.Context, category *string) ([]Equipment, error) {
	query := `SELECT e.id, e.name, e.description, e.image_url, e.price_per_hour, c.name AS category, COUNT(eu.id) AS quantity FROM equipment e
				LEFT JOIN equipment_categories c
				ON e.category_id = c.id
				LEFT JOIN equipment_units eu
				ON e.id = eu.equipment_id
				WHERE $1::text IS NULL OR c.code = $1 
				GROUP BY e.id, c.name
				ORDER BY e.id`

	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipments []Equipment

	for rows.Next() {
		var equipment Equipment
		err := rows.Scan(&equipment.ID, &equipment.Name, &equipment.Description, &equipment.ImageURL, &equipment.PricePerHour, &equipment.Category, &equipment.Quantity)
		if err != nil {
			return nil, err
		}
		equipments = append(equipments, equipment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return equipments, nil
}

var ErrEquipmentNotFound = errors.New("id not found")

func (r *PostgresRepository) GetEquipmentByID(ctx context.Context, id int) (Equipment, error) {
	query := `SELECT e.id, c.name AS category, e.name, e.description, e.image_url, e.price_per_hour, COUNT(eu.id) AS quantity
			FROM equipment e
			LEFT JOIN equipment_categories c
			ON e.category_id = c.id
			LEFT JOIN equipment_units eu
			ON e.id = eu.equipment_id
			WHERE e.id = $1
			GROUP BY e.id, c.name`

	row := r.pool.QueryRow(ctx, query, id)

	var equipment Equipment
	err := row.Scan(&equipment.ID, &equipment.Category, &equipment.Name, &equipment.Description, &equipment.ImageURL, &equipment.PricePerHour, &equipment.Quantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Equipment{}, ErrEquipmentNotFound
		}
		return Equipment{}, err
	}
	return equipment, nil
}

func (r *PostgresRepository) GetEquipmentCategories(ctx context.Context) ([]Category, error) {
	query := `SELECT id, name FROM equipment_categories`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)

	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

var ErrEquipmentSetNotFound = errors.New("set not found")

func (r *PostgresRepository) GetEquipmentSetByID(ctx context.Context, id int) (EquipmentSet, error) {
	querySet := `SELECT id, name, description, sport_type, age_group, price_per_hour, is_active
		FROM equipment_sets
		WHERE id = $1`

	querySetItems := `SELECT es.equipment_category_id, c.name, es.quantity
					FROM equipment_set_items es
					INNER JOIN equipment_categories c
					ON es.equipment_category_id = c.id
					WHERE es.equipment_set_id = $1`

	rowSet := r.pool.QueryRow(ctx, querySet, id)

	var set EquipmentSet
	err := rowSet.Scan(&set.ID, &set.Name, &set.Description, &set.SportType, &set.AgeGroup, &set.PricePerHour, &set.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EquipmentSet{}, ErrEquipmentSetNotFound
		}
		return EquipmentSet{}, err
	}

	rowsSetItems, err := r.pool.Query(ctx, querySetItems, id)
	if err != nil {
		return EquipmentSet{}, err
	}
	defer rowsSetItems.Close()

	set.Items = make([]EquipmentSetItem, 0)

	for rowsSetItems.Next() {
		var items EquipmentSetItem
		err := rowsSetItems.Scan(&items.CategoryID, &items.Category, &items.Quantity)
		if err != nil {
			return EquipmentSet{}, err
		}
		set.Items = append(set.Items, items)
	}

	if err := rowsSetItems.Err(); err != nil {
		return EquipmentSet{}, err
	}

	return set, nil

}

func (r *PostgresRepository) GetEquipmentSets(ctx context.Context) ([]EquipmentSet, error) {
	query := `SELECT id, name, description, sport_type, age_group, price_per_hour, is_active FROM equipment_sets ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets = make([]EquipmentSet, 0)
	for rows.Next() {
		var set EquipmentSet
		err := rows.Scan(&set.ID, &set.Name, &set.Description, &set.SportType, &set.AgeGroup, &set.PricePerHour, &set.IsActive)
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sets, nil
}
