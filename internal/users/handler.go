package users

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/axolotl-go/eternal_paw/internal/auth"
	"github.com/axolotl-go/eternal_paw/internal/db"
	"github.com/axolotl-go/eternal_paw/internal/hash"
	"github.com/axolotl-go/eternal_paw/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var validate = validator.New()

func Create(c *fiber.Ctx) error {
	var user User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			// "detail": err.Error(),
		})
	}

	if !utils.IsNotNull(
		user.Name,
		user.Email,
		user.Password,
		user.Phone,
		user.Address,
	) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email address",
		})
	}

	if err := ValidatePassword(user.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if CheckUserExists(user.Email) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User already exists",
		})
	}

	hashed, err := hash.Hash(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	user.Email = strings.ToLower(user.Email)
	user.Password = hashed

	if err := CreateUser(&user); err != nil {

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "User already exists",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"status":  "success",
	})
}

func Edit(c *fiber.Ctx) error {
	var update UpdateUser

	user, ok := c.Locals("user").(*User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	update.Name = strings.ToLower(strings.TrimSpace(update.Name))
	update.Phone = strings.TrimSpace(update.Phone)
	update.Address = strings.ToLower(strings.TrimSpace(update.Address))

	if err := validate.Struct(update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if update.Name != "" {
		user.Name = update.Name
	}

	if update.Phone != "" {
		user.Phone = update.Phone
	}

	if update.Address != "" {
		user.Address = update.Address
	}

	if err := db.DB.Save(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not update user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "successful",
	})
}

func Delete(c *fiber.Ctx) error {
	var user = c.Locals("user").(*User)

	if err := db.DB.Delete(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// All routes, login...
func Login(c *fiber.Ctx) error {
	var input LoginInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid body",
		})
	}

	input.Email = strings.ToLower(input.Email)

	user, err := AuthenticateUser(input.Email, input.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	if !ComparePassword(user.Password, input.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "None",
		MaxAge:   60 * 60 * 24,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":       user.ID,
		"username": user.Name,
		"email":    user.Email,
		"token":    token,
	})
}

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "None",
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func Me(c *fiber.Ctx) error {
	user := c.Locals("user").(*User)

	return c.JSON(fiber.Map{
		"authenticated": true,
		"role":          user.Role,
		"name":          user.Name,
		"email":         user.Email,
		"phone":         user.Phone,
		"address":       user.Address,
	})
}

func Verify(c *fiber.Ctx) error {
	var user User

	tokenString := c.Cookies("token")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No token provided",
		})
	}

	claims, err := auth.ParserJwt(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	if err := db.DB.Where("id = ?", claims["id"]).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"authenticated": true,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Name,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func Views(c *fiber.Ctx) error {
	var users []User

	if err := db.DB.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad Request",
		})
	}

	for i := range users {
		users[i].Password = ""
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": users,
	})
}

func DeleteByAdmin(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user id",
		})
	}

	result := db.DB.Unscoped().Delete(&User{}, id)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func EditByAdmin(c *fiber.Ctx) error {
	targetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID de usuario inválido",
		})
	}

	var user User
	if err := db.DB.First(&user, targetID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Usuario no encontrado",
		})
	}

	var update User
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cuerpo de la petición inválido",
		})
	}

	if update.Name != "" {
		user.Name = update.Name
	}
	if update.Email != "" {
		user.Email = update.Email
	}
	if update.Address != "" {
		user.Address = update.Address
	}
	if update.Phone != "" {
		user.Phone = update.Phone
	}
	if update.Role != "" {
		user.Role = update.Role
	}

	if update.Password != "" {
		pass, err := hash.Hash(update.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al procesar la contraseña",
			})
		}
		user.Password = pass
	}

	if err := db.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al actualizar el usuario en la base de datos",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Usuario actualizado exitosamente",
		"user":    user,
	})
}

func ChangePassword(c *fiber.Ctx) error {
	type Change struct {
		Password        string `json:"password" validate:"required,min=8,max=72"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
		CurrentPassword string `json:"current_password" validate:"required,min=8,max=72"`
	}

	user, ok := c.Locals("user").(*User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Usuario no autenticado",
		})
	}

	var change Change

	if err := c.BodyParser(&change); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos inválidos",
		})
	}

	if err := validate.Struct(change); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := hash.Verify(user.Password, change.CurrentPassword); err {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "La contraseña actual es incorrecta",
		})
	}

	hashedPassword, err := hash.Hash(change.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo actualizar la contraseña",
		})
	}

	user.Password = hashedPassword

	if err := db.DB.Save(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo actualizar la contraseña",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Contraseña actualizada correctamente",
	})
}
