package DB

type DB_Ka struct {
	Id           int     `json:"Id" gorm:"column:Id;primarykey"`
	AppId        int     `json:"AppId" gorm:"column:AppId;comment:所属应用 ;uniqueIndex:AppidName"`
	KaClassId    int     `json:"KaClassId" gorm:"column:KaClassId;comment:卡类id"`
	Name         string  `json:"Name" gorm:"column:Name;comment:卡号;size:191;uniqueIndex:AppidName"`
	Status       int     `json:"Status" gorm:"column:Status;default:1;comment:卡号状态"` // 1正常 2冻结
	RegisterUser string  `json:"RegisterUser" gorm:"column:RegisterUser;size:200;comment:制卡人账号"`
	RegisterTime int     `json:"RegisterTime" gorm:"column:RegisterTime;comment:制卡时间戳"`
	AdminNote    string  `json:"AdminNote" gorm:"column:AdminNote;size:1000;comment:管理员备注"`
	AgentNote    string  `json:"AgentNote" gorm:"column:AgentNote;size:1000;comment:代理备注"`
	VipTime      int64   `json:"VipTime" gorm:"column:VipTime;comment:增减时间秒数或点数"`
	InviteCount  int64   `json:"InviteCount" gorm:"column:InviteCount;comment:邀请人增减时间秒数或点数"`
	RMb          float64 `json:"RMb" gorm:"column:RMb;type:decimal(10,2);default:0;comment:余额增减"`
	VipNumber    float64 `json:"VipNumber" gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分增减"`
	Money        float64 `json:"Money" gorm:"column:Money;type:decimal(10,2);default:0;comment:用户价格"`
	AgentMoney   float64 `json:"AgentMoney" gorm:"column:AgentMoney;type:decimal(10,2);default:0;comment:代理价格"`
	UserClassId  int     `json:"UserClassId" gorm:"column:UserClassId;comment:用户分类id"`
	NoUserClass  int     `json:"NoUserClass" gorm:"column:NoUserClass;comment:用户类型不同处理方式,1自动根据权重转换分组,2禁止充值"`
	KaType       int     `json:"KaType" gorm:"column:KaType;comment:充值类型"`        //1=无限制只要有次数,2每个账号限制一次
	MaxOnline    int     `json:"MaxOnline" gorm:"column:MaxOnline;comment:最大在线数"` //修改可以修改App最大在线数量
	Num          int     `json:"Num" gorm:"column:Num;comment:已经使用次数"`
	NumMax       int     `json:"NumMax" gorm:"column:NumMax;comment:最大可用次数"`
	User         string  `json:"User" gorm:"column:User;size:1000;comment:充值用户User"` //不要id 还需要转换
	UserTime     string  `json:"UserTime" gorm:"column:UserTime;size:1000;comment:充值用户时间戳数组 1,2,3,4,5"`
	InviteUser   string  `json:"InviteUser" gorm:"column:InviteUser;size:1000;comment:邀请人用户名"`
	EndTime      int64   `json:"EndTime" gorm:"column:EndTime;comment:最后可用日期戳"` //9999999999为无限制,如果是时间戳,就有限制了
}

func (DB_Ka) TableName() string {
	return "db_Ka" //已生成卡列表
}
