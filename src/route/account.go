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

func GetInfoByQrcode(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var phone, token, qr string
	phone += fmt.Sprintf("%s", request["phone"])
	token += fmt.Sprintf("%s", request["token"])
	qr += fmt.Sprintf("%s", request["qrcode"])

	if !Helper.CheckCustomerToken(token, phone) {
		return echo.NewHTTPError(500, "token mismatch")
	}

	statement := `
		SELECT account_no, bank_id FROM Account WHERE account_no = (SELECT account_id 
		FROM AccountQRCode
		WHERE account_qr_code_ref = '` + qr + `'
		AND created_at > DATE_ADD(CURRENT_TIMESTAMP(), INTERVAL -30 MINUTE)
		ORDER BY created_at DESC
		LIMIT 1)
	`

	db := Service.InitialiedDb()
	result := map[string]interface{}{}
	err := db.Raw(statement).Find(&result).Error
	if err != nil {
		panic(err.Error())
	}
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}

func GetLastedTransfer(c echo.Context) error {
	request := Helper.GetJSONRawBody(c)
	var phone, token, accountNo string
	phone += fmt.Sprintf("%s", request["phone"])
	token += fmt.Sprintf("%s", request["token"])
	accountNo += fmt.Sprintf("%s", request["account_no"])

	if !Helper.CheckCustomerToken(token, phone) {
		return echo.NewHTTPError(500, "token mismatch")
	}

	statement := `
		SELECT a.account_no, a.account_name, a.bank_id 
		FROM Account AS a
		WHERE a.account_no 
		IN (
			SELECT DISTINCT(t.transaction_account_no_to) 
			FROM Transaction AS t 
			WHERE t.transaction_account_no_from = '` + accountNo + `' 
			AND t.transaction_type_id = 2
			AND t.transaction_account_no_to != '` + accountNo + `' 
		)
		LIMIT 10
	`

	result := []map[string]interface{}{}
	db := Service.InitialiedDb()
	err := db.Raw(statement).Find(&result).Error
	if err != nil {
		panic(err.Error())
	}
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
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
	fmt.Println(accountNo)
	fmt.Println(amount)
	fmt.Println(isIncrease)
	fmt.Println()
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
	accountNoTo += fmt.Sprintf("%g", request["accountNoTo"].(float64))
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
	go getAccountIncomeAndOutcome(db, result, accountNo, &wg)
	go getAccountTransaction(db, result, accountNo, &wg)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)
}

func getAccountIncomeAndOutcome(db *gorm.DB, res map[string][]map[string]interface{}, accountNo string, wg *sync.WaitGroup) {
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

func getAccountTransaction(db *gorm.DB, res map[string][]map[string]interface{}, accountNo string, wg *sync.WaitGroup) {
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
	SELECT account_no, account_name, account_balance, t.account_type_name, (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE (tran.transaction_account_no_from = account_no AND tran.transaction_type_id != 2 AND tran.transaction_type_id != 3)
		OR tran.transaction_account_no_to = account_no
		AND MONTH(tran.transaction_timestamp) = MONTH(CURRENT_DATE())
		AND YEAR(tran.transaction_timestamp) = YEAR(CURRENT_DATE())
		   ) AS income_current_month,
		  (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE (tran.transaction_account_no_from =account_no AND tran.transaction_type_id != 2 AND tran.transaction_type_id != 3)
		OR tran.transaction_account_no_to = account_no
		   ) AS income_all,
		   (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE tran.transaction_account_no_from = account_no
		AND tran.transaction_type_id != 1
		AND MONTH(tran.transaction_timestamp) = MONTH(CURRENT_DATE())
		AND YEAR(tran.transaction_timestamp) = YEAR(CURRENT_DATE())
		   ) AS outcome_current_month,
		   (
		SELECT SUM(tran.transaction_amount)
		FROM Transaction AS tran
		WHERE tran.transaction_account_no_from = account_no
		AND tran.transaction_type_id != 1
		   ) AS outcome_all
	   FROM Account AS a
		  INNER JOIN AccountType AS t
		  ON a.account_type_id = t.account_type_id
	   WHERE a.account_no IN (
		SELECT DISTINCT(account_no) FROM AccountOwner WHERE customer_id = (SELECT customer_id FROM Customer WHERE customer_phone_number = '` + phoneNumber + `')
	   )
	   AND a.account_status = 'active'
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

func GetAccountName(c echo.Context) error {
	result := map[string]interface{}{}

	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)

	if request["account_no"] == nil {
		return echo.NewHTTPError(500, "dont have account no")
	}

	id := fmt.Sprintf("%s", request["account_no"])

	err := db.Raw(`
	SELECT account_name FROM Account
	WHERE account_no = '` + id + `'
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

	var wg sync.WaitGroup

	result := map[string][]map[string]interface{}{}

	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)
	isTokenValid, _, branch_id, _ := CheckSessionToken(db, fmt.Sprintf("%s", request["token"]))
	if !isTokenValid {
		return echo.NewHTTPError(500, "Token mismatch")
	}
	wg.Add(7)
	go GetAccountList(db, result, &wg, branch_id)
	go GetSuspendAccountCount(db, result, &wg, branch_id)
	go GetActiveAccountCount(db, result, &wg, branch_id)
	go GetAccountCountThisMonth(db, result, &wg, branch_id)
	go GetAccountCount(db, result, &wg, branch_id)
	go GetMostAccountType(db, result, &wg, branch_id)
	go GetTotalDepositAmount(db, result, &wg, branch_id)
	wg.Wait()
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}

func GetAccountList(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Account
	LEFT JOIN AccountType ON Account.account_type_id=AccountType.account_type_id
	LEFT JOIN Branch ON Account.branch_id=Branch.branch_id
	WHERE Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	ORDER BY account_no DESC
	
	`).Scan(&result).Error
	res["account_list"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetAccountCount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count FROM Account
	WHERE Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	`).Scan(&result).Error
	res["account_count"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetAccountCountThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Account  WHERE MONTH(account_timestamp)=Month(now())
	AND Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	`).Scan(&result).Error
	res["account_count_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetMostAccountType(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT account_type_name, COUNT(*) as count  FROM Account LEFT JOIN AccountType ON Account.account_type_id=AccountType.account_type_id
	WHERE Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	GROUP BY Account.account_type_id ORDER BY COUNT(*) DESC LIMIT 1
	
	`).Scan(&result).Error
	res["most_account_type"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetActiveAccountCount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*)  as count FROM Account  WHERE account_status='active'
	AND Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	`).Scan(&result).Error
	res["active"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetSuspendAccountCount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Account  WHERE account_status='suspended'
	AND Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	`).Scan(&result).Error
	res["suspend"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetTotalDepositAmount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, branch_id int64) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT SUM(account_balance) as sum,AVG(account_balance) as avg FROM Account WHERE account_status='active' AND account_balance>0
	AND Account.branch_id='` + fmt.Sprintf("%d", branch_id) + `'
	`).Scan(&result).Error
	res["deposit"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func CreateAccount(c echo.Context) error {
	var wg sync.WaitGroup
	request := Helper.GetJSONRawBody(c)
	db := Service.InitialiedDb()
	isTokenValid, _, branch_id, _ := CheckSessionToken(db, fmt.Sprintf("%s", request["token"]))
	if !isTokenValid {
		return echo.NewHTTPError(500, "Token mismatch")
	}
	account_number := make(chan int)
	wg.Add(1)
	go CreateBankAccount(db, request, &wg, account_number, branch_id)

	switch request["info"].(type) {
	case interface{}:
		for _, v := range request["info"].([]interface{}) {
			person_info := make(map[string]string)
			for key, val := range v.(map[string]interface{}) {
				switch val.(type) {
				case string:
					if val != nil {
						person_info[key] = fmt.Sprintf("%s", val)
					} else {
						person_info[key] = "NULL"
					}
				case float64:
					person_info[key] = fmt.Sprintf("%.0f", val)
				}
			}
			wg.Add(1)
			go LinkAndCreateCustomer(db, person_info, &wg, account_number)
		}
	default:
		return c.String(500, "Invalid infomation")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}

	wg.Wait()
	defer sql.Close()
	return c.String(200, "create success")
}

func PrepareAccount(c echo.Context) error {
	var wg sync.WaitGroup

	start := time.Now()
	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(5)
	go GetBranchList(db, result, &wg)
	go GetAccountType(db, result, &wg)
	go GetCareer(db, result, &wg)
	go GetEducationLevel(db, result, &wg)
	go Helper.GetProvience(db, result, &wg)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)
}

func GetBranchList(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT branch_id,branch_name FROM Branch
	`).Scan(&result).Error
	res["branch"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func CreateBankAccount(db *gorm.DB, req map[string]interface{}, wg *sync.WaitGroup, account_number chan<- int, branch_id int64) {
	defer wg.Done()
	type Account struct {
		Account_no   int    `gorm:"primaryKey"`
		Account_name string `gorm:"column:account_name"`
		Account_type string `gorm:"column:account_type_id"`
		Branch       int64  `gorm:"column:branch_id"`
	}
	new_account := Account{
		Account_name: fmt.Sprintf("%s", req["account_name"]),
		Account_type: fmt.Sprintf("%.0f", req["account_type_selected"]),
		Branch:       branch_id,
	}
	db.Table("Account").Create(&new_account)
	for k, _ := range req["info"].([]interface{}) {
		account_number <- new_account.Account_no
		fmt.Println(k)
	}
	fmt.Println("create bank account success...")
}

func LinkAndCreateCustomer(db *gorm.DB, req map[string]string, wg *sync.WaitGroup, account_number <-chan int) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
		SELECT * FROM Customer WHERE customer_phone_number='` + fmt.Sprintf("%s", req["phone_number"]) + `'
	`).Scan(&result).Error
	fmt.Println("creating Channel.............")
	customer_id := make(chan string)
	switch {
	case len(result) == 0:
		wg.Add(2)
		go CreateNewCustomer(db, req, wg, customer_id)
		go CreateAccountOwner(db, wg, customer_id, account_number)
	case len(result) > 0:
		wg.Add(2)
		go UpdateCustomerInfo(db, req, wg, customer_id)
		go CreateAccountOwner(db, wg, customer_id, account_number)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func CreateAccountOwner(db *gorm.DB, wg *sync.WaitGroup, customer_id <-chan string, account_number <-chan int) {
	defer wg.Done()
	customer := <-customer_id
	account_no := <-account_number
	result := []map[string]interface{}{}
	err := db.Raw(`
	INSERT INTO AccountOwner (account_no, customer_id) VALUES (
		'` + fmt.Sprintf("%d", account_no) + `',
		'` + customer + `'
	)
	`).Scan(&result).Error
	fmt.Println("link account success")
	if err != nil {
		fmt.Println(err)
	}

}

func UpdateCustomerInfo(db *gorm.DB, req map[string]string, wg *sync.WaitGroup, customer_id chan<- string) {
	defer wg.Done()
	wg.Add(1)
	go FindCustomerByPhoneNumber(db, fmt.Sprintf("%s", req["phone_number"]), wg, customer_id)
	fmt.Println("updating customer...")
	result := []map[string]interface{}{}
	err := db.Raw(`
		UPDATE Customer SET
			customer_firstname='` + fmt.Sprintf("%s", req["firstname"]) + `',
			customer_middlename='` + fmt.Sprintf("%s", req["middlename"]) + `',
			customer_lastname='` + fmt.Sprintf("%s", req["lastname"]) + `',
			career_id='` + fmt.Sprintf("%s", req["career"]) + `',
			customer_income='` + fmt.Sprintf("%s", req["income"]) + `',
			customer_idcard_number='` + fmt.Sprintf("%s", req["idcard"]) + `',
			customer_gender='` + fmt.Sprintf("%s", req["gender"]) + `',
			customer_email='` + fmt.Sprintf("%s", req["email"]) + `',
			customer_status='` + fmt.Sprintf("%s", req["status"]) + `',
			customer_birthdate='` + fmt.Sprintf("%s", req["birthday"]) + `',
			customer_prefix='` + fmt.Sprintf("%s", req["prefix"]) + `',
			education_level_id='` + fmt.Sprintf("%s", req["education"]) + `',
			customer_contract_district_id='` + fmt.Sprintf("%s", req["contract_district_id"]) + `',
			customer_contract_address='` + fmt.Sprintf("%s", req["contract_address"]) + `',
			customer_contract_address_name='` + fmt.Sprintf("%s", req["contract_address_name"]) + `',
			customer_work_district_id='` + fmt.Sprintf("%s", req["work_district_id"]) + `',
			customer_work_address='` + fmt.Sprintf("%s", req["work_address"]) + `',
			customer_work_address_name='` + fmt.Sprintf("%s", req["work_address_name"]) + `'
		WHERE customer_phone_number='` + fmt.Sprintf("%s", req["phone_number"]) + `'
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}

}

func CreateNewCustomer(db *gorm.DB, req map[string]string, wg *sync.WaitGroup, customer_id chan<- string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	fmt.Println("creating customer...")
	err := db.Raw(`
		INSERT INTO Customer (customer_firstname,customer_middlename,customer_lastname,
		career_id,customer_income,customer_idcard_number,customer_gender,customer_email,
		customer_phone_number,customer_status,customer_birthdate,customer_prefix,education_level_id,
		customer_contract_district_id,customer_contract_address,customer_contract_address_name,
		customer_work_district_id,customer_work_address,customer_work_address_name,customer_passcode) VALUES (
			'` + fmt.Sprintf("%s", req["firstname"]) + `',
			'` + fmt.Sprintf("%s", req["middlename"]) + `',
			'` + fmt.Sprintf("%s", req["lastname"]) + `',
			'` + fmt.Sprintf("%s", req["career"]) + `',
			'` + fmt.Sprintf("%s", req["income"]) + `',
			'` + fmt.Sprintf("%s", req["idcard"]) + `',
			'` + fmt.Sprintf("%s", req["gender"]) + `',
			'` + fmt.Sprintf("%s", req["email"]) + `',
			'` + fmt.Sprintf("%s", req["phone_number"]) + `',
			'` + fmt.Sprintf("%s", req["status"]) + `',
			'` + fmt.Sprintf("%s", req["birthday"]) + `',
			'` + fmt.Sprintf("%s", req["prefix"]) + `',
			'` + fmt.Sprintf("%s", req["education"]) + `',
			'` + fmt.Sprintf("%s", req["contract_district_id"]) + `',
			'` + fmt.Sprintf("%s", req["contract_address"]) + `',
			'` + fmt.Sprintf("%s", req["contract_address_name"]) + `',
			'` + fmt.Sprintf("%s", req["work_district_id"]) + `',
			'` + fmt.Sprintf("%s", req["work_address"]) + `',
			'` + fmt.Sprintf("%s", req["work_address_name"]) + `', NULL
		)
	`).Scan(&result).Error
	wg.Add(1)
	go FindCustomerByPhoneNumber(db, fmt.Sprintf("%s", req["phone_number"]), wg, customer_id)
	if err != nil {
		fmt.Println(err)
	}
}

func FindCustomerByPhoneNumber(db *gorm.DB, phone_number string, wg *sync.WaitGroup, customer_id chan<- string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
		SELECT customer_id FROM Customer WHERE customer_phone_number='` + phone_number + `'
	`).Scan(&result).Error
	customer_id <- fmt.Sprintf("%d", result[0]["customer_id"])
	if err != nil {
		fmt.Println(err)
	}
}
func GetAccountType(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT account_type_id, account_type_name FROM AccountType
	`).Scan(&result).Error
	res["account_type"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetCareer(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT career_id,career_name FROM Career
	`).Scan(&result).Error
	res["career"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetEducationLevel(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT education_level_id, education_level_name FROM EducationLevel
	`).Scan(&result).Error
	res["education_level"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func SuspendAccount(c echo.Context) error {
	result := map[string]interface{}{}
	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)
	fmt.Println(fmt.Sprintf("%s", request["id"]))
	err := db.Raw(`
	UPDATE Account SET account_status='suspended' WHERE account_no='` + fmt.Sprintf("%s", request["id"]) + `'
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

func EditAccount(c echo.Context) error {
	result := map[string]interface{}{}
	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)
	err := db.Raw(`
	UPDATE Account SET 
	account_type_id='` + fmt.Sprintf("%.0f", request["account_type_selected"]) + `' , 
	account_name='` + fmt.Sprintf("%s", request["account_name"]) + `' 
	WHERE account_no='` + fmt.Sprintf("%s", request["account_no"]) + `'
	`).Find(&result).Error

	if err != nil {
		return echo.NewHTTPError(404, "เกิดข้อผิดพลาด")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}
