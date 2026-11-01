package users

import "time"

type User struct {
	ID           int
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}
