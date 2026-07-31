package db

// 日活月活表
type DB_LogUserActive struct {
	Id         int    `json:"id" gorm:"column:id;primarykey;comment:id"`
	CreatedAt  int64  `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	AppId      int    `json:"appId" gorm:"column:appId;uniqueIndex:uniq_app_date_type,priority:1;comment:应用id"`
	ActiveType int    `json:"activeType" gorm:"column:activeType;uniqueIndex:uniq_app_date_type,priority:3;comment:活跃类型  1=日活 2=月活"`    // 新增字段
	DateStr    string `json:"dateStr" gorm:"column:dateStr;size:10;uniqueIndex:uniq_app_date_type,priority:2;comment:日期字符串 YYYY-MM-DD"` // 日活 2022-03-03 月活 2022-02
	Count      int    `json:"count" gorm:"column:count;comment:总数"`
}

func (DB_LogUserActive) TableName() string {
	return "db_Log_UserActive"
}
