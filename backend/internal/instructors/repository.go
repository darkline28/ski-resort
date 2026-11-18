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

// Get /instructors
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

// Get /instructors/{id}
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

// Get /instructors/{id}/schedule
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

// Get /instructors/me/bookings
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
