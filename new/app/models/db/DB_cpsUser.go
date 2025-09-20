package db

type DB_CpsUser struct {
	Id        int   `json:"id" gorm:"column:id;primarykey;comment:自增ID"`
	AppId     int   `json:"appId" gorm:"column:appId;comment:appId;uniqueIndex:idx_app_user"`
	UserId    int   `json:"userId" gorm:"column:userId;comment:用户ID;uniqueIndex:idx_app_user"` // 明确用户ID
	CreatedAt int64 `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt int64 `json:"updatedAt" gorm:"column:updatedAt;comment:更新时间"`
	//CpsCode   string `json:"cpsCode" gorm:"column:cpsCode;size:255;default:'';unique;comment:分销邀请码"`  不再数据库存储,而是通过crc32生成
	Count         int     `json:"count" gorm:"column:count;default:0;comment:有效拉新计数缓存"`
	CumulativeRMB float64 `json:"cumulativeRMB" gorm:"column:cumulativeRMB;type:decimal(10,2);default:0;comment:累计收入缓存"`
}

func (DB_CpsUser) TableName() string {
	return "db_cps_user"
}
