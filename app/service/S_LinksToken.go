package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/models/request"
	"server/app/utils/Qqwry"
	"strings"
	"time"
)

// 在线列表 数据库处理
type LinksToken struct {
	db *gorm.DB
	c  *gin.Context
}

// NewLinksToken 创建 LinksToken 实例
func NewLinksToken(c *gin.Context, db *gorm.DB) *LinksToken {
	return &LinksToken{
		db: db,
		c:  c,
	}
}

// DeleteExpiredTokens 删除已过期的 token
func (s *LinksToken) S删除已过期的Token() error {
	// 删除已注销并 6 小时没活动的 token
	tx := s.db.Model(dbm.DB_LinksToken{}).Where("Status = 2").Where("LastTime < ?", time.Now().Unix()-21600).Delete("")
	return tx.Error
}

// RevokeExpiredTokens 定时注销已过期的 token
func (s *LinksToken) Z注销已过期的Token() error {
	// 注销超时的 token
	tx := s.db.Model(dbm.DB_LinksToken{}).Where("Status = 1").Where("LastTime + OutTime < ?", time.Now().Unix()).Updates(map[string]interface{}{"Status": 2, "LogoutCode": constant.Z注销_心跳超时自动注销})
	return tx.Error
}

// 增
func (s *LinksToken) Create(info dbm.DB_LinksToken) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(dbm.DB_LinksToken{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *LinksToken) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(dbm.DB_LinksToken{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(dbm.DB_LinksToken{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *LinksToken) GetList(请求 request.List, Status int) (int64, []dbm.DB_LinksToken, error) {
	tx := s.db
	if Status > 0 {
		tx = tx.Where("Status = ?", Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			tx = tx.Where("Id = ?", 请求.Keywords)
		case 2: //任务名称
			tx = tx.Where("Name LIKE ? ", "%"+请求.Keywords+"%")
		}
	}
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		tx.Count(&总数)
	}
	//处理排序
	switch 请求.Order {
	default:
		tx = tx.Order("Id ASC")
	case 2:
		tx = tx.Order("Id DESC")
	}
	var 局_数组 []dbm.DB_LinksToken
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *LinksToken) Info(id int) (info dbm.DB_LinksToken, err error) {
	tx := s.db.Model(dbm.DB_LinksToken{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *LinksToken) Infos(where map[string]interface{}) (info []dbm.DB_LinksToken, err error) {
	tx := s.db.Model(dbm.DB_LinksToken{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *LinksToken) InfosId排序(where map[string]interface{}) (info []dbm.DB_LinksToken, err error) {
	tx := s.db.Model(dbm.DB_LinksToken{}).Where(where).Order("Id ASC").Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *LinksToken) Update(id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_LinksToken{}).Where("Id = ?", id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 可指定AppId,0为全部注销
func (s *LinksToken) Set批量注销Uid数组(UId []int, AppId int, 注销原因 int) (err error) {
	db := s.db.Model(dbm.DB_LinksToken{}).Where("UId IN ? ", UId)
	if AppId != 0 {
		db.Where("LoginAppid =? ", AppId)
	}
	err = db.Updates(map[string]interface{}{"OutTime": 0, "Status": 2, "LogoutCode": 注销原因}).Error
	return
}

// 查
func (s *LinksToken) InfoToken(Token string) (info dbm.DB_LinksToken, err error) {
	tx := s.db.Model(dbm.DB_LinksToken{}).Where("Token = ?", Token).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *LinksToken) Updates(ids []int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_LinksToken{}).Where("Id IN ?", ids).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// Token取Name 按Token取用户名
func (s *LinksToken) Token取Name(Token string) string {
	var User string = ""
	s.db.Model(dbm.DB_LinksToken{}).Select("User").Where("Token=?", Token).First(&User)
	return User
}

// Token取User在线详情 按Token取在线详情
func (s *LinksToken) Token取User在线详情(Token string) (LinksToken dbm.DB_LinksToken, err error) {
	err = s.db.Model(dbm.DB_LinksToken{}).Where("Token=?", Token).First(&LinksToken).Error
	return LinksToken, err
}

// Lid增减风控分 按Lid增减风控分
func (s *LinksToken) Lid增减风控分(Lid, 风控分 int) (LinksToken dbm.DB_LinksToken, err error) {
	err = s.db.Model(dbm.DB_LinksToken{}).Where("Id=?", Lid).Update("RiskControl", gorm.Expr("RiskControl +?", 风控分)).Error
	return LinksToken, err
}

// New 创建新的在线Token
func (s *LinksToken) New(Uid, Status, LoginAppid, OutTIme int, User, Tab, Key, Ip, CryptoKeyAes string) (dbm.DB_LinksToken, error) {
	var DB_links_user dbm.DB_LinksToken
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
	DB_links_user.LoginTime = time.Now().Unix()
	DB_links_user.OutTime = OutTIme //退出时间 半小时
	DB_links_user.LastTime = DB_links_user.LoginTime

	DB_links_user.Token = strings.ToUpper(rand_string.RandomLetter(32))
	DB_links_user.LoginAppid = LoginAppid     //管理员后台代号1
	DB_links_user.CryptoKeyAes = CryptoKeyAes //通讯key
	err = s.db.Create(&DB_links_user).Error
	return DB_links_user, err
}

// NewWebApiToken 创建WebApi Token
func (s *LinksToken) NewWebApiToken(OutTIme int, Key, Tab string) (dbm.DB_LinksToken, error) {
	var DB_links_user dbm.DB_LinksToken
	DB_links_user.Uid = 0
	DB_links_user.User = strings.ToUpper(rand_string.RandomLetter(32))
	DB_links_user.Tab = Tab
	DB_links_user.Key = Key
	DB_links_user.Ip = ""
	DB_links_user.Status = 1
	DB_links_user.LoginTime = time.Now().Unix()
	DB_links_user.OutTime = OutTIme
	DB_links_user.LastTime = DB_links_user.LoginTime
	DB_links_user.Token = DB_links_user.User
	DB_links_user.LoginAppid = constant.APPID_WebApi
	DB_links_user.CryptoKeyAes = "" //通讯key
	err := s.db.Model(dbm.DB_LinksToken{}).Create(&DB_links_user).Error
	//因为有业务(任务池提交任务,获取任务列表)需要通过uid 判断是否为登陆状态,所以uid 必须大于0 这里直接设置为id
	s.db.Model(dbm.DB_LinksToken{}).Where("Id = ?", DB_links_user.Id).Update("Uid", DB_links_user.Id)

	return DB_links_user, err
}

// Set自动注销超时时间 设置自动注销超时时间
func (s *LinksToken) Set自动注销超时时间(OutTIme int, id []int) error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("id IN ?", id).Updates(map[string]interface{}{"OutTime": OutTIme}).Error
	return err
}

// Token更新最后活动时间 按Token更新最后活动时间
func (s *LinksToken) Token更新最后活动时间(Token string) {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Token = ?", Token).Update("LastTime", int(time.Now().Unix())).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Token更新最后活动时间失败:%v,%v", err.Error(), Token))
	}
	return
}

// Token更新在线ip 按Token更新在线ip
func (s *LinksToken) Token更新在线ip(Token, Ip string) {
	省市, 运行商, err := Qqwry.Ip查信息(Ip)
	var IPCity = ""
	if err == nil && 省市 != "" {
		IPCity = 省市 + " " + 运行商
	}
	err = s.db.Model(dbm.DB_LinksToken{}).Where("Token = ?", Token).Updates(map[string]interface{}{"Ip": Ip, "IPCity": IPCity}).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Token更新在线ip:%v,%v", err.Error(), Token))
	}
	return
}

// Id更新当前版本号 按Id更新当前版本号
func (s *LinksToken) Id更新当前版本号(Id int, 新应用版本号 string) {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Id = ?", Id).Update("AppVer", 新应用版本号).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Id更新当前版本号失败:%v,%v", err.Error(), 新应用版本号))
	}
	return
}

// Token风控分增减 按Token增减风控分
func (s *LinksToken) Token风控分增减(Token string, 增减值 int) {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Token = ?", Token).Update("RiskControl", gorm.Expr("RiskControl + ?", 增减值)).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Token风控分增减失败:%v,%v", err.Error(), Token))
	}
	return
}

// Set在线登录信息 设置在线登录信息
func (s *LinksToken) Set在线登录信息(Id, Uid int, 用户名, 绑定信息, 动态标签, 软件版本 string) error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Id = ?", Id).Updates(map[string]interface{}{"Uid": Uid, "User": 用户名, "Key": 绑定信息, "Tab": 动态标签, "AppVer": 软件版本}).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Set在线登录信息:%v,%v,%v,%v,%v", err.Error(), Id, 用户名, 绑定信息, 动态标签))
	}
	return err
}

// Get取在线数量 按AppId和Uid取在线数量
func (s *LinksToken) Get取在线数量(AppId, Uid int) []int {
	//返回数组排序为 先登录的在前面  id 即可也是值小的先登录 不用特意用时间排序
	var 局_在线ID []int
	_ = s.db.Model(dbm.DB_LinksToken{}).Select("Id").Where("Uid = ?", Uid).Where("Status = 1").Where("LoginAppid  = ?", AppId).Order("Id  ASC").Find(&局_在线ID).Error
	return 局_在线ID
}

// Get取在线总数 取在线总数
func (s *LinksToken) Get取在线总数(排除游客, 仅限正常状态 bool) int64 {

	//返回数组排序为 先登录的在前面  id 即可也是值小的先登录 不用特意用时间排序
	var 局_在线总数 int64
	db := s.db.Model(dbm.DB_LinksToken{})
	if 排除游客 {
		db.Where("User!=?", "游客")
	}

	if 仅限正常状态 {
		db.Where("Status=1")
	}
	_ = db.Count(&局_在线总数).Error
	return 局_在线总数
}

// Q指定应用真实在线 指定应用真实在线数量
func (s *LinksToken) Q指定应用真实在线(AppId int) int64 {
	var 局_在线总数 int64
	_ = s.db.Model(dbm.DB_LinksToken{}).Where("LoginAppid=?", AppId).Where("Status=1").Where("User!=?", "游客").Count(&局_在线总数).Error
	return 局_在线总数
}

// Set批量注销 批量注销
func (s *LinksToken) Set批量注销(Id []int, 注销原因 int) error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Id IN ? ", Id).Updates(map[string]interface{}{"OutTime": 0, "Status": 2, "LogoutCode": 注销原因}).Error
	return err
}

// Set批量注销Uid 按Uid批量注销
func (s *LinksToken) Set批量注销Uid(UId int, 注销原因 int) error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("UId = ? ", UId).Updates(map[string]interface{}{"OutTime": 0, "Status": 2, "LogoutCode": 注销原因}).Error
	return err
}

// Set批量注销全部代理 批量注销全部代理
func (s *LinksToken) Set批量注销全部代理() error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("LoginAppid = 2 ").Updates(map[string]interface{}{"OutTime": 0, "Status": 2}).Error
	return err
}

// Set批量注销User数组 按User数组批量注销
func (s *LinksToken) Set批量注销User数组(User []string, 注销原因 int) error {
	db := s.db.Model(dbm.DB_LinksToken{}).Where("User IN ? ", User)
	err := db.Updates(map[string]interface{}{"OutTime": 0, "Status": 2, "LogoutCode": 注销原因}).Error
	return err
}

// Set动态标签 设置动态标签
func (s *LinksToken) Set动态标签(Id int, 新动态标签 string) error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Id = ? ", Id).Updates(map[string]interface{}{"Tab": 新动态标签}).Error
	return err
}

// Set代理标志 设置代理标志
func (s *LinksToken) Set代理标志(Id int, 代理Uid int) error {
	err := s.db.Model(dbm.DB_LinksToken{}).Where("Id = ? ", Id).Updates(map[string]interface{}{"AgentUid": 代理Uid}).Error
	return err
}
