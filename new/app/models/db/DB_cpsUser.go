package db

type DB_CpsUser struct {
	Id        int   `json:"id" gorm:"column:id;primarykey;autoIncrement:false;comment:关联用户ID"` //只有账号模式,可用
	CreatedAt int64 `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt int64 `json:"updatedAt" gorm:"column:updatedAt;default:CURRENT_TIMESTAMP;comment:更新时间"`
	//CpsCode   string `json:"cpsCode" gorm:"column:cpsCode;size:255;default:'';unique;comment:分销邀请码"`  不再数据库存储,而是通过crc32生成
	Count int64 `json:"count" gorm:"column:count;default:0;comment:有效拉新计数缓存"`
}

func (DB_CpsUser) TableName() string {
	return "db_cps_user"
}
