package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_Log"
	"server/Service/Ser_UserClass"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/ka"
	"server/new/app/models/constant"
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
		Password  string `json:"password" binding:"required,min=6,max=190" zh:"密码"`     // 密码
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
	}{}
	var err error
	tx := *global.GVA_DB

	if info.appInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "AppId不存在")

		return
	}

	if info.appInfo.AppType >= 3 {
		info.kaInfo, err = service.NewKa(c, &tx).InfoKa(请求.UserOrKa)
		if err != nil {
			response.FailWithMessage(c, "卡号不存在")
			return
		}
		if info.kaInfo.Status == 2 {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "卡号已冻结", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "卡号已冻结")
			return
		}
		info.Uid = info.kaInfo.Id
	} else {
		info.user, err = service.NewUser(c, &tx).InfoName(请求.UserOrKa)
		// 没查到数据  或  取反(密码正确)
		if err != nil || !utils.BcryptCheck(请求.Password, info.user.PassWord) {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "用户或密码错误", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "用户或密码错误")
			return
		}
		if info.user.Status == 2 {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "账号已冻结", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "账号已冻结")
			return
		}
		if info.user.UPAgentId != 0 {
			go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "代理商请登录代理平台", constant.APPID_Web用户中心)
			response.FailWithMessage(c, "代理商请登录代理平台")
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
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", 1, time.Now().Unix(), 0, 0, "", info.DB_links_user.AgentUid)
		case 2: //账号限时
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", 1, 0, 0, 0, "", info.DB_links_user.AgentUid)
		case 3:
			//卡号模式,制卡人就是归属代理 如果是管理员制造的卡, 就使用代理标志为归属uid
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", S三元(info.kaInfo.MaxOnline == 0, 1, info.kaInfo.MaxOnline), time.Now().Unix()+info.kaInfo.VipTime, info.kaInfo.VipNumber, info.kaInfo.UserClassId, info.kaInfo.AdminNote, info.DB_links_user.AgentUid)
			_ = Ser_Ka.Ka修改已用次数加一([]int{info.Uid})
		case 4:
			//卡号模式,制卡人就是归属代理
			err = Ser_AppUser.New用户信息(info.appInfo.AppId, info.Uid, "", S三元(info.kaInfo.MaxOnline == 0, 1, info.kaInfo.MaxOnline), info.kaInfo.VipTime, info.kaInfo.VipNumber, info.kaInfo.UserClassId, info.kaInfo.AdminNote, info.DB_links_user.AgentUid)
			_ = Ser_Ka.Ka修改已用次数加一([]int{info.Uid})
		default:
			//???应该不会到这里
			response.FailWithMessage(c, "AppInfo.AppType错误")
		}

		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 请求.UserOrKa, info.appInfo.AppName, info.DB_links_user.AppVer, "新添加软件用户时失败报错信息:"+err.Error(), c.ClientIP())
			response.FailWithMessage(c, "New用户信息内部错误")
			return
		}
		// 注册送卡  只有 账号模式才使用
		if info.appInfo.RegisterGiveKaClassId > 0 && (info.appInfo.AppType == 1 || info.appInfo.AppType == 2) {
			_ = ka.L_ka.K卡类直冲_事务(c, info.appInfo.RegisterGiveKaClassId, info.Uid)
		}
		//重新读取用户信息
		info.appUser, err = service.NewAppUser(c, &tx, 请求.AppId).InfoUid(info.Uid)
	}

	info.DB_links_user.Uid = info.appUser.Id
	info.DB_links_user.User = 请求.UserOrKa
	info.DB_links_user.Tab = strconv.Itoa(请求.AppId)
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
	go Ser_Log.Log_写登录日志(请求.UserOrKa, c.ClientIP(), "Web用户中心平台("+strconv.Itoa(请求.AppId)+")登录", constant.APPID_Web用户中心)

	//账号模式登录成功把登录信息写到账号表
	if info.appInfo.AppType == 1 || info.appInfo.AppType == 2 {
		_, err = service.NewUser(c, &tx).Update(info.appInfo.AppId, map[string]interface{}{"LoginAppid": constant.APPID_Web用户中心, "LoginIp": c.ClientIP(), "LoginTime": time.Now().Unix()})
		if err != nil {
			局_log := "账号模式登录成功把登录最后时间信息写到账号表失败:" + err.Error()
			Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 请求.UserOrKa, "webUser", strconv.Itoa(请求.AppId), 局_log, c.ClientIP())
			return
		}
	}

	var 局_用户类型 DB.DB_UserClass
	局_用户类型, ok := Ser_UserClass.Id取详情(请求.AppId, info.appUser.UserClassId)
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

	response.OkWithDetailed(c, responseData, "登录成功")
	return
}
