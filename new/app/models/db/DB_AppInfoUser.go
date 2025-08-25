package db

type DB_AppInfoWebUser struct {
	Id           int `json:"id" gorm:"column:id;primarykey;autoIncrement:false"` // 关联appInfoAppId
	Status       int `json:"status" gorm:"column:status;default:2;comment:状态(1>启用,2>停用)"`
	CaptchaLogin int `json:"captchaLogin"  gorm:"column:captchaLogin;default:3;comment:登陆防爆破起始次数"`
}

func (DB_AppInfoWebUser) TableName() string {
	return "db_app_info_web_user"
}
