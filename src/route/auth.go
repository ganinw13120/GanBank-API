package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Login(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	result := map[string]interface{}{}

	db := Service.InitialiedDb()
	err := db.Raw(`
		SELECT * FROM Staff WHERE staff_phone_number='` + fmt.Sprintf("%s", request["phone_number"]) + `' 
	`).Scan(&result).Error
	if err != nil {
		return echo.NewHTTPError(404, "not fond")
	}
	sql, err := db.DB()
	defer sql.Close()
	if err != nil {
		panic(err.Error())
	}
	if result["staff_id"] != nil && Helper.ComparePasswords(fmt.Sprintf("%s", result["staff_auth_password"]), fmt.Sprintf("%s", request["password"])) {
		response := make(map[string]string)
		response["level"] = GetStaffPositionLevel(db, fmt.Sprintf("%d", result["staff_id"]))
		response["token"] = GenerateSession(db, fmt.Sprintf("%d", result["staff_id"]))
		response["prefix"], response["firstname"], response["middlename"], response["lastname"] = GetName(db, fmt.Sprintf("%d", result["staff_id"]))
		return echo.NewHTTPError(200, response)
	} else {
		return echo.NewHTTPError(404, "เบอร์โทรหรือรหัสผ่านไม่ถูกต้อง")
	}
}
func CheckSession(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	result := map[string]interface{}{}

	db := Service.InitialiedDb()
	err := db.Raw(`
		SELECT staff_id FROM StaffSession WHERE staff_session_token='` + fmt.Sprintf("%s", request["token"]) + `' AND DATEDIFF(now(),staff_session_timestamp) < 1
	`).Scan(&result).Error
	if err != nil {
		return echo.NewHTTPError(404, "not fond")
	}
	sql, err := db.DB()
	defer sql.Close()
	if err != nil {
		panic(err.Error())
	}
	if result["staff_id"] != nil {
		return echo.NewHTTPError(200, "success")
	} else {
		return echo.NewHTTPError(404, "not fond")
	}
}
func GetStaffPositionLevel(db *gorm.DB, staff_id string) string {
	result := map[string]interface{}{}
	err := db.Raw(`
	SELECT level FROM Staff LEFT JOIN Position ON Staff.position_id=Position.position_id WHERE Staff.staff_id='` + staff_id + `'
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("%s", result["level"])
}

func GenerateSession(db *gorm.DB, staff_id string) string {
	result := map[string]interface{}{}
	token := Helper.HashAndSalt("staff_id")
	err := db.Raw(`
	INSERT INTO StaffSession (staff_session_token,staff_id) VALUES (
		'` + token + `',
		'` + staff_id + `'
	)
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	return token
}

func GetName(db *gorm.DB, staff_id string) (string, string, string, string) {
	result := map[string]interface{}{}
	err := db.Raw(`
	SELECT staff_prefix,staff_firstname, staff_middlename, staff_lastname FROM Staff WHERE Staff.staff_id='` + staff_id + `'
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("%s", result["staff_prefix"]), fmt.Sprintf("%s", result["staff_firstname"]), fmt.Sprintf("%s", result["staff_middlename"]), fmt.Sprintf("%s", result["staff_lastname"])
}
