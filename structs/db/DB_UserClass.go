package DB

type DB_UserClass struct {
	Id     int    `json:"Id" gorm:"column:Id;primarykey"`
	AppId  int    `json:"AppId" gorm:"column:AppId;index;comment:属于哪个软件的用户分类"`
	Name   string `json:"Name" gorm:"column:Name;size:191;comment:分类名称"` //  vip1 初级中级高级
	Mark   int    `json:"Mark" gorm:"column:Mark;comment:分类整数代号"`
	Weight int64  `json:"Weight" gorm:"column:Weight;comment:权重"` //切换分组使用,剩余时间*旧分组权重/新分组权重=新剩余时间   未分类权重=1
}

func (DB_UserClass) TableName() string {
	return "db_UserClass" //(用户分类)
}
