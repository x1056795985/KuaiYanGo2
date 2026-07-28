package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/new/app/utils"
	"strconv"
	"time"
)

type User struct {
	db *gorm.DB
	c  *gin.Context
}

// NewUser 创建 User 实例
func NewUser(c *gin.Context, db *gorm.DB) *User {
	return &User{
		db: db,
		c:  c,
	}
}

// 增
func (s *User) Create(info *dbm.DB_User) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(dbm.DB_User{}).Create(info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *User) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(dbm.DB_User{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(dbm.DB_User{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *User) GetList(请求 request.List, Status int) (int64, []dbm.DB_User, error) {
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
	var 局_数组 []dbm.DB_User
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *User) Info(id int) (info dbm.DB_User, err error) {
	tx := s.db.Model(dbm.DB_User{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *User) Info2(where map[string]interface{}) (info dbm.DB_User, err error) {
	tx := s.db.Model(dbm.DB_User{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *User) Infos(where map[string]interface{}) (info []dbm.DB_User, err error) {
	tx := s.db.Model(dbm.DB_User{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *User) InfoName(name string) (info dbm.DB_User, err error) {
	tx := s.db.Model(dbm.DB_User{}).Where("User = ?", name).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *User) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_User{}).Where("Id = ?", id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

func (s *User) Id取Uid_批量(AppId int, id []int) []int {
	var Uid []int
	// 分批查询，避免占位符超限
	for i := 0; i < len(id); i += 5000 {
		end := i + 5000
		if end > len(id) {
			end = len(id)
		}
		var batch []int
		s.db.Raw("SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Id` IN  ? ", id[i:end]).Scan(&batch)
		Uid = append(Uid, batch...)
	}
	return Uid
}

// UserId是否存在
func (s *User) UserId是否存在(id int) bool {
	var Count int64
	result := s.db.Model(dbm.DB_User{}).Select("1").Where("Id=?", id).Take(&Count)
	return result.Error == nil
}

// User用户名取id
func (s *User) User用户名取id(用户名 string) int {
	if 用户名 == "" {
		return 0
	}
	var Id = 0
	_ = s.db.Model(dbm.DB_User{}).Select("Id").Where("User=?", 用户名).Take(&Id)
	return Id
}

// 负数会取管理员表的信息
func (s *User) Id取User(Id int) string {
	if Id == 0 {
		return ""
	}
	var 用户名 string
	if Id < 0 {
		s.db.Model(dbm.DB_Admin{}).Select("User").Where("Id=?", -Id).Scan(&用户名)
		return 用户名
	}
	err := s.db.Model(dbm.DB_User{}).Select("User").Where("Id=?", Id).Scan(&用户名).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	return 用户名
}

// 取用户表的信息_批量,仅限用户表
func (s *User) Id取User_批量(Id []int) map[int]string {
	if len(Id) == 0 {
		return map[int]string{}
	}
	var 用户名 []dbm.DB_User
	s.db.Model(dbm.DB_User{}).Select("Id,User").Where("Id IN ?", Id).Find(&用户名)
	var 局_返回 = make(map[int]string, len(用户名))
	for 索引, _ := range 用户名 {
		局_返回[用户名[索引].Id] = 用户名[索引].User
	}
	return 局_返回
}

// 负数会取管理员表的信息
func (s *User) Id取状态(Id int) int {
	if Id == 0 {
		return 1
	}
	var Status int
	if Id < 0 {
		s.db.Model(dbm.DB_Admin{}).Select("Status").Where("Id=?", -Id).First(&Status)
		return Status
	}
	s.db.Model(dbm.DB_User{}).Select("Status").Where("Id=?", Id).First(&Status)
	return Status
}

func (s *User) User取详情(User string) (用户详情 dbm.DB_User, ok bool) {
	err := s.db.Model(dbm.DB_User{}).Where("User=?", User).First(&用户详情).Error
	return 用户详情, err == nil
}

func (s *User) Id取详情(Id int) (用户详情 dbm.DB_User, ok bool) {
	err := s.db.Model(dbm.DB_User{}).Where("Id=?", Id).First(&用户详情).Error
	return 用户详情, err == nil
}

func (s *User) Id取详情_数组(Id []int) ([]dbm.DB_User, error) {
	var 局_用户详情 = make([]dbm.DB_User, 0, len(Id))
	if len(Id) == 0 {
		return 局_用户详情, nil
	}
	err := s.db.Model(dbm.DB_User{}).Where("Id IN ?", Id).Find(&局_用户详情).Error
	return 局_用户详情, err
}

func (s *User) Id取余额(Id int) (余额 float64) {
	_ = s.db.Model(dbm.DB_User{}).Select("Rmb").Where("Id=?", Id).First(&余额).Error
	return
}

func (s *User) Id置最后登录AppId(Id, AppId int, Ip string) {
	if Id == 0 {
		return
	}
	err := s.db.Model(dbm.DB_User{}).Where("Id = ?", Id).Updates(map[string]interface{}{"LoginAppid": AppId, "LoginIp": Ip, "LoginTime": time.Now().Unix()}).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Id置最后登录AppId失败ID:%v,AppId,%v,Ip,%v,%v", Id, AppId, Ip, err.Error()))
	}
	return
}

func (s *User) Id置QQ邮箱手机号(Id int, QQ, 邮箱, 手机号 string) error {
	if Id == 0 {
		return errors.New("id不能为空")
	}
	局data := map[string]interface{}{}
	if QQ != "" {
		局data["Qq"] = QQ
	}
	if 邮箱 != "" {
		局data["Email"] = 邮箱
	}
	if 手机号 != "" {
		局data["Phone"] = 手机号
	}
	err := s.db.Model(dbm.DB_User{}).Where("Id = ?", Id).Updates(&局data).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Id置QQ邮箱手机号失败ID:%v,%v,%v,%v,%v", Id, QQ, 邮箱, 手机号, err.Error()))
		return err
	}
	return nil
}

func (s *User) Id置新密码(Id int, NewPassWord string) error {
	if Id == 0 {
		return errors.New("Id不能为0")
	}
	err := s.db.Model(dbm.DB_User{}).Where("Id = ?", Id).Updates(map[string]interface{}{"PassWord": utils.Md5String(NewPassWord)}).Error
	if err != nil {
		global.GVA_LOG.Println(fmt.Sprintf("Id置新密码失败:%v,%v,%v", Id, NewPassWord, err.Error()))
		return errors.New("修改密码失败")
	}
	return nil
}

func (s *User) Q取总数() int64 {
	var 局_总数 int64
	_ = s.db.Model(dbm.DB_User{}).Count(&局_总数).Error
	return 局_总数
}

func (s *User) Id取上级代理ID(Id int) int {
	if Id == 0 {
		return 0
	}
	var 上级代理ID int
	s.db.Model(dbm.DB_User{}).Select("UPAgentId").Where("Id=?", Id).First(&上级代理ID)
	return 上级代理ID
}

func (s *User) Id取下级代理分成最高(Id int) int {
	if Id == 0 {
		return 0
	}
	var 上级代理ID = 0
	s.db.Model(dbm.DB_User{}).Select(" Max(AgentDiscount) ").Where("UPAgentId=?", Id).First(&上级代理ID)
	return 上级代理ID
}

// 0 非代理,1 一级代理 2 二级代理 3 三级代理
func (s *User) 取Id代理级别(用户ID int) int {
	var Count int64 = 0
	s.db.Model(dbm.Db_Agent_Level{}).Where("Uid=?", 用户ID).Count(&Count)
	return int(Count)
}
