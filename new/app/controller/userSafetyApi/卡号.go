package userSafetyApi

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/logic/common/blacklist"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/log"
	"server/new/app/models/constant"
	db2 "server/new/app/models/db"
	"server/new/app/service"
	"time"
)

func UserApi_取注册送卡(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if 局_ctx.AppInfo.AppType != 3 && 局_ctx.AppInfo.AppType != 4 {
		response.FailMsg(c, constant.Status_操作失败, "仅限卡号类型应用使用")
		return
	}

	//{"Api":"GetRegisterGiveKa","Key":"677F23CB3FA0055B5FD03916D6AB3C9A"}

	var 局_卡 db2.DB_Ka

	var err error
	if len(局_ctx.Q请求明文.Get("Key").String()) > 191 {
		response.FailMsg(c, constant.Status_操作失败, "绑定信息长度不能超过191")
		return
	}

	if blacklist.Is黑名单(局_ctx.Q请求明文.Get("Key").String(), 局_ctx.AppInfo.AppId) {
		response.FailMsg(c, constant.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}

	_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoKey(局_ctx.Q请求明文.Get("Key").String())
	if err == nil {
		response.FailMsg(c, constant.Status_操作失败, "已存在绑定信息,无法获取卡号")
		return
	}
	_, err = service.NewKa(c, &db).Info(局_ctx.AppInfo.RegisterGiveKaClassId)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "应用未设置赠送卡类,或卡类已删除")
		return
	}

	局_卡, err2 := ka.L_ka.Ka单卡创建(c, 局_ctx.AppInfo.RegisterGiveKaClassId, -1, "系统自动", "key测试卡:"+局_ctx.Q请求明文.Get("Key").String(), "", 0)
	if err2 != nil {
		response.FailMsg(c, constant.Status_操作失败, "卡号创建失败")
		return
	}

	var 局_AppUser db2.DB_AppUser
	局_AppUser.Id = 0
	局_AppUser.Uid = 局_卡.Id
	局_AppUser.Status = 1
	局_AppUser.Key = 局_ctx.Q请求明文.Get("Key").String()
	局_AppUser.VipNumber = 局_卡.VipNumber
	局_AppUser.Note = 局_卡.AdminNote
	局_AppUser.MaxOnline = S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline)
	局_AppUser.UserClassId = 局_卡.UserClassId
	局_AppUser.RegisterTime = time.Now().Unix()
	局_AppUser.AgentUid = 0 //不在这里赋值,单独处理

	//没有这个用户,应该是第一次登录应用,添加进去
	switch 局_ctx.AppInfo.AppType {
	case 3:
		局_AppUser.VipTime = time.Now().Unix() + 局_卡.VipTime
	case 4:
		局_AppUser.VipTime = 局_卡.VipTime
	default:
		//???应该不会到这里
		response.FailMsg(c, constant.Status_SQl错误, "AppInfo.AppType错误")
	}
	_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).Create(&局_AppUser)
	if err != nil {
		var 日志 = db2.DB_LogUserMsg{
			User:    "系统",
			App:     局_ctx.AppInfo.AppName,
			AppId:   局_ctx.AppInfo.AppId,
			AppVer:  局_ctx.Z在线信息.AppVer,
			MsgType: log.Log用户消息类型_系统执行错误,
			Note:    "新添加软件用户时失败报错信息:" + err.Error(),
			Ip:      c.ClientIP(),
		}
		err = log.L_log.S输出日志(c, 日志)
		if err != nil {
			return
		}
		response.FailMsg(c, constant.Status_SQl错误, "New用户信息内部错误")
		return
	}
	service.NewKa(c, &db).Update(局_卡.Id, map[string]interface{}{"UsedCount": 1}) //更新使用次数

	ka.L_ka.Z置归属代理(c, 局_ctx.AppInfo.AppId, 局_卡.Id, 局_ctx.Z在线信息.AgentUid) //失败也不影响
	//这里吧成功的状态
	response.OkData(c, gin.H{
		"Name":      局_卡.Name,
		"VipNumber": 局_卡.VipNumber,
		"VipTime":   局_卡.VipTime,
	})

}
