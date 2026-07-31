package bootstrap

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"

	"server/app/global"
	"server/app/logic/admin/KuaiYanUpdater"
	"server/app/router/middleware"
	utils2 "server/app/utils"
)

// F服务_运行 创建并启动 HTTP 服务，直到服务关闭后返回。
func F服务_运行(路由 http.Handler) {
	局_端口 := fmt.Sprintf(":%d", global.GVA_CONFIG.Port)
	global.GVA_Gin = &http.Server{
		Addr:           局_端口,
		Handler:        路由,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	服务_打印启动信息(局_端口)
	服务_初始化快验()
	服务_延迟加载快验令牌()
	KuaiYanUpdater.B宝塔_修改项目信息pid()

	局_错误 := global.GVA_Gin.ListenAndServe()
	if 局_错误 != nil && !errors.Is(局_错误, http.ErrServerClosed) {
		global.GVA_LOG.Println(局_错误.Error())
	}

	time.Sleep(10 * time.Second)
	fmt.Println("主线程等待10秒后自然关闭")
}

func 服务_打印启动信息(port string) {
	fmt.Printf(`
	欢迎使用 飞鸟快验后台管理
	当前版本:v.%s
	后台端口:%s
	管理员后台运行地址:http://127.0.0.1%s/%s
	代理后台运行地址:http://127.0.0.1%s/%s
	web用户中心运行地址:http://127.0.0.1%s/user/{AppId}
	webSocket链接地址:ws://127.0.0.1%s/ws/{Token}
`, global.X系统信息.B版本号当前, port, port, global.GVA_Viper.GetString("管理入口"), port, global.GVA_Viper.GetString("代理入口"), port, port)
	fmt.Printf("是否有读写文件权限:%v\n", utils2.X系统_权限检测())
}

func 服务_初始化快验() {
	global.Q快验.C初始化配置(string(B编码_BASE64解码("eyJBcHBXZWIiOiJodHRwOi8va3lhcGkuOXc5OS5jbi9BcGk/QXBwSWQ9MTAwMDEiLCJDcnlwdG9LZXlQdWJsaWMiOiItLS0tLUJFR0lOIFBVQkxJQyBLRVktLS0tLVxuTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDM0hib1NYQ0trcUdWaDBsaEt6cFNwaDFYRVxuNUtNYnBtYUhBTzIyNzdnOGtZaVVRRk5XU21PNlZ0RjJlcENKTURleTJjSVZEMk9OUnFZSkxLeWdYbHpkSGtZNlxuQU1Oa3AweTh5elVMQUlUSkgyOTkwbzJTb3VPdzdYQlBPNTN0Nk9URVJSTG92Lzh5YTVMNXJaK1NzM2R4RHNJVFxudkZqd0drYzZ5QlBBVFJKM1N3SURBUUFCXG4tLS0tLUVORCBQVUJMSUMgS0VZLS0tLS1cbiIsIkNyeXB0b1R5cGUiOjN9")))
	if global.GVA_Viper.GetInt("系统模式") == 1056795985 {
		return
	}
	局_接口地址 := string(B编码_BASE64解码("aHR0cHM6Ly9hcGl3Zy45dzk5LmNu"))
	局_响应, 局_错误 := req.C().EnableInsecureSkipVerify().R().Get(局_接口地址)
	if 局_错误 == nil && 局_响应.GetHeader("X-Ca-Error-Code") != "" {
		global.Q快验.SetAppWeb(局_接口地址)
		global.Q快验.J集_连接方式 = 1
	}
}

func 服务_延迟加载快验令牌() {
	if global.GVA_DB == nil {
		return
	}
	go func() {
		time.Sleep(5 * time.Second)
		middleware.D读取缓存Token()
		if global.GVA_Viper.GetInt("duid") > 0 {
			global.Q快验.Z置代理标志(global.GVA_Viper.GetInt("duid"))
		}
	}()
}
