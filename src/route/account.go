package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"sync"
	"time"

	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Transaction struct {
	Transaction_id              int       `gorm:"primaryKey"`
	Transaction_amount          float64   `gorm:"column:transaction_amount"`
	Transaction_account_no_to   int       `gorm:"column:transaction_account_no_to"`
	Transaction_type_id         int       `gorm:"column:transaction_type_id"`
	Transaction_timestamp       time.Time `gorm:"column:transaction_timestamp"`
	Staff_id                    int       `gorm:"column:staff_id"`
	Transaction_executor_name   string    `gorm:"column:transaction_executor_name"`
	Transaction_account_no_from int       `gorm:"column:transaction_account_no_from"`
	Transaction_bank_id_to      int       `gorm:"column:transaction_bank_id_to"`
	Transaction_memo            string    `gorm:"column:transaction_memo"`
	Period_id                   int       `gorm:"column:period_id"`
	Transaction_bank_id_from    int       `gorm:"column:transaction_bank_id_from"`
}

func checkAdequacyMoney(accountNo string, amount float64) bool {

	var result bool
	db := Service.InitialiedDb()
	err := db.Raw(`
	SELECT EXISTS(SELECT * FROM Account WHERE account_balance >= '` + fmt.Sprintf("%g", amount) + `' AND account_no = '` + accountNo + `') 
	`).Find(&result).Error
	if err != nil {
		return false
	}
	fmt.Println(`SELECT EXISTS(SELECT * FROM Account WHERE account_balance >= '` + fmt.Sprintf("%g", amount) + `' AND account_no = '` + accountNo + `') `)
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return result
}

func updateAccountBalance(accountNo string, amount string, db *gorm.DB, wg *sync.WaitGroup, isIncrease bool) {
	defer wg.Done()
	var operator string
	if isIncrease {
		operator = "+"
	} else {
		operator = "-"
	}
	result := map[string]interface{}{}
	err := db.Raw(`
		UPDATE Account SET account_balance = account_balance ` + operator + ` ` + amount + ` WHERE account_no = '` + accountNo + `'
	`).Find(&result).Error
	if err != nil {
		panic(err.Error())
	}
}

func createTrasaction(transfer *Transaction, db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	err := db.Table("Transaction").Create(&transfer).Error
	if err != nil {
		panic(err.Error())
	}
}

func Transfer(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var accountNoFrom, accountNoTo, amount, phone, token string
	accountNoFrom += fmt.Sprintf("%g", request["accountNoFrom"].(float64))
	accountNoTo += fmt.Sprintf("%s", request["accountNoTo"])
	amount += fmt.Sprintf("%g", request["amount"].(float64))
	phone += fmt.Sprintf("%s", request["phone"])
	token += fmt.Sprintf("%s", request["token"])

	if !Helper.CheckCustomerToken(token, phone) {
		return echo.NewHTTPError(500, "token mismatch")
	}

	if !checkAdequacyMoney(accountNoFrom, request["amount"].(float64)) {
		return echo.NewHTTPError(500, "เงินในบัญชีไม่เพียงพอต่อการทำรายการ")
	}

	transfer := Transaction{
		Transaction_amount:          request["amount"].(float64),
		Transaction_account_no_to:   int(request["accountNoTo"].(float64)),
		Transaction_type_id:         2,
		Transaction_timestamp:       time.Now(),
		Transaction_account_no_from: int(request["accountNoFrom"].(float64)),
		Transaction_bank_id_to:      int(request["bank_id_to"].(float64)),
		Transaction_memo:            request["memo"].(string),
		Transaction_bank_id_from:    1,
	}

	var wg sync.WaitGroup
	wg.Add(3)
	db := Service.InitialiedDb()
	go createTrasaction(&transfer, db, &wg)
	go updateAccountBalance(accountNoTo, amount, db, &wg, true)
	go updateAccountBalance(accountNoFrom, amount, db, &wg, false)
	wg.Wait()

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, fmt.Sprintf("%d", transfer.Transaction_id))
}

func GetAccountByID(c echo.Context) error {
	result := map[string]interface{}{}

	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)

	if request["account_no"] == nil {
		return echo.NewHTTPError(500, "dont have account no")
	}

	if request["bank_id"] == nil {
		return echo.NewHTTPError(500, "dont have bank id")
	}

	id := fmt.Sprintf("%s", request["account_no"])
	bank_id := fmt.Sprintf("%g", request["bank_id"].(float64))

	err := db.Raw(`
	SELECT a.account_no, a.account_name, b.bank_name, b.bank_logo, b.bank_color, b.bank_id
	FROM Account AS a
	INNER JOIN Bank AS b
	ON a.bank_id = b.bank_id
	WHERE a.account_no = '` + id + `'
	AND a.bank_id = '` + bank_id + `'
	`).Find(&result).Error

	if err != nil || len(result) == 0 {
		return echo.NewHTTPError(404, "not fond")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}

func PrepareTransaction(c echo.Context) error {
	var wg sync.WaitGroup
	start := time.Now()
	db := Service.InitialiedDb()
	request := Helper.GetJSONRawBody(c)
	var accountNo string
	accountNo += fmt.Sprintf("%s", request["account_no"])
	result := map[string][]map[string]interface{}{}
	wg.Add(2)
	go GetAccountIncomeAndOutcome(db, result, accountNo, &wg)
	go GetAccountTransaction(db, result, accountNo, &wg)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)
}

func GetAccountIncomeAndOutcome(db *gorm.DB, res map[string][]map[string]interface{}, accountNo string, wg *sync.WaitGroup) {
	result := []map[string]interface{}{}
	defer wg.Done()
	err := db.Raw(`
	SELECT (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE (tran.transaction_account_no_from = '` + accountNo + `' AND tran.transaction_type_id != 2 AND tran.transaction_type_id != 3)
		OR tran.transaction_account_no_to = '` + accountNo + `'
		AND MONTH(tran.transaction_timestamp) = MONTH(CURRENT_DATE())
		AND YEAR(tran.transaction_timestamp) = YEAR(CURRENT_DATE())
		   ) AS income_current_month,
		  (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE (tran.transaction_account_no_from = '` + accountNo + `' AND tran.transaction_type_id != 2 AND tran.transaction_type_id != 3)
		OR tran.transaction_account_no_to = '` + accountNo + `'
		   ) AS income_all,
		   (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE tran.transaction_account_no_from = '` + accountNo + `' 
		AND tran.transaction_type_id != 1
		AND MONTH(tran.transaction_timestamp) = MONTH(CURRENT_DATE())
		AND YEAR(tran.transaction_timestamp) = YEAR(CURRENT_DATE())
		   ) AS outcome_current_month,
		   (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE tran.transaction_account_no_from = '` + accountNo + `' 
		AND tran.transaction_type_id != 1
		   ) AS outcome_all
	`).Find(&result).Error
	if err != nil {
		fmt.Println(err)
	}

	res["income_outcome"] = result

}

func GetAccountTransaction(db *gorm.DB, res map[string][]map[string]interface{}, accountNo string, wg *sync.WaitGroup) {
	result := []map[string]interface{}{}
	defer wg.Done()
	err := db.Raw(`
	SELECT tran.transaction_amount, type.transaction_type_name, tran.transaction_timestamp, "from" AS account_way 
	FROM Transaction AS tran
	INNER JOIN TransactionType AS type
	ON tran.transaction_type_id = type.transaction_type_id
	WHERE tran.transaction_account_no_from = '` + accountNo + `'
	AND MONTH(tran.transaction_timestamp) = MONTH(CURRENT_DATE())
	AND YEAR(tran.transaction_timestamp) = YEAR(CURRENT_DATE())
	UNION
	SELECT tran.transaction_amount, type.transaction_type_name, tran.transaction_timestamp, "to" AS account_way 
	FROM Transaction AS tran
	INNER JOIN TransactionType AS type
	ON tran.transaction_type_id = type.transaction_type_id
	WHERE tran.transaction_account_no_to = '` + accountNo + `' 
	AND MONTH(tran.transaction_timestamp) = MONTH(CURRENT_DATE())
	AND YEAR(tran.transaction_timestamp) = YEAR(CURRENT_DATE())
	ORDER BY transaction_timestamp DESC
	`).Find(&result).Error
	if err != nil {
		fmt.Println(err)
	}

	res["transaction"] = result
}

func GetAccountByCustomer(c echo.Context) error {
	result := []map[string]interface{}{}
	request := Helper.GetJSONRawBody(c)
	var token string
	var phoneNumber string
	token += fmt.Sprintf("%s", request["token"])
	phoneNumber += fmt.Sprintf("%s", request["customer_phone_number"])
	db := Service.InitialiedDb()

	if !Helper.CheckCustomerToken(token, phoneNumber) {
		return echo.NewHTTPError(500, "token mismatch")
	}

	err := db.Raw(`
	SELECT account_no, account_name, account_balance, t.account_type_name
	FROM Account AS a
    INNER JOIN AccountType AS t
    ON a.account_type_id = t.account_type_id
	WHERE a.account_no IN (
		SELECT DISTINCT(account_no) FROM AccountOwner WHERE customer_id = (SELECT customer_id FROM Customer WHERE customer_phone_number = '` + phoneNumber + `')
	)
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
