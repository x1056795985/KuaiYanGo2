package middleware

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_User"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strings"
	"time"
)

const 管理员后台 = 1
const 代理后台 = 2
const WEBApi = 3

// Token有效的才放行,否则返回Ttoken失效
func IsTokenAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		Token := c.Request.Header.Get("Token")

		if Token == "" {
			response.FailTokenErr(gin.H{"reload": true}, "请先登录", c)
			c.Abort()
			return
		}

		if global.GVA_DB == nil {
			response.FailTokenErr(gin.H{"reload": true}, "数据库连接失败,请重新设置", c)
			c.Abort()
			return
		}

		var DB_LinksToken DB.DB_LinksToken
		//这里如果报错  invalid memory address or nil pointer dereference   可能是配置文件数据库配置北山,global.GVA_DB 值为空
		err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).Find(&DB_LinksToken).Error
		// 没查到数据 或状态不正常
		if err != nil || DB_LinksToken.Status != 1 {
			response.FailTokenErr(gin.H{"reload": true}, "令牌已失效", c)
			c.Abort()
			return
		}

		if DB_LinksToken.LoginAppid != 管理员后台 {
			response.FailTokenErr(gin.H{"reload": true}, "非管理员后台令牌,请重新登录.", c)
			c.Abort()
			return
		}
		//fmt.Println(DB_LinksToken)
		//更新最后活动时间
		global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Update("LastTime", int(time.Now().Unix()))
		//把 userID 保存到上下文,这样逻辑层就不用再查询了
		c.Set("Uid", DB_LinksToken.Uid)
		c.Set("User", DB_LinksToken.User)
		c.Set("局_在线信息", DB_LinksToken)

		if global.GVA_CONFIG.X系统设置.W系统模式 == 1 { //演示模式不强制跳
			c.Next()
			return
		}

		//如果会员没登录,且请求没有 kuaiyan 关键字,就跳转快验个人中心
		if global.X系统信息.H会员帐号 == "" && strings.Index(c.Request.RequestURI, "KuaiYan") == -1 && strings.Index(c.Request.RequestURI, "GetAdminInfo") == -1 {
			response.FailTokenErr(gin.H{"KuaiYan": true}, "", c)
			c.Abort()
			return
		}

		if strings.Index(c.Request.RequestURI, "KuaiYan") == -1 && strings.Index(c.Request.RequestURI, "GetAdminInfo") == -1 && global.X系统信息.D到期时间 < time.Now().Unix() {
			局_计数 := 0
			局_计数缓存, ok := global.H缓存.Get("在线数量")
			//在个人中心那里获取就可以了,如果超过100 写入缓存,这样不影响速度
			if !ok {
				//如果没有写入缓存,就永远是0,直接放行
				局_计数 = 0
			} else {
				局_计数 = 局_计数缓存.(int)
			}
			//"恭喜管理员同时在线用户数量已超过100,请开通商业会员,感谢您的支持"
			局_临时, _ := base64.StdEncoding.DecodeString("5oGt5Zac566h55CG5ZGY5ZCM5pe25Zyo57q/55So5oi35pWw6YeP5bey6LaF6L+HMTAwLOivt+W8gOmAmuWVhuS4muS8muWRmCzmhJ/osKLmgqjnmoTmlK/mjIE=")
			if 局_计数 > 120 { //做一个容错处理 容错20
				response.FailTokenErr(gin.H{"KuaiYan": true}, string(局_临时), c)
				c.Abort()
				return
			}

		}

		// 继续处理请求
		c.Next()
	}
}

func IsAdminHost() gin.HandlerFunc {
	return func(c *gin.Context) {
		//需要处理 外网->宝塔->Nginx转发->快验,这种情况host会变成127.0.0.1,所以检测  Origin Referer 也没有域名才拦截
		//Origin:[http://ky.9w99.cn] Pragma:[no-cache] Referer:[http://ky.9w99.cn/Admin/]
		局_host := global.GVA_CONFIG.X系统设置.G管理员后台Host
		if 局_host != "" && 局_host != c.Request.Host && strings.Index(c.Request.Header.Get("Origin"), "://"+局_host) == -1 && strings.Index(c.Request.Header.Get("Referer"), "://"+局_host+"/Admin") == -1 {
			//Get没有Origin Referer 所以如果是Get并且内部访问直接放行
			if c.Request.Method == "GET" && c.Request.Host[:10] == "127.0.0.1:" {
				c.Next()
				return
			}

			if global.GVA_CONFIG.X系统设置.W系统模式 == 1056795985 {
				c.String(404, fmt.Sprintf("%v", c.Request))
			} else {
				c.String(404, "") //fmt.Sprintf("%v", c.Request)
			}
			c.Abort()
			return
		}
		// 继续处理请求
		c.Next()
	}
}

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
		err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).Find(&DB_LinksToken).Error
		// 没查到数据 或状态不正常
		if err != nil || DB_LinksToken.Status != 1 {
			response.FailTokenErr(gin.H{"reload": true}, "令牌已失效", c)
			c.Abort()
			return
		}

		if DB_LinksToken.LoginAppid != 代理后台 {
			response.FailTokenErr(gin.H{"reload": true}, "非代理后台令牌,请重新登录.", c)
			c.Abort()
			return
		}
		//更新最后活动时间
		global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Update("LastTime", time.Now().Unix())
		go Ser_User.Id置最后登录AppId(DB_LinksToken.Uid, 2, c.ClientIP())
		//把 userID 保存到上下文,这样逻辑层就不用再查询了
		c.Set("Uid", DB_LinksToken.Uid)
		c.Set("User", DB_LinksToken.User)
		c.Set("局_在线信息", DB_LinksToken)

		// 继续处理请求
		c.Next()
	}
}
func IsAgentHost() gin.HandlerFunc {
	return func(c *gin.Context) {
		//需要处理 外网->宝塔->Nginx转发->快验,这种情况host会变成127.0.0.1,所以检测  Origin Referer 也没有域名才拦截
		//Origin:[http://ky.9w99.cn] Pragma:[no-cache] Referer:[http://ky.9w99.cn/Admin/]
		局_host := global.GVA_CONFIG.X系统设置.D代理后台Host
		if 局_host != "" && 局_host != c.Request.Host && strings.Index(c.Request.Header.Get("Origin"), "://"+局_host) == -1 && strings.Index(c.Request.Header.Get("Referer"), "://"+局_host+"/Admin") == -1 {
			//Get没有Origin Referer 所以如果是Get并且内部访问直接放行
			if c.Request.Method == "GET" && c.Request.Host[:10] == "127.0.0.1:" {
				c.Next()
				return
			}

			if global.GVA_CONFIG.X系统设置.W系统模式 == 1056795985 {
				c.String(404, fmt.Sprintf("%v", c.Request))
			} else {
				c.String(404, "") //fmt.Sprintf("%v", c.Request)
			}
			c.Abort()
			return
		}
		// 继续处理请求
		c.Next()
	}
}

func IsAgent是否关闭() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !global.GVA_CONFIG.X系统设置.D代理中心开关 {
			c.String(404, global.GVA_CONFIG.X系统设置.D代理中心关闭提示)
			c.Abort()
			return
		}
		// 继续处理请求
		c.Next()
	}
}
