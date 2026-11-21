package admin

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/admin/api"
	"github.com/darkline28/ski-resort/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func isAdmin(ctx context.Context) bool {
	role, ok := middleware.RoleFromContext(ctx)
	return ok && role == "admin"
}

func apiError(err error) api.Error {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
		return api.Error{Error: "invalid_input", Message: "Некорректные данные запроса"}
	case errors.Is(err, ErrNotFound):
		return api.Error{Error: "not_found", Message: "Ресурс не найден"}
	case errors.Is(err, ErrConflict), errors.Is(err, ErrInvalidTransition):
		return api.Error{Error: "conflict", Message: "Операция конфликтует с текущим состоянием данных"}
	default:
		slog.Error("admin operation failed", "error", err)
		return api.Error{Error: "internal_error", Message: "Внутренняя ошибка сервера"}
	}
}

func toAPIService(s *ServiceModel) api.AdminService {
	return api.AdminService{
		Id:              s.ID,
		Name:            s.Name,
		Description:     s.Description,
		DurationMinutes: s.DurationMinutes,
		Price:           s.Price,
		IsActive:        s.IsActive,
	}
}

func toAPIInstructor(i *Instructor) api.AdminInstructor {
	return api.AdminInstructor{
		Id:              i.ID,
		Name:            i.Name,
		Specialization:  i.Specialization,
		Description:     i.Description,
		ExperienceYears: i.ExperienceYears,
		ImageUrl:        i.ImageURL,
		IsActive:        i.IsActive,
	}
}

func toAPISchedule(s *InstructorSchedule) api.AdminSchedule {
	return api.AdminSchedule{
		Id:       s.ID,
		StartsAt: s.StartsAt,
		EndsAt:   s.EndsAt,
	}
}

func toAPILessonBooking(b *AdminLessonBooking) api.AdminLessonBooking {
	return api.AdminLessonBooking{
		Id:             b.ID,
		UserId:         b.UserID,
		ClientName:     b.ClientName,
		InstructorId:   b.InstructorID,
		InstructorName: b.InstructorName,
		ServiceId:      b.ServiceID,
		ServiceName:    b.ServiceName,
		StartsAt:       b.StartsAt,
		EndsAt:         b.EndsAt,
		Status:         api.AdminLessonStatus(b.Status),
		TotalPrice:     b.TotalPrice,
		CreatedAt:      b.CreatedAt,
	}
}

func toAPIRental(b *AdminEquipmentRental) api.AdminEquipmentRental {
	items := make([]api.AdminRentalItem, 0, len(b.Items))
	for _, item := range b.Items {
		items = append(items, api.AdminRentalItem{
			EquipmentUnitId: item.EquipmentUnitID,
			EquipmentId:     item.EquipmentID,
			Name:            item.Name,
			Size:            item.Size,
			BookedSetId:     item.BookedSetID,
			ItemTotalPrice:  item.ItemTotalPrice,
		})
	}

	sets := make([]api.AdminBookedRentalSet, 0, len(b.Sets))
	for _, set := range b.Sets {
		sets = append(sets, api.AdminBookedRentalSet{
			Id:             set.ID,
			EquipmentSetId: set.EquipmentSetID,
			Name:           set.Name,
			TotalPrice:     set.TotalPrice,
		})
	}

	return api.AdminEquipmentRental{
		Id:         b.ID,
		UserId:     b.UserID,
		ClientName: b.ClientName,
		StartsAt:   b.StartsAt,
		EndsAt:     b.EndsAt,
		Status:     api.AdminRentalStatus(b.Status),
		TotalPrice: b.TotalPrice,
		CreatedAt:  b.CreatedAt,
		Items:      items,
		Sets:       sets,
	}
}

func toAPIEquipment(e *Equipment) api.AdminEquipment {
	return api.AdminEquipment{
		Id:           e.ID,
		CategoryId:   e.CategoryID,
		Name:         e.Name,
		Description:  e.Description,
		ImageUrl:     e.ImageURL,
		PricePerHour: e.PricePerHour,
	}
}

func toAPIEquipmentUnit(u *EquipmentUnit) api.AdminEquipmentUnit {
	return api.AdminEquipmentUnit{
		Id:              u.ID,
		EquipmentId:     u.EquipmentID,
		InventoryNumber: u.InventoryNumber,
		Size:            u.Size,
		Status:          api.AdminUnitStatus(u.Status),
		CreatedAt:       u.CreatedAt,
	}
}

func toAPIEquipmentSet(s *EquipmentSet) api.AdminEquipmentSetDetail {
	items := make([]api.AdminEquipmentSetItem, 0, len(s.Items))
	for _, item := range s.Items {
		items = append(items, api.AdminEquipmentSetItem{
			CategoryId:   item.CategoryID,
			CategoryName: item.CategoryName,
			Quantity:     item.Quantity,
		})
	}

	return api.AdminEquipmentSetDetail{
		Id:           s.ID,
		Name:         s.Name,
		Description:  s.Description,
		SportType:    api.AdminEquipmentSetDetailSportType(s.SportType),
		AgeGroup:     api.AdminEquipmentSetDetailAgeGroup(s.AgeGroup),
		PricePerHour: s.PricePerHour,
		IsActive:     s.IsActive,
		Items:        items,
	}
}

func toSetItems(items []api.AdminEquipmentSetItemRequest) []EquipmentSetItemParams {
	result := make([]EquipmentSetItemParams, 0, len(items))
	for _, item := range items {
		result = append(result, EquipmentSetItemParams{
			CategoryID: item.CategoryId,
			Quantity:   item.Quantity,
		})
	}
	return result
}

// CLIENTS

func (h *Handler) GetAdminClients(ctx context.Context, request api.GetAdminClientsRequestObject) (api.GetAdminClientsResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminClients403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	clients, err := h.service.GetClients(ctx)
	if err != nil {
		return api.GetAdminClients500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
	}

	result := make([]api.AdminClient, 0, len(clients))
	for _, client := range clients {
		result = append(result, api.AdminClient{Id: client.ID, Name: client.Name, Email: client.Email, Phone: client.Phone})
	}
	return api.GetAdminClients200JSONResponse(result), nil
}

// SERVICES

func (h *Handler) GetAdminServices(ctx context.Context, request api.GetAdminServicesRequestObject) (api.GetAdminServicesResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminServices403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	services, err := h.service.GetServices(ctx)
	if err != nil {
		return api.GetAdminServices500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
	}

	result := make([]api.AdminService, 0, len(services))
	for _, service := range services {
		result = append(result, toAPIService(&service))
	}
	return api.GetAdminServices200JSONResponse(result), nil
}

func (h *Handler) GetAdminService(ctx context.Context, request api.GetAdminServiceRequestObject) (api.GetAdminServiceResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminService403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	service, err := h.service.GetService(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.GetAdminService404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminService500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.GetAdminService200JSONResponse(toAPIService(service)), nil
}

func (h *Handler) CreateAdminService(ctx context.Context, request api.CreateAdminServiceRequestObject) (api.CreateAdminServiceResponseObject, error) {
	if !isAdmin(ctx) {
		return api.CreateAdminService403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.CreateAdminService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	active := true
	if request.Body.IsActive != nil {
		active = *request.Body.IsActive
	}

	service, err := h.service.CreateService(ctx, CreateServiceParams{
		Name: request.Body.Name, Description: request.Body.Description,
		DurationMinutes: request.Body.DurationMinutes, Price: request.Body.Price, IsActive: active,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.CreateAdminService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.CreateAdminService409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.CreateAdminService500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.CreateAdminService201JSONResponse(toAPIService(service)), nil
}

func (h *Handler) UpdateAdminService(ctx context.Context, request api.UpdateAdminServiceRequestObject) (api.UpdateAdminServiceResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminService403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	service, err := h.service.UpdateService(ctx, request.Id, UpdateServiceParams{
		Name: request.Body.Name, Description: request.Body.Description,
		DurationMinutes: request.Body.DurationMinutes, Price: request.Body.Price, IsActive: request.Body.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.UpdateAdminService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminService404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.UpdateAdminService409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminService500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminService200JSONResponse(toAPIService(service)), nil
}

func (h *Handler) DeleteAdminService(ctx context.Context, request api.DeleteAdminServiceRequestObject) (api.DeleteAdminServiceResponseObject, error) {
	if !isAdmin(ctx) {
		return api.DeleteAdminService403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	err := h.service.DeleteService(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.DeleteAdminService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.DeleteAdminService404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.DeleteAdminService500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.DeleteAdminService204Response{}, nil
}

// INSTRUCTORS

func (h *Handler) GetAdminInstructors(ctx context.Context, request api.GetAdminInstructorsRequestObject) (api.GetAdminInstructorsResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminInstructors403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	instructors, err := h.service.GetInstructors(ctx)
	if err != nil {
		return api.GetAdminInstructors500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
	}

	result := make([]api.AdminInstructor, 0, len(instructors))
	for _, instructor := range instructors {
		result = append(result, toAPIInstructor(&instructor))
	}
	return api.GetAdminInstructors200JSONResponse(result), nil
}

func (h *Handler) CreateAdminInstructor(ctx context.Context, request api.CreateAdminInstructorRequestObject) (api.CreateAdminInstructorResponseObject, error) {
	if !isAdmin(ctx) {
		return api.CreateAdminInstructor403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.CreateAdminInstructor400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	active := true
	if request.Body.IsActive != nil {
		active = *request.Body.IsActive
	}

	instructor, err := h.service.CreateInstructor(ctx, CreateInstructorParams{
		UserID: request.Body.UserId, Specialization: request.Body.Specialization,
		Description: request.Body.Description, ExperienceYears: request.Body.ExperienceYears,
		ImageURL: request.Body.ImageUrl, IsActive: active,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.CreateAdminInstructor400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrNotFound):
			return api.CreateAdminInstructor404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.CreateAdminInstructor409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.CreateAdminInstructor500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.CreateAdminInstructor201JSONResponse(toAPIInstructor(instructor)), nil
}

func (h *Handler) GetAdminInstructor(ctx context.Context, request api.GetAdminInstructorRequestObject) (api.GetAdminInstructorResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminInstructor403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	instructor, err := h.service.GetInstructor(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminInstructor400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.GetAdminInstructor404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminInstructor500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.GetAdminInstructor200JSONResponse(toAPIInstructor(instructor)), nil
}

func (h *Handler) UpdateAdminInstructor(ctx context.Context, request api.UpdateAdminInstructorRequestObject) (api.UpdateAdminInstructorResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminInstructor403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminInstructor400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	instructor, err := h.service.UpdateInstructor(ctx, request.Id, UpdateInstructorParams{
		Specialization: request.Body.Specialization, Description: request.Body.Description,
		ExperienceYears: request.Body.ExperienceYears, ImageURL: request.Body.ImageUrl,
		IsActive: request.Body.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.UpdateAdminInstructor400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminInstructor404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.UpdateAdminInstructor409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminInstructor500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminInstructor200JSONResponse(toAPIInstructor(instructor)), nil
}

func (h *Handler) DeleteAdminInstructor(ctx context.Context, request api.DeleteAdminInstructorRequestObject) (api.DeleteAdminInstructorResponseObject, error) {
	if !isAdmin(ctx) {
		return api.DeleteAdminInstructor403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	err := h.service.DeleteInstructor(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.DeleteAdminInstructor400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.DeleteAdminInstructor404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.DeleteAdminInstructor500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.DeleteAdminInstructor204Response{}, nil
}

// INSTRUCTOR SERVICES

func (h *Handler) GetAdminInstructorServices(ctx context.Context, request api.GetAdminInstructorServicesRequestObject) (api.GetAdminInstructorServicesResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminInstructorServices403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	services, err := h.service.GetInstructorServices(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminInstructorServices400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.GetAdminInstructorServices404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminInstructorServices500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}

	result := make([]api.AdminService, 0, len(services))
	for _, service := range services {
		result = append(result, toAPIService(&service))
	}
	return api.GetAdminInstructorServices200JSONResponse(result), nil
}

func (h *Handler) AssignAdminInstructorService(ctx context.Context, request api.AssignAdminInstructorServiceRequestObject) (api.AssignAdminInstructorServiceResponseObject, error) {
	if !isAdmin(ctx) {
		return api.AssignAdminInstructorService403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.AssignAdminInstructorService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	if err := h.service.AssignInstructorService(ctx, request.Id, request.Body.ServiceId); err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.AssignAdminInstructorService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound), errors.Is(err, ErrInvalidReference):
			return api.AssignAdminInstructorService404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.AssignAdminInstructorService409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.AssignAdminInstructorService500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.AssignAdminInstructorService204Response{}, nil
}

func (h *Handler) RemoveAdminInstructorService(ctx context.Context, request api.RemoveAdminInstructorServiceRequestObject) (api.RemoveAdminInstructorServiceResponseObject, error) {
	if !isAdmin(ctx) {
		return api.RemoveAdminInstructorService403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	if err := h.service.RemoveInstructorService(ctx, request.Id, request.ServiceId); err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.RemoveAdminInstructorService400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.RemoveAdminInstructorService404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.RemoveAdminInstructorService500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.RemoveAdminInstructorService204Response{}, nil
}

// INSTRUCTOR SCHEDULES

func (h *Handler) GetAdminInstructorSchedules(ctx context.Context, request api.GetAdminInstructorSchedulesRequestObject) (api.GetAdminInstructorSchedulesResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminInstructorSchedules403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	filter := ScheduleFilter{}
	if request.Params.DateFrom != nil {
		date := request.Params.DateFrom.Time
		filter.DateFrom = &date
	}
	if request.Params.DateTo != nil {
		date := request.Params.DateTo.Time
		filter.DateTo = &date
	}

	schedules, err := h.service.GetInstructorSchedules(ctx, request.Id, filter)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminInstructorSchedules400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.GetAdminInstructorSchedules404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminInstructorSchedules500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}

	result := make([]api.AdminSchedule, 0, len(schedules))
	for _, schedule := range schedules {
		result = append(result, toAPISchedule(&schedule))
	}
	return api.GetAdminInstructorSchedules200JSONResponse(result), nil
}

func (h *Handler) CreateAdminInstructorSchedule(ctx context.Context, request api.CreateAdminInstructorScheduleRequestObject) (api.CreateAdminInstructorScheduleResponseObject, error) {
	if !isAdmin(ctx) {
		return api.CreateAdminInstructorSchedule403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.CreateAdminInstructorSchedule400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	schedule, err := h.service.CreateInstructorSchedule(ctx, request.Id, CreateScheduleParams{
		StartsAt: request.Body.StartsAt,
		EndsAt:   request.Body.EndsAt,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.CreateAdminInstructorSchedule400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrNotFound):
			return api.CreateAdminInstructorSchedule404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.CreateAdminInstructorSchedule409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.CreateAdminInstructorSchedule500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.CreateAdminInstructorSchedule201JSONResponse(toAPISchedule(schedule)), nil
}

func (h *Handler) UpdateAdminInstructorSchedule(ctx context.Context, request api.UpdateAdminInstructorScheduleRequestObject) (api.UpdateAdminInstructorScheduleResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminInstructorSchedule403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminInstructorSchedule400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	schedule, err := h.service.UpdateInstructorSchedule(ctx, request.Id, request.ScheduleId, UpdateScheduleParams{
		StartsAt: request.Body.StartsAt,
		EndsAt:   request.Body.EndsAt,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.UpdateAdminInstructorSchedule400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminInstructorSchedule404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.UpdateAdminInstructorSchedule409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminInstructorSchedule500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminInstructorSchedule200JSONResponse(toAPISchedule(schedule)), nil
}

func (h *Handler) DeleteAdminInstructorSchedule(ctx context.Context, request api.DeleteAdminInstructorScheduleRequestObject) (api.DeleteAdminInstructorScheduleResponseObject, error) {
	if !isAdmin(ctx) {
		return api.DeleteAdminInstructorSchedule403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	err := h.service.DeleteInstructorSchedule(ctx, request.Id, request.ScheduleId)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.DeleteAdminInstructorSchedule400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.DeleteAdminInstructorSchedule404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.DeleteAdminInstructorSchedule409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.DeleteAdminInstructorSchedule500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.DeleteAdminInstructorSchedule204Response{}, nil
}

// LESSON BOOKINGS

func (h *Handler) GetAdminLessonBookings(ctx context.Context, request api.GetAdminLessonBookingsRequestObject) (api.GetAdminLessonBookingsResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminLessonBookings403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	filter := LessonBookingFilter{
		UserID:       request.Params.UserId,
		InstructorID: request.Params.InstructorId,
	}
	if request.Params.Status != nil {
		status := string(*request.Params.Status)
		filter.Status = &status
	}
	if request.Params.Date != nil {
		date := request.Params.Date.Time
		filter.Date = &date
	}

	bookings, err := h.service.GetLessonBookings(ctx, filter)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminLessonBookings400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminLessonBookings500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}

	result := make([]api.AdminLessonBooking, 0, len(bookings))
	for _, booking := range bookings {
		result = append(result, toAPILessonBooking(&booking))
	}
	return api.GetAdminLessonBookings200JSONResponse(result), nil
}

func (h *Handler) GetAdminLessonBookingById(ctx context.Context, request api.GetAdminLessonBookingByIdRequestObject) (api.GetAdminLessonBookingByIdResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminLessonBookingById403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	booking, err := h.service.GetLessonBookingByID(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminLessonBookingById400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.GetAdminLessonBookingById404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminLessonBookingById500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.GetAdminLessonBookingById200JSONResponse(toAPILessonBooking(booking)), nil
}

func (h *Handler) UpdateAdminLessonBooking(ctx context.Context, request api.UpdateAdminLessonBookingRequestObject) (api.UpdateAdminLessonBookingResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminLessonBooking403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminLessonBooking400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	booking, err := h.service.UpdateLessonBookingStatus(ctx, request.Id, string(request.Body.Status))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.UpdateAdminLessonBooking400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminLessonBooking404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidTransition), errors.Is(err, ErrConflict):
			return api.UpdateAdminLessonBooking409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminLessonBooking500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminLessonBooking200JSONResponse(toAPILessonBooking(booking)), nil
}

// EQUIPMENT BOOKINGS

func (h *Handler) GetAdminEquipmentBookings(ctx context.Context, request api.GetAdminEquipmentBookingsRequestObject) (api.GetAdminEquipmentBookingsResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminEquipmentBookings403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	filter := EquipmentBookingFilter{UserID: request.Params.UserId}
	if request.Params.Status != nil {
		status := string(*request.Params.Status)
		filter.Status = &status
	}
	if request.Params.Date != nil {
		date := request.Params.Date.Time
		filter.Date = &date
	}

	bookings, err := h.service.GetEquipmentBookings(ctx, filter)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminEquipmentBookings400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminEquipmentBookings500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}

	result := make([]api.AdminEquipmentRental, 0, len(bookings))
	for _, booking := range bookings {
		result = append(result, toAPIRental(&booking))
	}
	return api.GetAdminEquipmentBookings200JSONResponse(result), nil
}

func (h *Handler) GetAdminEquipmentBookingById(ctx context.Context, request api.GetAdminEquipmentBookingByIdRequestObject) (api.GetAdminEquipmentBookingByIdResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminEquipmentBookingById403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	booking, err := h.service.GetEquipmentBookingByID(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminEquipmentBookingById400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.GetAdminEquipmentBookingById404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminEquipmentBookingById500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.GetAdminEquipmentBookingById200JSONResponse(toAPIRental(booking)), nil
}

func (h *Handler) UpdateAdminEquipmentBooking(ctx context.Context, request api.UpdateAdminEquipmentBookingRequestObject) (api.UpdateAdminEquipmentBookingResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminEquipmentBooking403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminEquipmentBooking400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	booking, err := h.service.UpdateEquipmentBookingStatus(ctx, request.Id, string(request.Body.Status))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.UpdateAdminEquipmentBooking400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminEquipmentBooking404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidTransition), errors.Is(err, ErrConflict):
			return api.UpdateAdminEquipmentBooking409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminEquipmentBooking500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminEquipmentBooking200JSONResponse(toAPIRental(booking)), nil
}

// EQUIPMENT

func (h *Handler) CreateAdminEquipment(ctx context.Context, request api.CreateAdminEquipmentRequestObject) (api.CreateAdminEquipmentResponseObject, error) {
	if !isAdmin(ctx) {
		return api.CreateAdminEquipment403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.CreateAdminEquipment400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	equipment, err := h.service.CreateEquipment(ctx, CreateEquipmentParams{
		CategoryID:   request.Body.CategoryId,
		Name:         request.Body.Name,
		Description:  request.Body.Description,
		ImageURL:     request.Body.ImageUrl,
		PricePerHour: request.Body.PricePerHour,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.CreateAdminEquipment400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrNotFound):
			return api.CreateAdminEquipment404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.CreateAdminEquipment409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.CreateAdminEquipment500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.CreateAdminEquipment201JSONResponse(toAPIEquipment(equipment)), nil
}

func (h *Handler) UpdateAdminEquipment(ctx context.Context, request api.UpdateAdminEquipmentRequestObject) (api.UpdateAdminEquipmentResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminEquipment403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminEquipment400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	equipment, err := h.service.UpdateEquipment(ctx, request.Id, UpdateEquipmentParams{
		CategoryID:   request.Body.CategoryId,
		Name:         request.Body.Name,
		Description:  request.Body.Description,
		ImageURL:     request.Body.ImageUrl,
		PricePerHour: request.Body.PricePerHour,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.UpdateAdminEquipment400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminEquipment404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.UpdateAdminEquipment409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminEquipment500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminEquipment200JSONResponse(toAPIEquipment(equipment)), nil
}

func (h *Handler) DeleteAdminEquipment(ctx context.Context, request api.DeleteAdminEquipmentRequestObject) (api.DeleteAdminEquipmentResponseObject, error) {
	if !isAdmin(ctx) {
		return api.DeleteAdminEquipment403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	err := h.service.DeleteEquipment(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.DeleteAdminEquipment400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.DeleteAdminEquipment404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.DeleteAdminEquipment409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.DeleteAdminEquipment500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.DeleteAdminEquipment204Response{}, nil
}

// EQUIPMENT UNITS

func (h *Handler) GetAdminEquipmentUnits(ctx context.Context, request api.GetAdminEquipmentUnitsRequestObject) (api.GetAdminEquipmentUnitsResponseObject, error) {
	if !isAdmin(ctx) {
		return api.GetAdminEquipmentUnits403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	units, err := h.service.GetEquipmentUnits(ctx, request.Params.EquipmentId)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.GetAdminEquipmentUnits400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		default:
			return api.GetAdminEquipmentUnits500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}

	result := make([]api.AdminEquipmentUnit, 0, len(units))
	for _, unit := range units {
		result = append(result, toAPIEquipmentUnit(&unit))
	}
	return api.GetAdminEquipmentUnits200JSONResponse(result), nil
}

func (h *Handler) CreateAdminEquipmentUnit(ctx context.Context, request api.CreateAdminEquipmentUnitRequestObject) (api.CreateAdminEquipmentUnitResponseObject, error) {
	if !isAdmin(ctx) {
		return api.CreateAdminEquipmentUnit403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.CreateAdminEquipmentUnit400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	unit, err := h.service.CreateEquipmentUnit(ctx, CreateEquipmentUnitParams{
		EquipmentID:     request.Body.EquipmentId,
		InventoryNumber: request.Body.InventoryNumber,
		Size:            request.Body.Size,
		Status:          string(request.Body.Status),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.CreateAdminEquipmentUnit400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrNotFound):
			return api.CreateAdminEquipmentUnit404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.CreateAdminEquipmentUnit409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.CreateAdminEquipmentUnit500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.CreateAdminEquipmentUnit201JSONResponse(toAPIEquipmentUnit(unit)), nil
}

func (h *Handler) UpdateAdminEquipmentUnit(ctx context.Context, request api.UpdateAdminEquipmentUnitRequestObject) (api.UpdateAdminEquipmentUnitResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminEquipmentUnit403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminEquipmentUnit400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	params := UpdateEquipmentUnitParams{
		InventoryNumber: request.Body.InventoryNumber,
		Size:            request.Body.Size,
	}
	if request.Body.Status != nil {
		status := string(*request.Body.Status)
		params.Status = &status
	}

	unit, err := h.service.UpdateEquipmentUnit(ctx, request.Id, params)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.UpdateAdminEquipmentUnit400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminEquipmentUnit404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.UpdateAdminEquipmentUnit409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminEquipmentUnit500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminEquipmentUnit200JSONResponse(toAPIEquipmentUnit(unit)), nil
}

// EQUIPMENT SETS

func (h *Handler) CreateAdminEquipmentSet(ctx context.Context, request api.CreateAdminEquipmentSetRequestObject) (api.CreateAdminEquipmentSetResponseObject, error) {
	if !isAdmin(ctx) {
		return api.CreateAdminEquipmentSet403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.CreateAdminEquipmentSet400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	active := true
	if request.Body.IsActive != nil {
		active = *request.Body.IsActive
	}

	set, err := h.service.CreateEquipmentSet(ctx, CreateEquipmentSetParams{
		Name:         request.Body.Name,
		Description:  request.Body.Description,
		SportType:    string(request.Body.SportType),
		AgeGroup:     string(request.Body.AgeGroup),
		PricePerHour: request.Body.PricePerHour,
		IsActive:     active,
		Items:        toSetItems(request.Body.Items),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.CreateAdminEquipmentSet400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrNotFound):
			return api.CreateAdminEquipmentSet404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.CreateAdminEquipmentSet409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.CreateAdminEquipmentSet500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.CreateAdminEquipmentSet201JSONResponse(toAPIEquipmentSet(set)), nil
}

func (h *Handler) UpdateAdminEquipmentSet(ctx context.Context, request api.UpdateAdminEquipmentSetRequestObject) (api.UpdateAdminEquipmentSetResponseObject, error) {
	if !isAdmin(ctx) {
		return api.UpdateAdminEquipmentSet403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}
	if request.Body == nil {
		return api.UpdateAdminEquipmentSet400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse{Error: "invalid_input", Message: "Отсутствует тело запроса"}}, nil
	}

	params := UpdateEquipmentSetParams{
		Name:         request.Body.Name,
		Description:  request.Body.Description,
		PricePerHour: request.Body.PricePerHour,
		IsActive:     request.Body.IsActive,
	}

	if request.Body.SportType != nil {
		sportType := string(*request.Body.SportType)
		params.SportType = &sportType
	}
	if request.Body.AgeGroup != nil {
		ageGroup := string(*request.Body.AgeGroup)
		params.AgeGroup = &ageGroup
	}
	if request.Body.Items != nil {
		items := toSetItems(*request.Body.Items)
		params.Items = &items
	}

	set, err := h.service.UpdateEquipmentSet(ctx, request.Id, params)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidReference):
			return api.UpdateAdminEquipmentSet400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.UpdateAdminEquipmentSet404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrConflict):
			return api.UpdateAdminEquipmentSet409JSONResponse{AdminConflictJSONResponse: api.AdminConflictJSONResponse(apiError(err))}, nil
		default:
			return api.UpdateAdminEquipmentSet500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.UpdateAdminEquipmentSet200JSONResponse(toAPIEquipmentSet(set)), nil
}

func (h *Handler) DeleteAdminEquipmentSet(ctx context.Context, request api.DeleteAdminEquipmentSetRequestObject) (api.DeleteAdminEquipmentSetResponseObject, error) {
	if !isAdmin(ctx) {
		return api.DeleteAdminEquipmentSet403JSONResponse{AdminForbiddenJSONResponse: api.AdminForbiddenJSONResponse{Error: "forbidden", Message: "Недостаточно прав"}}, nil
	}

	err := h.service.DeleteEquipmentSet(ctx, request.Id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return api.DeleteAdminEquipmentSet400JSONResponse{AdminBadRequestJSONResponse: api.AdminBadRequestJSONResponse(apiError(err))}, nil
		case errors.Is(err, ErrNotFound):
			return api.DeleteAdminEquipmentSet404JSONResponse{AdminNotFoundJSONResponse: api.AdminNotFoundJSONResponse(apiError(err))}, nil
		default:
			return api.DeleteAdminEquipmentSet500JSONResponse{AdminInternalErrorJSONResponse: api.AdminInternalErrorJSONResponse(apiError(err))}, nil
		}
	}
	return api.DeleteAdminEquipmentSet204Response{}, nil
}
