package jsEngine

import (
	"math"
	"strconv"
	"time"

	"server/app/global"
	"server/app/logic/common/appUser"
	"server/app/logic/common/log"
	"server/app/logic/common/publicData"
	"server/app/logic/common/user"
	"server/app/logic/common/userConfig"
	"server/app/logic/webSocket"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
)

func 脚本引擎_用户Id取详情(online dbm.DB_LinksToken) dbm.DB_User {
	局_上下文 := 脚本引擎_后台上下文()
	局_用户服务 := service.NewUser(局_上下文, global.GVA_DB)
	if online.Uid == 0 {
		online.Uid = 局_用户服务.User用户名取id(online.User)
	}
	局_用户信息, _ := 局_用户服务.Id取详情(online.Uid)
	return 局_用户信息
}

func 脚本引擎_卡号Id取详情(online dbm.DB_LinksToken) dbm.DB_Ka {
	局_上下文 := 脚本引擎_后台上下文()
	局_卡号服务 := service.NewKa(局_上下文, global.GVA_DB)
	if online.Uid == 0 {
		online.Uid = 局_卡号服务.Ka卡号取id(online.LoginAppid, online.User)
	}
	局_卡号信息, _ := 局_卡号服务.Id取详情(online.Uid)
	return 局_卡号信息
}

func 脚本引擎_取软件用户详情(online dbm.DB_LinksToken) dbm.DB_AppUser {
	局_上下文 := 脚本引擎_后台上下文()
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, online.LoginAppid)
	if online.Uid == 0 {
		局_Id := 局_软件用户服务.User或卡号取Id(online.LoginAppid, online.User)
		局_软件用户信息, _ := 局_软件用户服务.Id取详情(online.LoginAppid, 局_Id)
		return 局_软件用户信息
	}
	局_软件用户信息, _ := 局_软件用户服务.Uid取详情(online.LoginAppid, online.Uid)
	return 局_软件用户信息
}

func 脚本引擎_在线注销(online dbm.DB_LinksToken) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_在线服务 := service.NewLinksToken(局_上下文, global.GVA_DB)
	局_条件 := make(map[string]any, 1)
	switch {
	case online.Id != 0:
		局_条件["Id"] = online.Id
	case online.Uid != 0:
		局_条件["Uid"] = online.Uid
	case online.User != "":
		局_条件["User"] = online.User
	default:
		return 脚本引擎_失败消息("在线信息缺少Id、Uid或User")
	}
	局_在线数组, 局_错误 := 局_在线服务.Infos(局_条件)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	局_Id数组 := make([]int, 0, len(局_在线数组))
	for _, 局_在线信息 := range 局_在线数组 {
		局_Id数组 = append(局_Id数组, 局_在线信息.Id)
	}
	if 局_错误 = 局_在线服务.Set批量注销(局_Id数组, constant.Z注销_管理员手动注销); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	for _, 局_在线信息 := range 局_在线数组 {
		if 局_在线信息.LoginAppid == constant.APPID_WebSocket {
			webSocket.L_webSocket.RemoveConnection(局_在线信息.Id)
		}
	}
	return 脚本引擎_成功("注销成功", nil)
}

func 脚本引擎_用户Id增减余额(online dbm.DB_LinksToken, amount float64, reason string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	if online.Uid == 0 {
		online.Uid = service.NewUser(局_上下文, global.GVA_DB).User用户名取id(online.User)
	}
	局_新余额, 局_错误 := user.L_user.Id余额增减(局_上下文, online.Uid, math.Abs(amount), amount >= 0)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	go log.L_log.Log_写余额日志(online.User, online.Ip, reason+"|新余额≈"+strconv.FormatFloat(局_新余额, 'f', 2, 64), amount)
	return 脚本引擎_成功("", nil)
}

func 脚本引擎_用户Id增减积分(online dbm.DB_LinksToken, amount float64, reason string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, online.LoginAppid)
	局_Id := 局_软件用户服务.Uid取Id(online.LoginAppid, online.Uid)
	if online.Uid == 0 {
		局_Id = 局_软件用户服务.User或卡号取Id(online.LoginAppid, online.User)
	}
	if 局_错误 := appUser.L_appUser.Id积分增减(局_上下文, online.LoginAppid, 局_Id, math.Abs(amount), amount >= 0); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	go log.L_log.Log_写积分点数时间日志(online.User, online.Ip, reason, amount, online.LoginAppid, 1)
	return 脚本引擎_成功("", nil)
}

func 脚本引擎_用户Id增减时间点数(appID int, online dbm.DB_LinksToken, amount int, reason string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, appID)
	局_Id := 局_软件用户服务.User或卡号取Id(appID, online.User)
	if 局_Id == 0 {
		局_Id = 局_软件用户服务.Uid取Id(appID, online.Uid)
	}
	局_绝对值 := int64(amount)
	if 局_绝对值 < 0 {
		局_绝对值 = -局_绝对值
	}
	if 局_错误 := appUser.L_appUser.Id点数增减(局_上下文, appID, 局_Id, 局_绝对值, amount >= 0); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	局_日志类型 := 3
	if service.NewAppInfo(局_上下文, global.GVA_DB).App是否为计点(appID) {
		局_日志类型 = 2
	}
	go log.L_log.Log_写积分点数时间日志(online.User, online.Ip, reason, float64(amount), appID, 局_日志类型)
	return 脚本引擎_成功("", nil)
}

func 脚本引擎_读公共变量(name string) string {
	return publicData.L_publicData.Q取值(脚本引擎_后台上下文(), 1, name)
}

func 脚本引擎_置公共变量(name, value string) bool {
	局_上下文 := 脚本引擎_后台上下文()
	var 局_错误 error
	if publicData.L_publicData.Name是否存在(局_上下文, 1, name) {
		局_错误 = publicData.L_publicData.Z置值(局_上下文, 1, name, value)
	} else {
		局_错误 = publicData.L_publicData.C创建(局_上下文, dbm.DB_PublicData{
			AppId: 1, Name: name, Value: value, Type: 1, Time: time.Now().Unix(),
		})
	}
	return 局_错误 == nil
}

func 脚本引擎_置动态标记(online dbm.DB_LinksToken, tag string) bool {
	return service.NewLinksToken(脚本引擎_后台上下文(), global.GVA_DB).Set动态标签(online.Id, tag) == nil
}

func 脚本引擎_用户名或卡号取Uid(appID int, username string) int {
	局_上下文 := 脚本引擎_后台上下文()
	if service.NewAppInfo(局_上下文, global.GVA_DB).App是否为卡号(appID) {
		return service.NewKa(局_上下文, global.GVA_DB).Ka卡号取id(appID, username)
	}
	return service.NewUser(局_上下文, global.GVA_DB).User用户名取id(username)
}

func 脚本引擎_置黑名单(appID int, item, note string) 脚本引擎_Api结果 {
	局_错误 := (&service.S_Blacklist{}).Create(global.GVA_DB, dbm.DB_Blacklist{AppId: appID, ItemKey: item, Note: note})
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", nil)
}

func 脚本引擎_置用户云配置(online dbm.DB_LinksToken, name, value string) 脚本引擎_Api结果 {
	if name == "" {
		return 脚本引擎_失败消息("配置名称不能为空")
	}
	if online.LoginAppid <= 0 || online.Uid <= 0 {
		return 脚本引擎_失败消息("登录信息和Uid必须大于0")
	}
	局_上下文 := 脚本引擎_后台上下文()
	局_应用信息, 局_错误 := service.NewAppInfo(局_上下文, global.GVA_DB).Info(online.LoginAppid)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	if 局_错误 = userConfig.L_userConfig.Z置值_空删除(局_上下文, 局_应用信息, online.LoginAppid, online.Uid, name, value); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", nil)
}

func 脚本引擎_取用户云配置(online dbm.DB_LinksToken, name string) 脚本引擎_Api结果 {
	if name == "" {
		return 脚本引擎_失败消息("配置名称不能为空")
	}
	if online.LoginAppid <= 0 || online.Uid <= 0 {
		return 脚本引擎_失败消息("登录信息和Uid必须大于0")
	}
	局_值 := userConfig.L_userConfig.Q取值(脚本引擎_后台上下文(), online.LoginAppid, online.Uid, name)
	return 脚本引擎_成功("成功", 局_值)
}

func 脚本引擎_置软件用户状态(online dbm.DB_LinksToken, status int) 脚本引擎_Api结果 {
	if online.LoginAppid <= 10000 {
		return 脚本引擎_失败消息("AppId必须大于10000")
	}
	if status != 1 && status != 2 {
		return 脚本引擎_失败消息("修改失败:Status状态代码错误")
	}
	局_上下文 := 脚本引擎_后台上下文()
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, online.LoginAppid)
	if online.Uid == 0 {
		online.Uid = 局_软件用户服务.User或卡号取Uid(online.LoginAppid, online.User)
	}
	if online.Uid == 0 {
		return 脚本引擎_失败消息("Uid不能为0")
	}
	局_Id := 局_软件用户服务.Uid取Id(online.LoginAppid, online.Uid)
	if 局_Id == 0 {
		return 脚本引擎_失败消息("软件用户不存在")
	}
	if 局_错误 := appUser.L_appUser.Z置状态_同步卡号修改(局_上下文, online.LoginAppid, []int{局_Id}, status); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	if status == 2 {
		_ = service.NewLinksToken(局_上下文, global.GVA_DB).Set批量注销Uid数组([]int{online.Uid}, online.LoginAppid, constant.Z注销_管理员手动注销)
	}
	return 脚本引擎_成功("成功", nil)
}
