package DB

type DB_LinksToken struct {
	Id               int    `json:"Id" gorm:"column:Id;primarykey"`                              // id
	Uid              int    `json:"Uid" gorm:"column:Uid;index;comment:用户唯一id"`                  // uid  user id
	User             string `json:"User" gorm:"column:User;index;size:200;comment:用户登录名"`        // 用户登录名
	Status           int    `json:"Status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"` // 1正常 2冻结
	Token            string `json:"-" gorm:"column:Token;index;size:200;comment:令牌"`
	LoginAppid       int    `json:"LoginAppid" gorm:"column:LoginAppid;comment:登录位置id"` //1管理员后台 2代理后天  3WebApi令牌
	AppVer           string `json:"AppVer" gorm:"column:AppVer;size:128;comment:软件版本"`
	LoginTime        int64  `json:"LoginTime" gorm:"column:LoginTime;comment:登录时间"`
	LastTime         int64  `json:"LastTime" gorm:"column:LastTime;comment:最上次活动时间戳"`
	OutTime          int    `json:"OutTime" gorm:"column:OutTime;comment:上次活动时间+退出时间=退出时间戳"`
	Key              string `json:"Key" gorm:"column:Key;comment:绑定信息"` // key是关键字,坑点, 下次重构处理吧
	Tab              string `json:"Tab" gorm:"column:Tab;size:5000;comment:标记,动态标签软件版本之类的信息"`
	VerificationCode string `json:"-" gorm:"column:VerificationCode;comment:存储验证码信息"`
	Ip               string `json:"Ip" gorm:"column:Ip;size:200;comment:ip地址"`
	IPCity           string `json:"IPCity" gorm:"column:IPCity;size:500;comment:ip信息"`
	RiskControl      int    `json:"RiskControl" gorm:"column:RiskControl;comment:风控分数分数越高用户越可疑 "`
	CryptoKeyAes     string `json:"-" gorm:"column:CryptoKeyAes;size:24;comment:存储通讯KeyAes"`
	LogoutCode       int    `json:"LogoutCode" gorm:"column:LogoutCode;comment:注销原因代号"`
	AgentUid         int    `json:"AgentUid" gorm:"column:AgentUid;comment:代理标志,代理uid"`
}

func (DB_LinksToken) TableName() string {
	return "db_links_Token" //(在线用户令牌表)
}
