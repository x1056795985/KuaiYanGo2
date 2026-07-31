package db

// 绑定信息日志表
type DB_LogKey struct {
	Id     int     `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT"`
	Type   int     `json:"type" gorm:"column:type;comment:1 绑定,2解绑,3换绑"`
	User   string  `json:"user" gorm:"column:user;size:191;index;comment:用户名"`
	Uid    int     `json:"uid" gorm:"column:uid;comment:uid;index:idx_uid_app_id"`
	AppId  int     `json:"appId" gorm:"column:appId;comment:AppId;index:idx_uid_app_id"`
	OldKey string  `json:"oldKey" gorm:"column:oldKey;size:191;comment:旧绑定信息"`
	NewKey string  `json:"newKey" gorm:"column:newKey;size:191;comment:旧绑定信息"`
	Time   int64   `json:"time" gorm:"column:time;comment:时间"`
	Ip     string  `json:"ip" gorm:"column:ip;size:191;comment:登录ip地址"`
	Count  float64 `json:"count" gorm:"column:count;type:decimal(10,2);default:0;comment:数值"`
	Note   string  `json:"note" gorm:"column:note;size:1000;comment:消息"`
}

func (DB_LogKey) TableName() string {
	return "db_log_key"
}
