package userSafetyApi

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/logic/common/appUser"
	"server/new/app/logic/common/blacklist"
	"server/new/app/logic/common/log"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	utils2 "server/new/app/utils"
	"time"
)

// UserApi_置新绑定信息 置新绑定信息
func UserApi_置新绑定信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB

	// {"Api":"SetAppUserKey","NewKey":"8987657"}
	// 检查是否可以换换绑
	if 局_ctx.AppInfo.VerifyKey != 1 && 局_ctx.AppInfo.VerifyKey != 3 { //1和3 可以换绑
		response.FailMsg(c, constant.Status_操作失败, "应用禁止更换绑定信息.")
		return
	}

	局_Uid := 局_ctx.Z在线信息.Uid
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		局_账号 := 局_ctx.Q请求明文.Get("User").String()

		局密码 := 局_ctx.Q请求明文.Get("PassWord").String()
		if 局_账号 == "" {
			response.Fail(c, constant.Status_未登录)
			return
		} else {
			局_ctx.Z在线信息.User = 局_账号                                       //如果出错,写日志时会用到
			if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 { //是卡号
				局_卡, err := service.NewKa(c, &db).InfoKa(局_账号)
				if err != nil {
					response.FailMsg(c, constant.Status_操作失败, "卡号不存在.")
					return
				}
				局_Uid = 局_卡.Id
			} else {
				局_User, err := service.NewUser(c, &db).InfoName(局_账号)
				if err != nil {
					response.FailMsg(c, constant.Status_操作失败, "用户不存在.")
					return
				}
				if 局密码 == "" || !utils2.BcryptCheck(局密码, 局_User.PassWord) {
					go log.L_log.S输出日志(c, dbm.DB_LogLogin{
						User:      局_User.User,
						Ip:        c.ClientIP(),
						Note:      "更换绑定登录时密码错误:" + 局密码,
						LoginType: 局_ctx.AppInfo.AppId,
					})
					response.FailMsg(c, constant.Status_登录失败, "用户名或密码错误")
					return
				}
				局_Uid = 局_User.Id
			}
		}

	}

	局_信息绑定信息 := 局_ctx.Q请求明文.Get("NewKey").String()
	if 局_信息绑定信息 == "" {
		response.FailMsg(c, constant.Status_绑定信息验证失败, "新绑定信息不能为空.")
		return
	}

	// 检查是否可以绑定相同信息
	if 局_ctx.AppInfo.IsUserKeySame == 2 {
		_, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoKey(局_信息绑定信息)
		if err == nil {
			response.FailMsg(c, constant.Status_绑定信息已被其他用户使用, "绑定信息已被其他用户绑定.")
			return
		}
	}
	if blacklist.Is黑名单(局_信息绑定信息, 局_ctx.AppInfo.AppId) {
		response.FailMsg(c, constant.Status_黑名单信息, "绑定信息为黑名单信息")
		return
	}

	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.可能刚注册还没登录成功")
		return
	}

	err, 扣时间值 := 绑定信息更换规则校验(c, 局_ctx.AppInfo, 局_Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	// 如果换绑需要扣点,就执行扣点, 		且原来绑定信息不能为空
	if 扣时间值 > 0 && 局_AppUser.Key != "" {
		err = appUser.L_appUser.Id点数增减(c, 局_ctx.AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), false)
		if err != nil {
			response.FailMsg(c, constant.Status_Vip已到期, "剩余会员时间或点数不足.")
			return
		} else {
			局_日志 := "用户置新绑定,旧绑定信息:" + 局_AppUser.Key + ",新绑定信息:" + 局_信息绑定信息
			局_type := 3
			if 局_ctx.AppInfo.AppType == 2 || 局_ctx.AppInfo.AppType == 4 {
				局_type = 2
			}
			log.L_log.S输出日志(c, dbm.DB_LogVipNumber{
				User:  局_ctx.Z在线信息.User,
				Ip:    c.ClientIP(),
				Note:  局_日志,
				Count: D到数值(-扣时间值),
				AppId: 局_ctx.AppInfo.AppId,
				Type:  局_type,
			})
		}
	}
	_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).UpdateUid(局_Uid, map[string]interface{}{"Key": 局_信息绑定信息})
	if err == nil {
		// 取用户名用于日志
		局_用户名 := 取Uid对应名称(c, &db, 局_ctx.AppInfo.AppId, 局_Uid)
		_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
			Type:   constant.LogKey_换绑,
			User:   局_用户名,
			Uid:    局_Uid,
			AppId:  局_ctx.AppInfo.AppId,
			OldKey: 局_AppUser.Key,
			NewKey: 局_信息绑定信息,
			Time:   time.Now().Unix(),
			Ip:     c.ClientIP(),
			Count:  D到数值(-扣时间值),
			Note:   "置新绑定信息",
		})
		if err != nil {
			global.GVA_LOG.Println("修改绑定信息日志写入失败:" + err.Error())
		}
		response.OkData(c, gin.H{"ReduceVipTime": 扣时间值})
	} else {
		//退还已经扣除的点数
		if 扣时间值 > 0 {
			_ = appUser.L_appUser.Id点数增减(c, 局_ctx.AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), true)
		}
		response.Fail(c, constant.Status_SQl错误)
	}

	return
}

// UserApi_解除绑定信息 解除绑定信息
func UserApi_解除绑定信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB

	// {"Api":"SetAppUserKey"}
	// 检查是否可以换换绑
	if 局_ctx.AppInfo.VerifyKey != 1 && 局_ctx.AppInfo.VerifyKey != 3 { //1和3 可以换绑
		response.FailMsg(c, constant.Status_操作失败, "应用禁止更换绑定信息.")
		return
	}
	局_Uid := 局_ctx.Z在线信息.Uid
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		局_账号 := 局_ctx.Q请求明文.Get("User").String()
		局密码 := 局_ctx.Q请求明文.Get("PassWord").String()
		if 局_账号 == "" {
			response.Fail(c, constant.Status_未登录)
			return
		} else {
			局_ctx.Z在线信息.User = 局_账号                                       //如果出错,写日志时会用到
			if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 { //是卡号
				局_卡, err := service.NewKa(c, &db).InfoKa(局_账号)
				if err != nil {
					response.FailMsg(c, constant.Status_操作失败, "卡号不存在.")
					return
				}
				局_Uid = 局_卡.Id
			} else {
				局_User, err := service.NewUser(c, &db).InfoName(局_账号)
				if err != nil {
					response.FailMsg(c, constant.Status_操作失败, "用户不存在.")
					return
				}
				if 局密码 == "" || !utils2.BcryptCheck(局密码, 局_User.PassWord) {
					go log.L_log.S输出日志(c, dbm.DB_LogLogin{
						User:      局_User.User,
						Ip:        c.ClientIP(),
						Note:      "更换绑定登录时密码错误:" + 局密码,
						LoginType: 局_ctx.AppInfo.AppId,
					})
					response.FailMsg(c, constant.Status_登录失败, "用户名或密码错误")
					return
				}
				局_Uid = 局_User.Id
			}
		}

	}

	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.可能刚注册还没登录成功")
		return
	}
	if 局_AppUser.Key == "" {
		response.FailMsg(c, constant.Status_操作失败, "无绑定信息,无需解除")
		return
	}

	err, 扣时间值 := 绑定信息更换规则校验(c, 局_ctx.AppInfo, 局_Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	// 如果换绑需要扣点,就执行扣点
	if 扣时间值 > 0 {
		err = appUser.L_appUser.Id点数增减(c, 局_ctx.AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), false)
		if err != nil {
			response.FailMsg(c, constant.Status_Vip已到期, "剩余会员时间或点数不足.")
			return
		} else {
			局_日志 := "用户解除绑定信息,旧绑定信息:" + 局_AppUser.Key
			局_type := 3
			if 局_ctx.AppInfo.AppType == 2 || 局_ctx.AppInfo.AppType == 4 {
				局_type = 2
			}
			log.L_log.S输出日志(c, dbm.DB_LogVipNumber{
				User:  局_ctx.Z在线信息.User,
				Ip:    c.ClientIP(),
				Note:  局_日志,
				Count: D到数值(-扣时间值),
				AppId: 局_ctx.AppInfo.AppId,
				Type:  局_type,
			})
		}
	}

	_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).UpdateUid(局_Uid, map[string]interface{}{"Key": ""})

	if err == nil {
		// 取用户名用于日志
		局_用户名 := 取Uid对应名称(c, &db, 局_ctx.AppInfo.AppId, 局_Uid)
		_, err = service.NewLogKey(c, &db).Create(&dbm.DB_LogKey{
			Type:   constant.LogKey_解绑,
			User:   局_用户名,
			Uid:    局_Uid,
			AppId:  局_ctx.AppInfo.AppId,
			OldKey: 局_AppUser.Key,
			NewKey: "",
			Time:   time.Now().Unix(),
			Ip:     c.ClientIP(),
			Count:  D到数值(-扣时间值),
			Note:   "解除绑定信息",
		})
		if err != nil {
			global.GVA_LOG.Println("修改绑定信息日志写入失败:" + err.Error())
		}
		response.OkData(c, gin.H{"ReduceVipTime": 扣时间值})
	} else {
		//退还已经扣除的点数
		if 扣时间值 > 0 {
			_ = appUser.L_appUser.Id点数增减(c, 局_ctx.AppInfo.AppId, 局_AppUser.Id, int64(扣时间值), true)
		}
		// 暂时想不出什么情况会修改失败 概率较低
		response.Fail(c, constant.Status_SQl错误)
	}

	return
}

// 取Uid对应名称 根据AppType获取用户名称或卡号名称(用于日志记录)
func 取Uid对应名称(c *gin.Context, db *gorm.DB, AppId int, Uid int) string {
	// 通过AppInfo判断是卡号还是用户模式
	局_AppInfo, err := service.NewAppInfo(c, db).Info(AppId)
	if err != nil {
		return ""
	}
	if 局_AppInfo.AppType == 3 || 局_AppInfo.AppType == 4 {
		//卡号模式
		局_卡, err := service.NewKa(c, db).Info(Uid)
		if err != nil {
			return ""
		}
		return 局_卡.Name
	}
	//用户模式
	局_User, err := service.NewUser(c, db).Info(Uid)
	if err != nil {
		return ""
	}
	return 局_User.User
}
