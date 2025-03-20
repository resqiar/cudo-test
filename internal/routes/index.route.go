package routes

import (
	"cudo-test/internal/services"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func InitMainRoute(app *fiber.App, mainService services.MainService) {
	api := app.Group("api/v1")

	api.Get("/fraud-detection", func(c *fiber.Ctx) error {
		// limit query parameter
		limitStr := c.Query("limit")
		limit := 1000
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err != nil || parsedLimit <= 0 {
				return c.JSON(fiber.Map{"error": "limit must be a positive integer"})
			}
			limit = parsedLimit
		}

		// risk_level query parameter
		riskLevelStr := c.Query("risk_level")
		var riskLevels []string
		if riskLevelStr != "" {
			riskLevels = strings.Split(strings.ToLower(riskLevelStr), ",")

			validRiskLevels := map[string]bool{
				"low":    true,
				"medium": true,
				"high":   true,
			}

			for _, rl := range riskLevels {
				if !validRiskLevels[rl] {
					return c.JSON(fiber.Map{
						"error": "risk_level must be low, medium, or high",
					})
				}
			}
		}

		result, err := mainService.DetectFraud(int32(limit), riskLevels)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(result)
	})
}
