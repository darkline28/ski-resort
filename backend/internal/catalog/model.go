package catalog

type Equipment struct {
	ID           int
	Name         string
	Category     string
	PricePerHour int
	Quantity     int
	Description  *string
	ImageURL     *string
}

type Category struct {
	ID   int
	Name string
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

type EquipmentSetItem struct {
	CategoryID int
	Category   string
	Quantity   int
}
