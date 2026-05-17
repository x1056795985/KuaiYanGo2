package init

import (
	"errors"
	"time"

	"server/global"
	"server/new/app/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitGormMysql 初始化数据库并产生数据库全局变量
func InitGormMysql() (*gorm.DB, error) {
	m := global.GVA_CONFIG.Mysql
	if m.Dbname == "" {
		return nil, errors.New("数据库名称不能为空")
	}

	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(),
		DefaultStringSize:         191,
		SkipInitializeWithVersion: false,
	}

	if db, err := gorm.Open(mysql.New(mysqlConfig), (&utils.DbConfig{}).Config(m.Prefix)); err != nil {
		return nil, err
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(100 * time.Second)
		sqlDB.SetConnMaxIdleTime(90 * time.Second)
		return db, nil
	}
}
