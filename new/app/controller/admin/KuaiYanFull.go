package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/KuaiYanUpdater"
	"server/Service/Ser_Ka"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
	utils2 "server/utils"
)

type KuaiYan struct {
	Common.Common
}

func NewKuaiYanController() *KuaiYan {
	return &KuaiYan{}
}

type 结构请求_个人信息 struct {
	User     string `json:"user"`
	PassWord string `json:"passWord"`
	Qq       string `json:"qq"`
	CaptCha1 struct {
		CaptchaId    string `json:"captchaId"`
		CaptChaValue string `json:"captChaValue"`
	} `json:"captCha1"`
	CaptCha2 struct {
		CaptchaId    string `json:"captchaId"`
		CaptChaValue string `json:"captChaValue"`
	} `json:"captCha2"`
	CaptCha3 struct {
		CaptchaId    string `json:"captchaId"`
		CaptChaValue string `json:"captChaValue"`
	} `json:"captCha3"`
}

type 结构请求_KaCLassId struct {
	KaCLassId int `json:"kaCLassId"`
}

type 结构请求_充值 struct {
	Ka         string `json:"ka"`
	InviteUser string `json:"inviteUser"`
}

type 结构请求_余额充值 struct {
	Type  string  `json:"type"`
	C充值金额 float64 `json:"rmb"`
	D订单ID string  `json:"orderId"`
}

// GetCaptchaApiList 取开启验证码接口列表
func (k *KuaiYan) GetCaptchaApiList(c *gin.Context) {
	var 跳转过标记 = false
标记:
	var 响应信息 string
	if !global.Q快验.Q取开启验证码接口列表(&响应信息) {
		var 错误代码 = 0
		global.Q快验.Q取错误信息(&错误代码)
		if 错误代码 == 109 && !跳转过标记 {
			global.Q快验.J_Token = ""
			跳转过标记 = true
			goto 标记
		}
		response.FailWithMessage("取开启验证码接口列表失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	global.X系统信息.K开启验证码接口列表 = 响应信息
	response.OkWithDetailed(gin.H{"map": 响应信息}, "获取成功", c)
	return
}

// GetCaptcha 取英数验证码
func (k *KuaiYan) GetCaptcha(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 跳转过标记 = false
标记:
	var 响应信息 string
	if !global.Q快验.Q取验证码(&响应信息, 1) {
		var 错误代码 = 0
		global.Q快验.Q取错误信息(&错误代码)
		if 错误代码 == 109 && !跳转过标记 {
			global.Q快验.J_Token = "" //后台被强行提出就会失败 需要重新获取token
			跳转过标记 = true
			goto 标记
		}
		response.FailWithMessage("获取英数验证码失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, _ := fastjson.Parse(响应信息)
	response.OkWithDetailed(gin.H{"captchaId": string(局_json.GetStringBytes("CaptchaId")), "picPath": string(局_json.GetStringBytes("CaptChaImg"))}, "获取成功", c)
	return
}

// GetUserInfo 快验个人信息更新
func (k *KuaiYan) GetUserInfo(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 文件修改日期 = ""
	if len(os.Args) > 0 {
		文件修改日期 = os.Args[0]
		文件信息, err1 := os.Stat(文件修改日期)
		if 文件信息 != nil {
			if false {
				// linux only
			} else {
				文件修改日期 = 文件信息.ModTime().String()
			}
		} else {
			文件修改日期 = err1.Error()
		}
		if len(文件修改日期) > 19 {
			文件修改日期 = 文件修改日期[:19]
		}
	}

	var 局_软件用户信息 string
	if !global.Q快验.Q取软件用户信息(&局_软件用户信息, global.X系统信息.B版本号当前) {
		global.Q快验.J_Token = ""
		response.OkWithDetailed(gin.H{
			"user": "", "userClassName": "", "vipNumber": 0, "rmb": 0,
			"vipTime": 0, "email": "", "loginTime": 0, "loginIp": "",
			"qq": "", "registerTime": 0, "appVer": global.X系统信息.B版本号当前,
			"appVerUpdateTime": 文件修改日期, "appVerNew": global.X系统信息.B版本号最新,
			"linkTokenCount": 0, "agentUid": global.GVA_Viper.GetInt("duid"),
		}, "获取成功", c)
		return
	}

	局_应用用户信息json, _ := fastjson.Parse(局_软件用户信息)
	global.X系统信息.D到期时间 = 局_应用用户信息json.GetInt64("VipTime")
	global.X系统信息.B绑定信息 = string(局_应用用户信息json.GetStringBytes("Key"))
	global.X系统信息.Z注册时间 = 局_应用用户信息json.GetInt("RegisterTime")
	global.X系统信息.Y用户类型 = string(局_应用用户信息json.GetStringBytes("UserClassName"))
	global.X系统信息.J积分 = 局_应用用户信息json.GetFloat64("VipNumber")
	global.X系统信息.D登录时间 = 局_应用用户信息json.GetInt("LoginTime")

	局_是否需要更新 := false
	响应信息 := ""
	global.Q快验.Q取最新版本检测(&响应信息, global.X系统信息.B版本号当前, false, &局_是否需要更新, &global.X系统信息.B版本号最新)

	局_基础信息 := ""
	global.Q快验.Q取用户基础信息(&局_基础信息)
	局_基础信息json, _ := fastjson.Parse(局_基础信息)

	global.X系统信息.Y余额 = 局_基础信息json.GetFloat64("RMB")
	global.X系统信息.Y邮箱 = string(局_基础信息json.GetStringBytes("Email"))
	global.X系统信息.D登录IP = string(局_基础信息json.GetStringBytes("LoginIp"))
	global.X系统信息.Qq = string(局_基础信息json.GetStringBytes("Qq"))

	局_在线计数 := 0
	if global.X系统信息.D到期时间 < time.Now().Unix() {
		局_在线计数 = int(Ser_LinkUser.Get取在线总数(true, true))
		global.H缓存.Set("在线数量", 局_在线计数, time.Minute*10)
	}

	response.OkWithDetailed(gin.H{
		"user": global.X系统信息.H会员帐号, "userClassName": global.X系统信息.Y用户类型,
		"vipNumber": global.X系统信息.J积分, "rmb": global.X系统信息.Y余额,
		"vipTime": global.X系统信息.D到期时间, "email": global.X系统信息.Y邮箱,
		"loginTime": global.X系统信息.D登录时间, "loginIp": global.X系统信息.D登录IP,
		"qq": global.X系统信息.Qq, "registerTime": global.X系统信息.Z注册时间,
		"appVer": global.X系统信息.B版本号当前, "appVerUpdateTime": 文件修改日期,
		"appVerNew": global.X系统信息.B版本号最新, "linkTokenCount": 局_在线计数,
		"agentUid": global.GVA_Viper.GetInt("duid"),
	}, "获取成功", c)
	return
}

// GetSmsCaptcha 取短信验证码
func (k *KuaiYan) GetSmsCaptcha(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 响应信息 string
	if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"GetSMSCaptcha"`) {
		if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"GetSMSCaptcha":2`) {
			global.Q快验.Z置验证码信息(2, 请求.CaptCha2.CaptchaId, 请求.CaptCha2.CaptChaValue)
		} else {
			global.Q快验.Z置验证码信息(1, 请求.CaptCha1.CaptchaId, 请求.CaptCha1.CaptChaValue)
		}
	}
	if !global.Q快验.Q取短信验证码(&响应信息, 请求.User, "") {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, _ := fastjson.Parse(响应信息)
	response.OkWithDetailed(gin.H{"captchaId": string(局_json.GetStringBytes("CaptchaId"))}, "发送成功", c)
	return
}

// SetPassword 快验找回密码
func (k *KuaiYan) SetPassword(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if !global.Q快验.M密码找回或修改_绑定手机(请求.User, 请求.PassWord, 请求.CaptCha3.CaptchaId, 请求.CaptCha3.CaptChaValue) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithMessage("设置新密码["+请求.PassWord+"]成功", c)
	return
}

// Register 快验注册
func (k *KuaiYan) Register(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	global.Q快验.Z置验证码信息(3, 请求.CaptCha3.CaptchaId, 请求.CaptCha3.CaptChaValue)
	if !global.Q快验.Y用户注册(请求.User, 请求.PassWord, c.ClientIP(), utils.W文本_取随机字符串(10), 请求.Qq, 请求.Qq+"@qq.com", 请求.User) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithMessage("注册成功", c)
	return
}

// Login 快验登录
func (k *KuaiYan) Login(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 响应信息 string
	if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"UserLogin"`) {
		if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"UserLogin":2`) || 请求.CaptCha2.CaptChaValue != "" {
			global.Q快验.Z置验证码信息(2, 请求.CaptCha2.CaptchaId, 请求.CaptCha2.CaptChaValue)
		} else {
			global.Q快验.Z置验证码信息(1, 请求.CaptCha1.CaptchaId, 请求.CaptCha1.CaptChaValue)
		}
	}
	if !global.Q快验.D登录_通用(&响应信息, 请求.User, 请求.PassWord, c.ClientIP(), "", global.X系统信息.B版本号当前) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, _ := fastjson.Parse(响应信息)
	global.X系统信息.H会员帐号 = 请求.User
	global.X系统信息.H会员密码 = 请求.PassWord
	global.X系统信息.D到期时间 = 局_json.GetInt64("VipTime")
	global.X系统信息.B绑定信息 = string(局_json.GetStringBytes("Key"))
	global.X系统信息.Z注册时间 = 局_json.GetInt("RegisterTime")

	局_是否需要更新 := false
	global.Q快验.Q取最新版本检测(&响应信息, global.X系统信息.B版本号当前, false, &局_是否需要更新, &global.X系统信息.B版本号最新)
	global.Q快验.Q取用户余额(&global.X系统信息.Y余额)
	局_基础信息 := ""
	global.Q快验.Q取用户基础信息(&局_基础信息)
	局_基础信息json, _ := fastjson.Parse(响应信息)

	global.X系统信息.Y用户类型 = string(局_json.GetStringBytes("UserClassName"))
	global.X系统信息.J积分 = 局_json.GetFloat64("VipNumber")
	global.X系统信息.Y邮箱 = string(局_基础信息json.GetStringBytes("Email"))
	global.X系统信息.D登录时间 = 局_json.GetInt("LoginTime")
	global.X系统信息.D登录IP = string(局_基础信息json.GetStringBytes("LoginIp"))
	global.X系统信息.Qq = string(局_基础信息json.GetStringBytes("Qq"))

	response.OkWithDetailed(gin.H{
		"user": global.X系统信息.H会员帐号, "userClassName": global.X系统信息.Y用户类型,
		"vipNumber": global.X系统信息.J积分, "rmb": global.X系统信息.Y余额,
		"vipTime": global.X系统信息.D到期时间, "email": global.X系统信息.Y邮箱,
		"loginTime": global.X系统信息.D登录时间, "loginIp": global.X系统信息.D登录IP,
		"qq": global.X系统信息.Qq, "registerTime": global.X系统信息.Z注册时间,
		"appVer": global.X系统信息.B版本号当前, "appVerNew": global.X系统信息.B版本号最新,
	}, "获取成功", c)
	return
}

// OutLogin 快验注销
func (k *KuaiYan) OutLogin(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Y用户登录注销() {
		response.FailWithMessage("注销失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	global.X系统信息.K开启验证码接口列表 = 响应信息
	response.OkWithMessage("注销成功", c)
	return
}

// GetIsBuyKaList 取可购买充值卡列表
func (k *KuaiYan) GetIsBuyKaList(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Q取可购买卡类列表(&响应信息) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, _ := fastjson.Parse(响应信息)
	局_可买充值卡 := 局_json.GetArray("Data")
	var data = make([]gin.H, len(局_可买充值卡))
	for 索引, _ := range 局_可买充值卡 {
		data[索引] = gin.H{
			"id": 局_可买充值卡[索引].GetInt("Id"), "name": string(局_可买充值卡[索引].GetStringBytes("Name")),
			"money": 局_可买充值卡[索引].GetFloat64("Money"),
		}
	}
	response.OkWithDetailed(data, "获取成功", c)
	return
}

// GetPurchasedKaList 购买充值卡记录
func (k *KuaiYan) GetPurchasedKaList(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Q取已购买卡号列表(&响应信息, 5) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, _ := fastjson.Parse(响应信息)
	局_可买充值卡 := 局_json.GetArray("Data")
	var data = make([]gin.H, len(局_可买充值卡))
	for 索引, _ := range 局_可买充值卡 {
		data[索引] = gin.H{
			"id": 局_可买充值卡[索引].GetInt("Id"), "name": string(局_可买充值卡[索引].GetStringBytes("Name")),
			"money": 局_可买充值卡[索引].GetFloat64("Money"), "num": 局_可买充值卡[索引].GetInt("Num"),
			"numMax": 局_可买充值卡[索引].GetInt("NumMax"), "kaClassId": 局_可买充值卡[索引].GetInt("KaClassId"),
			"kaClassName":  string(局_可买充值卡[索引].GetStringBytes("KaClassName")),
			"registerTime": 局_可买充值卡[索引].GetInt("RegisterTime"),
		}
	}
	response.OkWithDetailed(data, "获取成功", c)
	return
}

// GetPayStatus 取支付通道状态
func (k *KuaiYan) GetPayStatus(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Q余额充值_取支付通道状态(&响应信息) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithData(响应信息, c)
	return
}

// PayMoneyToKa 余额购买充值卡
func (k *KuaiYan) PayMoneyToKa(c *gin.Context) {
	var 请求 结构请求_KaCLassId
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 响应信息 string
	if !global.Q快验.Y余额购买充值卡(&响应信息, 请求.KaCLassId) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, _ := fastjson.Parse(响应信息)
	response.OkWithDetailed(gin.H{"kaName": string(局_json.GetStringBytes("KaName"))}, "购买成功", c)
	return
}

// UseKa 卡号充值
func (k *KuaiYan) UseKa(c *gin.Context) {
	var 请求 结构请求_充值
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if !global.Q快验.K卡号充值(global.X系统信息.H会员帐号, 请求.Ka, 请求.InviteUser) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithMessage("充值成功", c)
	return
}

// GetPayPC 余额充值
func (k *KuaiYan) GetPayPC(c *gin.Context) {
	var 请求 结构请求_余额充值
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 响应信息 string
	var 订单ID = 请求.D订单ID
	if 订单ID != "" {
		if !global.Q快验.D订单_取状态(&响应信息, 订单ID) {
			局_错误代号 := 0
			response.FailWithMessage(global.Q快验.Q取错误信息(&局_错误代号), c)
		} else {
			parse, _ := fastjson.Parse(响应信息)
			response.OkWithData(gin.H{"status": parse.GetInt("Status")}, c)
		}
		return
	}
	if !global.Q快验.D订单_购买余额(&响应信息, 请求.Type, global.X系统信息.H会员帐号, 请求.C充值金额, &订单ID) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	var 局_MAP gin.H
	_ = json.Unmarshal([]byte(响应信息), &局_MAP)
	局_MAP["status"] = 1
	response.OkWithData(局_MAP, c)
	return
}

// Updater 更新程序
func (k *KuaiYan) Updater(c *gin.Context) {
	if KuaiYanUpdater.J_系统更新状态 > 0 {
		response.FailWithMessage("已经在更新中当前进度:"+KuaiYanUpdater.J_系统更新提示, c)
		if KuaiYanUpdater.J_系统更新状态 == 2 {
			KuaiYanUpdater.J_系统更新提示 = ""
			KuaiYanUpdater.J_系统更新状态 = 0
		}
		return
	}
	var 响应信息 string
	if !global.Q快验.Q取新版本下载地址(&响应信息) {
		response.FailWithMessage("获取新版本下载地址失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	if !utils2.X系统_权限检测() {
		response.FailWithMessage("系统无写入文件权限,请检查权限是否777或755,运行用户是否为root", c)
		return
	}
	KuaiYanUpdater.J_系统更新提示 = "准备中"
	go KuaiYanUpdater.K快验系统开始更新(响应信息, func(执行程序本地路径 string) {
		fmt.Printf("系统更新%v,马上启动新版本程序继承端口,退出旧版本\n\n", 执行程序本地路径)
		global.GVA_LOG.Info("系统更新,马上启动新版本程序继承端口,退出旧版本\n")
		err := os.Chmod(执行程序本地路径, 0755)
		if err != nil {
			KuaiYanUpdater.J_系统更新状态 = 2
			KuaiYanUpdater.J_系统更新提示 = fmt.Sprintf("启动修改下载文件权限失败:%v\n", err.Error())
			return
		}
		cmd := exec.Command(执行程序本地路径)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err = cmd.Start(); err != nil {
			KuaiYanUpdater.J_系统更新状态 = 2
			KuaiYanUpdater.J_系统更新提示 = fmt.Sprintf("启动新更新后程序失败:%v\n", err.Error())
			return
		}
		err = global.GVA_Gin.Shutdown(context.Background())
		if err != nil {
			KuaiYanUpdater.J_系统更新状态 = 2
			KuaiYanUpdater.J_系统更新提示 = "下载文件成功,重启失败,请手动重启系统:" + err.Error()
			return
		}
		os.Exit(0)
	})
	var filename string
	filename = path.Base(os.Args[0])
	response.OkWithMessage("开始更新,当前:"+filename, c)
	return
}

// K快验心跳 内部函数 - 保留原有心跳逻辑
var 心跳容错计数 = 0

func K快验心跳() {
	if global.Q快验.J_Token == "" {
		return
	}
	var 响应信息 string
	var 当前状态 int
	if !global.Q快验.X心跳(&响应信息, &当前状态) {
		心跳容错计数++
		if 心跳容错计数 >= 3 {
			global.Q快验.J_Token = ""
			global.X系统信息.H会员帐号 = ""
			global.X系统信息.D到期时间 = 0
			global.X系统信息.Y用户类型代号 = 0
		}
		return
	}
	心跳容错计数 = 0

	局_设备信息, err := getServerInfoForKuaiYan()
	if err != nil {
		return
	}
	局_动态标记 := fmt.Sprintf("%s %dH%.2fG %dG %d协程,用户数:%d,卡总数:%d,在线数:%d",
		utils.S三元(global.Q快验.J集_连接方式 == 0, "直连", "网关"),
		runtime.NumCPU(),
		utils.Float64除int64(utils.Int64到Float64(int64(局_设备信息.Ram.TotalMB)), 1024, 2),
		局_设备信息.Disk.TotalGB,
		runtime.NumGoroutine(),
		Ser_User.Q取总数(),
		Ser_Ka.Q取总数(),
		Ser_LinkUser.Get取在线总数(true, true),
	)
	if 局_设备信息.Os.GOOS != "linux" {
		局_动态标记 += " " + 局_设备信息.Os.GOOS
	}
	global.Q快验.Z置动态标记(局_动态标记)
}

func getServerInfoForKuaiYan() (*utils2.Server, error) {
	return getServerInfo()
}
