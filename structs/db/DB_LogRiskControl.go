package DB

// DB_LogRiskControl 风控日志
type DB_LogRiskControl struct {
	Id   int    `json:"Id" gorm:"column:Id;primarykey"`
	LId  int    `json:"LId" gorm:"column:触发在线id;"`
	User string `json:"User" gorm:"column:User;size:191;index;comment:用户名"`
	Type int    `json:"Type" gorm:"column:Type;comment:风控规则类型"`
	Time int64  `json:"Time" gorm:"column:Time;comment:时间"`
	Ip   string `json:"Ip" gorm:"column:Ip;size:191;comment:ip地址"`
	Note string `json:"Note" gorm:"column:Note;size:10000;comment:风控信息"`
}

func (DB_LogRiskControl) TableName() string {
	return "db_Log_RiskControl" //(用户风控日志)
}
