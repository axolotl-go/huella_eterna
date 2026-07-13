package http

import (
	servicetype "github.com/axolotl-go/eternal_paw/internal/ServiceType"
	formregister "github.com/axolotl-go/eternal_paw/internal/form_register"
	"github.com/axolotl-go/eternal_paw/internal/middleware"
	serviceorders "github.com/axolotl-go/eternal_paw/internal/service_orders"
	"github.com/axolotl-go/eternal_paw/internal/users"
	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App) {

	app.Get("/", func(c *fiber.Ctx) error {

		return c.SendString("Hello, little treea!")
	})

	api := app.Group("/api")

	// Landing FormContact
	api.Post("/form_register", formregister.Create)
	api.Get("/form_register", middleware.Role, formregister.Views)
	api.Get("/verify", users.Verify)

	// User
	api.Post("/login", users.Login)
	api.Get("/jwtVerify", users.JWT_Verify)
	api.Post("/user", users.Create)
	api.Post("/logout", users.Logout)

	// Pets
	// api.Post("/pets", pets.Create)
	// api.Get("/pet/:uuid", pets.Get)

	// Orders
	api.Get("/orders", middleware.Role, serviceorders.Views)
	api.Get("/orders/:folio", serviceorders.View)
	api.Post("/orders", serviceorders.Create)

	// // SeriveType
	// api.Post("/service_type", servicetype.Create)
	// api.Delete("/service_type/:id", servicetype.Delete)
	api.Get("/service_type", servicetype.Views)
	api.Get("/service_type/:id", servicetype.View)

	// Certificacion
}
