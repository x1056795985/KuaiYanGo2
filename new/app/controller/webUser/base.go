package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"net/http"
	"server/Service/Captcha"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_Log"
	"server/Service/Ser_UserClass"
	"server/config"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/webUser/appInfoWebUser"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"server/utils"
	"server/utils/Qqwry"
	"strconv"
	"strings"
	"time"
)

type Base struct {
	Common.Common
}

func NewBaseController() *Base {
	return &Base{}
}

// 销售统计
func (C *Base) LoginUserOrKa(c *gin.Context) {
	var 请求 struct {
		AppId     int    `json:"AppId" binging:"required,min=10000"`                    // Appid 必填
		UserOrKa  string `json:"UserOrKa" binding:"required,min=6,max=190" zh:"用户名或卡号"` // 用户名
		Password  string `json:"password"   zh:"密码"`                                    // 密码
		Captcha   string `json:"captcha"`                                               // 验证码
		CaptchaId string `json:"captchaId"`                                             // 验证码ID
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		Uid           int
		appInfo       DB.DB_AppInfo
		appUser       DB.DB_AppUser
		DB_links_user DB.DB_LinksToken
		kaInfo        DB.DB_Ka
		user          DB.DB_User
		网页用户中心配置      dbm.DB_AppInfoWebUser
	}{}

	客户端ip := c.ClientIP()
	// 判断验证码是否开启
	var err error
	tx := *global.GVA_DB

	if info.appInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "AppId不存在")

		return
	}
	info.网页用户中心配置, err = service.NewAppInfoWebUser(c, &tx).Info(请求.AppId)
	if err != nil || info.网页用户中心配置.Status != 1 {
		response.FailWithMessage(c, constant.C常_关闭提示)
		c.Abort()
		return
	}
	openCaptcha := info.网页用户中心配置.CaptchaLogin // 是否开启防暴次数
	openCaptchaTimeOut := 10                  // 缓存超时时间
	v, ok := global.H缓存.Get(客户端ip)            // 获取这个ip已经被请求次数
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
		if !Captcha.Captcha_Verify点选(请求.CaptchaId, 请求.Captcha, true) {
			response.FailWithMessage(c, "验证码错误")
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "["+strconv.Itoa(请求.AppId)+"]验证码错误:"+请求.Captcha, constant.APPID_Web用户中心)
			return
		}
	}

	if info.appInfo.AppType >= 3 {
		info.kaInfo, err = service.NewKa(c, &tx).InfoKa(请求.UserOrKa)
		if err != nil {
			response.FailWithMessage(c, "卡号不存在")
			return
		}
		if info.kaInfo.Status == 2 {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "["+strconv.Itoa(请求.AppId)+"]卡号已冻结", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "卡号已冻结")
			return
		}
		info.Uid = info.kaInfo.Id
	} else {
		info.user, err = service.NewUser(c, &tx).InfoName(请求.UserOrKa)
		// 没查到数据  或  取反(密码正确)
		if err != nil || !utils.BcryptCheck(请求.Password, info.user.PassWord) {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "["+strconv.Itoa(请求.AppId)+"]用户或密码错误", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "用户或密码错误")
			return
		}
		if info.user.Status == 2 {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "["+strconv.Itoa(请求.AppId)+"]账号已冻结", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "账号已冻结")
			return
		}
		if info.user.UPAgentId != 0 {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "["+strconv.Itoa(请求.AppId)+"]代理商请登录代理平台,禁止登陆用户中心", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "代理商请登录代理平台,禁止登陆用户中心")
			return
		}
		info.Uid = info.user.Id
	}
	info.appUser, err = service.NewAppUser(c, &tx, 请求.AppId).InfoUid(info.Uid)
	var 局_老用户 = info.appUser.Id > 0
	if 局_老用户 {
		//老用户不用处理
	} else {
		//新用户验证绑定信息设置空即可,等他登陆用户自动绑定了

		//没有这个用户,应该是第一次登录应用,添加进去
		switch info.appInfo.AppType {
		case 1:
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", 1, time.Now().Unix(), 0, 0, "")
		case 2: //账号限时
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", 1, 0, 0, 0, "")
		case 3:
			//卡号模式,制卡人就是归属代理 如果是管理员制造的卡, 就使用代理标志为归属uid
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", S三元(info.kaInfo.MaxOnline == 0, 1, info.kaInfo.MaxOnline), time.Now().Unix()+info.kaInfo.VipTime, info.kaInfo.VipNumber, info.kaInfo.UserClassId, info.kaInfo.AdminNote)
			_ = Ser_Ka.Ka修改已用次数加一([]int{info.Uid})
		case 4:
			//卡号模式,制卡人就是归属代理
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", S三元(info.kaInfo.MaxOnline == 0, 1, info.kaInfo.MaxOnline), info.kaInfo.VipTime, info.kaInfo.VipNumber, info.kaInfo.UserClassId, info.kaInfo.AdminNote)
			_ = Ser_Ka.Ka修改已用次数加一([]int{info.Uid})
		default:
			//???应该不会到这里
			response.FailWithMessage(c, "AppInfo.AppType错误")
		}

		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, info.appInfo.AppId, 请求.UserOrKa, info.appInfo.AppName, info.DB_links_user.AppVer, "新添加软件用户时失败报错信息:"+err.Error(), c.ClientIP())
			response.FailWithMessage(c, "New用户信息内部错误")
			return
		}
		// 注册送卡  只有 账号模式才使用
		if info.appInfo.RegisterGiveKaClassId > 0 && (info.appInfo.AppType == 1 || info.appInfo.AppType == 2) {
			_ = ka.L_ka.K卡类直冲_事务(c, info.appInfo.RegisterGiveKaClassId, info.Uid)
		}
		ka.L_ka.Z置归属代理(c, info.appInfo.AppId, info.Uid, info.DB_links_user.AgentUid) //失败也不影响
		//重新读取用户信息
		info.appUser, err = service.NewAppUser(c, &tx, 请求.AppId).InfoUid(info.Uid)
	}

	info.DB_links_user.Uid = info.appUser.Uid
	info.DB_links_user.User = 请求.UserOrKa
	info.DB_links_user.Tab = ""
	info.DB_links_user.AppIdEx = 请求.AppId
	info.DB_links_user.Key = info.appUser.Key
	info.DB_links_user.Ip = c.ClientIP()
	省市, 运行商, err := Qqwry.Ip查信息(info.DB_links_user.Ip)
	if err == nil && 省市 != "" {
		info.DB_links_user.IPCity = 省市 + " " + 运行商
	}
	info.DB_links_user.Status = 1
	info.DB_links_user.LoginTime = time.Now().Unix()
	info.DB_links_user.OutTime = 36000 //退出时间
	info.DB_links_user.LastTime = info.DB_links_user.LoginTime
	info.DB_links_user.Token = strings.ToUpper(rand_string.RandomLetter(32))
	info.DB_links_user.LoginAppid = constant.APPID_Web用户中心
	err = global.GVA_DB.Create(&info.DB_links_user).Error
	go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "["+strconv.Itoa(请求.AppId)+"]登录", constant.APPID_Web用户中心)

	//账号模式登录成功把登录信息写到账号表
	if info.appInfo.AppType == 1 || info.appInfo.AppType == 2 {
		_, err = service.NewUser(c, &tx).Update(info.appInfo.AppId, map[string]interface{}{"LoginAppid": constant.APPID_Web用户中心, "LoginIp": c.ClientIP(), "LoginTime": time.Now().Unix()})
		if err != nil {
			局_log := "账号模式登录成功把登录最后时间信息写到账号表失败:" + err.Error()
			Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, info.appInfo.AppId, 请求.UserOrKa, "webUser", strconv.Itoa(请求.AppId), 局_log, c.ClientIP())
			return
		}
	}

	var 局_用户类型 DB.DB_UserClass
	局_用户类型, ok = Ser_UserClass.Id取详情(请求.AppId, info.appUser.UserClassId)
	if !ok {
		局_用户类型.Name = "已删待改"
		局_用户类型.Mark = 0
	}

	// 用户信息结构体
	type userInfo struct {
		DB.DB_AppUser
		Name          string `json:"name"`
		UserClassMark int    `json:"userClassMark"`
		UserClassName string `json:"userClassName"`
		NewAppUser    bool   `json:"newAppUser"`
	}
	// 定义统一的响应结构体
	type loginResponse struct {
		UserInfo userInfo `json:"userInfo"`
		Token    string   `json:"token"`
	}
	// 构建响应数据
	responseData := loginResponse{
		Token: info.DB_links_user.Token,
		UserInfo: userInfo{
			DB_AppUser:    info.appUser,
			Name:          请求.UserOrKa,
			UserClassName: 局_用户类型.Name,
			UserClassMark: 局_用户类型.Mark,
			NewAppUser:    !局_老用户,
		},
	}
	global.H缓存.Delete(客户端ip) //重置防暴次数
	response.OkWithDetailed(c, responseData, "登录成功")
	return
}

func (C *Base) LoginKey(c *gin.Context) {
	//http://127.0.0.1:18888/userApi/base/loginKey?k=10001&j=\pages\user\home
	局_key := c.Query("k")
	局_jumpUrl := c.Query("j") //302 回跳的地址
	var info = struct {
		来源links_user  DB.DB_LinksToken
		DB_links_user DB.DB_LinksToken
		appInfo       DB.DB_AppInfo
		系统设置          config.X系统设置
		网页用户中心配置      dbm.DB_AppInfoWebUser
	}{}
	var err error
	tx := *global.GVA_DB
	局_key = constant.H缓存前缀_LoginURLPrefix + 局_key

	if Data缓存, ok := global.H缓存.Get(局_key); ok {
		info.来源links_user, err = service.NewLinksToken(c, &tx).Info(Data缓存.(int))
		if err != nil {
			goto 结束开始跳转
		}
	}

	if info.来源links_user.Status != 1 { //已经不是正常状态了 ,可能改过密码
		goto 结束开始跳转
	}
	info.appInfo, err = service.NewAppInfo(c, &tx).Info(info.来源links_user.LoginAppid)
	if err != nil {
		goto 结束开始跳转
	}

	info.网页用户中心配置, err = service.NewAppInfoWebUser(c, &tx).Info(info.来源links_user.LoginAppid)
	if err != nil || info.网页用户中心配置.Status != 1 {
		response.FailWithMessage(c, constant.C常_关闭提示)
		return
	}
	info.DB_links_user.Uid = info.来源links_user.Uid
	info.DB_links_user.User = info.来源links_user.User
	info.DB_links_user.Tab = ""
	info.DB_links_user.AppIdEx = info.来源links_user.LoginAppid
	info.DB_links_user.Key = info.来源links_user.Key
	info.DB_links_user.Ip = c.ClientIP()
	if 省市, 运行商, err2 := Qqwry.Ip查信息(info.DB_links_user.Ip); err2 == nil && 省市 != "" {
		info.DB_links_user.IPCity = 省市 + " " + 运行商
	}
	info.DB_links_user.Status = 1
	info.DB_links_user.LoginTime = time.Now().Unix()
	info.DB_links_user.OutTime = 36000 //退出时间
	info.DB_links_user.LastTime = info.DB_links_user.LoginTime
	info.DB_links_user.Token = strings.ToUpper(rand_string.RandomLetter(32))
	info.DB_links_user.LoginAppid = constant.APPID_Web用户中心
	err = global.GVA_DB.Create(&info.DB_links_user).Error
	if err != nil {
		goto 结束开始跳转
	}
	go Ser_Log.Log_写登录日志(info.来源links_user.User, c.ClientIP(), "["+strconv.Itoa(info.来源links_user.LoginAppid)+"]登录", constant.APPID_Web用户中心)

	//账号模式登录成功把登录信息写到账号表
	if info.appInfo.AppType == 1 || info.appInfo.AppType == 2 {
		_, err = service.NewUser(c, &tx).Update(info.appInfo.AppId, map[string]interface{}{"LoginAppid": constant.APPID_Web用户中心, "LoginIp": c.ClientIP(), "LoginTime": time.Now().Unix()})
		if err != nil {
			局_log := "账号模式登录成功把登录最后时间信息写到账号表失败:" + err.Error()
			Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, info.appInfo.AppId, info.来源links_user.User, "webUser", strconv.Itoa(info.appInfo.AppId), 局_log, c.ClientIP())
		}
	}

	//不管登陆成功还是失败,都需要跳转到这里
结束开始跳转:

	//设置302跳转
	局_jumpUrl = appInfoWebUser.L_appInfoWebUser.Q用户中心域名(c, info.网页用户中心配置.Id) + "/user/" + strconv.Itoa(info.来源links_user.LoginAppid) + "/#/" + 局_jumpUrl + "?tempToken=" + info.DB_links_user.Token
	//判断左边是否为https

	//设置临时token 前端路由守卫,会把cook放到token内 httpOnly 必须为false 否则js无法读取cookies
	cookie := &http.Cookie{
		Name:     "tempToken",
		Value:    info.DB_links_user.Token,
		MaxAge:   36000,
		Path:     "/",
		Domain:   "",                                    // 根据需要设置具体域名
		Secure:   strings.HasPrefix(局_jumpUrl, "https"), // 重要：SameSite=None 必须配合 HTTPS
		HttpOnly: false,                                 // 根据前端需求设置
		SameSite: http.SameSiteNoneMode,
	}

	c.Header("Set-Cookie", cookie.String())

	c.Redirect(302, 局_jumpUrl)
	return
}
