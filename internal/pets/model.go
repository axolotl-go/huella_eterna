package pets

import (
	"time"

	"gorm.io/gorm"
)

type Pet struct {
	gorm.Model

	UserID    uint       `json:"user_id"`
	ImageUrl  *string    `json:"image_url,omitempty"`
	Name      string     `json:"name" validate:"required"`
	Species   string     `json:"species" validate:"required"`
	Breed     string     `json:"breed"`
	Color     string     `json:"color" validate:"required"`
	Weight    float64    `json:"weight" validate:"gt=0"`
	DeathDate *time.Time `json:"death_date"`
}
