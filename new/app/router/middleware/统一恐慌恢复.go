package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"runtime/debug"
	"server/global"
)

// T统一恐慌恢复 全局恐慌恢复中间件
func T统一恐慌恢复() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				局_上报错误 := fmt.Sprintln("全局捕获错误:\n", err, "\n堆栈信息:\n", string(debug.Stack()))
				debug.PrintStack()
				log.Println("发生致命错误:", 局_上报错误)
				global.Q快验.Z置新用户消息(2, 局_上报错误)
			}
		}()
		c.Next()
	}
}
