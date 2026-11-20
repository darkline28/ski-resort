package bookings

import "time"

type CreateLessonBookingParams struct {
	UserID       int
	InstructorID int
	ServiceID    int
	StartsAt     time.Time
	EndsAt       time.Time
	TotalPrice   int
}

type LessonBooking struct {
	ID           int
	UserID       int
	InstructorID int
	ServiceID    int
	StartsAt     time.Time
	EndsAt       time.Time
	Status       string
	TotalPrice   int
	CreatedAt    time.Time
}

type BookingServiceInfo struct {
	DurationMinutes int
	Price           int
	IsActive        bool
}

type LessonBookingDetail struct {
	ID             int
	UserID         int
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
