package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"server/Service/Captcha"
	adminController "server/new/app/controller/admin"
	"server/new/app/global"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/log"
	"server/new/app/models/constant"
	"server/new/app/models/db"
	"server/new/app/models/old/response"
	"server/new/app/utils"
	"server/new/app/utils/Qqwry"
	"strings"
	"time"
)

type AgentBase struct{}

func NewAgentBaseController() *AgentBase {
	return &AgentBase{}
}

type Agent登录请求 struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Captcha   string `json:"captcha"`
	CaptchaId string `json:"captchaId"`
}

type Agent登录响应 struct {
	UserInfo db.DB_User `json:"userInfo"`
	Token    string     `json:"token"`
	KuaiYan  bool       `json:"kuaiYan"`
}

func (A *AgentBase) Captcha(c *gin.Context) {
	adminController.NewBaseController().Captcha2(c)
}

func (A *AgentBase) Login(c *gin.Context) {
	var 局_请求 Agent登录请求
	if err := c.ShouldBindJSON(&局_请求); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	局_客户端IP := c.ClientIP()
	局_是否校验验证码 := false
	局_开启验证码次数 := 1
	局_缓存超时时间 := 3600
	局_缓存值, ok := global.H缓存.Get(局_客户端IP)
	if !ok {
		global.H缓存.Set(局_客户端IP, 1, time.Second*time.Duration(局_缓存超时时间))
	}
	if 局_开启验证码次数 == 0 || 局_开启验证码次数 < interfaceToInt(局_缓存值) {
		局_是否校验验证码 = true
	}
	_ = global.H缓存.Increment(局_客户端IP, 1)
	if 局_是否校验验证码 {
		if !Captcha.Captcha_Verify点选(局_请求.CaptchaId, 局_请求.Captcha, true) {
			response.FailWithMessage("验证码错误", c)
			go log.L_log.Log_写登录日志(局_请求.Username, 局_客户端IP, "验证码错误:"+局_请求.Captcha, 3)
			return
		}
	}
	if global.GVA_DB == nil {
		response.FailWithMessage("请先初始化数据库", c)
		return
	}

	var 局_用户 db.DB_User
	if err := global.GVA_DB.Where("User = ?", 局_请求.Username).First(&局_用户).Error; err != nil || !utils.BcryptCheck(局_请求.Password, 局_用户.PassWord) {
		response.FailWithMessage("账号或密码错误", c)
		go log.L_log.Log_写登录日志(局_请求.Username, 局_客户端IP, "密码错误:"+局_请求.Password, 3)
		return
	}
	if 局_用户.Status != 1 {
		response.FailWithMessage("用户被禁止登录", c)
		go log.L_log.Log_写登录日志(局_请求.Username, 局_客户端IP, "用户被禁止登录代理平台", 3)
		return
	}

	局_代理级别 := agentLevel.L_agentLevel.Q取Id代理级别(c, 局_用户.Id)
	if 局_代理级别 < 1 || 局_代理级别 > 3 {
		response.FailWithMessage("非代理用户禁止登录平台", c)
		return
	}
	global.H缓存.Delete(局_客户端IP)

	var 局_在线信息 db.DB_LinksToken
	局_在线信息.Uid = 局_用户.Id
	局_在线信息.User = 局_用户.User
	局_在线信息.Tab = ""
	局_在线信息.Key = ""
	局_在线信息.Ip = 局_客户端IP
	局_省市, 局_运行商, err := Qqwry.Ip查信息(局_在线信息.Ip)
	if err == nil && 局_省市 != "" {
		局_在线信息.IPCity = 局_省市 + " " + 局_运行商
	}
	局_在线信息.Status = 局_用户.Status
	局_在线信息.LoginTime = time.Now().Unix()
	局_在线信息.OutTime = 36000
	局_在线信息.LastTime = 局_在线信息.LoginTime
	局_在线信息.Token = strings.ToUpper(rand_string.RandomLetter(32))
	局_在线信息.LoginAppid = constant.APPID_代理平台

	if err = global.GVA_DB.Create(&局_在线信息).Error; err != nil {
		response.FailWithMessage("登录失败:"+err.Error(), c)
		return
	}

	go log.L_log.Log_写登录日志(局_请求.Username, 局_客户端IP, "代理平台登录", 局_代理级别)
	response.OkWithDetailed(Agent登录响应{
		UserInfo: 局_用户,
		Token:    局_在线信息.Token,
		KuaiYan:  false,
	}, "登录成功", c)
}

func interfaceToInt(v interface{}) int {
	switch 局_值 := v.(type) {
	case int:
		return 局_值
	default:
		return 0
	}
}
