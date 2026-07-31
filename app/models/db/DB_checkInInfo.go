package db

// 签到活动配置表
type DB_CheckInInfo struct {
	Id               int    `json:"id" gorm:"column:id;primarykey;comment:关联活动表id"`
	CreateTime       int64  `json:"createTime" gorm:"column:createTime;comment:创建时间戳"`
	UpdateTime       int64  `json:"updateTime" gorm:"column:updateTime;comment:更新时间戳"`
	ShareGivePoints  int    `json:"shareGivePoints" gorm:"column:shareGivePoints;comment:分享任务赠送签到分,0关闭任务"`
	InviteGivePoints int    `json:"inviteGivePoints" gorm:"column:inviteGivePoints;comment:邀请任务赠送签到分,0关闭任务"`
	CardClassList    string `json:"cardClassList" gorm:"column:cardClassList;type:varchar(5000);comment:可兑换卡类列表"`
}

func (DB_CheckInInfo) TableName() string {
	return "db_check_in_info"
}
