package Ser_Init

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Init/internal"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_TaskPool"
	"server/Service/Ser_User"
	"server/api/Admin/App"
	"server/config"
	"server/global"
	"server/new/app/logic/common/appInfo"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/publicData"
	"server/new/app/logic/common/setting"
	dbm "server/new/app/models/db"

	"server/new/app/service"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"time"
)

// 初始化数据库并产生数据库全局变量
func InitGormMysql() (*gorm.DB, error) {
	m := global.GVA_CONFIG.Mysql
	//如果没有数据库名字,直接返回估计还没传配置,第一次启动
	if m.Dbname == "" {
		return nil, errors.New("数据库名称不能为空")
	}

	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}

	//连接并设置数据库连接池参数
	if db, err := gorm.Open(mysql.New(mysqlConfig), internal.Gorm.Config(m.Prefix)); err != nil {
		//链接失败了
		return nil, err
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns) //最大连接数
		sqlDB.SetMaxOpenConns(m.MaxOpenConns) //允许空闲数
		//设定数据库连接的最大生命周期 Mysql默认120秒 所以gorm 设置个比这个值小的数 防止断开连接时操作数据库失败
		sqlDB.SetConnMaxLifetime(100 * time.Second)
		return db, nil //返回连接好的db池
		//获取gorm db对象，其他包需要执行数据库查询的时候，只要通过	global.GVA_DB 获取db对象即可。
		//不用担心协程并发使用同样的db对象会共用同一个连接，db对象在调用他的方法的时候会从数据库连接池中获取新的连接
	}
}

// InitDbTables

func InitDbTables(c *gin.Context) {
	db := global.GVA_DB //全局变量赋值到局部

	//gorm:table_options 设置创建表强制为InnoDB引擎, 因为MyISAM不支持事务,回滚会失效所以要修改成InnoDB引擎,
	//参考地址 https://blog.csdn.net/qq_25436207/article/details/107533197

	// 分别迁移每个表来定位问题
	tables := []interface{}{
		// 系统模块表  数据库结构表
		DB.DB_PublicData{},
		DB.DB_PublicJs{},
		DB.DB_UserConfig{},

		DB.DB_Admin{},
		DB.DB_User{},
		DB.DB_LinksToken{},

		DB.DB_AppInfo{},
		// DB.DB_AppUser{}, //DB.DB_AppUser{},   //因为每个应用一个表 所以不在自动迁移里处理  只在创建应用时 创建 处理
		DB.DB_UserClass{},

		dbm.DB_KaClass{},
		DB.DB_Ka{},

		DB.DB_LogMoney{},
		DB.DB_LogLogin{},
		DB.DB_LogUserMsg{},
		DB.DB_LogRMBPayOrder{},
		DB.DB_LogKa{},
		DB.DB_LogRiskControl{},
		DB.DB_LogVipNumber{},
		DB.DB_LogAgentOtherFunc{},

		//	代理相关
		DB.Db_Agent_Level{},
		DB.Db_Agent_卡类授权{},
		DB.Db_Agent_库存日志{},
		DB.Db_Agent_库存卡包{},

		dbm.DB_Setting{},
		dbm.DB_Blacklist{},
		dbm.DB_Cron{},
		dbm.DB_Cron_log{},
		dbm.DB_PromotionCode{},
		dbm.DB_KaClassUpPrice{},
		dbm.DB_AppInfoWebUser{},
		dbm.DB_AppPromotionConfig{},

		dbm.DB_CpsInfo{},
		dbm.DB_CpsShortUrl{},
		dbm.DB_CpsInvitingRelation{},
		dbm.DB_CpsUser{},
		dbm.DB_CpsPayOrder{},

		dbm.DB_CheckInUser{},
		dbm.DB_CheckInScoreLog{},
		dbm.DB_CheckInLog{},

		//统计数据用的表
		dbm.DB_TongJiZaiXian{},

		//任务池数据库
		DB.TaskPool_类型{},
		DB.TaskPool_队列{},
		DB.DB_TaskPoolData{}, //任务池数据库 放到最后 业务字段可能和预设不同
	}
	for _, table := range tables {
		//AutoMigrate 自动迁移功能（如不存在会自动创建一个表）传过来的是  结构体HelloWorld实例的指针地址
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(table); err != nil {
			global.GVA_LOG.Error("表创建失败", zap.String("table", fmt.Sprintf("%T", table)), zap.Error(err))
			//os.Exit(0)                                                //结束程序  //20250703 不能结束,如果客户修改过字段类型,会导致进不去程序 比如修改任务池数据提交字段类型
		}
	}
	//global.GVA_LOG.Info("register table success(创建表成功)") //日志消息 表创建成功
	InitDbTable数据(c) //初始化数据

}

func InitDbTable数据(c *gin.Context) {
	db := global.GVA_DB //全局变量赋值到局部
	数据库兼容旧版本(c)         //需要先兼容, 迁移配置数据 才能使用setting
	局_例子记录 := setting.Q例子写出记录()
	if db == nil {
		return
	}
	//检查 admin表是否有账号============================================开始
	var 局_数量 int64
	db.Model(DB.DB_Admin{}).Count(&局_数量)
	if 局_数量 == 0 {
		entities := []DB.DB_Admin{{
			Id:            1,
			User:          "admin",
			PassWord:      utils2.BcryptHash("admin"),
			Phone:         "",
			Email:         "",
			Qq:            "",
			SuperPassWord: utils2.BcryptHash("admin"),
			Status:        1,
			Authority:     "All",
			AgentDiscount: 100,
		},
		}
		global.GVA_DB.Create(&entities)
	}
	//=============================================结束

	//检查 用户表  是否有数据没有
	局_例子版本 := 1
	if 局_例子记录.DbUser < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_User{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_User.New用户信息("test0001", "test0001", "test0001test0001", "10001", "10001@qq.com", "", "127.0.0.1", "", 0, 0, 0, "")
		}
		局_例子记录.DbUser = 局_例子版本
	}

	//-============================================结束==========================
	//检查 DB_AppInfo表是否有应用如果没有插入测试应用============================================
	局_例子版本 = 1
	if 局_例子记录.DbAppinfo < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_AppInfo{}).Count(&局_数量)
		if 局_数量 == 0 {
			_ = appInfo.L_appInfo.NewApp信息(c, 10001, 1, "演示对接账密限时Rsa交换密匙")
			Ser_AppUser.New用户信息(10001, 1, "测试绑定", 1, time.Now().Unix(), 11.02, 0, "", 0)
			卡类ID, _ := Ser_KaClass.KaClass创建New(10001, "天卡", "Y30", 2592000, 2592000, 0.01, 1.01, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
			卡类ID, _ = Ser_KaClass.KaClass创建New(10001, "月卡", "Y30", 2592000, 2592000, 0.01, 1.01, 100, 100, 0, 1, 25, 1, 1, 1, 1)
			卡信息, _ := Ser_Ka.Ka单卡创建(卡类ID, Ser_Admin.Id取User(1), "演示创建", "", 0)
			卡信息, _ = Ser_Ka.Ka单卡创建(卡类ID, Ser_Admin.Id取User(1), "演示创建可追回卡号", "", 0)
			ka.L_ka.K卡号充值_事务(c, 10001, 卡信息.Name, "test0001", "")
			_ = appInfo.L_appInfo.NewApp信息(c, 10002, 3, "演示对接卡号限时RSA通讯")
			卡类ID, _ = Ser_KaClass.KaClass创建New(10002, "天卡", "Y01", 86400, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
			卡类ID, _ = Ser_KaClass.KaClass创建New(10002, "周卡", "Y01", 604800, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
		}
		局_例子记录.DbAppinfo = 局_例子版本
	}
	//-============================================结束==========================

	//检查 余额充值订单 是否有应用如果没有插入测试应用============================================
	局_例子版本 = 1
	if 局_例子记录.DbLogrmbpayorder < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Count(&局_数量)
		if 局_数量 == 0 {
			订单创建, _ := Ser_RMBPayOrder.Order订单创建(1, 1, 0.01, "支付宝PC", "演示数据", "127.0.0.1", 0, "")
			Ser_RMBPayOrder.Order更新订单状态(订单创建.PayOrder, Ser_RMBPayOrder.D订单状态_成功)

			订单创建, _ = Ser_RMBPayOrder.Order订单创建(1, 1, 0.01, "微信支付", "演示数据", "127.0.0.1", 0, "")
			Ser_RMBPayOrder.Order更新订单状态(订单创建.PayOrder, Ser_RMBPayOrder.D订单状态_成功)

			订单创建, _ = Ser_RMBPayOrder.Order订单创建(1, 1, 0.01, "管理员手动充值", "演示数据", "127.0.0.1", 0, "")
			Ser_RMBPayOrder.Order更新订单状态(订单创建.PayOrder, Ser_RMBPayOrder.D订单状态_成功)
			订单创建, _ = Ser_RMBPayOrder.Order订单创建(1, 1, 0.01, "微信支付", "演示数据", "127.0.0.1", 0, "")
			Ser_RMBPayOrder.Order更新订单状态(订单创建.PayOrder, Ser_RMBPayOrder.D订单状态_等待支付)
			订单创建, _ = Ser_RMBPayOrder.Order订单创建(1, 1, 0.01, "支付宝PC", "演示数据", "127.0.0.1", 0, "")
			Ser_RMBPayOrder.Order更新订单状态(订单创建.PayOrder, Ser_RMBPayOrder.D订单状态_退款成功)
			go Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "管理员操作退款,余额充值订单:"+订单创建.PayOrder+",扣除用户已充值余额"+"|新余额≈"+utils.Float64到文本(0.01, 2), utils.Float64取负值(订单创建.Rmb))

			Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "看你长得帅,收费", -0.05)

			订单创建, _ = Ser_RMBPayOrder.Order订单创建(1, 1, 0.01, "微信支付", "演示数据", "127.0.0.1", 0, "")
			Ser_RMBPayOrder.Order更新订单状态(订单创建.PayOrder, Ser_RMBPayOrder.D订单状态_退款失败)
		}
		局_例子记录.DbLogrmbpayorder = 局_例子版本
	}
	//-============================================结束==========================
	//检查 余额日志  是否有数据没有
	局_例子版本 = 1
	if 局_例子记录.DbLogmoney < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_LogMoney{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", -0.01)
			Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", 0.01)
		}
		局_例子记录.DbLogmoney = 局_例子版本
	}
	//-============================================结束==========================
	//检查 积分点数  是否有数据没有
	局_例子版本 = 1
	if 局_例子记录.DbLogvipnumber < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_LogVipNumber{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示积分效果", -0.01, 10001, 1)
			Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示积分效果", 0.01, 10001, 1)
			Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示点数效果", 1, 10001, 2)
			Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示点数效果", -1, 10001, 2)
		}
		局_例子记录.DbLogvipnumber = 局_例子版本
	}
	//-============================================结束==========================

	//检查 公共变量表  是否有数据没有====================================================
	局_例子版本 = 1
	if 局_例子记录.DbPublicdata < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_PublicData{}).Count(&局_数量)
		if 局_数量 == 0 {

			_ = publicData.L_publicData.C创建(&gin.Context{}, DB.DB_PublicData{
				AppId: 1,
				Type:  3,
				Name:  "测试逻辑开关",
				Value: "1",
			})
			_ = publicData.L_publicData.C创建(&gin.Context{}, DB.DB_PublicData{
				AppId: 1,
				Type:  1,
				Name:  "系统名称",
				Value: "飞鸟快验应用管理后台",
			})
		}
		局_例子记录.DbPublicdata = 局_例子版本
	}
	//-============================================结束==========================

	//检查 公共js表  是否有数据没有
	插入公共js例子() //太长了,单独写个函数
	//-============================================结束==========================
	//检查 任务类型  是否有数据没有
	局_例子版本 = 1
	if 局_例子记录.Taskpool < 局_例子版本 {
		global.GVA_DB.Model(DB.TaskPool_类型{}).Count(&局_数量)
		if 局_数量 == 0 {
			_ = Ser_TaskPool.Task类型创建("测试任务1", "hook模板_任务创建入库前", "", "", "", "")
		}
		局_例子记录.Taskpool = 局_例子版本
	}
	//-============================================结束==========================

	//检查 任务类型  是否有数据没有
	局_例子版本 = 1
	if 局_例子记录.DbLogusermsg < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_LogUserMsg{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_Log.Log_写用户消息(3, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", "建议做个自动赚钱的功能,启动软件后,微信余额就蹭蹭涨", "127.0.0.1")
			Ser_Log.Log_写用户消息(2, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", `捕获到异常bug文件名:EDV8FCC.tmp句柄数:508,ExceptionText：运行时出错!\r\n\r\n错误代码：0\r\n\r\n错误信息：分配 1073741832 字节内存失败!\r\n0, 0\r\n\r\nCallStack:\r\n 0x024B7B4C\r\n  0x10063260\r\n   0x024A0410\r\n    0x024B5254\r\n     0x024B51B8\r\n      0x024B52A3\r\n       0x02300015\r\n\r\n异常调用过程： 0x024B8656\r\n  0x024B8A65\r\n   0x024AB2F5\r\n    0x024B7CA3\r\n     0x024B7B4C\r\n      0x024B7D74\r\n       0x10063260\r\n        0x024A0410\r\n         0x024B5254\r\n          0x024B51B8\r\n           0x024B52A3\r\n            0x02300015\r\n\r\n当前调用过程： 0x024B7B4C\r\n  0x10063260\r\n   0x024A0410\r\n    0x024B5254\r\n     0x024B51B8\r\n      0x024B52A3\r\n       0x02300015\r\n`, "127.0.0.1")
			Ser_Log.Log_写用户消息(2, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.3", "内存写入错误错误信息:11191919;2424233", "127.0.0.1")
		}
		局_例子记录.DbLogusermsg = 局_例子版本
	}
	//-============================================结束==========================

	//检查 代理数量  是否有数据没有
	局_例子版本 = 1
	if 局_例子记录.DbAgentLevel < 局_例子版本 {
		global.GVA_DB.Model(DB.Db_Agent_Level{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_User.New用户信息("刘备", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", -1, 50, 0, "")
			局_Uid := Ser_User.User用户名取id("刘备")
			if 局_Uid > 0 {
				Ser_User.New用户信息("关羽", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0, "")
				Ser_User.New用户信息("张飞", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0, "")
				Ser_User.New用户信息("诸葛亮", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0, "")
			}

			局_Uid = Ser_User.User用户名取id("关羽")
			if 局_Uid > 0 {
				Ser_User.New用户信息("关平", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 10, 0, "")
			}
			局_Uid = Ser_User.User用户名取id("张飞")
			if 局_Uid > 0 {
				Ser_User.New用户信息("张苞", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 10, 0, "")
			}
		}
		局_例子记录.DbAgentLevel = 局_例子版本
	}
	//-============================================结束==========================
	//检查 定时任务,例子
	局_例子版本 = 1
	if 局_例子记录.Cron < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_Cron{}).Count(&局_数量)
		if 局_数量 == 0 {
			var S = service.S_Cron{}
			tx := *global.GVA_DB
			_ = S.Create(&tx, dbm.DB_Cron{Name: "测试网页访问", Status: 2, IsLog: 2, Type: 1, Cron: `0 0 0 * * ?`, RunText: `https://www.baidu.com`, Note: "例子每分钟请求一次"})
			_ = S.Create(&tx, dbm.DB_Cron{Name: "测试公共函数", Status: 2, IsLog: 2, Type: 2, Cron: `0 0 0 * * ?`, RunText: `测试网页访问("aaa")`, Note: "例子每分钟执行一次公共函数"})
			_ = S.Create(&tx, dbm.DB_Cron{Name: "测试执行sql", Status: 2, IsLog: 2, Type: 3, Cron: `0 * * * * ?`, RunText: `DELETE FROM db_cron_log WHERE  RunTime<{{十位时间戳}}-86400`, Note: "例子每天请求一次,支持变量{{十位时间戳}}会替换当前时间戳"})
		}
		局_例子记录.Cron = 局_例子版本
	}
	//-============================================结束==========================
	//检查 卡号列表,执行修改旧卡的卡号使用时间
	局_例子版本 = 1
	if global.GVA_DB.Exec("Select 1 FROM `db_ka`  WHERE  `UserTime` != '' and UseTime=0").RowsAffected > 0 {
		局_sql := "UPDATE `db_ka`  SET `UseTime` = CAST(LEFT(`UserTime`, 10) AS UNSIGNED)  WHERE  `UserTime` != '' and UseTime=0"
		global.GVA_LOG.Info("兼容执行修改旧卡的卡号时间,执行数量:" + strconv.Itoa(int(global.GVA_DB.Exec(局_sql).RowsAffected)))
		局_例子记录.KaUseTime = 局_例子版本
	}

	err := setting.Z例子写出记录(&局_例子记录)
	if err != nil {
		return
	}
}
func 插入公共js例子() {
	局_例子版本 := 1
	if global.GVA_Viper.GetInt("test.DB_PublicJs") >= 局_例子版本 {
		return
	}

	var 局_数量 int64
	global.GVA_DB.Model(DB.DB_PublicJs{}).Count(&局_数量)
	if 局_数量 != 0 {
		return
	}
	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "测试1111",
		Value: `function 测试1111(JSON形参文本) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{
return 局_用户信息
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "测试网页访问",
		Value: `function 测试网页访问(JSON形参文本) {

    局_url = "https://www.baidu.com/sugrec?&prod=pc_his&from=pc_web&json=1&sid=38516_36555_38613_38538_38595_38581_36803_38485_38637_26350_38621_38663&hisdata=%5B%7B%22time%22%3A1675596837%2C%22kw%22%3A%22python%E8%BF%90%E8%A1%8C%E6%97%B6%E4%BF%AE%E6%94%B9%E4%BB%A3%E7%A0%81%22%7D%2C%7B%22time%22%3A1675605796%2C%22kw%22%3A%22%E7%86%8A%E7%8C%AB%208.23%20%E6%BA%90%E7%A0%81%22%2C%22fq%22%3A2%7D%2C%7B%22time%22%3A1675609301%2C%22kw%22%3A%22go%20%E4%BC%98%E7%A7%80%E9%A1%B9%E7%9B%AE%22%7D%2C%7B%22time%22%3A1675671958%2C%22kw%22%3A%22win11%20%E7%AA%97%E5%8F%A3%E6%97%A0%E6%B3%95%E6%8B%96%E6%96%87%E4%BB%B6%22%7D%2C%7B%22time%22%3A1675673946%2C%22kw%22%3A%22win11%20%E6%97%A0%E6%B3%95%E6%8B%96%E6%94%BE%22%2C%22fq%22%3A3%7D%2C%7B%22time%22%3A1676041744%2C%22kw%22%3A%22ns%20retroarch%20%E6%9A%82%E5%81%9C%22%7D%2C%7B%22time%22%3A1676042251%2C%22kw%22%3A%22ns%20retroarch%20%E6%B2%A1%E6%9C%89%E8%AE%BE%E7%BD%AE%22%7D%2C%7B%22time%22%3A1676096636%2C%22kw%22%3A%22gin%20%E5%92%8C%20beego%22%7D%2C%7B%22time%22%3A1676126858%2C%22kw%22%3A%22ruby%E5%92%8Cgo%22%7D%2C%7B%22time%22%3A1676177555%2C%22kw%22%3A%22%E7%A7%9F%E5%8F%B7%E7%8E%A9%E7%BD%91%E6%98%93%E4%B8%8A%E5%8F%B7%E5%99%A8%22%7D%5D&_t=1684664739701&req=2&csor=0"

    //返回对象 = $api_网页访问_GET(局_url, 15, "")
    协议头 = [
        "Accept: application/json, text/javascript, */*; q=0.01",
        "Content-Type: application/json",
        "User-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"
    ]
    返回对象 = $api_网页访问_POST(局_url, "api=123",协议头,"", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}
    return 返回对象.Body //只返回响应信息
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})
	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "任务池创建延迟查询结果例子",
		Value: `function 任务池创建查询例子(形参) {
    let 任务类型ID = 1
    let 结果 = $api_任务池_任务创建($用户在线信息, 任务类型ID, JSON形参文本)
    //{"IsOk":true,"Err":"","Data":{"TaskUuid":"1fb701a9-05c5-442a-8bcc-34bda07050ae"}}

    if (结果.IsOk) {
        let 局_任务对象 = 结果.Data
        let 任务结果
        for (let i = 0; i < 3; i++) {
            $程序_延时(5000); // 等待1秒
            任务结果 = $api_任务池_任务查询(局_任务对象.TaskUuid)
            if (任务结果.Data.Status !== 1 && 任务结果.Data.Status !== 2) { //不是刚创建, 也不是处理中,跳出循环
                break
            }
        }
        //{"IsOk":true,"Err":"","Data":{"ReturnData":"","Status":1,"TimeEnd":0,"TimeStart":1695016978}}
        if (任务结果.Data.Status === 3) {
            // 如果是成功,直接返回
            return {
                Code: 1,
                Msg: "ok",
                recognition: 任务结果.Data.ReturnData,
            }
        }
    }
    return {
        Code: -1,
        Msg: "失败",
    }
}

const js对象_通用返回 = { //api函数返回基本都是这个
    IsOk: false,
    Err: "",
    Data: {}
};`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "用户余额增减案例",
		Value: `function 用户余额增减案例(JSON形参文本) {
	return 0 // 默认函数安全防护 正式使用时请删掉本行
    JSON形参文本 = JSON形参文本.replace(/'/g, '"') //因为易语言 双引号不方便,所以到js里换成替换单引号成双引号 //注意永远不要相信客户端传参

    var 局_形参对象 = JSON.parse(JSON形参文本); //使用JSON.parse() 将JSON字符串转为JS对象;

    if (局_形参对象.a > 0) {
        $拦截原因 = "金额不能大于0"
        return {
            IsOk: false,
            Err: "金额不能大于0"
        }
    } else {
        局_结果 = $api_用户Id增减余额($用户在线信息, 局_形参对象.a, "测试公共函数扣余额")
    }


    return 局_结果 // {IsOk: true, Err: ""}
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "获取用户相关信息",
		Value: `function 获取用户相关信息(形参) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //Id":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
    //var 局_卡号信息 = $api_卡号Id取详情($用户在线信息)
	var 局_软件用户信息 = $api_取软件用户详情($用户在线信息) 

    $api_置动态标记($用户在线信息, $用户在线信息.Tab + "追加文本")

    return 局_用户信息
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "读写公共变量案例",
		Value: `function 读写公共变量案例(JSON形参文本) {

    var 待写入变量 = $api_读公共变量("系统名称")

    var 局_逻辑 = $api_置公共变量("系统名称", 待写入变量 + "追加1")

    return 局_逻辑
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "执行SQL功能测试",
		Value: `function 执行SQL功能测试(JSON形参文本) {
	return 0 // 默认函数安全防护 正式使用时请删掉本行
    var 局_结果对象 = $api_执行SQL功能("UPDATE db_public_js SET Type=Type+1 WHERE  Id=11") //获取公共函数数据库全部信息
    if (局_结果对象.isOk) {
        //这里说明成功了,
        let 影响行数 = Number(局_结果对象.Err)
        return 影响行数 //返回影响行数
    }

    return 局_结果对象.Err

}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "执行SQL查询测试",
		Value: `function 执行SQL查询测试(JSON形参文本) {
	return {} // 默认函数安全防护 正式使用时请删掉本行
    var 局_结果对象 = $api_执行SQL查询(" SELECT * FROM db_public_js")  

    if (局_结果对象.isOk) {
        //这里说明查询成功了,
        return 局_结果对象.Err
    }
    //return 局_结果对象.Err   //这个会把结果返回的文本
    return 局_结果对象.Data //这个会把结果转换成对象

}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})
	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "测试调用管理员后台接口冻结卡号",
		Value: `function 测试调用管理员后台接口冻结卡号(参数) {
    //详细说明 官网常见问题  http://www.fnkuaiyan.cn/%E6%8C%87%E5%8D%97/%E5%B8%B8%E8%A7%81%E9%97%AE%E9%A2%98.html#%E5%85%AC%E5%85%B1%E5%87%BD%E6%95%B0%E5%86%85token%E8%B0%83%E7%94%A8%E5%90%8E%E5%8F%B0%E6%88%96%E4%BB%A3%E7%90%86%E6%8E%A5%E5%8F%A3%E5%8A%9F%E8%83%BD
    局_url = "http://127.0.0.1:18888/Admin/AppUser/SetStatus"
    局_post = '{"AppId":10001,"Id":[69],"Status":2}' //这里可以根据需求自己修改参数, 这个id是卡号id,AppId是卡号归属id
    局_token = "WD3NMTTWNG40DERXA6WRZTK3BZZLTKMJ"  //这个需要自己抓包替换
    协议头 = "Token: " + 局_token
    返回对象 = $api_网页访问_POST(局_url, 局_post,协议头, "", 15, "")

    return 返回对象
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程调用管理员后台接口冻结卡号",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "WebApi_用户Id取详情",
		Value: `function WebApi_用户Id取详情(JSON形参文本) {
    JSON形参文本 = JSON形参文本.replace(/'/g, '"') //因为易语言 双引号不方便,所以到js里换成替换单引号成双引号 //注意永远不要相信客户端传参

    var 局_形参对象 = JSON.parse(JSON形参文本); //使用JSON.parse() 将JSON字符串转为JS对象;

    $用户在线信息.Uid = 局_形参对象.Uid //下边这个传对象,所以先赋值Uid 到对象内
    局_结果 = $api_用户Id取详情($用户在线信息)


    return 局_结果 
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 2,
		Name:  "hook模板_任务创建入库前",
		Value: `function hook模板_任务创建入库前(任务JSON格式参数) {
    
    return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    return $用户在线信息.Uid

  var 局_用户信息 = $api_用户Id取详情($用户在线信息) //
    例子随机 拦截任务提交

    任务JSON格式参数 = 任务JSON格式参数.replace(/'/g, '"') //因为易语言 双引号不方便,所以到js里换成替换单引号成双引号 //注意永远不要相信客户端传参,建议直接在hook函数内固定金额,这里只是测试
   var 局_形参对象 = JSON.parse(任务JSON格式参数); //使用JSON.parse() 将JSON字符串转为JS对象;
    局_结果 = $api_用户Id增减余额($用户在线信息, -局_形参对象.a, "测试任务池Hook内扣余额") //扣款需要时负数需要前面加 - 负号 直接操作就行, 内部会自动判断不用再js先判断余额是否充足,
    if (!局_结果.IsOk) {
        $拦截原因 = "扣费失败" + 局_结果.Err
    }

         if (Math.floor(Math.random() * 10) > 5) {
         $拦截原因 = "如果值不为空,则任务拦截,响应拦截原因"
         }
         //   $拦截原因 只要赋值了,就会被拦截,如果没赋值 就正常放行
    return 任务JSON格式参数 //任务JSON格式文本型参数,可以在这里修改内容  然后返回
}`,
		Type:  1,
		IsVip: 0,
		Note:  "任务池hook例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 3,
		Name:  "Hook_ApiHOOk执行前例子",
		Value: `function Hook_Api执行前HOOk例子` + App.Api之前Hook函数模板,
		Type:  2,
		IsVip: 0,
		Note:  "ApiHook例子,这个是演示hook登录接口进入前,先判断是否能访问百度,如果不能直接拦截响应失败,",
	})
	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 3,
		Name:  "Hook_ApiHOOk执行后例子",
		Value: `function Hook_Api执行后HOOk例子` + App.Api之后Hook函数模板,
		Type:  2,
		IsVip: 0,
		Note:  "ApiHook例子,这个是演示hook登录结束后,修改响应明文的例子,",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "置用户云配置",
		Value: `function 置用户云配置(JSON形参文本) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    let 配置名 = "窗口宽度";
    let 配置值 = "360px";

    // $用户在线信息.LoginAppid=10001    //通过修改登录appid,可以读取其他应用的云配置信息
    // 局uid=$api_用户名或卡号取uid($用户在线信息.LoginAppid,"aaaaaa")  // 可以通过这个获取指定用户的uid, 其实就是应用用户的来源id  无该用户返回0
    $用户在线信息.Uid = 57 //通过修改uid 用户或卡号id,可以读取其他用户的云配置信息


    var 局_结果对象 = $api_置用户云配置($用户在线信息, 配置名, 配置值)
    if (局_结果对象.IsOk) {
        return "写入成功"
    }

    return 局_结果对象.Err //这里存放错误信息,
}`,
		Type:  1,
		IsVip: 0,
		Note:  "置用户云配置例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "取用户云配置",
		Value: `function 取用户云配置(JSON形参文本) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    let 配置名 = "窗口宽度";

    // $用户在线信息.LoginAppid=10001    //通过修改登录appid,可以读取其他应用的云配置信息
    // 局uid=$api_用户名或卡号取uid($用户在线信息.LoginAppid,"aaaaaa")  // 可以通过这个获取指定用户的uid, 其实就是应用用户的来源id  无该用户返回0
    $用户在线信息.Uid = 57 //通过修改uid 用户或卡号id,可以读取其他用户的云配置信息


    var 局_结果对象 = $api_取用户云配置($用户在线信息, 配置名)
    if (局_结果对象.IsOk) {
        return 局_结果对象.Data //只要非异常都会返回值 即使没有这个配置也会返回空字符串
    }

    return 局_结果对象.Err //这里存放错误信息,只有参数不正确,或数据库无法连接的情况才会有错误
}`,
		Type:  1,
		IsVip: 0,
		Note:  "取用户云配置例程",
	})
	global.GVA_Viper.Set("test.DB_PublicJs", 局_例子版本)

}

func 数据库兼容旧版本(c *gin.Context) {

	db := *global.GVA_DB //全局变量赋值到局部
	var 局_待处理订单Id数组 []DB.DB_LogRMBPayOrder
	_ = db.Model(DB.DB_LogRMBPayOrder{}).Where("UidType is NULL ").Find(&局_待处理订单Id数组).Error
	for _, 局_订单 := range 局_待处理订单Id数组 {
		err := db.Model(DB.DB_LogRMBPayOrder{}).Where("Id = ?", 局_订单.Id).Updates(map[string]interface{}{
			"UidType":        1,
			"User":           Ser_User.Id取User(局_订单.Uid),
			"ProcessingType": 0,
			"Extra":          "",
		}).Error
		if err != nil {
			global.GVA_LOG.Info("支付支付订单,兼容旧版本处理失败ID:" + strconv.Itoa(局_订单.Uid))
		}
	}
	//2023/9/9  把支付方式  微信PC修改成 微信支付
	_ = db.Model(DB.DB_LogRMBPayOrder{}).Where("Type = ? ", "微信PC").Update("Type", "微信支付").Error
	//2023/9/16  把appUser 积分 字段类型 修改成  双精度小数型
	局_已有AppID := Ser_AppInfo.App取map列表String()
	for 值 := range 局_已有AppID {
		columnType := ""
		err := db.Raw("SELECT data_type FROM information_schema.columns WHERE table_name = 'db_AppUser_" + 值 + "' AND column_name = 'VipNumber'").Scan(&columnType).Error
		if columnType != "" && columnType != "decimal" {
			err = db.Exec("ALTER TABLE db_AppUser_" + 值 + " MODIFY COLUMN VipNumber DECIMAL(10,2)").Error
			if err != nil {
				fmt.Println("兼容就版本,%d成功修改字段类型为 DECIMAL(10, 2)报错:?", 值, err.Error())
			} else {
				fmt.Println("兼容就版本,%d成功修改字段类型为 DECIMAL(10, 2)", 值)
			}

		}

	}
	//2023/9/17  把任务信息数据库 生成信息和消费信息,字段修改长度为5000

	columnType := ""
	err := db.Raw("SELECT COLUMN_TYPE FROM information_schema.columns WHERE table_name = 'db_TaskPoolData' AND column_name = 'SubmitData'").Scan(&columnType).Error
	if columnType != "" && columnType != "varchar(8000)" {
		err = db.Exec("ALTER TABLE db_TaskPoolData MODIFY COLUMN ReturnData varchar(8000)").Error
		err = db.Exec("ALTER TABLE db_TaskPoolData MODIFY COLUMN SubmitData varchar(8000)").Error
		if err != nil {
			fmt.Println("兼容就版本,失败修改字段类型为 varchar(8000)", err.Error())
		} else {
			fmt.Println("兼容就版本,成功修改字段类型为 varchar(8000)")
		}

	}

	//2023/10/9 发现用户云配置联合主键有问题,有的会缺字段,所以增加一个判断
	var 局_主键信息 int
	//err = db.Raw(`SHOW INDEX FROM db_UserConfig  WHERE Key_name="PRIMARY"`).Scan(&局_主键信息).Error
	err = db.Raw(`SELECT COUNT(*) AS count FROM information_schema.statistics WHERE table_name = 'db_UserConfig' `).Scan(&局_主键信息).Error
	if 局_主键信息 < 4 {
		err = db.Exec("ALTER TABLE db_UserConfig DROP PRIMARY KEY").Error                   //删除旧联合主键
		err = db.Exec("ALTER TABLE db_UserConfig ADD PRIMARY KEY (AppId, Uid, Name)").Error //设置新联合主键
		fmt.Println("已处理并兼容用户云配置联合主键问题")
	}

	//2023/12/13  将配置信息改放到数据库,将旧的数据写入数据库
	var 局_总数 int64
	_ = db.Model(dbm.DB_Setting{}).Count(&局_总数).Error
	if 局_总数 == 0 && global.GVA_Viper.IsSet("系统设置.系统开关") {
		var Test config.Test
		Test.DbAgentLevel = global.GVA_Viper.GetInt("test.db_agent_level")
		Test.DbAppinfo = global.GVA_Viper.GetInt("test.db_appinfo")
		Test.DbLogmoney = global.GVA_Viper.GetInt("test.db_logmoney")
		Test.DbLogrmbpayorder = global.GVA_Viper.GetInt("test.db_logrmbpayorder")
		Test.DbLogusermsg = global.GVA_Viper.GetInt("test.db_logusermsg")
		Test.DbLogvipnumber = global.GVA_Viper.GetInt("test.db_logvipnumber")
		Test.DbPublicdata = global.GVA_Viper.GetInt("test.db_publicdata")
		Test.DbUser = global.GVA_Viper.GetInt("test.db_user")
		Test.Taskpool = global.GVA_Viper.GetInt("test.taskpool_类型")
		Test.User = global.GVA_Viper.GetInt("test.user")
		_ = setting.Z例子写出记录(&Test)

		fmt.Printf("已处理并兼容用户云配置联合主键问题\n")
	}

	//2024/05/12  appUser 缺少归属代理id,无法自动创建所以需要兼容处理
	局_所有应用信息, err := service.NewAppInfo(c, &db).Infos(map[string]interface{}{})
	for _, v := range 局_所有应用信息 {
		var 局_字段 []string
		局_sql := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = 'db_AppUser_%d' AND column_name IN ('AgentUid','Id')", v.AppId)

		err = db.Raw(局_sql).Scan(&局_字段).Error
		if len(局_字段) == 1 { //正常表应该有两个成员 id,AgentUid 只有一个说明缺少字段
			局_sql = fmt.Sprintf("ALTER TABLE `db_AppUser_%d` ADD COLUMN `AgentUid` BIGINT(20) NULL DEFAULT 0 COMMENT '归属代理Uid' AFTER `RegisterTime`", v.AppId)
			err = db.Exec(局_sql).Error
			if err != nil {
				fmt.Println("兼容就版本,软件用户表添加AgentUid", err.Error())
			} else {
				fmt.Println("兼容就版本,软件用户表成功修添加AgentUid")
			}
		}

	}
	//2025/01/05  修改webAppid=3 的 uid 为id 方便业务判断在线逻辑  任务池用的人比较少 基本无影响
	//err = db.Model(DB.DB_LinksToken{}).Where("LoginAppid = ?", 3).Where("Uid = ?", 0).Update("Uid", gorm.Expr("id")).Error
	//fmt.Println(err.Error())

	//2025/03/29  UniqueNumLog 缺少唯一减少积分表,自动创建所以需要兼容处理
	局_所有应用信息, err = service.NewAppInfo(c, &db).Infos(map[string]interface{}{})
	for _, v := range 局_所有应用信息 {
		//检查是否存在表 不存在则创建
		// 检查是否存在表 不存在则创建
		migrator := db.Migrator()
		tableName := dbm.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(v.AppId)
		// 检查表是否存在
		if migrator.HasTable(tableName) {
			continue //如果存在则跳到循环尾
		}
		// 创建唯一积分记录表
		if err = db.Set("gorm:table_options", "ENGINE=InnoDB").
			Table(dbm.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(v.AppId)).
			AutoMigrate(&dbm.DB_UniqueNumLog{}); err != nil {
			fmt.Println("积分记录表创建失败: ", err.Error())
		}
	}

}
