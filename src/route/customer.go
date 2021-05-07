package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"net/http"

	"fmt"

	"github.com/labstack/echo/v4"
)

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
		'` + fmt.Sprintf("%s", request["postcode"]) + `'
		)
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "create fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "create success")
}
