package Route

import (
	Service "GANBANKING_API/src/service"

	"github.com/labstack/echo/v4"
)

func GetBank(c echo.Context) error {
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT * FROM Bank
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(404, "not fond")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}
