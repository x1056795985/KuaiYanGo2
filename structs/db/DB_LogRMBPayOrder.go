package DB

// DB_LogRMBPayOrder 制卡日志 余额充值日志
type DB_LogRMBPayOrder struct {
	Id             int     `json:"Id" gorm:"column:Id;primarykey"`
	PayOrder       string  `json:"PayOrder" gorm:"column:PayOrder;size:200;index;comment:余额充值订单id"`
	PayOrder2      string  `json:"PayOrder2" gorm:"column:PayOrder2;size:200;comment:第三方订单id"`
	User           string  `json:"User" gorm:"column:User;size:200;comment:用户名或卡号"`
	Uid            int     `json:"Uid" gorm:"column:Uid;comment:充值用户Uid"`
	UidType        int     `json:"UidType" gorm:"column:UidType;comment:Uid类型"`                      //1 账号,2为卡号 或其他为账号
	Status         int     `json:"Status" gorm:"column:Status;comment:订单状态"`                         // 1  '等待支付'  2  '已付待充' 3 '充值成功' 4 退款中 5 ? 退款失败" : 6退款成功 7 订单关闭
	Type           string  `json:"Type" gorm:"column:Type;size:200;comment:支付类型"`                    //  支付宝PC  微信支付 管理员手动充值 小叮当
	ProcessingType int     `json:"ProcessingType" gorm:"column:ProcessingType;size:20;comment:处理类型"` //  0 余额充值 1 购卡直冲 2 应用积分充值
	Extra          string  `json:"Extra" gorm:"column:Extra;size:2000;comment:额外信息"`                 //购卡直冲 为卡类,
	Rmb            float64 `json:"Rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:充值金额"`
	Time           int64   `json:"Time" gorm:"column:Time;index;comment:时间"`
	Ip             string  `json:"Ip" gorm:"column:Ip;size:200;comment:ip地址"`
	Note           string  `json:"Note" gorm:"column:Note;size:5000;comment:信息"`
	ReceivedUid    int     `json:"ReceivedUid" gorm:"column:ReceivedUid;default:0;index;comment:代收款代理Uid"`
}

func (DB_LogRMBPayOrder) TableName() string {
	return "db_Log_RMBPayOrder"
}
