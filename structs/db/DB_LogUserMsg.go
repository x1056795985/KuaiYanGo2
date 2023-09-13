package DB

// DB_LogUserMsg 用户消息日志
type DB_LogUserMsg struct {
	Id           int    `json:"Id" gorm:"column:Id;primarykey"`
	User         string `json:"User" gorm:"column:User;size:200;index;comment:用户"`
	App          string `json:"App" gorm:"column:App;index;size:200;comment:App名称"`
	AppVer       string `json:"AppVer" gorm:"column:AppVer;size:200;comment:App版本"`
	MsgType      int    `json:"MsgType" gorm:"column:MsgType;comment:消息类型"` //  1 其他 2 bug提交  3 投诉建议 4 系统执行错误
	Time         int    `json:"Time" gorm:"column:Time;comment:时间"`
	Ip           string `json:"Ip" gorm:"column:Ip;size:200;comment:ip地址"`
	Note         string `json:"Note" gorm:"column:Note;size:10000;comment:消息"`
	IsReadIsRead bool   `json:"IsRead" gorm:"column:IsRead;comment:是否已阅读"`
}

func (DB_LogUserMsg) TableName() string {
	return "db_Log_UserMsg" //(用户消息类型)
}
