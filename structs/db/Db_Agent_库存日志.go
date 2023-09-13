package DB

type Db_Agent_库存日志 struct {
	ID          int    `json:"Id" gorm:"column:Id;primaryKey;comment:Id"`
	User1       string `json:"User1" gorm:"column:User1;size:200;comment:来源用户ID"`
	User1Role   int    `json:"User1Role" gorm:"column:User1Role;comment:来源用户ID角色"`
	User2       string `json:"User2" gorm:"column:User2;size:200;comment:去向用户ID"`
	User2Role   int    `json:"User2Role" gorm:"column:User2Role;comment:去向用户ID角色"`
	Num         int    `json:"Num" gorm:"column:Num;comment:转移数量"`
	Type        int    `json:"Type" gorm:"column:Type;comment:类型,1:1向2发送,2:1接收2"`
	InventoryId int    `json:"InventoryId" gorm:"column:InventoryId;comment:操作资源包ID"` //如果是发送,就是原始库存包,如果是接收,就是新库存包
	Time        int64  `json:"Time" gorm:"column:Time;comment:操作时间"`
	Note        string `json:"Note" gorm:"column:Note;comment:消息"`
	Ip          string `json:"Ip" gorm:"column:Ip;size:20;comment:操作Ip"`
}

func (Db_Agent_库存日志) TableName() string {
	return "db_Log_AgentInventory" //代理库存日志表
}
