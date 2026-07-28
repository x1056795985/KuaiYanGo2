package userSafetyApi

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/log"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"time"
)

// UserApi_心跳 心跳
func UserApi_心跳(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.OkData(c, gin.H{"Status": 1})
		return
	}

	if 局_ctx.AppInfo.Status == 2 { //应用免费模式直接返回 会员状态1
		response.OkData(c, gin.H{"Status": 1})
		return
	}

	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	Status := 1                                                   //1 正常  3 vip过期
	if 局_ctx.AppInfo.AppType == 2 || 局_ctx.AppInfo.AppType == 4 { //计点
		Status = S三元(局_AppUser.VipTime > 0, 1, 3) //'计点模式大于0'
	} else {
		Status = S三元(局_AppUser.VipTime > time.Now().Unix(), 1, 3) //账号模式大于当前时间戳
	}

	response.OkData(c, gin.H{"Status": Status})
	return
}

// UserApi_取动态标签 取动态标签
func UserApi_取动态标签(c *gin.Context) {
	局_ctx := 取上下文(c)
	response.OkData(c, gin.H{"Tab": 局_ctx.Z在线信息.Tab})
	return
}

// UserApi_置动态标签 置动态标签
func UserApi_置动态标签(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	_, err := service.NewLinksToken(c, &db).Update(局_ctx.Z在线信息.Id, map[string]interface{}{"Tab": 局_ctx.Q请求明文.Get("Tab").String()})
	if err != nil {
		response.Fail(c, constant.Status_操作失败)
		return
	}
	response.Ok(c)
	return
}

// UserApi_置代理标志 置代理标志
func UserApi_置代理标志(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	局_代理uid := 局_ctx.Q请求明文.Get("AgentUid").Int()
	局_推广码 := 局_ctx.Q请求明文.Get("PromotionCode").String() //如果有推广码 代理id失效
	if 局_推广码 != "" {
		局_临时, err := service.NewPromotionCode(c, &db).Info2(map[string]interface{}{"PromotionCode": 局_推广码})
		if err == nil {
			局_代理uid = 局_临时.Id
		} else {
			response.FailMsg(c, constant.Status_操作失败, "推广码错误")
			return
		}
	}

	if agentLevel.L_agentLevel.Q取Id代理级别(c, 局_代理uid) <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "AgentUid非代理Uid")
		return
	}

	_, err := service.NewLinksToken(c, &db).Update(局_ctx.Z在线信息.Id, map[string]interface{}{"AgentUid": 局_代理uid})
	if err != nil {
		response.Fail(c, constant.Status_操作失败)
		return
	}
	response.Ok(c)
	return
}

// UserApi_置新用户消息 置新用户消息
func UserApi_置新用户消息(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"SetUserMsg","MsgType":2,"Note":"内存写入错误错误信息:11191919;2424233"}
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	局_消息类型 := 局_ctx.Q请求明文.Get("MsgType").Int()
	if 局_消息类型 < 1 || 局_消息类型 == 4 {
		response.FailMsg(c, constant.Status_操作失败, "消息类型不正确")
		return
	}
	局_消息内容 := 局_ctx.Q请求明文.Get("Msg").String()
	if 局_消息内容 == "" {
		response.FailMsg(c, constant.Status_操作失败, "消息内容不能为空")
		return
	}
	go func() {
		var 日志 = dbm.DB_LogUserMsg{
			User:    局_ctx.Z在线信息.User,
			App:     局_ctx.AppInfo.AppName,
			AppId:   局_ctx.AppInfo.AppId,
			AppVer:  局_ctx.Z在线信息.AppVer,
			MsgType: 局_消息类型,
			Note:    局_消息内容,
			Ip:      c.ClientIP(),
		}
		log.L_log.S输出日志(c, 日志)
	}()
	response.Ok(c)
	return
}
