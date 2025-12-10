package db

// 签到日志
type DB_CheckInLog struct {
	Id        int    `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:自增ID"`
	AppId     int    `json:"appId" gorm:"column:appId;comment:appId;uniqueIndex:idx_app_user_day"`
	UserId    int    `json:"userId" gorm:"column:userId;comment:用户ID;uniqueIndex:idx_app_user_day"` // 明确用户ID
	CreatedAt int64  `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt int64  `json:"updatedAt" gorm:"column:updatedAt;comment:更新时间"`
	Day       string `json:"day" gorm:"column:day;size:8;comment:年月日 组成唯一;uniqueIndex:idx_app_user_day"` //联合索引防止多次签到
}

func (DB_CheckInLog) TableName() string {
	return "db_check_in_log"
}
