package services

type Service struct {
	ID              int
	Name            string
	Description     *string
	DurationMinutes int
	Price           int
	IsActive        bool
}
