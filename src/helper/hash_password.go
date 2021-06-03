package Helper

import (
	Service "GANBANKING_API/src/service"

	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(pwd string) string {
	bytePassword := []byte(pwd)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, 4)
	if err != nil {
		panic(err.Error())
	}
	return string(hash)
}

func ComparePasswords(hashedPwd string, plainPwd string) bool {
	bytePlainPassword := []byte(plainPwd)
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePlainPassword)
	return err == nil
}

func CheckCustomerToken(token string, phoneNumber string) bool {
	db := Service.InitialiedDb()

	var hashToken string

	err := db.Raw(`
	SELECT customer_session_token FROM CustomerSession 
	WHERE customer_id = (SELECT customer_id FROM Customer WHERE customer_phone_number = '` + phoneNumber + `')
	AND customer_session_status = 'login'
	ORDER BY customer_session_timestamp DESC
	LIMIT 1
	`).Scan(&hashToken).Error

	if err != nil || !ComparePasswords(hashToken, token) {
		return false
	}

	sql, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer sql.Close()

	return true
}
