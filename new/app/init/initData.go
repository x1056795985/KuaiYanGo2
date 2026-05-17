package init

import (
	"EFunc/utils"
	"fmt"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_PublicJs"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_TaskPool"
	"server/Service/Ser_User"
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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitDbTables 初始化数据库表
func InitDbTables(c *gin.Context) {
	db := global.GVA_DB

	tables := []interface{}{
		// 系统模块表
		DB.DB_PublicData{},
		DB.DB_PublicJs{},
		DB.DB_UserConfig{},

		DB.DB_Admin{},
		DB.DB_User{},
		DB.DB_LinksToken{},

		DB.DB_AppInfo{},
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
		dbm.DB_LogKey{},

		// 代理相关
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
		dbm.DB_ShortUrl{},
		dbm.DB_CpsInvitingRelation{},
		dbm.DB_CpsUser{},
		dbm.DB_CpsCode{},
		dbm.DB_CpsPayOrder{},

		dbm.DB_CheckInInfo{},
		dbm.DB_CheckInUser{},
		dbm.DB_CheckInScoreLog{},
		dbm.DB_CheckInLog{},
		dbm.DB_CheckInTaskLog{},

		// 统计数据用的表
		dbm.DB_LogUserActive{},
		dbm.DB_TongJiZaiXian{},

		// 任务池数据库
		DB.TaskPool_类型{},
		DB.TaskPool_队列{},
		DB.DB_TaskPoolData{},
	}

	for _, table := range tables {
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(table); err != nil {
			global.GVA_LOG.Error("表创建失败", zap.String("table", fmt.Sprintf("%T", table)), zap.Error(err))
		}
	}

	InitDbTableData(c)
}

// InitDbTableData 初始化示例数据
func InitDbTableData(c *gin.Context) {
	db := global.GVA_DB
	局_例子记录 := setting.Q例子写出记录()

	if db == nil {
		return
	}

	// 检查 admin表是否有账号
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
		}}
		global.GVA_DB.Create(&entities)
	}

	// 检查 用户表
	局_例子版本 := 1
	if 局_例子记录.DbUser < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_User{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_User.New用户信息("test0001", "test0001", "test0001test0001", "10001", "10001@qq.com", "", "127.0.0.1", "", 0, 0, 0, "")
		}
		局_例子记录.DbUser = 局_例子版本
	}

	// 检查 DB_AppInfo表
	局_例子版本 = 1
	if 局_例子记录.DbAppinfo < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_AppInfo{}).Count(&局_数量)
		if 局_数量 == 0 {
			_ = appInfo.L_appInfo.NewApp信息(c, 10001, 1, "演示对接账密限时Rsa交换密匙")
			Ser_AppUser.New用户信息(10001, 1, "测试绑定", 1, time.Now().Unix(), 11.02, 0, "")
			卡类ID, _ := Ser_KaClass.KaClass创建New(10001, "天卡", "Y30", 2592000, 2592000, 0.01, 1.01, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
			卡类ID, _ = Ser_KaClass.KaClass创建New(10001, "月卡", "Y30", 2592000, 2592000, 0.01, 1.01, 100, 100, 0, 1, 25, 1, 1, 1, 1)
			卡信息, _ := Ser_Ka.Ka单卡创建(卡类ID, -1, Ser_Admin.Id取User(1), "演示创建", "", 0)
			卡信息, _ = Ser_Ka.Ka单卡创建(卡类ID, -1, Ser_Admin.Id取User(1), "演示创建可追回卡号", "", 0)
			ka.L_ka.K卡号充值_事务(c, 10001, 卡信息.Name, "test0001", "")
			_ = appInfo.L_appInfo.NewApp信息(c, 10002, 3, "演示对接卡号限时RSA通讯")
			卡类ID, _ = Ser_KaClass.KaClass创建New(10002, "天卡", "Y01", 86400, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
			卡类ID, _ = Ser_KaClass.KaClass创建New(10002, "周卡", "Y01", 604800, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
		}
		局_例子记录.DbAppinfo = 局_例子版本
	}

	// 检查 余额充值订单
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

	// 检查 余额日志
	局_例子版本 = 1
	if 局_例子记录.DbLogmoney < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_LogMoney{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", -0.01)
			Ser_Log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", 0.01)
		}
		局_例子记录.DbLogmoney = 局_例子版本
	}

	// 检查 积分点数
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

	// 检查 公共变量表
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

	// 检查 公共js表
	插入公共Js例子()
	局_例子版本 = 1

	// 检查 任务类型
	if 局_例子记录.Taskpool < 局_例子版本 {
		global.GVA_DB.Model(DB.TaskPool_类型{}).Count(&局_数量)
		if 局_数量 == 0 {
			_ = Ser_TaskPool.Task类型创建("测试任务1", "hook模板_任务创建入库前", "", "", "")
		}
		局_例子记录.Taskpool = 局_例子版本
	}

	// 检查 用户消息
	局_例子版本 = 1
	if 局_例子记录.DbLogusermsg < 局_例子版本 {
		global.GVA_DB.Model(DB.DB_LogUserMsg{}).Count(&局_数量)
		if 局_数量 == 0 {
			Ser_Log.Log_写用户消息(3, 10001, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", "建议做个自动赚钱的功能,启动软件后,微信余额就蹭蹭涨", "127.0.0.1")
			Ser_Log.Log_写用户消息(2, 10001, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", "捕获到异常bug...", "127.0.0.1")
			Ser_Log.Log_写用户消息(2, 10001, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.3", "内存写入错误", "127.0.0.1")
		}
		局_例子记录.DbLogusermsg = 局_例子版本
	}

	// 检查 代理数量
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

	// 检查 定时任务
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

	// 检查 卡号列表,执行修改旧卡的卡号使用时间
	局_例子版本 = 1
	if global.GVA_DB.Exec("Select 1 FROM `db_Ka`  WHERE  `UserTime` != '' and UseTime=0").RowsAffected > 0 {
		局_sql := "UPDATE `db_Ka`  SET `UseTime` = CAST(LEFT(`UserTime`, 10) AS UNSIGNED)  WHERE  `UserTime` != '' and UseTime=0"
		global.GVA_LOG.Info("兼容执行修改旧卡的卡号时间,执行数量:" + strconv.Itoa(int(global.GVA_DB.Exec(局_sql).RowsAffected)))
		局_例子记录.KaUseTime = 局_例子版本
	}

	err := setting.Z例子写出记录(&局_例子记录)
	if err != nil {
		return
	}
}

// 插入公共Js例子 插入公共JS示例数据
func 插入公共Js例子() {
	局_例子版本 := 1
	if global.GVA_Viper.GetInt("test.DB_PublicJs") >= 局_例子版本 {
		return
	}

	var 局_数量 int64
	global.GVA_DB.Model(DB.DB_PublicJs{}).Count(&局_数量)
	if 局_数量 != 0 {
		return
	}

	// 测试公共JS函数
	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "测试1111",
		Value: `function 测试1111(JSON形参文本) {
    var 局_用户信息 = $api_用户Id取详情($用户在线信息)
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
    局_url = "https://www.baidu.com/sugrec?&prod=pc_his&from=pc_web"
    返回对象 = $api_网页访问_POST(局_url, "api=123",协议头,"", 15, "")
    return 返回对象.Body
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
    if (结果.IsOk) {
        let 局_任务对象 = 结果.Data
        let 任务结果
        for (let i = 0; i < 3; i++) {
            $程序_延时(5000);
            任务结果 = $api_任务池_任务查询(局_任务对象.TaskUuid)
            if (任务结果.Data.Status !== 1 && 任务结果.Data.Status !== 2) {
                break
            }
        }
        if (任务结果.Data.Status === 3) {
            return { Code: 1, Msg: "ok", recognition: 任务结果.Data.ReturnData }
        }
    }
    return { Code: -1, Msg: "失败" }
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "用户余额增减案例",
		Value: `function 用户余额增减案例(JSON形参文本) {
	return 0
    JSON形参文本 = JSON形参文本.replace(/'/g, '"')
    var 局_形参对象 = JSON.parse(JSON形参文本);
    if (局_形参对象.a > 0) {
        $拦截原因 = "金额不能大于0"
        return { IsOk: false, Err: "金额不能大于0" }
    } else {
        局_结果 = $api_用户Id增减余额($用户在线信息, 局_形参对象.a, "测试公共函数扣余额")
    }
    return 局_结果
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "获取用户相关信息",
		Value: `function 获取用户相关信息(形参) {
    var 局_用户信息 = $api_用户Id取详情($用户在线信息)
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
	return 0
    var 局_结果对象 = $api_执行SQL功能("UPDATE db_public_js SET Type=Type+1 WHERE  Id=11")
    if (局_结果对象.isOk) {
        let 影响行数 = Number(局_结果对象.Err)
        return 影响行数
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
	return {}
    var 局_结果对象 = $api_执行SQL查询(" SELECT * FROM db_public_js")
    if (局_结果对象.isOk) {
        return 局_结果对象.Data
    }
    return 局_结果对象.Data
}`,
		Type:  1,
		IsVip: 0,
		Note:  "例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "测试调用管理员后台接口冻结卡号",
		Value: `function 测试调用管理员后台接口冻结卡号(参数) {
    局_url = "http://127.0.0.1:18888/Admin/AppUser/SetStatus"
    局_post = '{"AppId":10001,"Id":[69],"Status":2}'
    局_token = "WD3NMTTWNG40DERXA6WRZTK3BZZLTKMJ"
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
    JSON形参文本 = JSON形参文本.replace(/'/g, '"')
    var 局_形参对象 = JSON.parse(JSON形参文本);
    $用户在线信息.Uid = 局_形参对象.Uid
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
    return $用户在线信息
}`,
		Type:  1,
		IsVip: 0,
		Note:  "任务池hook例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "置用户云配置",
		Value: `function 置用户云配置(JSON形参文本) {
    let 配置名 = "窗口宽度";
    let 配置值 = "360px";
    $用户在线信息.Uid = 57
    var 局_结果对象 = $api_置用户云配置($用户在线信息, 配置名, 配置值)
    if (局_结果对象.IsOk) {
        return "写入成功"
    }
    return 局_结果对象.Err
}`,
		Type:  1,
		IsVip: 0,
		Note:  "置用户云配置例程",
	})

	Ser_PublicJs.C创建(DB.DB_PublicJs{
		AppId: 1,
		Name:  "取用户云配置",
		Value: `function 取用户云配置(JSON形参文本) {
    let 配置名 = "窗口宽度";
    $用户在线信息.Uid = 57
    var 局_结果对象 = $api_取用户云配置($用户在线信息, 配置名)
    if (局_结果对象.IsOk) {
        return 局_结果对象.Data
    }
    return 局_结果对象.Err
}`,
		Type:  1,
		IsVip: 0,
		Note:  "取用户云配置例程",
	})

	global.GVA_Viper.Set("test.DB_PublicJs", 局_例子版本)
}

// 数据库兼容旧版本 数据库兼容旧版本升级
func 数据库兼容旧版本(c *gin.Context) {
	db := *global.GVA_DB
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

	// 把支付方式 微信PC修改成 微信支付
	_ = db.Model(DB.DB_LogRMBPayOrder{}).Where("Type = ? ", "微信PC").Update("Type", "微信支付").Error

	// 把appUser 积分 字段类型 修改成 双精度小数型
	局_已有AppID := Ser_AppInfo.App取map列表String(true)
	for 值 := range 局_已有AppID {
		columnType := ""
		err := db.Raw("SELECT data_type FROM information_schema.columns WHERE table_name = 'db_AppUser_" + 值 + "' AND column_name = 'VipNumber'").Scan(&columnType).Error
		if columnType != "" && columnType != "decimal" {
			err = db.Exec("ALTER TABLE db_AppUser_" + 值 + " MODIFY COLUMN VipNumber DECIMAL(10,2)").Error
			if err != nil {
				fmt.Println("兼容就版本,失败修改字段类型为 DECIMAL(10, 2)", err.Error())
			}
		}
	}

	// 把任务信息数据库 生成信息和消费信息,字段修改长度为5000
	columnType := ""
	err := db.Raw("SELECT COLUMN_TYPE FROM information_schema.columns WHERE table_name = 'db_TaskPoolData' AND column_name = 'SubmitData'").Scan(&columnType).Error
	if columnType != "" && columnType != "varchar(8000)" {
		err = db.Exec("ALTER TABLE db_TaskPoolData MODIFY COLUMN ReturnData varchar(8000)").Error
		err = db.Exec("ALTER TABLE db_TaskPoolData MODIFY COLUMN SubmitData varchar(8000)").Error
		if err != nil {
			fmt.Println("兼容就版本,失败修改字段类型为 varchar(8000)", err.Error())
		}
	}

	// 将配置信息改放到数据库,将旧的数据写入数据库
	var 局_总数 int64
	_ = db.Model(dbm.DB_Setting{}).Count(&局_总数).Error
	if 局_总数 == 0 && global.GVA_Viper.IsSet("系统设置.系统开关") {
		var Test = config.Test{
			DbAgentLevel:      global.GVA_Viper.GetInt("test.db_agent_level"),
			DbAppinfo:         global.GVA_Viper.GetInt("test.db_appinfo"),
			DbLogmoney:        global.GVA_Viper.GetInt("test.db_logmoney"),
			DbLogrmbpayorder:  global.GVA_Viper.GetInt("test.db_logrmbpayorder"),
			DbLogusermsg:      global.GVA_Viper.GetInt("test.db_logusermsg"),
			DbLogvipnumber:    global.GVA_Viper.GetInt("test.db_logvipnumber"),
			DbPublicdata:      global.GVA_Viper.GetInt("test.db_publicdata"),
			DbUser:            global.GVA_Viper.GetInt("test.db_user"),
			Taskpool:          global.GVA_Viper.GetInt("test.taskpool_类型"),
			User:              global.GVA_Viper.GetInt("test.user"),
		}
		_ = setting.Z例子写出记录(&Test)
	}

	// appUser 缺少归属代理id
	局_所有应用信息, err := service.NewAppInfo(c, &db).Infos(map[string]interface{}{})
	for _, v := range 局_所有应用信息 {
		var 局_字段 []string
		局_sql := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = 'db_AppUser_%d' AND column_name IN ('AgentUid','Id')", v.AppId)
		err = db.Raw(局_sql).Scan(&局_字段).Error
		if len(局_字段) == 1 {
			局_sql = fmt.Sprintf("ALTER TABLE `db_AppUser_%d` ADD COLUMN `AgentUid` BIGINT(20) NULL DEFAULT 0 COMMENT '归属代理Uid' AFTER `RegisterTime`", v.AppId)
			err = db.Exec(局_sql).Error
			if err != nil {
				fmt.Println("兼容就版本,软件用户表添加AgentUid", err.Error())
			}
		}
	}

	// 唯一积分表
	局_所有应用信息, err = service.NewAppInfo(c, &db).Infos(map[string]interface{}{})
	for _, v := range 局_所有应用信息 {
		migrator := db.Migrator()
		tableName := dbm.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(v.AppId)
		if migrator.HasTable(tableName) {
			continue
		}
		if err = db.Set("gorm:table_options", "ENGINE=InnoDB").
			Table(dbm.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(v.AppId)).
			AutoMigrate(&dbm.DB_UniqueNumLog{}); err != nil {
			fmt.Println("积分记录表创建失败: ", err.Error())
		}
	}

	// 用户消息新增 AppID字段
	_ = db.Model(DB.DB_LogUserMsg{}).Where("AppId = ?", 0).Count(&局_总数).Error
	if 局_总数 > 0 {
		db.Exec("UPDATE db_log_usermsg  AS a SET  AppId=(SELECT AppId FROM db_app_info WHERE AppName =a.App)")
		err = db.Model(DB.DB_LogUserMsg{}).Where("AppId = ?", 0).Update("AppId", gorm.Expr("App")).Error
		err = db.Model(DB.DB_LogUserMsg{}).Where("AppId IS NULL").Delete(&DB.DB_LogUserMsg{}).Error
	}
}
