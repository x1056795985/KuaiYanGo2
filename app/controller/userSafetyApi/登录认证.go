package userSafetyApi

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/logic/common/blacklist"
	"server/app/logic/common/ka"
	"server/app/logic/common/log"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	utils2 "server/app/utils"
	"strconv"
	"strings"
	"time"
)

// UserApi_用户登录 登录
func UserApi_用户登录(c *gin.Context) {
	局_ctx := 取上下文(c)
	//{"Api":"UserPassLogin","UserOrKa":"aaaaaa","PassWord":"AF15D5FDACD5FDFEA300E88A8E253E82","Key":"677F23CB3FA0055B5FD03916D6AB3C9A","Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","AppVer":"1.0.1","Captcha":{"Id":"","Value":""}}
	if !版本号_检测可用(局_ctx.Q请求明文.Get("AppVer").String(), 局_ctx.AppInfo.AppVer) {
		response.Fail(c, constant.Status_版本不可用)
		return
	}
	var 局_Uid = 0
	var 局_卡 dbm.DB_Ka
	var err error
	if len(局_ctx.Q请求明文.Get("Key").String()) > 191 {
		response.FailMsg(c, constant.Status_操作失败, "绑定信息长度不能超过191")
		return
	}
	if blacklist.Is黑名单(局_ctx.Q请求明文.Get("Key").String(), 局_ctx.AppInfo.AppId) {
		response.FailMsg(c, constant.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}
	if 局_ctx.Z在线信息.Uid != 0 {
		response.FailMsg(c, constant.Status_操作失败, "已登陆,无需重复登陆")
		return
	}

	var 局_卡号或用户名 = strings.TrimSpace(局_ctx.Q请求明文.Get("UserOrKa").String())
	db := *global.GVA_DB
	if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 {
		//卡号
		局_卡, err = service.NewKa(c, &db).InfoKa(局_卡号或用户名)
		if err != nil || 局_卡.AppId != 局_ctx.AppInfo.AppId {
			response.FailMsg(c, constant.Status_登录失败, "卡号不存在")
			return
		}
		if 局_卡.Status != 1 {
			go log.L_log.S输出日志(c, dbm.DB_LogLogin{
				User:      局_卡.Name,
				Ip:        c.ClientIP(),
				Note:      "卡号已冻结",
				LoginType: 局_ctx.Z在线信息.LoginAppid,
				Time:      time.Now().Unix(),
			})
			response.FailMsg(c, constant.Status_登录失败, "卡号已冻结")
			return
		}
		局_Uid = 局_卡.Id
		局_卡号或用户名 = 局_卡.Name
	} else {
		//账号
		var 局_User dbm.DB_User
		局_User, err = service.NewUser(c, &db).InfoName(局_卡号或用户名)
		if err != nil {
			response.FailMsg(c, constant.Status_登录失败, "用户不存在")
			return
		}

		if 局_User.PassWord == "" || !utils2.BcryptCheck(局_ctx.Q请求明文.Get("PassWord").String(), 局_User.PassWord) {
			go log.L_log.S输出日志(c, dbm.DB_LogLogin{
				User:      局_User.User,
				Ip:        c.ClientIP(),
				Note:      "密码错误:" + 局_ctx.Q请求明文.Get("PassWord").String(),
				LoginType: 局_ctx.Z在线信息.LoginAppid,
				Time:      time.Now().Unix(),
			})
			response.FailMsg(c, constant.Status_登录失败, "用户名或密码错误")
			return
		}
		if 局_User.Status != 1 {
			go log.L_log.S输出日志(c, dbm.DB_LogLogin{
				User:      局_User.User,
				Ip:        c.ClientIP(),
				Note:      "账号已冻结",
				LoginType: 局_ctx.Z在线信息.LoginAppid,
				Time:      time.Now().Unix(),
			})
			response.FailMsg(c, constant.Status_登录失败, "账号已冻结")
			return
		}
		if 局_User.UPAgentId != 0 {
			go log.L_log.S输出日志(c, dbm.DB_LogLogin{
				User:      局_User.User,
				Ip:        c.ClientIP(),
				Note:      "代理商请登录代理平台",
				LoginType: 局_ctx.Z在线信息.LoginAppid,
				Time:      time.Now().Unix(),
			})
			response.FailMsg(c, constant.Status_登录失败, "代理商请登录代理平台")
			return
		}
		局_Uid = 局_User.Id
		局_卡号或用户名 = 局_User.User
	}
	var 局_AppUser dbm.DB_AppUser
	局_老用户 := false
	局_AppUser, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_Uid)
	if err == nil {
		局_老用户 = true
	}
	if 局_老用户 {
		//如果用户key是空的直接重新绑定

		if 局_AppUser.Key == "" {
			//检查是否可以绑定相同信息
			if 局_ctx.AppInfo.IsUserKeySame == 2 {
				_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoKey(局_ctx.Q请求明文.Get("Key").String())
				if err == nil {
					response.FailMsg(c, constant.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
					return
				}
			}

			_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).UpdateUid(局_Uid, map[string]interface{}{"Key": 局_ctx.Q请求明文.Get("Key").String()})
			if err != nil {
				global.GVA_LOG.Println("设置绑定信息失败:" + err.Error())
			}

			_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
				Type:   constant.LogKey_绑定,
				User:   局_卡号或用户名,
				Uid:    局_Uid,
				AppId:  局_ctx.AppInfo.AppId,
				OldKey: 局_AppUser.Key,
				NewKey: 局_ctx.Q请求明文.Get("Key").String(),
				Time:   time.Now().Unix(),
				Ip:     c.ClientIP(),
				Note:   "无绑定信息登陆自动绑定",
			})
			if err != nil {
				global.GVA_LOG.Println("修改绑定信息日志写入失败:" + err.Error())
			}
			局_AppUser.Key = 局_ctx.Q请求明文.Get("Key").String()
		}

		//老用户验证绑定信息是否相同
		if 局_ctx.AppInfo.VerifyKey == 3 || 局_ctx.AppInfo.VerifyKey == 4 {
			//1 免验证可以换绑 2 免验证禁止换绑 3 验证可以换绑 4 验证禁止换
			if 局_AppUser.Key != 局_ctx.Q请求明文.Get("Key").String() {
				go log.L_log.S输出日志(c, dbm.DB_LogLogin{
					User:      局_卡号或用户名,
					Ip:        c.ClientIP(),
					Note:      "登录绑定信息验证失败:" + 局_ctx.Q请求明文.Get("Key").String(),
					LoginType: 局_ctx.Z在线信息.LoginAppid,
					Time:      time.Now().Unix(),
				})
				response.Fail(c, constant.Status_绑定信息验证失败)
				return
			}
		}

	} else {

		//新用户验证绑定信息是否存在
		if 局_ctx.AppInfo.IsUserKeySame == 2 {
			//1 免验证可以换绑 2 免验证禁止换绑 3 验证可以换绑 4 验证禁止换
			_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoKey(局_ctx.Q请求明文.Get("Key").String())
			if err == nil {
				go log.L_log.S输出日志(c, dbm.DB_LogLogin{
					User:      局_卡号或用户名,
					Ip:        c.ClientIP(),
					Note:      "登录注册绑定信息已存在:" + 局_ctx.Q请求明文.Get("Key").String(),
					LoginType: 局_ctx.Z在线信息.LoginAppid,
					Time:      time.Now().Unix(),
				})
				response.Fail(c, constant.Status_绑定信息已被其他用户使用)
				return
			}
		}

		if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 {
			if 局_卡.Num >= 局_卡.NumMax {
				response.FailMsg(c, constant.Status_登录失败, "卡号已经使用到最大次数")
				return
			}
		}

		//没有这个用户,应该是第一次登录应用,添加进去
		var 局_新AppUser dbm.DB_AppUser
		局_新AppUser.Id = 0
		局_新AppUser.Uid = 局_Uid
		局_新AppUser.Status = 1
		局_新AppUser.Key = 局_ctx.Q请求明文.Get("Key").String()
		局_新AppUser.RegisterTime = time.Now().Unix()
		局_新AppUser.AgentUid = 0 //不在这里赋值,单独处理

		switch 局_ctx.AppInfo.AppType {
		case 1:
			局_新AppUser.MaxOnline = 1
			局_新AppUser.VipTime = time.Now().Unix()
			局_新AppUser.VipNumber = 0
			局_新AppUser.UserClassId = 0
			局_新AppUser.Note = ""
		case 2: //账号限时
			局_新AppUser.MaxOnline = 1
			局_新AppUser.VipTime = 0
			局_新AppUser.VipNumber = 0
			局_新AppUser.UserClassId = 0
			局_新AppUser.Note = ""
		case 3:
			//卡号模式,制卡人就是归属代理 如果是管理员制造的卡, 就使用代理标志为归属uid
			局_新AppUser.MaxOnline = S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline)
			局_新AppUser.VipTime = time.Now().Unix() + 局_卡.VipTime
			局_新AppUser.VipNumber = 局_卡.VipNumber
			局_新AppUser.UserClassId = 局_卡.UserClassId
			局_新AppUser.Note = 局_卡.AdminNote
			// 卡号已用次数+1
			go service.NewKa(c, &db).Update(局_Uid, map[string]interface{}{"UsedCount": 1})
		case 4:
			//卡号模式,制卡人就是归属代理
			局_新AppUser.MaxOnline = S三元(局_卡.MaxOnline == 0, 1, 局_卡.MaxOnline)
			局_新AppUser.VipTime = 局_卡.VipTime
			局_新AppUser.VipNumber = 局_卡.VipNumber
			局_新AppUser.UserClassId = 局_卡.UserClassId
			局_新AppUser.Note = 局_卡.AdminNote
			// 卡号已用次数+1
			go service.NewKa(c, &db).Update(局_Uid, map[string]interface{}{"UsedCount": 1})
		default:
			//???应该不会到这里
			response.FailMsg(c, constant.Status_SQl错误, "AppInfo.AppType错误")
		}

		_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).Create(&局_新AppUser)
		if err != nil {
			go log.L_log.S输出日志(c, dbm.DB_LogUserMsg{
				User:    局_卡号或用户名,
				App:     局_ctx.AppInfo.AppName,
				AppId:   局_ctx.AppInfo.AppId,
				AppVer:  局_ctx.Z在线信息.AppVer,
				MsgType: constant.LogKey_绑定, //系统执行错误
				Note:    "新添加软件用户时失败报错信息:" + err.Error(),
				Ip:      c.ClientIP(),
			})
			response.FailMsg(c, constant.Status_SQl错误, "New用户信息内部错误")
			return
		}
		局_归属代理uid := 0
		if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 {
			//账号模式,制卡人就是归属代理 如果是管理员制造的卡, 就使用代理标志为归属uid
			局_归属代理User, err := service.NewUser(c, &db).InfoName(局_卡.RegisterUser)
			if err == nil {
				局_归属代理uid = 局_归属代理User.Id
			}
			if 局_归属代理uid == 0 {
				局_归属代理uid = 局_ctx.Z在线信息.AgentUid
			}
		} else {
			局_归属代理uid = 局_ctx.Z在线信息.AgentUid
		}
		ka.L_ka.Z置归属代理(c, 局_ctx.AppInfo.AppId, 局_Uid, S三元(局_ctx.AppInfo.AppType <= 2, 局_ctx.Z在线信息.AgentUid, 局_归属代理uid)) //失败也不影响
		_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
			Type:   constant.LogKey_绑定,
			User:   局_卡号或用户名,
			Uid:    局_Uid,
			AppId:  局_ctx.AppInfo.AppId,
			OldKey: "",
			NewKey: 局_ctx.Q请求明文.Get("Key").String(),
			Time:   time.Now().Unix(),
			Ip:     c.ClientIP(),
			Note:   "新用户登陆自动绑定",
		})
		if err != nil {
			global.GVA_LOG.Println("修改绑定信息日志写入失败:" + err.Error())
		}

		// 注册送卡  只有 账号模式才使用
		if 局_ctx.AppInfo.RegisterGiveKaClassId > 0 && (局_ctx.AppInfo.AppType == 1 || 局_ctx.AppInfo.AppType == 2) {
			_ = ka.L_ka.K卡类直冲_事务(c, 局_ctx.AppInfo.RegisterGiveKaClassId, 局_Uid)
			//局_注册送卡, 局_制卡结果 := Ser_Ka.Ka单卡创建(AppInfo.RegisterGiveKaClassId, "系统自动", "用户注册系统自动制卡赠送充值", "", 0)
			//if 局_制卡结果 == nil {
			//	_ = ka.L_ka.K卡号充值_事务(c, AppInfo.AppId, 局_注册送卡.Name, 局_卡号或用户名, "")
			//}
		}

	}

	局_AppUser, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_Uid) //充值之后重新读取一遍
	if err != nil {
		response.FailMsg(c, constant.Status_SQl错误, "获取用户信息失败:"+err.Error())
		return
	}
	if 局_AppUser.Status == 2 {
		go log.L_log.S输出日志(c, dbm.DB_LogLogin{
			User:      局_卡号或用户名,
			Ip:        c.ClientIP(),
			Note:      "已冻结无法登录",
			LoginType: 局_ctx.Z在线信息.LoginAppid,
			Time:      time.Now().Unix(),
		})
		response.Fail(c, constant.Status_已冻结无法登录)
		return
	}

	if 局_ctx.AppInfo.Status == 2 {
		//免费运营模式不检查时间直接登录成功
	} else {
		if 局_ctx.AppInfo.AppType == 2 || 局_ctx.AppInfo.AppType == 4 { //计点方式
			if 局_AppUser.VipTime <= 0 {
				go log.L_log.S输出日志(c, dbm.DB_LogLogin{
					User:      局_卡号或用户名,
					Ip:        c.ClientIP(),
					Note:      "非Vip禁止登录",
					LoginType: 局_ctx.Z在线信息.LoginAppid,
					Time:      time.Now().Unix(),
				})
				response.Fail(c, constant.Status_Vip已到期)
				return
			}
		} else { //计时模式
			if 局_AppUser.VipTime <= time.Now().Unix() { // 相等也限制登录, 防止刚注册 时间和过期正好相当
				go log.L_log.S输出日志(c, dbm.DB_LogLogin{
					User:      局_卡号或用户名,
					Ip:        c.ClientIP(),
					Note:      "Vip已过期",
					LoginType: 局_ctx.Z在线信息.LoginAppid,
					Time:      time.Now().Unix(),
				})
				response.Fail(c, constant.Status_Vip已到期)
				return
			}
		}
	}

	// 获取在线数量 - 单表查询LinksToken
	局_已经在线数量, err := service.NewLinksToken(c, &db).InfosId排序(map[string]interface{}{
		"Uid":        局_AppUser.Uid,
		"Status":     1,
		"LoginAppid": 局_ctx.AppInfo.AppId,
	})
	var 局_在线Id列表 []int
	if err == nil {
		for _, v := range 局_已经在线数量 {
			局_在线Id列表 = append(局_在线Id列表, v.Id)
		}
	}

	var 局_要踢掉的数量 = 0
	if len(局_在线Id列表) >= 局_AppUser.MaxOnline {
		if 局_ctx.AppInfo.ExceedMaxOnlineOut == 1 {
			//踢掉最早在线
			局_要踢掉的数量 = len(局_在线Id列表) - 局_AppUser.MaxOnline + 1
			_, err = service.NewLinksToken(c, &db).Updates(局_在线Id列表[:局_要踢掉的数量], map[string]interface{}{
				"OutTime":    0,
				"Status":     2,
				"LogoutCode": 1, //超过同时在线注销
			})
			//已经登录的数量-最大数量 +1
			go log.L_log.S输出日志(c, dbm.DB_LogLogin{
				User:      局_卡号或用户名,
				Ip:        c.ClientIP(),
				Note:      "登录同时在线超过最大值已注销最早登录:" + strconv.Itoa(局_要踢掉的数量),
				LoginType: 局_ctx.Z在线信息.LoginAppid,
				Time:      time.Now().Unix(),
			})

		} else if 局_ctx.AppInfo.ExceedMaxOnlineOut == 2 {
			//直接提示
			go log.L_log.S输出日志(c, dbm.DB_LogLogin{
				User:      局_卡号或用户名,
				Ip:        c.ClientIP(),
				Note:      "同时在线超过最大值",
				LoginType: 局_ctx.Z在线信息.LoginAppid,
				Time:      time.Now().Unix(),
			})
			response.Fail(c, constant.Status_同时在线超过最大值)
			return
		}

	}

	//登录成功吧数据写入在线信息内
	tx := *global.GVA_DB
	data := map[string]interface{}{
		"Uid":    局_Uid,
		"User":   局_卡号或用户名,
		"Key":    局_AppUser.Key,
		"Tab":    局_ctx.Q请求明文.Get("Tab").String(),
		"AppVer": 局_ctx.Q请求明文.Get("AppVer").String(),
	}
	_, err = service.NewLinksToken(c, &tx).Update(局_ctx.Z在线信息.Id, data)
	if err != nil {
		//mark 一个新奇的bug, Tab是ansi编码的中文, go字符串,类型为utf8 获取字节数组string转文本就会导致是乱码,导致修改数据库失败,看来得加参数校验了
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}
	//没有归属代理,但是在线信息已经有代理标志了 赋予软件用户归属代理
	if 局_AppUser.AgentUid == 0 && 局_ctx.Z在线信息.AgentUid != 0 {
		err = ka.L_ka.Z置归属代理(c, 局_ctx.Z在线信息.LoginAppid, 局_Uid, 局_ctx.Z在线信息.AgentUid) //失败也不影响
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, err.Error())
			return
		}
		局_AppUser.AgentUid = 局_ctx.Z在线信息.AgentUid
	}

	//用户已有归属代理,但是和在线信息代理标志不同,修改在线代理标志
	if 局_AppUser.AgentUid != 局_ctx.Z在线信息.AgentUid {
		_, err = service.NewLinksToken(c, &tx).Update(局_ctx.Z在线信息.Id, map[string]interface{}{"AgentUid": 局_AppUser.AgentUid})
		局_ctx.Z在线信息.AgentUid = 局_AppUser.AgentUid
	}

	//登录成功写日志
	_ = log.L_log.R日活月活增加_登陆处理(局_ctx.AppInfo.AppId, 局_卡号或用户名) //需要先处理日活,在写日志
	if 局_老用户 {
		go log.L_log.S输出日志(c, dbm.DB_LogLogin{
			User:      局_卡号或用户名,
			Ip:        c.ClientIP(),
			Note:      "用户登录",
			LoginType: 局_ctx.Z在线信息.LoginAppid,
			Time:      time.Now().Unix(),
		})
	} else {
		go log.L_log.S输出日志(c, dbm.DB_LogLogin{
			User:      局_卡号或用户名,
			Ip:        c.ClientIP(),
			Note:      "新用户登录注册",
			LoginType: 局_ctx.Z在线信息.LoginAppid,
			Time:      time.Now().Unix(),
		})
	}

	//账号模式登录成功把登录信息写到账号表
	if 局_ctx.AppInfo.AppType == 1 || 局_ctx.AppInfo.AppType == 2 {
		go service.NewUser(c, &db).Update(局_Uid, map[string]interface{}{
			"LoginAppid": 局_ctx.AppInfo.AppId,
			"LoginIp":    c.ClientIP(),
			"LoginTime":  time.Now().Unix(),
		})
	}

	var 局_用户类型 dbm.DB_UserClass
	局_用户类型, err = service.NewUserClass(c, &db).Info(局_AppUser.UserClassId)
	if err != nil {
		局_用户类型.Name = "已删待改"
		局_用户类型.Mark = 0
	}
	更新上下文缓存在线信息(c)
	//这里吧成功的状态
	response.OkData(c, gin.H{
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

// UserApi_取登录状态 取登录状态
func UserApi_取登录状态(c *gin.Context) {
	局_ctx := 取上下文(c)

	if 局_ctx.Z在线信息.Status == 1 && 局_ctx.Z在线信息.Uid > 0 {
		response.Ok(c)
		return
	}

	response.Fail(c, constant.Status_未登录)
	return
}

// UserApi_用户登录注销 用户登录注销
func UserApi_用户登录注销(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	//if !检测用户登录在线正常(&局_在线信息) {   //用户说不应该判断是否已登陆,调用就注销,我觉得有道理
	//	response.Fail(c, response.Status_未登录)
	//	return
	//}
	_, err := service.NewLinksToken(c, &db).Updates([]int{局_ctx.Z在线信息.Id}, map[string]interface{}{
		"OutTime":    0,
		"Status":     2,
		"LogoutCode": 3, //用户操作注销
	})
	更新上下文缓存在线信息(c)
	if err != nil {
		response.Fail(c, constant.Status_操作失败)
	} else {
		response.Ok(c)
	}
	return
}

// UserApi_用户登录远程注销 用户登录远程注销
func UserApi_用户登录远程注销(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	// {"Api":"RemoteLogOut","User":"aaaaaa","PassWord":"ssssss","Token":"","Time":1684069624,"Status":27417}'
	局_id := 0

	if 局_ctx.AppInfo.AppType == 1 || 局_ctx.AppInfo.AppType == 2 {
		局_User, err := service.NewUser(c, &db).InfoName(局_ctx.Q请求明文.Get("User").String())
		if err != nil {
			response.FailMsg(c, constant.Status_操作失败, "用户不存在")
			return
		}
		if !utils2.BcryptCheck(局_ctx.Q请求明文.Get("PassWord").String(), 局_User.PassWord) {
			response.FailMsg(c, constant.Status_操作失败, "用户名或密码错误")
			return
		}
		局_id = 局_User.Id

	} else {
		局_卡, err := service.NewKa(c, &db).InfoKa(局_ctx.Q请求明文.Get("User").String())
		if err != nil || 局_卡.AppId != 局_ctx.AppInfo.AppId {
			response.FailMsg(c, constant.Status_操作失败, "卡号不存在")
			return
		}
		局_id = 局_卡.Id
	}
	var err error
	var 局_指定token = 局_ctx.Q请求明文.Get("Token").String()
	if 局_指定token == "" {
		//用户远程注销
		err = service.NewLinksToken(c, &db).Set批量注销Uid数组([]int{局_id}, 局_ctx.AppInfo.AppId, constant.Z注销_用户远程注销)
	} else {
		var 局_临时在线信息 dbm.DB_LinksToken
		局_临时在线信息, err = service.NewLinksToken(c, &db).InfoToken(局_指定token)
		if err != nil {
			response.Fail(c, constant.Status_操作失败)
			return
		}
		if 局_临时在线信息.Uid != 局_id { //只允许注销已经登陆的token,并且uid是自己的
			response.FailMsg(c, constant.Status_操作失败, "用户没有权限注销此token")
			return
		}
		_, err = service.NewLinksToken(c, &db).Updates([]int{局_临时在线信息.Id}, map[string]interface{}{
			"OutTime":    0,
			"Status":     2,
			"LogoutCode": 4, //用户远程注销
		})
	}

	更新上下文缓存在线信息(c)
	if err != nil {
		response.Fail(c, constant.Status_操作失败)
	} else {
		response.Ok(c)
	}
	return
}
