package userSafetyApi

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Captcha"
	"server/global"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/logic/common/log"
	"server/new/app/logic/common/user"
	"server/new/app/models/constant"
	"server/new/app/service"
	DB "server/structs/db"
	"server/utils/Qqwry"
	"time"
)

func KyApiSendSms(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	//{"Api":"KyApiSendSms","Code":["123456"],"Phone":"13100000000"}

	局_User, err := service.NewUser(c, &db).Info(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}
	//局_错误信息 := ""
	局_手机号 := 局_ctx.Q请求明文.Get("Phone").String()
	/*	if !Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.FailMsg(c, response.Status_操作失败, 局_错误信息)
			return
	}*/
	局_参数验证码 := 局_ctx.Q请求明文.Get("Code", "0").String()
	if len(局_参数验证码) < 1 || len(局_参数验证码) > 6 {
		response.FailMsg(c, constant.Status_操作失败, "验证码长度为1-6")
		return
	}

	var 局_增减值 float64
	局_增减值 = 0.06 //
	if 局_User.Rmb < 局_增减值 {
		response.FailMsg(c, constant.Status_操作失败, "余额不足")
		return
	}

	新余额, err := user.L_user.Id余额增减(c, 局_User.Id, 局_增减值, false)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error()) //基本就是余额不足
		return
	}
	go log.L_log.S输出日志(c, DB.DB_LogMoney{
		User:  局_User.User,
		Ip:    c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP()),
		Time:  time.Now().Unix(),
		Count: Float64取负值(局_增减值),
		Note:  fmt.Sprintf("%s|新余额%v", "快验系统ApiSendSms"+局_手机号+","+局_参数验证码, 新余额),
	})
	err = Captcha.Sms_当前选择发送短信验证码([]string{局_参数验证码}, 局_手机号)
	if err == nil {
		response.Ok(c)
		return
	}
	response.FailMsg(c, constant.Status_操作失败, err.Error())
	//补偿扣款
	新余额2, err2 := user.L_user.Id余额增减(c, 局_User.Id, 局_增减值, true) // 'mark 隐患,增加值会失败,后期重构放事务内'
	if err2 != nil {
		局_log := fmt.Sprintf("%s|金额%v", "快验系统ApiSendSms"+局_手机号+","+局_参数验证码+"发送失败补偿单失败了,原因"+err2.Error(), 局_增减值)
		_ = log.L_log.S输出日志(c, DB.DB_LogUserMsg{
			User:    "系统",
			App:     局_ctx.AppInfo.AppName,
			AppId:   局_ctx.AppInfo.AppId,
			AppVer:  局_ctx.Z在线信息.AppVer,
			MsgType: 4, // Ser_Log.Log用户消息类型_系统执行错误
			Time:    time.Now().Unix(),
			Ip:      c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP()),
			Note:    局_log,
		})
	} else {
		go log.L_log.S输出日志(c, DB.DB_LogMoney{
			User:  局_User.User,
			Ip:    c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP()),
			Time:  time.Now().Unix(),
			Count: 局_增减值,
			Note:  fmt.Sprintf("%s|新余额%v", "快验系统ApiSendSms"+局_手机号+","+局_参数验证码+"发送失败补偿,原因"+err.Error()+"", 新余额2),
		})
	}

}

func K快验_极验验证码结果验证(c *gin.Context) {
	局_ctx := 取上下文(c)
	//{"Api":"KyApiJiYanVerifyTicket","CaptchaId":"123456","CaptchaValue":"ad16w41da135sdad"}

	局_结果 := Captcha.J极验_滑动验证码参数验证(
		局_ctx.Q请求明文.Get("CaptchaId").String(),
		局_ctx.Q请求明文.Get("CaptchaValue").String(),
	)

	response.OkData(c, gin.H{"Code": 局_结果 == nil})
	return
}
