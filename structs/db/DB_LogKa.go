package DB

// DB_LogKa 卡操作日志
type DB_LogKa struct {
	Id       int    `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
	User     string `json:"User" gorm:"column:User;size:191;index;comment:操作用户名"`
	UserType int    `json:"UserType" gorm:"column:UserType;comment:用户类型"` //0 普通用户  1 2 3 级代理  4  管理员  5 系统自动
	Ka       string `json:"Ka" gorm:"column:Ka;index;size:191;comment:卡号"`
	KaType   int    `json:"KaType" gorm:"column:KaType;comment:操作类型1增2删3改4查"`
	Time     int64  `json:"Time" gorm:"column:Time;comment:时间"`
	Ip       string `json:"Ip" gorm:"column:Ip;size:191;comment:ip地址"`
	Note     string `json:"Note" gorm:"column:Note;comment:信息"`
}

func (DB_LogKa) TableName() string {
	return "db_Log_LogKa" //(制卡日志)
}
