package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"strconv"
	"sync"
	"time"

	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func CreateTransaction(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()
	db = db.Begin()

	var wg sync.WaitGroup
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%s", request["transaction_amount"]), 64)
	sign := "-"
	if fmt.Sprintf("%s", request["transaction_type"]) == "3" && !CheckAvailableCash(db, amount, fmt.Sprintf("%s", request["account_no"])) {
		return echo.NewHTTPError(500, "Insufficient cash amount!")
	}
	if fmt.Sprintf("%s", request["transaction_type"]) == "1" {
		sign = "+"
	}
	fmt.Println("create success")
	wg.Add(1)
	go AdjustBalance(db, sign, amount, fmt.Sprintf("%s", request["account_no"]), &wg)
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
		db.Rollback()
		return echo.NewHTTPError(500, "create fail")
	}
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer db.Commit()
	defer sql.Close()
	return c.String(200, "create success")

}

func CreateTransferTransaction(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()
	db = db.Begin()

	var wg sync.WaitGroup
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%s", request["amount"]), 64)
	excecutor_name := fmt.Sprintf("%s", request["transaction_executor_name"])
	if request["transaction_executor_name"] == nil {
		excecutor_name = "system"
	}
	if !CheckAvailableCash(db, amount, fmt.Sprintf("%s", request["origin_no"])) {
		return echo.NewHTTPError(500, "Insufficient cash amount!")
	}
	wg.Add(2)
	go AdjustBalance(db, "+", amount, fmt.Sprintf("%s", request["dest_no"]), &wg)
	go AdjustBalance(db, "-", amount, fmt.Sprintf("%s", request["origin_no"]), &wg)
	err := db.Raw(`
		INSERT INTO Transaction (
			transaction_amount,
			transaction_account_no_to,
			transaction_account_no_from,
			transaction_bank_id_to,
			transaction_type_id,
			transaction_executor_name
		)
		VALUES (
			'` + fmt.Sprintf("%s", request["amount"]) + `',
			'` + fmt.Sprintf("%s", request["dest_no"]) + `',
			'` + fmt.Sprintf("%s", request["origin_no"]) + `',
			'` + fmt.Sprintf("%.0f", request["bank"]) + `', 2 ,
			'` + excecutor_name + `'
		)
	`).Scan(&result).Error
	if err != nil {
		db.Rollback()
		return echo.NewHTTPError(500, "create fail")
	}
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer db.Commit()
	defer sql.Close()
	return c.String(200, "create success")

}

func CheckAvailableCash(db *gorm.DB, amount float64, accountNo string) bool {
	var result bool
	err := db.Raw(`
	SELECT EXISTS(SELECT * FROM Account WHERE account_balance >= '` + fmt.Sprintf("%g", amount) + `' AND account_no = '` + accountNo + `') 
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	return result
}

func AdjustBalance(db *gorm.DB, sign string, amount float64, account_no string, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
		UPDATE Account SET account_balance=(account_balance` + sign + ` ` + fmt.Sprintf("%f", amount) + `) WHERE account_no='` + account_no + `'
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
}

func GetAllTransaction(c echo.Context) error {
	type Transaction struct {
		ID        int       `gorm:"column:transaction_id"`
		Amount    string    `gorm:"column:transaction_amount"`
		TimeStamp time.Time `gorm:"column:transaction_timestamp"`
		Branch    string    `gorm:"column:branch_id"`
		Typename  string    `gorm:"column:transaction_type_name"`
	}
	result := []Transaction{}
	db := Service.InitialiedDb()
	err := db.Raw(`
		SELECT * FROM Transaction LEFT JOIN TransactionType ON Transaction.transaction_type_id=TransactionType.transaction_type_id ORDER BY transaction_id DESC
	`).Find(&result).Error
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

func GetAllBank(c echo.Context) error {
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT bank_id, bank_name FROM Bank
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
