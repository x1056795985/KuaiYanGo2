package Ser_Log

import (
	"server/new/app/global"
	dbm "server/new/app/models/db"
)

func Y用户消息_取未读数量() int64 {

	var 未读数量 int64
	global.GVA_DB.Model(dbm.DB_LogUserMsg{}).Where("IsRead=0").Count(&未读数量)
	return 未读数量
}
