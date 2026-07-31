package middleware

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"server/app/global"
	"server/app/logic/common/blacklist"
	"server/app/logic/common/setting"
	"server/app/models/common"
	"server/app/models/constant"
	"server/app/service"
	"server/app/utils"
	"time"
)

// Token有效的才放行,否则返回Token失效
func X系统关闭则响应() gin.HandlerFunc {
	return func(c *gin.Context) {
		if setting.Q系统设置().X系统开关 {
			c.Next()
			return
		}

		c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), constant.Status_系统已关闭, setting.Q系统设置().X系统关闭提示})
		c.Abort()
	}
}

func UserApi检查数据库连接() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GVA_DB == nil {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), constant.Status_SQl错误, "服务器还未连接数据库,暂不可用,请管理员检查原因,或重启系统"})
			c.Abort()
		}
		c.Next()
	}
}

func C初始化上下文() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 构建请求上下文并注入
		var err error
		ctx := &common.Q请求_上下文{}
		ctx.AppInfo.AppId = D到整数(c.DefaultQuery("AppId", ""))
		if ctx.AppInfo.AppId < 10000 {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), constant.Status_App不存在, "App不存在"})
			c.Abort()
		}
		db := global.GVA_DB
		ctx.AppInfo, err = service.NewAppInfo(c, db).Info(ctx.AppInfo.AppId)
		if err != nil {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), constant.Status_App不存在, "App不存在"})
			c.Abort()
		}

		if ctx.AppInfo.Status == 1 {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), constant.Status_已停止运营, ctx.AppInfo.AppStatusMessage})
			c.Abort()
		}
		utils.Z置上下文(c, ctx)
		c.Next()
	}
}
func J检查黑名单() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.Q取上下文(c)
		if blacklist.Is黑名单(c.ClientIP(), ctx.AppInfo.AppId) {
			c.JSON(http.StatusOK, 请求响应_X响应状态{time.Now().Unix(), constant.Status_黑名单信息, "黑名单ip"})
			c.Abort()
		}
		c.Next()
	}
}
