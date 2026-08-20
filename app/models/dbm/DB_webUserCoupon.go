package dbm

type DB_WebUserCoupon struct {
	Id                  int     `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;comment:优惠券ID"`
	AppId               int     `json:"appId" gorm:"column:appId;index:idx_app_status_time;comment:所属应用ID"`
	Name                string  `json:"name" gorm:"column:name;size:100;comment:优惠券名称"`
	Type                int     `json:"type" gorm:"column:type;comment:1满减券2折扣券"`
	CouponValue         float64 `json:"couponValue" gorm:"column:couponValue;type:decimal(10,2);comment:满减金额或折扣支付比例"`
	MinPayAmount        float64 `json:"minPayAmount" gorm:"column:minPayAmount;type:decimal(10,2);default:0;comment:最低原支付金额"`
	MaxDiscountAmount   float64 `json:"maxDiscountAmount" gorm:"column:maxDiscountAmount;type:decimal(10,2);default:0;comment:折扣券最高减免金额"`
	KaClassIds          string  `json:"kaClassIds" gorm:"column:kaClassIds;size:2000;comment:可用卡类ID JSON数组"`
	TotalCount          int     `json:"totalCount" gorm:"column:totalCount;default:0;comment:总发放量,0不限量"`
	ReceivedCount       int     `json:"receivedCount" gorm:"column:receivedCount;default:0;comment:已领取数量"`
	UsedCount           int     `json:"usedCount" gorm:"column:usedCount;default:0;comment:已使用数量"`
	PerUserReceiveLimit int     `json:"perUserReceiveLimit" gorm:"column:perUserReceiveLimit;default:1;comment:单用户最多领取数量"`
	ReceiveStartTime    int64   `json:"receiveStartTime" gorm:"column:receiveStartTime;comment:领取开始时间"`
	ReceiveEndTime      int64   `json:"receiveEndTime" gorm:"column:receiveEndTime;comment:领取结束时间"`
	UseStartTime        int64   `json:"useStartTime" gorm:"column:useStartTime;comment:使用开始时间"`
	UseEndTime          int64   `json:"useEndTime" gorm:"column:useEndTime;comment:使用结束时间"`
	Status              int     `json:"status" gorm:"column:status;index:idx_app_status_time;default:1;comment:1启用2停用"`
	Note                string  `json:"note" gorm:"column:note;size:500;comment:使用说明"`
	CreateTime          int64   `json:"createTime" gorm:"column:createTime;comment:创建时间"`
	UpdateTime          int64   `json:"updateTime" gorm:"column:updateTime;comment:更新时间"`
}

func (DB_WebUserCoupon) TableName() string {
	return "db_WebUserCoupon"
}
