package Ser_Admin

import (
	"errors"
	"fmt"
	"server/global"
	DB "server/structs/db"
	. "server/utils"
)

func Id置新密码(Id int, NewPassWord string) error {
	if Id == 0 {
		return errors.New("id不能为0")
	}

	err := global.GVA_DB.Model(DB.DB_Admin{}).Where("Id = ?", Id).Updates(map[string]interface{}{"PassWord": Md5String(NewPassWord)}).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Id置新密码失败:%v,%v,%v", Id, NewPassWord, err.Error()))
		return errors.New("修改密码失败")
	}
	return nil

}
func Id取User(Id int) string {
	if Id == 0 {
		return ""
	}
	var 用户名 string
	global.GVA_DB.Model(DB.DB_Admin{}).Select("User").Where("Id=?", Id).Take(&用户名)
	return 用户名
}

func User用户名取id(用户名 string) int {
	if 用户名 == "" {
		return 0
	}

	var Id int
	db := *global.GVA_DB
	db.Model(DB.DB_Admin{}).Select("Id").Where("User=?", 用户名).Take(&Id)
	return Id
}
