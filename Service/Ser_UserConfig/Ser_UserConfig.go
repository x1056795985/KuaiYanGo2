package Ser_UserConfig

import (
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Ka"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
	"time"
)

func Name是否存在(AppId, Uid int, Name string) bool {
	var Count int64
	global.GVA_DB.Model(DB.DB_UserConfig{}).Select("1").Where("AppId=?", AppId).Where("Uid=?", Uid).Where("Name=?", Name).Take(&Count)
	return Count > 0

}

func Q取值(AppId, Uid int, Name string) string {
	var value string = ""
	global.GVA_DB.Model(DB.DB_UserConfig{}).Select("Value").Where("AppId=?", AppId).Where("Uid=?", Uid).Where("Name=?", Name).First(&value)
	return value
}
func Q取值2(AppId, Uid int, Name string) (DB.DB_UserConfig, error) {
	var value DB.DB_UserConfig
	err := global.GVA_DB.Model(DB.DB_UserConfig{}).Where("AppId=?", AppId).Where("Uid=?", Uid).Where("Name=?", Name).First(&value).Error
	return value, err
}
func Z置值(Appid, Uid int, Name string, Value string) error {
	db := global.GVA_DB.Model(DB.DB_UserConfig{})
	var err error
	if Name是否存在(Appid, Uid, Name) {
		updates := map[string]interface{}{
			"Value":      Value,
			"UpdateTime": time.Now().Unix(),
		}
		err = db.Where("AppId=?", Appid).Where("Uid=?", Uid).Where("Name=?", Name).Updates(updates).Error
	} else {
		var User = ""
		if Ser_AppInfo.App是否为卡号(Appid) {
			User = Ser_Ka.Id取卡号(Uid)
		} else {
			User = Ser_User.Id取User(Uid)
		}
		var 局_用户配置 = DB.DB_UserConfig{AppId: Appid, Uid: Uid, Name: Name, Value: Value, Time: time.Now().Unix(), UpdateTime: time.Now().Unix(), User: User}
		err = db.Create(&局_用户配置).Error
	}

	return err
}
func Z置值2(PublicData DB.DB_UserConfig) error {
	return global.GVA_DB.Model(DB.DB_UserConfig{}).Select("Value", "IsVip", "Note").Omit("Type", "AppId", "Name").Where("AppId=?", PublicData.AppId).Where("Name=?", PublicData.Name).Updates(PublicData).Error
}

func C创建(PublicData DB.DB_UserConfig) error {
	err := global.GVA_DB.Model(DB.DB_UserConfig{}).Create(&PublicData).Error
	return err
}

func P批量取值(Appid int) []DB.DB_UserConfig {
	var value []DB.DB_UserConfig
	global.GVA_DB.Model(DB.DB_UserConfig{}).Where("AppId=?", Appid).Find(&value)
	return value
}

func P批量置值(DB_PublicData []DB.DB_UserConfig) error {

	return global.GVA_DB.Model(DB.DB_UserConfig{}).Save(DB_PublicData).Error
}

func P批量置值2(Appid int, Uid []int, Name string, Value string) error {
	if Value == "" {
		return global.GVA_DB.Model(DB.DB_UserConfig{}).Where("AppId=?", Appid).Where("Uid IN ?", Uid).Where("Name=?", Name).Delete("").Error
	}

	var 局_数据 []DB.DB_UserConfig
	局_数据 = make([]DB.DB_UserConfig, len(Uid))
	for i, v := range Uid {
		局_数据[i].AppId = Appid
		局_数据[i].Uid = v
		局_数据[i].Name = Name
		局_数据[i].Value = Value
	}

	return global.GVA_DB.Model(DB.DB_UserConfig{}).Save(局_数据).Error
}
