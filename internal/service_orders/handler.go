package serviceorders

import (
	"errors"
	"strconv"

	servicetype "github.com/axolotl-go/eternal_paw/internal/ServiceType"
	"github.com/axolotl-go/eternal_paw/internal/db"
	"github.com/axolotl-go/eternal_paw/internal/pets"
	"github.com/axolotl-go/eternal_paw/internal/users"
	"github.com/axolotl-go/eternal_paw/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

const taxRate = 0.17

var validate = validator.New()

func Create(c *fiber.Ctx) error {

	var (
		user        = c.Locals("user").(*users.User)
		order       Order
		serviceType servicetype.ServiceType
	)

	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	order.Status = "pending"

	if err := validate.Struct(order); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	order.UserID = user.ID
	order.Pet.UserID = user.ID

	if err := db.DB.First(&serviceType, order.ServiceTypeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service not found",
		})
	}

	order.ServiceName = serviceType.Name
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

		if err := db.DB.First(&user, order.UserID).Error; err != nil {
			continue
		}

		response = append(response, OrderPreview{
			ID: order.ID,

			Folio:       order.OrderNumber,
			PetName:     order.Pet.Name,
			ServiceName: order.ServiceName,
			Status:      order.Status,
			Price:       float32(order.Price),
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

	response = OrderResponse{
		UserName: user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Address:  user.Address,

		PetName:     pet.Name,
		Species:     pet.Species,
		Breed:       pet.Breed,
		Weight:      pet.Weight,
		Color:       pet.Color,
		Sex:         pet.Sex,
		Observation: pet.Observation,
		Age:         pet.Age,
		DeathDate:   pet.DeathDate,

		PickupRequired: order.PickupRequired,
		PickupAddress:  order.PickupAddress,

		ServiceID:   order.ServiceTypeID,
		ServiceName: order.ServiceName,

		Folio:  order.OrderNumber,
		Active: order.Active,
		Price:  order.Price,
		Status: order.Status,
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}

func ViewMyOrders(c *fiber.Ctx) error {
	var (
		authUser      = *c.Locals("user").(*users.User)
		targetUser    users.User
		orders        []Order
		orderResponse []OrderViewProfile
	)

	var targetUserID uint

	if authUser.Role == "admin" {
		id, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid user id",
			})
		}

		targetUserID = uint(id)
	} else {
		targetUserID = authUser.ID
	}

	if err := db.DB.First(&targetUser, targetUserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if err := db.DB.
		Preload("Pet").
		Where("user_id = ?", targetUserID).
		Find(&orders).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	for _, order := range orders {
		orderResponse = append(orderResponse, OrderViewProfile{
			Folio:          order.OrderNumber,
			PetName:        order.Pet.Name,
			Species:        order.Pet.Species,
			DeathDate:      order.Pet.DeathDate,
			Status:         order.Status,
			Price:          order.Price,
			ServiceName:    order.ServiceName,
			PickupRequired: order.PickupRequired,
			CreationDate:   order.CreatedAt.Format("2006-01-02"),
		})
	}

	return c.JSON(fiber.Map{
		"creation_date": targetUser.CreatedAt.Format("2006-01-02"),
		"name":          targetUser.Name,
		"email":         targetUser.Email,
		"phone":         targetUser.Phone,
		"address":       targetUser.Address,
		"orders":        orderResponse,
		"role":          targetUser.Role,
	})
}

func Edit(c *fiber.Ctx) error {

	type StatusEdit struct {
		Status string `json:"status"`
	}

	var (
		order Order
		body  StatusEdit
	)

	folio := c.Params("folio")

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Body inválido",
		})
	}

	if err := db.DB.Where("order_number = ?", folio).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Orden no encontrada",
		})
	}

	order.Status = body.Status

	if err := db.DB.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo actualizar la orden",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Estado actualizado",
	})
}

func Delete(c *fiber.Ctx) error {
	return c.JSON("")
}
