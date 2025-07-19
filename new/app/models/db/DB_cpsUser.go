package db

type DB_CpsUser struct {
	Id        int    `json:"id" gorm:"column:id;primarykey;autoIncrement:false;comment:关联用户ID"` //只有账号模式,可用
	CreatedAt int64  `json:"createdAt" gorm:"column:created_at;comment:创建时间"`
	UpdatedAt int64  `json:"updatedAt" gorm:"column:updated_at;default:CURRENT_TIMESTAMP;comment:更新时间"`
	CpsCode   string `json:"cpsCode" gorm:"column:cpsCode;size:255;default:'';unique;comment:分销邀请码"`
}

func (DB_CpsUser) TableName() string {
	return "db_cps_user"
}
