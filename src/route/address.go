package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"sync"
	"time"

	"fmt"

	"github.com/labstack/echo/v4"
)

func GetProvince(c echo.Context) error {

	var wg sync.WaitGroup

	start := time.Now()
	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(1)
	go Helper.GetProvience(db, result, &wg)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)
}

func GetAmphur(c echo.Context) error {

	var wg sync.WaitGroup

	request := Helper.GetJSONRawBody(c)
	if request["province_id"] == nil {
		return echo.NewHTTPError(404, "Province not found")
	}

	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(1)
	go Helper.GetAmphur(db, result, &wg, fmt.Sprintf("%s", request["province_id"]))

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	return c.JSON(200, result)
}

func GetDistrict(c echo.Context) error {

	var wg sync.WaitGroup

	request := Helper.GetJSONRawBody(c)
	if request["amphur_id"] == nil {
		return echo.NewHTTPError(404, "Province not found")
	}

	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(1)
	go Helper.GetDistrict(db, result, &wg, fmt.Sprintf("%s", request["amphur_id"]))

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	return c.JSON(200, result)
}
