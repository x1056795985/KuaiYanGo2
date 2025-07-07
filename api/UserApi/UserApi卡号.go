package UserApi

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/api/UserApi/response"
	"server/new/app/logic/common/blacklist"
	DB "server/structs/db"
	"time"
)

func UserApi_取注册送卡(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if AppInfo.AppType != 3 && AppInfo.AppType != 4 {
		response.X响应状态消息(c, response.Status_操作失败, "仅限卡号类型应用使用")
		return
	}
	if !Ser_KaClass.KaClassId是否存在(AppInfo.RegisterGiveKaClassId) {
		response.X响应状态消息(c, response.Status_操作失败, "应用未设置赠送卡类,或卡类已删除")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetRegisterGiveKa","Key":"677F23CB3FA0055B5FD03916D6AB3C9A"}
	var 局_Uid = 0
	var 局_卡 DB.DB_Ka

	var err error
	if blacklist.Is黑名单(string(请求json.GetStringBytes("Key")), AppInfo.AppId) {
		response.X响应状态消息(c, response.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}

	if Ser_AppUser.B绑定信息是否存在(AppInfo.AppId, string(请求json.GetStringBytes("Key"))) {
		response.X响应状态消息(c, response.Status_操作失败, "已存在绑定信息,无法获取卡号")
		return
	}

	局_卡, err = Ser_Ka.Ka单卡创建(AppInfo.RegisterGiveKaClassId, Ser_Admin.Id取User(1), "key测试卡:"+string(请求json.GetStringBytes("Key")), "", 0)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "卡号创建失败")
		return
	}

	局_Uid = 局_卡.Id
	//没有这个用户,应该是第一次登录应用,添加进去
	switch AppInfo.AppType {
	case 3:
		//注册送卡一定是系统制卡,不会有制卡人 只能为在线代理标志uid
		err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline), time.Now().Unix()+局_卡.VipTime, 局_卡.VipNumber, 局_卡.UserClassId, 局_卡.AdminNote, 局_在线信息.AgentUid)
		_ = Ser_Ka.Ka修改已用次数加一([]int{局_Uid})
	case 4:
		//注册送卡一定是系统制卡,不会有制卡人 只能为在线代理标志uid
		err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline), 局_卡.VipTime, 局_卡.VipNumber, 局_卡.UserClassId, 局_卡.AdminNote, 局_在线信息.AgentUid)
		_ = Ser_Ka.Ka修改已用次数加一([]int{局_Uid})
	default:
		//???应该不会到这里
		response.X响应状态消息(c, response.Status_SQl错误, "AppInfo.AppType错误")
	}

	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 局_卡.Name, AppInfo.AppName, 局_在线信息.AppVer, "新添加软件用户时失败报错信息:"+err.Error(), c.ClientIP())
		response.X响应状态消息(c, response.Status_SQl错误, "New用户信息内部错误")
		return
	}

	//局_AppUser, _ = Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid) //充值之后重新读取一遍
	//这里吧成功的状态
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"Name":      局_卡.Name,
		"VipNumber": 局_卡.VipNumber,
		"VipTime":   局_卡.VipTime,
	})

}
