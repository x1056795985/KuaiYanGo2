package UserApi

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
	"server/Service/Captcha"
	"server/Service/Ser_Agent"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/api/UserApi/response"
	"server/global"
	"server/new/app/logic/common/blacklist"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"time"
)

func UserApi_取用户基础信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	if AppInfo.AppType != 1 && AppInfo.AppType != 2 { //账号计时账号计点
		response.X响应状态消息(c, response.Status_操作失败, "仅限账号登录用户可获取")
		return
	}
	var 局_User DB.DB_User
	局_User, ok := Ser_User.Id取详情(局_在线信息.Uid)

	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
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
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	if AppInfo.AppType != 1 && AppInfo.AppType != 2 { //账号计时账号计点
		response.X响应状态消息(c, response.Status_操作失败, "仅限账号登录用户可获取")
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetUserQqEmailPhone","Qq":"1059795985","Email":"测a试a个a中文1056795985@qq.com","Phone":"15666666666","Time":1683988985,"Status":37865}

	err := Ser_User.Id置QQ邮箱手机号(局_在线信息.Uid, string(请求json.GetStringBytes("Qq")), string(请求json.GetStringBytes("Email")), string(请求json.GetStringBytes("Phone")))

	if err == nil {
		response.X响应状态(c, c.GetInt("局_成功Status"))
	} else {
		response.X响应状态(c, response.Status_操作失败)
	}

	return
}

func UserApi_密码找回或修改_验证旧密码(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//'{"Api":"OldPassWordSetPassWord","User":"aaaaaa","OldPassWord":"aaaaaa","NewPassWord":"aaaaaa","Time":1684042764,"Status":27954}'
	局_用户Id := Ser_User.User用户名取id(string(请求json.GetStringBytes("User")))
	if 局_用户Id == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}

	msg := ""
	局_新密码 := string(请求json.GetStringBytes("NewPassWord"))
	if !utils.Z正则_校验密码(局_新密码, &msg) {
		response.X响应状态消息(c, response.Status_操作失败, "密码"+msg)
		return
	}

	局_user, _ := Ser_User.Id取详情(局_用户Id)
	if !utils2.BcryptCheck(string(请求json.GetStringBytes("OldPassWord")), 局_user.PassWord) {
		response.X响应状态消息(c, response.Status_操作失败, "旧密码错误.")
		return
	}

	if err := Ser_User.Id置新密码(局_用户Id, 局_新密码); err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "修改失败")
	} else {
		_ = Ser_LinkUser.Set批量注销Uid(局_用户Id, Ser_LinkUser.Z注销_用户改密注销)
		response.X响应状态消息(c, c.GetInt("局_成功Status"), "修改成功")
	}
	return

}
func UserApi_密码找回或修改_超级密码(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//'{"Api":"SetPassWord","Type":1,"User":"aaaaaa","SuperPassWord":"aaaaaa","NewPassWord":"aaaaaa","Time":1684042764,"Status":27954}'
	局_用户Id := Ser_User.User用户名取id(string(请求json.GetStringBytes("User")))
	if 局_用户Id == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}

	msg := ""
	局_新密码 := string(请求json.GetStringBytes("NewPassWord"))
	if !utils.Z正则_校验密码(局_新密码, &msg) {
		response.X响应状态消息(c, response.Status_操作失败, "密码"+msg)
		return
	}

	switch 请求json.GetInt("Type") {
	case 1:
		局_user, _ := Ser_User.Id取详情(局_用户Id)
		if !utils2.BcryptCheck(string(请求json.GetStringBytes("SuperPassWord")), 局_user.SuperPassWord) {
			response.X响应状态消息(c, response.Status_操作失败, "超级密码错误.")
			return
		}
	case 2:
		UserApi_密码找回或修改_密保手机(c) //兼容旧版本 1.0.148 版本之后,接口转成两种接口名称
		return
	default:
		response.X响应状态消息(c, response.Status_操作失败, "密码找回或修改方式参数错误")
		return
	}

	if err := Ser_User.Id置新密码(局_用户Id, 局_新密码); err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "修改失败")
	} else {
		_ = Ser_LinkUser.Set批量注销Uid(局_用户Id, Ser_LinkUser.Z注销_用户改密注销)
		response.X响应状态消息(c, c.GetInt("局_成功Status"), "修改成功")
	}
	return

}

func UserApi_密码找回或修改_密保手机(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//'{"Api":"SmsCodeSetPassWord","User":"aaaaaa","SuperPassWord":"aaaaaa","NewPassWord":"aaaaaa","Time":1684042764,"Status":27954}'
	局_用户Id := Ser_User.User用户名取id(string(请求json.GetStringBytes("User")))
	if 局_用户Id == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}

	msg := ""
	局_新密码 := string(请求json.GetStringBytes("NewPassWord"))
	if !utils.Z正则_校验密码(局_新密码, &msg) {
		response.X响应状态消息(c, response.Status_操作失败, "密码"+msg)
		return
	}

	局_短信验证码Id := string(请求json.GetStringBytes("PhoneCaptchaId"))
	局_短信验证码值 := string(请求json.GetStringBytes("PhoneCaptchaValue"))
	if 局_短信验证码Id == "" || 局_短信验证码值 == "" {
		response.X响应状态消息(c, response.Status_操作失败, "验证码错误.")
		return
	}

	if strings.Index(局_短信验证码Id, "Note") != 0 {
		go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, string(请求json.GetStringBytes("User")), c.ClientIP(), "使用绑定手机密码找回或修改,用户使用非短信验证码Id进行提交,可能是异常用户")
		response.X响应状态消息(c, response.Status_操作失败, "验证码错误.")
		return
	}

	局_User, ok := Ser_User.User取详情(string(请求json.GetStringBytes("User")))
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}
	if 局_User.Phone == "" || strings.Index(局_短信验证码Id, "Note"+utils2.Md5String(局_User.Phone)[:16]) == -1 {
		go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, string(请求json.GetStringBytes("User")), c.ClientIP(), "使用绑定手机密码找回或修改,用户使用非账号绑定的验证码进行提交,可能是异常用户")
		response.X响应状态消息(c, response.Status_操作失败, "验证码错误.")
		return
	}
	if !Captcha.H缓存验证码校验实例.Verify(局_短信验证码Id, 局_短信验证码值, false) {
		response.X响应状态消息(c, response.Status_操作失败, "短信验证码错误.")
		return
	}

	if err := Ser_User.Id置新密码(局_用户Id, 局_新密码); err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "修改失败")
	} else {
		_ = Ser_LinkUser.Set批量注销Uid(局_用户Id, Ser_LinkUser.Z注销_用户改密注销)
		response.X响应状态消息(c, c.GetInt("局_成功Status"), "修改成功")
	}
	return
}

func UserApi_取用户余额(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	局_User, ok := Ser_User.Id取详情(局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户信息失败.")
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Rmb": 局_User.Rmb})
	return
}

func UserApi_余额购买积分(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了

	局_花费金额 := 请求json.GetFloat64("Money")
	if 局_花费金额 <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "花费金额要求大于0")
		return
	}
	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "应用用户不存在")
		return
	}
	新余额, err := Ser_User.Id余额增减(局_在线信息.Uid, 局_花费金额, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	局_精确花费金额 := decimal.NewFromFloat(局_花费金额)
	局_精确乘数 := decimal.NewFromInt(int64(AppInfo.RmbToVipNumber))
	局_增减积分, _ := 局_精确花费金额.Mul(局_精确乘数).Float64()

	go Ser_Log.Log_写余额日志(局_在线信息.User, c.ClientIP(), fmt.Sprintf("购买积分:%.2f|新余额≈%.2f", 局_增减积分, 新余额), utils.Float64取负值(局_花费金额))
	err = Ser_AppUser.Id积分增减(AppInfo.AppId, 局_AppUser.Id, 局_增减积分, true)
	if err != nil {
		新余额, err = Ser_User.Id余额增减(局_在线信息.Uid, 局_花费金额, true)
		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 局_在线信息.User, AppInfo.AppName, 局_在线信息.AppVer, "用户余额购买积分,减余额成功,增加积分失败,请手动处理,本次错误原因:"+err.Error(), c.ClientIP())
			response.X响应状态消息(c, response.Status_操作失败, "扣费成功,但是积分增加失败,请联系开发者手动处理")
		} else {
			response.X响应状态消息(c, response.Status_操作失败, "购买积分失败,请重试")
			go Ser_Log.Log_写余额日志(局_在线信息.User, c.ClientIP(), fmt.Sprintf("购买积分失败:%.2f|新余额≈%.2f", 局_增减积分, 新余额), 局_花费金额)

		}
		return

	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AddVipNumber": 局_增减积分})
	return
}
func UserApi_余额购买充值卡(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"PayMoneyToKa","Money":1,"Time":1684550291,"Status":37674}
	var 局_卡类 DB.DB_KaClass
	局_卡类.Id = 请求json.GetInt("KaClassId")
	局_卡类, err := Ser_KaClass.KaClass取详细信息(局_卡类.Id)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "要购买的充值卡类型ID不存在")
		return
	}
	if AppInfo.AppId != 局_卡类.AppId || 局_卡类.Money <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "普通用户无法购买本类型充值卡")
		return
	}

	新余额, err := Ser_User.Id余额增减(局_在线信息.Uid, 局_卡类.Money, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "购买失败,"+err.Error())
		return
	}

	局_卡信息, err2 := Ser_Ka.Ka单卡创建(局_卡类.Id, 局_在线信息.User, "用户"+局_在线信息.User+"自助通过Api购卡", "", 0)
	if err2 != nil {
		新余额, err = Ser_User.Id余额增减(局_在线信息.Uid, 局_卡类.Money, true)
		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 局_在线信息.User, AppInfo.AppName, 局_在线信息.AppVer, "用户余额购卡,减余额成功,制卡失败,请手动处理,本次错误原因:"+err.Error(), c.ClientIP())
			response.X响应状态消息(c, response.Status_操作失败, "购卡失败,费用退还失败,请联系开发者手动处理")
		} else {
			response.X响应状态消息(c, response.Status_操作失败, "购卡失败,请重试")
		}
		return
	} else if 局_在线信息.AgentUid > 0 && 局_卡类.AgentMoney > 0 {
		go Ser_Log.Log_写余额日志(局_在线信息.User, c.ClientIP(), "自助购卡->"+AppInfo.AppName+`->卡ID:`+strconv.Itoa(局_卡类.Id)+",卡号:"+局_卡类.Name+":"+局_卡信息.Name+"|新余额≈"+utils.Float64到文本(新余额, 2), utils.Float64取负值(局_卡类.Money))

		//代理分成
		//开始分利润 20240202 mark处理重构以后改事务
		代理分成数据, err3 := Ser_Agent.D代理分成计算(局_在线信息.AgentUid, 局_卡类.AgentMoney)
		if err3 == nil {
			for 局_索引 := range 代理分成数据 {
				d := 代理分成数据[局_索引] //太长了,放个变量里
				新余额, err2 = Ser_User.Id余额增减(d.Uid, d.S实际分成金额, true)
				if err2 != nil {
					//,一般不会出现,除非用户不存在
					global.GVA_LOG.Error(fmt.Sprintf("用户余额代理分成余额增加失败:%s,代理ID:%d,金额¥%v,卡号ID:%d", err2.Error(), d.Uid, d.S实际分成金额, 局_卡信息.Id))
				} else {
					str := fmt.Sprintf("用户余额制卡ID:%d,分成:¥%s (¥%s*(%d%%-%d%%)),|新余额≈%s", 局_卡信息.Id, utils.Float64到文本(d.S实际分成金额, 2), utils.Float64到文本(局_卡类.Money, 2), d.F分成百分比, d.F分给下级百分比, utils.Float64到文本(新余额, 2))
					Ser_Log.Log_写余额日志(Ser_User.Id取User(d.Uid), c.ClientIP(), str, d.S实际分成金额)
				}
			}
		}
		// 分成结束==============
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppId": 局_卡信息.AppId, "KaClassId": 局_卡信息.KaClassId, "KaClassName": 局_卡类.Name, "KaName": 局_卡信息.Name})

	局_文本 := fmt.Sprintf("自助购卡应用:%s,卡类:%s,消费:%.2f)", AppInfo.AppName, 局_卡类.Name, 局_卡类.Money)
	go Ser_Log.Log_写卡号操作日志(局_在线信息.User, c.ClientIP(), 局_文本, []string{局_卡信息.Name}, 1, 0)
	return
}

func UserApi_订单_余额充值(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if AppInfo.AppId < 10000 {
		response.X响应状态消息(c, response.Status_操作失败, "应用不存在")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetAliPayPC","User":"aaaaaa","Money":0.01,"PayType":"小叮当","Time":1684152719,"Status":15959}

	局_用户名 := string(请求json.GetStringBytes("User"))
	局_卡号 := Ser_AppInfo.App是否为卡号(AppInfo.AppId)

	var 局_Uid = 0
	var 局_Uid类型 = 0
	if 局_卡号 {
		局_Uid类型 = 2
		局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_用户名)
	} else {
		局_Uid类型 = 1
		局_Uid = Ser_User.User用户名取id(局_用户名)
	}

	if 局_Uid == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户不存在")
		return
	}

	var err error
	局_支付方式 := strings.TrimSpace(string(请求json.GetStringBytes("PayType")))
	//==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 局_Uid
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.ReceivedUid = 局_在线信息.AgentUid
	参数.Rmb = 请求json.GetFloat64("Money")
	参数.ProcessingType = constant.D订单类型_余额充值
	参数.E额外信息 = gjson.New("{}")
	err = 参数.E额外信息.Set("AppId", AppInfo.AppId)

	响应数据2, err := rmbPay.L_rmbPay.D订单创建(c, 参数)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "充值方式["+string(请求json.GetStringBytes("PayType"))+"]"+err.Error())
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 响应数据2)
	}
	return
}
func UserApi_取用户是否存在(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetIsUser","User":"13188888888"}
	局_用户名 := string(请求json.GetStringBytes("User"))
	局_Uid := 0
	if AppInfo.AppType == 3 || AppInfo.AppType == 4 {
		//卡号
		局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_用户名)
	} else {
		局_Uid = Ser_User.User用户名取id(局_用户名)
	}
	if 局_Uid == 0 {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"IsUser": false})
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"IsUser": true})
	}

	return
}
func UserApi_用户注册(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	fmt.Printf(c.GetString("局_json明文"))
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"NewUserInfo","User":"aaaaaa","PassWord":"a123456","Key":"123465","SuperPassWord":"123465"
	//,"Qq":"1059795985","Email":"测a试a个a中文1056795985@qq.com","Phone":"15666666666","Time":1684034325,"Status":17533}

	//检查是否可以绑定相同信息
	if AppInfo.IsUserKeySame == 2 && string(请求json.GetStringBytes("Key")) == "" {
		response.X响应状态消息(c, response.Status_操作失败, "绑定信息不能为空.")
		return
	}

	if AppInfo.IsUserKeySame == 2 && Ser_AppUser.B绑定信息是否存在(AppInfo.AppId, string(请求json.GetStringBytes("Key"))) {
		response.X响应状态消息(c, response.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
		return
	}

	if blacklist.Is黑名单(string(请求json.GetStringBytes("Key")), AppInfo.AppId) {
		response.X响应状态消息(c, response.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}

	err := Ser_User.New用户信息(string(请求json.GetStringBytes("User")), string(请求json.GetStringBytes("PassWord")), string(请求json.GetStringBytes("SuperPassWord")), string(请求json.GetStringBytes("Qq")), string(请求json.GetStringBytes("Email")), string(请求json.GetStringBytes("Phone")), c.ClientIP(), "", 0, 0, 0, "")
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	var 局_VipNumber int64
	if !Ser_AppInfo.App是否为计点(AppInfo.AppId) {
		局_VipNumber = time.Now().Unix()
	} else {
		局_VipNumber = 0
	}

	//没有这个用户,应该是第一次登录应用,添加进去
	err = Ser_AppUser.New用户信息(AppInfo.AppId, Ser_User.User用户名取id(string(请求json.GetStringBytes("User"))), string(请求json.GetStringBytes("Key")), AppInfo.MaxOnline, 局_VipNumber, 0, 0, "")
	if err != nil {
		response.X响应状态消息(c, response.Status_SQl错误, "New用户信息内部错误,用户注册成功,注册软件用户失败")
		return
	}

	// 注册送卡
	if AppInfo.RegisterGiveKaClassId > 0 {
		局_注册送卡, 局_制卡结果 := Ser_Ka.Ka单卡创建(AppInfo.RegisterGiveKaClassId, "系统自动", "用户注册系统自动制卡赠送充值", "", 0)
		if 局_制卡结果 == nil {
			_, _ = Ser_Ka.K卡号充值_事务(AppInfo.AppId, 局_注册送卡.Name, string(请求json.GetStringBytes("User")), "", c.ClientIP())
		}
	}

	response.X响应状态消息(c, c.GetInt("局_成功Status"), "注册成功")
	return
}
func 检测_账密模式专用(c *gin.Context, AppInfo DB.DB_AppInfo) bool {
	if AppInfo.AppType > 2 {
		response.X响应状态消息(c, response.Status_操作失败, "本接口仅限应用账密模式可用")
		return false
	}
	return true
}
