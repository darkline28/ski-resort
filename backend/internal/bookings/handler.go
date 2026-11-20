package bookings

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/bookings/api"
	"github.com/darkline28/ski-resort/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateLessonBooking implements [api.StrictServerInterface].
func (h *Handler) CreateLessonBooking(ctx context.Context, request api.CreateLessonBookingRequestObject) (api.CreateLessonBookingResponseObject, error) {
	// Проверяем авторизацию
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.CreateLessonBooking401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	// Проверяем права доступа
	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.CreateLessonBooking403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав для создания бронирования",
		}, nil
	}

	if request.Body == nil {
		return api.CreateLessonBooking400JSONResponse{
			Error:   "invalid_input",
			Message: "Некорректные данные запроса",
		}, nil
	}

	// Вызываем бизнес-логику
	booking, err := h.service.CreateLessonBooking(ctx, userID,
		request.Body.InstructorId, request.Body.ServiceId, request.Body.StartsAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBookingInput), errors.Is(err, ErrBookingInPast):
			return api.CreateLessonBooking400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные данные бронирования",
			}, nil

		case errors.Is(err, ErrBookingServiceNotFound), errors.Is(err, ErrBookingInstructorInactive):
			return api.CreateLessonBooking404JSONResponse{
				Error:   "not_found",
				Message: "Услуга или инструктор не найдены",
			}, nil

		case errors.Is(err, ErrBookingServiceInactive), errors.Is(err, ErrInstructorNotProvide):
			return api.CreateLessonBooking422JSONResponse{
				Error:   "booking_unavailable",
				Message: "Выбранная услуга недоступна у инструктора",
			}, nil

		case errors.Is(err, ErrInstructorNotAvailable), errors.Is(err, ErrBookingConflict):
			return api.CreateLessonBooking409JSONResponse{
				Error:   "booking_conflict",
				Message: "Выбранное время недоступно",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to create lesson booking",
				"user_id", userID,
				"instructor_id", request.Body.InstructorId,
				"service_id", request.Body.ServiceId,
				"error", err,
			)

			return api.CreateLessonBooking500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := api.CreateLessonBooking201JSONResponse{
		Id:           int(booking.ID),
		UserId:       int(booking.UserID),
		InstructorId: int(booking.InstructorID),
		ServiceId:    int(booking.ServiceID),
		StartsAt:     booking.StartsAt,
		EndsAt:       booking.EndsAt,
		Status:       api.LessonBookingStatus(booking.Status),
		TotalPrice:   int(booking.TotalPrice),
		CreatedAt:    booking.CreatedAt,
	}

	slog.InfoContext(ctx,
		"lesson booking created",
		"booking_id", booking.ID,
		"user_id", userID,
		"instructor_id", booking.InstructorID,
	)

	return response, nil
}

// GetMyLessonBookings implements [api.StrictServerInterface].
func (h *Handler) GetMyLessonBookings(ctx context.Context, request api.GetMyLessonBookingsRequestObject) (api.GetMyLessonBookingsResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetMyLessonBookings401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.GetMyLessonBookings403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	var status *string

	if request.Params.Status != nil {
		if !request.Params.Status.Valid() {
			return api.GetMyLessonBookings400JSONResponse{
				Error:   "invalid_status",
				Message: "Некорректный статус бронирования",
			}, nil
		}

		value := string(*request.Params.Status)
		status = &value
	}

	bookings, err := h.service.GetMyLessonBookings(
		ctx, userID, status,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidBookingStatus) ||
			errors.Is(err, ErrInvalidBookingInput) {
			return api.GetMyLessonBookings400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные параметры запроса",
			}, nil
		}

		slog.ErrorContext(ctx,
			"failed to get my lesson bookings",
			"user_id", userID,
			"error", err,
		)

		return api.GetMyLessonBookings500JSONResponse{
			Error:   "internal_error",
			Message: "Внутренняя ошибка сервера",
		}, nil
	}

	response := make([]api.LessonBookingDetail, 0, len(bookings))

	for _, booking := range bookings {
		apiBooking := api.LessonBookingDetail{
			Id:             booking.ID,
			UserId:         booking.UserID,
			InstructorId:   booking.InstructorID,
			InstructorName: booking.InstructorName,
			ServiceId:      booking.ServiceID,
			ServiceName:    booking.ServiceName,
			StartsAt:       booking.StartsAt,
			EndsAt:         booking.EndsAt,
			Status:         api.LessonBookingDetailStatus(booking.Status),
			TotalPrice:     booking.TotalPrice,
			CreatedAt:      booking.CreatedAt,
		}
		response = append(response, apiBooking)

	}

	return api.GetMyLessonBookings200JSONResponse(response), nil
}

// GetLessonBookingById implements [api.StrictServerInterface].
func (h *Handler) GetLessonBookingById(ctx context.Context, request api.GetLessonBookingByIdRequestObject) (api.GetLessonBookingByIdResponseObject, error) {

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetLessonBookingById401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.GetLessonBookingById403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	if request.Id <= 0 {
		return api.GetLessonBookingById400JSONResponse{
			Error:   "invalid_id",
			Message: "Некорректный ID бронирования",
		}, nil
	}

	booking, err := h.service.GetLessonBookingByID(
		ctx, userID, request.Id,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBookingInput):
			return api.GetLessonBookingById400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные данные запроса",
			}, nil

		case errors.Is(err, ErrBookingNotFound):
			return api.GetLessonBookingById404JSONResponse{
				Error:   "booking_not_found",
				Message: "Бронирование не найдено",
			}, nil

		case errors.Is(err, ErrBookingForbidden):
			return api.GetLessonBookingById403JSONResponse{
				Error:   "forbidden",
				Message: "Нет доступа к этому бронированию",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to get lesson booking",
				"booking_id", request.Id,
				"user_id", userID,
				"error", err,
			)

			return api.GetLessonBookingById500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := api.LessonBookingDetail{
		Id:             booking.ID,
		UserId:         booking.UserID,
		InstructorId:   booking.InstructorID,
		InstructorName: booking.InstructorName,
		ServiceId:      booking.ServiceID,
		ServiceName:    booking.ServiceName,
		StartsAt:       booking.StartsAt,
		EndsAt:         booking.EndsAt,
		Status:         api.LessonBookingDetailStatus(booking.Status),
		TotalPrice:     booking.TotalPrice,
		CreatedAt:      booking.CreatedAt,
	}

	return api.GetLessonBookingById200JSONResponse(response), nil
}

// CancelLessonBooking implements [api.StrictServerInterface].
func (h *Handler) CancelLessonBooking(ctx context.Context, request api.CancelLessonBookingRequestObject) (api.CancelLessonBookingResponseObject, error) {

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.CancelLessonBooking401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.CancelLessonBooking403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	if request.Id <= 0 {
		return api.CancelLessonBooking400JSONResponse{
			Error:   "invalid_id",
			Message: "Некорректный ID бронирования",
		}, nil
	}

	booking, err := h.service.CancelLessonBooking(
		ctx, userID, request.Id,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBookingInput):
			return api.CancelLessonBooking400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные данные запроса",
			}, nil

		case errors.Is(err, ErrBookingNotFound):
			return api.CancelLessonBooking404JSONResponse{
				Error:   "booking_not_found",
				Message: "Бронирование не найдено",
			}, nil

		case errors.Is(err, ErrBookingForbidden):
			return api.CancelLessonBooking403JSONResponse{
				Error:   "forbidden",
				Message: "Нельзя отменить чужое бронирование",
			}, nil

		case errors.Is(err, ErrBookingCannotCancel):
			return api.CancelLessonBooking409JSONResponse{
				Error:   "booking_cannot_cancel",
				Message: "Это бронирование больше нельзя отменить",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to cancel lesson booking",
				"booking_id", request.Id,
				"user_id", userID,
				"error", err,
			)

			return api.CancelLessonBooking500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := api.LessonBookingDetail{
		Id:             booking.ID,
		UserId:         booking.UserID,
		InstructorId:   booking.InstructorID,
		InstructorName: booking.InstructorName,
		ServiceId:      booking.ServiceID,
		ServiceName:    booking.ServiceName,
		StartsAt:       booking.StartsAt,
		EndsAt:         booking.EndsAt,
		Status:         api.LessonBookingDetailStatus(booking.Status),
		TotalPrice:     booking.TotalPrice,
		CreatedAt:      booking.CreatedAt,
	}

	slog.InfoContext(ctx,
		"lesson booking cancelled",
		"booking_id", booking.ID,
		"user_id", userID,
	)

	return api.CancelLessonBooking200JSONResponse(response), nil
}
