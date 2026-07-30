package pets

import (
	"time"

	"gorm.io/gorm"
)

type Pet struct {
	gorm.Model

	UserID      uint       `json:"user_id"`
	ImageUrl    *string    `json:"image_url,omitempty" validate:"omitempty,max=255"`
	Name        string     `json:"name" validate:"required,max=255"`
	Species     string     `json:"species" validate:"required,max=255"`
	Breed       string     `json:"breed" validate:"max=255"`
	Color       string     `json:"color" validate:"required,max=255"`
	Weight      float64    `json:"weight" validate:"required,gt=0,lte=1000"`
	Sex         string     `validate:"required,oneof=macho hembra"`
	Observation string     `json:"observation" validate:"required,max=255"`
	Age         uint       `json:"age" validate:"gte=0,lte=100"`
	DeathDate   *time.Time `json:"death_date"`
}
