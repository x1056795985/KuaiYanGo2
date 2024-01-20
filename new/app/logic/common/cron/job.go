package cron

import (
	"server/new/app/models/db"
	"time"
)

type Job struct {
	R任务数据 db.DB_Cron
	H函数   func(int64, db.DB_Cron)
}

// 携带参数执行
func (j Job) Run() {
	局_时间戳 := time.Now().Unix()
	//hash := utils.Md5String(strconv.Itoa(j.R任务数据.Id) + strconv.Itoa(int(局_时间戳)))
	j.H函数(局_时间戳, j.R任务数据)
}
