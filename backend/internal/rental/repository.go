package rental

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

type rentalUnit struct {
	ID           int
	EquipmentID  int
	Name         string
	Size         string
	PricePerHour int
}

var (
	ErrEquipmentUnavailable = errors.New("equipment unavailable")
	ErrEquipmentNotFound    = errors.New("equipment not found")
	ErrEquipmentSetNotFound = errors.New("equipment set not found")
	ErrEquipmentSetInactive = errors.New("equipment set is inactive")
	ErrInvalidSetSizes      = errors.New("invalid equipment set sizes")
	ErrInvalidRentalInput   = errors.New("invalid rental input")
	ErrRentalCannotCancel   = errors.New("rental cannot be cancelled")
)

// Выбираем свободные экземпляры и блокируем их до конца транзакции.
func (r *PostgresRepository) getAvailableUnits(ctx context.Context, tx pgx.Tx, equipmentID, categoryID *int,
	size string, quantity int, startsAt, endsAt time.Time, excludedIDs []int) ([]rentalUnit, error) {
	query := `
		SELECT
			eu.id,
			eu.equipment_id,
			e.name,
			eu.size,
			e.price_per_hour
		FROM equipment_units eu
		JOIN equipment e ON e.id = eu.equipment_id
		WHERE eu.status = 'available'
		  AND ($1::int IS NULL OR eu.equipment_id = $1)
		  AND ($2::int IS NULL OR e.category_id = $2)
		  AND eu.size = $3
		  AND NOT (eu.id = ANY($4::int[]))
		  AND NOT EXISTS (
		      SELECT 1
		      FROM equipment_booking_items ebi
		      JOIN equipment_bookings eb ON eb.id = ebi.booking_id
		      WHERE ebi.equipment_unit_id = eu.id
		        AND eb.status IN ('pending', 'confirmed', 'issued')
		        AND eb.starts_at < $6
		        AND eb.ends_at > $5
		  )
		ORDER BY eu.id
		LIMIT $7
		FOR UPDATE OF eu
	`

	rows, err := tx.Query(ctx, query,
		equipmentID, categoryID, size, excludedIDs,
		startsAt, endsAt, quantity,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	units := make([]rentalUnit, 0, quantity)

	for rows.Next() {
		var unit rentalUnit

		err := rows.Scan(&unit.ID, &unit.EquipmentID, &unit.Name, &unit.Size, &unit.PricePerHour)
		if err != nil {
			return nil, err
		}

		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(units) != quantity {
		return nil, ErrEquipmentUnavailable
	}

	return units, nil
}

// Проверяем доступность моделей по размерам на выбранный период.
func (r *PostgresRepository) GetEquipmentAvailability(ctx context.Context, startsAt, endsAt time.Time, equipmentID *int) ([]EquipmentAvailability, error) {
	if equipmentID != nil {
		var exists bool

		err := r.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM equipment
				WHERE id = $1
			)`, *equipmentID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrEquipmentNotFound
		}
	}

	query := `
		SELECT
			eu.equipment_id,
			eu.size,
			COUNT(*) FILTER (
				WHERE eu.status = 'available'
				  AND NOT EXISTS (
				      SELECT 1
				      FROM equipment_booking_items ebi
				      JOIN equipment_bookings eb ON eb.id = ebi.booking_id
				      WHERE ebi.equipment_unit_id = eu.id
				        AND eb.status IN ('pending', 'confirmed', 'issued')
				        AND eb.starts_at < $2
				        AND eb.ends_at > $1
				  )
			) AS available_quantity
		FROM equipment_units eu
		WHERE ($3::int IS NULL OR eu.equipment_id = $3)
		GROUP BY eu.equipment_id, eu.size
		ORDER BY eu.equipment_id, eu.size
	`

	rows, err := r.pool.Query(ctx, query, startsAt, endsAt, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]EquipmentAvailability, 0)

	for rows.Next() {
		var item EquipmentAvailability

		err := rows.Scan(&item.EquipmentID, &item.Size, &item.AvailableQuantity)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Создаём аренду в одной транзакции.
func (r *PostgresRepository) CreateEquipmentBooking(ctx context.Context, userID int, params CreateEquipmentBookingParams) (EquipmentRental, error) {
	hours := int(params.EndsAt.Sub(params.StartsAt).Hours())
	if hours <= 0 || params.EndsAt.Sub(params.StartsAt) != time.Duration(hours)*time.Hour {
		return EquipmentRental{}, ErrInvalidRentalInput
	}
	if len(params.Items) == 0 && len(params.Sets) == 0 {
		return EquipmentRental{}, ErrInvalidRentalInput
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return EquipmentRental{}, err
	}
	defer tx.Rollback(ctx)

	// Одна транзакция распределения оборудования за раз
	// Все операции выделения экземпляров должны соблюдать этот протокол
	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(82416310)`)
	if err != nil {
		return EquipmentRental{}, err
	}

	var bookingID int

	err = tx.QueryRow(ctx, `
		INSERT INTO equipment_bookings (
			user_id, starts_at, ends_at, status, total_price
		)
		VALUES ($1, $2, $3, 'confirmed', 0)
		RETURNING id
	`, userID, params.StartsAt, params.EndsAt).Scan(&bookingID)
	if err != nil {
		return EquipmentRental{}, err
	}

	totalPrice := 0
	allocatedIDs := make([]int, 0)

	// Отдельные позиции оборудования.
	for _, item := range params.Items {
		if item.EquipmentID <= 0 || item.Quantity <= 0 || item.Size == "" {
			return EquipmentRental{}, ErrInvalidRentalInput
		}

		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM equipment WHERE id = $1
			)
		`, item.EquipmentID).Scan(&exists)
		if err != nil {
			return EquipmentRental{}, err
		}
		if !exists {
			return EquipmentRental{}, ErrEquipmentNotFound
		}

		units, err := r.getAvailableUnits(ctx, tx, &item.EquipmentID, nil, item.Size, item.Quantity,
			params.StartsAt, params.EndsAt, allocatedIDs)
		if err != nil {
			return EquipmentRental{}, err
		}

		for _, unit := range units {
			itemPrice := unit.PricePerHour * hours

			_, err := tx.Exec(ctx, `
				INSERT INTO equipment_booking_items (
					booking_id, equipment_unit_id, item_total_price
				)
				VALUES ($1, $2, $3)`, bookingID, unit.ID, itemPrice)
			if err != nil {
				return EquipmentRental{}, err
			}

			totalPrice += itemPrice
			allocatedIDs = append(allocatedIDs, unit.ID)
		}
	}

	// Комплекты оборудования.
	for _, item := range params.Sets {
		if item.EquipmentSetID <= 0 || item.Quantity <= 0 {
			return EquipmentRental{}, ErrInvalidRentalInput
		}

		var set EquipmentSetInfo

		err := tx.QueryRow(ctx, `
			SELECT id, price_per_hour, is_active
			FROM equipment_sets
			WHERE id = $1`, item.EquipmentSetID).Scan(&set.ID, &set.PricePerHour, &set.IsActive)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return EquipmentRental{}, ErrEquipmentSetNotFound
			}
			return EquipmentRental{}, err
		}

		if !set.IsActive {
			return EquipmentRental{}, ErrEquipmentSetInactive
		}

		rows, err := tx.Query(ctx, `
			SELECT equipment_category_id, quantity
			FROM equipment_set_items
			WHERE equipment_set_id = $1
			ORDER BY equipment_category_id
		`, item.EquipmentSetID)
		if err != nil {
			return EquipmentRental{}, err
		}

		setItems := make([]EquipmentSetItem, 0)

		for rows.Next() {
			var setItem EquipmentSetItem

			err := rows.Scan(&setItem.CategoryID, &setItem.Quantity)
			if err != nil {
				rows.Close()
				return EquipmentRental{}, err
			}
			setItems = append(setItems, setItem)
		}

		err = rows.Err()
		rows.Close()
		if err != nil {
			return EquipmentRental{}, err
		}

		if len(setItems) == 0 {
			return EquipmentRental{}, ErrInvalidSetSizes
		}

		sizes := make(map[int]string)

		for _, selected := range item.Sizes {
			if selected.CategoryID <= 0 || selected.Size == "" {
				return EquipmentRental{}, ErrInvalidSetSizes
			}
			if _, exists := sizes[selected.CategoryID]; exists {
				return EquipmentRental{}, ErrInvalidSetSizes
			}
			sizes[selected.CategoryID] = selected.Size
		}

		// Для каждой категории комплекта должен быть указан ровно один размер.
		if len(sizes) != len(setItems) {
			return EquipmentRental{}, ErrInvalidSetSizes
		}

		for _, setItem := range setItems {
			if sizes[setItem.CategoryID] == "" {
				return EquipmentRental{}, ErrInvalidSetSizes
			}
		}

		// quantity = число одинаковых комплектов с выбранными размерами.
		for n := 0; n < item.Quantity; n++ {
			setPrice := set.PricePerHour * hours
			var bookedSetID int

			err := tx.QueryRow(ctx, `
				INSERT INTO booked_equipment_sets (
					booking_id, equipment_set_id, total_price
				)
				VALUES ($1, $2, $3)
				RETURNING id`, bookingID, item.EquipmentSetID, setPrice).Scan(&bookedSetID)
			if err != nil {
				return EquipmentRental{}, err
			}

			for _, setItem := range setItems {
				units, err := r.getAvailableUnits(
					ctx, tx,
					nil, &setItem.CategoryID,
					sizes[setItem.CategoryID],
					setItem.Quantity,
					params.StartsAt, params.EndsAt,
					allocatedIDs,
				)
				if err != nil {
					return EquipmentRental{}, err
				}

				for _, unit := range units {
					_, err := tx.Exec(ctx, `
						INSERT INTO equipment_booking_items (
							booking_id,
							equipment_unit_id,
							booked_set_id
						)
						VALUES ($1, $2, $3)`, bookingID, unit.ID, bookedSetID)
					if err != nil {
						return EquipmentRental{}, err
					}

					allocatedIDs = append(allocatedIDs, unit.ID)
				}
			}

			totalPrice += setPrice
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE equipment_bookings
		SET total_price = $2
		WHERE id = $1`, bookingID, totalPrice)
	if err != nil {
		return EquipmentRental{}, err
	}

	result, err := loadEquipmentRental(ctx, tx, bookingID)
	if err != nil {
		return EquipmentRental{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return EquipmentRental{}, fmt.Errorf("commit equipment booking: %w", err)
	}

	return result, nil
}

type rentalQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func loadEquipmentRental(ctx context.Context, db rentalQuerier, bookingID int) (EquipmentRental, error) {
	query := `
		SELECT
			id, user_id, starts_at, ends_at,
			status, total_price, created_at
		FROM equipment_bookings
		WHERE id = $1`

	var rental EquipmentRental

	err := db.QueryRow(ctx, query, bookingID).Scan(&rental.ID, &rental.UserID, &rental.StartsAt,
		&rental.EndsAt, &rental.Status, &rental.TotalPrice, &rental.CreatedAt)
	if err != nil {
		return EquipmentRental{}, err
	}

	rental.Items = make([]RentalItem, 0)
	rental.Sets = make([]BookedRentalSet, 0)

	query = `
		SELECT
			ebi.equipment_unit_id,
			e.id,
			e.name,
			eu.size,
			ebi.booked_set_id,
			ebi.item_total_price
		FROM equipment_booking_items ebi
		JOIN equipment_units eu ON eu.id = ebi.equipment_unit_id
		JOIN equipment e ON e.id = eu.equipment_id
		WHERE ebi.booking_id = $1
		ORDER BY ebi.id`

	rows, err := db.Query(ctx, query, bookingID)
	if err != nil {
		return EquipmentRental{}, err
	}

	for rows.Next() {
		var item RentalItem

		err := rows.Scan(&item.EquipmentUnitID, &item.EquipmentID, &item.Name,
			&item.Size, &item.BookedSetID, &item.ItemTotalPrice,
		)
		if err != nil {
			rows.Close()
			return EquipmentRental{}, err
		}

		rental.Items = append(rental.Items, item)
	}

	err = rows.Err()
	rows.Close()
	if err != nil {
		return EquipmentRental{}, err
	}

	query = `
		SELECT
			bes.id,
			bes.equipment_set_id,
			es.name,
			bes.total_price
		FROM booked_equipment_sets bes
		JOIN equipment_sets es ON es.id = bes.equipment_set_id
		WHERE bes.booking_id = $1
		ORDER BY bes.id`

	rows, err = db.Query(ctx, query, bookingID)
	if err != nil {
		return EquipmentRental{}, err
	}

	for rows.Next() {
		var set BookedRentalSet

		err := rows.Scan(&set.ID, &set.EquipmentSetID, &set.Name, &set.TotalPrice)
		if err != nil {
			rows.Close()
			return EquipmentRental{}, err
		}

		rental.Sets = append(rental.Sets, set)
	}

	err = rows.Err()
	rows.Close()
	if err != nil {
		return EquipmentRental{}, err
	}

	return rental, nil
}

func (r *PostgresRepository) GetEquipmentBookingByID(ctx context.Context, bookingID int) (EquipmentRental, error) {
	return loadEquipmentRental(ctx, r.pool, bookingID)
}

func (r *PostgresRepository) GetMyEquipmentBookings(ctx context.Context, userID int, status *string) ([]EquipmentRental, error) {
	query := `
		SELECT id
		FROM equipment_bookings
		WHERE user_id = $1
		  AND ($2::text IS NULL OR status = $2)
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.pool.Query(ctx, query, userID, status)
	if err != nil {
		return nil, err
	}

	bookingIDs := make([]int, 0)

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}

		bookingIDs = append(bookingIDs, id)
	}

	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}

	result := make([]EquipmentRental, 0, len(bookingIDs))

	for _, id := range bookingIDs {
		rental, err := r.GetEquipmentBookingByID(ctx, id)
		if err != nil {
			return nil, err
		}

		result = append(result, rental)
	}

	return result, nil
}

func (r *PostgresRepository) CancelEquipmentBooking(ctx context.Context, bookingID, userID int) (EquipmentRental, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return EquipmentRental{}, err
	}
	defer tx.Rollback(ctx)

	var id int

	err = tx.QueryRow(ctx, `
		UPDATE equipment_bookings
		SET status = 'cancelled'
		WHERE id = $1
		  AND user_id = $2
		  AND status IN ('pending', 'confirmed')
		  AND starts_at > NOW()
		RETURNING id`, bookingID, userID).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EquipmentRental{}, ErrRentalCannotCancel
		}
		return EquipmentRental{}, err
	}

	rental, err := loadEquipmentRental(ctx, tx, id)
	if err != nil {
		return EquipmentRental{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return EquipmentRental{}, err
	}

	return rental, nil
}
