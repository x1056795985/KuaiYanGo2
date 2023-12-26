package db

type DB_Blacklist struct {
	Id      int    `json:"Id" gorm:"column:Id;primarykey;comment:id"`
	AppId   int    `json:"AppId" gorm:"column:AppId;comment:应用id,1为全局"`
	ItemKey string `json:"ItemKey" gorm:"column:ItemKey;size:190;index;comment:黑名单值可以是ip,或绑定信息"` // 索引最大长度767字节 除4 就是191  否则INNODB引擎报错  Specified key wastoo long; max key length is 767 bytes
	Time    int64  `json:"Time" gorm:"column:Time;comment:添加时间戳"`
	Count   int    `json:"Count" gorm:"column:Count;comment:拦截次数"`
	Note    string `json:"Note" gorm:"column:Note;size:1000;comment:备注"`
}

func (DB_Blacklist) TableName() string {
	return "db_blacklist" //(黑名单)
}
