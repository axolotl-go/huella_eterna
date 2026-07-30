package serviceorders

import (
	"time"

	"github.com/axolotl-go/eternal_paw/internal/pets"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model

	UserID uint     `json:"user_id"`
	PetID  uint     `json:"pet_id"`
	Pet    pets.Pet `json:"pet" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	OrderNumber   string `json:"order_number"`
	ServiceTypeID uint   `json:"service_type_id" validate:"required"`

	PickupRequired bool   `json:"pickup_required"`
	PickupAddress  string `json:"pickup_address"`

	Active bool    `json:"active" gorm:"default:false"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type OrderPreview struct {
	ID uint `json:"id"`

	PetName string `json:"pet_name"`

	ServiceName string  ` json:"service_name"`
	Folio       string  `json:"folio"`
	Status      string  `json:"status"`
	Price       float32 `json:"price"`

	Date *time.Time `json:"date"`
}

type OrderResponse struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`

	PetName     string     `json:"pet_name"`
	Species     string     `json:"species"`
	Breed       string     `json:"breed"`
	Weight      float64    `json:"weight"`
	Color       string     `json:"color"`
	Sex         string     `json:"sex"`
	Age         uint       `json:"age"`
	Observation string     `json:"observation"`
	DeathDate   *time.Time `json:"death_date"`

	PickupRequired bool   `json:"pickup_required"`
	PickupAddress  string `json:"pickup_address"`

	ServiceID   uint   `json:"service_id"`
	ServiceName string ` json:"service_name"`

	Folio  string  `json:"folio"`
	Active bool    `json:"active"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type OrderViewProfile struct {
	Folio          string     `json:"folio"`
	PetName        string     `json:"pet_name"`
	Species        string     `json:"species"`
	DeathDate      *time.Time `json:"death_date"`
	Status         string     `json:"status"`
	ServiceName    string     `json:"service_name"`
	Price          float64    `json:"price"`
	PickupRequired bool       `json:"pickup_required"`

	CreationDate string `json:"creation_date"`
}
