package public

import (
	"github.com/joexu01/container-dispatcher/log"
	"golang.org/x/crypto/bcrypt"
	"os"
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

//PathExists 判断一个文件或文件夹是否存在
//输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
