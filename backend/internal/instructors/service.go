package instructors

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

type InstructorRepository interface {
	ListInstructors(ctx context.Context) ([]Instructor, error)
	GetInstructorByID(ctx context.Context, id int) (Instructor, error)
	GetInstructorScheduleByID(ctx context.Context, instructorID int, date time.Time) ([]InstructorSchedule, error)
	GetBookedIntervals(ctx context.Context, instructorID int, date time.Time) ([]BookedInterval, error)
	GetMyInstructorBookings(ctx context.Context, userID int, date *time.Time) ([]InstructorBooking, error)
	GetInstructorByUserID(ctx context.Context, userID int) (int, error)
	GetServiceDuration(ctx context.Context, serviceID int) (int, error)
	InstructorProvidesService(ctx context.Context, instructorID, serviceID int) (bool, error)
	GetInstructorServices(ctx context.Context, instructorID int) ([]InstructorServiceInfo, error)
	UpdateInstructorBookingStatus(ctx context.Context, userID, bookingID int, status string) (InstructorBooking, error)
	InstructorBookingExists(ctx context.Context, userID, bookingID int) (bool, error)
}

type InstructorService struct {
	repo InstructorRepository
}

func NewInstructorService(repo InstructorRepository) *InstructorService {
	return &InstructorService{repo: repo}
}

func (s *InstructorService) ListInstructors(ctx context.Context) ([]Instructor, error) {
	instructors, err := s.repo.ListInstructors(ctx)

	if err != nil {
		return nil, fmt.Errorf("list instructors: %w", err)
	}

	return instructors, nil
}

var ErrInstructorInvalidID = errors.New("invalid id instructor")
var ErrInstructorNotFound = errors.New("instructor not found")

func (s *InstructorService) GetInstructorByID(ctx context.Context, id int) (Instructor, error) {
	if id <= 0 {
		return Instructor{}, ErrInstructorInvalidID
	}
	instructor, err := s.repo.GetInstructorByID(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Instructor{}, ErrInstructorNotFound
		}
		return Instructor{}, fmt.Errorf("getByID instructor: %w", err)
	}
	return instructor, nil
}

func (s *InstructorService) GetInstructorScheduleByID(ctx context.Context, id int, date time.Time) (InstructorScheduleResponse, error) {
	if id <= 0 {
		return InstructorScheduleResponse{}, ErrInstructorInvalidID
	}
	_, err := s.repo.GetInstructorByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InstructorScheduleResponse{}, ErrInstructorNotFound
		}
		return InstructorScheduleResponse{}, fmt.Errorf("getInstructorSchedule: getinstructorByID: %w", err)
	}
	schedule, err := s.repo.GetInstructorScheduleByID(ctx, id, date)
	if err != nil {
		return InstructorScheduleResponse{}, fmt.Errorf("getInstructorSchedule: schedule: %w", err)
	}
	bookedIntervals, err := s.repo.GetBookedIntervals(ctx, id, date)
	if err != nil {
		return InstructorScheduleResponse{}, fmt.Errorf("getInstructorSchedule: bookedInterval: %w", err)
	}

	response := InstructorScheduleResponse{
		InstructorID:    id,
		Date:            date.Format(time.DateOnly),
		Schedule:        schedule,
		BookedIntervals: bookedIntervals,
	}

	return response, nil
}

var ErrNotInstructor = errors.New("user is not an instructor")

func (s *InstructorService) GetMyInstructorBookings(ctx context.Context, userID int, date *time.Time) ([]InstructorBooking, error) {
	_, err := s.repo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []InstructorBooking{}, ErrNotInstructor
		}
		return []InstructorBooking{}, err
	}
	bookings, err := s.repo.GetMyInstructorBookings(ctx, userID, date)
	if err != nil {
		return []InstructorBooking{}, err
	}
	return bookings, nil
}

var ErrInvalidServiceID = errors.New("invalid service id")
var ErrInvalidDate = errors.New("date cannot be in the past")
var ErrInstructorNotWork = errors.New("instructor not work")
var ErrInstructorNotProvideService = errors.New("instructor not provied this service")
var ErrServiceNotFound = errors.New("service not found")

func (s *InstructorService) GetInstructorAvailableSlots(ctx context.Context, instructorID int, serviceID int, date time.Time) ([]AvailableSlot, error) {
	if instructorID <= 0 {
		return nil, ErrInstructorInvalidID
	}
	if serviceID <= 0 {
		return nil, ErrInvalidServiceID
	}

	// Работаем по московскому времени
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, err
	}

	now := time.Now().In(location)
	requestedDate := date.In(location)

	todayStart := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, location,
	)

	requestedDayStart := time.Date(
		requestedDate.Year(), requestedDate.Month(), requestedDate.Day(),
		0, 0, 0, 0, location,
	)

	requestedDayEnd := requestedDayStart.AddDate(0, 0, 1)

	// Запрещаем прошедшие даты
	if requestedDayStart.Before(todayStart) {
		return nil, ErrInvalidDate
	}

	// Проверяем инструктора
	instructor, err := s.repo.GetInstructorByID(ctx, instructorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstructorNotFound
		}
		return nil, err
	}

	if !instructor.IsActive {
		return nil, ErrInstructorNotWork
	}

	// Получаем длительность услуги
	duration, err := s.repo.GetServiceDuration(ctx, serviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	// Проверяем, оказывает ли инструктор эту услугу
	providesService, err := s.repo.InstructorProvidesService(
		ctx,
		instructorID,
		serviceID,
	)
	if err != nil {
		return nil, err
	}

	if !providesService {
		return nil, ErrInstructorNotProvideService
	}

	// Получаем рабочее расписание инструктора
	intervals, err := s.repo.GetInstructorScheduleByID(
		ctx,
		instructorID,
		requestedDayStart,
	)
	if err != nil {
		return nil, err
	}

	// Получаем занятые интервалы
	bookedIntervals, err := s.repo.GetBookedIntervals(
		ctx,
		instructorID,
		requestedDayStart,
	)
	if err != nil {
		return nil, err
	}

	// Создаём массив для свободных слотов
	slots := make([]AvailableSlot, 0)

	// Перебираем рабочие интервалы
	for _, interval := range intervals {
		workStart := interval.StartsAt
		workEnd := interval.EndsAt

		// Не начинаем генерацию раньше выбранного дня
		if workStart.Before(requestedDayStart) {
			workStart = requestedDayStart
		}

		// Генерируем слоты каждые 30 минут
		for slotStart := workStart; slotStart.Before(workEnd); slotStart = slotStart.Add(30 * time.Minute) {

			// Слот должен начинаться в выбранный день
			if !slotStart.Before(requestedDayEnd) {
				break
			}

			slotEnd := slotStart.Add(
				time.Duration(duration) * time.Minute,
			)

			// Не показываем слоты, которые уже начались
			if !slotStart.After(now) {
				continue
			}

			// Занятие должно помещаться в рабочее расписание
			if slotEnd.After(workEnd) {
				continue
			}

			// Проверяем пересечения с бронированиями
			isBusy := false

			for _, booking := range bookedIntervals {
				if slotStart.Before(booking.EndsAt) &&
					slotEnd.After(booking.StartsAt) {

					isBusy = true
					break
				}
			}

			if isBusy {
				continue
			}

			// Слот свободен
			slots = append(slots, AvailableSlot{
				StartsAt: slotStart.In(location),
				EndsAt:   slotEnd.In(location),
			})
		}
	}

	// Сортируем слоты по времени начала
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].StartsAt.Before(slots[j].StartsAt)
	})

	return slots, nil
}

func (s *InstructorService) GetInstructorServices(ctx context.Context, instructorID int) ([]InstructorServiceInfo, error) {
	if instructorID <= 0 {
		return nil, ErrInstructorInvalidID
	}

	_, err := s.repo.GetInstructorByID(ctx, instructorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstructorNotFound
		}
		return nil, err
	}

	services, err := s.repo.GetInstructorServices(ctx, instructorID)
	if err != nil {
		return nil, err
	}

	return services, nil
}

var ErrInvalidBookingStatus = errors.New("invalid booking status")
var ErrInstructorBookingNotFound = errors.New("instructor booking not found")
var ErrInstructorBookingConflict = errors.New("instructor booking status conflict")

func (s *InstructorService) UpdateInstructorBookingStatus(ctx context.Context, userID, bookingID int, status string) (InstructorBooking, error) {
	if bookingID <= 0 {
		return InstructorBooking{}, ErrInstructorInvalidID
	}

	if status != "completed" && status != "no_show" {
		return InstructorBooking{}, ErrInvalidBookingStatus
	}

	_, err := s.repo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InstructorBooking{}, ErrNotInstructor
		}
		return InstructorBooking{}, fmt.Errorf("get instructor by user ID: %w", err)
	}

	booking, err := s.repo.UpdateInstructorBookingStatus(ctx, userID, bookingID, status)
	if err == nil {
		return booking, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return InstructorBooking{}, fmt.Errorf("update instructor booking status: %w", err)
	}

	exists, err := s.repo.InstructorBookingExists(ctx, userID, bookingID)
	if err != nil {
		return InstructorBooking{}, fmt.Errorf("check instructor booking: %w", err)
	}

	if !exists {
		return InstructorBooking{}, ErrInstructorBookingNotFound
	}

	return InstructorBooking{}, ErrInstructorBookingConflict
}
