package models

import "time"

type Book struct {
	ID        uint      `json:"id,omitempty"`
	Title     string    `json:"title,omitempty" pagination:"search,filter"`
	Author    string    `json:"author,omitempty" pagination:"search,filter"`
	CreatedAt time.Time `json:"created_at,omitempty" pagination:"filter"`
	UpdatedAt time.Time `json:"updated_at,omitempty" pagination:"filter"`
}
