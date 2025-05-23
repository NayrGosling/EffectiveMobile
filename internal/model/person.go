package model

import "time"

type Person struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Patronymic  *string   `json:"patronymic,omitempty"`
	Age         *int      `json:"age,omitempty"`
	Gender      *string   `json:"gender,omitempty"`
	Nationality *string   `json:"nationality,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Message     string    `json:"message"`
}
