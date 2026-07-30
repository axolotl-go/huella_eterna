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
		return c.SendString("Hello, little tree!")
	})

	// ===========================
	// Public routes
	// ===========================

	public := app.Group("/api")

	// Auth
	public.Post("/login", users.Login)
	public.Post("/user", users.Create)
	public.Get("/verify", users.Verify)

	// Landing
	public.Post("/form_register", formregister.Create)

	// ===========================
	// Private routes
	// ===========================

	private := app.Group("/api")
	private.Use(middleware.CurrentUser)

	// User
	private.Get("/me", users.Me)
	private.Put("/user", users.Edit)
	private.Delete("/user", users.Delete)
	private.Post("/logout", users.Logout)

	// Orders
	private.Post("/order", serviceorders.Create)
	private.Get("/order/:folio", serviceorders.View)
	private.Get("/me/orders", serviceorders.ViewMyOrders)
	private.Put("/order/:id", serviceorders.Edit)
	private.Delete("/order/:id", serviceorders.Delete)

	// Services
	private.Get("/services", servicetype.Views)

	// ===========================
	// Admin routes
	// ===========================

	admin := private.Group("/")
	admin.Use(middleware.Role)

	// Users
	admin.Get("/users", users.Views)
	admin.Delete("/users/:id", users.DeleteByAdmin)
	admin.Get("/users/:id/orders", serviceorders.ViewMyOrders)

	// Contact forms
	admin.Get("/form_register", formregister.Views)

	// Orders
	admin.Get("/orders", serviceorders.Views)

	// Services
	admin.Post("/service", servicetype.Create)
	admin.Put("/service/:id", servicetype.Edit)
	admin.Delete("/service/:id", servicetype.Delete)
	admin.Get("/service/:id", servicetype.View)
}
