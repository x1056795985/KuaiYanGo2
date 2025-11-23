package UserApi

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/Captcha"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/api/UserApi/response"
	DB "server/structs/db"
)

func KyApiSendSms(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"KyApiSendSms","Code":["123456"],"Phone":"13100000000"}
	var 局_User DB.DB_User
	局_User, ok := Ser_User.Id取详情(局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}
	//局_错误信息 := ""
	局_手机号 := string(请求json.GetStringBytes("Phone"))
	/*	if !Z正则_校验手机号(局_手机号, &局_错误信息) {
		response.X响应状态消息(c, response.Status_操作失败, 局_错误信息)
		return
	}*/
	局_参数验证码 := string(请求json.GetStringBytes("Code", "0"))
	if len(局_参数验证码) < 1 || len(局_参数验证码) > 6 {
		response.X响应状态消息(c, response.Status_操作失败, "验证码长度为1-6")
		return
	}

	var 局_增减值 float64
	局_增减值 = 0.04 //短信价格0.47 //mark  正常一条0.047 但是系统只能计费到分,先这样,以后再说,下次重构系统,小数点要保留后4位备用 //更新 纠结,易语言只有双精度,4位易语言用不了,还是算了
	if 局_User.Rmb < 局_增减值 {
		response.X响应状态消息(c, response.Status_操作失败, "余额不足")
		return
	}

	新余额, err := Ser_User.Id余额增减(局_User.Id, 局_增减值, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error()) //基本就是余额不足
		return
	}
	go Ser_Log.Log_写余额日志(局_User.User, c.ClientIP(), fmt.Sprintf("%s|新余额%v", "快验系统ApiSendSms"+局_手机号+","+局_参数验证码, 新余额), Float64取负值(局_增减值))
	err = Captcha.Sms_当前选择发送短信验证码([]string{局_参数验证码}, 局_手机号)
	if err == nil {
		response.X响应状态(c, c.GetInt("局_成功Status"))
		return
	}
	response.X响应状态消息(c, response.Status_操作失败, err.Error())
	//补偿扣款
	新余额2, err2 := Ser_User.Id余额增减(局_User.Id, 局_增减值, true) // 'mark 隐患,增加值会失败,后期重构放事务内'
	if err2 != nil {
		局_log := fmt.Sprintf("%s|金额%v", "快验系统ApiSendSms"+局_手机号+","+局_参数验证码+"发送失败补偿单失败了,原因"+err2.Error(), 局_增减值)
		Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, AppInfo.AppId, 局_User.User, AppInfo.AppName, 局_在线信息.AppVer, 局_log, c.ClientIP())
	} else {
		go Ser_Log.Log_写余额日志(局_User.User, c.ClientIP(), fmt.Sprintf("%s|新余额%v", "快验系统ApiSendSms"+局_手机号+","+局_参数验证码+"发送失败补偿,原因"+err.Error()+"", 新余额2), 局_增减值)
	}

}

func K快验_极验验证码结果验证(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"KyApiJiYanVerifyTicket","CaptchaId":"123456","CaptchaValue":"ad16w41da135sdad"}

	局_结果 := Captcha.J极验_滑动验证码参数验证(
		string(请求json.GetStringBytes("CaptchaId")),
		string(请求json.GetStringBytes("CaptchaValue")),
	)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Code": 局_结果 == nil})
	return
}
