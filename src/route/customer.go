package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"

	"fmt"

	"github.com/labstack/echo/v4"
)

func UpdateCustomer(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["customer_id"] == nil {
		return echo.NewHTTPError(500, "dont have customer id")
	}

	id := fmt.Sprintf("%g", request["customer_id"].(float64))

	delete(request, "customer_id")

	db := Service.InitialiedDb()

	var condition string

	index := 0
	for key, value := range request {
		valueType := fmt.Sprintf("%T", value)

		if index == len(request)-1 || len(request) == 1 {
			if valueType == "float64" {
				condition += fmt.Sprintf("%s = '%g'", key, value.(float64))
			} else {
				condition += fmt.Sprintf("%s = '%s'", key, value)
			}

		} else {
			if valueType == "float64" {
				condition += fmt.Sprintf("%s = '%g',", key, value.(float64))
			} else {
				condition += fmt.Sprintf("%s = '%s',", key, value)
			}
		}

		index++
	}

	err := db.Raw(`
	UPDATE Customer SET ` + condition + ` WHERE customer_id = ` + id + `
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

func CreateCustomer(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	var middlename string

	if request["middlename"] != nil {

		middlename += fmt.Sprintf("'%s'", request["middlename"])
	} else {
		middlename += "NULL"
	}

	err := db.Raw(`
	INSERT INTO Customer (customer_id, customer_firstname, customer_middlename, customer_lastname, career_id, customer_income,
		 customer_idcard_number, customer_gender, customer_email, customer_phone_number, customer_status, customer_birthdate,
		  customer_prefix, education_level_id, customer_contract_district_id, customer_contract_address, customer_contract_address_name,
		   customer_work_district_id, customer_work_address, customer_work_address_name, customer_passcode)
	VALUES (NULL,
		'` + fmt.Sprintf("%s", request["fistname"]) + `',
		` + middlename + `,
		'` + fmt.Sprintf("%s", request["lastname"]) + `',
		'` + fmt.Sprintf("%g", request["carrer_id"].(float64)) + `',
		'` + fmt.Sprintf("%g", request["income"].(float64)) + `',
		'` + fmt.Sprintf("%s", request["idcard_number"]) + `',
		'` + fmt.Sprintf("%s", request["gender"]) + `',
		'` + fmt.Sprintf("%s", request["email"]) + `',
		'` + fmt.Sprintf("%s", request["phone"]) + `',
		'` + fmt.Sprintf("%s", request["status"]) + `',
		'` + fmt.Sprintf("%s", request["birthdate"]) + `',
		'` + fmt.Sprintf("%s", request["prefix"]) + `',
		'` + fmt.Sprintf("%g", request["education_level_id"].(float64)) + `',
		'` + fmt.Sprintf("%g", request["contract_district_id"].(float64)) + `',
		'` + fmt.Sprintf("%s", request["contract_address"]) + `',
		'` + fmt.Sprintf("%s", request["contract_address_name"]) + `',
		'` + fmt.Sprintf("%g", request["work_district_id"].(float64)) + `',
		'` + fmt.Sprintf("%s", request["work_address"]) + `',
		'` + fmt.Sprintf("%s", request["work_address_name"]) + `',
		'` + fmt.Sprintf("%s", request["passcode"]) + `'
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
