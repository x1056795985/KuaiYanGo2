package DB

// DB_LogVipNumber 积分点数变化日志
type DB_LogVipNumber struct {
	Id    int     `json:"Id" gorm:"column:Id;primarykey"`
	User  string  `json:"User" gorm:"column:User;size:200;index;comment:用户名"`
	AppId int     `json:"AppId" gorm:"column:AppId;comment:AppId"`
	Type  int     `json:"Type" gorm:"column:Type;comment:类型1积分2点数3时间"`
	Time  int64   `json:"Time" gorm:"column:Time;comment:时间"`
	Ip    string  `json:"Ip" gorm:"column:Ip;size:200;comment:登录ip地址"`
	Count float64 `json:"Count" gorm:"column:Count;type:decimal(10,2);default:0;comment:数值"`
	Note  string  `json:"Note" gorm:"column:Note;size:1000;comment:消息"`
}

func (DB_LogVipNumber) TableName() string {
	return "db_Log_VipNumber" //(积分点数增减日志)  //因为需要AppId 所以不能喝余额表放在一起
}
