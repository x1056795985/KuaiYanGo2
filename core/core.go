package core

import (
	"EFunc/utils"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
	"server/Service/Ser_Cron"
	"server/Service/Ser_TaskPool"
	"server/api/Admin/KuaiYan"
	"server/core/internal"
	"server/global"
	utils2 "server/utils"
)

// InitZap 日志系统初始化
// Author [SliverHorn](https://github.com/SliverHorn)
func InitZap() (logger *zap.Logger) {
	// 判断是否有Director文件夹  没有就创建
	if ok, _ := utils2.PathExists(global.GVA_CONFIG.Zap.Director); !ok {
		fmt.Printf("create %v directory\n", global.GVA_CONFIG.Zap.Director)
		_ = os.Mkdir(global.GVA_CONFIG.Zap.Director, os.ModePerm)
	}
	//获取日志配置信息
	cores := internal.Zap.GetZapCores()
	//创建日志记录器 logger
	logger = zap.New(zapcore.NewTee(cores...))
	//判断全局日志配饰 是否显示文件方法行号信息
	if global.GVA_CONFIG.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	//返回创建的日志记录器
	return logger
}

// InitViper //
// 新建并初始化配置读写器赋值给全局变量GVA_Viper 并把配置信息反序列化到全局文件
func InitViper() *viper.Viper {

	//新建一个viper 配置读写器

	v := viper.New()
	//写入默认值
	v.SetDefault("Port", 18888)
	//==================验证码默认配置
	v.SetDefault("captcha.open-captcha", 1)            //设置验证码默认ip防暴次数
	v.SetDefault("captcha.open-captcha-timeout", 3600) //防暴时间 被爆破后开启验证秒数
	v.SetDefault("captcha.img-height", 80)             //设置验证码高度
	v.SetDefault("captcha.img-width", 240)             //设置验证码宽度
	v.SetDefault("captcha.Key-long", 4)                //设置验证码长
	v.SetDefault("captcha.Type", 1)                    //设置验证码 类型   mark 后期类型拓展滑动验证码
	//==================日志默认配置
	v.SetDefault("zap.director", "log")                            //设置日志文件目录
	v.SetDefault("zap.encode-level", "LowercaseColorLevelEncoder") //设置日志文件编码
	v.SetDefault("zap.format", "console")                          //设置是否替换系统日志
	v.SetDefault("zap.level", "info")                              //设置是否替换系统日志
	v.SetDefault("zap.log-in-console", "true")                     //设置是否输出到控制台
	v.SetDefault("zap.max-age", 0)                                 //设置日志最大数量
	v.SetDefault("zap.show-line", true)                            //设置显示代码行号
	v.SetDefault("zap.stacktrace-key", "stacktrace")               //设置显示栈名
	//==================数据库默认配置
	v.SetDefault("mysql.Config", "")
	v.SetDefault("mysql.Dbname", "")
	v.SetDefault("mysql.LogMode", "error")
	v.SetDefault("mysql.LogZap", true)
	v.SetDefault("mysql.MaxIdleConns", 10)
	v.SetDefault("mysql.MaxOpenConns", 100)
	v.SetDefault("mysql.Path", "")
	v.SetDefault("mysql.Port", "3306")
	v.SetDefault("mysql.Prefix", "")
	v.SetDefault("mysql.Singular", false)
	v.SetDefault("mysql.Username", "")
	global.GVA_CONFIG.Q取运行目录 = utils.C程序_取运行目录()
	if runtime.GOOS == "windows" {
		global.GVA_CONFIG.Q取运行目录 = "."
	}
	path := global.GVA_CONFIG.Q取运行目录 + "/config.json" //设置文件目录   //注意设置 ./config.json  宝塔写文件不会写运行目录 文件会在 /www/server/panel 文件夹

	//设置路径
	v.SetConfigFile(path)
	v.SetConfigType("json")
	var err error
	//读入配置信息
	if utils.W文件_是否存在(path) {
		err = v.ReadInConfig()
		//如果err不动于空  就说明有问题,输出错误信息
		if err != nil {
			panic(fmt.Errorf("读入配置文件失败: %s \n", err))
		}
	} else {
		err = v.WriteConfig()
	}
	//viper支持监听配置文件，并会在配置文件发生变化，重新读取配置文件和回调函数，这样可以避免每次配置变化时，都需要重启启动应用的麻烦。
	// 监听配置文件 发生手动操作的变化自动读取
	//v.WatchConfig()

	////设置配置更新时处理函数   2023/4/23  发现自动回连续读取两次配置, 导致数据不正确, 停止自动读取,改为手动读取
	//v.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("配置文件已更改:", e.Name)
	//	//重新读取配置反序列化到全局配置结构里  global.GVA_CONFIG  config.Server   失败输出错误
	//	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
	//		fmt.Println("配置文件反序列化失败2:", err)
	//	}
	//})

	//读取配置反序列化到全局配置结构里  global.GVA_CONFIG  config.Server   失败输出错误
	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
		fmt.Println("配置文件反序列化失败:", err)
	}

	if global.GVA_CONFIG.Port == 0 {
		global.GVA_CONFIG.Port = 18888
	}

	return v
}

// InitCron定时任务 //
// 新建Cron定时任务,并插入需要初始化的任务
func InitCron定时任务() {
	global.Cron定时任务.Cron = cron.New(cron.WithSeconds()) //这里设置的 6位表达式 秒级

	//1  "0 0 0 * * *"  每天的 0点启动  * 通配符可以匹配任何数字
	//2  "*/5 * * * * *" 表示每隔5秒钟执行一次
	//3  "*/1 * * * *"  表示每隔1分钟执行一次，比秒级别解析器少了个*
	//4  "30 * * * *"  分钟域为 30，其他域都是*表示任意。每30分触发
	//5  "30 3-6,20-23 * * *"：分钟域为 30，小时域的3-6,20-23表示 3 点到 6 点和 20 点到 23 点。每小时的30 分钟触发.
	//6  "0 0 0 * * ?"  表示每天0点执行一次
	//7  "0 0 1 1 * ?"  表示每月1号凌晨1点执行一次
	//8  "0 1,2,3 * * * ?" 表示在1分，2分，3分执行一次
	//9  "0 0 0,1,2 * * ?" 表示每天的0点，1点，2点执行一次
	global.Cron定时任务.T添加任务("在线列表定时注销已过期", "0 */1 * * * *", Ser_Cron.Corn_在线列表定时注销已过期) //每分钟执行一次
	global.Cron定时任务.T添加任务("在线列表定时删除已过期", "0 */1 * * * *", Ser_Cron.Corn_在线列表定时删除已过期) //每分钟执行一次
	global.Cron定时任务.T添加任务("在线列表定时删除已过期", "0 0 0 * * ?", Ser_TaskPool.Task数据删除过期)     //每天0点执行一次
	global.Cron定时任务.T添加任务("快验心跳", "0 */5 * * * *", KuaiYan.K快验心跳)                    //5分钟心跳执行一次
	global.Cron定时任务.Cron.Start()
}
