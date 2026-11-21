package admin

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidInput = errors.New("invalid input")

type AdminRepository interface {
	// Services
	GetServices(ctx context.Context) ([]ServiceModel, error)
	GetServiceByID(ctx context.Context, id int) (*ServiceModel, error)
	CreateService(ctx context.Context, params CreateServiceParams) (*ServiceModel, error)
	UpdateService(ctx context.Context, id int, params UpdateServiceParams) (*ServiceModel, error)
	DeleteService(ctx context.Context, id int) error

	// Instructors
	GetInstructorByID(ctx context.Context, id int) (*Instructor, error)
	GetInstructors(ctx context.Context) ([]Instructor, error)
	CreateInstructor(ctx context.Context, params CreateInstructorParams) (*Instructor, error)
	UpdateInstructor(ctx context.Context, id int, params UpdateInstructorParams) (*Instructor, error)
	DeleteInstructor(ctx context.Context, id int) error

	// Instructor services
	GetInstructorServices(ctx context.Context, instructorID int) ([]ServiceModel, error)
	AssignInstructorService(ctx context.Context, instructorID, serviceID int) error
	RemoveInstructorService(ctx context.Context, instructorID, serviceID int) error

	// Instructor schedules
	GetInstructorSchedules(ctx context.Context, instructorID int, filter ScheduleFilter) ([]InstructorSchedule, error)
	CreateInstructorSchedule(ctx context.Context, instructorID int, params CreateScheduleParams) (*InstructorSchedule, error)
	UpdateInstructorSchedule(ctx context.Context, instructorID, scheduleID int, params UpdateScheduleParams) (*InstructorSchedule, error)
	DeleteInstructorSchedule(ctx context.Context, instructorID, scheduleID int) error

	// Lesson bookings
	GetLessonBookings(ctx context.Context, filter LessonBookingFilter) ([]AdminLessonBooking, error)
	GetLessonBookingByID(ctx context.Context, id int) (*AdminLessonBooking, error)
	UpdateLessonBookingStatus(ctx context.Context, id int, status string) (*AdminLessonBooking, error)
	UpdateOwnLessonBookingStatus(ctx context.Context, userID, bookingID int, status string) (*InstructorBooking, error)

	// Equipment bookings
	GetEquipmentBookings(ctx context.Context, filter EquipmentBookingFilter) ([]AdminEquipmentRental, error)
	GetEquipmentBookingByID(ctx context.Context, id int) (*AdminEquipmentRental, error)
	UpdateEquipmentBookingStatus(ctx context.Context, id int, status string) (*AdminEquipmentRental, error)

	// Equipment
	CreateEquipment(ctx context.Context, params CreateEquipmentParams) (*Equipment, error)
	UpdateEquipment(ctx context.Context, id int, params UpdateEquipmentParams) (*Equipment, error)
	DeleteEquipment(ctx context.Context, id int) error

	// Equipment units
	GetEquipmentUnits(ctx context.Context, equipmentID *int) ([]EquipmentUnit, error)
	CreateEquipmentUnit(ctx context.Context, params CreateEquipmentUnitParams) (*EquipmentUnit, error)
	UpdateEquipmentUnit(ctx context.Context, id int, params UpdateEquipmentUnitParams) (*EquipmentUnit, error)

	// Equipment sets
	CreateEquipmentSet(ctx context.Context, params CreateEquipmentSetParams) (*EquipmentSet, error)
	UpdateEquipmentSet(ctx context.Context, id int, params UpdateEquipmentSetParams) (*EquipmentSet, error)
	DeleteEquipmentSet(ctx context.Context, id int) error
}

type Service struct {
	repo AdminRepository
}

func NewService(repo AdminRepository) *Service {
	return &Service{repo: repo}
}

// SERVICES

func (s *Service) GetServices(ctx context.Context) ([]ServiceModel, error) {
	return s.repo.GetServices(ctx)
}

func (s *Service) GetService(ctx context.Context, id int) (*ServiceModel, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetServiceByID(ctx, id)
}

func (s *Service) CreateService(ctx context.Context, params CreateServiceParams) (*ServiceModel, error) {
	params.Name = strings.TrimSpace(params.Name)

	if params.Name == "" || params.DurationMinutes <= 0 || params.Price < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.CreateService(ctx, params)
}

func (s *Service) UpdateService(ctx context.Context, id int, params UpdateServiceParams) (*ServiceModel, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	if params.Name == nil && params.Description == nil && params.DurationMinutes == nil && params.Price == nil && params.IsActive == nil {
		return nil, ErrInvalidInput
	}

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)
		if name == "" {
			return nil, ErrInvalidInput
		}
		params.Name = &name
	}

	if params.DurationMinutes != nil && *params.DurationMinutes <= 0 {
		return nil, ErrInvalidInput
	}

	if params.Price != nil && *params.Price < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateService(ctx, id, params)
}

func (s *Service) DeleteService(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteService(ctx, id)
}

// INSTRUCTORS

func (s *Service) GetInstructors(ctx context.Context) ([]Instructor, error) {
	return s.repo.GetInstructors(ctx)
}

func (s *Service) GetInstructor(ctx context.Context, id int) (*Instructor, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetInstructorByID(ctx, id)
}

func (s *Service) CreateInstructor(ctx context.Context, params CreateInstructorParams) (*Instructor, error) {
	params.Specialization = strings.TrimSpace(params.Specialization)

	if params.UserID <= 0 || params.Specialization == "" || params.ExperienceYears < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.CreateInstructor(ctx, params)
}

func (s *Service) UpdateInstructor(ctx context.Context, id int, params UpdateInstructorParams) (*Instructor, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	if params.Specialization == nil && params.Description == nil && params.ExperienceYears == nil && params.ImageURL == nil && params.IsActive == nil {
		return nil, ErrInvalidInput
	}

	if params.Specialization != nil {
		specialization := strings.TrimSpace(*params.Specialization)
		if specialization == "" {
			return nil, ErrInvalidInput
		}
		params.Specialization = &specialization
	}

	if params.ExperienceYears != nil && *params.ExperienceYears < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateInstructor(ctx, id, params)
}

func (s *Service) DeleteInstructor(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteInstructor(ctx, id)
}

// INSTRUCTOR SERVICES

func (s *Service) GetInstructorServices(ctx context.Context, instructorID int) ([]ServiceModel, error) {
	if instructorID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetInstructorServices(ctx, instructorID)
}

func (s *Service) AssignInstructorService(ctx context.Context, instructorID, serviceID int) error {
	if instructorID <= 0 || serviceID <= 0 {
		return ErrInvalidInput
	}

	return s.repo.AssignInstructorService(ctx, instructorID, serviceID)
}

func (s *Service) RemoveInstructorService(ctx context.Context, instructorID, serviceID int) error {
	if instructorID <= 0 || serviceID <= 0 {
		return ErrInvalidInput
	}

	return s.repo.RemoveInstructorService(ctx, instructorID, serviceID)
}

// INSTRUCTOR SCHEDULES

func (s *Service) GetInstructorSchedules(ctx context.Context, instructorID int, filter ScheduleFilter) ([]InstructorSchedule, error) {
	if instructorID <= 0 {
		return nil, ErrInvalidInput
	}

	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo) {
		return nil, ErrInvalidInput
	}

	return s.repo.GetInstructorSchedules(ctx, instructorID, filter)
}

func (s *Service) CreateInstructorSchedule(ctx context.Context, instructorID int, params CreateScheduleParams) (*InstructorSchedule, error) {
	if instructorID <= 0 {
		return nil, ErrInvalidInput
	}

	if params.StartsAt.IsZero() || params.EndsAt.IsZero() {
		return nil, ErrInvalidInput
	}

	if !params.StartsAt.Before(params.EndsAt) {
		return nil, ErrInvalidInput
	}

	return s.repo.CreateInstructorSchedule(ctx, instructorID, params)
}

func (s *Service) UpdateInstructorSchedule(ctx context.Context, instructorID, scheduleID int, params UpdateScheduleParams) (*InstructorSchedule, error) {
	if instructorID <= 0 || scheduleID <= 0 {
		return nil, ErrInvalidInput
	}

	if params.StartsAt == nil && params.EndsAt == nil {
		return nil, ErrInvalidInput
	}

	if params.StartsAt != nil && params.StartsAt.IsZero() {
		return nil, ErrInvalidInput
	}

	if params.EndsAt != nil && params.EndsAt.IsZero() {
		return nil, ErrInvalidInput
	}

	if params.StartsAt != nil && params.EndsAt != nil && !params.StartsAt.Before(*params.EndsAt) {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateInstructorSchedule(ctx, instructorID, scheduleID, params)
}

func (s *Service) DeleteInstructorSchedule(ctx context.Context, instructorID, scheduleID int) error {
	if instructorID <= 0 || scheduleID <= 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteInstructorSchedule(ctx, instructorID, scheduleID)
}

// LESSON BOOKINGS

func (s *Service) GetLessonBookings(ctx context.Context, filter LessonBookingFilter) ([]AdminLessonBooking, error) {
	if filter.Status != nil && !validLessonStatus(*filter.Status) {
		return nil, ErrInvalidInput
	}

	if filter.UserID != nil && *filter.UserID <= 0 {
		return nil, ErrInvalidInput
	}

	if filter.InstructorID != nil && *filter.InstructorID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetLessonBookings(ctx, filter)
}

func (s *Service) GetLessonBookingByID(ctx context.Context, id int) (*AdminLessonBooking, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetLessonBookingByID(ctx, id)
}

func (s *Service) UpdateLessonBookingStatus(ctx context.Context, id int, status string) (*AdminLessonBooking, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	switch status {
	case "cancelled", "completed", "no_show":
		return s.repo.UpdateLessonBookingStatus(ctx, id, status)
	default:
		return nil, ErrInvalidInput
	}
}

func (s *Service) UpdateOwnLessonBookingStatus(ctx context.Context, userID, bookingID int, status string) (*InstructorBooking, error) {
	if userID <= 0 || bookingID <= 0 {
		return nil, ErrInvalidInput
	}

	if status != "completed" && status != "no_show" {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateOwnLessonBookingStatus(ctx, userID, bookingID, status)
}

// EQUIPMENT BOOKINGS

func (s *Service) GetEquipmentBookings(ctx context.Context, filter EquipmentBookingFilter) ([]AdminEquipmentRental, error) {
	if filter.Status != nil && !validRentalStatus(*filter.Status) {
		return nil, ErrInvalidInput
	}

	if filter.UserID != nil && *filter.UserID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetEquipmentBookings(ctx, filter)
}

func (s *Service) GetEquipmentBookingByID(ctx context.Context, id int) (*AdminEquipmentRental, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetEquipmentBookingByID(ctx, id)
}

func (s *Service) UpdateEquipmentBookingStatus(ctx context.Context, id int, status string) (*AdminEquipmentRental, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	switch status {
	case "confirmed", "cancelled", "issued", "returned":
		return s.repo.UpdateEquipmentBookingStatus(ctx, id, status)
	default:
		return nil, ErrInvalidInput
	}
}

// EQUIPMENT

func (s *Service) CreateEquipment(ctx context.Context, params CreateEquipmentParams) (*Equipment, error) {
	params.Name = strings.TrimSpace(params.Name)

	if params.CategoryID <= 0 || params.Name == "" || params.PricePerHour < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.CreateEquipment(ctx, params)
}

func (s *Service) UpdateEquipment(ctx context.Context, id int, params UpdateEquipmentParams) (*Equipment, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	if params.CategoryID == nil && params.Name == nil && params.Description == nil && params.ImageURL == nil && params.PricePerHour == nil {
		return nil, ErrInvalidInput
	}

	if params.CategoryID != nil && *params.CategoryID <= 0 {
		return nil, ErrInvalidInput
	}

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)
		if name == "" {
			return nil, ErrInvalidInput
		}
		params.Name = &name
	}

	if params.PricePerHour != nil && *params.PricePerHour < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateEquipment(ctx, id, params)
}

func (s *Service) DeleteEquipment(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteEquipment(ctx, id)
}

// EQUIPMENT UNITS

func (s *Service) GetEquipmentUnits(ctx context.Context, equipmentID *int) ([]EquipmentUnit, error) {
	if equipmentID != nil && *equipmentID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetEquipmentUnits(ctx, equipmentID)
}

func (s *Service) CreateEquipmentUnit(ctx context.Context, params CreateEquipmentUnitParams) (*EquipmentUnit, error) {
	params.InventoryNumber = strings.TrimSpace(params.InventoryNumber)
	params.Size = strings.TrimSpace(params.Size)

	if params.EquipmentID <= 0 || params.InventoryNumber == "" || params.Size == "" {
		return nil, ErrInvalidInput
	}

	if !validUnitStatus(params.Status) {
		return nil, ErrInvalidInput
	}

	return s.repo.CreateEquipmentUnit(ctx, params)
}

func (s *Service) UpdateEquipmentUnit(ctx context.Context, id int, params UpdateEquipmentUnitParams) (*EquipmentUnit, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	if params.InventoryNumber == nil && params.Size == nil && params.Status == nil {
		return nil, ErrInvalidInput
	}

	if params.InventoryNumber != nil {
		number := strings.TrimSpace(*params.InventoryNumber)
		if number == "" {
			return nil, ErrInvalidInput
		}
		params.InventoryNumber = &number
	}

	if params.Size != nil {
		size := strings.TrimSpace(*params.Size)
		if size == "" {
			return nil, ErrInvalidInput
		}
		params.Size = &size
	}

	if params.Status != nil && !validUnitStatus(*params.Status) {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateEquipmentUnit(ctx, id, params)
}

// EQUIPMENT SETS

func (s *Service) CreateEquipmentSet(ctx context.Context, params CreateEquipmentSetParams) (*EquipmentSet, error) {
	params.Name = strings.TrimSpace(params.Name)

	if params.Name == "" || params.PricePerHour < 0 {
		return nil, ErrInvalidInput
	}

	if !validSportType(params.SportType) || !validAgeGroup(params.AgeGroup) {
		return nil, ErrInvalidInput
	}

	if !validSetItems(params.Items) {
		return nil, ErrInvalidInput
	}

	return s.repo.CreateEquipmentSet(ctx, params)
}

func (s *Service) UpdateEquipmentSet(ctx context.Context, id int, params UpdateEquipmentSetParams) (*EquipmentSet, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	if params.Name == nil && params.Description == nil && params.SportType == nil && params.AgeGroup == nil && params.PricePerHour == nil && params.IsActive == nil && params.Items == nil {
		return nil, ErrInvalidInput
	}

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)
		if name == "" {
			return nil, ErrInvalidInput
		}
		params.Name = &name
	}

	if params.SportType != nil && !validSportType(*params.SportType) {
		return nil, ErrInvalidInput
	}

	if params.AgeGroup != nil && !validAgeGroup(*params.AgeGroup) {
		return nil, ErrInvalidInput
	}

	if params.PricePerHour != nil && *params.PricePerHour < 0 {
		return nil, ErrInvalidInput
	}

	if params.Items != nil && !validSetItems(*params.Items) {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateEquipmentSet(ctx, id, params)
}

func (s *Service) DeleteEquipmentSet(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteEquipmentSet(ctx, id)
}

// VALIDATION HELPERS

func validLessonStatus(status string) bool {
	switch status {
	case "confirmed", "cancelled", "completed", "no_show":
		return true
	default:
		return false
	}
}

func validRentalStatus(status string) bool {
	switch status {
	case "pending", "confirmed", "issued", "returned", "cancelled":
		return true
	default:
		return false
	}
}

func validUnitStatus(status string) bool {
	switch status {
	case "available", "maintenance", "written_off":
		return true
	default:
		return false
	}
}

func validSportType(sportType string) bool {
	return sportType == "ski" || sportType == "snowboard"
}

func validAgeGroup(ageGroup string) bool {
	return ageGroup == "adult" || ageGroup == "child"
}

func validSetItems(items []EquipmentSetItemParams) bool {
	if len(items) == 0 {
		return false
	}

	categories := make(map[int]bool, len(items))

	for _, item := range items {
		if item.CategoryID <= 0 || item.Quantity <= 0 {
			return false
		}

		if categories[item.CategoryID] {
			return false
		}

		categories[item.CategoryID] = true
	}

	return true
}
