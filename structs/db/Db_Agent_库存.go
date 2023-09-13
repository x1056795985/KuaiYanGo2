package DB

type Db_Agent_库存卡包 struct {
	Id             int    `json:"Id" gorm:"column:Id;primaryKey;comment:Id"`
	Uid            int    `json:"Uid" gorm:"column:Uid;comment:所属用户ID"`
	KaClassId      int    `json:"KaClassId" gorm:"column:KaClassId;comment:卡类Id"`
	Num            int    `json:"Num" gorm:"column:Num;comment:库存数量"`
	NumMax         int    `json:"NumMax" gorm:"column:NumMax;comment:最大数量"`
	RegisterUserId int    `json:"RegisterUserId" gorm:"column:RegisterUserId;comment:库存创建用户ID"` //只能创建人重置过期时间
	EndTime        int64  `json:"EndTime" gorm:"column:EndTime;comment:库存到期时间"`                 //到期后只能上级收回或废弃
	Note           string `json:"Note" gorm:"column:Note;comment:备注"`
	SourceID       int    `json:"SourceID" gorm:"column:SourceID;comment:来源用户库存卡包ID"` //0为直接购买,只有持有来源库存卡包ID的代理才可以收回库存,管理员为-1
	SourceUid      int    `json:"SourceUid" gorm:"column:SourceUid;comment:来源用户库存卡包所属Uid"`
	StartTime      int64  `json:"StartTime" gorm:"column:StartTime;comment:来源时间"`
}

func (Db_Agent_库存卡包) TableName() string {
	return "db_Agent_Inventory" //代理库存表
}
