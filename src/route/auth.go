package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"fmt"

	"github.com/labstack/echo/v4"
)

func Login(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	result := map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
		SELECT * FROM Staff WHERE staff_phone_number='` + fmt.Sprintf("%s", request["phone_number"]) + `' AND staff_auth_password='` + fmt.Sprintf("%s", request["password"]) + `'
	`).Scan(&result).Error
	if result["staff_id"] != nil {
		return echo.NewHTTPError(200, "success")
	} else {
		return echo.NewHTTPError(404, "เบอร์โทรหรือรหัสผ่านไม่ถูกต้อง")
	}
	if err != nil {
		return echo.NewHTTPError(404, "not fond")
	}
	fmt.Println(result)
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}
