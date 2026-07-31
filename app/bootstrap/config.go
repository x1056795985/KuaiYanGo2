package bootstrap

import (
	. "EFunc/utils"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"

	"server/app/global"
)

// P配置_初始化 创建配置读取器、写入默认配置并更新全局配置。
func P配置_初始化() *viper.Viper {
	局_配置器 := viper.New()
	配置_设置默认值(局_配置器)

	global.GVA_CONFIG.Q取运行目录 = C程序_取运行目录()
	if runtime.GOOS == "windows" {
		global.GVA_CONFIG.Q取运行目录 = "."
	}
	局_配置路径 := filepath.Join(global.GVA_CONFIG.Q取运行目录, "config.json")
	局_配置器.SetConfigFile(局_配置路径)
	局_配置器.SetConfigType("json")

	var 局_错误 error
	if W文件_是否存在(局_配置路径) {
		局_错误 = 局_配置器.ReadInConfig()
	} else {
		局_错误 = 局_配置器.WriteConfigAs(局_配置路径)
	}
	if 局_错误 != nil {
		panic(fmt.Errorf("读写配置文件失败: %w", 局_错误))
	}
	if 局_错误 = 局_配置器.Unmarshal(&global.GVA_CONFIG); 局_错误 != nil {
		panic(fmt.Errorf("配置文件反序列化失败: %w", 局_错误))
	}
	if global.GVA_CONFIG.Port == 0 {
		global.GVA_CONFIG.Port = 18888
	}
	return 局_配置器
}

func 配置_设置默认值(config *viper.Viper) {
	config.SetDefault("管理入口", "Admin")
	config.SetDefault("代理入口", "Agent")
	config.SetDefault("Port", 18888)
	config.SetDefault("系统模式", 0)

	config.SetDefault("captcha.open-captcha", 1)
	config.SetDefault("captcha.open-captcha-timeout", 3600)
	config.SetDefault("captcha.img-height", 80)
	config.SetDefault("captcha.img-width", 240)
	config.SetDefault("captcha.Key-long", 4)
	config.SetDefault("captcha.Type", 1)

	config.SetDefault("mysql.Config", "")
	config.SetDefault("mysql.Dbname", "")
	config.SetDefault("mysql.LogMode", "error")
	config.SetDefault("mysql.MaxIdleConns", 10)
	config.SetDefault("mysql.MaxOpenConns", 100)
	config.SetDefault("mysql.Path", "")
	config.SetDefault("mysql.Port", "3306")
	config.SetDefault("mysql.Prefix", "")
	config.SetDefault("mysql.Singular", false)
	config.SetDefault("mysql.Username", "")
}
