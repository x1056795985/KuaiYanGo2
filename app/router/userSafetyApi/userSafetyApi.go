package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"server/app/router/middleware"
)

type AllRouter struct {
}

func (r *AllRouter) InitWebApiRouter(router *gin.RouterGroup) {
	//注入路由分发函数指针到中间件包,避免循环依赖

	middleware.F解密Api名称 = 解密Api名称
	// 先注册中间件,再注册路由,否则中间件不生效
	局_安全api := router.Group("/Api")
	局_安全api.Use(middleware.UserApi检查数据库连接()) //检查数据库是否连接,连接后才可以使用用户Api,不然大量报错
	局_安全api.Use(middleware.X系统关闭则响应())       //检查系统是否已关闭,如果关闭直接返回
	局_安全api.Use(middleware.C初始化上下文())        //检查读取应用信息,初始化上下文变量
	// 必须放在可能 Abort 的中间件之前，才能在请求栈回退时统一写出响应。
	局_安全api.Use(middleware.C处理响应数据())
	局_安全api.Use(middleware.J检查黑名单())          //检查来源ip是否黑名单
	局_安全api.Use(middleware.UserApi无Token解密()) ///鉴权中间件 检查 token  单独优先处理
	局_安全api.Use(middleware.UserApi解密())       ///鉴权中间件 检查 token
	局_安全api.Use(middleware.J解密Api名称())
	局_安全api.POST("", F分发请求)
}
