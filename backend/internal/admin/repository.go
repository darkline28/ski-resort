package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	ErrInvalidReference  = errors.New("invalid reference")
	ErrInvalidTransition = errors.New("invalid status transition")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func dbError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23P01":
			return fmt.Errorf("%w: %v", ErrConflict, err)
		case "23503", "23514", "23502", "22001", "22P02":
			return fmt.Errorf("%w: %v", ErrInvalidReference, err)
		}
	}
	return err
}

// SERVICES

func (r *PostgresRepository) GetServices(ctx context.Context) ([]ServiceModel, error) {
	query := `SELECT id, name, description, duration_minutes, price, is_active FROM services ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get services: %w", err)
	}
	defer rows.Close()

	services := make([]ServiceModel, 0)
	for rows.Next() {
		var service ServiceModel
		if err := rows.Scan(&service.ID, &service.Name, &service.Description, &service.DurationMinutes, &service.Price, &service.IsActive); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read services: %w", err)
	}

	return services, nil
}

func (r *PostgresRepository) GetServiceByID(ctx context.Context, id int) (*ServiceModel, error) {
	query := `SELECT id, name, description, duration_minutes, price, is_active FROM services WHERE id = $1`

	var service ServiceModel
	err := r.pool.QueryRow(ctx, query, id).Scan(&service.ID, &service.Name, &service.Description, &service.DurationMinutes, &service.Price, &service.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get service: %w", err)
	}

	return &service, nil
}

func (r *PostgresRepository) CreateService(ctx context.Context, params CreateServiceParams) (*ServiceModel, error) {
	query := `
		INSERT INTO services (name, description, duration_minutes, price, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, duration_minutes, price, is_active
	`

	var service ServiceModel
	err := r.pool.QueryRow(ctx, query, params.Name, params.Description, params.DurationMinutes, params.Price, params.IsActive).Scan(&service.ID, &service.Name, &service.Description, &service.DurationMinutes, &service.Price, &service.IsActive)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", dbError(err))
	}

	return &service, nil
}

func (r *PostgresRepository) UpdateService(ctx context.Context, id int, params UpdateServiceParams) (*ServiceModel, error) {
	query := `
		UPDATE services
		SET name = COALESCE($2, name),
			description = COALESCE($3, description),
			duration_minutes = COALESCE($4, duration_minutes),
			price = COALESCE($5, price),
			is_active = COALESCE($6, is_active)
		WHERE id = $1
		RETURNING id, name, description, duration_minutes, price, is_active
	`

	var service ServiceModel
	err := r.pool.QueryRow(ctx, query, id, params.Name, params.Description, params.DurationMinutes, params.Price, params.IsActive).Scan(&service.ID, &service.Name, &service.Description, &service.DurationMinutes, &service.Price, &service.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update service: %w", dbError(err))
	}

	return &service, nil
}

func (r *PostgresRepository) DeleteService(ctx context.Context, id int) error {
	query := `UPDATE services SET is_active = false WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deactivate service: %w", dbError(err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// INSTRUCTORS

func (r *PostgresRepository) GetInstructorByID(ctx context.Context, id int) (*Instructor, error) {
	query := `
		SELECT i.id, u.name, i.specialization, i.description, i.experience_years, i.image_url, i.is_active
		FROM instructors i
		JOIN users u ON u.id = i.user_id
		WHERE i.id = $1
	`

	var instructor Instructor
	err := r.pool.QueryRow(ctx, query, id).Scan(&instructor.ID, &instructor.Name, &instructor.Specialization, &instructor.Description, &instructor.ExperienceYears, &instructor.ImageURL, &instructor.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get instructor: %w", err)
	}

	return &instructor, nil
}

func (r *PostgresRepository) GetInstructors(ctx context.Context) ([]Instructor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id
		FROM instructors
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("get instructor ids: %w", err)
	}

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan instructor id: %w", err)
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("read instructor ids: %w", err)
	}

	instructors := make([]Instructor, 0, len(ids))
	for _, id := range ids {
		instructor, err := r.GetInstructorByID(ctx, id)
		if err != nil {
			return nil, err
		}
		instructors = append(instructors, *instructor)
	}

	return instructors, nil
}

func (r *PostgresRepository) CreateInstructor(ctx context.Context, params CreateInstructorParams) (*Instructor, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET role = 'instructor'
		WHERE id = $1 AND role IN ('client', 'instructor')
		RETURNING id
	`, params.UserID).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidReference
	}
	if err != nil {
		return nil, fmt.Errorf("assign instructor role: %w", dbError(err))
	}

	var instructorID int
	err = tx.QueryRow(ctx, `
		INSERT INTO instructors (
			user_id, specialization, description,
			experience_years, image_url, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userID, params.Specialization, params.Description,
		params.ExperienceYears, params.ImageURL, params.IsActive,
	).Scan(&instructorID)

	if err != nil {
		return nil, fmt.Errorf("create instructor: %w", dbError(err))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit instructor creation: %w", err)
	}

	return r.GetInstructorByID(ctx, instructorID)
}

func (r *PostgresRepository) UpdateInstructor(ctx context.Context, id int, params UpdateInstructorParams) (*Instructor, error) {
	query := `
		UPDATE instructors
		SET specialization = COALESCE($2, specialization),
			description = COALESCE($3, description),
			experience_years = COALESCE($4, experience_years),
			image_url = COALESCE($5, image_url),
			is_active = COALESCE($6, is_active)
		WHERE id = $1
		RETURNING id
	`

	var updatedID int
	err := r.pool.QueryRow(ctx, query, id, params.Specialization, params.Description, params.ExperienceYears, params.ImageURL, params.IsActive).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update instructor: %w", dbError(err))
	}

	return r.GetInstructorByID(ctx, updatedID)
}

func (r *PostgresRepository) DeleteInstructor(ctx context.Context, id int) error {
	query := `UPDATE instructors SET is_active = false WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deactivate instructor: %w", dbError(err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// INSTRUCTOR SERVICES

func (r *PostgresRepository) GetInstructorServices(ctx context.Context, instructorID int) ([]ServiceModel, error) {
	if _, err := r.GetInstructorByID(ctx, instructorID); err != nil {
		return nil, err
	}

	query := `
		SELECT s.id, s.name, s.description, s.duration_minutes, s.price, s.is_active
		FROM instructor_services ins
		JOIN services s ON s.id = ins.service_id
		WHERE ins.instructor_id = $1
		ORDER BY s.id
	`

	rows, err := r.pool.Query(ctx, query, instructorID)
	if err != nil {
		return nil, fmt.Errorf("get instructor services: %w", err)
	}
	defer rows.Close()

	services := make([]ServiceModel, 0)
	for rows.Next() {
		var service ServiceModel
		if err := rows.Scan(&service.ID, &service.Name, &service.Description, &service.DurationMinutes, &service.Price, &service.IsActive); err != nil {
			return nil, fmt.Errorf("scan instructor service: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read instructor services: %w", err)
	}

	return services, nil
}

func (r *PostgresRepository) AssignInstructorService(ctx context.Context, instructorID, serviceID int) error {
	query := `
		INSERT INTO instructor_services (instructor_id, service_id)
		SELECT i.id, s.id
		FROM instructors i CROSS JOIN services s
		WHERE i.id = $1 AND s.id = $2
		RETURNING instructor_id
	`

	var id int
	err := r.pool.QueryRow(ctx, query, instructorID, serviceID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("assign instructor service: %w", dbError(err))
	}

	return nil
}

func (r *PostgresRepository) RemoveInstructorService(ctx context.Context, instructorID, serviceID int) error {
	query := `DELETE FROM instructor_services WHERE instructor_id = $1 AND service_id = $2`

	tag, err := r.pool.Exec(ctx, query, instructorID, serviceID)
	if err != nil {
		return fmt.Errorf("remove instructor service: %w", dbError(err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// INSTRUCTOR SCHEDULES

func (r *PostgresRepository) GetInstructorSchedules(ctx context.Context, instructorID int, filter ScheduleFilter) ([]InstructorSchedule, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM instructors WHERE id = $1)", instructorID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check instructor: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	query := `
		SELECT id, starts_at, ends_at
		FROM instructor_schedules
		WHERE instructor_id = $1
		  AND ($2::date IS NULL OR (ends_at AT TIME ZONE 'Europe/Moscow')::date >= $2::date)
		  AND ($3::date IS NULL OR (starts_at AT TIME ZONE 'Europe/Moscow')::date <= $3::date)
		ORDER BY starts_at, id
	`

	rows, err := r.pool.Query(ctx, query, instructorID, filter.DateFrom, filter.DateTo)
	if err != nil {
		return nil, fmt.Errorf("get instructor schedules: %w", err)
	}
	defer rows.Close()

	schedules := make([]InstructorSchedule, 0)
	for rows.Next() {
		var schedule InstructorSchedule
		err := rows.Scan(&schedule.ID, &schedule.StartsAt, &schedule.EndsAt)
		if err != nil {
			return nil, fmt.Errorf("scan instructor schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read instructor schedules: %w", err)
	}

	return schedules, nil
}

func (r *PostgresRepository) CreateInstructorSchedule(ctx context.Context, instructorID int, params CreateScheduleParams) (*InstructorSchedule, error) {
	query := `
		INSERT INTO instructor_schedules (instructor_id, starts_at, ends_at)
		VALUES ($1, $2, $3)
		RETURNING id, starts_at, ends_at
	`

	var schedule InstructorSchedule
	err := r.pool.QueryRow(ctx, query, instructorID, params.StartsAt, params.EndsAt).Scan(&schedule.ID, &schedule.StartsAt, &schedule.EndsAt)
	if err != nil {
		return nil, fmt.Errorf("create instructor schedule: %w", dbError(err))
	}

	return &schedule, nil
}

func (r *PostgresRepository) UpdateInstructorSchedule(ctx context.Context, instructorID, scheduleID int, params UpdateScheduleParams) (*InstructorSchedule, error) {
	query := `
		UPDATE instructor_schedules s
		SET starts_at = COALESCE($3, s.starts_at),
			ends_at = COALESCE($4, s.ends_at)
		WHERE s.id = $1
		  AND s.instructor_id = $2
		  AND NOT EXISTS (
			  SELECT 1
			  FROM lesson_bookings lb
			  WHERE lb.instructor_id = s.instructor_id
			    AND lb.status = 'confirmed'
			    AND lb.starts_at >= s.starts_at
			    AND lb.ends_at <= s.ends_at
			    AND (
			        lb.starts_at < COALESCE($3, s.starts_at)
			        OR lb.ends_at > COALESCE($4, s.ends_at)
			    )
		  )
		RETURNING s.id, s.starts_at, s.ends_at
	`

	var schedule InstructorSchedule
	err := r.pool.QueryRow(ctx, query, scheduleID, instructorID, params.StartsAt, params.EndsAt).Scan(&schedule.ID, &schedule.StartsAt, &schedule.EndsAt)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		checkErr := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM instructor_schedules WHERE id = $1 AND instructor_id = $2)", scheduleID, instructorID).Scan(&exists)
		if checkErr != nil {
			return nil, fmt.Errorf("check instructor schedule: %w", checkErr)
		}
		if exists {
			return nil, ErrConflict
		}
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update instructor schedule: %w", dbError(err))
	}

	return &schedule, nil
}

func (r *PostgresRepository) DeleteInstructorSchedule(ctx context.Context, instructorID, scheduleID int) error {
	query := `
		DELETE FROM instructor_schedules s
		WHERE s.id = $1
		  AND s.instructor_id = $2
		  AND NOT EXISTS (
			  SELECT 1
			  FROM lesson_bookings lb
			  WHERE lb.instructor_id = s.instructor_id
			    AND lb.status = 'confirmed'
			    AND lb.starts_at >= s.starts_at
			    AND lb.ends_at <= s.ends_at
		  )
	`

	tag, err := r.pool.Exec(ctx, query, scheduleID, instructorID)
	if err != nil {
		return fmt.Errorf("delete instructor schedule: %w", dbError(err))
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	var exists bool
	err = r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM instructor_schedules WHERE id = $1 AND instructor_id = $2)", scheduleID, instructorID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check instructor schedule: %w", err)
	}
	if exists {
		return ErrConflict
	}

	return ErrNotFound
}

// LESSON BOOKINGS

const lessonBookingSelect = `
	SELECT lb.id, lb.user_id, client.name, lb.instructor_id,
	       instructor_user.name, lb.service_id, s.name,
	       lb.starts_at, lb.ends_at, lb.status, lb.total_price, lb.created_at
	FROM lesson_bookings lb
	JOIN users client ON client.id = lb.user_id
	JOIN instructors i ON i.id = lb.instructor_id
	JOIN users instructor_user ON instructor_user.id = i.user_id
	JOIN services s ON s.id = lb.service_id
`

func (r *PostgresRepository) GetLessonBookingByID(ctx context.Context, id int) (*AdminLessonBooking, error) {
	query := lessonBookingSelect + " WHERE lb.id = $1"

	var booking AdminLessonBooking
	err := r.pool.QueryRow(ctx, query, id).Scan(&booking.ID, &booking.UserID, &booking.ClientName, &booking.InstructorID, &booking.InstructorName, &booking.ServiceID, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status, &booking.TotalPrice, &booking.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get lesson booking: %w", err)
	}

	return &booking, nil
}

func (r *PostgresRepository) GetLessonBookings(ctx context.Context, filter LessonBookingFilter) ([]AdminLessonBooking, error) {
	query := lessonBookingSelect + `
		WHERE ($1::text IS NULL OR lb.status = $1)
		  AND ($2::date IS NULL OR (lb.starts_at AT TIME ZONE 'Europe/Moscow')::date = $2::date)
		  AND ($3::integer IS NULL OR lb.user_id = $3)
		  AND ($4::integer IS NULL OR lb.instructor_id = $4)
		ORDER BY lb.starts_at DESC, lb.id DESC
	`

	rows, err := r.pool.Query(ctx, query, filter.Status, filter.Date, filter.UserID, filter.InstructorID)
	if err != nil {
		return nil, fmt.Errorf("get lesson bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]AdminLessonBooking, 0)
	for rows.Next() {
		var booking AdminLessonBooking
		err := rows.Scan(&booking.ID, &booking.UserID, &booking.ClientName, &booking.InstructorID, &booking.InstructorName, &booking.ServiceID, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status, &booking.TotalPrice, &booking.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan lesson booking: %w", err)
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read lesson bookings: %w", err)
	}

	return bookings, nil
}

func (r *PostgresRepository) UpdateLessonBookingStatus(ctx context.Context, id int, status string) (*AdminLessonBooking, error) {
	query := `
		UPDATE lesson_bookings
		SET status = $2
		WHERE id = $1
		  AND status = 'confirmed'
		  AND (
			  ($2 = 'cancelled' AND starts_at > NOW())
			  OR ($2 IN ('completed', 'no_show') AND starts_at <= NOW())
		  )
		RETURNING id
	`

	var updatedID int
	err := r.pool.QueryRow(ctx, query, id, status).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		checkErr := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM lesson_bookings WHERE id = $1)", id).Scan(&exists)
		if checkErr != nil {
			return nil, fmt.Errorf("check lesson booking: %w", checkErr)
		}
		if exists {
			return nil, ErrInvalidTransition
		}
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update lesson booking status: %w", dbError(err))
	}

	return r.GetLessonBookingByID(ctx, updatedID)
}

// INSTRUCTOR BOOKING STATUS

func (r *PostgresRepository) UpdateOwnLessonBookingStatus(ctx context.Context, userID, bookingID int, status string) (*InstructorBooking, error) {
	query := `
		UPDATE lesson_bookings lb
		SET status = $3
		FROM instructors i
		WHERE lb.id = $2
		  AND i.id = lb.instructor_id
		  AND i.user_id = $1
		  AND lb.status = 'confirmed'
		  AND lb.starts_at <= NOW()
		  AND $3 IN ('completed', 'no_show')
		RETURNING lb.id
	`

	var id int
	err := r.pool.QueryRow(ctx, query, userID, bookingID, status).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		checkErr := r.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM lesson_bookings lb
				JOIN instructors i ON i.id = lb.instructor_id
				WHERE lb.id = $1 AND i.user_id = $2
			)
		`, bookingID, userID).Scan(&exists)
		if checkErr != nil {
			return nil, fmt.Errorf("check instructor booking: %w", checkErr)
		}
		if exists {
			return nil, ErrInvalidTransition
		}
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update instructor booking status: %w", dbError(err))
	}

	query = `
		SELECT lb.id, u.name, s.name, lb.starts_at, lb.ends_at, lb.status
		FROM lesson_bookings lb
		JOIN users u ON u.id = lb.user_id
		JOIN services s ON s.id = lb.service_id
		WHERE lb.id = $1
	`

	var booking InstructorBooking
	err = r.pool.QueryRow(ctx, query, id).Scan(&booking.ID, &booking.ClientName, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status)
	if err != nil {
		return nil, fmt.Errorf("load instructor booking: %w", err)
	}

	return &booking, nil
}

// EQUIPMENT RENTALS

func (r *PostgresRepository) GetEquipmentBookingByID(ctx context.Context, id int) (*AdminEquipmentRental, error) {
	query := `
		SELECT eb.id, eb.user_id, u.name, eb.starts_at, eb.ends_at,
		       eb.status, eb.total_price, eb.created_at
		FROM equipment_bookings eb
		JOIN users u ON u.id = eb.user_id
		WHERE eb.id = $1
	`

	var booking AdminEquipmentRental
	err := r.pool.QueryRow(ctx, query, id).Scan(&booking.ID, &booking.UserID, &booking.ClientName, &booking.StartsAt, &booking.EndsAt, &booking.Status, &booking.TotalPrice, &booking.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get equipment booking: %w", err)
	}

	booking.Items = make([]RentalItem, 0)
	booking.Sets = make([]BookedRentalSet, 0)

	query = `
		SELECT ebi.equipment_unit_id, e.id, e.name, eu.size,
		       ebi.booked_set_id, ebi.item_total_price
		FROM equipment_booking_items ebi
		JOIN equipment_units eu ON eu.id = ebi.equipment_unit_id
		JOIN equipment e ON e.id = eu.equipment_id
		WHERE ebi.booking_id = $1
		ORDER BY ebi.id
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get rental items: %w", err)
	}

	for rows.Next() {
		var item RentalItem
		err = rows.Scan(&item.EquipmentUnitID, &item.EquipmentID, &item.Name, &item.Size, &item.BookedSetID, &item.ItemTotalPrice)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan rental item: %w", err)
		}
		booking.Items = append(booking.Items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("read rental items: %w", err)
	}

	query = `
		SELECT bes.id, bes.equipment_set_id, es.name, bes.total_price
		FROM booked_equipment_sets bes
		JOIN equipment_sets es ON es.id = bes.equipment_set_id
		WHERE bes.booking_id = $1
		ORDER BY bes.id
	`

	rows, err = r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get booked equipment sets: %w", err)
	}

	for rows.Next() {
		var set BookedRentalSet
		err = rows.Scan(&set.ID, &set.EquipmentSetID, &set.Name, &set.TotalPrice)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan booked set: %w", err)
		}
		booking.Sets = append(booking.Sets, set)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("read booked sets: %w", err)
	}

	return &booking, nil
}

func (r *PostgresRepository) GetEquipmentBookings(ctx context.Context, filter EquipmentBookingFilter) ([]AdminEquipmentRental, error) {
	query := `
		SELECT id
		FROM equipment_bookings
		WHERE ($1::text IS NULL OR status = $1)
		  AND ($2::date IS NULL OR (starts_at AT TIME ZONE 'Europe/Moscow')::date = $2::date)
		  AND ($3::integer IS NULL OR user_id = $3)
		ORDER BY starts_at DESC, id DESC
	`

	rows, err := r.pool.Query(ctx, query, filter.Status, filter.Date, filter.UserID)
	if err != nil {
		return nil, fmt.Errorf("get equipment bookings: %w", err)
	}

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan equipment booking ID: %w", err)
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("read equipment booking IDs: %w", err)
	}

	bookings := make([]AdminEquipmentRental, 0, len(ids))
	for _, id := range ids {
		booking, err := r.GetEquipmentBookingByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, *booking)
	}

	return bookings, nil
}

func (r *PostgresRepository) UpdateEquipmentBookingStatus(ctx context.Context, id int, status string) (*AdminEquipmentRental, error) {
	query := `
		UPDATE equipment_bookings
		SET status = $2
		WHERE id = $1
		  AND (
			  (status = 'pending' AND $2 IN ('confirmed', 'cancelled'))
			  OR (status = 'confirmed' AND $2 = 'issued')
			  OR (status = 'confirmed' AND $2 = 'cancelled' AND starts_at > NOW())
			  OR (status = 'issued' AND $2 = 'returned')
		  )
		RETURNING id
	`

	var updatedID int
	err := r.pool.QueryRow(ctx, query, id, status).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		checkErr := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM equipment_bookings WHERE id = $1)", id).Scan(&exists)
		if checkErr != nil {
			return nil, fmt.Errorf("check equipment booking: %w", checkErr)
		}
		if exists {
			return nil, ErrInvalidTransition
		}
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update equipment booking status: %w", dbError(err))
	}

	return r.GetEquipmentBookingByID(ctx, updatedID)
}

// EQUIPMENT

func (r *PostgresRepository) CreateEquipment(ctx context.Context, params CreateEquipmentParams) (*Equipment, error) {
	query := `
		INSERT INTO equipment (category_id, name, description, image_url, price_per_hour)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, category_id, name, description, image_url, price_per_hour
	`

	var equipment Equipment
	err := r.pool.QueryRow(ctx, query, params.CategoryID, params.Name, params.Description, params.ImageURL, params.PricePerHour).Scan(&equipment.ID, &equipment.CategoryID, &equipment.Name, &equipment.Description, &equipment.ImageURL, &equipment.PricePerHour)
	if err != nil {
		return nil, fmt.Errorf("create equipment: %w", dbError(err))
	}

	return &equipment, nil
}

func (r *PostgresRepository) UpdateEquipment(ctx context.Context, id int, params UpdateEquipmentParams) (*Equipment, error) {
	query := `
		UPDATE equipment
		SET category_id = COALESCE($2, category_id),
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			image_url = COALESCE($5, image_url),
			price_per_hour = COALESCE($6, price_per_hour)
		WHERE id = $1
		RETURNING id, category_id, name, description, image_url, price_per_hour
	`

	var equipment Equipment
	err := r.pool.QueryRow(ctx, query, id, params.CategoryID, params.Name, params.Description, params.ImageURL, params.PricePerHour).Scan(&equipment.ID, &equipment.CategoryID, &equipment.Name, &equipment.Description, &equipment.ImageURL, &equipment.PricePerHour)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update equipment: %w", dbError(err))
	}

	return &equipment, nil
}

func (r *PostgresRepository) DeleteEquipment(ctx context.Context, id int) error {
	query := `DELETE FROM equipment WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrConflict
		}
		return fmt.Errorf("delete equipment: %w", dbError(err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// EQUIPMENT UNITS

func (r *PostgresRepository) GetEquipmentUnits(ctx context.Context, equipmentID *int) ([]EquipmentUnit, error) {
	query := `
		SELECT id, equipment_id, inventory_number, size, status, created_at
		FROM equipment_units
		WHERE ($1::integer IS NULL OR equipment_id = $1)
		ORDER BY equipment_id, id
	`

	rows, err := r.pool.Query(ctx, query, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("get equipment units: %w", err)
	}
	defer rows.Close()

	units := make([]EquipmentUnit, 0)
	for rows.Next() {
		var unit EquipmentUnit
		err := rows.Scan(&unit.ID, &unit.EquipmentID, &unit.InventoryNumber, &unit.Size, &unit.Status, &unit.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan equipment unit: %w", err)
		}
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read equipment units: %w", err)
	}

	return units, nil
}

func (r *PostgresRepository) CreateEquipmentUnit(ctx context.Context, params CreateEquipmentUnitParams) (*EquipmentUnit, error) {
	query := `
		INSERT INTO equipment_units (equipment_id, inventory_number, size, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, equipment_id, inventory_number, size, status, created_at
	`

	var unit EquipmentUnit
	err := r.pool.QueryRow(ctx, query, params.EquipmentID, params.InventoryNumber, params.Size, params.Status).Scan(&unit.ID, &unit.EquipmentID, &unit.InventoryNumber, &unit.Size, &unit.Status, &unit.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create equipment unit: %w", dbError(err))
	}

	return &unit, nil
}

func (r *PostgresRepository) UpdateEquipmentUnit(ctx context.Context, id int, params UpdateEquipmentUnitParams) (*EquipmentUnit, error) {
	query := `
		UPDATE equipment_units
		SET inventory_number = COALESCE($2, inventory_number),
			size = COALESCE($3, size),
			status = COALESCE($4, status)
		WHERE id = $1
		RETURNING id, equipment_id, inventory_number, size, status, created_at
	`

	var unit EquipmentUnit
	err := r.pool.QueryRow(ctx, query, id, params.InventoryNumber, params.Size, params.Status).Scan(&unit.ID, &unit.EquipmentID, &unit.InventoryNumber, &unit.Size, &unit.Status, &unit.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update equipment unit: %w", dbError(err))
	}

	return &unit, nil
}

// EQUIPMENT SETS

func (r *PostgresRepository) CreateEquipmentSet(ctx context.Context, params CreateEquipmentSetParams) (*EquipmentSet, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create equipment set: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO equipment_sets (name, description, sport_type, age_group, price_per_hour, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err = tx.QueryRow(ctx, query, params.Name, params.Description, params.SportType, params.AgeGroup, params.PricePerHour, params.IsActive).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create equipment set: %w", dbError(err))
	}

	for _, item := range params.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO equipment_set_items (equipment_set_id, equipment_category_id, quantity)
			VALUES ($1, $2, $3)
		`, id, item.CategoryID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("create equipment set item: %w", dbError(err))
		}
	}

	var set EquipmentSet
	err = tx.QueryRow(ctx, `
		SELECT id, name, description, sport_type, age_group, price_per_hour, is_active
		FROM equipment_sets
		WHERE id = $1
	`, id).Scan(&set.ID, &set.Name, &set.Description, &set.SportType, &set.AgeGroup, &set.PricePerHour, &set.IsActive)
	if err != nil {
		return nil, fmt.Errorf("load created equipment set: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT esi.equipment_category_id, ec.name, esi.quantity
		FROM equipment_set_items esi
		JOIN equipment_categories ec ON ec.id = esi.equipment_category_id
		WHERE esi.equipment_set_id = $1
		ORDER BY esi.equipment_category_id
	`, id)
	if err != nil {
		return nil, fmt.Errorf("load created set items: %w", err)
	}

	set.Items = make([]EquipmentSetItem, 0)
	for rows.Next() {
		var item EquipmentSetItem
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.Quantity); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan created set item: %w", err)
		}
		set.Items = append(set.Items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("read created set items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create equipment set: %w", err)
	}

	return &set, nil
}

func (r *PostgresRepository) UpdateEquipmentSet(ctx context.Context, id int, params UpdateEquipmentSetParams) (*EquipmentSet, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update equipment set: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE equipment_sets
		SET name = COALESCE($2, name),
			description = COALESCE($3, description),
			sport_type = COALESCE($4, sport_type),
			age_group = COALESCE($5, age_group),
			price_per_hour = COALESCE($6, price_per_hour),
			is_active = COALESCE($7, is_active)
		WHERE id = $1
		RETURNING id
	`

	var updatedID int
	err = tx.QueryRow(ctx, query, id, params.Name, params.Description, params.SportType, params.AgeGroup, params.PricePerHour, params.IsActive).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update equipment set: %w", dbError(err))
	}

	if params.Items != nil {
		_, err = tx.Exec(ctx, "DELETE FROM equipment_set_items WHERE equipment_set_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("delete old equipment set items: %w", dbError(err))
		}

		for _, item := range *params.Items {
			_, err = tx.Exec(ctx, `
				INSERT INTO equipment_set_items (equipment_set_id, equipment_category_id, quantity)
				VALUES ($1, $2, $3)
			`, id, item.CategoryID, item.Quantity)
			if err != nil {
				return nil, fmt.Errorf("replace equipment set item: %w", dbError(err))
			}
		}
	}

	var set EquipmentSet
	err = tx.QueryRow(ctx, `
		SELECT id, name, description, sport_type, age_group, price_per_hour, is_active
		FROM equipment_sets
		WHERE id = $1
	`, updatedID).Scan(&set.ID, &set.Name, &set.Description, &set.SportType, &set.AgeGroup, &set.PricePerHour, &set.IsActive)
	if err != nil {
		return nil, fmt.Errorf("load updated equipment set: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT esi.equipment_category_id, ec.name, esi.quantity
		FROM equipment_set_items esi
		JOIN equipment_categories ec ON ec.id = esi.equipment_category_id
		WHERE esi.equipment_set_id = $1
		ORDER BY esi.equipment_category_id
	`, updatedID)
	if err != nil {
		return nil, fmt.Errorf("load updated set items: %w", err)
	}

	set.Items = make([]EquipmentSetItem, 0)
	for rows.Next() {
		var item EquipmentSetItem
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.Quantity); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan updated set item: %w", err)
		}
		set.Items = append(set.Items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("read updated set items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update equipment set: %w", err)
	}

	return &set, nil
}

func (r *PostgresRepository) DeleteEquipmentSet(ctx context.Context, id int) error {
	query := `UPDATE equipment_sets SET is_active = false WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deactivate equipment set: %w", dbError(err))
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
