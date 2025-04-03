package db

// 代理卡类调价表
type DB_KaClassUpPrice struct {
	Id        int     `json:"Id" gorm:"column:Id;primarykey;comment:id"`
	KaClassId int     `json:"KaClassId" gorm:"column:KaClassId;comment:卡类id;uniqueIndex:idx_ka_agent"`
	AgentId   int     `json:"AgentId" gorm:"column:AgentId;comment:代理id;uniqueIndex:idx_ka_agent"`
	Markup    float64 `json:"Markup" gorm:"column:Markup;type:decimal(10,2);default:0;comment:调价幅度"`
}

func (DB_KaClassUpPrice) TableName() string {
	return "db_ka_class_up_price"
}
