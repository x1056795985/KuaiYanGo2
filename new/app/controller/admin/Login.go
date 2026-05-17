package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"server/Service/Captcha"
	"server/Service/Ser_Log"
	"server/global"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"server/utils/Qqwry"
	"strings"
	"time"
)

type LoginCtrl struct {
	Common.Common
}

func NewLoginController() *LoginCtrl {
	return &LoginCtrl{}
}

// 登录请求结构体
type 结构_登录请求 struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Captcha   string `json:"captcha"`
	CaptchaId string `json:"captchaId"`
}

// 登录响应结构体
type 结构_登录响应 struct {
	UserInfo DB.DB_Admin `json:"userInfo"`
	Token    string      `json:"token"`
	KuaiYan  bool        `json:"kuaiYan"`
}

// Login 管理员登录
func (l *LoginCtrl) Login(c *gin.Context) {
	var Request 结构_登录请求
	err := c.ShouldBindJSON(&Request)
	客户端ip := c.ClientIP()

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 判断验证码是否开启
	openCaptcha := global.GVA_CONFIG.Captcha.OpenCaptcha
	openCaptchaTimeOut := global.GVA_CONFIG.Captcha.OpenCaptchaTimeOut
	v, ok := global.H缓存.Get(客户端ip)
	if !ok {
		global.H缓存.Set(客户端ip, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}

	var j校验验证码 = false
	if openCaptcha == 0 || openCaptcha < interfaceToInt(v) {
		j校验验证码 = true
	}

	_ = global.H缓存.Increment(客户端ip, 2)
	if j校验验证码 {
		if !Captcha.Captcha_Verify点选(Request.CaptchaId, Request.Captcha, true) {
			response.FailWithMessage("验证码错误", c)
			go Ser_Log.Log_写登录日志(Request.Username, c.ClientIP(), "验证码错误:"+Request.Captcha, 4)
			return
		}
	}

	if nil == global.GVA_DB {
		response.FailWithMessage("请先初始化数据库", c)
		return
	}

	var DB_user DB.DB_Admin
	err = global.GVA_DB.Where("User = ?", Request.Username).First(&DB_user).Error

	if err != nil || !utils.BcryptCheck(Request.Password, DB_user.PassWord) {
		if global.GVA_Viper.GetInt("系统模式") == 1 {
			response.FailWithMessage("账号或密码错误,当前为演示模式,账密都是 admin", c)
		} else {
			response.FailWithMessage("账号或密码错误", c)
		}
		go Ser_Log.Log_写登录日志(Request.Username, c.ClientIP(), "密码错误:"+Request.Password, 4)
		return
	}

	if DB_user.Status != 1 {
		response.FailWithMessage("用户被禁止登录", c)
		go Ser_Log.Log_写登录日志(Request.Username, c.ClientIP(), "用户被禁止登录", 4)
		return
	}
	global.H缓存.Delete(客户端ip)
	var DB_links_user DB.DB_LinksToken
	DB_links_user.Uid = DB_user.Id
	DB_links_user.User = DB_user.User
	DB_links_user.Tab = ""
	DB_links_user.Key = ""
	DB_links_user.Ip = c.ClientIP()
	省市, 运行商, err := Qqwry.Ip查信息(DB_links_user.Ip)
	if err == nil && 省市 != "" {
		DB_links_user.IPCity = 省市 + " " + 运行商
	}
	DB_links_user.Status = DB_user.Status
	DB_links_user.LoginTime = time.Now().Unix()
	DB_links_user.OutTime = 36000
	DB_links_user.LastTime = DB_links_user.LoginTime
	DB_links_user.Token = strings.ToUpper(rand_string.RandomLetter(32))
	DB_links_user.LoginAppid = 1 //管理员后台代号1

	err = global.GVA_DB.Create(&DB_links_user).Error
	go Ser_Log.Log_写登录日志(Request.Username, c.ClientIP(), "管理平台登录", 4)
	快验 := global.Q快验.Q取登录状态()

	response.OkWithDetailed(结构_登录响应{
		UserInfo: DB_user,
		Token:    DB_links_user.Token,
		KuaiYan:  !快验,
	}, "登录成功", c)
}
