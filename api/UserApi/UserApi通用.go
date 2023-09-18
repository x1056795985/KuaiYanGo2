package UserApi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	. "github.com/duolabmeng6/goefun/eCore"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
	"server/Service/Captcha"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Js"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_Pay"
	"server/Service/Ser_PublicData"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_TaskPool"
	"server/Service/Ser_User"
	"server/Service/Ser_UserClass"
	"server/api/UserApi/response"
	"server/global"
	DB "server/structs/db"
	"server/utils"
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
	if AppInfo.CryptoType == 3 {
		局_通讯AES密钥 = c.GetString("局_CryptoKeyAes")
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

func UserApi_用户登录(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserPassLogin","UserOrKa":"aaaaaa","PassWord":"AF15D5FDACD5FDFEA300E88A8E253E82","Key":"677F23CB3FA0055B5FD03916D6AB3C9A","Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","AppVer":"1.0.1","Captcha":{"Id":"","Value":""}}
	if !版本号_检测可用(string(请求json.GetStringBytes("AppVer")), AppInfo.AppVer) {
		response.X响应状态(c, response.Status_版本不可用)
		return
	}
	var 局_Uid = 0
	var 局_卡 DB.DB_Ka
	var err error
	var 局_卡号或用户名 = string(请求json.GetStringBytes("UserOrKa"))
	if AppInfo.AppType == 3 || AppInfo.AppType == 4 {
		//卡号
		局_卡, err = Ser_Ka.Ka卡号取详情(局_卡号或用户名)
		if err != nil || 局_卡.AppId != AppInfo.AppId {
			response.X响应状态消息(c, response.Status_登录失败, "卡号不存在")
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

		if 局_User.PassWord == "" || !utils.BcryptCheck(string(请求json.GetStringBytes("PassWord")), 局_User.PassWord) {
			go Ser_Log.Log_写登录日志(局_User.User, c.ClientIP(), "密码错误:"+string(请求json.GetStringBytes("PassWord")), 局_在线信息.LoginAppid)
			response.X响应状态消息(c, response.Status_登录失败, "用户名或密码错误")
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

		//没有这个用户,应该是第一次登录应用,添加进去
		if AppInfo.AppType == 3 {
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), AppInfo.MaxOnline, time.Now().Unix()+局_卡.VipTime, 局_卡.VipNumber)
			_ = Ser_Ka.Ka修改已用次数加一([]int{局_Uid})

		} else if AppInfo.AppType == 4 {
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), AppInfo.MaxOnline, 局_卡.VipTime, 局_卡.VipNumber)
			_ = Ser_Ka.Ka修改已用次数加一([]int{局_Uid})
		} else {
			err = Ser_AppUser.New用户信息(AppInfo.AppId, 局_Uid, string(请求json.GetStringBytes("Key")), AppInfo.MaxOnline, time.Now().Unix(), 0)
		}

		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, 局_卡号或用户名, AppInfo.AppName, 局_在线信息.AppVer, "新添加软件用户时失败报错信息:"+err.Error(), c.ClientIP())
			response.X响应状态消息(c, response.Status_SQl错误, "New用户信息内部错误")
			return
		}
		// 注册送卡  只有 账号模式才使用
		if AppInfo.RegisterGiveKaClassId > 0 && (AppInfo.AppType == 1 || AppInfo.AppType == 2) {
			局_注册送卡, 局_制卡结果 := Ser_Ka.Ka单卡创建(AppInfo.RegisterGiveKaClassId, "系统自动", "用户注册系统自动制卡赠送充值", "", 0)
			if 局_制卡结果 == nil {
				_, _ = Ser_Ka.K卡号充值_事务(AppInfo.AppId, 局_注册送卡.Name, 局_卡号或用户名, "", c.ClientIP())
			}
		}

		局_AppUser, _ = Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid)
	}

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
			_ = Ser_LinkUser.Set批量注销(局_已经在线数量[:局_要踢掉的数量])
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
	go Ser_User.Id置最后登录AppId(局_Uid, AppInfo.AppId, c.ClientIP())
	Ser_LinkUser.Set在线登录信息(局_在线信息.Id, 局_Uid, 局_卡号或用户名, 局_AppUser.Key, string(请求json.GetStringBytes("Tab")), string(请求json.GetStringBytes("AppVer")))
	//登录成功写日志
	if 局_老用户 {
		go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "用户登录", 局_在线信息.LoginAppid)
	} else {
		go Ser_Log.Log_写登录日志(局_卡号或用户名, c.ClientIP(), "新用户登录注册", 局_在线信息.LoginAppid)
	}

	var 局_用户类型 DB.DB_UserClass
	局_用户类型, ok := Ser_UserClass.Id取详情(AppInfo.AppId, 局_AppUser.UserClassId)
	if !ok {
		局_用户类型.Name = "已删待改"
		局_用户类型.Mark = 0
	}

	//这里吧成功的状态
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"VipTime":       局_AppUser.VipTime,
		"Key":           局_AppUser.Key,
		"OutUser":       局_要踢掉的数量,
		"UserClassMark": 局_用户类型.Mark,
		"UserClassName": 局_用户类型.Name,
		"VipNumber":     局_AppUser.VipNumber,
		"LoginTime":     time.Now().Unix(),
		"LoginIp":       c.ClientIP(),
		"RegisterTime":  局_AppUser.RegisterTime,
	})
}
func UserApi_GetUserIP(c *gin.Context) {
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"IP": c.ClientIP()})
}

func UserApi_用户减少余额(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserReduceMoney","Money":1.01,"Log":"看你长得帅,减些钱","AgentId":10,"AgentMoney":0,"AgentMoneyLog":"代理分成"}

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

	go Ser_Log.Log_写余额日志(局_User.User, c.ClientIP(), fmt.Sprintf("%s|新余额%v", 请求json.GetStringBytes("Log"), 新余额), utils.Float64取负值(局_增减值))
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Money": 局_User.Rmb})

	//用户减少成功,开始判断代理增加  不需要让用户知道,代理是否有分成,所以上面直接返回就行
	局_代理用户Id := 请求json.GetInt("AgentId")
	局_代理分成金额Id := 请求json.GetFloat64("AgentMoney")
	if 局_代理用户Id == 0 || 局_代理分成金额Id <= 0 {
		return
	}

	if 局_代理分成金额Id > 请求json.GetFloat64("Money") { //判断代理分成金额是否比用户减少余额还高
		go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, 局_User.User, c.ClientIP(), fmt.Sprintf("代理%v分成值(%v)大于用户减少余额(%v),可能非法用户尝试篡改应用数据", 局_代理用户Id, 局_代理分成金额Id, 请求json.GetFloat64("Money")))
		return
	}

	if 局_代理用户Id < 0 {
		局_代理用户Id = -局_代理用户Id
		局_管理员信息, ok2 := Ser_Admin.Id取详情(局_代理用户Id)
		if ok2 == false {
			//管理不存在就不用风控记录了,没什么用
			return
		}
		管理员新余额, 管理员err := Ser_Admin.Id余额增减(局_管理员信息.Id, 局_代理分成金额Id, true)
		if 管理员err == nil {
			go Ser_Log.Log_写余额日志(局_管理员信息.User, c.ClientIP(), fmt.Sprintf("%s|新余额%v", 请求json.GetStringBytes("AgentMoneyLog"), 管理员新余额), 局_代理分成金额Id)
		} else {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", AppInfo.AppName, 局_在线信息.AppVer, fmt.Sprintf("给管理员分成增加余额时失败失败:%v,%v", 管理员err.Error(), c.GetString("局_json明文")), c.ClientIP())
		}
	} else {
		局_代理信息, ok2 := Ser_User.Id取详情(局_代理用户Id)
		if ok2 == false {
			//用户不存在就不用风控记录了,没什么用
			return
		}
		if 局_代理信息.UPAgentId == 0 { //检测代理是否为普通用户,没有上级代理必然是普通用户,如果是普通用户触发风控
			//如果想检测是否为1级代理,可以改成 局_代理信息.UPAgentId >= 0  1级代理的上级代理,必然是管理员负数
			go Ser_Log.Log_写风控日志(局_在线信息.Id, Ser_Log.Log风控类型_Api异常调用, 局_User.User, c.ClientIP(), fmt.Sprintf("余额减少api,形参分成代理Id(%v)非代理id,可能非法用户尝试篡改应用数据", 局_代理用户Id))
		} else {
			代理新余额, 代理err := Ser_User.Id余额增减(局_代理信息.UPAgentId, 局_代理分成金额Id, true)
			if 代理err == nil {
				go Ser_Log.Log_写余额日志(局_代理信息.User, c.ClientIP(), fmt.Sprintf("%s|新余额%v", 请求json.GetStringBytes("AgentMoneyLog"), 代理新余额), 局_代理分成金额Id)
			} else {
				go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", AppInfo.AppName, 局_在线信息.AppVer, fmt.Sprintf("给代理分成增加余额时失败失败:%v,%v", 代理err.Error(), c.GetString("局_json明文")), c.ClientIP())
			}
		}
	}

}
func UserApi_用户减少点数(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserReduceMoney","VipTime":1.3,"Log":"看你长得帅,扣点钱"}
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
		response.X响应状态消息(c, response.Status_操作失败, "积分不足")
		return
	}

	err := Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, 局_增减值, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error()) //基本就是积分不足
		return
	}

	局_AppUser.VipTime -= 局_增减值
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"VipTime": 局_AppUser.VipTime})
	go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, c.ClientIP(), fmt.Sprintf("%s|剩余%v", 请求json.GetStringBytes("Log"), 局_AppUser.VipNumber), utils.Float64取负值(float64(局_增减值)), AppInfo.AppId, 2)
	return
}
func UserApi_用户减少积分(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"UserReduceMoney","VipNumber":1.3,"Log":"看你长得帅,扣点钱"}

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
	if 局_AppUser.VipNumber < 局_增减值 {
		response.X响应状态消息(c, response.Status_操作失败, "积分不足")
		return
	}

	err := Ser_AppUser.Id积分增减(AppInfo.AppId, 局_AppUser.Id, 局_增减值, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error()) //基本就是积分不足
		return
	}

	//flosat64 直接
	局_增减值D := decimal.NewFromFloat(局_增减值)
	局_用户积分D := decimal.NewFromFloat(局_AppUser.VipNumber)

	局_用户积分D = 局_用户积分D.Sub(局_增减值D)
	局_AppUser.VipNumber, _ = 局_用户积分D.Float64()

	局_增减值, _ = 局_增减值D.Mul(decimal.NewFromFloat(-1)).Float64() //乘-1 变成负数

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"VipNumber": 局_AppUser.VipNumber})
	go Ser_Log.Log_写积分点数时间日志(局_在线信息.User, c.ClientIP(), fmt.Sprintf("%s|剩余%v", 请求json.GetStringBytes("Log"), 局_AppUser.VipNumber), 局_增减值, AppInfo.AppId, 1)
	return
}

func UserApi_取服务器连接状态(c *gin.Context) {
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}

func UserApi_取登录状态(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppGongGao": AppInfo.AppGongGao})
	return
}
func UserApi_取新版本下载地址(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	if strings.Index(AppInfo.UrlDownload, "{{AppVer}}") != -1 && AppInfo.AppVer != "" {
		局_可用版本 := utils.W文本_分割文本(AppInfo.AppVer, "\n")
		if len(局_可用版本) > 0 {
			AppInfo.UrlDownload = strings.Replace(AppInfo.UrlDownload, "{{AppVer}}", 局_可用版本[0], -1)
		}

	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppUpDataJson": AppInfo.UrlDownload})
	return
}
func UserApi_取应用专属变量(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))
	局_云变量数据, err := Ser_PublicData.P取值2(AppInfo.AppId, 局_变量名)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := string(请求json.GetStringBytes("Name"))

	局_云变量数据, err := Ser_PublicData.P取值2(1, 局_变量名) //1>所以有软件公共读
	if err != nil || 局_云变量数据.Type > 3 {           //只允许变量  不允许读取云函数
		response.X响应状态消息(c, response.Status_操作失败, "变量不存在,请到后台应用编辑,添加专属变量")
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{局_变量名: 局_云变量数据.Value})
	//暂时公共变量直接响应数据 后期看情况
	/*	if 局_云变量数据.IsVip == 0 || 检测用户登录在线正常(&局_在线信息) {
			if 局_云变量数据.IsVip > 0 { //只有返回VIP变量时才强制
				c.Set("RSA强制", true)
			}

			response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{局_变量名: 局_云变量数据.Value})
		} else {
			response.X响应状态(c, response.Status_未登录)
		}*/
	return
}
func UserApi_取应用最新版本(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetAppVersion","Version":"1.3.5","IsVersionAll":true}
	局_可用版本 := utils.W文本_分割文本(AppInfo.AppVer, "\n")
	if len(局_可用版本) == 0 || AppInfo.AppVer == "" {
		response.X响应状态消息(c, response.Status_操作失败, "应用未设置版本号或格式不正确")
		return
	}

	局_分解版本号 := utils.W文本_分割文本(局_可用版本[0], ".")
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
		//只有大版本号
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"NewVersion": 局_可用版本[0], "Version": 局_分解版本号最新.大版本号, "IsUpdate": 局_是否更新})
		return
	} else {
		//有大小版本号
		局_小数运算, _ := decimal.NewFromString(局_分解版本号[0] + "." + 局_分解版本号[1])
		局_双精度版本, _ := 局_小数运算.Float64()
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"NewVersion": 局_可用版本[0], "Version": 局_双精度版本, "IsUpdate": 局_是否更新})
		return
	}
}
func UserApi_取应用主页Url(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppHomeUrl": AppInfo.UrlHome})
	return
}

// 1.0.42+版本添加可用
func UserApi_取应用基础信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{
		"AppId":   AppInfo.AppId,
		"AppType": AppInfo.AppType,
		"AppName": AppInfo.AppName,
		"AppWeb":  AppInfo.AppWeb,
	})
	return
}
func UserApi_置新绑定信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetAppUserKey","NewKey":"8987657"}
	//检查是否可以换换绑
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
				if 局密码 == "" || !utils.BcryptCheck(局密码, 局_User.PassWord) {
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
	//检查是否可以绑定相同信息
	if AppInfo.IsUserKeySame == 2 && Ser_AppUser.B绑定信息是否存在(AppInfo.AppId, 局_信息绑定信息) {
		response.X响应状态消息(c, response.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
		return
	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.可能刚注册还没登录成功")
		return
	}
	//如果换绑需要扣点,就执行扣点, 		且原来绑定信息不能为空
	if AppInfo.UpKeyData > 0 && 局_AppUser.Key != "" {
		err := Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(AppInfo.UpKeyData), false)
		if err != nil {
			response.X响应状态消息(c, response.Status_Vip已到期, "剩余会员时间或点数不足.")
			return
		}
	} else {
		AppInfo.UpKeyData = 0
	}
	err := Ser_AppUser.Set绑定信息(AppInfo.AppId, 局_Uid, 局_信息绑定信息)
	if err == nil {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"ReduceVipTime": AppInfo.UpKeyData})
	} else {
		_ = Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(AppInfo.UpKeyData), true) //退还已经扣除的点数
		response.X响应状态(c, response.Status_SQl错误)
	}

	return
}
func UserApi_解除绑定信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetAppUserKey"}
	//检查是否可以换换绑
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
				if 局密码 == "" || !utils.BcryptCheck(局密码, 局_User.PassWord) {
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
		response.X响应状态消息(c, response.Status_操作失败, "未绑定信息,无需解除")
		return
	}

	//如果换绑需要扣点,就执行扣点
	if AppInfo.UpKeyData > 0 {
		err := Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(AppInfo.UpKeyData), false)
		if err != nil {
			response.X响应状态消息(c, response.Status_Vip已到期, "剩余会员时间或点数不足.")
			return
		}
	}

	err := Ser_AppUser.Set绑定信息(AppInfo.AppId, 局_Uid, "")
	if err == nil {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"ReduceVipTime": AppInfo.UpKeyData})
	} else {
		_ = Ser_AppUser.Id点数增减(AppInfo.AppId, 局_AppUser.Id, int64(AppInfo.UpKeyData), true) //退还已经扣除的点数
		response.X响应状态(c, response.Status_SQl错误)
	}

	return
}
func UserApi_置新用户消息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"SetUserMsg","MsgType":2,"Note":"内存写入错误错误信息:11191919;2424233"}
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	局_消息类型 := 请求json.GetInt("MsgType")
	if 局_消息类型 < 1 || 局_消息类型 > 3 {
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
	/*	var AppInfo DB.DB_AppInfo
		var 局_在线信息 DB.DB_LinksToken
		用户数据信息还原(c, &AppInfo, &局_在线信息)*/
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
	/*	var AppInfo DB.DB_AppInfo
		var 局_在线信息 DB.DB_LinksToken
		用户数据信息还原(c, &AppInfo, &局_在线信息)*/
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
		if !utils.Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.X响应状态消息(c, response.Status_参数错误, "用户绑定手机号格式不正确")
			return
		}
	} else {
		if !utils.Z正则_校验手机号(局_手机号, &局_错误信息) {
			response.X响应状态消息(c, response.Status_参数错误, 局_错误信息)
			return
		}
	}

	局_验证码 := utils.W文本_取随机字符串_数字(6)
	局_验证码ID := "Note" + utils.Md5String(局_手机号)[:16] + utils.W文本_取随机字符串(15)

	err := Captcha.TX云_sms发送短信验证码([]string{局_验证码}, 局_手机号)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Time": time.Now().Unix()})
	return
}
func UserApi_取软件用户信息(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
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
	//{"Api":"GetAppUserInfo","AppVer":"1.0.15"}
	//飞鸟快验内部使用, 主要解决用户更新软件后,继承token,但是在线用户信息的版本号没有改变
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}
	err := Ser_LinkUser.Set批量注销([]int{局_在线信息.Id})
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	if AppInfo.Status == 2 {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 1})
		return
	}

	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_在线信息.Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	if 局_AppUser.VipTime > int64(time.Now().Unix()) {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 1})
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 3})
	}
	return
}

func UserApi_取用户积分(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

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
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), AppInfo.Captcha)
	return
}
func UserApi_用户登录远程注销(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"RemoteLogOut","User":"aaaaaa","PassWord":"ssssss","Time":1684069624,"Status":27417}'
	局_id := 0
	if AppInfo.AppType == 1 || AppInfo.AppType == 2 {
		局_User, ok := Ser_User.User取详情(string(请求json.GetStringBytes("User")))
		if !ok {
			response.X响应状态消息(c, response.Status_操作失败, "用户不存在")
			return
		}
		if !utils.BcryptCheck(string(请求json.GetStringBytes("PassWord")), 局_User.PassWord) {
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

	err := Ser_LinkUser.Set批量注销Uid(局_id)
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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Tab": 局_在线信息.Tab})
	return
}
func UserApi_置动态标签(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	err := Ser_LinkUser.Set动态标签(局_在线信息.Id, string(请求json.GetStringBytes("Tab")))
	if err != nil {
		response.X响应状态(c, response.Status_操作失败)
		return
	}
	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}
func UserApi_取支付通道状态(c *gin.Context) {
	局map := Ser_Pay.Pay_取支付通道状态()
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局map)
	return
}
func UserApi_取可购买卡类列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	var DB_KaClass []DB.DB_KaClass
	DB_KaClass = Ser_KaClass.KaClass取可购买卡类列表(AppInfo.AppId)

	var 卡类列表_简化 = make([]gin.H, 0, len(DB_KaClass))
	for 索引, _ := range DB_KaClass {
		卡类列表_简化 = append(卡类列表_简化, gin.H{"Id": DB_KaClass[索引].Id, "Name": DB_KaClass[索引].Name, "Money": DB_KaClass[索引].Money})
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 卡类列表_简化)
	return
}
func UserApi_取已购买充值卡列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	局_数量 := 10
	if 请求json.GetInt("Number") > 0 {
		局_数量 = 请求json.GetInt("Number")
	}
	卡类名称map := Ser_KaClass.KaClass取map列表Int(AppInfo.AppId)
	var DB_Ka []DB.DB_Ka
	DB_Ka, _ = Ser_Ka.Ka取已购卡列表(局_在线信息.User, 1, 局_数量)

	var 卡列表_简化 = make([]gin.H, len(DB_Ka), len(DB_Ka)+1)
	for 索引, _ := range DB_Ka {
		卡列表_简化[索引] = gin.H{
			"Id":           DB_Ka[索引].Id,
			"AppId":        DB_Ka[索引].AppId,
			"Name":         DB_Ka[索引].Name,
			"Money":        DB_Ka[索引].Money,
			"KaClassId":    DB_Ka[索引].KaClassId,
			"KaClassName":  卡类名称map[DB_Ka[索引].KaClassId],
			"Status":       DB_Ka[索引].Status,
			"Num":          DB_Ka[索引].Num,
			"NumMax":       DB_Ka[索引].NumMax,
			"RegisterTime": DB_Ka[索引].RegisterTime,
		}
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 卡列表_简化)
	return
}

func UserApi_取用户类型列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

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
	用户数据信息还原(c, &AppInfo, &局_在线信息)
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
		//用户类型不同, 根据权重处理
		if AppInfo.AppType == 2 || AppInfo.AppType == 4 {
			局_增减时间点数 := 局_App用户.VipTime * 局_旧用户类型.Weight / 局_新用户类型.Weight //转换结果值
			局_App用户.VipTime = 局_增减时间点数
		} else {
			if 局_App用户.VipTime < 局_现行时间戳 {
				//已经过期了直接赋值新类型 现行时间+新时间就可以了
				局_App用户.VipTime = 局_现行时间戳
			} else {
				局_App用户.VipTime = 局_App用户.VipTime - 局_现行时间戳                   //先计算还剩多长时间
				局_增减时间点数 := 局_App用户.VipTime * 局_旧用户类型.Weight / 局_新用户类型.Weight //剩余时间 权重转换转换结果值
				局_App用户.VipTime = 局_现行时间戳 + 局_增减时间点数                          // 现在时间 + 旧权重转换后的新权重时间+卡增减时间
			}
		}
		局_App用户.UserClassId = 局_新用户类型.Id //最后更换类型,防止前面用到卡类id,计算权重转换类型错误

	}
	err := Ser_AppUser.Ser用户类型Vip时间(AppInfo.AppId, 局_App用户.Id, 局_App用户.UserClassId, 局_App用户.VipTime)

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "写入新用户类型和Vip失败")
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"UserClassMark": 局_新用户类型.Mark, "UserClassName": 局_新用户类型.Name, "VipTime": 局_App用户.VipTime})
	return
}

func UserApi_云函数执行(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能JS函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能JS函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()

	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"RunJS","Parameter":"{'a':1}","JsName":"获取用户相关信息","IsGlobal":false,"Time":1684497856,"Status":30873}
	var 局_JSid = 0
	if 请求json.GetBool("IsGlobal") {
		局_JSid = Ser_PublicJs.Name取Id([]int{Ser_PublicJs.Js类型_公共函数}, string(请求json.GetStringBytes("JsName")))
	} else {
		局_JSid = Ser_PublicJs.Name取Id([]int{AppInfo.AppId}, string(请求json.GetStringBytes("JsName")))
	}
	if 局_JSid == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "JS公共函数不存在")
		return
	}
	局_耗时 := time.Now().UnixMilli()

	var 局_PublicJs DB.DB_PublicJs
	var err error
	局_PublicJs, err = Ser_PublicJs.Q取值2(局_JSid)

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "JS公共函数不存在")
		return
	}

	if E文件是否存在(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value) {
		局_PublicJs.Value = string(E读入文件(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value))
	} else {
		response.X响应状态消息(c, response.Status_操作失败, "js文件读取失败可能被删除")
		return
	}

	局_云函数型参数 := string(请求json.GetStringBytes("Parameter"))
	vm := Ser_Js.JS引擎初始化_用户(&AppInfo, &局_在线信息)
	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.X响应状态消息(c, response.Status_操作失败, "JS代码运行失败:"+局_详细错误.String())
		return
	}
	var 局_待执行js函数名 func(string) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		response.X响应状态消息(c, response.Status_操作失败, "Js中没有["+局_PublicJs.Name+"()]函数")
		return
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "Js绑定函数到变量失败")
		return
	}
	局_return := 局_待执行js函数名(局_云函数型参数)
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Return": 局_return, "Time": time.Now().UnixMilli() - 局_耗时})
	return
}
func UserApi_任务池_任务创建(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolNew","TaskTypeId":1,"Parameter":"{'a':1}","Time":1684752350,"Status":28986}

	局_任务类型, err := Ser_TaskPool.Task类型读取(请求json.GetInt("TaskTypeId"))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "任务类型Id不存在")
		return
	}

	if 局_任务类型.Status != 1 {
		response.X响应状态消息(c, response.Status_操作失败, "维护中")
		return
	}

	局_任务数据 := string(请求json.GetStringBytes("Parameter"))
	if 局_任务类型.HookSubmitDataStart != "" {
		局_任务数据, _, err = Ser_Js.JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 0)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}

	任务Id, err := Ser_TaskPool.Task数据创建加入队列(局_任务类型.Id, 局_任务数据)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "Task数据创建加入队列失败")
		return
	}

	if 局_任务类型.HookSubmitDataEnd != "" {
		局_任务数据, _, err = Ser_Js.JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 1)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"TaskUuid": 任务Id})
	return
}

func UserApi_任务池_任务查询(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolGetData","TaskUuid":"388f3cb1-ee27-4a5c-979d-a17cf3107dcd","Time":1684761030,"Status":12622}

	局_uuid := string(请求json.GetStringBytes("TaskUuid"))
	if len(局_uuid) != 36 { //提前筛选,优化
		response.X响应状态消息(c, response.Status_操作失败, "任务Uuid错误")
		return
	}
	局_任务数据, err := Ser_TaskPool.Task数据读取_单条(局_uuid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "任务Uuid错误")
		return
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"Status": 局_任务数据.Status, "ReturnData": 局_任务数据.ReturnData, "TimeStart": 局_任务数据.TimeStart, "TimeEnd": 局_任务数据.TimeEnd})
	return
}

func UserApi_任务池_任务处理获取(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolGetTask","GetTaskNumber":5,"GetTaskTypeId":[1],"Time":1684764215,"Status":18042}
	局_最大数量 := 请求json.GetInt("GetTaskNumber")
	局_临时 := 请求json.GetArray("GetTaskTypeId")
	var 局_可获取任务类型ID = make([]int, len(局_临时))
	for 索引, _ := range 局_临时 {
		局_可获取任务类型ID[索引], _ = 局_临时[索引].Int()
	}
	局_任务UUID := Ser_TaskPool.Task队列弹出任务(局_可获取任务类型ID, 局_最大数量)
	var 局_已获取任务数据 []DB.TaskPool_数据_精简
	if len(局_任务UUID) > 0 {
		局_已获取任务数据 = Ser_TaskPool.Task数据读取_数组(局_任务UUID)
	} else {
		局_已获取任务数据 = []DB.TaskPool_数据_精简{}
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局_已获取任务数据)
	return
}

type TaskPool_数据_精简 struct {
	Uuid string `json:"uuid" gorm:"column:uuid;size:36;primarykey;"`
	//LId        int    `json:"LId" gorm:"column:LId;comment:在线id,只允许相同的查询任务"` 直接用UUid,不可能重复的除了获取者别人也猜不到ID
	Tid        int    `json:"Tid" gorm:"column:Tid;comment:对应的任务类型Id"`
	TimeStart  int    `json:"TimeStart" gorm:"column:TimeStart;comment:任务创建时间戳"`
	TimeEnd    int    `json:"TimeEnd" gorm:"column:TimeEnd;comment:任务结束时间戳"`
	SubmitData string `json:"SubmitData" gorm:"column:SubmitData;comment:生产提交数据"`
	ReturnData string `json:"ReturnData" gorm:"column:ReturnData;comment:消费返回数据"`
	Status     int    `json:"Status" gorm:"column:Status;comment:任务状态,"` //1 已创建,2任务处理中,3成功,4任务失败
}

func UserApi_任务池_任务处理返回(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.X响应状态消息(c, response.Status_操作失败, "异常:可能Hook函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) { //强制登录才可以,不用检测ISVip了 必须登录
		response.X响应状态(c, response.Status_未登录)
		return
	}
	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"TaskPoolSetTask","TaskUuid":"f2e87ec0-4e0a-404d-a374-124d553a5a35","TaskStatus":40160,"TaskReturnData":"BB6CB5C68DF4652941CAF652A366F2D8","Time":1684769068}

	局_uuid := string(请求json.GetStringBytes("TaskUuid"))
	if len(局_uuid) != 36 { //提前筛选,优化
		response.X响应状态消息(c, response.Status_操作失败, "任务Uuid错误")
		return
	}

	局_Tid := Ser_TaskPool.Task数据读取Tid(局_uuid)

	局_任务类型, err := Ser_TaskPool.Task类型读取(局_Tid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "该UUID的任务类型Id不存在")
		return
	}
	局_任务数据 := string(请求json.GetStringBytes("TaskReturnData"))
	局_任务状态 := 请求json.GetInt("TaskStatus")
	if 局_任务类型.HookReturnDataStart != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataStart, 局_任务数据, 局_任务状态)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}

	err = Ser_TaskPool.Task数据修改(局_uuid, 局_任务状态, 局_任务数据)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "任务数据写入数据库失败")
		return
	}

	if 局_任务类型.HookReturnDataEnd != "" {
		局_任务数据, 局_任务状态, err = Ser_Js.JS引擎初始化_Hook处理(&AppInfo, &局_在线信息, 局_任务类型.HookSubmitDataEnd, 局_任务数据, 局_任务状态)
		if err != nil {
			response.X响应状态消息(c, response.Status_操作失败, err.Error())
			return
		}
	}

	response.X响应状态(c, c.GetInt("局_成功Status"))
	return
}

func UserApi_卡号充值(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"UseKa","User":"aaaaaa","Ka":"aaaaaa","InviteUser":"aaaaaa","Time":1684071722,"Status":41016}
	局_卡号 := string(请求json.GetStringBytes("Ka"))
	局_推荐人 := string(请求json.GetStringBytes("InviteUser"))
	err用户, err推荐人 := Ser_Ka.K卡号充值_事务(AppInfo.AppId, 局_卡号, string(请求json.GetStringBytes("User")), 局_推荐人, c.ClientIP())
	if err用户 != nil {
		response.X响应状态消息(c, response.Status_操作失败, err用户.Error())
		return
	}

	if 局_推荐人 == "" {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"InviteUser": false})
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"InviteUser": err推荐人 == nil})
	}
	return
}
func UserApi_订单_取状态(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetPayOrderStatus","OrderId":"","Time":1684152719,"Status":15959}
	局_订单Id := string(请求json.GetStringBytes("OrderId"))
	if 局_订单Id == "" {
		response.X响应状态消息(c, response.Status_操作失败, "订单不存在")
		return
	}

	局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(局_订单Id)
	if !ok {
		//如果失败了,在判断是不是上传的第三方订单号
		局_订单详细信息, ok = Ser_RMBPayOrder.Order取订单详细_第三方订单(局_订单Id)
	}

	//可能存在未登录充值的情况,所以不检测在线了
	if !ok { //|| 局_订单详细信息.Uid != 局_在线信息.Uid
		response.X响应状态消息(c, response.Status_操作失败, "订单不存在")
	} else {
		局_响应 := gin.H{"Status": 局_订单详细信息.Status}
		if 局_卡号 := utils.W文本_取出中间文本(局_订单详细信息.Extra, `"Name":"`, `"`); 局_卡号 != "" {
			局_响应["KaName"] = 局_卡号
		}
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局_响应)
	}
	return
}

func UserApi_订单_购卡直冲(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	if AppInfo.AppId < 10000 {
		response.X响应状态消息(c, response.Status_操作失败, "应用不存在")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetAliPayPC","User":"aaaaaa","KaClassId":1,"PayType":"小叮当","Time":1684152719,"Status":15959}

	局_用户名 := string(请求json.GetStringBytes("User"))
	局_卡号 := Ser_AppInfo.App是否为卡号(AppInfo.AppId)
	var 局_Uid = 0
	var 局_Uid类型 = 0
	var 局_AppUserID = 0

	if 局_卡号 {
		局_Uid类型 = 2
		局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_用户名)
		局_AppUserID = Ser_AppUser.Uid取Id(AppInfo.AppId, 局_Uid)
	} else {
		局_Uid类型 = 1
		局_Uid = Ser_User.User用户名取id(局_用户名)
		局_AppUserID = Ser_AppUser.Uid取Id(AppInfo.AppId, 局_Uid)
	}

	if 局_Uid == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户不存在")
		return
	}

	if 局_AppUserID == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户未登录过应用,请先操作登录一次")
		return
	}

	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求json.GetInt("KaClassId"))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "卡类不存在")
		return
	}
	if AppInfo.AppId != 局_卡类信息.AppId {
		response.X响应状态消息(c, response.Status_操作失败, "非本应用卡类")
		return
	}
	if 局_卡类信息.Money <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "该卡类用户价格小于0不可购买")
		return
	}

	局_额外数据 := fmt.Sprintf(`{"KaClassId":%d,"AppUserId":%d}`, 局_卡类信息.Id, 局_AppUserID)

	var 响应数据 gin.H
	局_支付方式 := strings.TrimSpace(string(请求json.GetStringBytes("PayType")))
	if 局_支付方式 == "" {
		response.X响应状态消息(c, response.Status_操作失败, "支付方式不能为空")
		return
	}
	//修改支付显示别名为原名称
	局_支付方式 = Ser_Pay.Pay_显示名称转原名(局_支付方式)

	fmt.Printf(局_支付方式)
	switch strings.TrimSpace(局_支付方式) {
	case "支付宝PC":
		err, 响应数据 = Ser_Pay.Pay_支付宝Pc_订单创建(局_Uid, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), 1, 局_额外数据)
	case "支付宝当面付":
		err, 响应数据 = Ser_Pay.Pay_支付宝当面付_订单创建(局_Uid, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), 1, 局_额外数据)
	case "微信支付":
		err, 响应数据 = Ser_Pay.Pay_微信Pc_订单创建(局_Uid, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), 1, 局_额外数据)
	case "小叮当":
		err, 响应数据 = Ser_Pay.Pay_小叮当_订单创建(局_Uid, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), 1, 局_额外数据)
	default:
		err = errors.New("充值方式[" + 局_支付方式 + "]不存在")
	}

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 响应数据)
	}
	return
}

func UserApi_订单_积分充值(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	if AppInfo.AppId < 10000 {
		response.X响应状态消息(c, response.Status_操作失败, "应用不存在")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetAliPayPC","User":"aaaaaa","Money":0.01,"PayType":"小叮当","Time":1684152719,"Status":15959}

	局_用户名 := string(请求json.GetStringBytes("User"))
	局_卡号 := Ser_AppInfo.App是否为卡号(AppInfo.AppId)
	var 局_Uid = 0
	var 局_Uid类型 = 0
	var 局_AppUserID = 0

	if 局_卡号 {
		局_Uid类型 = 2
		局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_用户名)
		局_AppUserID = Ser_AppUser.Uid取Id(AppInfo.AppId, 局_Uid)
	} else {
		局_Uid类型 = 1
		局_Uid = Ser_User.User用户名取id(局_用户名)
		局_AppUserID = Ser_AppUser.Uid取Id(AppInfo.AppId, 局_Uid)
	}

	if 局_Uid == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户不存在")
		return
	}

	if 局_AppUserID == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户未登录过应用,请先操作登录一次")
		return
	}

	局_额外数据 := fmt.Sprintf(`{"AppID":%d,"AppUserId":%d}`, AppInfo.AppId, 局_AppUserID)
	var err error
	var 响应数据 gin.H
	局_支付方式 := strings.TrimSpace(string(请求json.GetStringBytes("PayType")))
	if 局_支付方式 == "" {
		response.X响应状态消息(c, response.Status_操作失败, "支付方式不能为空")
		return
	}
	//修改支付显示别名为原名称
	局_支付方式 = Ser_Pay.Pay_显示名称转原名(局_支付方式)

	fmt.Printf(局_支付方式)
	switch strings.TrimSpace(局_支付方式) {
	case "支付宝PC":
		err, 响应数据 = Ser_Pay.Pay_支付宝Pc_订单创建(局_Uid, 局_Uid类型, 请求json.GetFloat64("Money"), c.ClientIP(), 2, 局_额外数据)
	case "支付宝当面付":
		err, 响应数据 = Ser_Pay.Pay_支付宝当面付_订单创建(局_Uid, 局_Uid类型, 请求json.GetFloat64("Money"), c.ClientIP(), 2, 局_额外数据)
	case "微信支付":
		err, 响应数据 = Ser_Pay.Pay_微信Pc_订单创建(局_Uid, 局_Uid类型, 请求json.GetFloat64("Money"), c.ClientIP(), 2, 局_额外数据)
	case "小叮当":
		err, 响应数据 = Ser_Pay.Pay_小叮当_订单创建(局_Uid, 局_Uid类型, 请求json.GetFloat64("Money"), c.ClientIP(), 2, 局_额外数据)
	default:
		err = errors.New("充值方式[" + 局_支付方式 + "]不存在")
	}

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "充值方式["+string(请求json.GetStringBytes("PayType"))+"]"+err.Error())
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 响应数据)
	}
	return
}
func UserApi_订单_支付购卡(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	用户数据信息还原(c, &AppInfo, &局_在线信息)

	if AppInfo.AppId < 10000 {
		response.X响应状态消息(c, response.Status_操作失败, "应用不存在")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"PayGetKa",,"KaClassId":1,"PayType":"小叮当","Time":1684152719,"Status":15959}

	var 局_Uid类型 = 0
	if Ser_AppInfo.App是否为卡号(AppInfo.AppId) {
		局_Uid类型 = 2
	} else {
		局_Uid类型 = 1
	}

	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求json.GetInt("KaClassId"))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "卡类不存在")
		return
	}
	if AppInfo.AppId != 局_卡类信息.AppId {
		response.X响应状态消息(c, response.Status_操作失败, "非本应用卡类")
		return
	}
	if 局_卡类信息.Money <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "该卡类用户价格小于0不可购买")
		return
	}

	局_额外数据 := fmt.Sprintf(`{"KaClassId":%d,"Ip":"%s"}`, 局_卡类信息.Id, c.ClientIP())

	var 响应数据 gin.H
	局_支付方式 := strings.TrimSpace(string(请求json.GetStringBytes("PayType")))
	if 局_支付方式 == "" {
		response.X响应状态消息(c, response.Status_操作失败, "支付方式不能为空")
		return
	}
	//修改支付显示别名为原名称
	局_支付方式 = Ser_Pay.Pay_显示名称转原名(局_支付方式)

	fmt.Printf(局_支付方式)
	switch strings.TrimSpace(局_支付方式) {
	case "支付宝PC":
		err, 响应数据 = Ser_Pay.Pay_支付宝Pc_订单创建(0, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), Ser_Pay.D订单_处理类型_支付购卡, 局_额外数据)
	case "支付宝当面付":
		err, 响应数据 = Ser_Pay.Pay_支付宝当面付_订单创建(0, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), Ser_Pay.D订单_处理类型_支付购卡, 局_额外数据)
	case "微信支付":
		err, 响应数据 = Ser_Pay.Pay_微信Pc_订单创建(0, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), Ser_Pay.D订单_处理类型_支付购卡, 局_额外数据)
	case "小叮当":
		err, 响应数据 = Ser_Pay.Pay_小叮当_订单创建(0, 局_Uid类型, 局_卡类信息.Money, c.ClientIP(), Ser_Pay.D订单_处理类型_支付购卡, 局_额外数据)
	default:
		err = errors.New("充值方式[" + 局_支付方式 + "]不存在")
	}

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 响应数据)
	}
	return
}
