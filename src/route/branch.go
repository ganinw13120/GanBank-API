package Route

import (
	Helper "GANBANKING_API/src/helper"
	Service "GANBANKING_API/src/service"

	"fmt"

	"github.com/labstack/echo/v4"
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
	result := []map[string]interface{}{}

	db := Service.InitialiedDb()

	err := db.Raw(`
	SELECT * FROM Branch LEFT JOIN District ON Branch.district_id=District.district_id
	LEFT JOIN Amphur ON District.amphur_id=Amphur.amphur_id
	LEFT JOIN Province ON Amphur.province_id=Province.province_id
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
