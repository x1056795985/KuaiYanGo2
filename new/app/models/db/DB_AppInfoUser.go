package db

// app配置扩展,user相关添加扩展字段
type DB_AppInfoWebUser struct {
	Id           int `json:"id" gorm:"column:id;primarykey;autoIncrement:false"` // 关联appInfoAppId
	Status       int `json:"status" gorm:"column:status;default:2;comment:状态(1>启用,2>停用)"`
	CaptchaLogin int `json:"captchaLogin"  gorm:"column:captchaLogin;default:3;comment:登陆防爆破起始次数"`

	//下方两个接口, 无法单独设置某个应用是否开启验证码 所以只能全局设置  注册可以用js算法,控制自动化,注册,即使算法被破解影响不大
	//但是发短信的接口强制启动验证码,防止暴力破解, 毕竟发短信是真的花钱,所以更严格一些
	//CaptchaReg   int `json:"captchaReg"  gorm:"column:captchaReg;default:2;comment:注册验证码,1启用,2停用"`
	//CaptchaSendSms int `json:"captchaSendSms"  gorm:"column:captchaSendSms;default:1;comment:发送短信启用验证码,1启用,2停用"`
}

func (DB_AppInfoWebUser) TableName() string {
	return "db_App_Info_User"
}
