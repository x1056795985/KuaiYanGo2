package DB

// DB_LogAgentOtherFunc 代理操作日志
type DB_LogAgentOtherFunc struct {
	Id        int    `json:"Id" gorm:"column:Id;primarykey"`
	AgentType int    `json:"AgentType" gorm:"column:AgentType;comment:用户类型"` //0 普通用户  1 2 3 级代理  4  管理员
	AgentUid  int    `json:"AgentUid" gorm:"column:AgentUid;comment:代理uid"`
	AppId     int    `json:"AppId" gorm:"column:AppId;comment:应用id"`
	AppUser   string `json:"AppUser" gorm:"column:AppUser;comment:应用用户名,存这里不用连表速度快"`
	AppUserid int    `json:"AppUserid" gorm:"column:AppUserid;comment:应用用户id"`
	Func      int    `json:"Func" gorm:"column:Func;comment:操作功能id"`
	Note      string `json:"Note" gorm:"column:Note;size:5000;comment:其他信息"`
	Time      int64  `json:"Time" gorm:"column:Time;comment:时间"`
	Ip        string `json:"Ip" gorm:"column:Ip;size:200;comment:ip地址"`
}

func (DB_LogAgentOtherFunc) TableName() string {
	return "db_Log_AgentOtherFunc" //(操作日志)
}
