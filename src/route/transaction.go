package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"strconv"
	"sync"

	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func CreateTransaction(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()
	db = db.Begin()

	// db.Rollback()

	isvalid := make(chan bool)

	var wg sync.WaitGroup
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%s", request["transaction_amount"]), 64)
	sign := "-"
	if fmt.Sprintf("%s", request["transaction_type"]) == "1" {
		sign = "+"
	}
	// CheckAvailableCash(db, amount, fmt.Sprintf("%s", request["account_no"]), isvalid)
	success := <-isvalid
	fmt.Println(success)
	wg.Add(1)
	go AdjustBalance(db, sign, amount, fmt.Sprintf("%s", request["account_no"]), &wg, isvalid)
	err := db.Raw(`
		INSERT INTO Transaction (
			transaction_amount,
			transaction_account_no_from,
			transaction_type_id,
			transaction_executor_name
		)
		VALUES (
			'` + fmt.Sprintf("%s", request["transaction_amount"]) + `',
			'` + fmt.Sprintf("%s", request["account_no"]) + `',
			'` + fmt.Sprintf("%s", request["transaction_type"]) + `',
			'` + fmt.Sprintf("%s", request["transaction_executor_name"]) + `'
		)
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(500, "create fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	fmt.Println(success)
	if !success {
		// db.Rollback()
		return c.String(200, "create failed for some reason")
	}
	defer db.Commit()
	defer sql.Close()
	return c.String(200, "create success")

}

func AdjustBalance(db *gorm.DB, sign string, amount float64, account_no string, wg *sync.WaitGroup, isvalid chan<- bool) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
		UPDATE Account SET account_balance=(account_balance` + sign + ` ` + fmt.Sprintf("%f", amount) + `)
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
}
