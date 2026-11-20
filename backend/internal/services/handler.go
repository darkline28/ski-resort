package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/services/api"
)

type Handler struct {
	service *ServiceService
}

func NewHandler(service *ServiceService) *Handler {
	return &Handler{service: service}
}

// ListServices implements [api.StrictServerInterface].
func (h *Handler) ListServices(ctx context.Context, request api.ListServicesRequestObject) (api.ListServicesResponseObject, error) {
	serverServices, err := h.service.ListServices(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list services", "error", err)

		return api.ListServices500JSONResponse{
			Error:   "internal_server_error",
			Message: "Внутренняя ошибка сервера",
		}, nil
	}

	response := make([]api.Service, 0, len(serverServices))

	for _, service := range serverServices {
		apiService := api.Service{
			Id:              service.ID,
			Description:     service.Description,
			DurationMinutes: service.DurationMinutes,
			IsActive:        service.IsActive,
			Name:            service.Name,
			Price:           service.Price,
		}
		response = append(response, apiService)
	}

	return api.ListServices200JSONResponse(response), nil

}

// GetServiceByID implements [api.StrictServerInterface].
func (h *Handler) GetServiceByID(ctx context.Context, request api.GetServiceByIDRequestObject) (api.GetServiceByIDResponseObject, error) {
	service, err := h.service.GetServiceByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, ErrInvalidServiceID) {
			return api.GetServiceByID400JSONResponse{
				Error:   "invalid_service_id",
				Message: "Некорректный ID услуги",
			}, nil
		}

		if errors.Is(err, ErrServiceNotFound) {
			return api.GetServiceByID404JSONResponse{
				Error:   "service_not_found",
				Message: "Услуга не найдена",
			}, nil
		}

		slog.ErrorContext(ctx, "failed to get service by ID",
			"service_id", request.Id,
			"error", err,
		)

		return api.GetServiceByID500JSONResponse{
			Error:   "internal_server_error",
			Message: "Внутренняя ошибка сервера",
		}, nil
	}

	apiService := api.Service{
		Id:              service.ID,
		Name:            service.Name,
		Description:     service.Description,
		DurationMinutes: service.DurationMinutes,
		Price:           service.Price,
		IsActive:        service.IsActive,
	}

	return api.GetServiceByID200JSONResponse(apiService), nil
}
