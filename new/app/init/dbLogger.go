package init

import (
	"fmt"
	"server/new/app/global"

	"gorm.io/gorm/logger"
)

type dbWriter struct {
	logger.Writer
}

// NewDbWriter writer 构造函数
func NewDbWriter(w logger.Writer) *dbWriter {
	return &dbWriter{Writer: w}
}

// Printf 格式化打印日志
func (w *dbWriter) Printf(message string, data ...interface{}) {
	global.GVA_LOG.Print(fmt.Sprintf(message, data...))
}
