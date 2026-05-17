package middleware

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/setting"
	"server/new/app/models/constant"
	"server/structs/Http/response"
	DB "server/structs/db"
	"time"
)

// Token有效的才放行,否则返回Ttoken失效
func IsTokenAgent() gin.HandlerFunc {
	return func(c *gin.Context) {

		Token := c.Request.Header.Get("Token")
		if Token == "" {
			response.FailTokenErr(gin.H{"reload": true}, "请先登录", c)
			c.Abort()
			return
		}

		if global.GVA_DB == nil {
			response.FailTokenErr(gin.H{"reload": true}, "数据库连接失败,请联系管理员", c)
			c.Abort()
			return
		}
		var DB_LinksToken DB.DB_LinksToken
		//这里如果报错  invalid memory address or nil pointer dereference   可能是配置文件数据库配置北山,global.GVA_DB 值为空
		err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).First(&DB_LinksToken).Error
		// 没查到数据 或状态不正常
		if err != nil || DB_LinksToken.Status != 1 {
			response.FailTokenErr(gin.H{"reload": true}, "令牌已失效", c)
			c.Abort()
			return
		}

		if DB_LinksToken.LoginAppid != constant.APPID_代理平台 {
			response.FailTokenErr(gin.H{"reload": true}, "非代理后台令牌,请重新登录.", c)
			c.Abort()
			return
		}
		//更新最后活动时间
		if time.Now().Unix()-DB_LinksToken.LastTime > 60 { //超过1分钟,更新最后活动时间
			global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Update("LastTime", time.Now().Unix())
		}
		go Ser_User.Id置最后登录AppId(DB_LinksToken.Uid, 2, c.ClientIP())
		//把 userID 保存到上下文,这样逻辑层就不用再查询了
		c.Set("Uid", DB_LinksToken.Uid)
		c.Set("User", DB_LinksToken.User)
		c.Set("局_在线信息", DB_LinksToken)

		// 继续处理请求
		c.Next()
	}
}

func IsAgent是否关闭() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !setting.Q系统设置().D代理中心开关 {
			c.String(404, setting.Q系统设置().D代理中心关闭提示)
			c.Abort()
			return
		}
		// 继续处理请求
		c.Next()
	}
}
