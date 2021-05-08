package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"

	"fmt"

	"github.com/labstack/echo/v4"
)

func DeletePosition(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["position_id"] == nil {
		return echo.NewHTTPError(500, "dont have position id")
	}

	id := fmt.Sprintf("%s", request["position_id"])

	db := Service.InitialiedDb()

	err := db.Raw(`
	DELETE FROM Position WHERE position_id = ` + id + `
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

func UpdatePosition(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["position_id"] == nil {
		return echo.NewHTTPError(500, "dont have position id")
	}

	id := fmt.Sprintf("%s", request["position_id"])

	delete(request, "position_id")

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
	UPDATE Position SET ` + condition + ` WHERE position_id = ` + id + `
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

func CreatePosition(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	err := db.Raw(`
	INSERT INTO Position (position_id, position_name, level) 
	VALUES (
		NULL, 
		'` + fmt.Sprintf("%s", request["position_name"]) + `',
		'` + fmt.Sprintf("%s", request["level"]) + `'
		);
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
