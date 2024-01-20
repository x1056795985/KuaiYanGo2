package db

// 定时任务表
type DB_Cron struct {
	Id      int    `json:"Id" gorm:"column:Id;primarykey;comment:id"`
	Name    string `json:"Name" gorm:"column:Name;comment:任务名称"`
	Status  int    `json:"Status" gorm:"column:Status;comment:状态,1启用,2停用"`
	IsLog   int    `json:"IsLog" gorm:"column:IsLog;comment:是否写出日志,1写日志,2忽略"`
	Cron    string `json:"Cron" gorm:"column:Cron;comment:Cron表达式"`
	Type    int    `json:"Type" gorm:"column:Type;comment:任务类型,1,http请求,2公共js函数,3 SQL 4 shell"`
	RunText string `json:"RunText" gorm:"column:RunText;size:1000;comment:运行数据,get网址,云函数名称(参数),shell命令行"`
	Note    string `json:"Note" gorm:"column:Note;size:1000;comment:备注"`
}

func (DB_Cron) TableName() string {
	return "db_cron" //(定时任务)
}
