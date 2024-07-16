package DB

// 可制卡类权限 和功能权限合并在一起 ,KId负数就是功能id权限,正数为可制卡类
// 如果增加工功能 全局搜索  Q取全部代理功能ID_MAP  这个函数内也要增加!!!!!!!!!!!!
const D代理功能_卡号冻结 = -1
const D代理功能_卡号解冻 = -2
const D代理功能_更换卡号 = -3
const D代理功能_删除卡号 = -4 //附加信息  {}
const D代理功能_余额充值 = -5
const D代理功能_发展下级代理 = -6
const D代理功能_卡号追回 = -7
const D代理功能_修改用户绑定 = -8
const D代理功能_转账 = -9
const D代理功能_代收款 = -10
const D代理功能_查看归属软件用户 = -11
const D代理功能_冻结软件用户 = -12
const D代理功能_解冻软件用户 = -13
const D代理功能_修改用户密码 = -14

type Db_Agent_卡类授权 struct {
	ID   int    `json:"Id" gorm:"column:Id;primaryKey;comment:Id"`
	Uid  int    `json:"Uid" gorm:"column:Uid;comment:代理用户ID"`
	KId  int    `json:"KId" gorm:"column:KId;comment:授权制卡的卡类Id"`
	Info string `json:"Info" gorm:"column:Info;comment:附加信息功能描述对应参数等等"`
}

func (Db_Agent_卡类授权) TableName() string {
	return "db_Agent_KaClassAuthority" //制卡类授权
}
