package userSafetyApi

import (
	"fmt"
	"server/app/global"
	"sync"

	"github.com/gin-gonic/gin"
	controller "server/app/controller/userSafetyApi"
	"server/app/logic/common/setting"
	utils2 "server/app/utils"
)

// 键名不能有长度正好32的名称, 否则可能会和md5(用户api) 冲突隐患
var J集_UserAPi路由 = map[string]路由信息{
	//"GetToken": UserApi_GetToken,   //通过中间件单独处理了,不放在路由内,防止重复调用
	"NewUserInfo":            {"用户注册", controller.UserApi_用户注册, true},
	"UserLogin":              {"用户登录", controller.UserApi_用户登录, true},
	"UseKa":                  {"卡号充值", controller.UserApi_卡号充值, true},
	"UserReduceMoney":        {"用户减少余额", controller.UserApi_用户减少余额, true},
	"UserReduceVipNumber":    {"用户减少积分", controller.UserApi_用户减少积分, true},
	"UserReduceVipTime":      {"用户减少点数", controller.UserApi_用户减少点数, true},
	"IsServerLink":           {"取服务器连接状态", controller.UserApi_取服务器连接状态, true},
	"IsLogin":                {"取登录状态", controller.UserApi_取登录状态, true},
	"GetVipData":             {"取Vip数据", controller.UserApi_取Vip数据, true},
	"GetAppGongGao":          {"取应用公告", controller.UserApi_取应用公告, true},
	"GetAppUpDataJson":       {"取新版本下载地址", controller.UserApi_取新版本下载地址, true},
	"GetAppPublicData":       {"取应用专属变量", controller.UserApi_取应用专属变量, true},
	"GetPublicData":          {"取公共变量", controller.UserApi_取公共变量, true},
	"SetPublicData":          {"置公共变量", controller.UserApi_置公共变量, true},
	"GetAgentConfig":         {"取代理云配置", controller.UserApi_取代理云配置, true},
	"GetAppVersion":          {"取应用最新版本", controller.UserApi_取应用最新版本, true},
	"GetAppHomeUrl":          {"取应用主页Url", controller.UserApi_取应用主页Url, true},
	"SetAppUserKey":          {"置新绑定信息", controller.UserApi_置新绑定信息, true},
	"DeleteAppUserKey":       {"解除绑定信息", controller.UserApi_解除绑定信息, true},
	"SetNewUserMsg":          {"置新用户消息", controller.UserApi_置新用户消息, true},
	"GetCaptcha":             {"取验证码信息", controller.UserApi_取验证码信息, true},
	"GetSMSCaptcha":          {"取短信验证码信息", controller.UserApi_取短信验证码信息, true},
	"GetAppUserKey":          {"取用户绑定信息", controller.UserApi_取用户绑定信息, true},
	"GetIsUser":              {"取用户是否存在", controller.UserApi_取用户是否存在, true},
	"GetAppUserInfo":         {"取软件用户信息", controller.UserApi_取软件用户信息, true},
	"GetAppInfo":             {"取应用基础信息", controller.UserApi_取应用基础信息, true},
	"GetUserInfo":            {"取用户基础信息", controller.UserApi_取用户基础信息, true},
	"SetUserQqEmailPhone":    {"置用户基础信息", controller.UserApi_置用户基础信息, true},
	"GetUserIP":              {"取用户IP", controller.UserApi_GetUserIP, true},
	"GetSystemTime":          {"取系统时间戳", controller.UserApi_取系统时间戳, true},
	"GetAppUserVipTime":      {"取Vip到期时间戳", controller.UserApi_取Vip到期时间戳, true},
	"GetAppUserNote":         {"取软件用户备注", controller.UserApi_取软件用户备注, true},
	"LogOut":                 {"用户登录注销", controller.UserApi_用户登录注销, true},
	"RemoteLogOut":           {"用户登录远程注销", controller.UserApi_用户登录远程注销, true},
	"HeartBeat":              {"心跳", controller.UserApi_心跳, true},
	"OldPassWordSetPassWord": {"密码找回或修改_旧密码", controller.UserApi_密码找回或修改_验证旧密码, true},
	"SetPassWord":            {"密码找回或修改_超级密码", controller.UserApi_密码找回或修改_超级密码, true},
	"SmsCodeSetPassWord":     {"密码找回或修改_密保手机", controller.UserApi_密码找回或修改_密保手机, true},
	"GetUserRmb":             {"取用户余额", controller.UserApi_取用户余额, true},
	"GetAppUserVipNumber":    {"取用户积分", controller.UserApi_取用户积分, true},
	"GetCaptchaApiList":      {"取开启验证码接口", controller.UserApi_取开启验证码接口, true},

	"GetTab":            {"取动态标签", controller.UserApi_取动态标签, true},
	"GetRegisterGiveKa": {"取注册送卡", controller.UserApi_取注册送卡, true},
	"SetTab":            {"置动态标签", controller.UserApi_置动态标签, true},
	"GetPayOrderStatus": {"订单_取状态", controller.UserApi_订单_取状态, true},
	"PayKaUsa":          {"订单_购卡直冲", controller.UserApi_订单_购卡直冲, true},
	"PayUserMoney":      {"订单_余额充值", controller.UserApi_订单_余额充值, true},

	"PayGetKa":           {"订单_支付购卡", controller.UserApi_订单_支付购卡, true},
	"GetPayStatus":       {"取支付通道状态", controller.UserApi_取支付通道状态, true},
	"GetPayKaList":       {"取可购买卡类列表", controller.UserApi_取可购买卡类列表, true},
	"GetPurchasedKaList": {"取已购买充值卡列表", controller.UserApi_取已购买充值卡列表, true},

	"PayMoneyToKa":          {"余额购买充值卡", controller.UserApi_余额购买充值卡, true},
	"GetUserClassList":      {"取用户类型列表", controller.UserApi_取用户类型列表, true},
	"SetUserClass":          {"置用户类型", controller.UserApi_置用户类型, true},
	"RunJS":                 {"云函数执行", controller.UserApi_云函数执行, true},
	"TaskPoolNewData":       {"任务池_任务创建", controller.UserApi_任务池_任务创建, true},
	"TaskPoolGetData":       {"任务池_任务查询", controller.UserApi_任务池_任务查询, true},
	"TaskPoolGetDataList":   {"任务池_取任务列表", controller.UserApi_任务池_取任务列表, true},
	"TaskPoolGetTask":       {"任务池_任务处理获取", controller.UserApi_任务池_任务处理获取, true},
	"TaskPoolSetTask":       {"任务池_任务处理返回", controller.UserApi_任务池_任务处理返回, true},
	"TaskPoolGetTypeStatus": {"任务池_取类型状态", controller.UserApi_任务池_取类型状态, true},
	"GetUserConfig":         {"取用户云配置", controller.UserApi_取用户云配置, true},
	"SetUserConfig":         {"置用户云配置", controller.UserApi_置用户云配置, true},
	"SetAgentUid":           {"置代理标志", controller.UserApi_置代理标志, true},
	"GetKaInfo":             {"取卡号详情", controller.UserApi_取卡号详情, true},
	"GetJwtToken":           {"取jwt令牌", controller.UserApi_取jwtToken, true},
	"GetUploadToken":        {"云存储_取文件上传授权", controller.UserApi_云存储_取文件上传授权, true},
	"LoginShortUrl":         {"取登陆短链", controller.UserApi_取登陆短链, true},

	//快验Api
	"VmpComputeAuth":         {"VMP计算授权码", controller.UserApi_VMP计算授权码, false},
	"VmpComputeAuthRoot":     {"VMP计算授权码防山寨", controller.UserApi_VMP计算授权码防山寨, false},
	"KyApiSendSms":           {"快验发送验证码短信", controller.KyApiSendSms, false},
	"KyApiJiYanVerifyTicket": {"快验_极验验证码结果验证", controller.K快验_极验验证码结果验证, false},
}

type 路由信息 struct {
	Z中文名  string
	Z指向函数 func(*gin.Context)
	X显示   bool //是否显示到前段
}

var J集_UserAPi路由_加密 加密路由信息

type 加密路由信息 struct {
	L路由md5 string //每次更新加密路由缓存, 都更新这个索引,每次读取路由,都检测索引是否和缓存相同,如果不同,更新索引
	J加密路由  map[string]string
	D读写锁   sync.RWMutex
}

func (j *加密路由信息) G更新md5APi名称(盐值 string) {
	if 盐值 == "" {
		j.J加密路由 = make(map[string]string, 0)
		return
	}
	if !j.D读写锁.TryLock() {
		return
	}
	defer j.D读写锁.Unlock()

	j.J加密路由 = make(map[string]string, len(J集_UserAPi路由)+1)
	局_临时文本 := utils2.Md5String("GetToken" + 盐值)
	j.J加密路由[局_临时文本] = "GetToken"

	fmt.Printf("API名称加密已更新:%s => %s\n", j.J加密路由[局_临时文本], 局_临时文本)
	局_路由md5原值 := ""
	for 局_用户api := range J集_UserAPi路由 {
		局_哈希后的值 := utils2.Md5String(局_用户api + 盐值)
		j.J加密路由[局_哈希后的值] = 局_用户api
		fmt.Printf("API名称加密已更新:%s => %s\n", 局_用户api, 局_哈希后的值)
		局_路由md5原值 = 局_路由md5原值 + 局_哈希后的值
	}

	j.L路由md5 = utils2.Md5String(局_路由md5原值)
	global.H缓存.Set("J集_UserAPi加密路由md5", j.L路由md5, -1)
}

func (j *加密路由信息) Q取md5APi名称(md5值 string) (string, bool) {
	if len(md5值) != 32 {
		return "", false
	}
	局_用户api, ok := j.J加密路由[md5值]
	if ok {
		return 局_用户api, ok
	}

	云_L路由md5, ok := global.H缓存.Get("J集_UserAPi加密路由md5")
	if !ok || 云_L路由md5 != j.L路由md5 {
		j.G更新md5APi名称(setting.Q系统设置().Y用户API加密盐)
	}

	局_用户api, ok = j.J加密路由[md5值]
	return 局_用户api, ok
}
