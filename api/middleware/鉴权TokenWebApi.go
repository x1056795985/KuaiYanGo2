package middleware

import (
	"bytes"
	E "github.com/duolabmeng6/goefun/eTool"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strings"
	"time"
)

const WebApi = 3

// Token有效的才放行,否则返回Ttoken失效
func IsTokenWebApi() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GVA_CONFIG.X系统设置.WebApiHost != "" && global.GVA_CONFIG.X系统设置.WebApiHost != c.Request.Host {
			c.String(404, "")
			c.Abort()
			return
		}

		Token := c.Request.Header.Get("Token")
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

		if DB_LinksToken.LoginAppid != WebApi {
			response.FailTokenErr(gin.H{}, "非WebApi令牌,请管理员到在线列表创建WebApi令牌", c)
			c.Abort()
			return
		}

		//判断令牌是否有接口权限
		if strings.Index(DB_LinksToken.Key, E.E文本_取右边(c.Request.URL.Path, "/WebApi/")) == -1 {
			response.FailTokenErr(gin.H{}, "令牌无本接口权限", c)
			c.Abort()
			return
		}

		//更新最后活动时间
		data, err := c.GetRawData() //GetRawData只能使用一次
		if err != nil {
			response.FailTokenErr(gin.H{}, "参数错误", c)
			c.Abort()
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data)) // 关键点 //通过这写回post数据,就可以多次读取了

		c.Set("局_json明文", string(data))
		global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Updates(map[string]interface{}{"LastTime": int(time.Now().Unix()), "Ip": c.ClientIP()})
		// 继续处理请求
		c.Next()
	}
}
