package dbm

// DB_LogRMBPayOrder 制卡日志 余额充值日志
type DB_LogRMBPayOrder struct {
	Id             int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
	PayOrder       string  `json:"PayOrder" gorm:"column:PayOrder;size:191;index;comment:余额充值订单id"`
	PayOrder2      string  `json:"PayOrder2" gorm:"column:PayOrder2;size:191;comment:第三方订单id"`
	User           string  `json:"User" gorm:"column:User;size:191;comment:用户名或卡号"`
	Uid            int     `json:"Uid" gorm:"column:Uid;comment:充值用户Uid"`
	UidType        int     `json:"UidType" gorm:"column:UidType;comment:Uid类型"`
	Status         int     `json:"Status" gorm:"column:Status;comment:订单状态"`
	Type           string  `json:"Type" gorm:"column:Type;size:191;comment:支付类型"`
	ProcessingType int     `json:"ProcessingType" gorm:"column:ProcessingType;size:20;comment:处理类型"`
	Extra          string  `json:"Extra" gorm:"column:Extra;size:1910;comment:额外信息"`
	Rmb            float64 `json:"Rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:充值金额"`
	CouponUserId   int     `json:"CouponUserId" gorm:"column:CouponUserId;default:0;index;comment:使用的用户优惠券ID"`
	CouponDiscount float64 `json:"CouponDiscount" gorm:"column:CouponDiscount;type:decimal(10,2);default:0;comment:优惠金额"`
	Time           int64   `json:"Time" gorm:"column:Time;index;comment:时间"`
	Ip             string  `json:"Ip" gorm:"column:Ip;size:191;comment:ip地址"`
	Note           string  `json:"Note" gorm:"column:Note;size:5000;comment:信息"`
	ReceivedUid    int     `json:"ReceivedUid" gorm:"column:ReceivedUid;default:0;index;comment:代收款代理Uid"`
}

func (DB_LogRMBPayOrder) TableName() string {
	return "db_Log_RMBPayOrder"
}
