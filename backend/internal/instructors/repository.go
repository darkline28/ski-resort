package instructors

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// ListInstructors returns all instructors
func (r *PostgresRepository) ListInstructors(ctx context.Context) ([]Instructor, error) {
	query := `SELECT i.id, u.name AS name, i.specialization, i.description, i.experience_years, i.is_active, i.image_url
	FROM instructors i
	INNER JOIN users u
	ON i.user_id = u.id
	ORDER BY i.is_active DESC, i.id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instructors := make([]Instructor, 0)

	for rows.Next() {
		var instructor Instructor

		err := rows.Scan(&instructor.ID, &instructor.Name, &instructor.Specialization, &instructor.Description, &instructor.ExperienceYears, &instructor.IsActive, &instructor.ImageURL)
		if err != nil {
			return nil, err
		}
		instructors = append(instructors, instructor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instructors, nil
}

// GetInstructorByID returns an instructor by instructor ID
func (r *PostgresRepository) GetInstructorByID(ctx context.Context, instructorID int) (Instructor, error) {
	query := `SELECT i.id, u.name AS name, i.specialization, i.description, i.experience_years, i.is_active, i.image_url
	FROM instructors i
	INNER JOIN users u
	ON i.user_id = u.id
	WHERE i.id = $1`

	row := r.pool.QueryRow(ctx, query, instructorID)

	var instructor Instructor

	err := row.Scan(&instructor.ID, &instructor.Name, &instructor.Specialization, &instructor.Description, &instructor.ExperienceYears, &instructor.IsActive, &instructor.ImageURL)
	if err != nil {
		return Instructor{}, err
	}

	return instructor, nil
}

// GetInstructorScheduleByID returns instructor working intervals for a date
func (r *PostgresRepository) GetInstructorScheduleByID(ctx context.Context, instructorID int, date time.Time) ([]InstructorSchedule, error) {
	query := `SELECT id, starts_at, ends_at
        FROM instructor_schedules
        WHERE instructor_id = $1
          AND starts_at < $3
          AND ends_at > $2
        ORDER BY starts_at`

	startOfDay := date
	endOfDay := startOfDay.AddDate(0, 0, 1)

	rows, err := r.pool.Query(ctx, query, instructorID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedule := make([]InstructorSchedule, 0)

	for rows.Next() {
		var interval InstructorSchedule

		if err := rows.Scan(&interval.ID, &interval.StartsAt, &interval.EndsAt); err != nil {
			return nil, err
		}

		schedule = append(schedule, interval)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedule, nil
}

// GetBookedIntervals returns occupied time intervals for a date
func (r *PostgresRepository) GetBookedIntervals(ctx context.Context, instructorID int, date time.Time) ([]BookedInterval, error) {
	query := `SELECT starts_at, ends_at
        FROM lesson_bookings
        WHERE instructor_id = $1
          AND starts_at < $3
          AND ends_at > $2
          AND status IN ('confirmed', 'completed', 'no_show')
        ORDER BY starts_at`

	startOfDay := date
	endOfDay := startOfDay.AddDate(0, 0, 1)

	rows, err := r.pool.Query(ctx, query, instructorID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	intervals := make([]BookedInterval, 0)

	for rows.Next() {
		var interval BookedInterval

		if err := rows.Scan(&interval.StartsAt, &interval.EndsAt); err != nil {
			return nil, err
		}

		intervals = append(intervals, interval)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return intervals, nil
}

// GetMyInstructorBookings returns bookings for the authenticated instructor
func (r *PostgresRepository) GetMyInstructorBookings(ctx context.Context, userID int, date *time.Time) ([]InstructorBooking, error) {
	query := `SELECT b.id, u.name AS client_name, s.name AS service_name, b.starts_at, b.ends_at, b.status
	FROM lesson_bookings b
	INNER JOIN services s
	ON b.service_id = s.id
	INNER JOIN users u
	ON b.user_id = u.id
	INNER JOIN instructors i
	ON b.instructor_id = i.id
	WHERE i.user_id = $1
	  AND (
      $2::timestamptz IS NULL
      OR (
          b.starts_at < $3
          AND b.ends_at > $2
     	 )
  		)	
	ORDER BY b.starts_at`

	var startOfDay, endOfDay *time.Time
	if date != nil {
		startOfDay = date
		nextDay := startOfDay.AddDate(0, 0, 1)
		endOfDay = &nextDay
	}

	rows, err := r.pool.Query(ctx, query, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings = make([]InstructorBooking, 0)

	for rows.Next() {
		var booking InstructorBooking
		if err := rows.Scan(&booking.ID, &booking.ClientName, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil

}

// GetInstructorByUserID finds an instructor ID by user ID
func (r *PostgresRepository) GetInstructorByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT i.id AS instructor_id
	FROM instructors i
	WHERE i.user_id=$1`

	row := r.pool.QueryRow(ctx, query, userID)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil

}

// GetServiceDuration возвращает продолжительность активной услуги в минутах
func (r *PostgresRepository) GetServiceDuration(ctx context.Context, serviceID int) (int, error) {
	query := `
		SELECT duration_minutes
		FROM services
		WHERE id = $1 AND is_active = true
	`

	var duration int

	row := r.pool.QueryRow(ctx, query, serviceID)
	err := row.Scan(&duration)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// InstructorProvidesService проверяет, оказывает ли инструктор выбранную услугу
func (r *PostgresRepository) InstructorProvidesService(ctx context.Context, instructorID, serviceID int) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM instructor_services
			WHERE instructor_id = $1
			  AND service_id = $2
		)`

	var exists bool

	row := r.pool.QueryRow(ctx, query, instructorID, serviceID)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresRepository) GetInstructorServices(ctx context.Context, instructorID int) ([]InstructorServiceInfo, error) {
	query := `SELECT s.id, s.name, s.description, s.duration_minutes, s.price, s.is_active
        FROM instructor_services ins
        INNER JOIN services s
            ON ins.service_id = s.id
        WHERE ins.instructor_id = $1
          AND s.is_active = true
        ORDER BY s.id
    `

	rows, err := r.pool.Query(ctx, query, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]InstructorServiceInfo, 0)

	for rows.Next() {
		var service InstructorServiceInfo

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

		result = append(result, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresRepository) UpdateInstructorBookingStatus(ctx context.Context, userID, bookingID int, status string) (InstructorBooking, error) {
	query := `
		UPDATE lesson_bookings b
		SET status = $3
		FROM instructors i, users u, services s
		WHERE b.id = $2
		  AND b.instructor_id = i.id
		  AND i.user_id = $1
		  AND b.user_id = u.id
		  AND b.service_id = s.id
		  AND b.status = 'confirmed'
		  AND b.starts_at <= NOW()
		RETURNING b.id, u.name, s.name, b.starts_at, b.ends_at, b.status
	`

	var booking InstructorBooking

	err := r.pool.QueryRow(ctx, query, userID, bookingID, status).Scan(&booking.ID, &booking.ClientName, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status)
	if err != nil {
		return InstructorBooking{}, err
	}

	return booking, nil
}

func (r *PostgresRepository) InstructorBookingExists(ctx context.Context, userID, bookingID int) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM lesson_bookings b
			INNER JOIN instructors i ON b.instructor_id = i.id
			WHERE b.id = $1 AND i.user_id = $2
		)
	`

	var exists bool

	err := r.pool.QueryRow(ctx, query, bookingID, userID).Scan(&exists)
	return exists, err
}
