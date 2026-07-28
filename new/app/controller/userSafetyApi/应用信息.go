package userSafetyApi

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"regexp"
	"server/new/app/global"
	"server/new/app/models/constant"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/logic/common/cloudStorage"
	"server/new/app/logic/common/publicData"
	dbm "server/new/app/models/db"
	"server/new/app/service"
)

// UserApi_取应用公告 取应用公告
func UserApi_取应用公告(c *gin.Context) {
	局_ctx := 取上下文(c)
	response.OkData(c, gin.H{"AppGongGao": 局_ctx.AppInfo.AppGongGao})
	return
}

// UserApi_取新版本下载地址 取新版本下载地址
func UserApi_取新版本下载地址(c *gin.Context) {
	局_ctx := 取上下文(c)

	局_下载地址 := App下载更新地址变量处理(局_ctx.AppInfo)

	response.OkData(c, gin.H{"AppUpDataJson": 局_下载地址})
	return
}

// UserApi_取应用专属变量 取应用专属变量
func UserApi_取应用专属变量(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := 局_ctx.Q请求明文.Get("Name").String()
	局_云变量数据, err := publicData.L_publicData.Q取值2(c, 局_ctx.AppInfo.AppId, 局_变量名)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "变量不存在,请到后台应用编辑,添加专属变量")
		return
	}
	if 局_云变量数据.IsVip == 0 || 检测用户登录在线正常(&局_ctx.Z在线信息) {
		if 局_云变量数据.IsVip > 0 { //只有返回VIP变量时才强制
			局_ctx.RSA强制 = true
		}

		response.OkData(c, gin.H{局_变量名: 局_云变量数据.Value})
	} else {
		response.Fail(c, constant.Status_未登录)
	}
	return
}

// UserApi_取公共变量 取公共变量
func UserApi_取公共变量(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"GetPublicData","Name":"会员数据a"}
	局_变量名 := 局_ctx.Q请求明文.Get("Name").String()
	取值2, err := publicData.L_publicData.Q取值2(c, 1, 局_变量名)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	response.OkData(c, gin.H{局_变量名: 取值2.Value, "QueueCount": 取值2})
	return
}

// UserApi_置公共变量 置公共变量
func UserApi_置公共变量(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"GetPublicData","Name":"会员数据a","Value":"aaaaa"}
	局_变量名 := 局_ctx.Q请求明文.Get("Name").String()
	局_变量值 := 局_ctx.Q请求明文.Get("Value").String()
	err := publicData.L_publicData.Z置值(c, 1, 局_变量名, 局_变量值)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}
	response.Ok(c)
	return
}

// UserApi_取应用最新版本 取应用最新版本
func UserApi_取应用最新版本(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"GetAppVersion","Version":"1.3.5","IsVersionAll":true}
	局_可用版本 := W文本_分割文本(局_ctx.AppInfo.AppVer, "\n")
	if len(局_可用版本) == 0 || 局_ctx.AppInfo.AppVer == "" {
		response.FailMsg(c, constant.Status_操作失败, "应用未设置版本号或格式不正确")
		return
	}

	局_分解版本号 := W文本_分割文本(局_可用版本[0], ".")
	局_分解版本号最新 := 版本号_分解(局_可用版本[0])
	局_版本号当前 := 局_ctx.Q请求明文.Get("Version").String()

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
				if 局_ctx.Q请求明文.Get("IsVersionAll").Bool() {
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
		response.OkData(c, gin.H{"NewVersion": 局_可用版本[0], "Version": 局_分解版本号最新.大版本号, "IsUpdate": 局_是否更新})
		return
	} else {
		// 有大小版本号
		局_小数运算, _ := decimal.NewFromString(局_分解版本号[0] + "." + 局_分解版本号[1])
		局_双精度版本, _ := 局_小数运算.Float64()
		response.OkData(c, gin.H{"NewVersion": 局_可用版本[0], "Version": 局_双精度版本, "IsUpdate": 局_是否更新})
		return
	}
}

// UserApi_取应用主页Url 取应用主页Url
func UserApi_取应用主页Url(c *gin.Context) {
	局_ctx := 取上下文(c)
	response.OkData(c, gin.H{"AppHomeUrl": 局_ctx.AppInfo.UrlHome})
	return
}

// UserApi_取应用基础信息 1.0.42+版本添加可用
func UserApi_取应用基础信息(c *gin.Context) {
	局_ctx := 取上下文(c)
	db := global.GVA_DB
	var AppInfoWebUser dbm.DB_AppInfoWebUser
	AppInfoWebUser, _ = service.NewAppInfoWebUser(c, db).Info(局_ctx.AppInfo.AppId)

	response.OkData(c, gin.H{
		"AppId":            局_ctx.AppInfo.AppId,
		"AppType":          局_ctx.AppInfo.AppType,
		"AppName":          局_ctx.AppInfo.AppName,
		"AppWeb":           局_ctx.AppInfo.AppWeb,
		"Status":           局_ctx.AppInfo.Status,
		"AppStatusMessage": 局_ctx.AppInfo.AppStatusMessage,
		"WebUserStatus":    AppInfoWebUser.Status,
		"WebUserDomain":    S三元(AppInfoWebUser.Status == 1, AppInfoWebUser.WebUserDomain, ""),
	})
	return
}

// UserApi_取服务器连接状态 取服务器连接状态
func UserApi_取服务器连接状态(c *gin.Context) {
	response.Ok(c)
	return
}

// UserApi_取开启验证码接口 取开启验证码接口
func UserApi_取开启验证码接口(c *gin.Context) {
	局_ctx := 取上下文(c)

	response.OkData(c, 局_ctx.AppInfo.Captcha)
	return
}

// App下载更新地址变量处理 处理应用下载更新地址中的模板变量
func App下载更新地址变量处理(DB_AppInfo dbm.DB_AppInfo) string {
	局_新文本 := DB_AppInfo.UrlDownload

	局_新文本 = strings.Replace(局_新文本, "{{AppName}}", DB_AppInfo.AppName, -1)

	if strings.Index(局_新文本, "{{AppVer}}") != -1 && DB_AppInfo.AppVer != "" {
		局_可用版本 := W文本_分割文本(DB_AppInfo.AppVer, "\n")
		if len(局_可用版本) > 0 {
			局_新文本 = strings.Replace(局_新文本, "{{AppVer}}", 局_可用版本[0], -1)
		}
	}

	//{{(.*?)\((.*?)\)}}  正则匹配指令,  子匹配1为指令名 子匹配2为参数
	if strings.Index(局_新文本, "{{") != -1 { //判断是否还有变量
		re := regexp.MustCompile(`{{(.*?)\((.*?)\)}}`)
		result := re.FindAllStringSubmatch(局_新文本, -1)
		for i, _ := range result {
			局_完整文本 := result[i][0]
			局_指令名 := result[i][1]
			局_参数 := W文本_分割文本(result[i][2], ",")
			switch 局_指令名 {
			case "云存储_取外链":
				if len(局_参数) == 2 {
					下载地址, err := cloudStorage.L_云存储.Q取外链地址(&gin.Context{}, strings.Trim(局_参数[0], "'"), gconv.Int64(局_参数[1]))
					if err == nil {
						局_新文本 = strings.Replace(局_新文本, 局_完整文本, 下载地址, -1)
					}
				}
			}
		}
	}

	return 局_新文本
}
