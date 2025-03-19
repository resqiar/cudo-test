package routes

import (
	"cudo-test/internal/services"

	"github.com/gofiber/fiber/v2"
)

func InitMainRoute(app *fiber.App, mainService services.MainService) {
	api := app.Group("api/v1")

	api.Get("/transactions", func(c *fiber.Ctx) error {
		result, err := mainService.GetAllTransactions()
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{
			"data": result,
		})
	})

	api.Get("/fraud-detection", func(c *fiber.Ctx) error {
		mainService.DetectFraud()
		return c.SendStatus(fiber.StatusOK)
	})
}
