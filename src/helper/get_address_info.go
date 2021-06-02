package Helper

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

var wg sync.WaitGroup

func GetProvience(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM Province
	`).Scan(&result).Error
	res["Province"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetDistrict(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, amphur_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	err := db.Raw(`
	SELECT * FROM District WHERE amphur_id='` + amphur_id + `'
	`).Scan(&result).Error
	res["District"] = result
	if err != nil {
		fmt.Println(err)
	}
}
func GetAmphur(db *gorm.DB, res map[string][]map[string]interface{}, wg *sync.WaitGroup, province_id string) {
	defer wg.Done()
	result := []map[string]interface{}{}
	fmt.Println(province_id)
	err := db.Raw(`
	SELECT * FROM Amphur WHERE province_id='` + province_id + `'
	`).Scan(&result).Error
	res["Amphur"] = result
	if err != nil {
		fmt.Println(err)
	}
}
