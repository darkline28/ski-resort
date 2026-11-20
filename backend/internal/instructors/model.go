package instructors

import "time"

type Instructor struct {
	ID              int
	Name            string
	Specialization  string
	Description     *string
	ExperienceYears int
	IsActive        bool
	ImageURL        *string
}

type InstructorSchedule struct {
	ID       int
	StartsAt time.Time
	EndsAt   time.Time
}

type BookedInterval struct {
	StartsAt time.Time
	EndsAt   time.Time
}

type InstructorScheduleResponse struct {
	InstructorID    int
	Date            string
	Schedule        []InstructorSchedule
	BookedIntervals []BookedInterval
}

type InstructorBooking struct {
	ID          int
	ClientName  string
	ServiceName string
	StartsAt    time.Time
	EndsAt      time.Time
	Status      string
}

type AvailableSlot struct {
	StartsAt time.Time
	EndsAt   time.Time
}

type InstructorServiceInfo struct {
	ID              int
	Name            string
	Description     *string
	DurationMinutes int
	Price           int
	IsActive        bool
}
