package Ser_Admin

import (
	"server/new/app/global"
	dbm "server/new/app/models/db"
)

func Id取User(Id int) string {
	if Id == 0 {
		return ""
	}
	var 用户名 string
	global.GVA_DB.Model(dbm.DB_Admin{}).Select("User").Where("Id=?", Id).Take(&用户名)
	return 用户名
}
