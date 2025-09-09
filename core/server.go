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
	管理员后台运行地址:http://127.0.0.1%s/%s
	代理后台运行地址:http://127.0.0.1%s/%s
	web用户中心运行地址:http://127.0.0.1%s/user/{AppId}
`, global.X系统信息.B版本号当前, 端口, 端口, global.GVA_Viper.GetString("管理入口"), 端口, global.GVA_Viper.GetString("代理入口"), 端口)
	fmt.Printf("是否有读写文件权限:%v\n", utils2.X系统_权限检测())
	global.Q快验.C初始化配置(string(utils.B编码_BASE64解码("eyJBcHBXZWIiOiJodHRwOi8va3lhcGkuOXc5OS5jbi9BcGk/QXBwSWQ9MTAwMDEiLCJDcnlwdG9LZXlQdWJsaWMiOiItLS0tLUJFR0lOIFBVQkxJQyBLRVktLS0tLVxuTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDM0hib1NYQ0trcUdWaDBsaEt6cFNwaDFYRVxuNUtNYnBtYUhBTzIyNzdnOGtZaVVRRk5XU21PNlZ0RjJlcENKTURleTJjSVZEMk9OUnFZSkxLeWdYbHpkSGtZNlxuQU1Oa3AweTh5elVMQUlUSkgyOTkwbzJTb3VPdzdYQlBPNTN0Nk9URVJSTG92Lzh5YTVMNXJaK1NzM2R4RHNJVFxudkZqd0drYzZ5QlBBVFJKM1N3SURBUUFCXG4tLS0tLUVORCBQVUJMSUMgS0VZLS0tLS1cbiIsIkNyeXB0b1R5cGUiOjN9")))

	if global.GVA_Viper.GetInt("系统模式") != 1056795985 {
		ret, err := req.C().EnableInsecureSkipVerify().R().Get(string(utils.B编码_BASE64解码("aHR0cHM6Ly9hcGl3Zy45dzk5LmNu")))
		if err == nil && ret.GetHeader("X-Ca-Error-Code") != "" {
			//需要放在这里,不然无法执行 //初始化取token,放中间件内了,可以自验证
			global.Q快验.SetAppWeb(string(utils.B编码_BASE64解码("aHR0cHM6Ly9hcGl3Zy45dzk5LmNu")))
			global.Q快验.J集_连接方式 = 1
		}
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
	err := global.GVA_Gin.ListenAndServe() //执行到此处会暂停,直到系统退出
	if err != nil {
		global.GVA_LOG.Error(err.Error())
	}

	//global.GVA_Gin.Shutdown()  这句话可以停止侦听关闭端口
	time.Sleep(10 * time.Second) //延迟10秒在关闭主程序,因为可能是关闭了gin 后面还要输出日志重启
	fmt.Println("主线程等待10秒后自然关闭,")
}
