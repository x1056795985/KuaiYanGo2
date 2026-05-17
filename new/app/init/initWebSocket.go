package init

import (
	"github.com/gin-gonic/gin"
)

// InitWebSocket 初始化WebSocket服务
func InitWebSocket() {
}

// InitAll 初始化所有服务
func InitAll(c *gin.Context) {
	InitDbTables(c)
}
