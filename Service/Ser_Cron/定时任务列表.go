package Ser_Cron

import (
	"server/global"
	DB "server/structs/db"
	"time"
)

func Corn_在线列表定时删除已过期() {
	if global.GVA_DB != nil {
		//删除已注销并21600秒(6小时)没活动的token
		_ = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Status=2").Where("LastTime<?", time.Now().Unix()-21600).Delete("").RowsAffected
		//fmt.Printf("定时删除已注销并21600秒(6小时)没活动的token:%v\n", 局_数量)
	}
}

func Corn_在线列表定时注销已过期() {
	if global.GVA_DB != nil {
		//注销超时的token
		_ = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Status=1").Where("LastTime+OutTime<?", time.Now().Unix()).Update("Status", 2).RowsAffected
		//fmt.Printf("定时注销已过期:%v\n", 局_数量)
	}
}
