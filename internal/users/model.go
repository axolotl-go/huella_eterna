package users

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string `gorm:"not null; size:255;" json:"name"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"password_hash"`
	Phone    string `gorm:"size:15" json:"phone"`
	Address  string `json:"address"`
	Role     string `gorm:"not null;default:client" json:"role"`
}

type UpdateUser struct {
	Name    string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone   string `json:"phone" validate:"omitempty,min=10,max=15,numeric"`
	Address string `json:"address" validate:"omitempty,min=5,max=255"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
