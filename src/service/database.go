package Service

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// http://ec2-54-169-70-154.ap-southeast-1.compute.amazonaws.com/
func InitialiedDb() *gorm.DB {

	dsn := "root:DatabaseProject1234@tcp(ec2-54-169-70-154.ap-southeast-1.compute.amazonaws.com)/gan_banking?parseTime=true"
	fmt.Print("c")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	fmt.Print("b")

	// dsn := "sqlserver://root:DatabaseProject1234@db.gan-mkk.com:3306?database=gan_banking"
	// db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	return db
}
