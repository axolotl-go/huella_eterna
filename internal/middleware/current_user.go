package middleware

import (
	"github.com/axolotl-go/eternal_paw/internal/auth"
	"github.com/axolotl-go/eternal_paw/internal/db"
	"github.com/axolotl-go/eternal_paw/internal/users"
	"github.com/gofiber/fiber/v2"
)

func CurrentUser(c *fiber.Ctx) error {
	token := c.Cookies("token")

	claims, err := auth.ParserJwt(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	userID := uint(claims["id"].(float64))

	var user users.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	c.Locals("user", &user)

	return c.Next()
}
