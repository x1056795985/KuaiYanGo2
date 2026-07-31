package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/logic/common/ka"
	"server/app/models/constant"
	"server/app/service"
	"strings"
)

// UserApi_卡号充值 卡号充值
func UserApi_卡号充值(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"UseKa","User":"aaaaaa","Ka":"aaaaaa","InviteUser":"aaaaaa","Time":1684071722,"Status":41016}
	局_用户 := 局_ctx.Q请求明文.Get("User").String()
	if 局_用户 == "" && 局_ctx.Z在线信息.Uid > 0 { //如果获取不到就充值在线用户
		局_用户 = 局_ctx.Z在线信息.User
	}
	局_卡号 := strings.TrimSpace(局_ctx.Q请求明文.Get("Ka").String())
	局_推荐人 := strings.TrimSpace(局_ctx.Q请求明文.Get("InviteUser").String())
	err := ka.L_ka.K卡号充值_事务(c, 局_ctx.AppInfo.AppId, 局_卡号, 局_用户, 局_推荐人)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	response.OkData(c, gin.H{"InviteUser": 局_推荐人 != ""})

	return
}

// UserApi_取卡号详情 1.0.277+版本添加可用
func UserApi_取卡号详情(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	// {"Api":"GetKaInfo","ka":"8987657"}
	kaInfo, err := service.NewKa(c, &db).InfoKa(局_ctx.Q请求明文.Get("Ka").String())
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "卡号不存在")
		return
	}
	if kaInfo.AppId != 局_ctx.AppInfo.AppId {
		response.FailMsg(c, constant.Status_操作失败, "非本应用卡号")
		return
	}

	response.OkData(c, gin.H{
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
