package db

// cps分销订单信息  客户支付后,用于统计cps拉新的客户分成信息
type DB_CpsPayOrder struct {
	Id       int     `json:"Id" gorm:"column:Id;primarykey"`
	PayOrder string  `json:"payOrder" gorm:"column:payOrder;size:191;uniqueIndex;comment:余额充值订单id"` //唯一索引,那个线程插入成功,那个处理
	Time     int64   `json:"time" gorm:"column:time;index;comment:时间"`
	AppId    int     `json:"appId" gorm:"column:appId;comment:应用id"`
	Uid      int     `json:"uid" gorm:"column:uid;comment:充值用户Uid"`
	Rmb      float64 `json:"rmb" gorm:"column:rmb;type:decimal(10,2);default:0;comment:订单实际付款金额"` //不能用订单的实付金额,而是用订单的卡类金额,因为代理有代理调价功能,可能导致实付金额和卡类金额不一致

	InviterId       int     `json:"inviterId" gorm:"column:inviterId;index;comment:邀请人id"`
	InviterDiscount int     `json:"inviterDiscount" gorm:"column:inviterDiscount;default:0;comment:佣金百分比"` //当时的邀请人折扣
	InviterRMB      float64 `json:"inviterRMB" gorm:"column:inviterRMB;type:decimal(10,2);comment:实际佣金"`
	InviterStatus   int     `json:"inviterStatus" gorm:"column:inviterStatus;comment:佣金状态"`

	GrandpaId       int     `json:"grandpaId" gorm:"column:grandpaId;index;comment:上上级邀请人id"`
	GrandpaDiscount int     `json:"grandpaDiscount" gorm:"column:grandpaDiscount;default:0;comment:上上级佣金百分比"` //当时的徒孙订单分成比例
	GrandpaRMB      float64 `json:"grandpaRMB" gorm:"column:grandpaRMB;type:decimal(10,2);comment:上上级实际佣金"`
	GrandpaStatus   int     `json:"grandpaStatus" gorm:"column:grandpaStatus;comment:上上级佣金状态"`
	Note            string  `json:"Note" gorm:"column:Note;size:5000;comment:信息"`
	Extra           string  `json:"extra" gorm:"column:extra;size:1910;comment:额外信息"`
}

func (DB_CpsPayOrder) TableName() string {
	return "db_cps_pay_order"
}
