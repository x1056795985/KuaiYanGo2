package core

import (
	"EFunc/utils"
	"fmt"
	"github.com/imroc/req/v3"
	"net/http"
	"server/Service/KuaiYanUpdater"
	"server/api/middleware"
	utils2 "server/utils"

	"server/global"
	"time"
)

func RunWindowsServer() {

	Router := InitRouters() //注册路由 和绑定 具体实现函数

	// 关键点【解决页面刷新404的问题】
	/*	Router.NoRoute(func(c *gin.Context) {
		c.String(404, "")
		return
	})*/

	端口 := fmt.Sprintf(":%d", global.GVA_CONFIG.Port) //:18888

	global.GVA_Gin = &http.Server{
		Addr:           端口,
		Handler:        Router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 保证文本顺序输出
	// In order to ensure that the text order output can be deleted
	time.Sleep(10 * time.Microsecond)
	//global.GVA_LOG.Info("web 服务器启动成功", zap.String("端口", 端口))

	fmt.Printf(`
	欢迎使用 飞鸟快验后台管理
	当前版本:v.%s
	后台端口:%s
	管理员后台运行地址:http://127.0.0.1%s/Admin
	代理后台运行地址:http://127.0.0.1%s/Agent
`, global.X系统信息.B版本号当前, 端口, 端口, 端口)
	fmt.Printf("是否有读写文件权限:%v\n", utils2.X系统_权限检测())

	ret, err := req.Get(string(utils.B编码_BASE64解码("aHR0cHM6Ly9reWFwaS45dzk5LmNu")))
	if err == nil && ret.String() != "" {
		global.Q快验.C初始化配置(string(utils.B编码_BASE64解码("eyJBcHBXZWIiOiJodHRwczovL2t5YXBpLjl3OTkuY24vQXBpP0FwcElkPTEwMDAxIiwiQ3J5cHRvS2V5UHVibGljIjoiLS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS1cbk1JR2ZNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0R05BRENCaVFLQmdRQzNIYm9TWENLa3FHVmgwbGhLenBTcGgxWEVcbjVLTWJwbWFIQU8yMjc3ZzhrWWlVUUZOV1NtTzZWdEYyZXBDSk1EZXkyY0lWRDJPTlJxWUpMS3lnWGx6ZEhrWTZcbkFNTmtwMHk4eXpVTEFJVEpIMjk5MG8yU291T3c3WEJQTzUzdDZPVEVSUkxvdi84eWE1TDVyWitTczNkeERzSVRcbnZGandHa2M2eUJQQVRSSjNTd0lEQVFBQlxuLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tXG4iLCJDcnlwdG9UeXBlIjozfQ==")))
	} else {
		//需要放在这里,不然无法执行 //初始化取token,放中间件内了,可以自验证
		global.Q快验.C初始化配置(string(utils.B编码_BASE64解码("eyJBcHBXZWIiOiJodHRwOi8va3lhcGkuZm5rdWFpeWFuLmNuL0FwaT9BcHBJZD0xMDAwMSIsIkNyeXB0b0tleVB1YmxpYyI6Ii0tLS0tQkVHSU4gUFVCTElDIEtFWS0tLS0tXG5NSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUUMzSGJvU1hDS2txR1ZoMGxoS3pwU3BoMVhFXG41S01icG1hSEFPMjI3N2c4a1lpVVFGTldTbU82VnRGMmVwQ0pNRGV5MmNJVkQyT05ScVlKTEt5Z1hsemRIa1k2XG5BTU5rcDB5OHl6VUxBSVRKSDI5OTBvMlNvdU93N1hCUE81M3Q2T1RFUlJMb3YvOHlhNUw1clorU3MzZHhEc0lUXG52Rmp3R2tjNnlCUEFUUkozU3dJREFRQUJcbi0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLVxuIiwiQ3J5cHRvVHlwZSI6M30=")))
	}

	if global.GVA_DB != nil {
		go func() { //启动web后,在协程内获取token,也可以解决自验证的问题,
			time.Sleep(5 * time.Second) //延迟5秒在在获取Token, 中间件获取可能导致,进入个人中心,获取验证码列表,可能因为速度太快还没获取token,报错验签失败,或加解密失败
			middleware.D读取缓存Token()
			if global.GVA_Viper.GetInt("duid") > 0 {
				global.Q快验.Z置代理标志(global.GVA_Viper.GetInt("duid"))
			}
		}()
	}
	KuaiYanUpdater.B宝塔_修改项目信息pid()
	err = global.GVA_Gin.ListenAndServe() //执行到此处会暂停,直到系统退出
	if err != nil {
		global.GVA_LOG.Error(err.Error())
	}

	//global.GVA_Gin.Shutdown()  这句话可以停止侦听关闭端口
	time.Sleep(10 * time.Second) //延迟10秒在关闭主程序,因为可能是关闭了gin 后面还要输出日志重启
	fmt.Println("主线程等待10秒后自然关闭,")
}
