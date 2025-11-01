package middleware

import (
	"EFunc/utils"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"server/global"
	"server/new/app/logic/common/setting"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strings"
	"time"
)

const WebApi = 3

func IsWebApiHost() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GVA_DB == nil {
			c.String(404, "数据库连接失败,请重新设置", c)
			c.Abort()
			return
		}
		//需要处理 外网->宝塔->Nginx转发->快验,这种情况host会变成127.0.0.1,所以检测  Origin Referer 也没有域名才拦截
		局_host := setting.Q系统设置().WebApiHost
		if 局_host != "" && 局_host != c.Request.Host && strings.Index(c.Request.Header.Get("Origin"), "://"+局_host) == -1 && strings.Index(c.Request.Header.Get("Referer"), "://"+局_host+"/Admin") == -1 {
			/*			//Get没有Origin Referer 所以如果是Get并且内部访问直接放行  WebApi没有Get 必须带 Referer
						//如果伪造请求过多,直接连Origin Referer 都禁止,开发者去宝塔配置Nginx转发 让其转发host
						if c.Request.Method == "GET" && c.Request.Host[:10] == "127.0.0.1:" {
							c.Next()
							return
						}*/

			if global.GVA_Viper.GetInt("系统模式") == 1056795985 {
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
func IsTokenWebApi() gin.HandlerFunc {
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

		if DB_LinksToken.LoginAppid != WebApi {
			response.FailTokenErr(gin.H{}, "非WebApi令牌,请管理员到在线列表创建WebApi令牌", c)
			c.Abort()
			return
		}
		//20251101  进行权限改变增加公共函数,更细致化权限控制 //因为权限长度不能超过191
		//旧权限例子  任务处理获取(TaskPoolGetTask),任务处理返回(TaskPoolSetTask),任务池_任务创建(TaskPoolNewData),运行公共js函数(RunJs),取公共变量(GetPublicData),置公共变量(SetPublicData),任务池_任务查询(TaskPoolGetData),
		//新权限例子  (TaskPoolGetTask)(TaskPoolSetTask)(TaskPoolNewData)(RunJs)(GetPublicData)(SetPublicData)(TaskPoolGetData)[公共函数名称][aaaxxx]

		//判断令牌是否有接口权限
		局_接口名称 := utils.W文本_取文本右边(c.Request.URL.Path, "/WebApi/")
		//两种格式的接口,
		//第一种  (接口名称)
		//第二种  如果存在 (RunJs/:JsName)
		//局_公共函数名 := c.Param("JsName") //取url内的参数
		//在检测公共函数的名字 [公共函数名]
		var 有权限 bool
		if strings.HasPrefix(局_接口名称, "RunJs/") {
			局_接口名称 = c.Param("JsName") //取url内的参数
			//查找到,且不为空才有权限
			有权限 = strings.Index(DB_LinksToken.Key, "["+局_接口名称+"]") != -1 && 局_接口名称 != ""
		} else {
			有权限 = strings.Index(DB_LinksToken.Key, "("+局_接口名称+")") != -1
		}

		if !有权限 {
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
		c.Set("DB_LinksToken", DB_LinksToken)
		global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Updates(map[string]interface{}{"LastTime": int(time.Now().Unix()), "Ip": c.ClientIP()})
		// 继续处理请求
		c.Next()
	}
}
