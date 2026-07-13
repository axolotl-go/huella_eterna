package middleware

import (
	"github.com/axolotl-go/eternal_paw/internal/auth"
	"github.com/axolotl-go/eternal_paw/internal/db"
	"github.com/axolotl-go/eternal_paw/internal/users"
	"github.com/gofiber/fiber/v2"
)

func Role(c *fiber.Ctx) error {
	var user users.User
	token := c.Cookies("token")

	claims, err := auth.ParserJwt(token)
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

	if user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized",
		})
	}

	return c.Next()
}
