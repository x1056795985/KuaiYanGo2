package Ser_Js

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/ka"
	"server/new/app/service"
	DB "server/structs/db"
	"server/utils"
	"time"
)

func jS_批量注册(局_在线信息 DB.DB_LinksToken, user []string, 密码 string) js对象_通用返回 {

	if 局_在线信息.LoginAppid <= 10000 {
		return js对象_通用返回{IsOk: false, Err: "AppId必须大于10000"}
	}
	if Ser_AppInfo.App是否为卡号(局_在线信息.LoginAppid) {
		return js对象_通用返回{IsOk: false, Err: "仅限账号模式应用调用"}
	}

	type 临时 struct {
		User    DB.DB_User
		AppUser DB.DB_AppUser
		Name    string
		IsOk    bool
		Msg     string
	}

	var 局_软件用户信息 = make([]临时, 0, len(user))
	// 创建 Context
	// 创建一个 ResponseWriter 和 Request 对象
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	db := *global.GVA_DB
	var appInfo DB.DB_AppInfo
	appInfo, err2 := service.NewAppInfo(c, &db).Info(局_在线信息.LoginAppid)
	if err2 != nil {
		return js对象_通用返回{IsOk: false, Err: "无该应用信息"}
	}

	for i := range user {
		用户信息, err := Ser_User.New用户信息(user[i], 密码, 密码, "", "", "", "127.0.0.1", "批量注册", 0, 0, 0, "")
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				Name: user[i],
				Msg:  err.Error(),
			})
			continue
		}
		err = Ser_AppUser.New用户信息(appInfo.AppId, 用户信息.Id, "", 1, time.Now().Unix(), 0, 0, "", 0)
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				User: 用户信息,
				Name: user[i],
				Msg:  err.Error(),
			})
			continue
		}

		// 注册送卡
		if appInfo.RegisterGiveKaClassId > 0 {
			_ = ka.L_ka.K卡类直冲_事务(c, appInfo.RegisterGiveKaClassId, 用户信息.Id)
		}
		appUser, err := service.NewAppUser(c, &db, 局_在线信息.LoginAppid).InfoUid(用户信息.Id)
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				User:    用户信息,
				AppUser: DB.DB_AppUser{},
				Name:    user[i],
				IsOk:    false,
				Msg:     "无软件用户信息:" + err.Error(),
			})
			continue
		}
		局_软件用户信息 = append(局_软件用户信息, 临时{
			User:    用户信息,
			AppUser: appUser,
			Name:    user[i],
			IsOk:    true,
			Msg:     "成功",
		})
	}

	return js对象_通用返回{IsOk: true, Err: "成功", Data: 局_软件用户信息}
}

func jS_批量充值(局_在线信息 DB.DB_LinksToken, user, 卡号列表 []string) js对象_通用返回 {

	if 局_在线信息.LoginAppid <= 10000 {
		return js对象_通用返回{IsOk: false, Err: "AppId必须大于10000"}
	}
	if Ser_AppInfo.App是否为卡号(局_在线信息.LoginAppid) {
		return js对象_通用返回{IsOk: false, Err: "仅限账号模式应用调用"}
	}
	if len(S数组_去重复(卡号列表)) != len(卡号列表) {
		return js对象_通用返回{IsOk: false, Err: "卡号有重复不可充值"}
	}

	//先检查所有卡号是否可用
	for i := range 卡号列表 {
		局_卡号详情, err := Ser_Ka.Ka卡号取详情(卡号列表[i])
		if err != nil {
			return js对象_通用返回{IsOk: false, Err: "卡号[" + 卡号列表[i] + "]不存在"}
		}
		if 局_卡号详情.NumMax <= 局_卡号详情.Num {
			return js对象_通用返回{IsOk: false, Err: "卡号[" + 卡号列表[i] + "]已耗尽使用次数"}
		}
	}

	//先检查所有账号是否存在
	for i := range user {
		局_uid := Ser_AppUser.User或卡号取Id(局_在线信息.LoginAppid, user[i])
		if 局_uid == 0 {
			return js对象_通用返回{IsOk: false, Err: "账号[" + user[i] + "不存在"}
		}
	}

	type 临时 struct {
		User    DB.DB_User
		AppUser DB.DB_AppUser
		UseKa   string
		Name    string
		IsOk    bool
		Msg     string
	}

	var 局_软件用户信息 = make([]临时, 0, len(user))
	// 创建 Context
	// 创建一个 ResponseWriter 和 Request 对象
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	db := *global.GVA_DB
	var appInfo DB.DB_AppInfo
	appInfo, err2 := service.NewAppInfo(c, &db).Info(局_在线信息.LoginAppid)
	if err2 != nil {
		return js对象_通用返回{IsOk: false, Err: "无该应用信息"}
	}

	for i := range user {
		if i >= len(卡号列表) {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				Name: user[i],
				Msg:  "无卡号",
			})
			break
		}
		err := ka.L_ka.K卡号充值_事务(c, appInfo.AppId, 卡号列表[i], user[i], "")
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				UseKa: 卡号列表[i],
				Name:  user[i],
				Msg:   err.Error(),
			})
			continue
		}
		用户信息, err := service.NewUser(c, &db).InfoName(user[i])
		appUser, err := service.NewAppUser(c, &db, 局_在线信息.LoginAppid).InfoUid(用户信息.Id)
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				User:    用户信息,
				AppUser: DB.DB_AppUser{},
				UseKa:   卡号列表[i],
				Name:    user[i],
				IsOk:    false,
				Msg:     "无软件用户信息",
			})
			continue
		}
		局_软件用户信息 = append(局_软件用户信息, 临时{
			User:    用户信息,
			AppUser: appUser,
			UseKa:   卡号列表[i],
			Name:    user[i],
			IsOk:    true,
			Msg:     "成功",
		})
	}

	return js对象_通用返回{IsOk: true, Err: "成功", Data: 局_软件用户信息}
}
func jS_批量取账号信息(局_在线信息 DB.DB_LinksToken, user []string, 密码 string) js对象_通用返回 {

	if 局_在线信息.LoginAppid <= 10000 {
		return js对象_通用返回{IsOk: false, Err: "AppId必须大于10000"}
	}
	if Ser_AppInfo.App是否为卡号(局_在线信息.LoginAppid) {
		return js对象_通用返回{IsOk: false, Err: "仅限账号模式应用调用"}
	}

	type 临时 struct {
		User    DB.DB_User
		AppUser DB.DB_AppUser
		Name    string
		IsOk    bool
		Msg     string
	}

	var 局_软件用户信息 = make([]临时, 0, len(user))
	// 创建 Context
	// 创建一个 ResponseWriter 和 Request 对象
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	db := *global.GVA_DB
	for i := range user {
		info, err := service.NewUser(c, &db).InfoName(user[i])
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				AppUser: DB.DB_AppUser{},
				Name:    user[i],
				IsOk:    false,
				Msg:     err.Error(),
			})
			continue
		}
		if !utils.BcryptCheck(密码, info.PassWord) {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				AppUser: DB.DB_AppUser{},
				Name:    user[i],
				IsOk:    false,
				Msg:     "密码错误",
			})
			continue
		}
		appUser, err := service.NewAppUser(c, &db, 局_在线信息.LoginAppid).InfoUid(info.Id)
		if err != nil {
			局_软件用户信息 = append(局_软件用户信息, 临时{
				User:    info,
				AppUser: DB.DB_AppUser{},
				Name:    user[i],
				IsOk:    false,
				Msg:     "无软件用户信息",
			})
			continue
		}
		局_软件用户信息 = append(局_软件用户信息, 临时{
			User:    info,
			AppUser: appUser,
			Name:    user[i],
			IsOk:    true,
			Msg:     "成功",
		})
	}

	return js对象_通用返回{IsOk: true, Err: "成功", Data: 局_软件用户信息}
}
