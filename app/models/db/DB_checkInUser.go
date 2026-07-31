package db

// 签到用户数据
type DB_CheckInUser struct {
	Id            int   `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:自增ID"`
	AppId         int   `json:"appId" gorm:"column:appId;comment:appId;uniqueIndex:idx_app_user"`
	UserId        int   `json:"userId" gorm:"column:userId;comment:用户ID;uniqueIndex:idx_app_user"` // 明确用户ID
	CreatedAt     int64 `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt     int64 `json:"updatedAt" gorm:"column:updatedAt;comment:更新时间"`
	CheckInScore  int   `json:"checkInScore" gorm:"column:checkInScore;default:0;comment:签到分"`
	ContinuousDay int   `json:"continuousDay" gorm:"column:continuousDay;default:0;comment:连续签到天数"`
}

func (DB_CheckInUser) TableName() string {
	return "db_check_in_user"
}
