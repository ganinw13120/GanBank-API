package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"

	"fmt"

	"github.com/labstack/echo/v4"
)

func GetAccountByID(c echo.Context) error {
	result := map[string]interface{}{}

	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)

	if request["account_no"] == nil {
		return echo.NewHTTPError(500, "dont have account no")
	}

	id := fmt.Sprintf("%s", request["account_no"])

	err := db.Raw(`
	SELECT a.account_no, a.account_name, a.account_balance, a.account_timestamp, t.account_type_name, t.account_type_interest_rate, b.branch_name, c.customer_firstname, c.customer_middlename, c.customer_lastname
	FROM Account AS a
	INNER JOIN AccountType AS t
	ON a.account_type_id = t.account_type_id
	INNER JOIN AccountOwner AS o
	ON o.account_no = a.account_no
	INNER JOIN Customer AS c
	ON c.customer_id = o.customer_id
	INNER JOIN Branch AS b
	ON b.branch_id = a.branch_id
	WHERE a.account_no = '` + id + `'
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

func GetAccount(c echo.Context) error {
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT a.account_no, a.account_name, a.account_balance, a.account_timestamp, t.account_type_name, t.account_type_interest_rate, b.branch_name, c.customer_firstname, c.customer_middlename, c.customer_lastname
	FROM Account AS a
	INNER JOIN AccountType AS t
	ON a.account_type_id = t.account_type_id
	INNER JOIN AccountOwner AS o
	ON o.account_no = a.account_no
	INNER JOIN Customer AS c
	ON c.customer_id = o.customer_id
	INNER JOIN Branch AS b
	ON b.branch_id = a.branch_id
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

func CreateAccount(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	err := db.Raw(`
	INSERT INTO Account (account_no, account_name, account_balance, account_timestamp, account_type_id, branch_id) 
	VALUES (
		NULL, 
		'` + fmt.Sprintf("%s", request["account_name"]) + `',
		'0', 
		current_timestamp(), 
		'` + fmt.Sprintf("%s", request["account_type_id"]) + `',
		'` + fmt.Sprintf("%s", request["branch_id"]) + `'
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
