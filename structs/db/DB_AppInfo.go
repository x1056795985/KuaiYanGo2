package DB

type DB_AppInfo struct {
	AppId                 int    `json:"AppId" gorm:"column:AppId;primarykey"` // id
	AppWeb                string `json:"AppWeb" gorm:"column:AppWeb;size:200;comment:服务器地址"`
	AppName               string `json:"AppName" gorm:"column:AppName;size:200;comment:应用名称"`
	Status                int    `json:"Status" gorm:"column:Status;default:3;comment:状态(1>停止运营,2>免费模式,3>收费模式)"`
	AppStatusMessage      string `json:"AppStatusMessage" gorm:"column:AppStatusMessage;comment:状态原因"`
	AppVer                string `json:"AppVer"  gorm:"column:AppVer;default:1.0.0;comment:软件版本"`
	RegisterGiveKaClassId int    `json:"RegisterGiveKaClassId"  gorm:"column:RegisterGiveKaClassId;comment:注册赠送卡类id"`

	VerifyKey     int `json:"VerifyKey"  gorm:"column:VerifyKey;default:1;comment:绑定模式"`                 //1 免验证可以换绑 2 免验证禁止换绑 3 验证可以换绑  4 验证禁止换绑
	IsUserKeySame int `json:"IsUserKeySame"  gorm:"column:IsUserKeySame;default:1;comment:绑定信息不同用户可否相同"` //1 不同用户可以相同 2 不同用户不可相同
	UpKeyData     int `json:"UpKeyData"  gorm:"column:UpKeyData;comment:修改绑定key增减值"`

	OutTime            int    `json:"OutTime"  gorm:"column:OutTime;default:1800;comment:心跳超时"`
	UrlHome            string `json:"UrlHome"  gorm:"column:UrlHome;default:https://www.baidu.com/;comment:首页Url"`
	UrlDownload        string `json:"UrlDownload"  gorm:"column:UrlDownload;size:5000;comment:下载地址json"`
	AppGongGao         string `json:"AppGongGao"  gorm:"column:AppGongGao;size:1000;comment:公告"`
	VipData            string `json:"VipData"  gorm:"column:VipData;size:5000;comment: vip可获取json数据"`
	CryptoType         int    `json:"CryptoType"  gorm:"column:CryptoType;default:1;comment:加密类型"` //加密类型 1: 明文 2AES 3 Rsa交换密匙 aes加密、私钥签名、公钥验签。
	CryptoKeyAes       string `json:"CryptoKeyAes"  gorm:"column:CryptoKeyAes;comment:加密通信Aes密匙"`
	CryptoKeyPrivate   string `json:"CryptoKeyPrivate"  gorm:"column:CryptoKeyPrivate;size:1024;comment:加密通信私钥签名"`
	CryptoKeyPublic    string `json:"CryptoKeyPublic"  gorm:"column:CryptoKeyPublic;size:1024;comment:加密通信公钥验签"`
	MaxOnline          int    `json:"MaxOnline"  gorm:"column:MaxOnline;default:999;comment:默认在线最大数量"`                     //在线最大数量 这个是默认 只有用户最大值是0时才用这个
	ExceedMaxOnlineOut int    `json:"ExceedMaxOnlineOut"  gorm:"column:ExceedMaxOnlineOut;default:1;comment:超过在线最大数量处理方式"` //1踢掉最先登录的账号  2 直接提示
	AppType            int    `json:"AppType"  gorm:"column:AppType;default:1;comment:软件类型"`                               //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	RmbToVipNumber     int    `json:"RmbToVipNumber"  gorm:"column:RmbToVipNumber;default:1;comment:1人民币换多少积分"`
	Captcha            string `json:"Captcha"  gorm:"column:Captcha;;comment:需要验证码的接口"`
}

func (DB_AppInfo) TableName() string {
	return "db_App_Info" //(软件信息表)
}
