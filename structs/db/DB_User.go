package DB

type DB_User struct {
	Id                  int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`                        // id
	User                string  `json:"User" gorm:"column:User;size:191;UNIQUE;index;comment:用户登录名"`          // 用户登录名
	PassWord            string  `json:"PassWord" gorm:"column:PassWord;size:191;comment:用户登录密码"`              // 用户登录密码
	Phone               string  `json:"Phone" gorm:"column:Phone;size:191;comment:用户手机号"`                     // 用户手机号
	Email               string  `json:"Email" gorm:"column:Email;size:191;comment:用户邮箱"`                      // 用户邮箱
	Qq                  string  `json:"Qq" gorm:"column:Qq;size:191;comment:Qq"`                              // 用户Qq
	SuperPassWord       string  `json:"SuperPassWord" gorm:"column:SuperPassWord;size:191;comment:超级密码或密保答案"` // 超级密码或密保答案
	Status              int     `json:"Status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"`          // 1正常 2冻结
	Rmb                 float64 `json:"Rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:余额"`
	Note                string  `json:"Note" gorm:"column:Note;size:1000;comment:备注"`
	RealNameAttestation string  `json:"RealNameAttestation" gorm:"column:RealNameAttestation;comment:实名认证,认证成功直接填写姓名未认证空"`
	UPAgentId           int     `json:"UPAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount       int     `json:"AgentDiscount" gorm:"column:AgentDiscount;default:0;comment:分成百分比"`
	LoginAppid          int     `json:"LoginAppid" gorm:"column:LoginAppid;comment:登录appid"`
	LoginIp             string  `json:"LoginIp" gorm:"column:LoginIp;size:191;comment:登录ip"`
	LoginTime           int64   `json:"LoginTime" gorm:"column:LoginTime;comment:登录时间"`
	RegisterIp          string  `json:"RegisterIp" gorm:"column:RegisterIp;size:191;comment:注册ip"`
	RegisterTime        int64   `json:"RegisterTime" gorm:"column:RegisterTime;comment:注册时间"`
	Sort                int64   `json:"Sort" gorm:"column:Sort;default:0;comment:排序权重; "`
}

func (DB_User) TableName() string {
	return "db_User" //(用户表)
}
