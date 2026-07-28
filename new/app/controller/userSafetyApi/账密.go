package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/Service/Captcha"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/logic/common/blacklist"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/log"
	"server/new/app/logic/common/rmbPay"
	logicUser "server/new/app/logic/userSafetyApi/user"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"server/new/app/utils"
	"strings"
	"time"
)

func UserApi_取用户基础信息(c *gin.Context) {
	//{"Api":"GetUserInfo"}
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	if 局_ctx.AppInfo.AppType != 1 && 局_ctx.AppInfo.AppType != 2 { //账号计时账号计点
		response.FailMsg(c, constant.Status_操作失败, "仅限账号登录用户可获取")
		return
	}

	db := *global.GVA_DB
	局_User, err := service.NewUser(c, &db).Info(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}

	response.OkData(c, gin.H{
		"Id":                  局_User.Id,
		"User":                局_User.User,
		"RegisterIp":          局_User.RegisterIp,
		"RegisterTime":        局_User.RegisterTime,
		"LoginAppid":          局_User.LoginAppid,
		"LoginIp":             局_User.LoginIp,
		"LoginTime":           局_User.LoginTime,
		"RealNameAttestation": 局_User.RealNameAttestation != "",
		"Qq":                  局_User.Qq,
		"Phone":               局_User.Phone,
		"Email":               局_User.Email,
		"RMB":                 局_User.Rmb,
	})
	return
}
func UserApi_置用户基础信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	局data := map[string]interface{}{}
	局_Qq := 局_ctx.Q请求明文.Get("Qq").String()
	局_Email := 局_ctx.Q请求明文.Get("Email").String()
	局_Phone := 局_ctx.Q请求明文.Get("Phone").String()
	if 局_Qq != "" {
		局data["Qq"] = 局_Qq
	}
	if 局_Email != "" {
		局data["Email"] = 局_Email
	}
	if 局_Phone != "" {
		局data["Phone"] = 局_Phone
	}

	db := *global.GVA_DB
	_, err := service.NewUser(c, &db).Update(局_ctx.Z在线信息.Uid, 局data)

	if err == nil {
		response.Ok(c)
	} else {
		response.Fail(c, constant.Status_操作失败)
	}

	return
}

func UserApi_密码找回或修改_验证旧密码(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	局_用户名 := 局_ctx.Q请求明文.Get("User").String()
	局_用户Id := 0
	if 局_用户名 != "" {
		db := *global.GVA_DB
		局_userInfo, err := service.NewUser(c, &db).InfoName(局_用户名)
		if err == nil {
			局_用户Id = 局_userInfo.Id
		}
	}
	if 局_用户Id == 0 {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}

	msg := ""
	局_新密码 := 局_ctx.Q请求明文.Get("NewPassWord").String()
	if !utils.Z正则_校验密码(局_新密码, &msg) {
		response.FailMsg(c, constant.Status_操作失败, "密码"+msg)
		return
	}

	db := *global.GVA_DB
	局_user, _ := service.NewUser(c, &db).Info(局_用户Id)
	if !utils.BcryptCheck(局_ctx.Q请求明文.Get("OldPassWord").String(), 局_user.PassWord) {
		response.FailMsg(c, constant.Status_操作失败, "旧密码错误.")
		return
	}

	_, err := service.NewUser(c, &db).Update(局_用户Id, map[string]interface{}{"PassWord": utils.Md5String(局_新密码)})
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "修改失败")
	} else {
		_ = service.NewLinksToken(c, &db).Set批量注销Uid数组([]int{局_用户Id}, 0, constant.Z注销_用户改密注销)
		response.OkMsg(c, "修改成功")
	}
	return

}
func UserApi_密码找回或修改_超级密码(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	局_用户名 := 局_ctx.Q请求明文.Get("User").String()
	局_用户Id := 0
	if 局_用户名 != "" {
		db := *global.GVA_DB
		局_userInfo, err := service.NewUser(c, &db).InfoName(局_用户名)
		if err == nil {
			局_用户Id = 局_userInfo.Id
		}
	}
	if 局_用户Id == 0 {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}

	msg := ""
	局_新密码 := 局_ctx.Q请求明文.Get("NewPassWord").String()
	if !utils.Z正则_校验密码(局_新密码, &msg) {
		response.FailMsg(c, constant.Status_操作失败, "密码"+msg)
		return
	}

	db := *global.GVA_DB

	switch 局_ctx.Q请求明文.Get("Type").Int() {
	case 1:
		局_user, _ := service.NewUser(c, &db).Info(局_用户Id)
		if !utils.BcryptCheck(局_ctx.Q请求明文.Get("SuperPassWord").String(), 局_user.SuperPassWord) {
			response.FailMsg(c, constant.Status_操作失败, "超级密码错误.")
			return
		}
	case 2:
		UserApi_密码找回或修改_密保手机(c) //兼容旧版本 1.0.148 版本之后,接口转成两种接口名称
		return
	default:
		response.FailMsg(c, constant.Status_操作失败, "密码找回或修改方式参数错误")
		return
	}

	_, err := service.NewUser(c, &db).Update(局_用户Id, map[string]interface{}{"PassWord": utils.Md5String(局_新密码)})
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "修改失败")
	} else {
		_ = service.NewLinksToken(c, &db).Set批量注销Uid数组([]int{局_用户Id}, 0, constant.Z注销_用户改密注销)
		response.OkMsg(c, "修改成功")
	}
	return

}

func UserApi_密码找回或修改_密保手机(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	局_用户名 := 局_ctx.Q请求明文.Get("User").String()
	局_用户Id := 0
	if 局_用户名 != "" {
		db := *global.GVA_DB
		局_userInfo, err := service.NewUser(c, &db).InfoName(局_用户名)
		if err == nil {
			局_用户Id = 局_userInfo.Id
		}
	}
	if 局_用户Id == 0 {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}

	msg := ""
	局_新密码 := 局_ctx.Q请求明文.Get("NewPassWord").String()
	if !utils.Z正则_校验密码(局_新密码, &msg) {
		response.FailMsg(c, constant.Status_操作失败, "密码"+msg)
		return
	}

	局_短信验证码Id := 局_ctx.Q请求明文.Get("PhoneCaptchaId").String()
	局_短信验证码值 := 局_ctx.Q请求明文.Get("PhoneCaptchaValue").String()
	if 局_短信验证码Id == "" || 局_短信验证码值 == "" {
		response.FailMsg(c, constant.Status_操作失败, "验证码错误.")
		return
	}

	if strings.Index(局_短信验证码Id, "Note") != 0 {
		go log.L_log.S写风控日志(c, 局_ctx.Z在线信息.Id, constant.Log风控类型_Api异常调用, 局_用户名, c.ClientIP(), "使用绑定手机密码找回或修改,用户使用非短信验证码Id进行提交,可能是异常用户")
		response.FailMsg(c, constant.Status_操作失败, "验证码错误.")
		return
	}

	db := *global.GVA_DB
	局_User, err := service.NewUser(c, &db).InfoName(局_用户名)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}

	if !Captcha.H缓存验证码校验实例.Verify(局_短信验证码Id, 局_短信验证码值, false) {
		response.FailMsg(c, constant.Status_操作失败, "短信验证码错误.")
		return
	}
	if 局_User.Phone == "" || strings.Index(局_短信验证码Id, "Note"+utils.Md5String(局_User.Phone)[:16]) == -1 {
		go log.L_log.S写风控日志(c, 局_ctx.Z在线信息.Id, constant.Log风控类型_Api异常调用, 局_用户名, c.ClientIP(), "使用绑定手机密码找回或修改,用户使用非账号绑定的验证码进行提交,可能是异常用户")
		response.FailMsg(c, constant.Status_操作失败, "验证码错误.")
		return
	}

	_, err = service.NewUser(c, &db).Update(局_用户Id, map[string]interface{}{"PassWord": utils.Md5String(局_新密码)})
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "修改失败")
	} else {
		_ = service.NewLinksToken(c, &db).Set批量注销Uid数组([]int{局_用户Id}, 0, constant.Z注销_用户改密注销)
		response.OkMsg(c, "修改成功")
	}
	return
}

func UserApi_取用户余额(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	db := *global.GVA_DB
	局_User, err := service.NewUser(c, &db).Info(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户信息失败.")
		return
	}
	response.OkData(c, gin.H{"Rmb": 局_User.Rmb})
	return
}

func UserApi_订单_余额充值(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	if 局_ctx.AppInfo.AppId < 10000 {
		response.FailMsg(c, constant.Status_操作失败, "应用不存在")
		return
	}

	db := *global.GVA_DB
	局_用户名 := 局_ctx.Q请求明文.Get("User").String()
	局_卡号 := 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4

	var 局_Uid = 0

	var 局_Uid类型 = 0
	if 局_卡号 {
		局_Uid类型 = 2
		局_卡信息, err := service.NewKa(c, &db).Info2(map[string]interface{}{"Name": 局_用户名, "AppId": 局_ctx.AppInfo.AppId})
		if err == nil {
			局_Uid = 局_卡信息.Id
		}
	} else {
		局_Uid类型 = 1
		局_userInfo, err := service.NewUser(c, &db).InfoName(局_用户名)
		if err == nil {
			局_Uid = 局_userInfo.Id
		}
	}
	局_appUser, _ := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_Uid)
	if 局_appUser.Uid == 0 {
		response.FailMsg(c, constant.Status_操作失败, "要充值的用户不存在")
		return
	}

	var err error
	局_支付方式 := strings.TrimSpace(局_ctx.Q请求明文.Get("PayType").String())
	//==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 局_appUser.Uid
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.ReceivedUid = 局_appUser.AgentUid
	参数.Rmb = 局_ctx.Q请求明文.Get("Money").Float64()
	参数.ProcessingType = constant.D订单类型_余额充值
	参数.E额外信息 = gjson.New("{}")
	err = 参数.E额外信息.Set("AppId", 局_ctx.AppInfo.AppId)
	err = 参数.E额外信息.Set("在线信息AgentUid", 局_ctx.Z在线信息.AgentUid)

	响应数据2, err := rmbPay.L_rmbPay.D订单创建(c, 参数)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "充值方式["+局_ctx.Q请求明文.Get("PayType").String()+"]"+err.Error())
	} else {
		response.OkData(c, 响应数据2)
	}
	return
}
func UserApi_取用户是否存在(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	db := *global.GVA_DB
	局_用户名 := 局_ctx.Q请求明文.Get("User").String()
	局_Uid := 0
	if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 {
		//卡号
		局_卡信息, err := service.NewKa(c, &db).Info2(map[string]interface{}{"Name": 局_用户名, "AppId": 局_ctx.AppInfo.AppId})
		if err == nil {
			局_Uid = 局_卡信息.Id
		}
	} else {
		局_userInfo, err := service.NewUser(c, &db).InfoName(局_用户名)
		if err == nil {
			局_Uid = 局_userInfo.Id
		}
	}
	if 局_Uid == 0 {
		response.OkData(c, gin.H{"IsUser": false})
	} else {
		response.OkData(c, gin.H{"IsUser": true})
	}

	return
}
func UserApi_用户注册(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测_账密模式专用(c) {
		return
	}

	//检查是否可以绑定相同信息
	局_Key := 局_ctx.Q请求明文.Get("Key").String()
	if 局_ctx.AppInfo.IsUserKeySame == 2 && 局_Key == "" {
		response.FailMsg(c, constant.Status_操作失败, "绑定信息不能为空.")
		return
	}
	if len(局_Key) > 191 {
		response.FailMsg(c, constant.Status_操作失败, "绑定信息长度不能超过191")
		return
	}

	db := *global.GVA_DB
	if 局_ctx.AppInfo.IsUserKeySame == 2 && 局_Key != "" {
		_, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoKey(局_Key)
		if err == nil {
			response.FailMsg(c, constant.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
			return
		}
	}

	if blacklist.Is黑名单(局_Key, 局_ctx.AppInfo.AppId) {
		response.FailMsg(c, constant.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}

	_, err := logicUser.L_user.New用户信息(c, 局_ctx.Q请求明文.Get("User").String(), 局_ctx.Q请求明文.Get("PassWord").String(), 局_ctx.Q请求明文.Get("SuperPassWord").String(), 局_ctx.Q请求明文.Get("Qq").String(), 局_ctx.Q请求明文.Get("Email").String(), 局_ctx.Q请求明文.Get("Phone").String(), c.ClientIP(), "", 0, 0, 0, "")
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	var 局_VipNumber int64
	if 局_ctx.AppInfo.AppType != 2 && 局_ctx.AppInfo.AppType != 4 { //非计点模式
		局_VipNumber = time.Now().Unix()
	} else {
		局_VipNumber = 0
	}

	//没有这个用户,应该是第一次登录应用,添加进去
	局_userInfo, _ := service.NewUser(c, &db).InfoName(局_ctx.Q请求明文.Get("User").String())
	局_Uid := 局_userInfo.Id

	var 局_AppUser dbm.DB_AppUser
	局_AppUser.Id = 0
	局_AppUser.Uid = 局_Uid
	局_AppUser.Status = 1
	局_AppUser.Key = 局_Key
	局_AppUser.VipTime = 局_VipNumber
	局_AppUser.Note = ""
	局_AppUser.MaxOnline = 1
	局_AppUser.RegisterTime = time.Now().Unix()
	局_AppUser.AgentUid = 0

	_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).Create(&局_AppUser)
	if err != nil {
		response.FailMsg(c, constant.Status_SQl错误, "New用户信息内部错误,用户注册成功,注册软件用户失败"+err.Error())
		return
	}
	ka.L_ka.Z置归属代理(c, 局_ctx.AppInfo.AppId, 局_Uid, 局_ctx.Z在线信息.AgentUid) //失败也不影响
	// 注册送卡
	if 局_ctx.AppInfo.RegisterGiveKaClassId > 0 {
		_ = ka.L_ka.K卡类直冲_事务(c, 局_ctx.AppInfo.RegisterGiveKaClassId, 局_Uid)
	}

	response.OkMsg(c, "注册成功")
	return
}
