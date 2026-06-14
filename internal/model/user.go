package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` //for never including this in the json
	CreateAt     time.Time `json:"created_at"`
}
