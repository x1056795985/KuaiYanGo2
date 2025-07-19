package db

// App活动配置表
type DB_AppPromotionConfig struct {
	Id               int    `json:"id" gorm:"column:id;primarykey;comment:ID"`
	Name             string `json:"name" gorm:"column:name;comment:活动名称"` //比如教师节活动
	AppId            int    `json:"appId" gorm:"column:appId;index;comment:关联appId"`
	CreateTime       int64  `json:"createTime" gorm:"column:createTime;comment:创建时间戳"`
	UpdateTime       int64  `json:"updateTime" gorm:"column:updateTime;comment:更新时间戳"`
	StartTime        int64  `json:"startTime" gorm:"column:startTime;comment:开始时间戳"` //让活动有即将开始的状态
	EndTime          int64  `json:"endTime" gorm:"column:endTime;comment:结束时间戳"`
	PromotionType    int    `json:"promotionType" gorm:"column:promotionType;comment:类型 1-cps 2-签到"`
	TypeAssociatedId int    `json:"typeAssociatedId" gorm:"column:typeAssociatedId;comment:类型关联对应表id"`
	Sort             int64  `json:"sort" gorm:"column:sort;default:0;comment:排序权重; "`
}

func (DB_AppPromotionConfig) TableName() string {
	return "db_app_promotion_config"
}
