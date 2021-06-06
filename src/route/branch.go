package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"
	"sync"

	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func DeleteBranch(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["branch_id"] == nil {
		return echo.NewHTTPError(500, "dont have branch id")
	}

	id := fmt.Sprintf("%s", request["branch_id"])

	db := Service.InitialiedDb()

	err := db.Raw(`
	DELETE FROM Branch WHERE branch_id = ` + id + `
	`).Scan(&result).Error

	if err != nil {
		return echo.NewHTTPError(500, "delete fail")
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.String(200, "delete success")

}

func UpdateBranch(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["branch_id"] == nil {
		return echo.NewHTTPError(500, "dont have branch id")
	}

	id := fmt.Sprintf("%s", request["branch_id"])

	delete(request, "branch_id")

	db := Service.InitialiedDb()

	var condition string

	index := 0
	for key, value := range request {

		if index == len(request)-1 || len(request) == 1 {
			condition += fmt.Sprintf("%s = '%s'", key, value)
		} else {
			condition += fmt.Sprintf("%s = '%s',", key, value)
		}

		index++
	}

	err := db.Raw(`
	UPDATE Branch SET ` + condition + ` WHERE branch_id = ` + id + `
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

func CreateBranch(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	err := db.Raw(`
	INSERT INTO Branch (branch_id, branch_name, branch_address, district_id) 
	VALUES (
		NULL, 
		'` + fmt.Sprintf("%s", request["branch_name"]) + `',
		'` + fmt.Sprintf("%s", request["branch_address"]) + `',
		'` + fmt.Sprintf("%s", request["district_id"]) + `'
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

func GetAllBranch(c echo.Context) error {
	result := map[string][]map[string]interface{}{}

	var wg sync.WaitGroup
	db := Service.InitialiedDb()

	wg.Add(5)
	go GetBranchCount(db, result, &wg)
	go GetBranchListConcurrent(db, result, &wg)
	go GetBranchAmount_Transfer(db, result, &wg)
	go GetBranchAmount_Deposit(db, result, &wg)
	go GetBranchAmount_Withdraw(db, result, &wg)
	wg.Wait()
	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return c.JSON(200, result)
}

func GetBranchCount(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count FROM Branch
	`).Scan(&result).Error
	res["branch_list"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetBranchListConcurrent(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT COUNT(*) as count  FROM Branch
	`).Scan(&result).Error
	res["branch_count"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetBranchAmount_Transfer(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT Branch.branch_name, SUM(transaction_amount) as sum  FROM Transaction 
	LEFT JOIN Staff ON Staff.staff_id=Transaction.staff_id
	LEFT JOIN Branch ON Staff.branch_id=Branch.branch_id
	WHERE Transaction.transaction_type_id='2' 
    GROUP BY Branch.branch_id
	ORDER BY SUM(transaction_amount) LIMIT 1
	`).Scan(&result).Error
	res["transfer"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetBranchAmount_Deposit(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT Branch.branch_name, SUM(transaction_amount) as sum  FROM Transaction 
	LEFT JOIN Staff ON Staff.staff_id=Transaction.staff_id
	LEFT JOIN Branch ON Staff.branch_id=Branch.branch_id
	WHERE Transaction.transaction_type_id='1'
    GROUP BY Branch.branch_id
	ORDER BY SUM(transaction_amount) LIMIT 1
	`).Scan(&result).Error
	res["deposit"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetBranchAmount_Withdraw(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT Branch.branch_name, SUM(transaction_amount)  as sum FROM Transaction 
	LEFT JOIN Staff ON Staff.staff_id=Transaction.staff_id
	LEFT JOIN Branch ON Staff.branch_id=Branch.branch_id
	WHERE Transaction.transaction_type_id='3'
    GROUP BY Branch.branch_id
	ORDER BY SUM(transaction_amount) LIMIT 1
	`).Scan(&result).Error
	res["withdraw"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func DeleteBranchByID(c echo.Context) error {
	result := map[string]interface{}{}
	db := Service.InitialiedDb()

	request := Helper.GetJSONRawBody(c)
	err := db.Raw(`
	DELETE FROM Branch  
	WHERE branch_id='` + fmt.Sprintf("%s", request["branch_id"]) + `'
	`).Find(&result).Error
	fmt.Println(`
	DELETE FROM Branch  
	WHERE branch_id='` + fmt.Sprintf("%s", request["branch_id"]) + `'
	`)
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

func EditBranch(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	db := Service.InitialiedDb()

	err := db.Raw(`
	UPDATE Branch SET 
		branch_name='` + fmt.Sprintf("%s", request["branch_name"]) + `',
		branch_address='` + fmt.Sprintf("%s", request["branch_address"]) + `',
		district_id='` + fmt.Sprintf("%s", request["district_id"]) + `'
		WHERE branch_id='` + fmt.Sprintf("%s", request["branch_id"]) + `'
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
