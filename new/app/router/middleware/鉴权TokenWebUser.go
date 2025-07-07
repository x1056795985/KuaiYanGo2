package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"server/global"
	"server/new/app/models/constant"
	"server/structs/Http/response"
	DB "server/structs/db"
	"time"
)

// Token有效的才放行,否则返回Ttoken失效
func IsTokenWebUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		Token := c.Request.Header.Get("Token") //优先协议头的,Token
		if Token == "" {                       //如果协议头没有,再读取,url内的
			Token = c.Request.FormValue("Token")
		}

		if Token == "" {
			response.FailTokenErr(gin.H{"reload": true}, "请先登录", c)
			c.Abort()
			return
		}
		var DB_LinksToken DB.DB_LinksToken
		//这里如果报错  invalid memory address or nil pointer dereference   可能是配置文件数据库配置北山,global.GVA_DB 值为空
		err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).First(&DB_LinksToken).Error
		// 没查到数据 或状态不正常
		if err != nil || DB_LinksToken.Status != 1 {
			response.FailTokenErr(gin.H{}, "令牌已失效", c)
			c.Abort()
			return
		}

		if DB_LinksToken.LoginAppid != constant.APPID_Web用户中心 {
			response.FailTokenErr(gin.H{}, "非WebUser令牌,请重新登录", c)
			c.Abort()
			return
		}

		data, err := c.GetRawData() //GetRawData只能使用一次
		if err != nil {
			response.FailTokenErr(gin.H{}, "参数错误", c)
			c.Abort()
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data)) // 关键点 //通过这写回post数据,就可以多次读取了
		c.Set("DB_LinksToken", DB_LinksToken)
		//更新最后活动时间
		global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Updates(map[string]interface{}{"LastTime": int(time.Now().Unix()), "Ip": c.ClientIP()})
		// 继续处理请求
		c.Next()
	}
}
