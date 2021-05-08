package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"

	"fmt"

	"github.com/labstack/echo/v4"
)

func DeleteStaff(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["staff_id"] == nil {
		return echo.NewHTTPError(500, "dont have staff id")
	}

	id := fmt.Sprintf("%s", request["staff_id"])

	db := Service.InitialiedDb()

	err := db.Raw(`
	DELETE FROM Staff WHERE staff_id = ` + id + `
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

func UpdateStaff(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["staff_id"] == nil {
		return echo.NewHTTPError(500, "dont have staff id")
	}

	id := fmt.Sprintf("%s", request["staff_id"])

	delete(request, "staff_id")

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
	UPDATE Staff SET ` + condition + ` WHERE staff_id = ` + id + `
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

func CreateStaff(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	var middlename string

	if request["middlename"] != nil {

		middlename += fmt.Sprintf("'%s'", request["staff_middlename"])
	} else {
		middlename += "NULL"
	}

	err := db.Raw(`
	INSERT INTO Staff (staff_id, branch_id, staff_firstname, staff_middlename, staff_lastname, 
		position_id, staff_phone_number, staff_idcard_number, staff_district_id, staff_address, 
		staff_address_name, staff_status, staff_gender, education_level_id, staff_auth_email, 
		staff_auth_password) 
		VALUES (
			NULL, 
			'` + fmt.Sprintf("%s", request["branch_id"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_firstname"]) + `',  
			'` + middlename + `',  
			'` + fmt.Sprintf("%s", request["staff_lastname"]) + `', 
			'` + fmt.Sprintf("%s", request["position_id"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_phone_number"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_idcard_number"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_district_id"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_address"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_address_name"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_status"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_gender"]) + `', 
			'` + fmt.Sprintf("%s", request["education_level_id"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_auth_email"]) + `', 
			'` + fmt.Sprintf("%s", request["staff_auth_password"]) + `'
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
