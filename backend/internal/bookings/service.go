package bookings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BookingRepository interface {
	CreateLessonBooking(ctx context.Context, booking CreateLessonBookingParams) (LessonBooking, error)
	IsInstructorAvailable(ctx context.Context, instructorID int, startsAt, endsAt time.Time) (bool, error)
	InstructorProvidesService(ctx context.Context, instructorID, serviceID int) (bool, error)
	GetServiceForBooking(ctx context.Context, serviceID int) (BookingServiceInfo, error)
	IsInstructorActive(ctx context.Context, instructorID int) (bool, error)
	HasBookingConflict(ctx context.Context, instructorID int, startsAt, endsAt time.Time) (bool, error)
	GetMyLessonBookings(ctx context.Context, userID int, status *string) ([]LessonBookingDetail, error)
	GetLessonBookingByID(ctx context.Context, bookingID int) (LessonBookingDetail, error)
	CancelLessonBooking(ctx context.Context, bookingID, userID int) (LessonBookingDetail, error)
}

type Service struct {
	repo BookingRepository
}

func NewService(repo BookingRepository) *Service {
	return &Service{repo: repo}
}

var (
	ErrInvalidBookingInput       = errors.New("invalid booking input")
	ErrBookingInPast             = errors.New("booking start time must be in the future")
	ErrBookingServiceNotFound    = errors.New("service not found")
	ErrBookingServiceInactive    = errors.New("service is inactive")
	ErrBookingInstructorInactive = errors.New("instructor not found or inactive")
	ErrInstructorNotProvide      = errors.New("instructor does not provide this service")
	ErrInstructorNotAvailable    = errors.New("instructor is not available")
	ErrBookingConflict           = errors.New("booking time conflict")
)

func (s *Service) CreateLessonBooking(ctx context.Context, userID, instructorID, serviceID int, startsAt time.Time) (LessonBooking, error) {

	if userID <= 0 || instructorID <= 0 || serviceID <= 0 || startsAt.IsZero() {
		return LessonBooking{}, ErrInvalidBookingInput
	}

	// Нельзя записаться в прошлое
	if !startsAt.After(time.Now()) {
		return LessonBooking{}, ErrBookingInPast
	}

	// Получаем информацию об услуге
	service, err := s.repo.GetServiceForBooking(ctx, serviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LessonBooking{}, ErrBookingServiceNotFound
		}
		return LessonBooking{}, fmt.Errorf("get service for booking: %w", err)
	}

	if !service.IsActive {
		return LessonBooking{}, ErrBookingServiceInactive
	}

	// Проверяем активность инструктора
	active, err := s.repo.IsInstructorActive(ctx, instructorID)
	if err != nil {
		return LessonBooking{}, fmt.Errorf("check instructor activity: %w", err)
	}
	if !active {
		return LessonBooking{}, ErrBookingInstructorInactive
	}

	// Проверяем, оказывает ли инструктор эту услугу
	providesService, err := s.repo.InstructorProvidesService(ctx, instructorID, serviceID)
	if err != nil {
		return LessonBooking{}, fmt.Errorf("check instructor service: %w", err)
	}
	if !providesService {
		return LessonBooking{}, ErrInstructorNotProvide
	}

	// Рассчитываем окончание занятия
	endsAt := startsAt.Add(time.Duration(service.DurationMinutes) * time.Minute)

	// Проверяем рабочее расписание инструктора
	available, err := s.repo.IsInstructorAvailable(ctx, instructorID, startsAt, endsAt)
	if err != nil {
		return LessonBooking{}, fmt.Errorf("check instructor availability: %w", err)
	}
	if !available {
		return LessonBooking{}, ErrInstructorNotAvailable
	}

	hasConflict, err := s.repo.HasBookingConflict(ctx, instructorID, startsAt, endsAt)
	if err != nil {
		return LessonBooking{}, fmt.Errorf("check booking conflict: %w", err)
	}

	if hasConflict {
		return LessonBooking{}, ErrBookingConflict
	}

	// Формируем данные для бронирования
	params := CreateLessonBookingParams{
		UserID:       userID,
		InstructorID: instructorID,
		ServiceID:    serviceID,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		TotalPrice:   service.Price,
	}

	// Создаём бронирование
	booking, err := s.repo.CreateLessonBooking(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return LessonBooking{}, ErrBookingConflict
		}

		return LessonBooking{}, fmt.Errorf("create lesson booking: %w", err)
	}

	return booking, nil
}

var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingForbidden     = errors.New("booking belongs to another user")
	ErrBookingCannotCancel  = errors.New("booking cannot be cancelled")
	ErrInvalidBookingStatus = errors.New("invalid booking status")
)

func (s *Service) GetMyLessonBookings(ctx context.Context, userID int, status *string) ([]LessonBookingDetail, error) {
	if userID <= 0 {
		return nil, ErrInvalidBookingInput
	}

	if status != nil {
		switch *status {
		case "confirmed", "cancelled", "completed", "no_show":
		default:
			return nil, ErrInvalidBookingStatus
		}
	}

	bookings, err := s.repo.GetMyLessonBookings(
		ctx, userID, status,
	)
	if err != nil {
		return nil, fmt.Errorf("get my lesson bookings: %w", err)
	}

	return bookings, nil
}

func (s *Service) GetLessonBookingByID(ctx context.Context, userID, bookingID int) (LessonBookingDetail, error) {
	if userID <= 0 || bookingID <= 0 {
		return LessonBookingDetail{}, ErrInvalidBookingInput
	}

	booking, err := s.repo.GetLessonBookingByID(
		ctx, bookingID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LessonBookingDetail{}, ErrBookingNotFound
		}

		return LessonBookingDetail{},
			fmt.Errorf("get lesson booking: %w", err)
	}

	if booking.UserID != userID {
		return LessonBookingDetail{}, ErrBookingForbidden
	}

	return booking, nil
}

func (s *Service) CancelLessonBooking(ctx context.Context, userID, bookingID int) (LessonBookingDetail, error) {
	if userID <= 0 || bookingID <= 0 {
		return LessonBookingDetail{}, ErrInvalidBookingInput
	}

	// Проверяем существование и владельца
	booking, err := s.GetLessonBookingByID(
		ctx, userID, bookingID,
	)
	if err != nil {
		return LessonBookingDetail{}, err
	}

	// Отменять можно только подтверждённое занятие
	if booking.Status != "confirmed" {
		return LessonBookingDetail{}, ErrBookingCannotCancel
	}

	// Нельзя отменять занятие после его начала
	if !booking.StartsAt.After(time.Now()) {
		return LessonBookingDetail{}, ErrBookingCannotCancel
	}

	// Атомарно меняем статус в PostgreSQL
	cancelled, err := s.repo.CancelLessonBooking(
		ctx, bookingID, userID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LessonBookingDetail{}, ErrBookingCannotCancel
		}

		return LessonBookingDetail{},
			fmt.Errorf("cancel lesson booking: %w", err)
	}

	return cancelled, nil
}
