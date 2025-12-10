package DB

// DB_LogMoney 余额日志
type DB_LogMoney struct {
	Id    int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
	User  string  `json:"User" gorm:"column:User;size:191;index;comment:用户名"`
	Time  int64   `json:"Time" gorm:"column:Time;comment:时间"`
	Ip    string  `json:"Ip" gorm:"column:Ip;size:191;comment:登录ip地址"`
	Count float64 `json:"Count" gorm:"column:Count;type:decimal(10,2);default:0;comment:数值"`
	Note  string  `json:"Note" gorm:"column:Note;type:text;comment:原因"` // 修改此处
}

func (DB_LogMoney) TableName() string {
	return "db_Log_Money" //(钱增减日志)
}
