package DB

// db_admin  管理员表 结构体
type DB_Admin struct {
	Id            int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`             // 用户ID
	User          string  `json:"User" gorm:"column:User;index;size:191;comment:用户登录名"`      // 用户登录名
	PassWord      string  `json:"-" gorm:"column:PassWord;size:191; comment:用户登录密码"`         // 用户登录密码
	Phone         string  `json:"phone" gorm:"column:phone;size:191;comment:用户手机号"`          // 用户手机号
	Email         string  `json:"Email" gorm:"column:Email;size:191;comment:用户邮箱"`           // 用户邮箱
	Qq            string  `json:"Qq" gorm:"column:Qq;size:191;comment:Qq"`                   // 用户Qq
	SuperPassWord string  `json:"-" gorm:"column:SuperPassWord;size:191; comment:超级密码或密保答案"` // 超级密码或密保答案
	Status        int     `json:"Status" gorm:"column:Status;comment:用户是否被冻结 1正常 2冻结"`       //用户是否被冻结 1正常 2冻结
	Rmb           float64 `json:"Rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:余额"`
	Authority     string  `json:"authority" gorm:"column:authority;comment:用户权限"` // 用户角色
	AgentDiscount int     `json:"AgentDiscount" gorm:"column:AgentDiscount;default:0;comment:分成百分比"`
}

func (DB_Admin) TableName() string {
	return "db_Admin" //(管理员表)
}
