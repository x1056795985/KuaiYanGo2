package utils

import (
	"fmt"

	"gorm.io/gorm/logger"
	"server/global"
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
	var logZap bool
	logZap = global.GVA_CONFIG.Mysql.LogZap
	if logZap {
		global.GVA_LOG.Info(fmt.Sprintf(message+"\n", data...))
	} else {
		w.Writer.Printf(message, data...)
	}
}
