package DB

// DB_LogLogin 登录日志
type DB_LogLogin struct {
	Id        int    `json:"Id" gorm:"column:Id;primarykey"`
	User      string `json:"User" gorm:"column:User;index;size:191;comment:用户名"`
	Time      int64  `json:"Time" gorm:"column:Time;comment:时间"`
	Ip        string `json:"Ip" gorm:"column:Ip;size:191;comment:登录ip地址"`
	LoginType int    `json:"LoginType" gorm:"column:LoginType;comment:登录用户类型"` //0 普通用户  1 2 3 级代理  4  管理员 5 web用户中心
	Note      string `json:"Note" gorm:"column:Note;comment:消息"`               //软件登录代号 代理平台登录)
}

func (DB_LogLogin) TableName() string {
	return "db_Log_Login" //(登录日志)
}
