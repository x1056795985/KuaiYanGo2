package db

type DB_KaClass struct {
	Id           int     `json:"Id" gorm:"column:Id;primarykey"`
	AppId        int     `json:"AppId" gorm:"column:AppId;所属软件id"`
	Name         string  `json:"Name" gorm:"column:Name;size:200;comment:卡名称"` //  年卡月卡季卡等等
	Note         string  `json:"Note" gorm:"column:Note;size:400;comment:卡备注"` //  年卡月卡季卡等等
	Prefix       string  `json:"Prefix" gorm:"column:Prefix;size:200;comment:卡前缀"`
	VipTime      int64   `json:"VipTime" gorm:"column:VipTime;comment:增减时间秒数或点数"`
	InviteCount  int64   `json:"InviteCount" gorm:"column:InviteCount;comment:邀请人增减时间秒数或点数"`
	RMb          float64 `json:"RMb" gorm:"column:RMb;type:decimal(10,2);default:0;comment:余额增减"`
	VipNumber    float64 `json:"VipNumber" gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分增减"`
	Money        float64 `json:"Money" gorm:"column:Money;type:decimal(10,2);default:0;comment:用户价格"`
	AgentMoney   float64 `json:"AgentMoney" gorm:"column:AgentMoney;type:decimal(10,2);default:0;comment:代理价格"`
	UserClassId  int     `json:"UserClassId" gorm:"column:UserClassId;comment:用户分类id"`
	NoUserClass  int     `json:"NoUserClass" gorm:"column:NoUserClass;comment:用户类型不同处理方式,1自动根据权重转换分组,2禁止充值"`
	KaLength     int     `json:"KaLength" gorm:"column:KaLength;comment:卡长度"`
	KaStringType int     `json:"KaStringType" gorm:"column:KaStringType;comment:卡内容字符集"` // 1 大小写字母+数字  2 大写字母 +数字  3 小写字母+数字
	Num          int     `json:"Num" gorm:"column:Num;comment:可以充值次数"`
	KaType       int     `json:"KaType" gorm:"column:KaType;comment:充值类型"`        //1=无限制只要有次数,2每个账号限制一次
	MaxOnline    int     `json:"MaxOnline" gorm:"column:MaxOnline;comment:最大在线数"` //修改可以修改App最大在线数量
}

func (DB_KaClass) TableName() string {
	return "db_Ka_Class" //(卡类属性)
}
