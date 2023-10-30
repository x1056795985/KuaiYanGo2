package InitDB

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"server/Service/Ser_Init"
	DB "server/structs/db"
	"server/utils"

	"server/global"
	"server/structs/Http/request"
	"server/structs/Http/response"
)

// DBApi  初始化数据库api 具体实现位置
type DBApi struct{}

// CheckDB
// @Tags     CheckDB
// @Summary  检测用户数据库
// @Produce  application/json
// @Success  200  {object}  response.Response{data=map[string]interface{},msg=string}  "初始化用户数据库"
// @Router   /init/checkdb [post]
func (i *DBApi) CheckDB(c *gin.Context) {
	var (
		message  = "前往初始化数据库"
		needInit = true
	)
	var 局_数量 int64
	if global.GVA_DB == nil {
		goto 结果
	}

	global.GVA_DB.Model(DB.DB_Admin{}).Count(&局_数量)

	if 局_数量 >= 1 { // 判断是否存在数据  如果没存在就返回假
		message = "数据库无需初始化"
		needInit = false
		goto 结果
	}

结果:
	//global.GVA_LOG.Info(message)
	//响应成功 并传入消息和数据
	response.OkWithDetailed(gin.H{"needInit": needInit, "ServerName": global.GVA_CONFIG.X系统设置.X系统名称, "Filing": global.GVA_CONFIG.X系统设置.B备案号}, message, c)

}

// InitDB
// @Tags     InitDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Param    data  body      request.InitDB                  true  "初始化数据库参数"
// @Success  200   {object}  response.Response{data=string}  "初始化用户数据库"
// @Router   /init/initdb [post]
func (i *DBApi) InitDB(c *gin.Context) {
	if !utils.X系统_权限检测() {
		response.FailWithMessage("进程权限不足,请前往宝塔设置权限777,读取写入都要", c)
		return
	}

	var J_数量 int64
	if global.GVA_DB != nil {
		global.GVA_DB.Model(DB.DB_Admin{}).Count(&J_数量)
		if J_数量 != 0 {
			global.GVA_LOG.Error("已存在数据库配置!")
			response.FailWithMessage("已存在数据库配置", c)
			return
		}
	}
	var 请求 request.InitDB //声明一个 initDB结果剖提
	// 反序列化参数到结构体
	if err := c.ShouldBindJSON(&请求); err != nil {
		global.GVA_LOG.Error("参数校验不通过!", zap.Error(err))
		response.FailWithMessage("参数校验不通过", c) //响应
		return
	}

	global.GVA_CONFIG.Mysql.Username = 请求.UserName
	global.GVA_CONFIG.Mysql.Password = 请求.Password
	global.GVA_CONFIG.Mysql.Path = 请求.Host
	global.GVA_CONFIG.Mysql.Port = 请求.Port
	global.GVA_CONFIG.Mysql.Dbname = 请求.DBName
	global.GVA_CONFIG.Mysql.Config = "charset=utf8mb4&parseTime=True&loc=Local"
	global.GVA_CONFIG.Mysql.MaxIdleConns = 10
	global.GVA_CONFIG.Mysql.MaxOpenConns = 100
	global.GVA_CONFIG.Mysql.LogMode = "error"

	global.GVA_DB = Ser_Init.InitGormMysql() // gorm连接数据库  Gorm参考资料https://www.cnblogs.com/davis12/p/16365213.html

	//开始创建数据库
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("自动创建数据库失败!")
		response.FailWithMessage("自动创建数据库失败，确认已启动Mysql,检查参数后再次进行初始化\r\n", c) //响应
		return
	}
	global.GVA_Viper.Set("Mysql.Username", 请求.UserName)
	global.GVA_Viper.Set("Mysql.Password", 请求.Password)
	global.GVA_Viper.Set("Mysql.Path", 请求.Host)
	global.GVA_Viper.Set("Mysql.Port", 请求.Port)
	global.GVA_Viper.Set("Mysql.Dbname", 请求.DBName)
	global.GVA_Viper.Set("Mysql.Config", "charset=utf8mb4&parseTime=True&loc=Local")
	global.GVA_Viper.Set("Mysql.MaxIdleConns", 10)
	global.GVA_Viper.Set("Mysql.MaxOpenConns", 100)
	global.GVA_Viper.Set("Mysql.LogMode", "error")
	global.GVA_Viper.WriteConfig()

	Ser_Init.InitDbTables()  //自动创建迁移表
	Ser_Init.InitDbTable数据() //自动创建数据

	global.GVA_Viper.SetConfigFile(global.GVA_CONFIG.Q取运行目录 + "/config.json")
	global.GVA_Viper.SetConfigType("json")
	err := global.GVA_Viper.WriteConfig()
	if err != nil {
		response.OkWithMessage("自动创建数据库成功,写配置文件失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("自动创建数据库成功,默认账密都是admin", c)
	return
}
