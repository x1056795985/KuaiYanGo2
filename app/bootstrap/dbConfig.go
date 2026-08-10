package bootstrap

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"server/app/global"
	"server/app/monitoring"
)

type DBBASE interface {
	GetLogMode() string
}

// DbConfig gorm 配置
type DbConfig struct{}

func (g *DbConfig) Config(表前缀 string) *gorm.Config {
	局_配置 := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   表前缀,
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	局_默认日志器 := logger.New(NewDbWriter(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	var 局_日志模式 DBBASE
	局_日志模式 = &global.GVA_CONFIG.Mysql
	switch 局_日志模式.GetLogMode() {
	case "silent", "Silent":
		局_配置.Logger = monitoring.C初始化Gorm日志器(局_默认日志器.LogMode(logger.Silent), 200*time.Millisecond)
	case "error", "Error":
		局_配置.Logger = monitoring.C初始化Gorm日志器(局_默认日志器.LogMode(logger.Error), 200*time.Millisecond)
	case "warn", "Warn":
		局_配置.Logger = monitoring.C初始化Gorm日志器(局_默认日志器.LogMode(logger.Warn), 200*time.Millisecond)
	case "info", "Info":
		局_配置.Logger = monitoring.C初始化Gorm日志器(局_默认日志器.LogMode(logger.Info), 200*time.Millisecond)
	default:
		局_配置.Logger = monitoring.C初始化Gorm日志器(局_默认日志器.LogMode(logger.Info), 200*time.Millisecond)
	}

	return 局_配置
}
