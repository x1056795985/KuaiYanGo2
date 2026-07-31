package utils

import (
	"github.com/gin-gonic/gin"
	"server/app/models/common"
)

// 请求上下文包,存放请求生命周期上下文结构体
// 独立为子包避免 userSafetyApi 和 response 之间的循环依赖
// 请求上下文 贯穿整个请求生命周期(解密→handler→加密响应)
// 通过 c.Set("ctx", ctx) 一次注入, handler 内通过 取上下文(c) 获取

const ctxKey = "安全api上下文"

// 从 gin.Context 取出请求上下文, 不存在则返回 nil
func Q取上下文(c *gin.Context) *common.Q请求_上下文 {
	v, ok := c.Get(ctxKey)
	if !ok {
		c.Set(ctxKey, &common.Q请求_上下文{}) //如果没有就返回一个空的,保证一定可以取到
		v, _ = c.Get(ctxKey)
	}
	return v.(*common.Q请求_上下文)
}
func Z置上下文(c *gin.Context, ctx *common.Q请求_上下文) {
	c.Set(ctxKey, ctx)
}
