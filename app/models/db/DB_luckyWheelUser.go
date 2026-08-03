package db

// 大转盘用户数据
type DB_LuckyWheelUser struct {
	Id               int    `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:自增ID"`
	AppId            int    `json:"appId" gorm:"column:appId;comment:appId;uniqueIndex:idx_app_user"`
	UserId           int    `json:"userId" gorm:"column:userId;comment:用户ID;uniqueIndex:idx_app_user"`
	CreateTime       int64  `json:"createTime" gorm:"column:createTime;comment:创建时间戳"`
	UpdateTime       int64  `json:"updateTime" gorm:"column:updateTime;comment:更新时间戳"`
	RemainCount      int    `json:"remainCount" gorm:"column:remainCount;default:0;comment:剩余抽奖次数"`
	DailyFreeDate    string `json:"dailyFreeDate" gorm:"column:dailyFreeDate;size:10;default:'';comment:每日免费领取日期标记"`
	DailyFreeUsed    int    `json:"dailyFreeUsed" gorm:"column:dailyFreeUsed;default:0;comment:今日已领取免费次数"`
	TotalInviteCount int    `json:"totalInviteCount" gorm:"column:totalInviteCount;default:0;comment:累计拉新获得次数"`
}

func (DB_LuckyWheelUser) TableName() string {
	return "db_lucky_wheel_user"
}
