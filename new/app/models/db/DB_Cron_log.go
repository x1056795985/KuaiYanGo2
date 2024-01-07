package db

// 定时任务日志表
type DB_Cron_log struct {
	Id         int    `json:"Id" gorm:"column:Id;primarykey;comment:id"`
	CronID     int    `json:"CronID" gorm:"column:CronID;comment:定时任务id"`
	RunTime    int64  `json:"RunTime" gorm:"column:RunTime;comment:定时执行时间戳"`
	Type       int    `json:"Type" gorm:"column:Type;comment:任务类型,1,http请求,2公共js函数,3 shell"`
	RunText    string `json:"RunText" gorm:"column:RunText;size:5000;comment:运行数据,get网址,云函数名称(参数),shell命令行"`
	ReturnText string `json:"ReturnText" gorm:"column:ReturnText;size:5000;comment:任务返回数据"`
}

func (DB_Cron_log) TableName() string {
	return "db_cron_log" //(定时任务日志表)
}
