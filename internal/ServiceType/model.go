package servicetype

import "gorm.io/gorm"

type ServiceType struct {
	gorm.Model

	Name        string  `gorm:"uniqueIndex;not null" json:"name"`
	Description string  `json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
}

type UpdateService struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
