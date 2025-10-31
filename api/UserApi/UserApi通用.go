package UserApi

import (
	. "EFunc/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
	"server/Service/Captcha"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/Service/Ser_UserClass"
	"server/Service/Ser_UserConfig"
	"server/api/UserApi/response"
	"server/global"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/appUser"
	"server/new/app/logic/common/blacklist"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/publicData"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"time"
)

func UserApi_Api不存在(c *gin.Context) {
	response.X响应状态(c, response.Status_Api不存在)
	return
}

func UserApi_GetToken(c *gin.Context) {
	//请求json, err := fastjson.Parse(c.GetString("局_json明文"))
	局_临时通用, _ := c.Get("AppInfo")
	AppInfo := 局_临时通用.(DB.DB_AppInfo)
	var 局_通讯AES密钥 = ""
	if AppInfo.CryptoType == 3 || AppInfo.CryptoType == 2 {
		局_通讯AES密钥 = c.GetString("局_CryptoKeyAes")
	} else if AppInfo.CryptoType == 2 {
		局_通讯AES密钥 = AppInfo.CryptoKeyAes
	}
	在线信息, err2 := Ser_LinkUser.New(0, 1, AppInfo.AppId, AppInfo.OutTime, "游客", "", "", c.ClientIP(), 局_通讯AES密钥)

	if err2 != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", AppInfo.AppName, "", fmt.Sprintf("用户获取Token插入新值失败:%v", err2.Error()), c.ClientIP())
		response.X响应状态(c, response.Status_SQl错误)
		return
	}

	//这里吧成功的状态
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 请求响应_Data_GetToken{Token: 在线信息.Token, CryptoKeyAes: 局_通讯AES密钥, IP: c.ClientIP()})
}

// 回复json结构体
type 请求响应_Data_GetToken struct {
	Token        string `json:"Token"`
	CryptoKeyAes string `json:"CryptoKeyAes"`
	IP           string `json:"IP"`
}

// 登陆
func UserApi_用户登录(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserPassLogin","UserOrKa":"aaaaaa","PassWord":"AF15D5FDACD5FDFEA300E88A8E253E82","Key":"677F23CB3FA0055B5FD03916D6AB3C9A","Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","AppVer":"1.0.1","Captcha":{"Id":"","Value":""}}
	if !版本号_检测可用(string(请求json.GetStringBytes("AppVer")), AppInfo.AppVer) {
		response.X响应状态(c, response.Status_版本不可用)
		return
	}
	var 局_Uid = 0
	var 局_卡 DB.DB_Ka
	var err error
	if len(请求json.GetStringBytes("Key")) > 191 {
		response.X响应状态消息(c, response.Status_操作失败, "绑定信息长度不能超过191")
		return
	}
	if blacklist.Is黑名单(string(请求json.GetStringBytes("Key")), AppInfo.AppId) {
		response.X响应状态消息(c, response.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}
	if 局_在线信息.Uid != 0 {
		response.X响应状态消息(c, response.Status_操作失败, "已登陆,无需重复登陆")
		return
	}

	var 局_卡号或用户名 = strings.TrimSpace(string(请求json.GetStringBytes("UserOrKa")))
	if AppInfo.AppType == 3 || AppInfo.AppType == 4 {
		//卡号
		局_卡, err = Ser_Ka.Ka卡号取详情(局_卡号或用户名)
		if err != nil || 局_卡.AppId != AppInfo.AppId {
			response.X响应状态消息(c, response.Status_登录失败, "卡号不存在")
			return
		}
		if 局_卡.Status != 1 {
			go Ser_Log.Log_写登录日志(局_卡.Name, c.ClientIP(), "卡号已冻结", 局_在线信息.LoginAppid)
			response.X响应状态消息(c, response.Status_登录失败, "卡号已冻结")
			return
		}
		局_Uid = 局_卡.Id
		局_卡号或用户名 = 局_卡.Name
	} else {
		//账号
		var 局_User DB.DB_User
		局_User, ok := Ser_User.User取详情(局_卡号或用户名)
		if !ok {
			response.X响应状态消息(c, response.Status_登录失败, "用户不存在")
			return
		}

		if 局_User.PassWord == "" || !utils2.BcryptCheck(string(请求json.GetStringBytes("PassWord")), 局_User.PassWord) {
			go Ser_Log.Log_写登录日志(局_User.User, c.ClientIP(), "密码错误:"+string(请求json.GetStringBytes("PassWord")), 局_在线信息.LoginAppid)
			response.X响应状态消息(c, response.Status_登录失败, "用户名或密码错误")
			return
		}
		if 局_User.Status != 1 {
			go Ser_Log.Log_写登录日志(局_User.User, c.ClientIP(), "账号已冻结", 局_在线信息.LoginAppid)
			response.X响应状态消息(c, response.Status_登录失败, "账号已冻结")
			return
		}
		if 局_User.UPAgentId != 0 {
			go Ser_Log.Log_写登录日志(局_User.User, c.ClientIP(), "代理商请登录代理平台", 局_在线信息.LoginAppid)
			response.X响应状态消息(c, response.Status_登录失败, "代理商请登录代理平台")
			return
		}
		局_Uid = 局_User.Id
		局_卡号或用户名 = 局_User.User
	}
	db := *global.GVA_DB
	var 局_AppUser DB.DB_AppUser
	var 局_老用户 = Ser_AppUser.Uid是否存在(AppInfo.AppId, 局_Uid)
	if 局_老用户 {
		局_AppUser, _ = Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid)
		//如果用户key是空的直接重新绑定

		if 局_AppUser.Key == "" {
			//检查是否可以绑定相同信息
			if AppInfo.IsUserKeySame == 2 && Ser_AppUser.B绑定信息是否存在(AppInfo.AppId, string(请求json.GetStringBytes("Key"))) {
				response.X响应状态消息(c, response.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
				return
			}

			Ser_AppUser.Set绑定信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")))

			_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
				Type:   constant.LogKey_绑定,
				User:   局_卡号或用户名,
				Uid:    局_Uid,
				AppId:  AppInfo.AppId,
				OldKey: 局_AppUser.Key,
				NewKey: string(请求json.GetStringBytes("Key")),
				Time:   time.Now().Unix(),
				Ip:     c.ClientIP(),
				Note:   "无绑定信息登陆自动绑定",
			})
			if err != nil {
				global.GVA_LOG.Error("修改绑定信息日志写入失败:" + err.Error())
			}
			局_AppUser.Key = string(请求json.GetStringBytes("Key"))
		}

		//老用户验证绑定信息是否相同
		if AppInfo.VerifyKey == 3 || AppInfo.VerifyKey == 4 {
			//1 免验证可以换绑 2 免验证禁止换绑 3 验证可以换绑 4 验证禁止换
			if 局_AppUser.Key != string(请求json.GetStringBytes("Key")) {
				go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "登录绑定信息验证失败:"+string(请求json.GetStringBytes("Key")), 局_在线信息.LoginAppid)
				response.X响应状态(c, response.Status_绑定信息验证失败)
				return
			}
		}

	} else {

		//新用户验证绑定信息是否存在
		if AppInfo.IsUserKeySame == 2 {
			//1 免验证可以换绑 2 免验证禁止换绑 3 验证可以换绑 4 验证禁止换
			if Ser_AppUser.B绑定信息是否存在(AppInfo.AppId, string(请求json.GetStringBytes("Key"))) {
				go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "登录注册绑定信息已存在:"+string(请求json.GetStringBytes("Key")), 局_在线信息.LoginAppid)
				response.X响应状态(c, response.Status_绑定信息已被其他用户使用)
				return
			}
		}

		if AppInfo.AppType == 3 || AppInfo.AppType == 4 {
			if 局_卡.Num >= 局_卡.NumMax {
				response.X响应状态消息(c, response.Status_登录失败, "卡号已经使用到最大次数")
				return
			}
		}
		局_归属代理uid := 0
		if AppInfo.AppType == 3 || AppInfo.AppType == 4 {
			//账号模式,制卡人就是归属代理 如果是管理员制造的卡, 就使用代理标志为归属uid
			局_归属代理uid = Ser_User.User用户名取id(局_卡.RegisterUser)
			if 局_归属代理uid == 0 {
				局_归属代理uid = 局_在线信息.AgentUid
			}
		}
		//没有这个用户,应该是第一次登录应用,添加进去
		switch AppInfo.AppType {
		case 1:
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), 1, time.Now().Unix(), 0, 0, "", 局_在线信息.AgentUid)
		case 2: //账号限时
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), 1, 0, 0, 0, "", 局_在线信息.AgentUid)
		case 3:
			//卡号模式,制卡人就是归属代理 如果是管理员制造的卡, 就使用代理标志为归属uid
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline), time.Now().Unix()+局_卡.VipTime, 局_卡.VipNumber, 局_卡.UserClassId, 局_卡.AdminNote, 局_归属代理uid)
			_ = Ser_Ka.Ka修改已用次数加一([]int{局_Uid})
		case 4:
			//卡号模式,制卡人就是归属代理
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline), 局_卡.VipTime, 局_卡.VipNumber, 局_卡.UserClassId, 局_卡.AdminNote, 局_归属代理uid)
			_ = Ser_Ka.Ka修改已用次数加一([]int{局_Uid})
		default:
			//???应该不会到这里
			response.X响应状态消息(c, response.Status_SQl错误, "AppInfo.AppType错误")
		}

		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 局_卡号或用户名, AppInfo.AppName, 局_在线信息.AppVer, "新添加软件用户时失败报错信息:"+err.Error(), c.ClientIP())
			response.X响应状态消息(c, response.Status_SQl错误, "New用户信息内部错误")
			return
		}

		_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
			Type:   constant.LogKey_绑定,
			User:   局_卡号或用户名,
			Uid:    局_Uid,
			AppId:  AppInfo.AppId,
			OldKey: "",
			NewKey: string(请求json.GetStringBytes("Key")),
			Time:   time.Now().Unix(),
			Ip:     c.ClientIP(),
			Note:   "新用户登陆自动绑定",
		})
		if err != nil {
			global.GVA_LOG.Error("修改绑定信息日志写入失败:" + err.Error())
		}

		// 注册送卡  只有 账号模式才使用
		if AppInfo.RegisterGiveKaClassId > 0 && (AppInfo.AppType == 1 || AppInfo.AppType == 2) {
			_ = ka.L_ka.K卡类直冲_事务(c, AppInfo.RegisterGiveKaClassId, 局_Uid)
			//局_注册送卡, 局_制卡结果 := Ser_Ka.Ka单卡创建(AppInfo.RegisterGiveKaClassId, "系统自动", "用户注册系统自动制卡赠送充值", "", 0)
			//if 局_制卡结果 == nil {
			//	_ = ka.L_ka.K卡号充值_事务(c, AppInfo.AppId, 局_注册送卡.Name, 局_卡号或用户名, "")
			//}
		}

	}
	局_AppUser, _ = Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid) //充值之后重新读取一遍
	if 局_AppUser.Status == 2 {
		go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "已冻结无法登录", 局_在线信息.LoginAppid)
		response.X响应状态(c, response.Status_已冻结无法登录)
		return
	}

	if AppInfo.Status == 2 {
		//免费运营模式不检查时间直接登录成功
	} else {
		if AppInfo.AppType == 2 || AppInfo.AppType == 4 { //计点方式
			if 局_AppUser.VipTime <= 0 {
				go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "非Vip禁止登录", 局_在线信息.LoginAppid)
				response.X响应状态(c, response.Status_Vip已到期)
				return
			}
		} else { //计时模式
			if 局_AppUser.VipTime <= time.Now().Unix() { // 相等也限制登录, 防止刚注册 时间和过期正好相当
				go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "Vip已过期", 局_在线信息.LoginAppid)
				response.X响应状态(c, response.Status_Vip已到期)
				return
			}
		}
	}

	局_已经在线数量 := Ser_LinkUser.Get取在线数量(AppInfo.AppId, 局_AppUser.Uid)
	var 局_要踢掉的数量 = 0
	if len(局_已经在线数量) >= 局_AppUser.MaxOnline {
		if AppInfo.ExceedMaxOnlineOut == 1 {
			//踢掉最早在线
			局_要踢掉的数量 = len(局_已经在线数量) - 局_AppUser.MaxOnline + 1
			_ = Ser_LinkUser.Set批量注销(局_已经在线数量[:局_要踢掉的数量], Ser_LinkUser.Z注销_超过同时在线注销)
			//已经登录的数量-最大数量 +1
			go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "登录同时在线超过最大值已注销最早登录:"+strconv.Itoa(局_要踢掉的数量), 局_在线信息.LoginAppid)

		} else if AppInfo.ExceedMaxOnlineOut == 2 {
			//直接提示
			go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "同时在线超过最大值", 局_在线信息.LoginAppid)
			response.X响应状态(c, response.Status_同时在线超过最大值)
			return
		}

	}

	//登录成功吧数据写入在线信息内
	tx := *global.GVA_DB
	data := map[string]interface{}{
		"Uid":    局_Uid,
		"User":   局_卡号或用户名,
		"Key":    局_AppUser.Key,
		"Tab":    string(请求json.GetStringBytes("Tab")),
		"AppVer": string(请求json.GetStringBytes("AppVer")),
	}
	_, err = service.NewLinksToken(c, &tx).Update(局_在线信息.Id, data)
	if err != nil {
		//mark 一个新奇的bug, Tab是ansi编码的中文, go字符串,类型为utf8 获取字节数组string转文本就会导致是乱码,导致修改数据库失败,看来得加参数校验了
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}
	//没有归属代理,但是在线信息已经有代理标志了 赋予软件用户归属代理
	if 局_AppUser.AgentUid == 0 && 局_在线信息.AgentUid != 0 {
		_, err = service.NewAppUser(c, &tx, 局_在线信息.LoginAppid).UpdateUid(局_Uid, map[string]interface{}{"AgentUid": 局_在线信息.AgentUid})
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
		局_AppUser.AgentUid = 局_在线信息.AgentUid
	}

	//用户已有归属代理,但是和在线信息代理标志不同,修改在线代理标志
	if 局_AppUser.AgentUid != 局_在线信息.AgentUid {
		_, err = service.NewLinksToken(c, &tx).Update(局_在线信息.Id, map[string]interface{}{"AgentUid": 局_AppUser.AgentUid})
		局_在线信息.AgentUid = 局_AppUser.AgentUid
	}

	//登录成功写日志
	if 局_老用户 {
		go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "用户登录", 局_在线信息.LoginAppid)
	} else {
		go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "新用户登录注册", 局_在线信息.LoginAppid)
	}

	//账号模式登录成功把登录信息写到账号表
	if AppInfo.AppType == 1 || AppInfo.AppType == 2 {
		go Ser_User.Id置最后登录AppId(局_Uid, AppInfo.AppId, c.ClientIP())
	}

	var 局_用户类型 DB.DB_UserClass
	局_用户类型, ok := Ser_UserClass.Id取详情(AppInfo.AppId, 局_AppUser.UserClassId)
	if !ok {
		局_用户类型.Name = "已删待改"
		局_用户类型.Mark = 0
	}
	更新上下文缓存在线信息(c)
	//这里吧成功的状态
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"User":          局_卡号或用户名,
		"VipTime":       局_AppUser.VipTime,
		"Key":           局_AppUser.Key,
		"OutUser":       局_要踢掉的数量,
		"UserClassMark": 局_用户类型.Mark,
		"UserClassName": 局_用户类型.Name,
		"VipNumber":     局_AppUser.VipNumber,
		"LoginTime":     time.Now().Unix(),
		"LoginIp":       c.ClientIP(),
		"RegisterTime":  局_AppUser.RegisterTime,
		"NewAppUser":    !局_老用户,
		"AgentUid":      局_AppUser.AgentUid,
	})

}
func UserApi_GetUserIP(c *gin.Context) {
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"IP": c.ClientIP()})
}

func UserApi_用户减少余额(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserReduceMoney","Money":1.01,"Log":"看你长得帅,减些钱","AgentId":10,"AgentMoney":0,"AgentMoneyLog":"代理分成"}

	if 请求json.GetInt("AgentId") > 0 {
		response.X响应状态消息(c, response.Status_操作失败, "服务端1.0.363+该功能已删除,如有需要请使用更安全的apiHook实现")
		return
	}
	var 局_User DB.DB_User
	局_User, ok := Ser_User.Id取详情(局_在线信息.Uid)

	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}
	var 局_增减值 float64
	局_增减值 = 请求json.GetFloat64("Money")
	if 局_增减值 <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "不能为小于等于0")
		return
	}
	if 局_User.Rmb < 局_增减值 {
		response.X响应状态消息(c, response.Status_操作失败, "余额不足")
		return
	}

	新余额, err := Ser_User.Id余额增减(局_User.Id, 局_增减值, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error()) //基本就是余额不足
		return
	}

	go Ser_Log.Log_写余额日志(局_User.User, c.ClientIP(), fmt.Sprintf("%s|新余额%v", 请求json.GetStringBytes("Log"), 新余额), Float64取负值(局_增减值))
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Money": 新余额})
}
func UserApi_用户减少点数(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"UserReduceMoney","VipTime":1.3,"Log":"看你长得帅,扣点钱"}
	if AppInfo.AppType != 2 && AppInfo.AppType != 4 { //检查是不是计点模式
		response.X响应状态消息(c, response.Status_操作失败, "应用非计点模式不可使用")
		return
	}
	var 局_AppUser DB.DB_AppUser
	局_AppUser, ok := Ser_AppUser.Uid取详情(局_在线信息.LoginAppid, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}
	var 局_增减值 = 请求json.GetInt64("VipTime")
	if 局_增减值 <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "不能为小于等于0")
		return
	}
	if 局_AppUser.VipTime < 局_增减值 {
		response.X响应状态消息(c, response.Status_操作失败, "点数不足")
		return
	}

	err := Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, 局_增减值, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error()) //基本就是点数不足
		return
	}

	局_AppUser.VipTime -= 局_增减值
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"VipTime": 局_AppUser.VipTime})
	go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, c.ClientIP(), fmt.Sprintf("%s|剩余%v", 请求json.GetStringBytes("Log"), 局_AppUser.VipNumber), Float64取负值(float64(局_增减值)), AppInfo.AppId, 2)
	return
}
func UserApi_用户减少积分(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserReduceMoney","VipNumber":1.3,"Log":"看你长得帅,扣点钱","UniqueStr":"",UniqueTime:0}
	var 局_AppUser DB.DB_AppUser
	局_AppUser, ok := Ser_AppUser.Uid取详情(局_在线信息.LoginAppid, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}
	var 局_增减值 = 请求json.GetFloat64("VipNumber")

	if 局_增减值 <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "不能为小于等于0")
		return
	}

	if 请求json.GetInt("AgentId") > 0 {
		response.X响应状态消息(c, response.Status_操作失败, "服务端1.0.363+该功能已删除,如有需要请使用更安全的apiHook实现")
		return
	}

	if 局_AppUser.VipNumber < 局_增减值 {
		response.X响应状态消息(c, response.Status_操作失败, "积分不足")
		return
	}

	局_唯一标志 := string(请求json.GetStringBytes("UniqueStr"))

	err := appUser.L_appUser.Uid积分减少(c, AppInfo.AppId, 局_AppUser.Uid, 局_增减值, 局_唯一标志, 请求json.GetInt64("UniqueTime"))
	if err != nil && strings.Contains(err.Error(), "唯一标识") {
		response.X响应状态消息(c, response.Status_唯一标识重复, err.Error())
		return
	}
	if err != nil && strings.Contains(err.Error(), "积分不足") { //基本就是积分不足
		response.X响应状态消息(c, response.Status_积分不足, err.Error())
		return
	}

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	// flosat64 直接
	局_增减值D := decimal.NewFromFloat(局_增减值)
	局_用户积分D := decimal.NewFromFloat(局_AppUser.VipNumber)

	局_用户积分D = 局_用户积分D.Sub(局_增减值D)
	局_AppUser.VipNumber, _ = 局_用户积分D.Float64()

	局_增减值, _ = 局_增减值D.Mul(decimal.NewFromFloat(-1)).Float64() //乘-1 变成负数

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"VipNumber": 局_AppUser.VipNumber})
	go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, c.ClientIP(), fmt.Sprintf("%s|≈%v", 请求json.GetStringBytes("Log"), 局_AppUser.VipNumber), 局_增减值, AppInfo.AppId, 1)
	return
}

func UserApi_取服务器连接状态(c *gin.Context) {
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}

func UserApi_取登录状态(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if 局_在线信息.Status == 1 && 局_在线信息.Uid > 0 {
		response.X响应状态(c, c.GetInt("局_成功Status"))
		return
	}

	response.X响应状态(c, response.Status_未登录)
	return
}

func UserApi_取Vip数据(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if 局_在线信息.Uid == 0 || 局_在线信息.Status != 1 {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态(c, response.Status_Vip已到期)
		return
	}

	var 局_比较值 int64
	if AppInfo.AppType == 2 || AppInfo.AppType == 4 {
		局_比较值 = 0
	} else {
		局_比较值 = time.Now().Unix()
	}

	if 局_AppUser.VipTime > 局_比较值 || AppInfo.AppType == 2 {
		var VipData interface{}
		err := json.Unmarshal([]byte(AppInfo.VipData), &VipData) //VipData被强制Json了 可以直接反序列化
		if err == nil {
			response.X响应状态带数据(c, c.GetInt("局_成功Status"), VipData)
		} else {
			response.X响应状态消息(c, response.Status_操作失败, "Vip数据非标准Json")
		}

		return
	}
	response.X响应状态(c, response.Status_Vip已到期)
	return
}
func UserApi_取应用公告(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppGongGao": AppInfo.AppGongGao})
	return
}
func UserApi_取新版本下载地址(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	AppInfo.UrlDownload = Ser_AppInfo.App下载更新地址变量处理(AppInfo)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppUpDataJson": AppInfo.UrlDownload})
	return
}
func UserApi_取应用专属变量(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	局_云变量数据, err := publicData.L_publicData.Q取值2(c, AppInfo.AppId, 局_变量名)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "变量不存在,请到后台应用编辑,添加专属变量")
		return
	}
	if 局_云变量数据.IsVip == 0 || 检测用户登录在线正常(&局_在线信息) {
		if 局_云变量数据.IsVip > 0 { //只有返回VIP变量时才强制
			c.Set("RSA强制", true)
		}

		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{局_变量名: 局_云变量数据.Value})
	} else {
		response.X响应状态(c, response.Status_未登录)
	}
	return
}
func UserApi_取公共变量(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	取值2, err := publicData.L_publicData.Q取值2(c, 1, 局_变量名)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{局_变量名: 取值2.Value, "QueueCount": 取值2})
	return
}

func UserApi_置公共变量(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a","Value":"aaaaa"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	局_变量值 := string(请求json.GetStringBytes("Value"))
	err := publicData.L_publicData.Z置值(c, 1, 局_变量名, 局_变量值)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}
func UserApi_取代理云配置(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetAgentConfig","Name":"配置1"}
	局_配置名 := string(请求json.GetStringBytes("Name"))
	局_AppUserInfo, _ := Ser_AppUser.Uid取详情(局_在线信息.LoginAppid, 局_在线信息.Uid)

	局_配置值 := Ser_UserConfig.Q取值(50, 局_AppUserInfo.AgentUid, 局_配置名)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{局_配置名: 局_配置值})
	return
}
func UserApi_取用户云配置(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetUserConfig","Name":"配置1"}
	局_配置名 := string(请求json.GetStringBytes("Name"))
	局_配置值 := Ser_UserConfig.Q取值(局_在线信息.LoginAppid, 局_在线信息.Uid, 局_配置名)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{局_配置名: 局_配置值})
	return
}
func UserApi_置用户云配置(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetUserConfig","Name":"配置1","Value":"值"}

	局_配置名 := string(请求json.GetStringBytes("Name"))
	if 局_配置名 == "" {
		response.X响应状态消息(c, response.Status_操作失败, "云配置名不能为空")
		return
	}
	局_配置值 := string(请求json.GetStringBytes("Value"))
	if 局_配置值 == "" { //值为空则删
		global.GVA_DB.Model(DB.DB_UserConfig{}).Delete(DB.DB_UserConfig{
			AppId: 局_在线信息.LoginAppid,
			Uid:   局_在线信息.Uid,
			Name:  局_配置名,
		})
	} else {
		_ = Ser_UserConfig.Z置值(局_在线信息.LoginAppid, 局_在线信息.Uid, 局_配置名, 局_配置值)
	}
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}
func UserApi_取应用最新版本(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetAppVersion","Version":"1.3.5","IsVersionAll":true}
	局_可用版本 := W文本_分割文本(AppInfo.AppVer, "\n")
	if len(局_可用版本) == 0 || AppInfo.AppVer == "" {
		response.X响应状态消息(c, response.Status_操作失败, "应用未设置版本号或格式不正确")
		return
	}

	局_分解版本号 := W文本_分割文本(局_可用版本[0], ".")
	局_分解版本号最新 := 版本号_分解(局_可用版本[0])
	局_版本号当前 := string(请求json.GetStringBytes("Version"))

	局_是否更新 := false
	if 局_版本号当前 != "" {
		局_分解版本号当前 := 版本号_分解(局_版本号当前)
		for I := 0; I < 3; I++ {
			switch I {
			case 0:
				局_是否更新 = 局_分解版本号最新.大版本号 > 局_分解版本号当前.大版本号
			case 1:
				局_是否更新 = 局_分解版本号最新.小版本号 > 局_分解版本号当前.小版本号
			case 2:
				if 请求json.GetBool("IsVersionAll") {
					局_是否更新 = 局_分解版本号最新.编译版本号 > 局_分解版本号当前.编译版本号
				}
			}

			if 局_是否更新 {
				break
			}
		}
	}

	if len(局_分解版本号) == 1 {
		// 只有大版本号
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"NewVersion": 局_可用版本[0], "Version": 局_分解版本号最新.大版本号, "IsUpdate": 局_是否更新})
		return
	} else {
		// 有大小版本号
		局_小数运算, _ := decimal.NewFromString(局_分解版本号[0] + "." + 局_分解版本号[1])
		局_双精度版本, _ := 局_小数运算.Float64()
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"NewVersion": 局_可用版本[0], "Version": 局_双精度版本, "IsUpdate": 局_是否更新})
		return
	}
}
func UserApi_取应用主页Url(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppHomeUrl": AppInfo.UrlHome})
	return
}

// 1.0.42+版本添加可用
func UserApi_取应用基础信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"AppId":            AppInfo.AppId,
		"AppType":          AppInfo.AppType,
		"AppName":          AppInfo.AppName,
		"AppWeb":           AppInfo.AppWeb,
		"Status":           AppInfo.Status,
		"AppStatusMessage": AppInfo.AppStatusMessage,
	})
	return
}
func UserApi_置新绑定信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetAppUserKey","NewKey":"8987657"}
	// 检查是否可以换换绑
	if AppInfo.VerifyKey != 1 && AppInfo.VerifyKey != 3 { //1和3 可以换绑
		response.X响应状态消息(c, response.Status_操作失败, "应用禁止更换绑定信息.")
		return
	}

	局_Uid := 局_在线信息.Uid
	if !检测用户登录在线正常(&局_在线信息) {
		局_账号 := string(请求json.GetStringBytes("User"))

		局密码 := string(请求json.GetStringBytes("PassWord"))
		if 局_账号 == "" {
			response.X响应状态(c, response.Status_未登录)
			return
		} else {
			局_在线信息.User = 局_账号                                //如果出错,写日志时会用到
			if AppInfo.AppType == 3 || AppInfo.AppType == 4 { //是卡号
				局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_账号)
				if 局_Uid == 0 {
					response.X响应状态消息(c, response.Status_操作失败, "卡号不存在.")
					return
				}
			} else {
				局_User, ok := Ser_User.User取详情(局_账号)
				if !ok {
					response.X响应状态消息(c, response.Status_操作失败, "用户不存在.")
					return
				}
				if 局密码 == "" || !utils2.BcryptCheck(局密码, 局_User.PassWord) {
					go Ser_Log.Log_写登录日志(局_User.User, c.ClientIP(), "更换绑定登录时密码错误:"+局密码, AppInfo.AppId)
					response.X响应状态消息(c, response.Status_登录失败, "用户名或密码错误")
					return
				}
				局_Uid = 局_User.Id
			}
		}

	}

	局_信息绑定信息 := string(请求json.GetStringBytes("NewKey"))
	if 局_信息绑定信息 == "" {
		response.X响应状态消息(c, response.Status_绑定信息验证失败, "新绑定信息不能为空.")
		return
	}

	// 检查是否可以绑定相同信息
	if AppInfo.IsUserKeySame == 2 && Ser_AppUser.B绑定信息是否存在(AppInfo.AppId, 局_信息绑定信息) {
		response.X响应状态消息(c, response.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
		return
	}
	if blacklist.Is黑名单(局_信息绑定信息, AppInfo.AppId) {
		response.X响应状态消息(c, response.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.可能刚注册还没登录成功")
		return
	}

	err, 扣时间值 := 绑定信息更换规则校验(c, AppInfo, 局_Uid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	// 如果换绑需要扣点,就执行扣点, 		且原来绑定信息不能为空
	if 扣时间值 > 0 && 局_AppUser.Key != "" {
		err = Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), false)
		if err != nil {
			response.X响应状态消息(c, response.Status_Vip已到期, "剩余会员时间或点数不足.")
			return
		} else {
			局_日志 := "用户置新绑定,旧绑定信息:" + 局_AppUser.Key + ",新绑定信息:" + 局_信息绑定信息
			局_type := 3
			if AppInfo.AppType == 2 || AppInfo.AppType == 4 {
				局_type = 2
			}
			Ser_Log.Log_写积分点数时间日志(局_在线信息.User, c.ClientIP(), 局_日志, D到数值(-扣时间值), AppInfo.AppId, 局_type)

		}
	}
	err = Ser_AppUser.Set绑定信息(AppInfo.AppId, 局_Uid, 局_信息绑定信息)
	if err == nil {
		db := *global.GVA_DB
		_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
			Type:   constant.LogKey_换绑,
			User:   Ser_AppUser.Uid取User(AppInfo.AppId, 局_Uid),
			Uid:    局_Uid,
			AppId:  AppInfo.AppId,
			OldKey: 局_AppUser.Key,
			NewKey: 局_信息绑定信息,
			Time:   time.Now().Unix(),
			Ip:     c.ClientIP(),
			Count:  D到数值(-扣时间值),
			Note:   "置新绑定信息",
		})
		if err != nil {
			global.GVA_LOG.Error("修改绑定信息日志写入失败:" + err.Error())
		}
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"ReduceVipTime": 扣时间值})
	} else {

		_ = Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), true) //退还已经扣除的点数
		response.X响应状态(c, response.Status_SQl错误)
	}

	return
}
func UserApi_解除绑定信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetAppUserKey"}
	// 检查是否可以换换绑
	if AppInfo.VerifyKey != 1 && AppInfo.VerifyKey != 3 { //1和3 可以换绑
		response.X响应状态消息(c, response.Status_操作失败, "应用禁止更换绑定信息.")
		return
	}
	局_Uid := 局_在线信息.Uid
	if !检测用户登录在线正常(&局_在线信息) {
		局_账号 := string(请求json.GetStringBytes("User"))
		局密码 := string(请求json.GetStringBytes("PassWord"))
		if 局_账号 == "" {
			response.X响应状态(c, response.Status_未登录)
			return
		} else {
			局_在线信息.User = 局_账号                                //如果出错,写日志时会用到
			if AppInfo.AppType == 3 || AppInfo.AppType == 4 { //是卡号
				局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_账号)
				if 局_Uid == 0 {
					response.X响应状态消息(c, response.Status_操作失败, "卡号不存在.")
					return
				}
			} else {
				局_User, ok := Ser_User.User取详情(局_账号)
				if !ok {
					response.X响应状态消息(c, response.Status_操作失败, "用户不存在.")
					return
				}
				if 局密码 == "" || !utils2.BcryptCheck(局密码, 局_User.PassWord) {
					go Ser_Log.Log_写登录日志(局_User.User, c.ClientIP(), "更换绑定登录时密码错误:"+局密码, AppInfo.AppId)
					response.X响应状态消息(c, response.Status_登录失败, "用户名或密码错误")
					return
				}
				局_Uid = 局_User.Id
			}
		}

	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.可能刚注册还没登录成功")
		return
	}
	if 局_AppUser.Key == "" {
		response.X响应状态消息(c, response.Status_操作失败, "无绑定信息,无需解除")
		return
	}

	err, 扣时间值 := 绑定信息更换规则校验(c, AppInfo, 局_Uid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	// 如果换绑需要扣点,就执行扣点
	if 扣时间值 > 0 {
		err = Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), false)
		if err != nil {
			response.X响应状态消息(c, response.Status_Vip已到期, "剩余会员时间或点数不足.")
			return
		} else {
			局_日志 := "用户解除绑定信息,旧绑定信息:" + 局_AppUser.Key
			局_type := 3
			if AppInfo.AppType == 2 || AppInfo.AppType == 4 {
				局_type = 2
			}
			Ser_Log.Log_写积分点数时间日志(局_在线信息.User, c.ClientIP(), 局_日志, D到数值(-扣时间值), AppInfo.AppId, 局_type)
		}
	}

	err = Ser_AppUser.Set绑定信息(AppInfo.AppId, 局_Uid, "")

	if err == nil {
		db := *global.GVA_DB
		_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
			Type:   constant.LogKey_解绑,
			User:   Ser_AppUser.Uid取User(AppInfo.AppId, 局_Uid),
			Uid:    局_Uid,
			AppId:  AppInfo.AppId,
			OldKey: 局_AppUser.Key,
			NewKey: "",
			Time:   time.Now().Unix(),
			Ip:     c.ClientIP(),
			Count:  D到数值(-扣时间值),
			Note:   "解除绑定信息",
		})
		if err != nil {
			global.GVA_LOG.Error("修改绑定信息日志写入失败:" + err.Error())
		}
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"ReduceVipTime": 扣时间值})
	} else {
		_ = Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), true) //退还已经扣除的点数
		// 暂时想不出什么情况会修改失败 概率较低
		response.X响应状态(c, response.Status_SQl错误)
	}

	return
}
func UserApi_置新用户消息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetUserMsg","MsgType":2,"Note":"内存写入错误错误信息:11191919;2424233"}
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	局_消息类型 := 请求json.GetInt("MsgType")
	if 局_消息类型 < 1 || 局_消息类型 == 4 {
		response.X响应状态消息(c, response.Status_操作失败, "消息类型不正确")
		return
	}
	局_消息内容 := string(请求json.GetStringBytes("Msg"))
	if 局_消息内容 == "" {
		response.X响应状态消息(c, response.Status_操作失败, "消息内容不能为空")
		return
	}
	go Ser_Log.Log_写用户消息(局_消息类型, 局_在线信息.User, AppInfo.AppName, 局_在线信息.AppVer, 局_消息内容, c.ClientIP())
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}

func UserApi_取验证码信息(c *gin.Context) {
	/*
	   var AppInfo DB.DB_AppInfo
	   var 局_在线信息 DB.DB_LinksToken
	   Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	*/
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetCaptcha","CaptchaType":2}

	if 请求json.GetInt("Captcha") == 2 {
		response.X响应状态消息(c, response.Status_操作失败, "滑动无需获取,直接置值验证即可")
		return
	}
	局_验证码id, 局_base64验证码内容, err := Captcha.Captcha_取英数验证码()
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "生成失败")
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"CaptchaType": 1, "CaptchaId": 局_验证码id, "CaptChaImg": 局_base64验证码内容})
	return
}

func UserApi_取短信验证码信息(c *gin.Context) {
	/*
	   var AppInfo DB.DB_AppInfo
	   var 局_在线信息 DB.DB_LinksToken
	   Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	*/
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPhoneCaptcha","Phone":"13188888888","User":"13188888888"}

	局_手机号 := string(请求json.GetStringBytes("Phone"))

	局_错误信息 := ""
	if 局_手机号 == "" {
		局_User, ok := Ser_User.User取详情(string(请求json.GetStringBytes("User")))
		if !ok {
			response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
			return
		}
		局_手机号 = 局_User.Phone
		if !utils2.Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.X响应状态消息(c, response.Status_参数错误, "用户绑定手机号格式不正确")
			return
		}
	} else {
		if !utils2.Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.X响应状态消息(c, response.Status_参数错误, 局_错误信息)
			return
		}
	}

	局_验证码 := W文本_取随机字符串_数字(6)
	局_验证码ID := "Note" + utils2.Md5String(局_手机号)[:16] + W文本_取随机字符串(15)

	err := Captcha.Sms_当前选择发送短信验证码([]string{局_验证码}, 局_手机号)
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("短信验证码发送失败:%v,%v,%v", 局_验证码, 局_手机号, err.Error()))
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}
	Captcha.H缓存验证码校验实例.Set(局_验证码ID, 局_验证码)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"CaptchaType": 3, "CaptchaId": 局_验证码ID})
	return
}

func UserApi_取用户绑定信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.")
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Key": 局_AppUser.Key})
	return
}

func UserApi_取系统时间戳(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Time": time.Now().Unix()})
	return
}
func UserApi_取软件用户信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	var 局_AppUser DB.DB_AppUser
	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)

	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetAppUserInfo","AppVer":"1.0.15"}
	// 飞鸟快验内部使用, 主要解决用户更新软件后,继承token,但是在线用户信息的版本号没有改变
	局_应用版本 := string(请求json.GetStringBytes("AppVer"))
	if 局_应用版本 != "" {
		Ser_LinkUser.Id更新当前版本号(局_在线信息.Id, 局_应用版本)
	}

	var 局_UserClass DB.DB_UserClass
	局_UserClass, _ = Ser_UserClass.Id取详情(AppInfo.AppId, 局_AppUser.UserClassId)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"Id":              局_AppUser.Id,
		"Uid":             局_AppUser.Uid,
		"User":            局_在线信息.User,
		"Key":             局_AppUser.Key,
		"VipTime":         局_AppUser.VipTime,
		"VipNumber":       局_AppUser.VipNumber,
		"Status":          局_AppUser.Status,
		"MaxOnline":       局_AppUser.MaxOnline,
		"AgentUid":        局_AppUser.AgentUid,
		"LoginTime":       局_在线信息.LoginTime,
		"LoginIp":         局_在线信息.Ip,
		"RegisterTime":    局_AppUser.RegisterTime,
		"UserClassId":     局_AppUser.UserClassId,
		"UserClassName":   局_UserClass.Name,
		"UserClassMark":   局_UserClass.Mark,
		"UserClassWeight": 局_UserClass.Weight,
	})

	return
}
func UserApi_取软件用户备注(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Note": 局_AppUser.Note})
	return
}
func UserApi_取Vip到期时间戳(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"VipTime": 局_AppUser.VipTime})
	return
}
func UserApi_用户登录注销(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	err := Ser_LinkUser.Set批量注销([]int{局_在线信息.Id}, Ser_LinkUser.Z注销_用户操作注销)
	更新上下文缓存在线信息(c)
	if err != nil {
		response.X响应状态(c, response.Status_操作失败)
	} else {
		response.X响应状态(c, c.GetInt("局_成功Status"))
	}
	return
}
func UserApi_心跳(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 1})
		return
	}

	if AppInfo.Status == 2 { //应用免费模式直接返回 会员状态1
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 1})
		return
	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	Status := 1                                       //1 正常  3 vip过期
	if AppInfo.AppType == 2 || AppInfo.AppType == 4 { //计点
		Status = S三元(局_AppUser.VipTime > 0, 1, 3) //'计点模式大于0'
	} else {
		Status = S三元(局_AppUser.VipTime > time.Now().Unix(), 1, 3) //账号模式大于当前时间戳
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": Status})
	return
}

func UserApi_取用户积分(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if 局_在线信息.Status != 1 || 局_在线信息.Uid == 0 {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取应用用户信息失败.")
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"VipNumber": 局_AppUser.VipNumber})
	return
}
func UserApi_取开启验证码接口(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), AppInfo.Captcha)
	return
}
func UserApi_用户登录远程注销(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"RemoteLogOut","User":"aaaaaa","PassWord":"ssssss","Token":"","Time":1684069624,"Status":27417}'
	局_id := 0

	if AppInfo.AppType == 1 || AppInfo.AppType == 2 {
		局_User, ok := Ser_User.User取详情(string(请求json.GetStringBytes("User")))
		if !ok {
			response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
			return
		}
		if !utils2.BcryptCheck(string(请求json.GetStringBytes("PassWord")), 局_User.PassWord) {
			response.X响应状态消息(c, response.Status_操作失败, "用户名或密码错误")
			return
		}
		局_id = 局_User.Id

	} else {
		局_id = Ser_Ka.Ka卡号取id(AppInfo.AppId, string(请求json.GetStringBytes("User")))
		if 局_id == 0 {
			response.X响应状态消息(c, response.Status_操作失败, "卡号不存在")
			return
		}
	}
	var err error
	var 局_指定token = string(请求json.GetStringBytes("Token"))
	if 局_指定token == "" {
		err = Ser_LinkUser.Set批量注销Uid(局_id, Ser_LinkUser.Z注销_用户远程注销)
	} else {
		var 局_临时在线信息 DB.DB_LinksToken
		局_临时在线信息, err = Ser_LinkUser.Token取User在线详情(局_指定token)
		if err != nil {
			response.X响应状态(c, response.Status_操作失败)
			return
		}
		if 局_临时在线信息.Uid != 局_id { //只允许注销已经登陆的token,并且uid是自己的
			response.X响应状态消息(c, response.Status_操作失败, "用户没有权限注销此token")
			return
		}
		err = Ser_LinkUser.Set批量注销([]int{局_临时在线信息.Id}, Ser_LinkUser.Z注销_用户远程注销)
	}

	更新上下文缓存在线信息(c)
	if err != nil {
		response.X响应状态(c, response.Status_操作失败)
	} else {
		response.X响应状态(c, c.GetInt("局_成功Status"))
	}
	return
}

func UserApi_取动态标签(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Tab": 局_在线信息.Tab})
	return
}
func UserApi_置动态标签(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	err := Ser_LinkUser.Set动态标签(局_在线信息.Id, string(请求json.GetStringBytes("Tab")))
	if err != nil {
		response.X响应状态(c, response.Status_操作失败)
		return
	}
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}

func UserApi_取用户类型列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	局_列表 := Ser_UserClass.UserClass取AppId用户类型列表(AppInfo.AppId)
	var 局_响应 []gin.H

	for _, 单列表 := range 局_列表 {
		局_响应 = append(局_响应, gin.H{"Name": 单列表.Name, "Mark": 单列表.Mark, "Weight": 单列表.Weight})
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局_响应)
	return
}

func UserApi_置用户类型(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	局_新用户类型, OK := Ser_UserClass.Mark取详情(AppInfo.AppId, 请求json.GetInt("Mark"))
	if !OK {
		response.X响应状态消息(c, response.Status_操作失败, "用户类型代号不存在")
		return
	}
	局_App用户, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
		return
	}

	局_旧用户类型, OK := Ser_UserClass.Mark取详情(AppInfo.AppId, 局_App用户.UserClassId)
	if !OK { //如果是没有的类型就赋值 未分类
		局_旧用户类型 = DB.DB_UserClass{AppId: AppInfo.AppId, Name: "未分类", Weight: 1}
	}

	if 局_旧用户类型.Mark == 局_新用户类型.Mark { //代号相同,直接转换即可
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"UserClassMark": 局_新用户类型.Mark, "UserClassName": 局_新用户类型.Name, "VipTime": 局_App用户.VipTime})
		return
	} else {
		局_现行时间戳 := time.Now().Unix()
		// 用户类型不同, 根据权重处理
		if AppInfo.AppType == 2 || AppInfo.AppType == 4 {
			局_增减时间点数 := 局_App用户.VipTime * 局_旧用户类型.Weight / 局_新用户类型.Weight //转换结果值
			局_App用户.VipTime = 局_增减时间点数
		} else {
			if 局_App用户.VipTime < 局_现行时间戳 {
				// 已经过期了直接赋值新类型 现行时间+新时间就可以了
				局_App用户.VipTime = 局_现行时间戳
			} else {
				局_App用户.VipTime = 局_App用户.VipTime - 局_现行时间戳                   //先计算还剩多长时间
				局_增减时间点数 := 局_App用户.VipTime * 局_旧用户类型.Weight / 局_新用户类型.Weight //剩余时间 权重转换转换结果值
				局_App用户.VipTime = 局_现行时间戳 + 局_增减时间点数                          // 现在时间 + 旧权重转换后的新权重时间+卡增减时间
			}
		}
		局_App用户.UserClassId = 局_新用户类型.Id //最后更换类型,防止前面用到卡类id,计算权重转换类型错误
	}
	err := Ser_AppUser.Ser用户类型Vip时间(AppInfo.AppId, 局_App用户.Uid, 局_App用户.UserClassId, 局_App用户.VipTime)

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "写入新用户类型和Vip失败")
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"UserClassMark": 局_新用户类型.Mark, "UserClassName": 局_新用户类型.Name, "VipTime": 局_App用户.VipTime})
	return
}

func UserApi_卡号充值(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"UseKa","User":"aaaaaa","Ka":"aaaaaa","InviteUser":"aaaaaa","Time":1684071722,"Status":41016}
	局_用户 := string(请求json.GetStringBytes("User"))
	if 局_用户 == "" && 局_在线信息.Uid > 0 { //如果获取不到就充值在线用户
		局_用户 = 局_在线信息.User
	}
	局_卡号 := strings.TrimSpace(string(请求json.GetStringBytes("Ka")))
	局_推荐人 := strings.TrimSpace(string(请求json.GetStringBytes("InviteUser")))
	err := ka.L_ka.K卡号充值_事务(c, AppInfo.AppId, 局_卡号, 局_用户, 局_推荐人)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"InviteUser": 局_推荐人 != ""})

	return
}

func UserApi_置代理标志(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	局_代理uid := 请求json.GetInt("AgentUid")
	局_推广码 := string(请求json.GetStringBytes("PromotionCode")) //如果有推广码 代理id失效
	if 局_推广码 != "" {
		tx := *global.GVA_DB
		局_临时, err := service.NewPromotionCode(c, &tx).Info2(map[string]interface{}{"PromotionCode": 局_推广码})
		if err == nil {
			局_代理uid = 局_临时.Id
		} else {
			response.X响应状态消息(c, response.Status_操作失败, "推广码错误")
			return
		}
	}

	if agentLevel.L_agentLevel.Q取Id代理级别(c, 局_代理uid) <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "AgentUid非代理Uid")
		return
	}

	err := Ser_LinkUser.Set代理标志(局_在线信息.Id, 局_代理uid)
	if err != nil {
		response.X响应状态(c, response.Status_操作失败)
		return
	}
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}

// 1.0.277+版本添加可用
func UserApi_取卡号详情(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetKaInfo","ka":"8987657"}
	kaInfo, err := Ser_Ka.Ka卡号取详情(string(请求json.GetStringBytes("Ka")))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "卡号不存在")
		return
	}
	if kaInfo.AppId != AppInfo.AppId {
		response.X响应状态消息(c, response.Status_操作失败, "非本应用卡号")
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"Name":         kaInfo.Name,
		"KaClassId":    kaInfo.KaClassId,
		"UserClassId":  kaInfo.UserClassId,
		"AppId":        kaInfo.AppId,
		"VipTime":      kaInfo.VipTime,
		"VipNumber":    kaInfo.VipNumber,
		"EndTime":      kaInfo.EndTime,
		"InviteCount":  kaInfo.InviteCount,
		"Id":           kaInfo.Id,
		"Num":          kaInfo.Num,
		"NumMax":       kaInfo.NumMax,
		"KaType":       kaInfo.KaType,
		"Money":        kaInfo.Money,
		"MaxOnline":    kaInfo.MaxOnline,
		"NoUserClass":  kaInfo.NoUserClass,
		"RMb":          kaInfo.RMb,
		"RegisterTime": kaInfo.RegisterTime,
		"Status":       kaInfo.Status,
	})
	return
}
