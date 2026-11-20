package rental

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type RentalRepository interface {
	GetEquipmentAvailability(ctx context.Context, startsAt, endsAt time.Time, equipmentID *int) ([]EquipmentAvailability, error)
	CreateEquipmentBooking(ctx context.Context, userID int, params CreateEquipmentBookingParams) (EquipmentRental, error)
	GetMyEquipmentBookings(ctx context.Context, userID int, status *string) ([]EquipmentRental, error)
	GetEquipmentBookingByID(ctx context.Context, bookingID int) (EquipmentRental, error)
	CancelEquipmentBooking(ctx context.Context, bookingID, userID int) (EquipmentRental, error)
}

type Service struct {
	repo RentalRepository
}

func NewService(repo RentalRepository) *Service {
	return &Service{repo: repo}
}

var (
	ErrRentalNotFound      = errors.New("rental not found")
	ErrRentalForbidden     = errors.New("rental belongs to another user")
	ErrInvalidRentalPeriod = errors.New("invalid rental period")
	ErrInvalidRentalStatus = errors.New("invalid rental status")
)

// Проверяем период аренды.
func validateRentalPeriod(startsAt, endsAt time.Time) error {
	if startsAt.IsZero() || endsAt.IsZero() {
		return ErrInvalidRentalPeriod
	}

	if !startsAt.After(time.Now()) {
		return ErrInvalidRentalPeriod
	}

	if !endsAt.After(startsAt) {
		return ErrInvalidRentalPeriod
	}

	duration := endsAt.Sub(startsAt)

	if duration < time.Hour || duration%time.Hour != 0 {
		return ErrInvalidRentalPeriod
	}

	return nil
}

// Получаем доступность оборудования.
func (s *Service) GetEquipmentAvailability(ctx context.Context, startsAt, endsAt time.Time, equipmentID *int) ([]EquipmentAvailability, error) {
	if err := validateRentalPeriod(startsAt, endsAt); err != nil {
		return nil, err
	}

	if equipmentID != nil && *equipmentID <= 0 {
		return nil, ErrInvalidRentalInput
	}

	items, err := s.repo.GetEquipmentAvailability(
		ctx, startsAt, endsAt, equipmentID,
	)
	if err != nil {
		if errors.Is(err, ErrEquipmentNotFound) {
			return nil, ErrEquipmentNotFound
		}

		return nil, fmt.Errorf("get equipment availability: %w", err)
	}

	return items, nil
}

// Создаём бронирование оборудования.
func (s *Service) CreateEquipmentBooking(ctx context.Context, userID int, params CreateEquipmentBookingParams) (EquipmentRental, error) {
	if userID <= 0 {
		return EquipmentRental{}, ErrInvalidRentalInput
	}

	if err := validateRentalPeriod(params.StartsAt, params.EndsAt); err != nil {
		return EquipmentRental{}, err
	}

	if len(params.Items) == 0 && len(params.Sets) == 0 {
		return EquipmentRental{}, ErrInvalidRentalInput
	}

	// Проверяем отдельные позиции оборудования.
	for _, item := range params.Items {
		if item.EquipmentID <= 0 ||
			item.Quantity <= 0 ||
			strings.TrimSpace(item.Size) == "" {
			return EquipmentRental{}, ErrInvalidRentalInput
		}
	}

	// Проверяем комплекты.
	for _, set := range params.Sets {
		if set.EquipmentSetID <= 0 ||
			set.Quantity <= 0 ||
			len(set.Sizes) == 0 {
			return EquipmentRental{}, ErrInvalidRentalInput
		}

		usedCategories := make(map[int]bool)

		for _, size := range set.Sizes {
			if size.CategoryID <= 0 ||
				strings.TrimSpace(size.Size) == "" {
				return EquipmentRental{}, ErrInvalidRentalInput
			}

			if usedCategories[size.CategoryID] {
				return EquipmentRental{}, ErrInvalidSetSizes
			}

			usedCategories[size.CategoryID] = true
		}
	}

	booking, err := s.repo.CreateEquipmentBooking(
		ctx, userID, params,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrEquipmentUnavailable):
			return EquipmentRental{}, ErrEquipmentUnavailable

		case errors.Is(err, ErrEquipmentNotFound):
			return EquipmentRental{}, ErrEquipmentNotFound

		case errors.Is(err, ErrEquipmentSetNotFound):
			return EquipmentRental{}, ErrEquipmentSetNotFound

		case errors.Is(err, ErrEquipmentSetInactive):
			return EquipmentRental{}, ErrEquipmentSetInactive

		case errors.Is(err, ErrInvalidSetSizes):
			return EquipmentRental{}, ErrInvalidSetSizes

		case errors.Is(err, ErrInvalidRentalInput):
			return EquipmentRental{}, ErrInvalidRentalInput

		default:
			return EquipmentRental{},
				fmt.Errorf("create equipment booking: %w", err)
		}
	}

	return booking, nil
}

// Получаем свои бронирования оборудования.
func (s *Service) GetMyEquipmentBookings(ctx context.Context, userID int, status *string) ([]EquipmentRental, error) {
	if userID <= 0 {
		return nil, ErrInvalidRentalInput
	}

	if status != nil {
		switch *status {
		case "pending", "confirmed", "issued", "returned", "cancelled":
		default:
			return nil, ErrInvalidRentalStatus
		}
	}

	bookings, err := s.repo.GetMyEquipmentBookings(
		ctx, userID, status,
	)
	if err != nil {
		return nil, fmt.Errorf("get my equipment bookings: %w", err)
	}

	return bookings, nil
}

// Получаем конкретное бронирование оборудования.
func (s *Service) GetEquipmentBookingByID(ctx context.Context, userID, bookingID int) (EquipmentRental, error) {
	if userID <= 0 || bookingID <= 0 {
		return EquipmentRental{}, ErrInvalidRentalInput
	}

	booking, err := s.repo.GetEquipmentBookingByID(
		ctx, bookingID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EquipmentRental{}, ErrRentalNotFound
		}

		return EquipmentRental{},
			fmt.Errorf("get equipment booking: %w", err)
	}

	if booking.UserID != userID {
		return EquipmentRental{}, ErrRentalForbidden
	}

	return booking, nil
}

// Отменяем бронирование оборудования.
func (s *Service) CancelEquipmentBooking(ctx context.Context, userID, bookingID int) (EquipmentRental, error) {
	if userID <= 0 || bookingID <= 0 {
		return EquipmentRental{}, ErrInvalidRentalInput
	}

	booking, err := s.GetEquipmentBookingByID(
		ctx, userID, bookingID,
	)
	if err != nil {
		return EquipmentRental{}, err
	}

	// Отменить можно только ожидающую или подтверждённую аренду.
	if booking.Status != "pending" && booking.Status != "confirmed" {
		return EquipmentRental{}, ErrRentalCannotCancel
	}

	// После начала аренды отмена недоступна.
	if !booking.StartsAt.After(time.Now()) {
		return EquipmentRental{}, ErrRentalCannotCancel
	}

	cancelled, err := s.repo.CancelEquipmentBooking(
		ctx, bookingID, userID,
	)
	if err != nil {
		if errors.Is(err, ErrRentalCannotCancel) {
			return EquipmentRental{}, ErrRentalCannotCancel
		}

		return EquipmentRental{},
			fmt.Errorf("cancel equipment booking: %w", err)
	}

	return cancelled, nil
}
