package db

type DB_CpsCode struct {
	Id        int    `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:自增ID"`
	UserId    int    `json:"userId" gorm:"column:userId;comment:用户ID;uniqueIndex:idx_app_user"` // 明确用户ID
	CreatedAt int64  `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt int64  `json:"updatedAt" gorm:"column:updatedAt;comment:更新时间"`
	CpsCode   string `json:"cpsCode" gorm:"column:cpsCode;size:191;uniqueIndex:idx_app_cpsCode;comment:分销邀请码"`
}

func (DB_CpsCode) TableName() string {
	return "db_cps_Code"
}
