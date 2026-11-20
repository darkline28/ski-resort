package rental

import "time"

type EquipmentAvailability struct {
	EquipmentID       int
	Size              string
	AvailableQuantity int
}

type RentalItemRequest struct {
	EquipmentID int
	Size        string
	Quantity    int
}

type RentalSetSize struct {
	CategoryID int
	Size       string
}

type RentalSetRequest struct {
	EquipmentSetID int
	Quantity       int
	Sizes          []RentalSetSize
}

type CreateEquipmentBookingParams struct {
	StartsAt time.Time
	EndsAt   time.Time
	Items    []RentalItemRequest
	Sets     []RentalSetRequest
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

type EquipmentRental struct {
	ID         int
	UserID     int
	StartsAt   time.Time
	EndsAt     time.Time
	Status     string
	TotalPrice int
	CreatedAt  time.Time
	Items      []RentalItem
	Sets       []BookedRentalSet
}

type EquipmentSetInfo struct {
	ID           int
	PricePerHour int
	IsActive     bool
}

type EquipmentSetItem struct {
	CategoryID int
	Quantity   int
}
