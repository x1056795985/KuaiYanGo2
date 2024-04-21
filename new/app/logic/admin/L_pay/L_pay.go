package L_pay

import (
	"server/global"
	"server/new/app/service"
	"time"
)

func G关闭超时订单() error {
	db := service.S_pay{}
	tx := *global.GVA_DB
	err := db.G关闭超时订单(&tx, time.Now().Unix()-86400) //暂时固定超时1天
	return err
}
