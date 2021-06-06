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
	isTokenValid, staff_id, _, _ := CheckSessionToken(db, fmt.Sprintf("%s", request["token"]))
	if !isTokenValid {
		return echo.NewHTTPError(500, "Token mismatch")
	}
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
			transaction_executor_name,
			staff_id
		)
		VALUES (
			'` + fmt.Sprintf("%s", request["transaction_amount"]) + `',
			'` + fmt.Sprintf("%s", request["account_no"]) + `',
			'` + fmt.Sprintf("%s", request["transaction_type"]) + `',
			'` + fmt.Sprintf("%s", request["transaction_executor_name"]) + `',
			'` + fmt.Sprintf("%d", staff_id) + `'
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
	isTokenValid, staff_id, _, _ := CheckSessionToken(db, fmt.Sprintf("%s", request["token"]))
	if !isTokenValid {
		return echo.NewHTTPError(500, "Token mismatch")
	}

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
			transaction_executor_name,
			staff_id
		)
		VALUES (
			'` + fmt.Sprintf("%s", request["amount"]) + `',
			'` + fmt.Sprintf("%s", request["dest_no"]) + `',
			'` + fmt.Sprintf("%s", request["origin_no"]) + `',
			'` + fmt.Sprintf("%.0f", request["bank"]) + `', 2 ,
			'` + excecutor_name + `',
			'` + fmt.Sprintf("%d", staff_id) + `'
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
	result := map[string][]map[string]interface{}{}

	var wg sync.WaitGroup
	db := Service.InitialiedDb()

	wg.Add(9)
	go GetTransactionList(db, result, &wg)
	go GetTransactionCountThisMonth(db, result, &wg)
	go GetTransactionCount(db, result, &wg)
	go GetTransactionWithdrawThisMonth(db, result, &wg)
	go GetTransactionWithdraw(db, result, &wg)
	go GetTransactionDepositThisMonth(db, result, &wg)
	go GetTransactionDeposit(db, result, &wg)
	go GetTransactionTransferThisMonth(db, result, &wg)
	go GetTransactionTransfer(db, result, &wg)
	wg.Wait()
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}

func GetTransactionList(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Transaction LEFT JOIN TransactionType ON TransactionType.transaction_type_id=Transaction.transaction_type_id
	`).Scan(&result).Error
	res["transaction_list"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionCount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count FROM Transaction 
	`).Scan(&result).Error
	res["transaction_count"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionCountThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Transaction WHERE MONTH(transaction_timestamp)=Month(now())
	`).Scan(&result).Error
	res["transaction_count_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionTransfer(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Transaction WHERE transaction_type_id='2'
	`).Scan(&result).Error
	res["transfer"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionTransferThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Transaction WHERE MONTH(transaction_timestamp)=Month(now()) AND transaction_type_id='2'
	`).Scan(&result).Error
	res["transfer_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionDeposit(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Transaction WHERE transaction_type_id='1'
	`).Scan(&result).Error
	res["deposit"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionDepositThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Transaction WHERE MONTH(transaction_timestamp)=Month(now()) AND transaction_type_id='1'
	`).Scan(&result).Error
	res["deposit_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionWithdraw(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count FROM Transaction WHERE transaction_type_id='3'
	`).Scan(&result).Error
	res["withdraw"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTransactionWithdrawThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Transaction WHERE MONTH(transaction_timestamp)=Month(now()) AND transaction_type_id='3'
	`).Scan(&result).Error
	res["withdraw_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
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
