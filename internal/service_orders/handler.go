package serviceorders

import (
	"errors"

	servicetype "github.com/axolotl-go/eternal_paw/internal/ServiceType"
	"github.com/axolotl-go/eternal_paw/internal/auth"
	"github.com/axolotl-go/eternal_paw/internal/db"
	"github.com/axolotl-go/eternal_paw/internal/pets"
	"github.com/axolotl-go/eternal_paw/internal/users"
	"github.com/axolotl-go/eternal_paw/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

const taxRate = 0.17

func Create(c *fiber.Ctx) error {
	var order Order
	var user users.User
	var serviceType servicetype.ServiceType

	token := c.Cookies("token")

	claims, err := auth.ParserJwt(token)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := validate.Struct(order); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := db.DB.Where("id = ?", claims["id"]).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	order.UserID = user.ID
	order.Pet.UserID = user.ID
	order.Pet.ID = order.Pet.ID

	if err := db.DB.Where("id = ?", order.ServiceTypeID).First(&serviceType).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service not found",
		})
	}

	order.Active = false
	order.OrderNumber = utils.GenerateOrder()

	// Pricing
	if order.Pet.Weight <= 0 {
		return errors.New("invalid pet weight")
	}

	if serviceType.Price <= 0 {
		return errors.New("invalid service price")
	}

	subtotal := order.Pet.Weight * serviceType.Price
	tax := subtotal * taxRate
	order.Price = subtotal + tax

	// Address
	order.Status = "pending"

	order.PickupAddress = ""
	if order.PickupRequired {
		if user.Address == "" {
			return errors.New("user has no address")
		}
		order.PickupAddress = user.Address
	}

	if err := db.DB.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Created successfully",
		"Order":   order,
	})
}

func Views(c *fiber.Ctx) error {
	var orders []Order
	var response []OrderPreview

	if err := db.DB.Preload("Pet").Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	for _, order := range orders {
		var user users.User
		var service servicetype.ServiceType

		if err := db.DB.First(&user, order.UserID).Error; err != nil {
			continue
		}

		if err := db.DB.First(&service, order.ServiceTypeID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		response = append(response, OrderPreview{
			ID: order.ID,

			Folio:       order.OrderNumber,
			PetName:     order.Pet.Name,
			ServiceName: service.Name,
			Status:      order.Status,
			Date:        &order.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(
		response,
	)
}

func View(c *fiber.Ctx) error {
	var (
		order    Order
		user     users.User
		pet      pets.Pet
		service  servicetype.ServiceType
		response OrderResponse
	)

	folio := c.Params("folio")

	if err := db.DB.Where("order_number = ?", folio).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Folio Not Found",
		})
	}

	if err := db.DB.First(&user, order.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User Not Found",
		})
	}

	if err := db.DB.First(&pet, order.PetID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Pet Not Found",
		})
	}

	if err := db.DB.First(&service, order.ServiceTypeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service Not Found",
		})
	}

	response = OrderResponse{
		UserName: user.Name,
		Email:    user.Email,
		Phone:    user.Phone,

		PetName:   pet.Name,
		Species:   pet.Species,
		Breed:     pet.Breed,
		Weight:    pet.Weight,
		DeathDate: pet.DeathDate,

		PickupRequired: order.PickupRequired,
		PickupAddress:  order.PickupAddress,

		ServiceID:   service.ID,
		ServiceName: service.Name,

		Folio:  order.OrderNumber,
		Active: order.Active,
		Price:  order.Price,
		Status: order.Status,
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}
