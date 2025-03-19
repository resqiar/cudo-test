package main

import (
	"cudo-test/configs"
	"cudo-test/gen"
	"cudo-test/internal/repos"
	"cudo-test/internal/routes"
	"cudo-test/internal/services"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// call db
	db := configs.InitDBCon()

	// call sqlc
	repository := gen.New(db)

	app := fiber.New()

	// init repositories
	mainRepo := repos.InitMainRepo(repository)

	// init services
	mainService := services.InitMainService(mainRepo)

	// init routes
	routes.InitMainRoute(app, mainService)

	PORT := os.Getenv("PORT")
	if err := app.Listen(fmt.Sprintf(":%s", PORT)); err != nil {
		log.Fatal(err.Error())
	}
}
