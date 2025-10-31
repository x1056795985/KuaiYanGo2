package log

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
	"server/global"
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
