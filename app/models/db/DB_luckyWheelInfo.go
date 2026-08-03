package db

// 大转盘活动配置表
type DB_LuckyWheelInfo struct {
	Id             int    `json:"id" gorm:"column:id;primarykey;comment:关联活动表id"`
	CreateTime     int64  `json:"createTime" gorm:"column:createTime;comment:创建时间戳"`
	UpdateTime     int64  `json:"updateTime" gorm:"column:updateTime;comment:更新时间戳"`
	DailyFreeCount int    `json:"dailyFreeCount" gorm:"column:dailyFreeCount;default:1;comment:每日免费领取次数,0关闭"`
	InviteGiveCount int   `json:"inviteGiveCount" gorm:"column:inviteGiveCount;default:1;comment:拉新奖励抽奖次数,0关闭"`
	PrizeList      string `json:"prizeList" gorm:"column:prizeList;type:varchar(5000);comment:奖品列表JSON,卡类id与概率"`
	ThemeColor     string `json:"themeColor" gorm:"column:themeColor;size:20;default:'';comment:转盘主题色"`
}

func (DB_LuckyWheelInfo) TableName() string {
	return "db_lucky_wheel_info"
}
