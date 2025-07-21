package db

type DB_TongJiZaiXian struct {
	Id        int   `json:"id" gorm:"column:id;primarykey;comment:关联用户ID"` //只有账号模式,可用
	AppId     int64 `json:"appId" gorm:"column:appId;index;comment:应用id"`
	CreatedAt int64 `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	Count     int64 `json:"count" gorm:"column:count;comment:在线总数"`
}

func (DB_TongJiZaiXian) TableName() string {
	return "db_tong_ji_zai_xian"
}
