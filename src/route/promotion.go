package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"fmt"

	"github.com/labstack/echo/v4"
)

func GetPromotion(c echo.Context) error {
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT * FROM Promotion
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

func DeletePromotion(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["promotion_id"] == nil {
		return echo.NewHTTPError(500, "dont have promotion id")
	}

	id := fmt.Sprintf("%s", request["promotion_id"])

	db := Service.InitialiedDb()

	err := db.Raw(`
	DELETE FROM Promotion WHERE promotion_id = ` + id + `
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(500, "delete fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "delete success")

}

func UpdatePromotion(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["promotion_id"] == nil {
		return echo.NewHTTPError(500, "dont have promotion id")
	}

	id := fmt.Sprintf("%s", request["promotion_id"])

	delete(request, "promotion_id")

	db := Service.InitialiedDb()

	var condition string

	index := 0
	for key, value := range request {

		if index == len(request)-1 || len(request) == 1 {
			condition += fmt.Sprintf("%s = '%s'", key, value)
		} else {
			condition += fmt.Sprintf("%s = '%s',", key, value)
		}

		index++
	}

	err := db.Raw(`
	UPDATE Promotion SET ` + condition + ` WHERE promotion_id = ` + id + `
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(500, "update fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "update success")

}

func CreatePromotion(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	err := db.Raw(`
	INSERT INTO Promotion (promotion_id, promotion_title, promotion_detail, promotion_image_path) 
	VALUES (
		NULL, 
		'` + fmt.Sprintf("%s", request["promotion_title"]) + `', 
		'` + fmt.Sprintf("%s", request["promotion_detail"]) + `', 
		'` + fmt.Sprintf("%s", request["promotion_image_path"]) + `'
	)
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(500, "create fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "create success")
}
