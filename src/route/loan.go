package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetPrepareLoan(c echo.Context) error {
	var wg sync.WaitGroup

	start := time.Now()
	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(4)
	go GetLoanType(db, result, &wg)
	go GetGuarantorRelation(db, result, &wg)
	go GetCareer(db, result, &wg)
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

func GetLoanType(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM LoanType
	`).Scan(&result).Error
	res["loan_type"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetGuarantorRelation(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM GuarantorRelation
	`).Scan(&result).Error
	res["relation_list"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func CreateLoan(c echo.Context) error {
	var wg sync.WaitGroup
	account_no := make(chan int)
	loanId := make(chan int)
	request := Helper.GetJSONRawBody(c)
	start := time.Now()
	db := Service.InitialiedDb()
	if !CheckIfCustomerExist(db, fmt.Sprintf("%s", request["phone_number"])) {
		return echo.NewHTTPError(500, "Customer Not Found!")
	}
	result := map[string][]map[string]interface{}{}
	wg.Add(2)
	go CreateLoanAccount(db, request, &wg, account_no)
	go InsertLoan(db, request, &wg, account_no, loanId)
	loan_id := <-loanId
	switch request["person_info"].(type) {
	case interface{}:
		for _, v := range request["person_info"].([]interface{}) {
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
			go InsertGuarantor(db, person_info, &wg, loan_id)
		}
	default:
		fmt.Println("No person infomation...")
	}

	switch request["property_info"].(type) {
	case interface{}:
		for _, v := range request["property_info"].([]interface{}) {
			property_info := make(map[string]string)
			owner_amount := 0
			property_id := make(chan int64)
			for key, val := range v.(map[string]interface{}) {
				switch val.(type) {
				case string:
					if val != nil {
						property_info[key] = fmt.Sprintf("%s", val)
					} else {
						property_info[key] = "NULL"
					}
				case float64:
					property_info[key] = fmt.Sprintf("%.0f", val)
				case []interface{}:
					for _, name := range val.([]interface{}) {
						wg.Add(1)
						go InsertPropertyOwner(db, name.(string), &wg, property_id)
						owner_amount += 1
					}
				}
			}
			wg.Add(1)
			InsertProperty(db, property_info, &wg, property_id, owner_amount, loan_id)
		}
	default:
		fmt.Println("No guarantor infomation...")
	}

	switch request["other_info"].(type) {
	case interface{}:
		for _, v := range request["other_info"].([]interface{}) {
			other_info := make(map[string]string)
			owner_amount := 0
			guarantee_id := make(chan int64)
			for key, val := range v.(map[string]interface{}) {
				switch val.(type) {
				case string:
					if val != nil {
						other_info[key] = fmt.Sprintf("%s", val)
					} else {
						other_info[key] = "NULL"
					}
				case float64:
					other_info[key] = fmt.Sprintf("%.0f", val)
				case []interface{}:
					for _, name := range val.([]interface{}) {
						wg.Add(1)
						go InsertGuaranteeOwner(db, name.(string), &wg, guarantee_id)
						owner_amount += 1
					}
				}
			}
			wg.Add(1)
			InsertGuarantee(db, other_info, &wg, guarantee_id, owner_amount, loan_id)
		}
	default:
		fmt.Println("No guarantee infomation...")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)
}

func InsertLoan(db *gorm.DB, req map[string]interface{}, wg *sync.WaitGroup, account_no <-chan int, loanId chan<- int) {
	defer wg.Done()
	acc_no := <-account_no
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%s", req["amount"]), 64)
	loan_type_id, _ := strconv.ParseInt(fmt.Sprintf("%.0f", req["loan_type"]), 10, 64)
	type Loan struct {
		LoanID       int     `gorm:"primaryKey"`
		Loan_type_id int64   `gorm:"column:loan_type_id"`
		Account_No   int     `gorm:"column:account_no"`
		Loan_amount  float64 `gorm:"column:loan_amount"`
		StartDate    string  `gorm:"column:loan_start_date"`
		StopDate     string  `gorm:"column:loan_end_date"`
		RequestName  string  `gorm:"column:loan_request_name"`
		Purpose      string  `gorm:"column:loan_purpose"`
	}
	new_loan := Loan{
		Loan_type_id: loan_type_id,
		Account_No:   acc_no,
		Loan_amount:  amount,
		StartDate:    fmt.Sprintf("%s", req["start_date"]),
		StopDate:     fmt.Sprintf("%s", req["stop_date"]),
		RequestName:  fmt.Sprintf("%s", req["request_name"]),
		Purpose:      fmt.Sprintf("%s", req["purpose"]),
	}
	err := db.Table("Loan").Create(&new_loan).Error
	loanId <- new_loan.LoanID
	fmt.Println("Insert Loan Complete")
	if err != nil {
		fmt.Println(err)
	}
}
func CheckIfCustomerExist(db *gorm.DB, phone_number string) bool {
	var result bool
	err := db.Raw(`
	SELECT EXISTS(SELECT * FROM Customer WHERE customer_phone_number='` + phone_number + `') 
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	return result
}

func CreateLoanAccount(db *gorm.DB, req map[string]interface{}, wg *sync.WaitGroup, account_no chan<- int) {
	defer wg.Done()
	type Account struct {
		Account_no   int    `gorm:"primaryKey"`
		Account_name string `gorm:"column:account_name"`
		Account_type string `gorm:"column:account_type_id"`
		Branch       string `gorm:"column:branch_id"`
	}
	new_account := Account{
		Account_name: "บัญชีเงินกู้",
		Account_type: "3",
		Branch:       fmt.Sprintf("%.0f", req["branch"]),
	}
	db.Table("Account").Create(&new_account)
	account_no <- new_account.Account_no
	fmt.Println("create bank account success...")
}

func InsertGuarantor(db *gorm.DB, data map[string]string, wg *sync.WaitGroup, loan_id int) {
	defer wg.Done()
	result := []map[string]interface{}{}
	error := db.Raw(`
		INSERT INTO Guarantor (
			guarantor_firstname,
			guarantor_middlename,
			guarantor_lastname,
			guarantor_idcard_number,
			guarantor_prefix,
			career_id,
			guarantor_phone_number,
			guarantor_email,
			guarantor_income,
			guarantor_outcome,
			guarantor_district_id,
			guarantor_address,
			guarantor_address_name,
			guarantor_relation_id,
			loan_id
		)
		VALUES (
			'` + data["firstname"] + `',
			'` + data["middlename"] + `',
			'` + data["lastname"] + `',
			'` + data["idcard"] + `',
			'` + data["prefix"] + `',
			'` + data["career"] + `',
			'` + data["tel"] + `',
			'` + data["email"] + `',
			'` + data["income"] + `',
			'` + data["expenditure"] + `',
			'` + data["district_id"] + `',
			'` + data["address"] + `',
			'` + data["address_name"] + `',
			'` + data["relation"] + `' ,
			'` + fmt.Sprintf("%d", loan_id) + `'
		)
	`).Scan(&result).Error
	if error != nil {
		fmt.Println(error)
	}
	fmt.Println("Insert guarantor...")

}

func InsertPropertyOwner(db *gorm.DB, name string, wg *sync.WaitGroup, property_id <-chan int64) {
	defer wg.Done()
	propertyId := <-property_id
	result := []map[string]interface{}{}
	error := db.Raw(`
		INSERT INTO PropertyOwner (
			property_owner_name,
			property_id
		)
		VALUES (
			'` + name + `',
			'` + fmt.Sprintf("%d", propertyId) + `'
		)
	`).Scan(&result).Error
	if error != nil {
		fmt.Println(error)
	}
	fmt.Println("Insert property owner...")

}
func InsertProperty(db *gorm.DB, req map[string]string, wg *sync.WaitGroup, property_id chan<- int64, owner_amount int, loan_id int) {
	defer wg.Done()
	area, _ := strconv.ParseFloat(fmt.Sprintf("%s", req["area"]), 64)
	district_id, _ := strconv.ParseInt(fmt.Sprintf("%s", req["district_id"]), 10, 64)
	type Property struct {
		PropertyID  int64   `gorm:"primaryKey"`
		Name        string  `gorm:"column:property_name"`
		Area        float64 `gorm:"column:property_area"`
		DistrictId  int64   `gorm:"column:district_id"`
		Address     string  `gorm:"column:property_address"`
		AddressName string  `gorm:"column:property_address_name"`
		LoanID      int     `gorm:"column:loan_id"`
	}
	new_property := Property{
		Name:        req["detail"],
		Area:        area,
		DistrictId:  district_id,
		Address:     fmt.Sprintf("%s", req["address"]),
		AddressName: fmt.Sprintf("%s", req["address_name"]),
		LoanID:      loan_id,
	}
	err := db.Table("Property").Create(&new_property).Error
	for i := 0; i < owner_amount; i++ {
		property_id <- new_property.PropertyID
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Insert property...")
}

func InsertGuaranteeOwner(db *gorm.DB, name string, wg *sync.WaitGroup, guarantee_id <-chan int64) {
	defer wg.Done()
	guaranteeId := <-guarantee_id
	result := []map[string]interface{}{}
	error := db.Raw(`
		INSERT INTO GuaranteeOwner (
			guarantee_owner_name,
			guarantee_id
		)
		VALUES (
			'` + name + `',
			'` + fmt.Sprintf("%d", guaranteeId) + `'
		)
	`).Scan(&result).Error
	if error != nil {
		fmt.Println(error)
	}
	fmt.Println("Insert guarantee owner...")

}
func InsertGuarantee(db *gorm.DB, req map[string]string, wg *sync.WaitGroup, guarantee_id chan<- int64, owner_amount int, loan_id int) {
	defer wg.Done()
	price, _ := strconv.ParseFloat(fmt.Sprintf("%s", req["price"]), 64)
	district_id, _ := strconv.ParseInt(fmt.Sprintf("%s", req["district_id"]), 10, 64)
	type Guarantee struct {
		ID          int64   `gorm:"primaryKey"`
		Name        string  `gorm:"column:guarantee_name"`
		Price       float64 `gorm:"column:guarantee_price"`
		DistrictId  int64   `gorm:"column:district_id"`
		Address     string  `gorm:"column:guarantee_address"`
		AddressName string  `gorm:"column:guarantee_address_name"`
		LoanID      int     `gorm:"column:loan_id"`
	}
	new_guarantee := Guarantee{
		Name:        req["detail"],
		Price:       price,
		DistrictId:  district_id,
		Address:     fmt.Sprintf("%s", req["address"]),
		AddressName: fmt.Sprintf("%s", req["address_name"]),
		LoanID:      loan_id,
	}
	err := db.Table("Guarantee").Create(&new_guarantee).Error
	for i := 0; i < owner_amount; i++ {
		guarantee_id <- new_guarantee.ID
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Insert guarantee...")
}

func GetAllLoan(c echo.Context) error {

	var wg sync.WaitGroup

	start := time.Now()
	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(7)
	go GetLoanList(db, result, &wg)
	go GetLoanCountThisMonth(db, result, &wg)
	go GetLoanCount(db, result, &wg)
	go GetLoanTotalAmount(db, result, &wg)
	go GetLoanTotalAmountThisMonth(db, result, &wg)
	go GetApprovedLoanTotalAmount(db, result, &wg)
	go GetApprovedLoanTotalAmountThisMonth(db, result, &wg)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)

}
func GetLoanList(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Loan LEFT JOIN LoanType ON Loan.loan_type_id=LoanType.loan_type_id ORDER BY loan_id DESC
	`).Scan(&result).Error
	res["loan_list"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetLoanCountThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count FROM Loan WHERE MONTH(loan_timestamp)=Month(now())
	`).Scan(&result).Error
	res["loan_count_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetLoanCount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count FROM Loan
	`).Scan(&result).Error
	res["loan_count"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetLoanTotalAmount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT SUM(loan_amount) as sum FROM Loan
	`).Scan(&result).Error
	res["total_amount"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetLoanTotalAmountThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT SUM(loan_amount) as sum FROM Loan  WHERE MONTH(loan_timestamp)=Month(now())
	`).Scan(&result).Error
	res["total_amount_this_month"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetApprovedLoanTotalAmount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT SUM(loan_amount) as sum FROM Loan WHERE loan_status='accepted'
	`).Scan(&result).Error
	res["approved_total_amount"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetApprovedLoanTotalAmountThisMonth(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT SUM(loan_amount) as sum FROM Loan  WHERE MONTH(loan_timestamp)=Month(now()) AND loan_status='pending'
	`).Scan(&result).Error
	res["unapproved_total_amount"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetDebtor(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(DISTINCT Customer.customer_id) as sum FROM Loan  LEFT JOIN Account ON Loan.account_no=Account.account_no LEFT JOIN Customer ON Customer.customer_id=AccountOwner.customer_id
	`).Scan(&result).Error
	res["unapproved_total_amount"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetLoanByID(c echo.Context) error {

	var wg sync.WaitGroup

	start := time.Now()
	result := map[string][]map[string]interface{}{}

	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)

	if request["loan_id"] == nil {
		return echo.NewHTTPError(500, "loan not found no")
	}

	id := fmt.Sprintf("%s", request["loan_id"])

	wg.Add(6)
	go GetLoanByID_Data(db, result, &wg, id)
	go GetLoanByID_Person(db, result, &wg, id)
	go GetLoanByID_Property_Owner(db, result, &wg, id)
	go GetLoanByID_Property(db, result, &wg, id)
	go GetLoanByID_Guarantee_Owner(db, result, &wg, id)
	go GetLoanByID_Guarantee(db, result, &wg, id)

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	wg.Wait()
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)

}

func GetLoanByID_Data(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, loan_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Loan
	LEFT JOIN LoanType ON Loan.loan_type_id=LoanType.loan_type_id
	WHERE loan_id = '` + loan_id + `'
	`).Scan(&result).Error
	res["data"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetLoanByID_Person(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, loan_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Guarantor 
	LEFT JOIN GuarantorRelation ON Guarantor.guarantor_relation_id=GuarantorRelation.guarantor_relation_id
	LEFT JOIN District ON Guarantor.district_id=District.district_id
	LEFT JOIN Amphur ON District.amphur_id=Amphur.amphur_id
	LEFT JOIN Province ON Amphur.province_id=Province.province_id
	LEFT JOIN Career ON Career.career_id=Guarantor.career_id
	WHERE loan_id='` + loan_id + `'
	`).Scan(&result).Error
	res["person"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetLoanByID_Property(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, loan_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Property 
	LEFT JOIN District ON Property.district_id=District.district_id
	LEFT JOIN Amphur ON District.amphur_id=Amphur.amphur_id
	LEFT JOIN Province ON Amphur.province_id=Province.province_id
	WHERE loan_id='` + loan_id + `'
	`).Scan(&result).Error
	res["property"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetLoanByID_Property_Owner(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, loan_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM PropertyOwner
	LEFT JOIN Property ON Property.property_id=PropertyOwner.property_id
	WHERE Property.loan_id='` + loan_id + `'
	`).Scan(&result).Error
	res["property_owner"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetLoanByID_Guarantee(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, loan_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Guarantee 
	LEFT JOIN District ON Guarantee.district_id=District.district_id
	LEFT JOIN Amphur ON District.amphur_id=Amphur.amphur_id
	LEFT JOIN Province ON Amphur.province_id=Province.province_id
	WHERE loan_id='` + loan_id + `'
	`).Scan(&result).Error
	res["guarantee"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetLoanByID_Guarantee_Owner(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, loan_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM GuaranteeOwner
	LEFT JOIN Guarantee ON Guarantee.guarantee_id=GuaranteeOwner.guarantee_id
	WHERE Guarantee.loan_id='` + loan_id + `'
	`).Scan(&result).Error
	res["guarantee_owner"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func UpdateLoanStatus(c echo.Context) error {
	var wg sync.WaitGroup
	start := time.Now()
	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	request := Helper.GetJSONRawBody(c)
	fmt.Println("%s", request["id"])
	if GetLoanStatus(db, fmt.Sprintf("%s", request["id"])) == "pending" && fmt.Sprintf("%s", request["new_status"]) == "accepted" {
		account_no, amount := GetLoanInfo(db, fmt.Sprintf("%s", request["id"]))
		wg.Add(1)
		go AdjustBalance(db, "+", amount, account_no, &wg)
	}
	err := db.Raw(`
	UPDATE Loan SET loan_status='` + fmt.Sprintf("%s", request["new_status"]) + `' WHERE loan_id='` + fmt.Sprintf("%s", request["id"]) + `'
	`).Scan(&result).Error
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()
	defer fmt.Println(time.Since(start))
	return c.JSON(200, result)
}

func GetLoanStatus(db *gorm.DB, loan_id string) string {
	result := map[string]interface{}{}
	err := db.Raw(`
	SELECT loan_status FROM Loan
	WHERE loan_id='` + loan_id + `' LIMIT 1
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	return result["loan_status"].(string)
}

func GetLoanInfo(db *gorm.DB, loan_id string) (string, float64) {
	result := map[string]interface{}{}
	err := db.Raw(`
	SELECT account_no,loan_amount FROM Loan
	WHERE loan_id='` + loan_id + `' LIMIT 1
	`).Scan(&result).Error
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
	return fmt.Sprintf("%s", result["account_no"]), result["loan_amount"].(float64)
}
