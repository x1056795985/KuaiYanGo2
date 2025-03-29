package db

type DB_UniqueNumLog struct {
	Id         int    `json:"Id" gorm:"column:Id;primarykey;comment:id"`
	ItemKey    string `json:"ItemKey" gorm:"column:ItemKey;size:190;UNIQUE;index;comment:uid_唯一标志"` // 索引最大长度767字节 除4 就是191  否则INNODB引擎报错  Specified key wastoo long; max key length is 767 bytes
	CreateTime int64  `json:"CreateTime" gorm:"column:CreateTime;comment:添加时间戳"`
	EndTime    int64  `json:"EndTime" gorm:"column:EndTime;comment:有效期时间戳"`
}

func (DB_UniqueNumLog) TableName() string {
	return "db_unique_num_log"
}
