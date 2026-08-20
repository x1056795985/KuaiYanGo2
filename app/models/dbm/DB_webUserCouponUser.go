package dbm

type DB_WebUserCouponUser struct {
	Id                int     `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:用户优惠券ID"`
	CouponId          int     `json:"couponId" gorm:"column:couponId;index:idx_coupon_user;comment:优惠券模板ID"`
	AppId             int     `json:"appId" gorm:"column:appId;index:idx_app_uid_status;comment:所属应用ID"`
	Uid               int     `json:"uid" gorm:"column:uid;index:idx_app_uid_status;comment:软件用户UID"`
	CouponName        string  `json:"couponName" gorm:"column:couponName;size:100;comment:优惠券名称快照"`
	CouponType        int     `json:"couponType" gorm:"column:couponType;comment:1满减券2折扣券快照"`
	CouponValue       float64 `json:"couponValue" gorm:"column:couponValue;type:decimal(10,2);comment:面额或折扣快照"`
	MinPayAmount      float64 `json:"minPayAmount" gorm:"column:minPayAmount;type:decimal(10,2);comment:最低原支付金额快照"`
	MaxDiscountAmount float64 `json:"maxDiscountAmount" gorm:"column:maxDiscountAmount;type:decimal(10,2);comment:最高减免快照"`
	KaClassIds        string  `json:"kaClassIds" gorm:"column:kaClassIds;size:2000;comment:可用卡类ID快照JSON"`
	Status            int     `json:"status" gorm:"column:status;index:idx_app_uid_status;default:1;comment:1未使用2已锁定3已使用4已过期5已作废"`
	ReceiveTime       int64   `json:"receiveTime" gorm:"column:receiveTime;comment:领取时间"`
	UseStartTime      int64   `json:"useStartTime" gorm:"column:useStartTime;comment:使用开始时间快照"`
	UseEndTime        int64   `json:"useEndTime" gorm:"column:useEndTime;index;comment:使用结束时间快照"`
	LockTime          int64   `json:"lockTime" gorm:"column:lockTime;comment:订单锁定时间"`
	LockExpireTime    int64   `json:"lockExpireTime" gorm:"column:lockExpireTime;index;comment:锁定到期时间"`
	PayOrder          string  `json:"payOrder" gorm:"column:payOrder;size:191;index;comment:关联支付订单号"`
	UseTime           int64   `json:"useTime" gorm:"column:useTime;comment:实际使用成功时间"`
	DiscountAmount    float64 `json:"discountAmount" gorm:"column:discountAmount;type:decimal(10,2);default:0;comment:实际优惠金额"`
	Source            int     `json:"source" gorm:"column:source;default:1;comment:1用户领取2系统发放"`
	Note              string  `json:"note" gorm:"column:note;size:500;comment:备注"`
}

func (DB_WebUserCouponUser) TableName() string {
	return "db_WebUserCouponUser"
}
