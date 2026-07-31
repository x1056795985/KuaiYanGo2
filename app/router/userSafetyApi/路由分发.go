package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/userSafetyApi"
	"server/app/controller/userSafetyApi/response"
	"server/app/models/constant"
	"server/app/utils"
)

// 解密Api名称 将md5加密的Api名称还原为明文
func 解密Api名称(c *gin.Context, Api string) (string, bool) {
	if len(J集_UserAPi路由_加密.J加密路由) == 0 {
		return Api, true
	}
	局_Api, ok := J集_UserAPi路由_加密.Q取md5APi名称(Api)
	if !ok {
		response.FailMsg(c, constant.Status_Api不存在, "API名称加密错误")
		c.Abort()
		return "", false
	}
	return 局_Api, true
}

// 根据Api名称查找路由表并执行对应handler,找到返回true
func F分发请求(c *gin.Context) {
	局_ctx := utils.Q取上下文(c)

	if 局_ctx.Api == "GetToken" { // 获取token 单独处理
		userSafetyApi.UserApi_GetToken(c)
		return
	}

	局_路由信息, ok := J集_UserAPi路由[局_ctx.Api]
	if !ok {
		response.FailMsg(c, constant.Status_Api不存在, 局_ctx.Api+",API不存在")
		return
	}

	局_路由信息.Z指向函数(c)
	return
}
