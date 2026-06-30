package db

type DB_RmbWithdraw struct {
	Id           int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
	OrderNo      string  `json:"OrderNo" gorm:"column:OrderNo;size:64;uniqueIndex;comment:提现单号"`
	RequestId    string  `json:"RequestId" gorm:"column:RequestId;size:64;uniqueIndex:idx_uid_request;comment:幂等请求id"`
	Uid          int     `json:"Uid" gorm:"column:Uid;index:idx_uid_status,priority:1;uniqueIndex:idx_uid_request;comment:用户id"`
	User         string  `json:"User" gorm:"column:User;size:191;index;comment:用户名快照"`
	WithdrawType int     `json:"WithdrawType" gorm:"column:WithdrawType;default:1;comment:1代理提现 2用户提现预留"`
	Amount       float64 `json:"Amount" gorm:"column:Amount;type:decimal(10,2);default:0;comment:提现金额"`
	Status       int     `json:"Status" gorm:"column:Status;index:idx_uid_status,priority:2;index:idx_status_create,priority:1;comment:状态"`
	UserNote     string  `json:"UserNote" gorm:"column:UserNote;type:text;comment:用户备注"`
	AdminReply   string  `json:"AdminReply" gorm:"column:AdminReply;type:text;comment:管理员回复"`
	PayeeType    int     `json:"PayeeType" gorm:"column:PayeeType;default:1;comment:1收款码 2收款账号"`
	PayeeQrPath  string  `json:"PayeeQrPath" gorm:"column:PayeeQrPath;size:500;comment:收款码快照"`
	PayeeAccount string  `json:"PayeeAccount" gorm:"column:PayeeAccount;size:191;comment:收款账号"`
	PayeeName    string  `json:"PayeeName" gorm:"column:PayeeName;size:191;comment:收款人姓名"`
	PayeeRawInfo string  `json:"PayeeRawInfo" gorm:"column:PayeeRawInfo;type:text;comment:收款信息快照json"`
	VoucherPath  string  `json:"VoucherPath" gorm:"column:VoucherPath;size:500;comment:付款凭证"`
	RiskTag      string  `json:"RiskTag" gorm:"column:RiskTag;type:text;comment:风险标签json"`
	CreateTime   int64   `json:"CreateTime" gorm:"column:CreateTime;index;index:idx_status_create,priority:2;comment:创建时间"`
	AuditTime    int64   `json:"AuditTime" gorm:"column:AuditTime;default:0;comment:审核时间"`
	PayTime      int64   `json:"PayTime" gorm:"column:PayTime;default:0;comment:付款时间"`
	CancelTime   int64   `json:"CancelTime" gorm:"column:CancelTime;default:0;comment:取消时间"`
	OperatorId   int     `json:"OperatorId" gorm:"column:OperatorId;default:0;comment:最后操作管理员id"`
	OperatorUser string  `json:"OperatorUser" gorm:"column:OperatorUser;size:191;comment:最后操作管理员"`
	UpdateTime   int64   `json:"UpdateTime" gorm:"column:UpdateTime;default:0;comment:更新时间"`
	Ip           string  `json:"Ip" gorm:"column:Ip;size:191;comment:用户提交ip"`
}

func (DB_RmbWithdraw) TableName() string {
	return "db_rmb_withdraw"
}

type DB_RmbWithdrawLog struct {
	Id           int    `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`
	WithdrawId   int    `json:"WithdrawId" gorm:"column:WithdrawId;index;comment:提现单id"`
	OrderNo      string `json:"OrderNo" gorm:"column:OrderNo;size:64;index;comment:提现单号"`
	Uid          int    `json:"Uid" gorm:"column:Uid;index;comment:用户id"`
	BeforeStatus int    `json:"BeforeStatus" gorm:"column:BeforeStatus;default:0;comment:操作前状态"`
	AfterStatus  int    `json:"AfterStatus" gorm:"column:AfterStatus;default:0;comment:操作后状态"`
	Action       int    `json:"Action" gorm:"column:Action;index;comment:操作类型"`
	OperatorId   int    `json:"OperatorId" gorm:"column:OperatorId;default:0;comment:操作人id"`
	OperatorUser string `json:"OperatorUser" gorm:"column:OperatorUser;size:191;comment:操作人"`
	OperatorType int    `json:"OperatorType" gorm:"column:OperatorType;default:1;comment:1用户 2管理员 3系统"`
	Ip           string `json:"Ip" gorm:"column:Ip;size:191;comment:ip"`
	Note         string `json:"Note" gorm:"column:Note;type:text;comment:备注"`
	Time         int64  `json:"Time" gorm:"column:Time;index;comment:时间"`
}

func (DB_RmbWithdrawLog) TableName() string {
	return "db_rmb_withdraw_log"
}
