package bootstrap

import (
	"EFunc/utils"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/app/global"
	"server/app/logic/common/appInfo"
	"server/app/logic/common/appUser"
	"server/app/logic/common/ka"
	"server/app/logic/common/log"
	"server/app/logic/common/publicData"
	"server/app/logic/common/publicJs"
	"server/app/logic/common/rmbPay"
	"server/app/logic/common/setting"
	"server/app/logic/common/user"
	"server/app/models/constant"
	"server/app/models/dbm"

	"server/app/service"
	utils2 "server/app/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// InitDbTables 初始化数据库表
func InitDbTables(c *gin.Context) {
	db := *global.GVA_DB

	tables := []interface{}{
		// 系统模块表
		dbm.DB_PublicData{},
		dbm.DB_PublicJs{},
		dbm.DB_PublicJsCategory{},
		dbm.DB_UserConfig{},

		dbm.DB_Admin{},
		dbm.DB_User{},
		dbm.DB_LinksToken{},

		dbm.DB_AppInfo{},
		dbm.DB_UserClass{},

		dbm.DB_Gateway{},

		dbm.DB_KaClass{},
		dbm.DB_Ka{},

		dbm.DB_LogMoney{},
		dbm.DB_LogLogin{},
		dbm.DB_RmbWithdraw{},
		dbm.DB_RmbWithdrawLog{},

		dbm.DB_LogUserMsg{},
		dbm.DB_LogRMBPayOrder{},
		dbm.DB_WebUserCoupon{},
		dbm.DB_WebUserCouponUser{},
		dbm.DB_WebUserCouponLog{},
		dbm.DB_LogKa{},
		dbm.DB_LogRiskControl{},
		dbm.DB_LogVipNumber{},
		dbm.DB_LogAgentOtherFunc{},
		dbm.DB_LogKey{},

		// 代理相关
		dbm.Db_Agent_Level{},
		dbm.Db_Agent_卡类授权{},
		dbm.Db_Agent_库存日志{},
		dbm.Db_Agent_库存卡包{},

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

		dbm.DB_LuckyWheelInfo{},
		dbm.DB_LuckyWheelUser{},
		dbm.DB_LuckyWheelLog{},

		// 统计数据用的表
		dbm.DB_LogUserActive{},
		dbm.DB_TongJiZaiXian{},

		// 任务池数据库
		dbm.TaskPool_类型{},
		dbm.TaskPool_队列{},
		dbm.DB_TaskPoolData{},
	}

	for _, table := range tables {
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(table); err != nil {
			global.GVA_LOG.Println("表创建失败", "table", fmt.Sprintf("%T", table), err)
		}
	}

	InitDbTableData(c)
}

// InitDbTableData 初始化示例数据
func InitDbTableData(c *gin.Context) {
	if global.GVA_DB == nil {
		return
	}
	db := *global.GVA_DB
	局_例子记录 := setting.Q例子写出记录()
	// 检查 admin表是否有账号
	var 局_数量 int64
	db.Model(dbm.DB_Admin{}).Count(&局_数量)
	if 局_数量 == 0 {
		entities := []dbm.DB_Admin{{
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
		db.Create(&entities)
	}

	// 检查 用户表
	局_例子版本 := 1
	if 局_例子记录.DbUser < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_User{}).Count(&局_数量)
		if 局_数量 == 0 {
			user.L_user.New用户信息(c, "test0001", "test0001", "test0001test0001", "10001", "10001@qq.com", "", "127.0.0.1", "", 0, 0, 0, "")
		}
		局_例子记录.DbUser = 局_例子版本
	}

	// 检查 DB_AppInfo表
	局_例子版本 = 1
	if 局_例子记录.DbAppinfo < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_AppInfo{}).Count(&局_数量)
		if 局_数量 == 0 {
			_ = appInfo.L_appInfo.NewApp信息(c, 10001, 1, "演示对接账密限时Rsa交换密匙")
			appUser.L_appUser.New用户信息(c, 10001, 1, "测试绑定", 1, time.Now().Unix(), 11.02, 0, "")
			卡类ID, _ := ka.L_ka.KaClass创建New(c, 10001, "天卡", "Y30", 2592000, 2592000, 0.01, 1.01, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
			卡类ID, _ = ka.L_ka.KaClass创建New(c, 10001, "月卡", "Y30", 2592000, 2592000, 0.01, 1.01, 100, 100, 0, 1, 25, 1, 1, 1, 1)
			卡信息, _ := ka.L_ka.Ka单卡创建(c, 卡类ID, -1, service.NewAdmin(c, &db).Id取User(1), "演示创建", "", 0)
			卡信息, _ = ka.L_ka.Ka单卡创建(c, 卡类ID, -1, service.NewAdmin(c, &db).Id取User(1), "演示创建可追回卡号", "", 0)
			ka.L_ka.K卡号充值_事务(c, 10001, 卡信息.Name, "test0001", "")
			_ = appInfo.L_appInfo.NewApp信息(c, 10002, 3, "演示对接卡号限时RSA通讯")
			卡类ID, _ = ka.L_ka.KaClass创建New(c, 10002, "天卡", "Y01", 86400, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
			卡类ID, _ = ka.L_ka.KaClass创建New(c, 10002, "周卡", "Y01", 604800, 0, 0, 0, 0.02, 0.02, 0, 1, 25, 1, 1, 1, 1)
		}
		局_例子记录.DbAppinfo = 局_例子版本
	}

	// 检查 余额充值订单
	局_例子版本 = 1
	if 局_例子记录.DbLogrmbpayorder < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_LogRMBPayOrder{}).Count(&局_数量)
		if 局_数量 == 0 {
			订单创建, _ := rmbPay.L_rmbPay.Order订单创建(c, 1, 1, 10001, 0.01, "支付宝PC", "演示数据", "127.0.0.1", 0, "")
			service.NewRmbPayService(&db).Order更新订单状态(订单创建.PayOrder, constant.D订单状态_成功)

			订单创建, _ = rmbPay.L_rmbPay.Order订单创建(c, 1, 1, 10001, 0.01, "微信支付", "演示数据", "127.0.0.1", 0, "")
			service.NewRmbPayService(&db).Order更新订单状态(订单创建.PayOrder, constant.D订单状态_成功)

			订单创建, _ = rmbPay.L_rmbPay.Order订单创建(c, 1, 1, 10001, 0.01, "管理员手动充值", "演示数据", "127.0.0.1", 0, "")
			service.NewRmbPayService(&db).Order更新订单状态(订单创建.PayOrder, constant.D订单状态_成功)
			订单创建, _ = rmbPay.L_rmbPay.Order订单创建(c, 1, 1, 10001, 0.01, "微信支付", "演示数据", "127.0.0.1", 0, "")
			service.NewRmbPayService(&db).Order更新订单状态(订单创建.PayOrder, constant.D订单状态_等待支付)
			订单创建, _ = rmbPay.L_rmbPay.Order订单创建(c, 1, 1, 10001, 0.01, "支付宝PC", "演示数据", "127.0.0.1", 0, "")
			service.NewRmbPayService(&db).Order更新订单状态(订单创建.PayOrder, constant.D订单状态_退款成功)
			go log.L_log.Log_写余额日志("test0001", "127.0.0.1", "管理员操作退款,余额充值订单:"+订单创建.PayOrder+",扣除用户已充值余额"+"|新余额≈"+utils.Float64到文本(0.01, 2), utils.Float64取负值(订单创建.Rmb))

			log.L_log.Log_写余额日志("test0001", "127.0.0.1", "看你长得帅,收费", -0.05)

			订单创建, _ = rmbPay.L_rmbPay.Order订单创建(c, 1, 1, 10001, 0.01, "微信支付", "演示数据", "127.0.0.1", 0, "")
			service.NewRmbPayService(&db).Order更新订单状态(订单创建.PayOrder, constant.D订单状态_退款失败)
		}
		局_例子记录.DbLogrmbpayorder = 局_例子版本
	}

	// 检查 余额日志
	局_例子版本 = 1
	if 局_例子记录.DbLogmoney < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_LogMoney{}).Count(&局_数量)
		if 局_数量 == 0 {
			log.L_log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", -0.01)
			log.L_log.Log_写余额日志("test0001", "127.0.0.1", "演示积分效果", 0.01)
		}
		局_例子记录.DbLogmoney = 局_例子版本
	}

	// 检查 积分点数
	局_例子版本 = 1
	if 局_例子记录.DbLogvipnumber < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_LogVipNumber{}).Count(&局_数量)
		if 局_数量 == 0 {
			log.L_log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示积分效果", -0.01, 10001, 1)
			log.L_log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示积分效果", 0.01, 10001, 1)
			log.L_log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示点数效果", 1, 10001, 2)
			log.L_log.Log_写积分点数时间日志("test0001", "127.0.0.1", "演示点数效果", -1, 10001, 2)
		}
		局_例子记录.DbLogvipnumber = 局_例子版本
	}

	// 检查 公共变量表
	局_例子版本 = 1
	if 局_例子记录.DbPublicdata < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_PublicData{}).Count(&局_数量)
		if 局_数量 == 0 {
			_ = publicData.L_publicData.C创建(&gin.Context{}, dbm.DB_PublicData{
				AppId: 1,
				Type:  3,
				Name:  "测试逻辑开关",
				Value: "1",
			})
			_ = publicData.L_publicData.C创建(&gin.Context{}, dbm.DB_PublicData{
				AppId: 1,
				Type:  1,
				Name:  "系统名称",
				Value: "飞鸟快验应用管理后台",
			})
		}
		局_例子记录.DbPublicdata = 局_例子版本
	}

	// 检查 公共js表
	插入公共Js例子(c)
	局_例子版本 = 1

	// 检查 任务类型
	if 局_例子记录.Taskpool < 局_例子版本 {
		global.GVA_DB.Model(dbm.TaskPool_类型{}).Count(&局_数量)
		if 局_数量 == 0 {
			_, _ = service.NewTaskPoolType(c, &db).Create(dbm.TaskPool_类型{Name: "测试任务1", HookSubmitDataStart: "hook模板_任务创建入库前"})
		}
		局_例子记录.Taskpool = 局_例子版本
	}

	// 检查 用户消息
	局_例子版本 = 1
	if 局_例子记录.DbLogusermsg < 局_例子版本 {
		global.GVA_DB.Model(dbm.DB_LogUserMsg{}).Count(&局_数量)
		if 局_数量 == 0 {
			log.L_log.Log_写用户消息(3, 10001, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", "建议做个自动赚钱的功能,启动软件后,微信余额就蹭蹭涨", "127.0.0.1")
			log.L_log.Log_写用户消息(2, 10001, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.0", "捕获到异常bug...", "127.0.0.1")
			log.L_log.Log_写用户消息(2, 10001, "test0001", "演示对接账密限时Rsa交换密匙", "1.0.3", "内存写入错误", "127.0.0.1")
		}
		局_例子记录.DbLogusermsg = 局_例子版本
	}

	// 检查 代理数量
	局_例子版本 = 1
	if 局_例子记录.DbAgentLevel < 局_例子版本 {
		global.GVA_DB.Model(dbm.Db_Agent_Level{}).Count(&局_数量)
		if 局_数量 == 0 {
			user.L_user.New用户信息(c, "刘备", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", -1, 50, 0, "")
			局_Uid := service.NewUser(c, &db).User用户名取id("刘备")
			if 局_Uid > 0 {
				user.L_user.New用户信息(c, "关羽", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0, "")
				user.L_user.New用户信息(c, "张飞", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0, "")
				user.L_user.New用户信息(c, "诸葛亮", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 30, 0, "")
			}

			局_Uid = service.NewUser(c, &db).User用户名取id("关羽")
			if 局_Uid > 0 {
				user.L_user.New用户信息(c, "关平", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 10, 0, "")
			}
			局_Uid = service.NewUser(c, &db).User用户名取id("张飞")
			if 局_Uid > 0 {
				user.L_user.New用户信息(c, "张苞", "a"+strconv.FormatInt(time.Now().Unix(), 10), "a"+strconv.FormatInt(time.Now().Unix(), 10), "", "", "", "127.0.0.1", "代理数量=0,系统创建演示", 局_Uid, 10, 0, "")
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
		global.GVA_LOG.Println("兼容执行修改旧卡的卡号时间,执行数量:" + strconv.Itoa(int(global.GVA_DB.Exec(局_sql).RowsAffected)))
		局_例子记录.KaUseTime = 局_例子版本
	}
	数据库兼容旧版本(c)
	err := setting.Z例子写出记录(&局_例子记录)
	if err != nil {
		return
	}
}

// 插入公共Js例子 插入公共JS示例数据
func 插入公共Js例子(c *gin.Context) {
	局_例子版本 := 1
	if global.GVA_Viper.GetInt("test.DB_PublicJs") >= 局_例子版本 {
		return
	}

	var 局_数量 int64
	global.GVA_DB.Model(dbm.DB_PublicJs{}).Count(&局_数量)
	if 局_数量 != 0 {
		return
	}

	// 测试公共JS函数
	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
		AppId: 2,
		Name:  "hook模板_任务创建入库前",
		Value: `function hook模板_任务创建入库前(任务JSON格式参数) {
    return $用户在线信息
}`,
		Type:  1,
		IsVip: 0,
		Note:  "任务池hook例程",
	})

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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

	publicJs.L_publicJs.C创建(c, dbm.DB_PublicJs{
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
	db := global.Get局db()
	// 兼容 余额充值订单,旧订单来源AppId存于额外信息,统一回填到AppId字段
	var 局_AppId为0订单 []dbm.DB_LogRMBPayOrder
	if err := db.Model(dbm.DB_LogRMBPayOrder{}).Where("AppId = ?", 0).Find(&局_AppId为0订单).Error; err == nil && len(局_AppId为0订单) > 0 {
		for _, 局_订单 := range 局_AppId为0订单 {
			局_AppId := gjson.New(局_订单.Extra).Get("AppId").Int()
			if 局_AppId <= 0 {
				continue //额外信息里也没有来源AppId,无法回填
			}
			if err = db.Model(dbm.DB_LogRMBPayOrder{}).Where("Id = ?", 局_订单.Id).Update("AppId", 局_AppId).Error; err != nil {
				global.GVA_LOG.Println("兼容支付订单回填来源AppId失败,订单ID:" + strconv.Itoa(局_订单.Id) + "," + err.Error())
				continue
			}
			//global.GVA_LOG.Println("兼容支付订单回填来源AppId,订单ID:" + strconv.Itoa(局_订单.Id) + ",AppId:" + strconv.Itoa(局_AppId))
		}
	}

}
