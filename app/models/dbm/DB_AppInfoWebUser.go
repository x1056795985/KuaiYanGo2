package dbm

type DB_AppInfoWebUser struct {
	Id             int    `json:"id" gorm:"column:id;primarykey;autoIncrement:false"` // 关联appInfoAppId
	Status         int    `json:"status" gorm:"column:status;default:2;comment:状态(1>启用,2>停用)"`
	UrlDownload    string `json:"urlDownload"  gorm:"column:urlDownload;size:1000;comment:下载地址"`
	CaptchaLogin   int    `json:"captchaLogin"  gorm:"column:captchaLogin;default:3;comment:登陆防爆破起始次数"`
	CaptchaReg     int    `json:"captchaReg"  gorm:"column:captchaReg;default:2;comment:注册是否要验证码"`          // 1需要 2不需要
	CaptchaSendSms int    `json:"captchaSendSms"  gorm:"column:captchaSendSms;default:1;comment:发短信是否要验证码"` // 1需要 2不需要
	WebUserDomain  string `json:"webUserDomain"  gorm:"column:webUserDomain;default:'';comment:用户中心域名"`     // 网页用户中心地址
	AgentOnlyOrder int    `json:"agentOnlyOrder"  gorm:"column:agentOnlyOrder;default:2;comment:限代理用户充值"`   //1需要 2不需要
	//实名认证(仅账号模式应用有效),状态取值见 constant包 S实名状态_ 常量
	RealNameStatus int `json:"realNameStatus" gorm:"column:realNameStatus;default:0;comment:实名认证(0不启用,1非强制启用,2强制启用预留)"`
	//实名奖励,实名认证成功后直充该卡类,0不赠送
	RealNameGiftKaClassId int `json:"realNameGiftKaClassId" gorm:"column:realNameGiftKaClassId;default:0;comment:实名奖励卡类id,0不赠送"`
}

func (DB_AppInfoWebUser) TableName() string {
	return "db_app_info_web_user"
}
