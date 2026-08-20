package dbm

type DB_WebUserCouponLog struct {
	Id           int     `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:流水ID"`
	AppId        int     `json:"appId" gorm:"column:appId;index;comment:所属应用ID"`
	CouponId     int     `json:"couponId" gorm:"column:couponId;index;comment:优惠券模板ID"`
	CouponUserId int     `json:"couponUserId" gorm:"column:couponUserId;index;comment:用户优惠券ID"`
	Uid          int     `json:"uid" gorm:"column:uid;index;comment:软件用户UID"`
	EventType    int     `json:"eventType" gorm:"column:eventType;comment:1领取2锁定3释放4使用5过期6作废"`
	PayOrder     string  `json:"payOrder" gorm:"column:payOrder;size:191;index;comment:关联订单号"`
	Amount       float64 `json:"amount" gorm:"column:amount;type:decimal(10,2);default:0;comment:实际优惠金额"`
	Operator     string  `json:"operator" gorm:"column:operator;size:100;comment:操作人或系统"`
	Time         int64   `json:"time" gorm:"column:time;index;comment:操作时间"`
	Note         string  `json:"note" gorm:"column:note;size:500;comment:操作说明"`
}

func (DB_WebUserCouponLog) TableName() string {
	return "db_WebUserCouponLog"
}
