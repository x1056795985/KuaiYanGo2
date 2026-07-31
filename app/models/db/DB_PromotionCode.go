package db

type DB_PromotionCode struct {
	Id            int    `json:"Id" gorm:"column:Id;primarykey;comment:用户id"`
	PromotionCode string `json:"PromotionCode" gorm:"column:PromotionCode;uniqueIndex;comment:推广代码"`
}

func (DB_PromotionCode) TableName() string {
	return "db_promotion_code" //(推广代码)
}
