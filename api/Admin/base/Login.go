package base

import (
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"server/Service/Captcha"
	"server/Service/Ser_Log"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"server/utils/Qqwry"
	"strings"
	"time"
)

// Login
// @Tags     Base
// @Summary  用户登录
// @Produce   application/json
// @Param    data  body      systemReq.Login                                             true  "用户名, 密码, 验证码"
// @Success  200   {object}  response.Response{data=systemRes.结构_,msg=string}  "返回包括用户信息,token,过期时间"
// @Router   /base/login [post]
func (b *BaseApi) Login(c *gin.Context) {
	var Request 结构_登录请求
	err := c.ShouldBindJSON(&Request)
	客户端ip := c.ClientIP()

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	//校验数据格式 ,感觉没什么用先不校验了
	//err = utils.Verify(Request, utils.LoginVerify)
	//if err != nil {
	//	response.FailWithMessage(err.Error(), c)
	//	return
	//}

	// 判断验证码是否开启
	openCaptcha := global.GVA_CONFIG.Captcha.OpenCaptcha               // 是否开启防暴次数
	openCaptchaTimeOut := global.GVA_CONFIG.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	v, ok := global.H缓存.Get(客户端ip)                                // 获取这个ip已经被请求次数
	if !ok {
		// 获取这个ip已经被请求次数  如果没请求过, 设置值为1
		global.H缓存.Set(客户端ip, 1, time.Second*time.Duration(openCaptchaTimeOut))

	}

	//如果 防暴次数次数=0  或 已请求次数大于 防暴次数  校验验证码
	var j校验验证码 = false
	if openCaptcha == 0 || openCaptcha < interfaceToInt(v) {
		j校验验证码 = true
	}

	_ = global.H缓存.Increment(客户端ip, 1) //这个ip防爆次数 + 1
	// j校验验证码
	if j校验验证码 {
		//验证码验证码正确 = 真
		if !Captcha.H缓存验证码校验实例.Verify(Request.CaptchaId, Request.Captcha, true) {
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

	// 没查到数据  或  取反(密码正确)
	if err != nil || !utils.BcryptCheck(Request.Password, DB_user.PassWord) {
		if global.GVA_CONFIG.X系统设置.W系统模式 == 1 {
			response.FailWithMessage("账号或密码错误,当前为演示模式,账密都是 Admin", c)
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
	global.H缓存.Delete(客户端ip) //重置防暴次数
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
	DB_links_user.OutTime = 36000 //退出时间
	DB_links_user.LastTime = DB_links_user.LoginTime
	//DB_links_user.Token = utils.BcryptHash(DB_links_user.User + string(DB_links_user.LoginTime) + string(DB_links_user.OutTIme) + DB_links_user.Key + rand_string.RandStringBytesMaskImprSrc(25))
	DB_links_user.Token = strings.ToUpper(rand_string.RandStringBytesMaskImprSrc(32))
	DB_links_user.LoginAppid = 1 //管理员后台代号1

	err = global.GVA_DB.Create(&DB_links_user).Error
	go Ser_Log.Log_写登录日志(Request.Username, c.ClientIP(), "管理平台登录", 4)
	快验 := global.X系统信息.D到期时间 < time.Now().Unix()
	if global.GVA_CONFIG.X系统设置.W系统模式 == 1 {
		快验 = false
	}
	response.OkWithDetailed(结构_登录响应{
		UserInfo: DB_user,
		Token:    DB_links_user.Token,
		KuaiYan:  快验,
	}, "登录成功", c)
	return

}

// 登录请求结构体
type 结构_登录请求 struct {
	Username  string `json:"username"`  // 用户名
	Password  string `json:"password"`  // 密码
	Captcha   string `json:"captcha"`   // 验证码
	CaptchaId string `json:"captchaId"` // 验证码ID

}

// 登录响应结构体
type 结构_登录响应 struct {
	UserInfo DB.DB_Admin `json:"UserInfo"`
	Token    string      `json:"Token"`
	KuaiYan  bool        `json:"KuaiYan"`
}
