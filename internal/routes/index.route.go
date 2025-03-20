package routes

import (
	"cudo-test/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func InitMainRoute(app *fiber.App, mainService services.MainService) {
	api := app.Group("api/v1")

	api.Get("/fraud-detection", func(c *fiber.Ctx) error {
		userIDStr := c.Query("user_id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
		}

		result := mainService.DetectFraud(userID)

		return c.JSON(fiber.Map{
			"user_id":         userID,
			"frequency_score": result.FrequencyScore,
			"amount_score":    result.AmountScore,
			"pattern_score":   result.PatternScore,
			"final_score":     result.FinalScore,
			"risk_level":      result.RiskLevel,
		})
	})
}
