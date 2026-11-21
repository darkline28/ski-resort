package admin

import "time"

// Services

type ServiceModel struct {
	ID              int
	Name            string
	Description     *string
	DurationMinutes int
	Price           int
	IsActive        bool
}

type CreateServiceParams struct {
	Name            string
	Description     *string
	DurationMinutes int
	Price           int
	IsActive        bool
}

type UpdateServiceParams struct {
	Name            *string
	Description     *string
	DurationMinutes *int
	Price           *int
	IsActive        *bool
}

// Instructors

type Instructor struct {
	ID              int
	Name            string
	Specialization  string
	Description     *string
	ExperienceYears int
	ImageURL        *string
	IsActive        bool
}

type CreateInstructorParams struct {
	UserID          int
	Specialization  string
	Description     *string
	ExperienceYears int
	ImageURL        *string
	IsActive        bool
}

type UpdateInstructorParams struct {
	Specialization  *string
	Description     *string
	ExperienceYears *int
	ImageURL        *string
	IsActive        *bool
}

// Instructor schedules

type InstructorSchedule struct {
	ID       int
	StartsAt time.Time
	EndsAt   time.Time
}

type ScheduleFilter struct {
	DateFrom *time.Time
	DateTo   *time.Time
}

type CreateScheduleParams struct {
	StartsAt time.Time
	EndsAt   time.Time
}

type UpdateScheduleParams struct {
	StartsAt *time.Time
	EndsAt   *time.Time
}

// Lesson bookings

type LessonBookingFilter struct {
	Status       *string
	Date         *time.Time
	UserID       *int
	InstructorID *int
}

type AdminLessonBooking struct {
	ID             int
	UserID         int
	ClientName     string
	InstructorID   int
	InstructorName string
	ServiceID      int
	ServiceName    string
	StartsAt       time.Time
	EndsAt         time.Time
	Status         string
	TotalPrice     int
	CreatedAt      time.Time
}

type InstructorBooking struct {
	ID          int
	ClientName  string
	ServiceName string
	StartsAt    time.Time
	EndsAt      time.Time
	Status      string
}

// Equipment rentals

type EquipmentBookingFilter struct {
	Status *string
	Date   *time.Time
	UserID *int
}

type RentalItem struct {
	EquipmentUnitID int
	EquipmentID     int
	Name            string
	Size            string
	BookedSetID     *int
	ItemTotalPrice  *int
}

type BookedRentalSet struct {
	ID             int
	EquipmentSetID int
	Name           string
	TotalPrice     int
}

type AdminEquipmentRental struct {
	ID         int
	UserID     int
	ClientName string
	StartsAt   time.Time
	EndsAt     time.Time
	Status     string
	TotalPrice int
	CreatedAt  time.Time
	Items      []RentalItem
	Sets       []BookedRentalSet
}

// Equipment

type Equipment struct {
	ID           int
	CategoryID   int
	Name         string
	Description  *string
	ImageURL     *string
	PricePerHour int
}

type CreateEquipmentParams struct {
	CategoryID   int
	Name         string
	Description  *string
	ImageURL     *string
	PricePerHour int
}

type UpdateEquipmentParams struct {
	CategoryID   *int
	Name         *string
	Description  *string
	ImageURL     *string
	PricePerHour *int
}

// Equipment units

type EquipmentUnit struct {
	ID              int
	EquipmentID     int
	InventoryNumber string
	Size            string
	Status          string
	CreatedAt       time.Time
}

type CreateEquipmentUnitParams struct {
	EquipmentID     int
	InventoryNumber string
	Size            string
	Status          string
}

type UpdateEquipmentUnitParams struct {
	InventoryNumber *string
	Size            *string
	Status          *string
}

// Equipment sets

type EquipmentSetItemParams struct {
	CategoryID int
	Quantity   int
}

type EquipmentSetItem struct {
	CategoryID   int
	CategoryName string
	Quantity     int
}

type EquipmentSet struct {
	ID           int
	Name         string
	Description  *string
	SportType    string
	AgeGroup     string
	PricePerHour int
	IsActive     bool
	Items        []EquipmentSetItem
}

type CreateEquipmentSetParams struct {
	Name         string
	Description  *string
	SportType    string
	AgeGroup     string
	PricePerHour int
	IsActive     bool
	Items        []EquipmentSetItemParams
}

type UpdateEquipmentSetParams struct {
	Name         *string
	Description  *string
	SportType    *string
	AgeGroup     *string
	PricePerHour *int
	IsActive     *bool
	Items        *[]EquipmentSetItemParams
}
