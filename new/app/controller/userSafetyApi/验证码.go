package userSafetyApi

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Captcha"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/models/constant"
	"server/new/app/service"
	utils2 "server/new/app/utils"
)

// UserApi_取验证码信息 取验证码信息
func UserApi_取验证码信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	//{"Api":"GetCaptcha","CaptchaType":2}

	if 局_ctx.Q请求明文.Get("Captcha").Int() == 2 {
		response.FailMsg(c, constant.Status_操作失败, "滑动无需获取,直接置值验证即可")
		return
	}
	局_验证码id, 局_base64验证码内容, err := Captcha.Captcha_取英数验证码()
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "生成失败")
	}
	response.OkData(c, gin.H{"CaptchaType": 1, "CaptchaId": 局_验证码id, "CaptChaImg": 局_base64验证码内容})
	return
}

// UserApi_取短信验证码信息 取短信验证码信息
func UserApi_取短信验证码信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"GetPhoneCaptcha","Phone":"13188888888","User":"13188888888"}

	局_手机号 := 局_ctx.Q请求明文.Get("Phone").String()

	局_错误信息 := ""
	if 局_手机号 == "" {
		db := *global.GVA_DB
		局_User, err := service.NewUser(c, &db).InfoName(局_ctx.Q请求明文.Get("User").String())
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, "用户不存在")
			return
		}
		局_手机号 = 局_User.Phone
		if !utils2.Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.FailMsg(c, constant.Status_参数错误, "用户绑定手机号格式不正确")
			return
		}
	} else {
		if !utils2.Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.FailMsg(c, constant.Status_参数错误, 局_错误信息)
			return
		}
	}

	局_验证码 := W文本_取随机字符串_数字(6)
	局_验证码ID := "Note" + utils2.Md5String(局_手机号)[:16] + W文本_取随机字符串(15)

	err := Captcha.Sms_当前选择发送短信验证码([]string{局_验证码}, 局_手机号)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("短信验证码发送失败:%v,%v,%v", 局_验证码, 局_手机号, err.Error()))
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}
	Captcha.H缓存验证码校验实例.Set(局_验证码ID, 局_验证码)
	response.OkData(c, gin.H{"CaptchaType": 3, "CaptchaId": 局_验证码ID})
	return
}
