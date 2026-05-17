package utils

import (
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"server/global"
)

type DBBASE interface {
	GetLogMode() string
}

// DbConfig gorm 配置
type DbConfig struct{}

func (g *DbConfig) Config(表前缀 string) *gorm.Config {
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   表前缀,
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	_default := logger.New(NewDbWriter(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	var logMode DBBASE
	logMode = &global.GVA_CONFIG.Mysql
	switch logMode.GetLogMode() {
	case "silent", "Silent":
		config.Logger = _default.LogMode(logger.Silent)
	case "error", "Error":
		config.Logger = _default.LogMode(logger.Error)
	case "warn", "Warn":
		config.Logger = _default.LogMode(logger.Warn)
	case "info", "Info":
		config.Logger = _default.LogMode(logger.Info)
	default:
		config.Logger = _default.LogMode(logger.Info)
	}
	return config
}
