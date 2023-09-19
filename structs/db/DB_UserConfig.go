package DB

type DB_UserConfig struct {
	AppId      int    `json:"AppId" gorm:"column:AppId;primarykey;comment:归属APPID" sql:"type:INT(10) UNSIGNED NOT NULL"` //10000+软件专属读取
	Uid        int    `json:"Uid" gorm:"column:Uid;comment:软件用户UiD; "`                                                   //考虑了一下还是用Uid,不然无法找到用户名
	User       string `json:"User" gorm:"column:User;index;comment:用户登录名"`
	Name       string `json:"Name" gorm:"column:Name;primarykey;size:100;comment:变量名"` // 需要支持中文 区分大小写 和 AppId    //不要用 key  会各种问题
	Value      string `json:"Value" gorm:"column:Value;size:15000;comment:变量值"`
	Time       int64  `json:"Time" gorm:"column:Time;comment:创建时间,方便排序 "`
	UpdateTime int64  `json:"UpdateTime" gorm:"column:UpdateTime;comment:更新数据时间戳; "`
}

func (DB_UserConfig) TableName() string {
	return "db_UserConfig" //(用户云配置表)
}
