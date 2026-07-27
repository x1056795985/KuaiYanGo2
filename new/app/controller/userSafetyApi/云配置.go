package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/logic/common/userConfig"
	"server/new/app/models/constant"
	"server/new/app/service"
)

// UserApi_取代理云配置 取代理云配置
func UserApi_取代理云配置(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	// {"Api":"GetAgentConfig","Name":"配置1"}
	局_配置名 := 局_ctx.Q请求明文.Get("Name").String()
	局_AppUserInfo, _ := service.NewAppUser(c, &db, 局_ctx.Z在线信息.LoginAppid).InfoUid(局_ctx.Z在线信息.Uid)

	局_配置值 := userConfig.L_userConfig.Q取值(c, 50, 局_AppUserInfo.AgentUid, 局_配置名)
	response.OkData(c, gin.H{局_配置名: 局_配置值})
	return
}

// UserApi_取用户云配置 取用户云配置
func UserApi_取用户云配置(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	// {"Api":"GetUserConfig","Name":"配置1"}
	局_配置名 := 局_ctx.Q请求明文.Get("Name").String()
	局_配置值 := userConfig.L_userConfig.Q取值(c, 局_ctx.Z在线信息.LoginAppid, 局_ctx.Z在线信息.Uid, 局_配置名)
	response.OkData(c, gin.H{局_配置名: 局_配置值})
	return
}

// UserApi_置用户云配置 置用户云配置
func UserApi_置用户云配置(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	// {"Api":"GetUserConfig","Name":"配置1","Value":"值"}

	局_配置名 := 局_ctx.Q请求明文.Get("Name").String()
	if 局_配置名 == "" {
		response.FailMsg(c, constant.Status_操作失败, "云配置名不能为空")
		return
	}
	局_配置值 := 局_ctx.Q请求明文.Get("Value").String()
	_ = userConfig.L_userConfig.Z置值_空删除(c, 局_ctx.AppInfo, 局_ctx.Z在线信息.LoginAppid, 局_ctx.Z在线信息.Uid, 局_配置名, 局_配置值)
	response.Ok(c)
	return
}
