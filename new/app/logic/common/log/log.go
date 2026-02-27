package log

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
	"server/global"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
	"time"
)

var L_log log

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
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}
	//在事务中执行数据库操作，使用的是tx变量，不是db。
	err = tx.Transaction(func(tx *gorm.DB) (err3 error) {
		switch v := logData.(type) {
		default:
			return errors.New(fmt.Sprintf("不支持的日志类型:%v", logData))
		case DB.DB_LogRMBPayOrder: //支付信息日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogRMBPayOrder{}).Create(&v).Error
		case []DB.DB_LogRMBPayOrder: //支付信息日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogRMBPayOrder{}).Create(&v[i]).Error
			}

		case DB.DB_LogVipNumber: //积分点数时间日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogVipNumber{}).Create(&v).Error
		case []DB.DB_LogVipNumber: //积分点数时间日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogVipNumber{}).Create(&v[i]).Error
			}
		case DB.DB_LogMoney: //余额日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogMoney{}).Create(&v).Error
		case []DB.DB_LogMoney: //余额日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogMoney{}).Create(&v[i]).Error
			}
		case DB.DB_LogKa: //卡号操作日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogKa{}).Create(&v).Error
		case []DB.DB_LogKa: //卡号操作日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogKa{}).Create(&v[i]).Error
			}
		case DB.DB_LogLogin: //登录日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogLogin{}).Create(&v).Error
		case []DB.DB_LogLogin: //登录日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogLogin{}).Create(&v[i]).Error
			}
		case DB.DB_LogAgentOtherFunc: //代理操作日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogAgentOtherFunc{}).Create(&v).Error
		case []DB.DB_LogAgentOtherFunc: //代理操作日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogAgentOtherFunc{}).Create(&v[i]).Error
			}
		case DB.DB_LogUserMsg: //用户消息日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogUserMsg{}).Create(&v).Error
		case []DB.DB_LogUserMsg: //用户消息日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogUserMsg{}).Create(&v[i]).Error
			}
		case DB.DB_LogRiskControl: //风控日志
			if v.Time == 0 {
				v.Time = time.Now().Unix()
			}
			err3 = tx.Model(DB.DB_LogUserMsg{}).Create(&v).Error
		case []DB.DB_LogRiskControl: //风控日志
			for i := range v {
				if v[i].Time == 0 {
					v[i].Time = time.Now().Unix()
				}
				err3 = tx.Model(DB.DB_LogUserMsg{}).Create(&v[i]).Error
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
	var 时间戳 DB.DB_LogLogin
	db.Model(DB.DB_LogLogin{}).Where("LoginType = ? and user = ?  AND (Note=? OR Note =?)", AppId, user, "用户登录", "新用户登录注册").
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
