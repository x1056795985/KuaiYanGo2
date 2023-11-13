package Ser_UserClass

import (
	"server/global"
	DB "server/structs/db"
	"strconv"
)

func UserClassId是否存在(id int) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.DB_UserClass{}).Select("1").Where("Id=?", id).Take(&Count)
	return result.Error == nil

}

func UserClassMark是否存在(AppId int, Mark int) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.DB_UserClass{}).Select("1").Where("Mark=?", Mark).Where("AppId=?", AppId).Take(&Count)
	return result.Error == nil

}
func UserClassName是否存在(AppId int, Name string) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.DB_UserClass{}).Select("1").Where("Name=?", Name).Where("AppId=?", AppId).Take(&Count)
	return result.Error == nil

}
func UserClass取map列表String(Appid int) map[string]string {

	var DB_UserClass []DB.DB_UserClass
	var 总数 int64
	_ = global.GVA_DB.Model(DB.DB_UserClass{}).Select("Id", "Name").Where("Appid=?", Appid).Count(&总数).Find(&DB_UserClass).Error
	var AppName = make(map[string]string, 总数)

	//吧 id 和 app名字 放入map
	for 索引 := range DB_UserClass {
		AppName[strconv.Itoa(int(DB_UserClass[索引].Id))] = DB_UserClass[索引].Name
	}
	return AppName
}

func UserClass取map列表Int(Appid int) map[int]string {

	var DB_UserClass []DB.DB_UserClass
	var 总数 int64
	_ = global.GVA_DB.Model(DB.DB_UserClass{}).Select("Id", "Name").Where("Appid=?", Appid).Count(&总数).Find(&DB_UserClass).Error
	var AppName = make(map[int]string, 总数)

	//吧 id 和 app名字 放入map
	for 索引 := range DB_UserClass {
		AppName[int(DB_UserClass[索引].Id)] = DB_UserClass[索引].Name
	}
	return AppName
}

func UserClass取AppId用户类型列表(Appid int) []DB.DB_UserClass {

	var DB_UserClass []DB.DB_UserClass
	_ = global.GVA_DB.Model(DB.DB_UserClass{}).Where("Appid=?", Appid).Find(&DB_UserClass).Error

	return DB_UserClass
}

// 整数代号是否存在
func IsMark存在数量(Appid, NewMark int, 排除ID []int) int64 {
	var Count int64
	global.GVA_DB.Model(DB.DB_UserClass{}).Where("Appid=?", Appid).Where("Id Not IN ?", 排除ID).Where("Mark=?", NewMark).Count(&Count)
	return Count
}

func Mark取详情(Appid, Mark int) (DB.DB_UserClass, bool) {
	var UserClass DB.DB_UserClass
	err := global.GVA_DB.Model(DB.DB_UserClass{}).Where("Appid=?", Appid).Where("Mark=?", Mark).First(&UserClass).Error

	return UserClass, err == nil
}

func Id取详情(Appid, Id int) (DB.DB_UserClass, bool) {

	var UserClass DB.DB_UserClass
	if Id == 0 {
		UserClass.Name = "未分类"
		UserClass.Mark = 0
		UserClass.Weight = 1
		return UserClass, true
	}
	err := global.GVA_DB.Model(DB.DB_UserClass{}).Where("Appid=?", Appid).Where("Id=?", Id).First(&UserClass).Error

	return UserClass, err == nil
}

func Get权重(用户类型Id int) (权重 int64) {
	权重 = 1
	if 用户类型Id == 0 { //未分类直接返回1
		return
	}
	_ = global.GVA_DB.Model(DB.DB_UserClass{}).Select("Weight").Where("Id=?", 用户类型Id).First(&权重).Error
	return 权重
}
