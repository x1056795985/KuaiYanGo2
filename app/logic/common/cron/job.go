package cron

import (
	"server/app/models/dbm"
	"time"
)

type Job struct {
	R任务数据 dbm.DB_Cron
	H函数     func(int64, dbm.DB_Cron)
}

// 携带参数执行
func (j Job) Run() {
	局_时间戳 := time.Now().Unix()
	//hash := utils.Md5String(strconv.Itoa(j.R任务数据.Id) + strconv.Itoa(int(局_时间戳)))
	j.H函数(局_时间戳, j.R任务数据)
}
