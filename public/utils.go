package public

import (
	"github.com/joexu01/container-dispatcher/log"
	"golang.org/x/crypto/bcrypt"
)

func GeneratePwdHash(pwd []byte) (string, error) {
	pwdHash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(pwdHash), nil
}

func ComparePwdAndHash(pwd []byte, hashedPwd string) bool {
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, pwd)
	if err != nil {
		log.Error("Comparing hashed password: %s", err.Error())
		return false
	}
	return true
}
