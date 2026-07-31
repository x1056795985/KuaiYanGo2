package userConfig

import (
	"github.com/gin-gonic/gin"
	"server/app/global"
	db2 "server/app/models/db"
	"server/app/service"
	"time"
)

var L_userConfig userConfig

func init() {
	L_userConfig = userConfig{}
}

type userConfig struct {
}

// Q取值 取用户云配置值
func (j *userConfig) Q取值(c *gin.Context, AppId, Uid int, Name string) string {
	db := *global.GVA_DB
	局_配置, err := service.NewUserConfig(c, &db).Info2(map[string]interface{}{
		"AppId": AppId,
		"Uid":   Uid,
		"Name":  Name,
	})
	if err != nil {
		return ""
	}
	return 局_配置.Value
}

// Z置值 置用户云配置值(多表操作: 根据App类型从Ka或User表读取用户名)
// AppInfo: 应用信息(用于判断卡号模式), AppId: 配置归属AppId, Uid, Name, Value: 配置项
func (j *userConfig) Z置值(c *gin.Context, AppInfo db2.DB_AppInfo, AppId, Uid int, Name, Value string) error {
	db := *global.GVA_DB
	局_服务 := service.NewUserConfig(c, &db)
	//先查是否存在
	_, err := 局_服务.Info2(map[string]interface{}{
		"AppId": AppId,
		"Uid":   Uid,
		"Name":  Name,
	})
	if err == nil {
		//存在则更新
		_, err = 局_服务.Update(map[string]interface{}{
			"AppId": AppId,
			"Uid":   Uid,
			"Name":  Name,
		}, map[string]interface{}{
			"Value":      Value,
			"UpdateTime": time.Now().Unix(),
		})
		return err
	}
	//不存在则创建,需要读取用户名
	var 局_用户名 string
	if AppInfo.AppType > 2 {
		//卡号模式
		局_卡, e := service.NewKa(c, &db).Info(Uid)
		if e == nil {
			局_用户名 = 局_卡.Name
		}
	} else {
		//账密模式
		局_User, e := service.NewUser(c, &db).Info(Uid)
		if e == nil {
			局_用户名 = 局_User.User
		}
	}
	局_用户配置 := db2.DB_UserConfig{
		AppId:      AppId,
		Uid:        Uid,
		Name:       Name,
		Value:      Value,
		Time:       time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
		User:       局_用户名,
	}
	_, err = 局_服务.Create(局_用户配置)
	return err
}

// Z置值_空删除 置用户云配置值,值为空则删除配置
func (j *userConfig) Z置值_空删除(c *gin.Context, AppInfo db2.DB_AppInfo, AppId, Uid int, Name, Value string) error {
	db := *global.GVA_DB
	局_服务 := service.NewUserConfig(c, &db)
	if Value == "" { //值为空则删
		_, err := 局_服务.Delete2(map[string]interface{}{
			"AppId": AppId,
			"Uid":   Uid,
			"Name":  Name,
		})
		return err
	}
	return j.Z置值(c, AppInfo, AppId, Uid, Name, Value)
}
