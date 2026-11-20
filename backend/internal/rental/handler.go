package rental

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/middleware"
	"github.com/darkline28/ski-resort/backend/internal/rental/api"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetEquipmentAvailability implements [api.StrictServerInterface].
func (h *Handler) GetEquipmentAvailability(ctx context.Context, request api.GetEquipmentAvailabilityRequestObject) (api.GetEquipmentAvailabilityResponseObject, error) {

	items, err := h.service.GetEquipmentAvailability(
		ctx,
		request.Params.StartsAt,
		request.Params.EndsAt,
		request.Params.EquipmentId,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRentalPeriod),
			errors.Is(err, ErrInvalidRentalInput):

			return api.GetEquipmentAvailability400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные параметры периода аренды",
			}, nil

		case errors.Is(err, ErrEquipmentNotFound):
			return api.GetEquipmentAvailability404JSONResponse{
				Error:   "equipment_not_found",
				Message: "Оборудование не найдено",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to get equipment availability",
				"error", err,
			)

			return api.GetEquipmentAvailability500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := make([]api.EquipmentAvailability, 0, len(items))

	for _, item := range items {
		apiItem := api.EquipmentAvailability{
			EquipmentId:       item.EquipmentID,
			Size:              item.Size,
			AvailableQuantity: item.AvailableQuantity,
		}

		response = append(response, apiItem)
	}

	return api.GetEquipmentAvailability200JSONResponse(response), nil
}

// CreateEquipmentBooking implements [api.StrictServerInterface].
func (h *Handler) CreateEquipmentBooking(ctx context.Context, request api.CreateEquipmentBookingRequestObject) (api.CreateEquipmentBookingResponseObject, error) {

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.CreateEquipmentBooking401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.CreateEquipmentBooking403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	if request.Body == nil {
		return api.CreateEquipmentBooking400JSONResponse{
			Error:   "invalid_input",
			Message: "Некорректные данные запроса",
		}, nil
	}

	params := CreateEquipmentBookingParams{
		StartsAt: request.Body.StartsAt,
		EndsAt:   request.Body.EndsAt,
		Items:    make([]RentalItemRequest, 0),
		Sets:     make([]RentalSetRequest, 0),
	}
	if request.Body.Items != nil {
		for _, item := range *request.Body.Items {
			params.Items = append(params.Items, RentalItemRequest{
				EquipmentID: item.EquipmentId,
				Size:        item.Size,
				Quantity:    item.Quantity,
			})
		}
	}
	if request.Body.Sets != nil {
		for _, set := range *request.Body.Sets {
			sizes := make([]RentalSetSize, 0, len(set.Sizes))

			for _, size := range set.Sizes {
				sizes = append(sizes, RentalSetSize{
					CategoryID: size.CategoryId,
					Size:       size.Size,
				})
			}

			params.Sets = append(params.Sets, RentalSetRequest{
				EquipmentSetID: set.EquipmentSetId,
				Quantity:       set.Quantity,
				Sizes:          sizes,
			})
		}
	}

	booking, err := h.service.CreateEquipmentBooking(
		ctx, userID, params,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRentalInput),
			errors.Is(err, ErrInvalidRentalPeriod),
			errors.Is(err, ErrInvalidSetSizes):

			return api.CreateEquipmentBooking400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные данные бронирования",
			}, nil

		case errors.Is(err, ErrEquipmentNotFound),
			errors.Is(err, ErrEquipmentSetNotFound):

			return api.CreateEquipmentBooking404JSONResponse{
				Error:   "not_found",
				Message: "Оборудование или комплект не найдены",
			}, nil

		case errors.Is(err, ErrEquipmentSetInactive):
			return api.CreateEquipmentBooking422JSONResponse{
				Error:   "equipment_inactive",
				Message: "Выбранный комплект недоступен",
			}, nil

		case errors.Is(err, ErrEquipmentUnavailable):
			return api.CreateEquipmentBooking409JSONResponse{
				Error:   "equipment_unavailable",
				Message: "Выбранное оборудование недоступно на указанный период",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to create equipment booking",
				"user_id", userID,
				"error", err,
			)

			return api.CreateEquipmentBooking500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := api.CreateEquipmentBooking201JSONResponse{
		Id:         booking.ID,
		UserId:     booking.UserID,
		StartsAt:   booking.StartsAt,
		EndsAt:     booking.EndsAt,
		Status:     api.EquipmentRentalStatus(booking.Status),
		TotalPrice: booking.TotalPrice,
		CreatedAt:  booking.CreatedAt,
		Items:      make([]api.EquipmentRentalItem, 0, len(booking.Items)),
		Sets:       make([]api.EquipmentRentalSet, 0, len(booking.Sets)),
	}

	for _, item := range booking.Items {
		response.Items = append(response.Items, api.EquipmentRentalItem{
			EquipmentUnitId: item.EquipmentUnitID,
			EquipmentId:     item.EquipmentID,
			Name:            item.Name,
			Size:            item.Size,
			BookedSetId:     item.BookedSetID,
			ItemTotalPrice:  item.ItemTotalPrice,
		})
	}

	for _, set := range booking.Sets {
		response.Sets = append(response.Sets, api.EquipmentRentalSet{
			Id:             set.ID,
			EquipmentSetId: set.EquipmentSetID,
			Name:           set.Name,
			TotalPrice:     set.TotalPrice,
		})
	}

	slog.InfoContext(ctx,
		"equipment booking created",
		"booking_id", booking.ID,
		"user_id", userID,
	)

	return response, nil
}

// GetMyEquipmentBookings implements [api.StrictServerInterface].
func (h *Handler) GetMyEquipmentBookings(ctx context.Context, request api.GetMyEquipmentBookingsRequestObject) (api.GetMyEquipmentBookingsResponseObject, error) {

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetMyEquipmentBookings401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.GetMyEquipmentBookings403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	var status *string

	if request.Params.Status != nil {
		value := string(*request.Params.Status)
		status = &value
	}

	bookings, err := h.service.GetMyEquipmentBookings(
		ctx, userID, status,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidRentalStatus) ||
			errors.Is(err, ErrInvalidRentalInput) {

			return api.GetMyEquipmentBookings400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные параметры запроса",
			}, nil
		}

		slog.ErrorContext(ctx,
			"failed to get my equipment bookings",
			"user_id", userID,
			"error", err,
		)

		return api.GetMyEquipmentBookings500JSONResponse{
			Error:   "internal_error",
			Message: "Внутренняя ошибка сервера",
		}, nil
	}

	response := make(
		[]api.EquipmentRental,
		0,
		len(bookings),
	)

	for _, booking := range bookings {
		apiBooking := api.EquipmentRental{
			Id:         booking.ID,
			UserId:     booking.UserID,
			StartsAt:   booking.StartsAt,
			EndsAt:     booking.EndsAt,
			Status:     api.EquipmentRentalStatus(booking.Status),
			TotalPrice: booking.TotalPrice,
			CreatedAt:  booking.CreatedAt,
			Items:      make([]api.EquipmentRentalItem, 0, len(booking.Items)),
			Sets:       make([]api.EquipmentRentalSet, 0, len(booking.Sets)),
		}

		for _, item := range booking.Items {
			apiBooking.Items = append(
				apiBooking.Items,
				api.EquipmentRentalItem{
					EquipmentUnitId: item.EquipmentUnitID,
					EquipmentId:     item.EquipmentID,
					Name:            item.Name,
					Size:            item.Size,
					BookedSetId:     item.BookedSetID,
					ItemTotalPrice:  item.ItemTotalPrice,
				},
			)
		}

		for _, set := range booking.Sets {
			apiBooking.Sets = append(
				apiBooking.Sets,
				api.EquipmentRentalSet{
					Id:             set.ID,
					EquipmentSetId: set.EquipmentSetID,
					Name:           set.Name,
					TotalPrice:     set.TotalPrice,
				},
			)
		}

		response = append(response, apiBooking)
	}

	return api.GetMyEquipmentBookings200JSONResponse(response), nil
}

// GetEquipmentBookingById implements [api.StrictServerInterface].
func (h *Handler) GetEquipmentBookingById(ctx context.Context, request api.GetEquipmentBookingByIdRequestObject) (api.GetEquipmentBookingByIdResponseObject, error) {

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetEquipmentBookingById401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.GetEquipmentBookingById403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	if request.Id <= 0 {
		return api.GetEquipmentBookingById400JSONResponse{
			Error:   "invalid_id",
			Message: "Некорректный ID бронирования",
		}, nil
	}

	booking, err := h.service.GetEquipmentBookingByID(
		ctx, userID, request.Id,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRentalInput):
			return api.GetEquipmentBookingById400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные данные запроса",
			}, nil

		case errors.Is(err, ErrRentalNotFound):
			return api.GetEquipmentBookingById404JSONResponse{
				Error:   "booking_not_found",
				Message: "Бронирование не найдено",
			}, nil

		case errors.Is(err, ErrRentalForbidden):
			return api.GetEquipmentBookingById403JSONResponse{
				Error:   "forbidden",
				Message: "Нет доступа к этому бронированию",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to get equipment booking",
				"booking_id", request.Id,
				"user_id", userID,
				"error", err,
			)

			return api.GetEquipmentBookingById500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := api.GetEquipmentBookingById200JSONResponse{
		Id:         booking.ID,
		UserId:     booking.UserID,
		StartsAt:   booking.StartsAt,
		EndsAt:     booking.EndsAt,
		Status:     api.EquipmentRentalStatus(booking.Status),
		TotalPrice: booking.TotalPrice,
		CreatedAt:  booking.CreatedAt,
		Items:      make([]api.EquipmentRentalItem, 0, len(booking.Items)),
		Sets:       make([]api.EquipmentRentalSet, 0, len(booking.Sets)),
	}

	for _, item := range booking.Items {
		response.Items = append(response.Items, api.EquipmentRentalItem{
			EquipmentUnitId: item.EquipmentUnitID,
			EquipmentId:     item.EquipmentID,
			Name:            item.Name,
			Size:            item.Size,
			BookedSetId:     item.BookedSetID,
			ItemTotalPrice:  item.ItemTotalPrice,
		})
	}

	for _, set := range booking.Sets {
		response.Sets = append(response.Sets, api.EquipmentRentalSet{
			Id:             set.ID,
			EquipmentSetId: set.EquipmentSetID,
			Name:           set.Name,
			TotalPrice:     set.TotalPrice,
		})
	}

	return response, nil
}

// CancelEquipmentBooking implements [api.StrictServerInterface].
func (h *Handler) CancelEquipmentBooking(ctx context.Context, request api.CancelEquipmentBookingRequestObject) (api.CancelEquipmentBookingResponseObject, error) {

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.CancelEquipmentBooking401JSONResponse{
			Error:   "unauthorized",
			Message: "Пользователь не авторизован",
		}, nil
	}

	role, ok := middleware.RoleFromContext(ctx)
	if !ok || role != "client" {
		return api.CancelEquipmentBooking403JSONResponse{
			Error:   "forbidden",
			Message: "Недостаточно прав",
		}, nil
	}

	if request.Id <= 0 {
		return api.CancelEquipmentBooking400JSONResponse{
			Error:   "invalid_id",
			Message: "Некорректный ID бронирования",
		}, nil
	}

	booking, err := h.service.CancelEquipmentBooking(
		ctx, userID, request.Id,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRentalInput):
			return api.CancelEquipmentBooking400JSONResponse{
				Error:   "invalid_input",
				Message: "Некорректные данные запроса",
			}, nil

		case errors.Is(err, ErrRentalNotFound):
			return api.CancelEquipmentBooking404JSONResponse{
				Error:   "booking_not_found",
				Message: "Бронирование не найдено",
			}, nil

		case errors.Is(err, ErrRentalForbidden):
			return api.CancelEquipmentBooking403JSONResponse{
				Error:   "forbidden",
				Message: "Нельзя отменить чужое бронирование",
			}, nil

		case errors.Is(err, ErrRentalCannotCancel):
			return api.CancelEquipmentBooking409JSONResponse{
				Error:   "invalid_status_transition",
				Message: "Изменение статуса недоступно",
			}, nil

		default:
			slog.ErrorContext(ctx,
				"failed to cancel equipment booking",
				"booking_id", request.Id,
				"user_id", userID,
				"error", err,
			)

			return api.CancelEquipmentBooking500JSONResponse{
				Error:   "internal_error",
				Message: "Внутренняя ошибка сервера",
			}, nil
		}
	}

	response := api.CancelEquipmentBooking200JSONResponse{
		Id:         booking.ID,
		UserId:     booking.UserID,
		StartsAt:   booking.StartsAt,
		EndsAt:     booking.EndsAt,
		Status:     api.EquipmentRentalStatus(booking.Status),
		TotalPrice: booking.TotalPrice,
		CreatedAt:  booking.CreatedAt,
		Items:      make([]api.EquipmentRentalItem, 0, len(booking.Items)),
		Sets:       make([]api.EquipmentRentalSet, 0, len(booking.Sets)),
	}

	for _, item := range booking.Items {
		response.Items = append(response.Items, api.EquipmentRentalItem{
			EquipmentUnitId: item.EquipmentUnitID,
			EquipmentId:     item.EquipmentID,
			Name:            item.Name,
			Size:            item.Size,
			BookedSetId:     item.BookedSetID,
			ItemTotalPrice:  item.ItemTotalPrice,
		})
	}

	for _, set := range booking.Sets {
		response.Sets = append(response.Sets, api.EquipmentRentalSet{
			Id:             set.ID,
			EquipmentSetId: set.EquipmentSetID,
			Name:           set.Name,
			TotalPrice:     set.TotalPrice,
		})
	}

	slog.InfoContext(ctx,
		"equipment booking cancelled",
		"booking_id", booking.ID,
		"user_id", userID,
	)

	return response, nil
}
