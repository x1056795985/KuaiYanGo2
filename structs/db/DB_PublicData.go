package DB

type DB_PublicData struct {
	AppId int    `json:"AppId" gorm:"column:AppId;primarykey;comment:权限" sql:"type:INT(10) UNSIGNED NOT NULL"` //1=所以有软件公共读变量,2,2系统内部读,3云函数,10000+软件专属读取
	Name  string `json:"Name" gorm:"column:Name;primarykey;size:100;comment:变量名"`                              // 需要支持中文 区分大小写 和 AppId    //不要用 key  会各种问题
	Value string `json:"Value" gorm:"column:Value;type:MEDIUMTEXT;comment:变量值"`
	Type  int    `json:"Type" gorm:"column:Type;comment:数据类型"`                 //数据类型 1=单行文本  2=多行文本  3=开关(逻辑型)
	IsVip int    `json:"IsVip" gorm:"column:IsVip;comment:Vip点数限制;default:0 "` //Vip点数大于值才可以获取
	Time  int    `json:"Time" gorm:"column:Time;comment:添加时间戳主要用来排序; "`
	Note  string `json:"Note" gorm:"column:Note;size:1000;comment:备注; "`
}

func (DB_PublicData) TableName() string {
	return "db_Public_Data" //(公共变量表)
}
