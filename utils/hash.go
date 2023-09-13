package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// BcryptHash 使用 bcrypt 对密码进行md5加密
func BcryptHash(password string) string {
	bytes := Md5String(password)
	return bytes
}

// BcryptCheck 对比明文密码和数据库的哈希值
func BcryptCheck(password, Md5PassWord string) bool {
	return strings.ToUpper(Md5String(password)) == strings.ToUpper(Md5PassWord)

}
func Md5String(str string) string {
	data := []byte(str) //切片
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}
