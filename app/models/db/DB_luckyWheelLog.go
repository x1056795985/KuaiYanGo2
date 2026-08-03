package db

// 大转盘抽奖记录
type DB_LuckyWheelLog struct {
	Id          int    `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:自增ID"`
	AppId       int    `json:"appId" gorm:"column:appId;index;comment:appId"`
	UserId      int    `json:"userId" gorm:"column:userId;index;comment:用户ID"`
	CreateTime  int64  `json:"createTime" gorm:"column:createTime;comment:抽奖时间戳"`
	KaClassId   int    `json:"kaClassId" gorm:"column:kaClassId;comment:中奖卡类ID,0为未中奖"`
	KaClassName string `json:"kaClassName" gorm:"column:kaClassName;size:100;comment:中奖卡类名称"`
	Source      int    `json:"source" gorm:"column:source;comment:抽奖来源1每日免费2拉新奖励"`
}

func (DB_LuckyWheelLog) TableName() string {
	return "db_lucky_wheel_log"
}
