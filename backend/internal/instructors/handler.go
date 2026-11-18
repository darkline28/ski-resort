package instructors

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/darkline28/ski-resort/backend/internal/instructors/api"
	"github.com/darkline28/ski-resort/backend/internal/middleware"
)

type Handler struct {
	service *InstructorService
}

func NewHandler(service *InstructorService) *Handler {
	return &Handler{service: service}
}

// ListInstructors implements [api.StrictServerInterface].
func (h *Handler) ListInstructors(ctx context.Context, request api.ListInstructorsRequestObject) (api.ListInstructorsResponseObject, error) {
	serviceInstructors, err := h.service.ListInstructors(ctx)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to list instructors",
			"error", err,
		)
		return api.ListInstructors500JSONResponse{Error: "internal_server_error", Message: "Внутренняя ошибка сервера"}, nil
	}
	slog.InfoContext(ctx,
		"list instructors",
		"count", len(serviceInstructors))

	apiInstructors := make([]api.Instructors, 0, len(serviceInstructors))
	for _, inst := range serviceInstructors {
		instructor := api.Instructors{
			Id:              inst.ID,
			Name:            inst.Name,
			Description:     inst.Description,
			ExperienceYears: inst.ExperienceYears,
			Specialization:  inst.Specialization,
			IsActive:        inst.IsActive,
			ImageUrl:        inst.ImageURL,
		}
		apiInstructors = append(apiInstructors, instructor)
	}

	return api.ListInstructors200JSONResponse(apiInstructors), nil

}

// GetInstructorByID implements [api.StrictServerInterface].
func (h *Handler) GetInstructorByID(ctx context.Context, request api.GetInstructorByIDRequestObject) (api.GetInstructorByIDResponseObject, error) {
	instructor, err := h.service.GetInstructorByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, ErrInstructorInvalidID) {
			return api.GetInstructorByID400JSONResponse{Error: "invalid_instructor_id", Message: "Некорректный ID инструктора"}, nil
		}
		if errors.Is(err, ErrInstructorNotFound) {
			return api.GetInstructorByID404JSONResponse{Error: "instructor_not_found", Message: "Инструктор не найден"}, nil
		}
		slog.ErrorContext(ctx,
			"failed to get instructor by ID",
			"instructor_id", request.Id,
			"error", err,
		)
		return api.GetInstructorByID500JSONResponse{Error: "internal_server_error", Message: "Внутренняя ошибка сервера"}, nil
	}

	response := api.Instructors{
		Id:              instructor.ID,
		Name:            instructor.Name,
		Specialization:  instructor.Specialization,
		ExperienceYears: instructor.ExperienceYears,
		Description:     instructor.Description,
		IsActive:        instructor.IsActive,
		ImageUrl:        instructor.ImageURL,
	}
	return api.GetInstructorByID200JSONResponse(response), nil
}

// GetInstructorScheduleByID implements [api.StrictServerInterface].
func (h *Handler) GetInstructorScheduleByID(ctx context.Context, request api.GetInstructorScheduleByIDRequestObject) (api.GetInstructorScheduleByIDResponseObject, error) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		slog.ErrorContext(ctx, "failed to load timezone", "error", err)
		return api.GetInstructorScheduleByID500JSONResponse{Error: "internal_server_error", Message: "Внутренняя ошибка сервера"}, nil
	}
	date := request.Params.Date
	startOfDay := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		0, 0, 0, 0,
		loc,
	)

	scheduleResponse, err := h.service.GetInstructorScheduleByID(ctx, request.Id, startOfDay)
	if err != nil {
		if errors.Is(err, ErrInstructorInvalidID) {
			return api.GetInstructorScheduleByID400JSONResponse{Error: "invalid_instructor_id", Message: "Некорректный ID инструктора"}, nil
		}
		if errors.Is(err, ErrInstructorNotFound) {
			return api.GetInstructorScheduleByID404JSONResponse{Error: "instructor_not_found", Message: "Инструктор не найден"}, nil
		}
		slog.ErrorContext(ctx, "GetInstructorScheduleByID", "error", err)
		return api.GetInstructorScheduleByID500JSONResponse{Error: "internal_server_error", Message: "Внутренняя ошибка сервера"}, nil
	}

	apiBookedIntervals := make([]api.BookedInterval, 0, len(scheduleResponse.BookedIntervals))
	for _, interval := range scheduleResponse.BookedIntervals {
		apiInterval := api.BookedInterval{
			EndsAt:   interval.EndsAt,
			StartsAt: interval.StartsAt,
		}
		apiBookedIntervals = append(apiBookedIntervals, apiInterval)
	}

	apiSchedule := make([]api.InstructorSchedule, 0, len(scheduleResponse.Schedule))
	for _, interval := range scheduleResponse.Schedule {
		apiInterval := api.InstructorSchedule{
			EndsAt:   interval.EndsAt,
			Id:       interval.ID,
			StartsAt: interval.StartsAt,
		}
		apiSchedule = append(apiSchedule, apiInterval)
	}
	response := api.InstructorScheduleResponse{
		BookedIntervals: apiBookedIntervals,
		Date:            date,
		InstructorId:    request.Id,
		Schedule:        apiSchedule,
	}

	return api.GetInstructorScheduleByID200JSONResponse(response), nil

}

// GetMyInstructorBookings implements [api.StrictServerInterface].
func (h *Handler) GetMyInstructorBookings(ctx context.Context, request api.GetMyInstructorBookingsRequestObject) (api.GetMyInstructorBookingsResponseObject, error) {
	var startOfDay *time.Time
	date := request.Params.Date
	if date != nil {
		loc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			slog.ErrorContext(ctx, "failed to load timezone", "error", err)
			return api.GetMyInstructorBookings500JSONResponse{Error: "internal_server_error", Message: "Внутренняя ошибка сервера"}, nil
		}
		day := time.Date(
			date.Year(),
			date.Month(),
			date.Day(),
			0, 0, 0, 0,
			loc,
		)
		startOfDay = &day
	}

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetMyInstructorBookings401JSONResponse{Error: "unauthorized", Message: "Пользователь не авторизирован"}, nil
	}

	instructorBookings, err := h.service.GetMyInstructorBookings(ctx, userID, startOfDay)
	if err != nil {
		if errors.Is(err, ErrNotInstructor) {
			return api.GetMyInstructorBookings403JSONResponse{Error: "forbidden", Message: "Доступ разрешён только инструкторам"}, nil
		}
		slog.ErrorContext(ctx, "getMyInstructorBookings", "error", err)
		return api.GetMyInstructorBookings500JSONResponse{Error: "internal_server_error", Message: "Внутренняя ошибка сервера"}, nil
	}
	apiResponse := make([]api.InstructorBooking, 0, len(instructorBookings))

	for _, interval := range instructorBookings {
		booking := api.InstructorBooking{
			ClientName:  interval.ClientName,
			EndsAt:      interval.EndsAt,
			Id:          interval.ID,
			ServiceName: interval.ServiceName,
			StartsAt:    interval.StartsAt,
			Status:      api.InstructorBookingStatus(interval.Status),
		}
		apiResponse = append(apiResponse, booking)
	}

	return api.GetMyInstructorBookings200JSONResponse(apiResponse), nil
}
