package userSafetyApi

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/logic/common/appUser"
	"server/app/logic/common/log"
	"server/app/logic/common/user"
	"server/app/models/constant"
	db2 "server/app/models/db"
	"server/app/service"
	"server/app/utils/Qqwry"
	"strings"
	"time"
)

// UserApi_用户减少余额 用户减少余额
func UserApi_用户减少余额(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	//{"Api":"UserReduceMoney","Money":1.01,"Log":"看你长得帅,减些钱","AgentId":10,"AgentMoney":0,"AgentMoneyLog":"代理分成"}

	if 局_ctx.Q请求明文.Get("AgentId").Int() > 0 {
		response.FailMsg(c, constant.Status_操作失败, "服务端1.0.363+该功能已删除,如有需要请使用更安全的apiHook实现")
		return
	}
	局_User, err := service.NewUser(c, &db).Info(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}
	var 局_增减值 float64
	局_增减值 = 局_ctx.Q请求明文.Get("Money").Float64()
	if 局_增减值 <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "不能为小于等于0")
		return
	}
	if 局_User.Rmb < 局_增减值 {
		response.FailMsg(c, constant.Status_操作失败, "余额不足")
		return
	}

	新余额, err := user.L_user.Id余额增减(c, 局_User.Id, 局_增减值, false)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error()) //基本就是余额不足
		return
	}

	go log.L_log.S输出日志(c, db2.DB_LogMoney{
		User:  局_User.User,
		Ip:    c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP()),
		Time:  time.Now().Unix(),
		Count: Float64取负值(局_增减值),
		Note:  fmt.Sprintf("%s|新余额%v", 局_ctx.Q请求明文.Get("Log").String(), 新余额),
	})
	response.OkData(c, gin.H{"Money": 新余额})
}

// UserApi_用户减少点数 用户减少点数
func UserApi_用户减少点数(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	// {"Api":"UserReduceMoney","VipTime":1.3,"Log":"看你长得帅,扣点钱"}
	if 局_ctx.AppInfo.AppType != 2 && 局_ctx.AppInfo.AppType != 4 { //检查是不是计点模式
		response.FailMsg(c, constant.Status_操作失败, "应用非计点模式不可使用")
		return
	}
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.Z在线信息.LoginAppid).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}
	var 局_增减值 = 局_ctx.Q请求明文.Get("VipTime").Int64()
	if 局_增减值 <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "不能为小于等于0")
		return
	}
	if 局_AppUser.VipTime < 局_增减值 {
		response.FailMsg(c, constant.Status_操作失败, "点数不足")
		return
	}

	err = appUser.L_appUser.Id点数增减(c, 局_ctx.AppInfo.AppId, 局_AppUser.Id, 局_增减值, false)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error()) //基本就是点数不足
		return
	}

	局_AppUser.VipTime -= 局_增减值
	response.OkData(c, gin.H{"VipTime": 局_AppUser.VipTime})
	go log.L_log.S输出日志(c, db2.DB_LogVipNumber{
		User:  局_ctx.Z在线信息.User,
		AppId: 局_ctx.AppInfo.AppId,
		Type:  2,
		Ip:    c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP()),
		Time:  time.Now().Unix(),
		Count: Float64取负值(float64(局_增减值)),
		Note:  fmt.Sprintf("%s|剩余%v", 局_ctx.Q请求明文.Get("Log").String(), 局_AppUser.VipNumber),
	})
	return
}

// UserApi_用户减少积分 用户减少积分
func UserApi_用户减少积分(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	//{"Api":"UserReduceMoney","VipNumber":1.3,"Log":"看你长得帅,扣点钱","UniqueStr":"",UniqueTime:0}
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.Z在线信息.LoginAppid).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}
	var 局_增减值 = 局_ctx.Q请求明文.Get("VipNumber").Float64()

	if 局_增减值 <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "不能为小于等于0")
		return
	}

	if 局_ctx.Q请求明文.Get("AgentId").Int() > 0 {
		response.FailMsg(c, constant.Status_操作失败, "服务端1.0.363+该功能已删除,如有需要请使用更安全的apiHook实现")
		return
	}

	if 局_AppUser.VipNumber < 局_增减值 {
		response.FailMsg(c, constant.Status_操作失败, "积分不足")
		return
	}

	局_唯一标志 := 局_ctx.Q请求明文.Get("UniqueStr").String()

	err = appUser.L_appUser.Uid积分减少(c, 局_ctx.AppInfo.AppId, 局_AppUser.Uid, 局_增减值, 局_唯一标志, 局_ctx.Q请求明文.Get("UniqueTime").Int64())
	if err != nil && strings.Contains(err.Error(), "唯一标识") {
		response.FailMsg(c, constant.Status_唯一标识重复, err.Error())
		return
	}
	if err != nil && strings.Contains(err.Error(), "积分不足") { //基本就是积分不足
		response.FailMsg(c, constant.Status_积分不足, err.Error())
		return
	}

	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	// flosat64 直接
	局_增减值D := decimal.NewFromFloat(局_增减值)
	局_用户积分D := decimal.NewFromFloat(局_AppUser.VipNumber)

	局_用户积分D = 局_用户积分D.Sub(局_增减值D)
	局_AppUser.VipNumber, _ = 局_用户积分D.Float64()

	局_增减值, _ = 局_增减值D.Mul(decimal.NewFromFloat(-1)).Float64() //乘-1 变成负数

	response.OkData(c, gin.H{"VipNumber": 局_AppUser.VipNumber})
	go log.L_log.S输出日志(c, db2.DB_LogVipNumber{
		User:  局_ctx.Z在线信息.User,
		AppId: 局_ctx.AppInfo.AppId,
		Type:  1,
		Ip:    c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP()),
		Time:  time.Now().Unix(),
		Count: 局_增减值,
		Note:  fmt.Sprintf("%s|≈%v", 局_ctx.Q请求明文.Get("Log").String(), 局_AppUser.VipNumber),
	})
	return
}
