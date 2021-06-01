package Helper

import "golang.org/x/crypto/bcrypt"

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
