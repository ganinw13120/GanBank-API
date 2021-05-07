package main

import (
	Route "GANBANKING_API/src/route"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	godotenv.Load()
	e := echo.New()

	mobileGroup := e.Group("/mobile")

	customerGroup := mobileGroup.Group("/customer")
	customerGroup.POST("/create", Route.CreateCustomer)
	customerGroup.POST("/update", Route.UpdateCustomer)

	e.Start(":8080")
}
