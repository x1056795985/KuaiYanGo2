package DB

type DB_PublicJs struct {
	Id    int    `json:"Id" gorm:"column:Id;primarykey;comment:id;AUTO_INCREMENT"`
	AppId int    `json:"AppId" gorm:"column:AppId;comment:AppId;"`       // 函数归属 1 全局
	Name  string `json:"Name" gorm:"column:Name;size:100;comment:公共函数名"` // 需要支持中文 区分大小写 和 AppId
	Value string `json:"Value" gorm:"column:Value;size:15000;comment:变量值"`
	Type  int    `json:"Type" gorm:"column:Type;comment:数据类型"`                        //数据类型 1=普通函数  2=Hook函数
	IsVip int    `json:"IsVip" gorm:"column:IsVip;comment:Vip点数或用户类型代号限制;default:0 "` //Vip点数大于值才可以获取 或可扩展指定用户类型代号(id不能,必须代号因为代号可以自己编辑,Id不能自己编辑)可以用
	Note  string `json:"Note" gorm:"column:Note;size:1000;comment:备注; "`
}

func (DB_PublicJs) TableName() string {
	return "db_Public_Js" //(公共变量表)
}
