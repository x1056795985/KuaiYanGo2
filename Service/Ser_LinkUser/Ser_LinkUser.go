package Ser_LinkUser

import (
	"fmt"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"gorm.io/gorm"
	"server/global"
	DB "server/structs/db"
	"server/utils/Qqwry"
	"strings"
	"time"
)

func Token取Name(Token string) string {
	var User string = ""
	global.GVA_DB.Model(DB.DB_LinksToken{}).Select("User").Where("Token=?", Token).First(&User)
	return User
}

func Token取User在线详情(Token string) (LinksToken DB.DB_LinksToken, err error) {
	err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token=?", Token).First(&LinksToken).Error
	return LinksToken, err
}

func Lid增减风控分(Lid, 风控分 int) (LinksToken DB.DB_LinksToken, err error) {
	err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id=?", Lid).Update("RiskControl", gorm.Expr("RiskControl +?", 风控分)).Error
	return LinksToken, err
}
func New(Uid, Status, LoginAppid, OutTIme int, User, Tab, Key, Ip, CryptoKeyAes string) (DB.DB_LinksToken, error) {
	var DB_links_user DB.DB_LinksToken
	DB_links_user.Uid = Uid
	DB_links_user.User = User
	DB_links_user.Tab = Tab
	DB_links_user.Key = Key
	DB_links_user.Ip = Ip
	省市, 运行商, err := Qqwry.Ip查信息(DB_links_user.Ip)
	if err == nil && 省市 != "" {
		DB_links_user.IPCity = 省市 + " " + 运行商
	}
	DB_links_user.Status = Status
	DB_links_user.LoginTime = int64(time.Now().Unix())
	DB_links_user.OutTime = OutTIme //退出时间 半小时
	DB_links_user.LastTime = DB_links_user.LoginTime
	DB_links_user.Token = strings.ToUpper(rand_string.RandStringBytesMaskImprSrc(32))
	DB_links_user.LoginAppid = LoginAppid     //管理员后台代号1
	DB_links_user.CryptoKeyAes = CryptoKeyAes //通讯key
	err = global.GVA_DB.Create(&DB_links_user).Error
	return DB_links_user, err
}
func NewWebApiToken(OutTIme int, Key, Tab string) (DB.DB_LinksToken, error) {
	var DB_links_user DB.DB_LinksToken
	DB_links_user.Uid = 0
	DB_links_user.User = strings.ToUpper(rand_string.RandStringBytesMaskImprSrc(32))
	DB_links_user.Tab = Tab
	DB_links_user.Key = Key
	DB_links_user.Ip = ""
	DB_links_user.Status = 1
	DB_links_user.LoginTime = time.Now().Unix()
	DB_links_user.OutTime = OutTIme
	DB_links_user.LastTime = DB_links_user.LoginTime
	DB_links_user.Token = DB_links_user.User
	DB_links_user.LoginAppid = 3
	DB_links_user.CryptoKeyAes = "" //通讯key
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Create(&DB_links_user).Error
	return DB_links_user, err
}
func Token更新最后活动时间(Token string) {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).Update("LastTime", int(time.Now().Unix())).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Token更新最后活动时间失败:%v,%v", err.Error(), Token))
	}
	return
}
func Token更新在线ip(Token, Ip string) {
	省市, 运行商, err := Qqwry.Ip查信息(Ip)
	var IPCity = ""
	if err == nil && 省市 != "" {
		IPCity = 省市 + " " + 运行商
	}
	err = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).Updates(map[string]interface{}{"Ip": Ip, "IPCity": IPCity}).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Token更新在线ip:%v,%v", err.Error(), Token))
	}
	return
}
func Id更新当前版本号(Id int, 新应用版本号 string) {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", Id).Update("AppVer", 新应用版本号).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Id更新当前版本号失败:%v,%v", err.Error(), 新应用版本号))
	}
	return
}

func Token风控分增减(Token string, 增减值 int) {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).Update("RiskControl", gorm.Expr("RiskControl + ?", 增减值)).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Token风控分增减失败:%v,%v", err.Error(), Token))
	}
	return
}
func Set在线登录信息(Id, Uid int, 用户名, 绑定信息, 动态标签, 软件版本 string) error {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", Id).Updates(map[string]interface{}{"Uid": Uid, "User": 用户名, "Key": 绑定信息, "Tab": 动态标签, "AppVer": 软件版本}).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Set在线登录信息:%v,%v,%v,%v", err.Error(), Id, 用户名, 绑定信息, 动态标签))
	}
	return err
}

func Get取在线数量(AppId, Uid int) []int {
	//返回数组排序为 先登录的在前面  id 即可也是值小的先登录 不用特意用时间排序
	var 局_在线ID []int
	_ = global.GVA_DB.Model(DB.DB_LinksToken{}).Select("Id").Where("Uid = ?", Uid).Where("Status = 1").Where("LoginAppid  = ?", AppId).Order("Id  ASC").Find(&局_在线ID).Error
	return 局_在线ID
}
func Get取在线总数(排除游客, 仅限正常状态 bool) int64 {

	//返回数组排序为 先登录的在前面  id 即可也是值小的先登录 不用特意用时间排序
	var 局_在线总数 int64
	db := global.GVA_DB.Model(DB.DB_LinksToken{})
	if 排除游客 {
		db.Where("User!=?", "游客")
	}

	if 仅限正常状态 {
		db.Where("Status=1")
	}
	_ = db.Count(&局_在线总数).Error
	return 局_在线总数
}
func Q指定应用真实在线(AppId int) int64 {
	var 局_在线总数 int64
	_ = global.GVA_DB.Model(DB.DB_LinksToken{}).Where("LoginAppid=?", AppId).Where("Status=1").Where("User!=?", "游客").Count(&局_在线总数).Error
	return 局_在线总数
}
func Set批量注销(Id []int) error {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id IN ? ", Id).Updates(map[string]interface{}{"OutTime": 0, "Status": 2}).Error
	return err
}
func Set批量注销Uid(UId int) error {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("UId = ? ", UId).Updates(map[string]interface{}{"OutTime": 0, "Status": 2}).Error
	return err
}
func Set批量注销全部代理() error {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("LoginAppid = 2 ").Updates(map[string]interface{}{"OutTime": 0, "Status": 2}).Error
	return err
}

// 可指定AppId,0为全部注销
func Set批量注销Uid数组(UId []int, AppId int) error {
	db := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("UId IN ? ", UId)
	if AppId != 0 {
		db.Where("LoginAppid =? ", AppId)
	}
	err := db.Updates(map[string]interface{}{"OutTime": 0, "Status": 2}).Error
	return err
}
func Set批量注销User数组(User []string) error {
	db := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("User IN ? ", User)
	err := db.Updates(map[string]interface{}{"OutTime": 0, "Status": 2}).Error
	return err
}
func Set动态标签(Id int, 新动态标签 string) error {
	err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ? ", Id).Updates(map[string]interface{}{"Tab": 新动态标签}).Error
	return err
}
