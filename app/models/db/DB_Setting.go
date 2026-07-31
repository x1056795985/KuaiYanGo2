package db

type DB_Setting struct {
	ItemKey   string `json:"ItemKey" gorm:"column:ItemKey;primarykey;size:190;comment:键名"` // 索引最大长度767字节 除4 就是191  否则INNODB引擎报错  Specified key wastoo long; max key length is 767 bytes
	ItemValue string `json:"ItemValue" gorm:"column:ItemValue;size:16100;comment:键值"`      // utf8mb4 65535/4=(max = 16383) 最多这个值
}

func (DB_Setting) TableName() string {
	return "db_Setting" //(系统设置表)
}
