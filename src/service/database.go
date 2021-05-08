package Service

import (
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitialiedDb() *gorm.DB {

	dsn := "root:DatabaseProject1234@tcp(ec2-54-169-70-154.ap-southeast-1.compute.amazonaws.com)/gan_banking?parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	return db
}
