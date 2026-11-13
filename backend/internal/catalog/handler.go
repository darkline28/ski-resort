package catalog

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/catalog/api"
)

type Handler struct {
	service *EquipmentService
}

func NewHandler(service *EquipmentService) *Handler {
	return &Handler{service: service}
}

// ListEquipment implements [api.StrictServerInterface].
func (h *Handler) ListEquipment(ctx context.Context, request api.ListEquipmentRequestObject) (api.ListEquipmentResponseObject, error) {
	serviceEquipments, err := h.service.GetListEquipments(ctx, request.Params.Category)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to list equipment",
			"error", err,
		)
		return api.ListEquipment500JSONResponse{Error: "500", Message: "Internal Server Error"}, nil
	}
	slog.InfoContext(ctx,
		"list equipment",
		"count", len(serviceEquipments))

	apiEquipments := make([]api.Equipment, 0, len(serviceEquipments))
	for _, eq := range serviceEquipments {
		equipment := api.Equipment{
			Id:           eq.ID,
			Name:         eq.Name,
			Category:     eq.Category,
			Description:  eq.Description,
			ImageUrl:     eq.ImageURL,
			PricePerHour: eq.PricePerHour,
			Quantity:     eq.Quantity,
		}
		apiEquipments = append(apiEquipments, equipment)
	}
	return api.ListEquipment200JSONResponse(apiEquipments), nil
}

// GetEquipmentByID implements [api.StrictServerInterface].
func (h *Handler) GetEquipmentByID(ctx context.Context, request api.GetEquipmentByIDRequestObject) (api.GetEquipmentByIDResponseObject, error) {
	equipment, err := h.service.GetEquipmentByID(ctx, request.Id)

	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			return api.GetEquipmentByID400JSONResponse{Error: "400", Message: "invalid id"}, nil
		}
		if errors.Is(err, ErrEquipmentNotFound) {
			return api.GetEquipmentByID404JSONResponse{Error: "404", Message: "id not found"}, nil
		}
		slog.ErrorContext(
			ctx,
			"failed to get equipment by id",
			"equipment_id", request.Id,
			"error", err,
		)
		return api.GetEquipmentByID500JSONResponse{Error: "500", Message: "internal server error"}, nil
	}

	var equipmentRes = api.Equipment{
		Id:           equipment.ID,
		Name:         equipment.Name,
		Category:     equipment.Category,
		Quantity:     equipment.Quantity,
		PricePerHour: equipment.PricePerHour,
		Description:  equipment.Description,
		ImageUrl:     equipment.ImageURL,
	}
	return api.GetEquipmentByID200JSONResponse(equipmentRes), nil

}

// GetEquipmentCategories implements [api.StrictServerInterface].
func (h *Handler) GetEquipmentCategories(ctx context.Context, request api.GetEquipmentCategoriesRequestObject) (api.GetEquipmentCategoriesResponseObject, error) {
	serviceCategories, err := h.service.GetEquipmentCategories(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get equipment categories", "error", err)
		return api.GetEquipmentCategories500JSONResponse{Error: "500", Message: "internal server error"}, nil
	}

	apiCategories := make([]api.EquipmentCategory, 0, len(serviceCategories))

	for _, ctgr := range serviceCategories {
		category := api.EquipmentCategory{
			Id:   ctgr.ID,
			Name: ctgr.Name,
		}
		apiCategories = append(apiCategories, category)
	}
	return api.GetEquipmentCategories200JSONResponse(apiCategories), nil
}

// GetEquipmentSetByID implements [api.StrictServerInterface].
func (h *Handler) GetEquipmentSetByID(ctx context.Context, request api.GetEquipmentSetByIDRequestObject) (api.GetEquipmentSetByIDResponseObject, error) {
	serviceEquipmentSet, err := h.service.GetEquipmentSetByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			return api.GetEquipmentSetByID400JSONResponse{Error: "400", Message: "invalid id"}, nil
		}
		if errors.Is(err, ErrEquipmentSetNotFound) {
			return api.GetEquipmentSetByID404JSONResponse{Error: "404", Message: "set not found"}, nil
		}
		slog.ErrorContext(ctx, "failed to get equipment sets", "error", err)
		return api.GetEquipmentSetByID500JSONResponse{Error: "500", Message: "internal server error"}, nil
	}

	apiSetItems := make([]api.EquipmentSetItem, 0, len(serviceEquipmentSet.Items))
	for _, it := range serviceEquipmentSet.Items {
		item := api.EquipmentSetItem{
			CategoryId:   it.CategoryID,
			CategoryName: it.Category,
			Quantity:     it.Quantity,
		}
		apiSetItems = append(apiSetItems, item)
	}
	set := api.EquipmentSetDetail{
		Id:           serviceEquipmentSet.ID,
		Name:         serviceEquipmentSet.Name,
		Description:  serviceEquipmentSet.Description,
		SportType:    api.EquipmentSetDetailSportType(serviceEquipmentSet.SportType),
		AgeGroup:     api.EquipmentSetDetailAgeGroup(serviceEquipmentSet.AgeGroup),
		PricePerHour: serviceEquipmentSet.PricePerHour,
		IsActive:     serviceEquipmentSet.IsActive,
		Items:        apiSetItems,
	}
	return api.GetEquipmentSetByID200JSONResponse(set), nil
}

// GetEquipmentSets implements [api.StrictServerInterface].
func (h *Handler) GetEquipmentSets(ctx context.Context, request api.GetEquipmentSetsRequestObject) (api.GetEquipmentSetsResponseObject, error) {
	serviceSets, err := h.service.GetEquipmentSets(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get equipment sets", "error", err)
		return api.GetEquipmentSets500JSONResponse{}, nil
	}

	apiSets := make([]api.EquipmentSet, 0, len(serviceSets))

	for _, st := range serviceSets {
		set := api.EquipmentSet{
			Id:           st.ID,
			Name:         st.Name,
			AgeGroup:     api.EquipmentSetAgeGroup(st.AgeGroup),
			SportType:    api.EquipmentSetSportType(st.SportType),
			Description:  st.Description,
			PricePerHour: st.PricePerHour,
			IsActive:     st.IsActive,
		}
		apiSets = append(apiSets, set)
	}
	return api.GetEquipmentSets200JSONResponse(apiSets), nil

}
