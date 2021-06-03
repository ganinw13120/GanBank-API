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
	var wg sync.WaitGroup
	request := Helper.GetJSONRawBody(c)
	account_number := make(chan int)
	db := Service.InitialiedDb()
	wg.Add(1)
	go CreateBankAccount(db, request, &wg, account_number)

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

func CreateBankAccount(db *gorm.DB, req map[string]interface{}, wg *sync.WaitGroup, account_number chan<- int) {
	defer wg.Done()
	type Account struct {
		Account_no   int    `gorm:"primaryKey"`
		Account_name string `gorm:"column:account_name"`
		Account_type string `gorm:"column:account_type_id"`
		Branch       string `gorm:"column:branch_id"`
	}
	// db.Raw(`
	// 	INSERT INTO Account (account_name, account_type_id, branch_id) VALUES (
	// 		'` + fmt.Sprintf("%s", req["account_name"]) + `',
	// 		'` + fmt.Sprintf("%.0f", req["account_type_selected"]) + `',
	// 		'` + fmt.Sprintf("%.0f", req["branch_selected"]) + `'
	// 	)
	// `)
	new_account := Account{
		Account_name: fmt.Sprintf("%s", req["account_name"]),
		Account_type: fmt.Sprintf("%.0f", req["account_type_selected"]),
		Branch:       fmt.Sprintf("%.0f", req["branch_selected"]),
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
