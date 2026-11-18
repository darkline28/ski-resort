package instructors

import (
	"context"
	"errors"
	"fmt"
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
