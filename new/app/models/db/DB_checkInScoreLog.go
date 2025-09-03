package db

// 签到分日志
type DB_CheckInScoreLog struct {
	Id           int    `json:"id" gorm:"column:id;primarykey;comment:自增ID"`
	AppId        int    `json:"appId" gorm:"column:appId;comment:appId;"`
	UserId       int    `json:"userId" gorm:"column:userId;comment:用户ID"` // 明确用户ID
	CreatedAt    int64  `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	Number       int64  `json:"number" gorm:"column:number;default:0;comment:变动值"`
	Msg          string `json:"msg" gorm:"column:msg;;comment:消息"`
	NumberBefore int    `json:"numberBefore" gorm:"column:numberBefore;default:0;comment:变动前"`
	NumberAfter  int    `json:"numberAfter" gorm:"column:numberAfter;default:0;comment:变动后"`
}

func (DB_CheckInScoreLog) TableName() string {
	return "db_check_in_score_log"
}
