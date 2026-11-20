package bookings

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

func (r *PostgresRepository) CreateLessonBooking(ctx context.Context, booking CreateLessonBookingParams) (LessonBooking, error) {
	query := `
        INSERT INTO lesson_bookings (user_id, instructor_id, service_id, starts_at, ends_at, status, total_price)
        VALUES ($1, $2, $3, $4, $5, 'confirmed', $6)
        RETURNING
            id,
            user_id,
            instructor_id,
            service_id,
            starts_at,
            ends_at,
            status,
            total_price,
            created_at`

	var result LessonBooking

	err := r.pool.QueryRow(ctx, query, booking.UserID,
		booking.InstructorID, booking.ServiceID,
		booking.StartsAt, booking.EndsAt, booking.TotalPrice,
	).Scan(&result.ID, &result.UserID, &result.InstructorID,
		&result.ServiceID, &result.StartsAt,
		&result.EndsAt, &result.Status,
		&result.TotalPrice, &result.CreatedAt,
	)

	if err != nil {
		return LessonBooking{}, err
	}

	return result, nil
}

func (r *PostgresRepository) IsInstructorAvailable(ctx context.Context, instructorID int, startsAt, endsAt time.Time) (bool, error) {
	query := `SELECT EXISTS (
    		SELECT 1
    		FROM instructor_schedules
    		WHERE instructor_id = $1
      			AND starts_at <= $2
     			AND ends_at >= $3
				)`

	var ok bool
	err := r.pool.QueryRow(ctx, query, instructorID, startsAt, endsAt).Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *PostgresRepository) InstructorProvidesService(ctx context.Context, instructorID, serviceID int) (bool, error) {
	query := `SELECT EXISTS(
			SELECT 1
			FROM instructor_services
			WHERE instructor_id = $1
			AND service_id = $2
			)`

	var ok bool

	err := r.pool.QueryRow(ctx, query, instructorID, serviceID).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *PostgresRepository) GetServiceForBooking(ctx context.Context, serviceID int) (BookingServiceInfo, error) {
	query := `SELECT duration_minutes, price, is_active
				FROM services
				WHERE id = $1`

	var bookingService BookingServiceInfo

	err := r.pool.QueryRow(ctx, query, serviceID).Scan(&bookingService.DurationMinutes, &bookingService.Price, &bookingService.IsActive)
	if err != nil {
		return BookingServiceInfo{}, err
	}

	return bookingService, nil
}

func (r *PostgresRepository) IsInstructorActive(ctx context.Context, instructorID int) (bool, error) {
	query := `SELECT EXISTS(
			SELECT 1
			FROM instructors
			WHERE id = $1
			AND is_active = true
	)`

	var ok bool

	err := r.pool.QueryRow(ctx, query, instructorID).Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil

}

func (r *PostgresRepository) HasBookingConflict(ctx context.Context, instructorID int, startsAt, endsAt time.Time) (bool, error) {
	query := `SELECT EXISTS (
    	SELECT 1
    	FROM lesson_bookings
    	WHERE instructor_id = $1
      		AND starts_at < $3
      		AND ends_at > $2
      		AND status IN ('confirmed', 'completed', 'no_show')
	)`

	var ok bool

	err := r.pool.QueryRow(ctx, query, instructorID, startsAt, endsAt).Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *PostgresRepository) GetMyLessonBookings(ctx context.Context, userID int, status *string) ([]LessonBookingDetail, error) {
	query := `SELECT lb.id, lb.user_id, lb.instructor_id, u.name, lb.service_id, s.name, lb.starts_at, lb.ends_at, lb.status, lb.total_price, lb.created_at
		FROM lesson_bookings lb
		JOIN instructors i ON i.id = lb.instructor_id
		JOIN users u ON u.id = i.user_id
		JOIN services s ON s.id = lb.service_id
		WHERE lb.user_id = $1
		  AND ($2::text IS NULL OR lb.status = $2)
		ORDER BY lb.starts_at DESC, lb.id DESC
	`

	rows, err := r.pool.Query(ctx, query, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]LessonBookingDetail, 0)

	for rows.Next() {
		var booking LessonBookingDetail

		err := rows.Scan(&booking.ID, &booking.UserID, &booking.InstructorID, &booking.InstructorName, &booking.ServiceID,
			&booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status, &booking.TotalPrice, &booking.CreatedAt)
		if err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *PostgresRepository) GetLessonBookingByID(ctx context.Context, bookingID int) (LessonBookingDetail, error) {
	query := `SELECT lb.id,lb.user_id, lb.instructor_id, u.name, lb.service_id, s.name, lb.starts_at, lb.ends_at, lb.status, lb.total_price, lb.created_at
		FROM lesson_bookings lb
		JOIN instructors i ON i.id = lb.instructor_id
		JOIN users u ON u.id = i.user_id
		JOIN services s ON s.id = lb.service_id
		WHERE lb.id = $1
	`

	var booking LessonBookingDetail

	err := r.pool.QueryRow(ctx, query, bookingID).Scan(&booking.ID, &booking.UserID, &booking.InstructorID, &booking.InstructorName,
		&booking.ServiceID, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status,
		&booking.TotalPrice, &booking.CreatedAt)
	if err != nil {
		return LessonBookingDetail{}, err
	}

	return booking, nil
}

func (r *PostgresRepository) CancelLessonBooking(ctx context.Context, bookingID, userID int) (LessonBookingDetail, error) {
	query := `WITH updated AS (
			UPDATE lesson_bookings
			SET status = 'cancelled'
			WHERE id = $1
			  AND user_id = $2
			  AND status = 'confirmed'
			  AND starts_at > NOW()
			RETURNING
				id, user_id, instructor_id, service_id,
				starts_at, ends_at, status, total_price,
				created_at
		)
		SELECT lb.id, lb.user_id, lb.instructor_id, u.name, lb.service_id, s.name, lb.starts_at, lb.ends_at, lb.status, lb.total_price, lb.created_at
		FROM updated lb
		JOIN instructors i ON i.id = lb.instructor_id
		JOIN users u ON u.id = i.user_id
		JOIN services s ON s.id = lb.service_id
	`

	var booking LessonBookingDetail

	err := r.pool.QueryRow(ctx, query, bookingID, userID).Scan(&booking.ID, &booking.UserID, &booking.InstructorID, &booking.InstructorName,
		&booking.ServiceID, &booking.ServiceName, &booking.StartsAt, &booking.EndsAt, &booking.Status,
		&booking.TotalPrice, &booking.CreatedAt)
	if err != nil {
		return LessonBookingDetail{}, err
	}

	return booking, nil
}
