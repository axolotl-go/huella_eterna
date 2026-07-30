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
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
