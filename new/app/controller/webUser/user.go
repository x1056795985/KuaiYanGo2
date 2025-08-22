package controller

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Captcha"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/webUser/cpsInvitingRelation"
	"server/new/app/models/constant"
	"server/new/app/service"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Common.Common
}

func NewUserController() *User {
	return &User{}
}

// 注册
func (C *User) NewUserInfo(c *gin.Context) {
	var 请求 struct {
		User          string `json:"user" binding:"required,min=6,max=190" zh:"用户名"`    // 用户名
		Password      string `json:"password" binding:"required,min=6,max=190" zh:"密码"` // 密码
		QQ            string `json:"qq"`
		Phone         string `json:"phone"`
		SuperPassword string `json:"superPassword"   zh:"超级密码"` // 密码`
		Email         string `json:"email"`
		PromotionCode int    `json:"promotionCode"`
		AppId         int    `json:"appId"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		user DB.DB_User
	}{}
	var err error
	tx := *global.GVA_DB
	局_临时文本 := ""
	if !utils2.Z正则_校验密码(请求.Password, &局_临时文本) {
		response.FailWithMessage(c, "密码"+局_临时文本)
		return
	}

	if 请求.SuperPassword != "" && !utils2.Z正则_校验密码(请求.SuperPassword, &局_临时文本) {
		response.FailWithMessage(c, "超级密码"+局_临时文本)
		return
	}

	info.user.User = 请求.User
	info.user.PassWord = 请求.Password
	info.user.Qq = 请求.QQ
	info.user.Phone = 请求.Phone
	info.user.RegisterIp = c.ClientIP()
	info.user.RegisterTime = time.Now().Unix()
	//如果为空, 则说明客户不需要超级密码修改密码功能,直接随机一个防止密码被猜到
	info.user.SuperPassWord = S三元(请求.SuperPassword == "", W文本_取随机字符串(24), 请求.SuperPassword)
	info.user.Email = 请求.Email

	_, err = service.NewUser(c, &tx).Create(&info.user)
	if err != nil {
		//Duplicate entry 'aaaaaa' for key 'User'"
		response.FailWithMessage(c, "用户名已存在")
		return
	}

	response.OkWithMessage(c, "注册成功")
	//如果appid和邀请码有值,则进行邀请码处理
	if 请求.AppId > 10000 && 请求.PromotionCode > 0 && info.user.Id > 0 && 请求.PromotionCode != info.user.Id {
		err = cpsInvitingRelation.L_CpsInvitingRelation.S设置邀请人(c, 请求.AppId, 请求.PromotionCode, info.user.Id, c.GetHeader("Referer"))
		//忽略错误
	}

	return
}

// 密码找回或修改_密保手机
func (C *User) GetPwSendSms(c *gin.Context) {
	var 请求 struct {
		User string `json:"user" binding:"required,min=6,max=190" zh:"用户名"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		user DB.DB_User
	}{}
	var err error
	tx := *global.GVA_DB
	info.user, err = service.NewUser(c, &tx).InfoName(请求.User)
	if err != nil {
		response.FailWithMessage(c, "用户不存在")
		return
	}
	if info.user.UPAgentId > 0 {
		response.FailWithMessage(c, "代理账号,不可操作")
		return
	}
	局_msg := "绑定手机非正确手机号格式"
	if !utils2.Z正则_校验手机号(info.user.Phone, &局_msg) {
		response.FailWithMessage(c, 局_msg)
		return
	}

	局_验证码 := W文本_取随机字符串_数字(6)
	局_验证码ID := "Note" + utils2.Md5String(info.user.Phone)[:16] + W文本_取随机字符串(15)
	err = Captcha.Sms_当前选择发送短信验证码([]string{局_验证码}, info.user.Phone)
	if err != nil {
		Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 请求.User, strconv.Itoa(constant.APPID_Web用户中心), "", fmt.Sprintf("短信验证码发送失败:%v,%v,%v", 局_验证码, info.user.Phone, err.Error()), c.ClientIP())
		response.FailWithMessage(c, "发送失败")
		return
	}
	Captcha.H缓存验证码校验实例.Set(局_验证码ID, 局_验证码)
	response.OkWithData(c, gin.H{"captchaType": 3, "captchaId": 局_验证码ID})
	return
}
func (C *User) GetInfo(c *gin.Context) {
	var info = struct {
		user     DB.DB_User
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
	}{}
	var err error
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	info.user, err = service.NewUser(c, &tx).Info(info.likeInfo.Uid)
	if err != nil {
		response.FailWithMessage(c, "用户不存在")
		return
	}
	if info.user.UPAgentId > 0 {
		response.FailWithMessage(c, "代理账号,不可操作")
		return
	}

	response.OkWithData(c, gin.H{
		"user":                info.user.User,
		"phone":               info.user.Phone,
		"qq":                  info.user.Qq,
		"email":               info.user.Email,
		"upAgentId":           info.user.UPAgentId,
		"realNameAttestation": info.user.RealNameAttestation,
		"loginIp":             info.user.LoginIp,
		"loginTime":           info.user.LoginTime,
		"registerIp":          info.user.RegisterIp,
		"registerTime":        info.user.RegisterTime,
	})

	return
}

func (C *User) SmsCodeSetPassWord(c *gin.Context) {
	var 请求 struct {
		User              string `json:"user" binding:"required,min=6,max=190" zh:"用户名"`
		NewPassWord       string `json:"newPassword" binding:"required,min=6,max=18" zh:"密码"`
		PhoneCaptchaId    string `json:"phoneCaptchaId" binding:"required,min=6,max=180" zh:"验证码ID"`
		PhoneCaptchaValue string `json:"phoneCaptchaValue" binding:"required,min=6,max=18" zh:"验证码"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		user DB.DB_User
	}{}
	var err error

	局_临时文本 := ""
	if !utils2.Z正则_校验密码(请求.NewPassWord, &局_临时文本) {
		response.FailWithMessage(c, 局_临时文本)
		return
	}
	tx := *global.GVA_DB
	info.user, err = service.NewUser(c, &tx).InfoName(请求.User)
	if err != nil {
		response.FailWithMessage(c, "用户不存在")
		return
	}

	if strings.Index(请求.PhoneCaptchaId, "Note") != 0 {
		go Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, 请求.User, c.ClientIP(), "使用绑定手机密码找回或修改,用户使用非短信验证码Id进行提交,可能是异常用户")
		response.FailWithMessage(c, "验证码错误.")
		return
	}

	if info.user.Phone == "" || strings.Index(请求.PhoneCaptchaId, "Note"+utils2.Md5String(info.user.Phone)[:16]) == -1 {
		go Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, info.user.User, c.ClientIP(), "使用绑定手机密码找回或修改,用户使用非账号绑定的验证码进行提交,可能是异常用户")
		response.FailWithMessage(c, "验证码错误.")
		return
	}
	if !Captcha.H缓存验证码校验实例.Verify(请求.PhoneCaptchaId, 请求.PhoneCaptchaValue, false) {
		response.FailWithMessage(c, "短信验证码错误.")
		return
	}

	_, err = service.NewUser(c, &tx).Update(info.user.Id, map[string]interface{}{"PassWord": utils2.Md5String(请求.NewPassWord)})
	if err != nil {
		response.FailWithMessage(c, "修改失败")
	} else {
		_ = Ser_LinkUser.Set批量注销Uid(info.user.Id, Ser_LinkUser.Z注销_用户改密注销)
		response.OkWithMessage(c, "修改成功")
	}
	return
}

func (C *User) Logout(c *gin.Context) {
	var err error
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)

	err = Ser_LinkUser.Set批量注销([]int{info.likeInfo.Id}, Ser_LinkUser.Z注销_用户操作注销)
	if err != nil {
		response.FailWithMessage(c, "注销失败")
		return
	}
	response.OkWithMessage(c, "注销成功")
	return
}
func (C *User) SetBaseInfo(c *gin.Context) {
	var 请求 struct {
		Type         string `json:"type" binding:"required" zh:"类型"`
		Value        string `json:"value" binding:"required" zh:"值"`
		CaptchaId    string `json:"captchaId" binding:"" zh:"验证码id"`
		CaptchaValue string `json:"captchaValue" binding:"" zh:"验证码值"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	if !Y限账号模式应用(c, &info.appInfo) {
		return
	}
	tx := *global.GVA_DB
	switch 请求.Type {
	case "email":
		_, err = service.NewUser(c, &tx).Update(info.likeInfo.Uid, map[string]interface{}{"Email": 请求.Value})
	case "qq":
		_, err = service.NewUser(c, &tx).Update(info.likeInfo.Uid, map[string]interface{}{"Qq": 请求.Value})
	case "phone":
		//校验验证码是否正确,

		if !Captcha.H缓存验证码校验实例.Verify(请求.CaptchaId, 请求.CaptchaValue, false) {
			response.FailWithMessage(c, "短信验证码错误.")
			return
		}
		if 请求.Value == "" || strings.Index(请求.CaptchaId, "Note"+utils2.Md5String(请求.Value)[:16]) == -1 {
			go Ser_Log.Log_写风控日志(info.likeInfo.Id, Ser_Log.Log风控类型_Api异常调用, info.likeInfo.User, c.ClientIP(), "用户使用非新手机号绑定的验证码进行提交,更换绑定手机,可能是异常用户")
			response.FailWithMessage(c, "验证码错误.")
			return
		}
		_, err = service.NewUser(c, &tx).Update(info.likeInfo.Uid, map[string]interface{}{"Phone": 请求.Value})

	}
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithMessage(c, "注销成功")
	return
}

func (C *User) SendSms(c *gin.Context) {
	var 请求 struct {
		Phone string `json:"phone" binding:"required,len=11" zh:"手机号"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	if !Y限账号模式应用(c, &info.appInfo) {
		return
	}
	局_msg := "手机号码非正确手机号格式"
	if !utils2.Z正则_校验手机号(请求.Phone, &局_msg) {
		response.FailWithMessage(c, 局_msg)
		return
	}
	局_验证码 := W文本_取随机字符串_数字(6)
	局_验证码ID := "Note" + utils2.Md5String(请求.Phone)[:16] + W文本_取随机字符串(15)
	err = Captcha.Sms_当前选择发送短信验证码([]string{局_验证码}, 请求.Phone)
	if err != nil {
		Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, info.likeInfo.User, strconv.Itoa(constant.APPID_Web用户中心), "", fmt.Sprintf("短信验证码发送失败:%v,%v,%v", 局_验证码, 请求.Phone, err.Error()), c.ClientIP())
		response.FailWithMessage(c, "发送失败")
		return
	}
	Captcha.H缓存验证码校验实例.Set(局_验证码ID, 局_验证码)
	response.OkWithData(c, gin.H{"captchaType": 3, "captchaId": 局_验证码ID})
	return
}
