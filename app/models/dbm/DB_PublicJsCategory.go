package dbm

type DB_PublicJsCategory struct {
	Id       int    `json:"Id" gorm:"column:Id;primarykey;comment:id;AUTO_INCREMENT"`
	ParentId int    `json:"ParentId" gorm:"column:ParentId;default:0;comment:父分类Id,0为顶级分类;index"`
	Name     string `json:"Name" gorm:"column:Name;size:100;uniqueIndex;comment:分类名"`
	Sort     int64  `json:"Sort" gorm:"column:Sort;default:0;comment:排序权重;"`
	Note     string `json:"Note" gorm:"column:Note;size:1000;comment:备注;"`
}

func (DB_PublicJsCategory) TableName() string {
	return "db_public_js_category" //(公共函数分类表)
}
