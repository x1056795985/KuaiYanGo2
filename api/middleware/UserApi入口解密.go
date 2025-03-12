package middleware

import (
	"EFunc/utils"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"net/http"
	"server/Service/Captcha"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Js"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/api/UserApi"
	"server/api/UserApi/response"
	"server/global"
	"server/new/app/logic/common/blacklist"
	"server/new/app/logic/common/setting"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 键名不能有长度正好32的名称, 否则可能会和md5(用户api) 冲突隐患
var J集_UserAPi路由 = map[string]路由信息{
	//"GetToken": UserApi.UserApi_GetToken,   //通过中间件单独处理了,不放在路由内,防止重复调用
	"NewUserInfo":            {"用户注册", UserApi.UserApi_用户注册, true},
	"UserLogin":              {"用户登录", UserApi.UserApi_用户登录, true},
	"UseKa":                  {"卡号充值", UserApi.UserApi_卡号充值, true},
	"UserReduceMoney":        {"用户减少余额", UserApi.UserApi_用户减少余额, true},
	"UserReduceVipNumber":    {"用户减少积分", UserApi.UserApi_用户减少积分, true},
	"UserReduceVipTime":      {"用户减少点数", UserApi.UserApi_用户减少点数, true},
	"IsServerLink":           {"取服务器连接状态", UserApi.UserApi_取服务器连接状态, true},
	"IsLogin":                {"取登录状态", UserApi.UserApi_取登录状态, true},
	"GetVipData":             {"取Vip数据", UserApi.UserApi_取Vip数据, true},
	"GetAppGongGao":          {"取应用公告", UserApi.UserApi_取应用公告, true},
	"GetAppUpDataJson":       {"取新版本下载地址", UserApi.UserApi_取新版本下载地址, true},
	"GetAppPublicData":       {"取应用专属变量", UserApi.UserApi_取应用专属变量, true},
	"GetPublicData":          {"取公共变量", UserApi.UserApi_取公共变量, true},
	"SetPublicData":          {"取公共变量", UserApi.UserApi_置公共变量, true},
	"GetAgentConfig":         {"取代理云配置", UserApi.UserApi_取代理云配置, true},
	"GetAppVersion":          {"取应用最新版本", UserApi.UserApi_取应用最新版本, true},
	"GetAppHomeUrl":          {"取应用主页Url", UserApi.UserApi_取应用主页Url, true},
	"SetAppUserKey":          {"置新绑定信息", UserApi.UserApi_置新绑定信息, true},
	"DeleteAppUserKey":       {"解除绑定信息", UserApi.UserApi_解除绑定信息, true},
	"SetNewUserMsg":          {"置新用户消息", UserApi.UserApi_置新用户消息, true},
	"GetCaptcha":             {"取验证码信息", UserApi.UserApi_取验证码信息, true},
	"GetSMSCaptcha":          {"取短信验证码信息", UserApi.UserApi_取短信验证码信息, true},
	"GetAppUserKey":          {"取用户绑定信息", UserApi.UserApi_取用户绑定信息, true},
	"GetIsUser":              {"取用户是否存在", UserApi.UserApi_取用户是否存在, true},
	"GetAppUserInfo":         {"取软件用户信息", UserApi.UserApi_取软件用户信息, true},
	"GetAppInfo":             {"取应用基础信息", UserApi.UserApi_取应用基础信息, true},
	"GetUserInfo":            {"取用户基础信息", UserApi.UserApi_取用户基础信息, true},
	"SetUserQqEmailPhone":    {"置用户基础信息", UserApi.UserApi_置用户基础信息, true},
	"GetUserIP":              {"取用户IP", UserApi.UserApi_GetUserIP, true},
	"GetSystemTime":          {"取系统时间戳", UserApi.UserApi_取系统时间戳, true},
	"GetAppUserVipTime":      {"取Vip到期时间戳", UserApi.UserApi_取Vip到期时间戳, true},
	"GetAppUserNote":         {"取软件用户备注", UserApi.UserApi_取软件用户备注, true},
	"LogOut":                 {"用户登录注销", UserApi.UserApi_用户登录注销, true},
	"RemoteLogOut":           {"用户登录远程注销", UserApi.UserApi_用户登录远程注销, true},
	"HeartBeat":              {"心跳", UserApi.UserApi_心跳, true},
	"OldPassWordSetPassWord": {"密码找回或修改_旧密码", UserApi.UserApi_密码找回或修改_验证旧密码, true},
	"SetPassWord":            {"密码找回或修改_超级密码", UserApi.UserApi_密码找回或修改_超级密码, true},
	"SmsCodeSetPassWord":     {"密码找回或修改_密保手机", UserApi.UserApi_密码找回或修改_密保手机, true},
	"GetUserRmb":             {"取用户余额", UserApi.UserApi_取用户余额, true},
	"GetAppUserVipNumber":    {"取用户积分", UserApi.UserApi_取用户积分, true},
	"GetCaptchaApiList":      {"取开启验证码接口", UserApi.UserApi_取开启验证码接口, true},

	"GetTab":            {"取动态标签", UserApi.UserApi_取动态标签, true},
	"GetRegisterGiveKa": {"取注册送卡", UserApi.UserApi_取注册送卡, true},
	"SetTab":            {"置动态标签", UserApi.UserApi_置动态标签, true},
	"GetPayOrderStatus": {"订单_取状态", UserApi.UserApi_订单_取状态, true},
	"PayKaUsa":          {"订单_购卡直冲", UserApi.UserApi_订单_购卡直冲, true},
	"PayUserMoney":      {"订单_余额充值", UserApi.UserApi_订单_余额充值, true},
	"PayUserVipNumber":  {"订单_积分充值", UserApi.UserApi_订单_积分充值, true},
	"PayGetKa":          {"订单_支付购卡", UserApi.UserApi_订单_支付购卡, true},
	//下方接口为历史遗留20240429 已移除
	//"GetAliPayPC":         {"订单_余额充值_支付宝PC支付", UserApi.UserApi_订单_余额充值_支付宝PC支付, true},
	//"GetWXPayPC":          {"订单_余额充值_微信支付支付", UserApi.UserApi_订单_余额充值_微信支付支付, true},
	"GetPayStatus":          {"取支付通道状态", UserApi.UserApi_取支付通道状态, true},
	"GetPayKaList":          {"取可购买卡类列表", UserApi.UserApi_取可购买卡类列表, true},
	"GetPurchasedKaList":    {"取已购买充值卡列表", UserApi.UserApi_取已购买充值卡列表, true},
	"PayMoneyToVipNumber":   {"余额购买积分", UserApi.UserApi_余额购买积分, true},
	"PayMoneyToKa":          {"余额购买充值卡", UserApi.UserApi_余额购买充值卡, true},
	"GetUserClassList":      {"取用户类型列表", UserApi.UserApi_取用户类型列表, true},
	"SetUserClass":          {"置用户类型", UserApi.UserApi_置用户类型, true},
	"RunJS":                 {"云函数执行", UserApi.UserApi_云函数执行, true},
	"TaskPoolNewData":       {"任务池_任务创建", UserApi.UserApi_任务池_任务创建, true},
	"TaskPoolGetData":       {"任务池_任务查询", UserApi.UserApi_任务池_任务查询, true},
	"TaskPoolGetDataList":   {"任务池_取任务列表", UserApi.UserApi_任务池_取任务列表, true},
	"TaskPoolGetTask":       {"任务池_任务处理获取", UserApi.UserApi_任务池_任务处理获取, true},
	"TaskPoolSetTask":       {"任务池_任务处理返回", UserApi.UserApi_任务池_任务处理返回, true},
	"TaskPoolGetTypeStatus": {"任务池_取类型状态", UserApi.UserApi_任务池_取类型状态, true},
	"GetUserConfig":         {"取用户云配置", UserApi.UserApi_取用户云配置, true},
	"SetUserConfig":         {"置用户云配置", UserApi.UserApi_置用户云配置, true},
	"SetAgentUid":           {"置代理标识", UserApi.UserApi_置代理标志, true},
	"GetKaInfo":             {"取卡号详情", UserApi.UserApi_取卡号详情, true},
	"GetJwtToken":           {"取jwt令牌", UserApi.UserApi_取jwtToken, true},
	"GetUploadToken":        {"云存储_取文件上传授权", UserApi.UserApi_云存储_取文件上传授权, true},

	//快验Api
	"KyApiSendSms":           {"快验发送验证码短信", UserApi.KyApiSendSms, false},
	"KyApiJiYanVerifyTicket": {"快验_极验验证码结果验证", UserApi.K快验_极验验证码结果验证, false},
}

type 路由信息 struct {
	Z中文名  string
	Z指向函数 func(*gin.Context)
	X显示   bool //是否显示到前段
}

var 集_UserAPi路由强制RSA = map[string]int{
	"GetToken":            1,
	"UserLogin":           1,
	"UserReduceMoney":     1,
	"UserReduceVipNumber": 1,
	"UserReduceVipTime":   1,
	"GetVipData":          1,
}

func init() {
	fmt.Sprintln("集_UserAPi路由被初始化了")
}

var J集_UserAPi路由_加密 加密路由信息

type 加密路由信息 struct {
	L路由md5 string //每次更新加密路由缓存, 都更新这个索引,每次读取路由,都检测索引是否和缓存相同,如果不同,更新索引
	J加密路由  map[string]string
	D读写锁   sync.RWMutex
}

func (this *加密路由信息) G更新md5APi名称(盐值 string) {

	if 盐值 == "" {
		//无加密 清空加密路由表
		this.J加密路由 = make(map[string]string, 0)
		return
	}
	if !this.D读写锁.TryLock() { // 尝试进入,如果成功继续,如果失败,说明有其他线程在更新,直接返回即可
		return
	}
	defer this.D读写锁.Unlock()

	//更新加密路由表
	this.J加密路由 = make(map[string]string, len(J集_UserAPi路由)+1)
	局_临时文本 := utils2.Md5String("GetToken" + 盐值)
	this.J加密路由[局_临时文本] = "GetToken"

	fmt.Printf("API名称加密已更新:%s => %s\n", this.J加密路由[局_临时文本], 局_临时文本)
	局_路由md5原值 := ""
	for 局_用户api := range J集_UserAPi路由 {
		局_哈希后的值 := utils2.Md5String(局_用户api + 盐值)
		this.J加密路由[局_哈希后的值] = 局_用户api
		fmt.Printf("API名称加密已更新:%s => %s\n", 局_用户api, 局_哈希后的值)
		局_路由md5原值 = 局_路由md5原值 + 局_哈希后的值
	}

	this.L路由md5 = utils2.Md5String(局_路由md5原值)
	global.H缓存.Set("J集_UserAPi加密路由md5", this.L路由md5, -1) //写到redis

}

func (this *加密路由信息) Q取md5APi名称(md5值 string) (string, bool) {
	if len(md5值) != 32 {
		return "", false
	}
	//大部分都是直接有结果的,直接返回即可
	局_用户api, ok := this.J加密路由[md5值]
	if ok {
		return 局_用户api, ok
	}
	// 失败了
	//检测是否和redis 加密数据md5一样,如果不相同,说明有其他端更新了,本端没更新
	云_L路由md5, ok := global.H缓存.Get("J集_UserAPi加密路由md5")
	if !ok || 云_L路由md5 != this.L路由md5 {
		this.G更新md5APi名称(setting.Q系统设置().Y用户API加密盐)
	}

	//更新完后在返回
	局_用户api, ok = this.J加密路由[md5值]
	return 局_用户api, ok

}

// UserApi解密 Token有效的才放行,否则返回Token失效
func UserApi解密() gin.HandlerFunc {
	return func(c *gin.Context) {

		if !setting.Q系统设置().X系统开关 {
			//什么都不返回,直接关闭
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_系统已关闭, setting.Q系统设置().X系统关闭提示})
			c.Abort()
			return
		}

		Token := c.Request.Header.Get("Token")
		if Token == "" {
			c.Next()
			return
		}

		//还是得先获取App信息,因为token不存在响应加密也需要
		Appid, _ := strconv.Atoi(c.DefaultQuery("AppId", ""))
		if Appid < 10000 || Ser_AppInfo.AppId是否存在(Appid) == false {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_App不存在, "App不存在"})
			c.Abort()
			return
		}

		if blacklist.Is黑名单(c.ClientIP(), Appid) {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_黑名单信息, "黑名单ip"})
			c.Abort()
			return
		}

		AppInfo := Ser_AppInfo.App取App详情(Appid)
		c.Set("AppInfo", AppInfo) //必须先置入 防止响应信息时加密失败

		if AppInfo.Status == 1 {
			response.X响应状态消息(c, response.Status_已停止运营, AppInfo.AppStatusMessage)
			c.Abort()
			return
		}

		局_在线信息, err := Ser_LinkUser.Token取User在线详情(Token)
		//防止越权 appid1的令牌操作 appid2的功能
		if err != nil || 局_在线信息.LoginAppid != Appid {
			response.X响应状态(c, response.Status_Token无效)
			c.Abort()
			return
		}
		if 局_在线信息.Status != 1 {
			response.X响应状态带数据(c, response.Status_Token已注销, gin.H{"LogoutCode": 局_在线信息.LogoutCode})
			c.Abort()
			return
		}

		go Ser_LinkUser.Token更新最后活动时间(Token)
		if 局_在线信息.Ip != c.ClientIP() {
			go Ser_LinkUser.Token更新在线ip(Token, c.ClientIP())
			//go Ser_Log.Log_写登录日志(局_在线信息.User, c.ClientIP(), "在线用户更换ip"+局_在线信息.Ip+"->"+c.ClientIP(), 局_在线信息.LoginAppid)
		}

		c.Set("局_在线信息", 局_在线信息)

		//密文解密成明文
		var 局_json明文 string
		var 结构加密包 请求响应_加密包
		var 局_临时字节集 []byte
		if AppInfo.CryptoType == 2 {
			c.Set("局_CryptoKeyAes", AppInfo.CryptoKeyAes)
		}

		if AppInfo.CryptoType == 2 || AppInfo.CryptoType == 3 {
			err = c.ShouldBindJSON(&结构加密包)
			if err != nil {
				response.X响应状态(c, response.Status_参数错误)
				c.Abort()
				return
			}
			局_临时字节集, _ = base64.StdEncoding.DecodeString(结构加密包.A密文)
		} else {
			局_临时字节集, _ = c.GetRawData()
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(局_临时字节集)) // 关键点 //通过这写回post数据,就可以多次读取了
		}
		//先检查签名  签名争取,设置AES密匙  明文的是不用AES密匙的
		var 局_临时AES密匙 []byte
		if len(结构加密包.B签名) == 32 {
			//签名都转大写防止误判
			if strings.ToUpper(结构加密包.B签名) == strings.ToUpper(utils2.Md5String(结构加密包.A密文+局_在线信息.CryptoKeyAes)) {
				局_临时AES密匙 = []byte(局_在线信息.CryptoKeyAes)
			} else {
				fmt.Printf("结构加密包.B签名校验失败%v!=%v", 结构加密包.B签名, strings.ToUpper(utils2.Md5String(结构加密包.A密文+局_在线信息.CryptoKeyAes)))
			}
		}

		if len(结构加密包.B签名) > 32 {
			局_签名解密, _ := base64.StdEncoding.DecodeString(结构加密包.B签名)
			局_临时AES密匙 = utils2.Rsa私钥解密2([]byte(AppInfo.CryptoKeyPrivate), 局_签名解密)
		}

		if len(局_临时AES密匙) == 0 && AppInfo.CryptoType != 1 { //只有明文才不检查
			go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, 局_在线信息.User, c.ClientIP(), "用户发送错误签名封包,可能在尝试破解")
			response.X响应状态(c, response.Status_签名错误)
			c.Abort()
			return
		}

		//===========有token解密明文======================
		if AppInfo.CryptoType == 3 || AppInfo.CryptoType == 2 {
			局_json明文 = utils2.Aes解密_cbc192字节集(局_临时字节集, 局_临时AES密匙)
		} else if AppInfo.CryptoType == 1 {
			局_json明文 = string(局_临时字节集)
		}

		if 局_json明文 == "" {
			go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, 局_在线信息.User, c.ClientIP(), "用户发送签名正确密文解密错误封包,可能在尝试破解")
			response.X响应状态(c, response.Status_加解密失败)
			c.Abort()
			return
		}

		局_fastjson, err := fastjson.Parse(局_json明文)
		if err != nil {
			response.X响应状态(c, response.Status_参数错误)
			c.Abort()
			return
		}

		c.Set("局_CryptoKeyAes", 局_在线信息.CryptoKeyAes) //不管用不用到都放里
		局_Time := 局_fastjson.GetInt("Time")
		if AppInfo.PackTimeOut != 0 && int(time.Now().Unix())-局_Time > AppInfo.PackTimeOut {
			response.X响应状态(c, response.Status_封包超时)
			c.Abort()
			return
		}

		局_成功Status := 局_fastjson.GetInt("Status")
		if 局_成功Status < 10000 {
			response.X响应状态(c, response.Status_状态码错误)
			c.Abort()
			return
		}
		局_Api := strings.TrimSpace(string(局_fastjson.GetStringBytes("Api")))

		ok := false
		// 如果有加密后的API,就会赋值原始APi到变量,如果失败,不会改变
		if len(J集_UserAPi路由_加密.J加密路由) > 0 { //如果>0说明启用Api加密了,
			if 局_Api, ok = J集_UserAPi路由_加密.Q取md5APi名称(局_Api); !ok {
				response.X响应状态消息(c, response.Status_Api不存在, "API名称加密错误")
				c.Abort()
				return
			}
		}

		c.Set("局_Api", 局_Api) //ApiHook后可能会用到
		//fmt.Printf("用户发送数据明文:%v", 局_json明文)
		//==========================ApiHook之前====================================
		err = apiHook之前(c, &AppInfo, &局_在线信息, 局_Api, &局_json明文)
		if err != nil {
			return
		}
		//=============================ApiHook结束=================================

		//==========================判断验证码===============================================
		if utils.W文本_是否包含关键字(AppInfo.Captcha, `"`+局_Api+`"`) { //先判断Api是否需要验证码

			//AppInfo.Captcha内容 {"UserReduceMoney":1,"UserReduceVipNumber":1,"UserLogin":1}
			//{"Api":"GetCaptcha","CaptchaType":2,"Time":1683629194,"Status":18518,"Captcha":{"Type":1,"Id":"123456789","Value":"8888"}}
			局_验证码类型 := 局_fastjson.Get("Captcha").GetInt("Type")
			局_验证码ID := string(局_fastjson.Get("Captcha").GetStringBytes("Id"))
			局_验证码内容 := string(局_fastjson.Get("Captcha").GetStringBytes("Value"))

			if 局_验证码类型 == 2 && (setting.Q行为验证码平台配置().J极验行为验证4.Y验证_ID == "" || setting.Q行为验证码平台配置().J极验行为验证4.Y验证_KEY == "") {
				response.X响应状态消息(c, response.Status_验证码错误, "系统未设置行为验证码Id或Key,系统设置->行为验证码平台配置")
				c.Abort()
				return
			}

			if 局_验证码类型 == 1 && utils.W文本_是否包含关键字(AppInfo.Captcha, `"`+局_Api+`":1`) && Captcha.H缓存验证码校验实例.Verify(局_验证码ID, 局_验证码内容, true) {
				//提交的验证码类型为1 英数   设置的也为1 英数, 验证没问题
				goto 验证码正确
			}
			if 局_验证码类型 == 2 && utils.W文本_是否包含关键字(AppInfo.Captcha, `"`+局_Api+`":2`) && Captcha.CaptCha_行为验证码验证(局_验证码ID, 局_验证码内容) == nil {
				//提交的验证码类型为2 行为验证   设置的也为1 英数, 验证没问题
				goto 验证码正确
			}
			if 局_验证码类型 == 3 && utils.W文本_是否包含关键字(AppInfo.Captcha, `"`+局_Api+`":3`) && Captcha.H缓存验证码校验实例.Verify(局_验证码ID, 局_验证码内容, false) {
				//提交的验证码类型为3 英数   设置的也为1 英数, 验证没问题
				goto 验证码正确
			}

			response.X响应状态(c, response.Status_验证码错误)
			c.Abort()
			return
		验证码正确:
		}
		//==========================验证码结束===============================================

		if 集_UserAPi路由强制RSA[局_Api] == 1 && AppInfo.CryptoType == 3 {
			//如果API是强制AES通讯的, 但是使用AES加密发过来的,估计是破解者测试包
			if len(结构加密包.B签名) == 32 {
				go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, 局_在线信息.User, c.ClientIP(), fmt.Sprintf("强制RSA封包Api,用户使用了AES加密方式可能非法用户在尝试破解,并已Hook到Aes密钥"))
				response.X响应状态(c, response.Status_加解密失败)
				c.Abort()
				return
			}
		}

		var 局_路由信息 路由信息
		局_路由信息, ok = J集_UserAPi路由[局_Api]

		c.Set("RSA强制", 集_UserAPi路由强制RSA[局_Api] == 1)

		if ok { //如果有这个api 就跳转执行, 如果没有就最终走向 返回无Api的函数

			c.Set("局_json明文", 局_json明文)
			c.Set("局_成功Status", 局_成功Status)
			局_路由信息.Z指向函数(c)
			c.Abort()
			return
		}

		c.Next()
		return
	}
}

type 请求响应_加密包 struct {
	A密文 string `json:"a"`
	B签名 string `json:"b"`
}

// 回复json结构体
type 请求响应_X响应状态 struct {
	Time   int64  `json:"Time"`
	Status int    `json:"Status"`
	Msg    string `json:"Msg"`
}

func UserApi无Token解密() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !setting.Q系统设置().X系统开关 {
			//什么都不返回,直接关闭
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_系统已关闭, setting.Q系统设置().X系统关闭提示})
			c.Abort()
			return
		}
		Token := c.Request.Header.Get("Token")
		if Token != "" {
			c.Next()
			return
		}

		Appid, _ := strconv.Atoi(c.DefaultQuery("AppId", ""))
		if Appid < 10000 || Ser_AppInfo.AppId是否存在(Appid) == false {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_App不存在, "App不存在"})
			c.Abort()
			return
		}
		if blacklist.Is黑名单(c.ClientIP(), Appid) {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_黑名单信息, "黑名单ip"})
			c.Abort()
			return
		}

		AppInfo := Ser_AppInfo.App取App详情(Appid)

		c.Set("AppInfo", AppInfo)
		//密文解密成明文
		var 局_json明文 string
		var 结构加密包 请求响应_加密包
		var 局_临时字节集 []byte
		if AppInfo.CryptoType == 2 || AppInfo.CryptoType == 3 {
			err := c.ShouldBindJSON(&结构加密包)
			if err != nil {
				response.X响应状态(c, response.Status_参数错误)
				c.Abort()
				return
			}
			局_临时字节集, _ = base64.StdEncoding.DecodeString(结构加密包.A密文)
		} else {
			局_临时字节集, _ = c.GetRawData()                                 //GetRawData只能使用一次 且会使反序列化无效
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(局_临时字节集)) // 关键点 //通过这写回post数据,就可以多次读取了
		}

		//===========无token解密明文======================
		if AppInfo.CryptoType == 3 {

			//强制这个接口必须走RSA方式
			局_签名解密, _ := base64.StdEncoding.DecodeString(结构加密包.B签名)

			局_临时AES密匙 := utils2.Rsa私钥解密2([]byte(AppInfo.CryptoKeyPrivate), 局_签名解密)

			局_json明文 = utils2.Aes解密_cbc192字节集(局_临时字节集, 局_临时AES密匙)

		} else if AppInfo.CryptoType == 2 {

			局_json明文 = utils2.Aes解密_cbc192(局_临时字节集, AppInfo.CryptoKeyAes)
		} else if AppInfo.CryptoType == 1 {

			局_json明文 = string(局_临时字节集)
		}
		c.Set("RSA强制", true) //必须放在 响应消息前面 否则getToken不能强制RSA加密 返回b值为md5 本地验签会错误

		if 局_json明文 == "" {
			response.X响应状态(c, response.Status_加解密失败)
			c.Abort()
			return
		}

		局_fastjson, err := fastjson.Parse(局_json明文)
		if err != nil {
			response.X响应状态(c, response.Status_参数错误)
			c.Abort()
			return
		}

		局_Api := strings.TrimSpace(string(局_fastjson.GetStringBytes("Api")))
		ok := false
		// 如果有加密后的API,就会赋值原始APi到变量,如果失败,不会改变
		if len(J集_UserAPi路由_加密.J加密路由) > 0 { //如果>0说明启用Api加密了,
			if 局_Api, ok = J集_UserAPi路由_加密.Q取md5APi名称(局_Api); !ok {
				response.X响应状态消息(c, response.Status_Api不存在, "API名称加密错误")
				c.Abort()
				return
			}
		}
		c.Set("局_Api", 局_Api) //ApiHook后可能会用到
		if 局_Api != "GetToken" {
			response.X响应状态(c, response.Status_Token无效)
			c.Abort()
			return
		}

		if string(局_fastjson.GetStringBytes("Key")) != "" {
			c.Set("局_CryptoKeyAes", utils.W文本_取随机字符串(24)) //随机生产AES密钥
		} else {
			c.Set("局_CryptoKeyAes", AppInfo.CryptoKeyAes)
		}

		局_Time := 局_fastjson.GetInt("Time")
		if AppInfo.PackTimeOut != 0 && int(time.Now().Unix())-局_Time > AppInfo.PackTimeOut {
			response.X响应状态(c, response.Status_封包超时)
			c.Abort()
			return
		}

		局_成功Status := 局_fastjson.GetInt("Status")
		if 局_成功Status < 10000 {
			response.X响应状态(c, response.Status_状态码错误)
			c.Abort()
			return
		}

		c.Set("局_json明文", 局_json明文)
		c.Set("局_成功Status", 局_成功Status)

		UserApi.UserApi_GetToken(c)
		c.Abort()
		return
	}
}
func UserApi检查数据库连接() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GVA_DB == nil {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), response.Status_SQl错误, "服务器还未连接数据库,暂不可用,请管理员检查原因,或重启系统"})
			c.Abort()
		} else {
			c.Next()
		}
		return
	}
}
func apiHook之前(c *gin.Context, AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken, Api string, json明文 *string) error {
	//==========================ApiHook之前====================================
	if utils.W文本_是否包含关键字(AppInfo.ApiHook, `"`+Api+`"`) { //先判断Api是否需要Hook
		//{"UserLogin":{"Before":"hook登录前","After":"hook登录后"}}
		局_hookBefore := utils.W文本_取出中间文本(AppInfo.ApiHook, `"`+Api+`":{"Before":"`, `"`)
		if 局_hookBefore == "" { //判断是否执行前hook
			return nil
		}

		局_json明文, err := Ser_Js.JS引擎初始化_ApiHook处理(AppInfo, 在线信息, 局_hookBefore, *json明文, c)
		*json明文 = 局_json明文
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			c.Abort()
			return err
		}
	}
	return nil
	//=============================ApiHook结束=================================
}
