package DB

type DB_AppUser struct {
	Id           int     `json:"Id" gorm:"column:Id;primarykey"`
	Uid          int     `json:"Uid" gorm:"column:Uid;index;UNIQUE;comment:用户的id关联到user表"`
	Status       int     `json:"Status" gorm:"column:Status;default:1;comment:本应用用户状态 1正常 2冻结"`
	Key          string  `json:"Key" gorm:"column:Key;comment:绑定信息"` // key是mysql关键字,踩坑点, 下次重构处理吧
	VipTime      int64   `VipTime:"VipTime"  gorm:"column:VipTime;index; comment:到期时间或剩余点数"`
	VipNumber    float64 `VipTime:"VipNumber"  gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分单独备用"`
	Note         string  `json:"Note" gorm:"column:Note;size:5000;comment:备注"`
	MaxOnline    int     `json:"MaxOnline" gorm:"column:MaxOnline;comment:在线最大数量"`
	UserClassId  int     `json:"UserClassId" gorm:"column:UserClassId;comment:用户分类类型"` // 0 未分类   1,2,3
	RegisterTime int     `json:"RegisterTime" gorm:"column:RegisterTime;comment:注册时间"`
	AgentUid     int     `json:"AgentUid" gorm:"column:AgentUid;Index;default:0;comment:归属代理uid"`
}

func (DB_AppUser) TableName() string {
	return "db_AppUser" //(软件用户表,每个软件一个表)

}
