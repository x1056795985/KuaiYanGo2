package log

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/utils/Qqwry"
	"strconv"
	"strings"
	"time"
)

var L_log log

const Log用户消息类型_其他 = 1
const Log用户消息类型_bug提交 = 2
const Log用户消息类型_投诉建议 = 4
const Log用户消息类型_系统执行错误 = 4

const Log风控类型_Api异常调用 = 1

func init() {
	L_log = log{}

}

type log struct {
}

func (j *log) S输出日志(c *gin.Context, logData interface{}) (err error) {
	if logData == nil {
		return nil
	}
	// 开启事务,检测上层是否有事务,如果有直接使用,没有就创建一个
	var tx *gorm.DB
	if c != nil {
		if tempObj, ok := c.Get("tx"); ok {
			tx = tempObj.(*gorm.DB)
		}
	}
	if tx == nil {
		db := *global.GVA_DB
		tx = &db
	}
	//在事务中执行数据库操作，使用的是tx变量，不是db。
	err = tx.Transaction(func(tx *gorm.DB) (err3 error) {
		switch v := logData.(type) {
		default:
			return errors.New(fmt.Sprintf("不支持的日志类型:%v", logData))
		case dbm.DB_LogRMBPayOrder: //支付信息日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogRMBPayOrder{}).Create(&v).Error
		case []dbm.DB_LogRMBPayOrder: //支付信息日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogRMBPayOrder{}).Create(&v[i]).Error
			}

		case dbm.DB_LogVipNumber: //积分点数时间日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogVipNumber{}).Create(&v).Error
		case []dbm.DB_LogVipNumber: //积分点数时间日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogVipNumber{}).Create(&v[i]).Error
			}
		case dbm.DB_LogMoney: //余额日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogMoney{}).Create(&v).Error
		case []dbm.DB_LogMoney: //余额日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogMoney{}).Create(&v[i]).Error
			}
		case dbm.DB_LogKa: //卡号操作日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogKa{}).Create(&v).Error
		case []dbm.DB_LogKa: //卡号操作日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogKa{}).Create(&v[i]).Error
			}
		case dbm.DB_LogLogin: //登录日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogLogin{}).Create(&v).Error
		case []dbm.DB_LogLogin: //登录日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogLogin{}).Create(&v[i]).Error
			}
		case dbm.DB_LogAgentOtherFunc: //代理操作日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogAgentOtherFunc{}).Create(&v).Error
		case []dbm.DB_LogAgentOtherFunc: //代理操作日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogAgentOtherFunc{}).Create(&v[i]).Error
			}
		case dbm.DB_LogUserMsg: //用户消息日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogUserMsg{}).Create(&v).Error
		case []dbm.DB_LogUserMsg: //用户消息日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogUserMsg{}).Create(&v[i]).Error
			}
		case dbm.DB_LogRiskControl: //风控日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(dbm.DB_LogUserMsg{}).Create(&v).Error
		case []dbm.DB_LogRiskControl: //风控日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(dbm.DB_LogUserMsg{}).Create(&v[i]).Error
			}
		}
		return
	})
	return err
}
func isInterfaceAnArray(i interface{}) bool {
	// 获取接口中实际存储的值的 reflect.Value
	value := reflect.ValueOf(i)

	// 检查其 Kind 是否为数组
	return value.Kind() == reflect.Array || value.Kind() == reflect.Slice
}

func (j *log) S上报异常(异常内容 string) (err error) {
	if len(异常内容) >= 10000 {
		return
	}
	global.Q快验.Z置新用户消息(2, 异常内容)
	print(异常内容)
	return err
}

// 用户登陆后调用, 检测登陆日志,当日是否登陆过,当月是否登陆过,如果没有 ,则日活表 值+1
func (j *log) R日活月活增加_登陆处理(AppId int, user string) (err error) {

	db := *global.GVA_DB // 创建用户活跃服务
	//上次登陆日志
	var 时间戳 dbm.DB_LogLogin
	db.Model(dbm.DB_LogLogin{}).Where("LoginType = ? and user = ?  AND (Note=? OR Note =?)", AppId, user, "用户登录", "新用户登录注册").
		Order("Id DESC").First(&时间戳)
	//如果不是今日,则日活+1
	DateStr := time.Now().Format("2006-01-02")
	if 时间戳.Id == 0 || time.Unix(时间戳.Time, 0).Format("2006-01-02") != DateStr {
		db.Model(dbm.DB_LogUserActive{}).Where("AppId = ? and DateStr = ?", AppId, DateStr).UpdateColumn("count", gorm.Expr("count + ?", 1))
	}

	//如果不是今月,则月活+1
	DateStr = time.Now().Format("2006-01")
	if 时间戳.Id == 0 || time.Unix(时间戳.Time, 0).Format("2006-01") != DateStr {
		db.Model(dbm.DB_LogUserActive{}).Where("AppId = ? and DateStr = ?", AppId, DateStr).UpdateColumn("count", gorm.Expr("count + ?", 1))
	}

	return nil
}

// S写风控日志 写风控日志并增加风控分(多表操作: LinksToken表风控分+LogRiskControl表插入)
func (j *log) S写风控日志(c *gin.Context, LId, 风控规则类型 int, User, IP, 风控信息 string) {
	db := *global.GVA_DB
	// 增加风控分
	_ = db.Model(dbm.DB_LinksToken{}).Where("Id=?", LId).Update("RiskControl", gorm.Expr("RiskControl +?", 1)).Error
	// 写风控日志
	login := dbm.DB_LogRiskControl{
		LId:  LId,
		User: User,
		Ip:   IP,
		Time: time.Now().Unix(),
		Type: 风控规则类型,
		Note: 风控信息,
	}
	_ = db.Model(dbm.DB_LogRiskControl{}).Create(&login).Error
}

// Log_写登录日志 写登录日志
func (j *log) Log_写登录日志(User, IP, Note string, LoginType int) {
	login := dbm.DB_LogLogin{
		Id:        0,
		User:      User,
		Ip:        IP + " " + Qqwry.Ip查信息2(IP),
		Time:      time.Now().Unix(),
		LoginType: LoginType,
		Note:      Note,
	}

	err := j.S输出日志(nil, login)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写登录日志失败:%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, LoginType))
	}
	return
}

// Log_写卡号操作日志 msg 支持 变量 {{卡号}} {{卡号索引}} 索引从1开始
// UserType 0 普通用户  1 2 3 级代理  4  管理员  5 系统自动
func (j *log) Log_写卡号操作日志(User, IP, Note string, Ka []string, 卡操作类型, UserType int) {
	logins := make([]dbm.DB_LogKa, 0, len(Ka))
	卡号批次 := strconv.FormatInt(time.Now().UnixNano(), 10)
	for 索引, ka := range Ka {
		login := dbm.DB_LogKa{
			Id:       0,
			User:     User,
			UserType: UserType,
			KaType:   卡操作类型,
			Ka:       ka,
			Ip:       IP + " " + Qqwry.Ip查信息2(IP),
			Time:     time.Now().Unix(),
			Note:     strings.Replace(strings.Replace(strings.Replace(Note, "{{批次id}}", 卡号批次, -1), "{{卡号}}", ka, -1), "{{卡号索引}}", strconv.Itoa(索引+1), -1),
		}
		logins = append(logins, login)
	}
	if len(logins) == 0 {
		return
	}
	err := j.S输出日志(nil, logins)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写卡操作日志失败:%v,%v,%v,%v,%v,%v,%v", err.Error(), User, IP, Note, Ka, 卡操作类型, UserType))
	}
	return
}

// Log_写用户消息 写用户消息日志
func (j *log) Log_写用户消息(消息类型, AppId int, User, App名称, AppVer, 消息内容, IP string) {
	login := dbm.DB_LogUserMsg{
		Id:      0,
		User:    User,
		App:     App名称,
		AppId:   AppId,
		AppVer:  AppVer,
		MsgType: 消息类型,
		Time:    time.Now().Unix(),
		Ip:      IP + " " + Qqwry.Ip查信息2(IP),
		Note:    消息内容,
		IsRead:  false,
	}
	err := j.S输出日志(nil, login)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写用户消息失败:%v,%v,%v,%v,%v,%v", err.Error(), 消息类型, User, App名称, 消息内容, IP))
	}
	return
}

// Log_写余额日志 写余额日志
func (j *log) Log_写余额日志(User, IP, Note string, Count float64) {
	LogMoney := dbm.DB_LogMoney{
		Id:    0,
		User:  User,
		Ip:    IP + " " + Qqwry.Ip查信息2(IP),
		Time:  time.Now().Unix(),
		Count: Count,
		Note:  Note,
	}
	err := j.S输出日志(nil, LogMoney)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写余额日志失败:%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, Count))
	}
	return
}

// Log_写积分点数时间日志 类型 1 积分 2 点数 3 时间
func (j *log) Log_写积分点数时间日志(User, IP, Note string, Count float64, AppId, Type int) {
	DB_LogVipNumber := dbm.DB_LogVipNumber{
		Id:    0,
		User:  User,
		AppId: AppId,
		Type:  Type,
		Ip:    IP + " " + Qqwry.Ip查信息2(IP),
		Time:  time.Now().Unix(),
		Count: Count,
		Note:  Note,
	}
	err := j.S输出日志(nil, DB_LogVipNumber)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写积分点数日志失败:%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, Count))
	}
	return
}

// Log_写库存转移日志 操作库存ID 转出就填原始id,转入就填写,新生成ID
// 类型 1转出,2转入 3创建
func (j *log) Log_写库存转移日志(操作库存ID, 数量, 类型 int, User1 string, User1角色 int, User2 string, User2角色 int, IP, Note string) {

	Log := dbm.Db_Agent_库存日志{
		ID:          0,
		User1:       User1,
		User1Role:   User1角色,
		User2:       User2,
		User2Role:   User2角色,
		Num:         数量,
		Type:        类型,
		InventoryId: 操作库存ID,
		Time:        time.Now().Unix(),
		Note:        Note,
		Ip:          IP,
	}

	db := *global.GVA_DB
	err := db.Model(dbm.Db_Agent_库存日志{}).Create(&Log).Error

	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写库存转移日志:%v,%v,%v,%v,%v,%v,%v,%v,", err.Error(), 操作库存ID, 数量, 类型, User1, User2, IP, Note))
	}
	return
}

// Log_写代理操作日志 写操作日志,主要是代理的操作用户,比如修改用户绑定信息
func (j *log) Log_写代理操作日志(AgentUid, AgentType, AppId, AppUserid int, AppUser string, Func int, IP string, Note string) {
	login := dbm.DB_LogAgentOtherFunc{
		Id:        0,
		AgentType: AgentType,
		AgentUid:  AgentUid,
		AppId:     AppId,
		AppUserid: AppUserid,
		AppUser:   AppUser,
		Func:      Func,
		Note:      Note,
		Ip:        IP + " " + Qqwry.Ip查信息2(IP),
		Time:      time.Now().Unix(),
	}

	err := j.S输出日志(nil, login)
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Log_写操作失败:%v,%v,%v,%v,%v,%v,%v,%v", err.Error(), AgentUid, AgentType, AppId, AppUserid, Func, IP, Note))
	}
	return
}

// Y用户消息_取未读数量 取用户未读消息数量
func (j *log) Y用户消息_取未读数量(User string) int64 {
	var Count int64
	_ = global.GVA_DB.Model(dbm.DB_LogUserMsg{}).Where("IsRead = ?", false).Count(&Count)
	return Count
}
