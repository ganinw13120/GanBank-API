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

func DeleteStaff(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["staff_id"] == nil {
		return echo.NewHTTPError(500, "dont have staff id")
	}

	id := fmt.Sprintf("%s", request["staff_id"])

	db := Service.InitialiedDb()

	err := db.Raw(`
	DELETE FROM Staff WHERE staff_id = ` + id + `
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

func UpdateStaff(c echo.Context) error {
	result := []map[string]interface{}{}

	request := Helper.GetJSONRawBody(c)

	if request["staff_id"] == nil {
		return echo.NewHTTPError(500, "dont have staff id")
	}

	id := fmt.Sprintf("%s", request["staff_id"])

	delete(request, "staff_id")

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
	UPDATE Staff SET ` + condition + ` WHERE staff_id = ` + id + `
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

func CreateStaff(c echo.Context) error {
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
	INSERT INTO Staff (staff_id, branch_id, staff_firstname, staff_middlename, staff_lastname, 
		position_id, staff_phone_number, staff_idcard_number, staff_district_id, staff_address, 
		staff_address_name, staff_status, staff_gender, education_level_id, staff_auth_email, 
		staff_auth_password) 
		VALUES (
			NULL, 
			'` + fmt.Sprintf("%.0f", request["branch_id"]) + `', 
			'` + fmt.Sprintf("%s", request["firstname"]) + `',  
			` + middlename + `,  
			'` + fmt.Sprintf("%s", request["lastname"]) + `', 
			'` + fmt.Sprintf("%.0f", request["position_id"]) + `', 
			'` + fmt.Sprintf("%s", request["phone_number"]) + `', 
			'` + fmt.Sprintf("%s", request["idcard"]) + `', 
			'` + fmt.Sprintf("%.0f", request["district_id"]) + `', 
			'` + fmt.Sprintf("%s", request["address"]) + `', 
			'` + fmt.Sprintf("%s", request["address_name"]) + `', 
			'` + fmt.Sprintf("%s", request["status"]) + `', 
			'` + fmt.Sprintf("%s", request["gender"]) + `', 
			'` + fmt.Sprintf("%.0f", request["education"]) + `', 
			'` + fmt.Sprintf("%s", request["email"]) + `', 
			'` + fmt.Sprintf("%s", request["password"]) + `'
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

func GetPrepareStaff(c echo.Context) error {
	var wg sync.WaitGroup
	start := time.Now()
	db := Service.InitialiedDb()
	result := map[string][]map[string]interface{}{}
	wg.Add(4)
	go GetBranchList(db, result, &wg)
	go GetEducationLevel(db, result, &wg)
	go GetPosition(db, result, &wg)
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

func GetPosition(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT position_id, position_name FROM Position
	`).Scan(&result).Error
	res["Position"] = result
	if err != nil {
		fmt.Println(err)
	}
}

func GetAllStaff(c echo.Context) error {
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT * FROM Staff LEFT JOIN District ON Staff.staff_district_id=District.district_id
	LEFT JOIN Amphur ON District.amphur_id=Amphur.amphur_id
	LEFT JOIN Province ON Amphur.province_id=Province.province_id
	LEFT JOIN Position ON Staff.position_id=Position.position_id
	LEFT JOIN Branch ON Staff.branch_id=Branch.branch_id
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
