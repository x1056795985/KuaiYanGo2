package Ser_PublicJs

import (
	"errors"
	. "github.com/duolabmeng6/efun/efun"
	E "github.com/duolabmeng6/goefun/eTool"
	"server/global"
	DB "server/structs/db"
	"time"
)

const Js类型_公共函数 = 1
const Js类型_任务池Hook函数 = 2
const Js类型_ApiHook函数 = 3

func Id是否存在(Id int) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.DB_PublicJs{}).Select("1").Where("Id=?", Id).First(&Count)
	return result.Error == nil
}
func Name是否存在(AppId int, Name string) bool {

	var Count int64
	_ = global.GVA_DB.Model(DB.DB_PublicJs{}).Select("1").Where("AppId=?", AppId).Where("Name=?", Name).Take(&Count)
	return Count > 0
}
func Q取值(AppId int, Name string) string {
	var value string
	global.GVA_DB.Model(DB.DB_PublicJs{}).Select("Value").Where("AppId=?", AppId).Where("Name=?", Name).First(&value)
	return value
}
func Q取值2(id int) (DB.DB_PublicJs, error) {
	var value DB.DB_PublicJs
	err := global.GVA_DB.Model(DB.DB_PublicJs{}).Where("Id=?", id).First(&value).Error
	return value, err
}
func Name取Id(AppId []int, Name string) int {
	if Name == "" {
		return 0
	}
	var Id int

	global.GVA_DB.Model(DB.DB_PublicJs{}).Select("Id").Where("AppId IN ?", AppId).Where("Name=?", Name).First(&Id)
	return Id
}
func Z置值(id int, Value string) error {
	return global.GVA_DB.Model(DB.DB_PublicJs{}).Select("Value").Where("id=?", id).Update("Value", Value).Error
}
func Z置值2(PublicJs DB.DB_PublicJs) error {
	//注意宝塔写文件 文件会在 /www/server/panel 文件夹
	err := E.E文件_保存(global.GVA_CONFIG.Q取运行目录+"/云函数/"+PublicJs.Name+".js", PublicJs.Value)
	if err != nil {
		return err
	}
	PublicJs.Value = "/云函数/" + PublicJs.Name + ".js"

	m := map[string]interface{}{}
	m["AppId"] = PublicJs.AppId
	m["Name"] = PublicJs.Name
	m["Value"] = PublicJs.Value
	m["IsVip"] = PublicJs.IsVip
	m["Note"] = PublicJs.Note
	err = global.GVA_DB.Model(DB.DB_PublicJs{}).Where("Id=?", PublicJs.Id).Updates(&m).Error
	if err == nil { //删除缓存
		global.H缓存.Delete(global.GVA_CONFIG.Q取运行目录 + PublicJs.Value)
	}
	return err
}
func C创建(PublicJs DB.DB_PublicJs) error {
	//注意宝塔写文件 文件会在 /www/server/panel 文件夹
	err := E.E文件_保存(global.GVA_CONFIG.Q取运行目录+"/云函数/"+PublicJs.Name+".js", PublicJs.Value)
	if err != nil {
		return errors.New("Js写入文件失败:" + err.Error())
	}
	PublicJs.Value = "/云函数/" + PublicJs.Name + ".js"
	err = global.GVA_DB.Model(DB.DB_PublicJs{}).Create(&PublicJs).Error
	return err
}

func P批量修改IsVip(Id []int, IsVip int) error {
	return global.GVA_DB.Model(DB.DB_PublicJs{}).Where("Id in ?", Id).Update("IsVip", IsVip).Error
}

func P取值2(Appid int, Name string) (DB.DB_PublicJs, error) {
	var 局_PublicJs DB.DB_PublicJs
	err := global.GVA_DB.Model(DB.DB_PublicJs{}).Where("AppId=?", Appid).Where("Name=?", Name).First(&局_PublicJs).Error
	if err != nil {
		return 局_PublicJs, errors.New("[" + Name + "],Hook函数不存在")
	}

	局_临时, ok := global.H缓存.Get(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value)
	if ok {
		局_PublicJs.Value = 局_临时.(string)
	} else {
		if E文件是否存在(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value) {
			局_PublicJs.Value = string(E读入文件(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value))
			global.H缓存.Set(global.GVA_CONFIG.Q取运行目录+局_PublicJs.Value, 局_PublicJs.Value, time.Hour*720)
		} else {
			return 局_PublicJs, errors.New(Name + ".js文件读取失败可能被删除,请重新编辑公共函数")
		}
	}

	return 局_PublicJs, err
}
