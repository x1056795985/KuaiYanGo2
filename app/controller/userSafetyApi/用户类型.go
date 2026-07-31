package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	"time"
)

// UserApi_取用户类型列表 取用户类型列表
func UserApi_取用户类型列表(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB

	局_列表, err := service.NewUserClass(c, &db).GetAllByAppId(局_ctx.AppInfo.AppId)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "获取用户类型列表失败")
		return
	}
	var 局_响应 []gin.H

	for _, 单列表 := range 局_列表 {
		局_响应 = append(局_响应, gin.H{"Name": 单列表.Name, "Mark": 单列表.Mark, "Weight": 单列表.Weight})
	}

	response.OkData(c, 局_响应)
	return
}

// UserApi_置用户类型 置用户类型
func UserApi_置用户类型(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := *global.GVA_DB
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	局_新用户类型, err := service.NewUserClass(c, &db).InfoByMark(局_ctx.AppInfo.AppId, int(局_ctx.Q请求明文.Get("Mark").Int()))
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户类型代号不存在")
		return
	}
	局_App用户, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "用户不存在")
		return
	}

	局_旧用户类型, err := service.NewUserClass(c, &db).InfoByMark(局_ctx.AppInfo.AppId, 局_App用户.UserClassId)
	if err != nil { //如果是没有的类型就赋值 未分类
		局_旧用户类型 = dbm.DB_UserClass{AppId: 局_ctx.AppInfo.AppId, Name: "未分类", Weight: 1}
	}

	if 局_旧用户类型.Mark == 局_新用户类型.Mark { //代号相同,直接转换即可
		response.OkData(c, gin.H{"UserClassMark": 局_新用户类型.Mark, "UserClassName": 局_新用户类型.Name, "VipTime": 局_App用户.VipTime})
		return
	} else {
		局_现行时间戳 := time.Now().Unix()
		// 用户类型不同, 根据权重处理
		if 局_ctx.AppInfo.AppType == 2 || 局_ctx.AppInfo.AppType == 4 {
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
	_, err = service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).UpdateUid(局_App用户.Uid, map[string]interface{}{
		"UserClassId": 局_App用户.UserClassId,
		"VipTime":     局_App用户.VipTime,
	})

	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "写入新用户类型和Vip失败")
		return
	}
	response.OkData(c, gin.H{"UserClassMark": 局_新用户类型.Mark, "UserClassName": 局_新用户类型.Name, "VipTime": 局_App用户.VipTime})
	return
}
