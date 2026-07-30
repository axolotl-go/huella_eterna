package main

import (
	"log"

	servicetype "github.com/axolotl-go/eternal_paw/internal/ServiceType"
	"github.com/axolotl-go/eternal_paw/internal/config"
	"github.com/axolotl-go/eternal_paw/internal/db"
	formregister "github.com/axolotl-go/eternal_paw/internal/form_register"
	"github.com/axolotl-go/eternal_paw/internal/http"
	"github.com/axolotl-go/eternal_paw/internal/pets"
	serviceorders "github.com/axolotl-go/eternal_paw/internal/service_orders"
	"github.com/axolotl-go/eternal_paw/internal/users"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var serverxport string

func init() {
	cfg := config.Load()
	serverxport = cfg.ServerPort

	if serverxport == "" {
		serverxport = "3000"
	}
}

func main() {

	// file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Delete all info
	// db.DB.Migrator().DropTable(
	// 	&users.User{},
	// 	&pets.Pet{},
	// 	&serviceorders.Order{},
	// 	&servicetype.ServiceType{},
	// 	&formregister.Appointment{},
	// )

	db.DB.AutoMigrate(
		&users.User{},
		&pets.Pet{},
		&serviceorders.Order{},
		&servicetype.ServiceType{},
		&formregister.Appointment{},
	)

	app := fiber.New()
	app.Use(cors.New(config.CorsConfig()))

	// app.Use(logger.New(logger.Config{
	// 	Format: "${time} | ${status} | ${method} | ${path} | ${latency}\n",
	// 	Output: file,
	// }))

	http.SetupRouter(app)
	log.Fatal(app.Listen(serverxport))

}
