package Ser_Admin

import (
	"server/global"
	DB "server/structs/db"
)

func Id取User(Id int) string {
	if Id == 0 {
		return ""
	}
	var 用户名 string
	global.GVA_DB.Model(DB.DB_Admin{}).Select("User").Where("Id=?", Id).Take(&用户名)
	return 用户名
}
