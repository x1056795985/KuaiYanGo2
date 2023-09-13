package Ser_PublicData

import (
	"server/global"
	DB "server/structs/db"
)

func Name是否存在(AppId int, Name string) bool {
	var Count int64
	global.GVA_DB.Model(DB.DB_PublicData{}).Select("1").Where("Name=?", Name).Where("AppId=?", AppId).Take(&Count)
	return Count > 0

}

func P取值(Appid int, Name string) string {
	var value string
	global.GVA_DB.Model(DB.DB_PublicData{}).Select("Value").Where("AppId=?", Appid).Where("Name=?", Name).First(&value)
	return value
}
func P取值2(Appid int, Name string) (DB.DB_PublicData, error) {
	var value DB.DB_PublicData
	err := global.GVA_DB.Model(DB.DB_PublicData{}).Where("AppId=?", Appid).Where("Name=?", Name).First(&value).Error
	return value, err
}
func P置值(Appid int, Name string, Value string) error {
	return global.GVA_DB.Model(DB.DB_PublicData{}).Select("Value").Where("AppId=?", Appid).Where("Name=?", Name).Update("Value", Value).Error
}
func P置值2(PublicData DB.DB_PublicData) error {
	return global.GVA_DB.Model(DB.DB_PublicData{}).Select("Value", "IsVip", "Note").Omit("Type", "AppId", "Name").Where("AppId=?", PublicData.AppId).Where("Name=?", PublicData.Name).Updates(PublicData).Error
}

func C创建(PublicData DB.DB_PublicData) error {
	err := global.GVA_DB.Model(DB.DB_PublicData{}).Create(&PublicData).Error
	return err
}

func P批量取值(Appid int) []DB.DB_PublicData {
	var value []DB.DB_PublicData
	global.GVA_DB.Model(DB.DB_PublicData{}).Where("AppId=?", Appid).Find(&value)
	return value
}

func P批量置值(DB_PublicData []DB.DB_PublicData) error {
	return global.GVA_DB.Model(DB.DB_PublicData{}).Save(DB_PublicData).Error
}

func P批量修改IsVip(AppId int, Name []string, IsVip int) error {
	return global.GVA_DB.Model(DB.DB_PublicData{}).Where("AppId=?", AppId).Where("Name in ?", Name).Update("IsVip", IsVip).Error
}
