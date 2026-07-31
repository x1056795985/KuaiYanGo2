package userSafetyApi

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	"time"
)

// UserApi_取软件用户信息 取软件用户信息
func UserApi_取软件用户信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}
	// {"Api":"GetAppUserInfo","AppVer":"1.0.15"}
	// 飞鸟快验内部使用, 主要解决用户更新软件后,继承token,但是在线用户信息的版本号没有改变
	局_应用版本 := 局_ctx.Q请求明文.Get("AppVer").String()
	if 局_应用版本 != "" {
		service.NewLinksToken(c, &db).Update(局_ctx.Z在线信息.Id, map[string]interface{}{"AppVer": 局_应用版本})
	}

	var 局_UserClass dbm.DB_UserClass
	局_UserClass, _ = service.NewUserClass(c, &db).Info(局_AppUser.UserClassId)

	response.OkData(c, gin.H{
		"Id":              局_AppUser.Id,
		"Uid":             局_AppUser.Uid,
		"User":            局_ctx.Z在线信息.User,
		"Key":             局_AppUser.Key,
		"VipTime":         局_AppUser.VipTime,
		"VipNumber":       局_AppUser.VipNumber,
		"Status":          局_AppUser.Status,
		"MaxOnline":       局_AppUser.MaxOnline,
		"AgentUid":        局_AppUser.AgentUid,
		"LoginTime":       局_ctx.Z在线信息.LoginTime,
		"LoginIp":         局_ctx.Z在线信息.Ip,
		"RegisterTime":    局_AppUser.RegisterTime,
		"UserClassId":     局_AppUser.UserClassId,
		"UserClassName":   局_UserClass.Name,
		"UserClassMark":   局_UserClass.Mark,
		"UserClassWeight": 局_UserClass.Weight,
	})

	return
}

// UserApi_取软件用户备注 取软件用户备注
func UserApi_取软件用户备注(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	response.OkData(c, gin.H{"Note": 局_AppUser.Note})
	return
}

// UserApi_取Vip到期时间戳 取Vip到期时间戳
func UserApi_取Vip到期时间戳(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	response.OkData(c, gin.H{"VipTime": 局_AppUser.VipTime})
	return
}

// UserApi_取Vip数据 取Vip数据
func UserApi_取Vip数据(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if 局_ctx.Z在线信息.Uid == 0 || 局_ctx.Z在线信息.Status != 1 {
		response.Fail(c, constant.Status_未登录)
		return
	}

	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.Fail(c, constant.Status_Vip已到期)
		return
	}

	var 局_比较值 int64
	if 局_ctx.AppInfo.AppType == 2 || 局_ctx.AppInfo.AppType == 4 {
		局_比较值 = 0
	} else {
		局_比较值 = time.Now().Unix()
	}

	if 局_AppUser.VipTime > 局_比较值 || 局_ctx.AppInfo.AppType == 2 {
		var VipData interface{}
		err := json.Unmarshal([]byte(局_ctx.AppInfo.VipData), &VipData) //VipData被强制Json了 可以直接反序列化
		if err == nil {
			response.OkData(c, VipData)
		} else {
			response.FailMsg(c, constant.Status_操作失败, "Vip数据非标准Json")
		}

		return
	}
	response.Fail(c, constant.Status_Vip已到期)
	return
}

// UserApi_取用户积分 取用户积分
func UserApi_取用户积分(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB

	if 局_ctx.Z在线信息.Status != 1 || 局_ctx.Z在线信息.Uid == 0 {
		response.Fail(c, constant.Status_未登录)
		return
	}

	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取应用用户信息失败.")
		return
	}
	response.OkData(c, gin.H{"VipNumber": 局_AppUser.VipNumber})
	return
}

// UserApi_取用户绑定信息 取用户绑定信息
func UserApi_取用户绑定信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.")
		return
	}

	response.OkData(c, gin.H{"Key": 局_AppUser.Key})
	return
}

// UserApi_GetUserIP 取用户IP
func UserApi_GetUserIP(c *gin.Context) {
	response.OkData(c, gin.H{"IP": c.ClientIP()})
}

// UserApi_取系统时间戳 取系统时间戳
func UserApi_取系统时间戳(c *gin.Context) {
	response.OkData(c, gin.H{"Time": time.Now().Unix()})
	return
}
