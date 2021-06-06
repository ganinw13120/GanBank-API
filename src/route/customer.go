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

func HasCustomer(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var phoneNumber string
	phoneNumber += fmt.Sprintf("%s", request["customer_phone_number"])

	var exists bool

	db := Service.InitialiedDb()

	db.Raw(`
	SELECT EXISTS(SELECT customer_phone_number FROM Customer
		WHERE customer_phone_number = '` + phoneNumber + `')
	`).Scan(&exists)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, exists)
}

func HasCustomerSession(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var phoneNumber string
	var token string
	phoneNumber += fmt.Sprintf("%s", request["customer_phone_number"])
	token += fmt.Sprintf("%s", request["token"])

	return c.JSON(200, Helper.CheckCustomerToken(token, phoneNumber))
}

func HasCustomerKey(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var phoneNumber string
	phoneNumber += fmt.Sprintf("%s", request["customer_phone_number"])

	var exists bool

	db := Service.InitialiedDb()

	db.Raw(`
	SELECT EXISTS(SELECT * 
		FROM Customer 
		WHERE customer_passcode IS NOT NULL) 
	`).Scan(&exists)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, exists)
}

func SignoutCustomerSession(c echo.Context) error {
	var result interface{}
	request := Helper.GetJSONRawBody(c)
	var phoneNumber string
	phoneNumber += fmt.Sprintf("%s", request["customer_phone_number"])

	db := Service.InitialiedDb()

	err := db.Raw(`
	UPDATE CustomerSession SET customer_session_status = 'logout' WHERE customer_session_id = 
	(SELECT customer_session_id FROM CustomerSession 
	WHERE customer_id = (SELECT customer_id FROM Customer WHERE customer_phone_number = '` + phoneNumber + `')
	AND customer_session_status = 'login'
	AND customer_session_timestamp > DATE_ADD(CURRENT_TIMESTAMP(), INTERVAL -30 MINUTE)
	ORDER BY customer_session_timestamp DESC
	LIMIT 1)
	`).Scan(&result).Error

	if err != nil {
		panic(err.Error())
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "Success")
}

func CreateCustomerSession(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)

	var token string
	var phoneNumber string
	var inputPasscode string
	token += fmt.Sprintf("%s", request["token"])
	phoneNumber += fmt.Sprintf("%s", request["phoneNumber"])
	inputPasscode += fmt.Sprintf("%s", request["passcode"])

	hashToken := Helper.HashAndSalt(token)
	fmt.Println(token)
	fmt.Println(hashToken)
	var result interface{}
	db := Service.InitialiedDb()

	var customerPasscode string
	err2 := db.Raw(`
	SELECT customer_passcode FROM Customer WHERE customer_phone_number = '` + phoneNumber + `'
	`).Scan(&customerPasscode).Error
	if err2 != nil {
		return echo.NewHTTPError(500, "not found")
	}

	if Helper.ComparePasswords(customerPasscode, inputPasscode) {
		return echo.NewHTTPError(500, "password not correct")
	}

	err := db.Raw(`
	INSERT INTO CustomerSession (customer_session_id, customer_session_token, customer_session_timestamp, customer_session_status, customer_id) 
	VALUES (NULL, '` + hashToken + `', current_timestamp(), 'login', (SELECT customer_id FROM Customer WHERE customer_phone_number = '` + phoneNumber + `'))
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(500, "create fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "Success")

}

func CreateCustomerKey(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var pwd string
	var token string

	var phoneNumber string
	pwd += fmt.Sprintf("%s", request["customer_passcode"])
	phoneNumber += fmt.Sprintf("%s", request["customer_phone_number"])
	token += fmt.Sprintf("%s", request["token"])

	if !Helper.CheckCustomerToken(token, phoneNumber) {
		return echo.NewHTTPError(500, "token mismatch")
	}

	db := Service.InitialiedDb()

	hashPassword := Helper.HashAndSalt(pwd)
	var result interface{}

	err2 := db.Raw(`
	UPDATE Customer SET customer_passcode = '` + hashPassword + `' WHERE customer_phone_number = '` + phoneNumber + `'
	`).Scan(&result).Error

	if err2 != nil {
		return echo.NewHTTPError(500, "create fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "Success")
}

func GetAllCustomer(c echo.Context) error {
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT * FROM Customer LEFT JOIN Career ON Customer.career_id=Career.career_id
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
