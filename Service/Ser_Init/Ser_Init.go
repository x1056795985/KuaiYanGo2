package Ser_Init

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Init/internal"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_PublicData"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_TaskPool"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
	"server/utils"
	"strconv"
	"time"
)

// InitGormMysql 初始化数据库并产生数据库全局变量
func InitGormMysql() *gorm.DB {
	m := global.GVA_CONFIG.Mysql
	//如果没有数据库名字,直接返回估计还没传配置,第一次启动
	if m.Dbname == "" {
		return nil
	}

	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}

	//连接并设置数据库连接池参数
	if db, err := gorm.Open(mysql.New(mysqlConfig), internal.Gorm.Config(m.Prefix)); err != nil {
		//链接失败了
		return nil
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns) //最大连接数
		sqlDB.SetMaxOpenConns(m.MaxOpenConns) //允许空闲数
		//设定数据库连接的最大生命周期 Mysql默认120秒 所以gorm 设置个比这个值小的数 防止断开连接时操作数据库失败
		sqlDB.SetConnMaxLifetime(100 * time.Second)
		return db //返回连接好的db池
		//获取gorm db对象，其他包需要执行数据库查询的时候，只要通过	global.GVA_DB 获取db对象即可。
		//不用担心协程并发使用同样的db对象会共用同一个连接，db对象在调用他的方法的时候会从数据库连接池中获取新的连接
	}
}

// InitDbTables

func InitDbTables() {
	db := global.GVA_DB //全局变量赋值到局部

	//gorm:table_options 设置创建表强制为InnoDB引擎, 因为MyISAM不支持事务,回滚会失效所以要修改成InnoDB引擎,
	//参考地址 https://blog.csdn.net/qq_25436207/article/details/107533197

	//AutoMigrate 自动迁移功能（如不存在会自动创建一个表）传过来的是  结构体HelloWorld实例的指针地址
	err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		// 系统模块表  数据库结构表
		DB.DB_PublicData{},
		DB.DB_PublicJs{},

		DB.DB_Admin{},
		DB.DB_User{},
		DB.DB_LinksToken{},

		DB.DB_AppInfo{},
		// DB.DB_AppUser{}, //DB.DB_AppUser{},   //因为每个应用一个表 所以不在自动迁移里处理  只在创建应用时 创建 处理
		DB.DB_UserClass{},

		DB.DB_KaClass{},
		DB.DB_Ka{},

		DB.DB_LogMoney{},
		DB.DB_LogLogin{},
		DB.DB_LogUserMsg{},
		DB.DB_LogRMBPayOrder{},
		DB.DB_LogKa{},
		DB.DB_LogRiskControl{},
		DB.DB_LogVipNumber{},
		//任务池数据库
		DB.TaskPool_类型{},
		DB.TaskPool_队列{},
		DB.TaskPool_数据{},
		//	代理相关
		DB.Db_Agent_Level{},
		DB.Db_Agent_卡类授权{},
		DB.Db_Agent_库存日志{},
		DB.Db_Agent_库存卡包{},
	)

	if err != nil {
		global.GVA_LOG.Error("InitDbTables表创建失败", zap.Error(err)) //日志错误 表创建成功
		os.Exit(0)                                                //结束程序
	}
	//global.GVA_LOG.Info("register table success(创建表成功)") //日志消息 表创建成功
	InitDbTable数据() //初始化数据

}

func InitDbTable数据() {
	db := global.GVA_DB //全局变量赋值到局部
	if db == nil {
		return
	}
	//检查 admin表是否有admin账号============================================开始
	var 局_数量 int64
	db.Model(DB.DB_Admin{}).Where("User = ?", "admin").Count(&局_数量)
	if 局_数量 == 0 {
		entities := []DB.DB_Admin{{
			Id:            1,
			User:          "admin",
			PassWord:      utils.BcryptHash("admin"),
			Phone:         "",
			Email:         "",
			Qq:            "",
			SuperPassWord: utils.BcryptHash("admin"),
			Status:        1,
			Authority:     "All",
			AgentDiscount: 100,
		},
		}
		global.GVA_DB.Create(&entities)
	}
	//=============================================结束

	//检查 用户表  是否有数据没有
	global.GVA_DB.Model(DB.DB_User{}).Count(&局_数量)
	if 局_数量 == 0 {
		Ser_User.New用户信息("test0001", "test0001", "test0001test0001", "10001", "10001@qq.com", "", "127.0.0.1", "", 0, 0, 0)
	}
	//-============================================结束==========================
	//检查 DB_AppInfo表是否有应用如果没有插入测试应用============================================
	global.GVA_DB.Model(DB.DB_AppInfo{}).Count(&局_数量)
	if 局_数量 == 0 {
		_ = Ser_AppInfo.NewApp信息(10001, 1, "演示对接账密限时Rsa交换密匙")

		Ser_AppUser.New用户信息(10001, 1, "测试绑定", 1, time.Now().Unix(), 11.02)
		卡类ID, _ := Ser_KaClass.KaClass创建New(10001, "天卡", "Y30", 2592000, 2592000, 0.01, 1.01, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
		卡类ID, _ = Ser_KaClass.KaClass创建New(10001, "月卡", "Y30", 2592000, 2592000, 0.01, 1.01, 100, 100, 0, 1, 25, 1, 1, 1, 1)
		卡信息, _ := Ser_Ka.Ka单卡创建(卡类ID, "admin", "演示创建", "", 0)
		卡信息, _ = Ser_Ka.Ka单卡创建(卡类ID, "admin", "演示创建可追回卡号", "", 0)
		Ser_Ka.K卡号充值_事务(10001, 卡信息.Name, "test0001", "", "127.0.0.1")
		_ = Ser_AppInfo.NewApp信息(10002, 3, "演示对接卡号限时RSA通讯")
		卡类ID, _ = Ser_KaClass.KaClass创建New(10002, "天卡", "Y01", 86400, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
		卡类ID, _ = Ser_KaClass.KaClass创建New(10002, "周卡", "Y01", 604800, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
	}

	//-============================================结束==========================

	//检查 余额充值订单 是否有应用如果没有插入测试应用============================================
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
	//-============================================结束==========================
	//检查 余额日志  是否有数据没有
	global.GVA_DB.Model(DB.DB_LogMoney{}).Count(&局_数量)
	if 局_数量 == 0 {
		Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", -0.01)
		Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", 0.01)
	}
	//-============================================结束==========================
	//检查 积分点数  是否有数据没有
	global.GVA_DB.Model(DB.DB_LogVipNumber{}).Count(&局_数量)
	if 局_数量 == 0 {
		Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示积分效果", -0.01, 10001, 1)
		Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示积分效果", 0.01, 10001, 1)
		Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示点数效果", 1, 10001, 2)
		Ser_Log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示点数效果", -1, 10001, 2)
	}
	//-============================================结束==========================

	//检查 公共变量表  是否有数据没有====================================================
	global.GVA_DB.Model(DB.DB_PublicData{}).Count(&局_数量)
	if 局_数量 == 0 {
		_ = Ser_PublicData.C创建(DB.DB_PublicData{
			AppId: 1,
			Type:  3,
			Name:  "测试逻辑开关",
			Value: "1",
		})
		_ = Ser_PublicData.C创建(DB.DB_PublicData{
			AppId: 1,
			Type:  1,
			Name:  "系统名称",
			Value: "飞鸟快验应用管理后台",
		})
	}
	//-============================================结束==========================

	//检查 公共js表  是否有数据没有
	插入公共js例子() //太长了,单独写个函数
	//-============================================结束==========================
	//检查 任务类型  是否有数据没有
	global.GVA_DB.Model(DB.TaskPool_类型{}).Count(&局_数量)
	if 局_数量 == 0 {
		_ = Ser_TaskPool.Task类型创建("测试任务1", "hook模板_任务创建入库前", "", "", "")
	}
	//-============================================结束==========================

	//检查 任务类型  是否有数据没有
	global.GVA_DB.Model(DB.DB_LogUserMsg{}).Count(&局_数量)
	if 局_数量 == 0 {
		Ser_Log.Log_写用户消息(3, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", "建议做个自动赚钱的功能,启动软件后,微信余额就蹭蹭涨", "127.0.0.1")
		Ser_Log.Log_写用户消息(2, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", `捕获到异常bug文件名:EDV8FCC.tmp句柄数:508,ExceptionText：运行时出错!\r\n\r\n错误代码：0\r\n\r\n错误信息：分配 1073741832 字节内存失败!\r\n0, 0\r\n\r\nCallStack:\r\n 0x024B7B4C\r\n  0x10063260\r\n   0x024A0410\r\n    0x024B5254\r\n     0x024B51B8\r\n      0x024B52A3\r\n       0x02300015\r\n\r\n异常调用过程： 0x024B8656\r\n  0x024B8A65\r\n   0x024AB2F5\r\n    0x024B7CA3\r\n     0x024B7B4C\r\n      0x024B7D74\r\n       0x10063260\r\n        0x024A0410\r\n         0x024B5254\r\n          0x024B51B8\r\n           0x024B52A3\r\n            0x02300015\r\n\r\n当前调用过程： 0x024B7B4C\r\n  0x10063260\r\n   0x024A0410\r\n    0x024B5254\r\n     0x024B51B8\r\n      0x024B52A3\r\n       0x02300015\r\n`, "127.0.0.1")
		Ser_Log.Log_写用户消息(2, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.3", "内存写入错误错误信息:11191919;2424233", "127.0.0.1")
	}
	//-============================================结束==========================

	//检查 代理数量  是否有数据没有
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Count(&局_数量)
	if 局_数量 == 0 {
		Ser_User.New用户信息("刘备", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", -1, 50, 0)
		局_Uid := Ser_User.User用户名取id("刘备")
		if 局_Uid > 0 {
			Ser_User.New用户信息("关羽", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0)
			Ser_User.New用户信息("张飞", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0)
			Ser_User.New用户信息("诸葛亮", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0)
		}
		局_Uid = Ser_User.User用户名取id("关羽")
		if 局_Uid > 0 {
			Ser_User.New用户信息("关平", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 10, 0)
		}
		局_Uid = Ser_User.User用户名取id("张飞")
		if 局_Uid > 0 {
			Ser_User.New用户信息("张苞", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 10, 0)
		}
	}
	//-============================================结束==========================
	数据库兼容旧版本()
}
func 插入公共js例子() {

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

    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{"Id":21,"User":"aaaaaa","PassWord":"af15d5fdacd5fdfea300e88a8e253e82","Phone":"13109812593","Email":"1056795985@qq.com","Qq":"1059795985","SuperPassWord":"af15d5fdacd5fdfea300e88a8e253e82","Status":1,"Rmb":91.39,"Note":"","RealNameAttestation":"","Role":0,"UPAgentId":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
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

    返回对象 = $api_网页访问_GET(局_url, 15, "")
    //返回对象 = $api_网页访问_POST(局_url, "api=123", 15, "")
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

    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{"Id":21,"User":"aaaaaa","PassWord":"af15d5fdacd5fdfea300e88a8e253e82","Phone":"13109812593","Email":"1056795985@qq.com","Qq":"1059795985","SuperPassWord":"af15d5fdacd5fdfea300e88a8e253e82","Status":1,"Rmb":91.39,"Note":"","RealNameAttestation":"","Role":0,"UPAgentId":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
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

    var 局_结果对象 = $api_执行SQL查询(" SELECT * FROM 'db_public_js'")
	//这里'db_public_js' 两侧可能是单引号,或键盘TAb上上方的按键
    if (局_结果对象.isOk) {
        //这里说明查询成功了,
        return 局_结果对象.Err
    }

    return 局_结果对象.Err

}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
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
    /*
    return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    return $用户在线信息.Uid

  var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{"Id":21,"User":"aaaaaa","PassWord":"af15d5fdacd5fdfea300e88a8e253e82","Phone":"13109812593","Email":"1056795985@qq.com","Qq":"1059795985","SuperPassWord":"af15d5fdacd5fdfea300e88a8e253e82","Status":1,"Rmb":91.39,"Note":"","RealNameAttestation":"","Role":0,"UPAgentId":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
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
         */
    return 任务JSON格式参数 //任务JSON格式文本型参数,可以在这里修改内容  然后返回
}`,
		Type:  1,
		IsVip: 0,
		Note:  "任务池hook例程",
	})
}

func 数据库兼容旧版本() {
	db := global.GVA_DB //全局变量赋值到局部
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
			fmt.Println("兼容就版本,成功修改字段类型为 varchar(8000)", err.Error())
		} else {
			fmt.Println("兼容就版本,成功修改字段类型为 varchar(8000)")
		}

	}

}
