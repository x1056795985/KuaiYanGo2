package db

// 用户活动配置表
type UserPromotionConfig struct {
	Id               int    `json:"id" gorm:"column:id;primarykey;autoIncrement:false;comment:关联用户ID"` //只有账号模式,可用
	Name             string `json:"name" gorm:"column:name;comment:活动名称"`                              //比如教师节活动
	AppId            int    `json:"appId" gorm:"column:appId;index;comment:关联appId"`
	CreateTime       int64  `json:"createTime" gorm:"column:createTime;comment:创建时间戳"`
	UpdateTime       int64  `json:"updateTime" gorm:"column:updateTime;comment:更新时间戳"`
	StartTime        int64  `json:"startTime" gorm:"column:startTime;comment:开始时间戳"` //让活动有即将开始的状态
	EndTime          int64  `json:"EndTime" gorm:"column:EndTime;comment:结束时间戳"`
	Type             int    `json:"type" gorm:"column:type;comment:类型 1-cps 2-签到"`
	TypeAssociatedId int    `json:"typeAssociatedId" gorm:"column:typeAssociatedId;comment:类型关联对应表id"`
}

func (UserPromotionConfig) TableName() string {
	return "db_promotion_config"
}
