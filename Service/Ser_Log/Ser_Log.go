package Ser_Log

import (
	"fmt"
	"server/Service/Ser_LinkUser"
	"server/global"
	DB "server/structs/db"
	"server/utils/Qqwry"
	"strconv"
	"strings"
	"time"
)

func Log_写登录日志(User, IP, Note string, LoginType int) {
	login := DB.DB_LogLogin{
		Id:        0,
		User:      User,
		Ip:        IP + " " + Qqwry.Ip查信息2(IP),
		Time:      time.Now().Unix(),
		LoginType: LoginType,
		Note:      Note,
	}

	err := global.GVA_DB.Model(DB.DB_LogLogin{}).Create(&login).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写登录日志失败:%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, LoginType))
	}
	return
}

// msg 支持 变量 {{卡号}} {{卡号索引}} 索引从1开始
// UserType 0 普通用户  1 2 3 级代理  4  管理员  5 系统自动
func Log_写卡号操作日志(User, IP, Note string, Ka []string, 卡操作类型, UserType int) {
	logins := make([]DB.DB_LogKa, 0, len(Ka))
	for 索引, ka := range Ka {
		login := DB.DB_LogKa{
			Id:       0,
			User:     User,
			UserType: UserType,
			KaType:   卡操作类型,
			Ka:       ka,
			Ip:       IP + " " + Qqwry.Ip查信息2(IP),
			Time:     time.Now().Unix(),
			Note:     strings.Replace(strings.Replace(Note, "{{卡号}}", ka, -1), "{{卡号索引}}", strconv.Itoa(索引+1), -1),
		}
		logins = append(logins, login)
	}
	err := global.GVA_DB.Model(DB.DB_LogKa{}).Create(&logins).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写卡操作日志失败:%v,%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, Ka, 卡操作类型, UserType))
	}
	return
}

const Log风控类型_Api异常调用 = 1

func Log_写风控日志(LId, 风控规则类型 int, User, IP, 风控信息 string) {
	Ser_LinkUser.Lid增减风控分(LId, 1)
	login := DB.DB_LogRiskControl{
		Id:   0,
		LId:  LId,
		User: User,
		Ip:   IP + " " + Qqwry.Ip查信息2(IP),
		Time: int(time.Now().Unix()),
		Type: 风控规则类型,
		Note: 风控信息,
	}
	err := global.GVA_DB.Model(DB.DB_LogRiskControl{}).Create(&login).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写登录日志失败:%v,%v,%v,%v,%v,%v", err.Error(), LId, 风控规则类型, User, IP, 风控信息))
	}
	return
}

const Log用户消息类型_其他 = 1
const Log用户消息类型_bug提交 = 2
const Log用户消息类型_投诉建议 = 4
const Log用户消息类型_系统执行错误 = 4

func Log_写用户消息(消息类型 int, User, App名称, AppVer, 消息内容, IP string) {
	login := DB.DB_LogUserMsg{
		Id:           0,
		User:         User,
		App:          App名称,
		AppVer:       AppVer,
		MsgType:      消息类型,
		Time:         int(time.Now().Unix()),
		Ip:           IP + " " + Qqwry.Ip查信息2(IP),
		Note:         消息内容,
		IsReadIsRead: false,
	}
	err := global.GVA_DB.Model(DB.DB_LogUserMsg{}).Create(&login).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写用户消息失败:%v,%v,%v,%v,%v,%v", err.Error(), 消息类型, User, App名称, 消息内容, IP))
	}
	return
}

func Log_写余额日志(User, IP, Note string, Count float64) {
	LogMoney := DB.DB_LogMoney{
		Id:    0,
		User:  User,
		Ip:    IP + " " + Qqwry.Ip查信息2(IP),
		Time:  int(time.Now().Unix()),
		Count: Count,
		Note:  Note,
	}
	err := global.GVA_DB.Model(DB.DB_LogMoney{}).Create(&LogMoney).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写余额日志失败:%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, Count))
	}
	return
}

// Log_写积分点数时间日志 类型 1 积分 2 点数 3 时间
func Log_写积分点数时间日志(User, IP, Note string, Count float64, AppId, Type int) {
	DB_LogVipNumber := DB.DB_LogVipNumber{
		Id:    0,
		User:  User,
		AppId: AppId,
		Type:  Type,
		Ip:    IP + " " + Qqwry.Ip查信息2(IP),
		Time:  int(time.Now().Unix()),
		Count: Count,
		Note:  Note,
	}
	err := global.GVA_DB.Model(DB.DB_LogVipNumber{}).Create(&DB_LogVipNumber).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写积分点数日志失败:%v,%v,%v,%v,%v,", err.Error(), User, IP, Note, Count))
	}
	return
}

// 操作库存ID 转出就填原始id,转入就填写,新生成ID
// 类型 1转出,2转入 3创建
func Log_写库存转移日志(操作库存ID, 数量, 类型 int, User1 string, User1角色 int, User2 string, User2角色 int, IP, Note string) {

	Log := DB.Db_Agent_库存日志{
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
		Ip:          IP + " " + Qqwry.Ip查信息2(IP),
	}
	err := global.GVA_DB.Model(DB.Db_Agent_库存日志{}).Create(&Log).Error

	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Log_写库存转移日志:%v,%v,%v,%v,%v,", err.Error(), 操作库存ID, 数量, 类型, User1, User2, IP, Note))
	}
	return
}
