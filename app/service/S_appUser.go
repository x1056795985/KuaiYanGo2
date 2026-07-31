package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/app/models/db"
	"strconv"
	"time"
)

type AppUser struct {
	db    *gorm.DB
	c     *gin.Context
	appid int
}

// NewAppUser 创建 AppUser 实例
func NewAppUser(c *gin.Context, db *gorm.DB, appId int) *AppUser {
	return &AppUser{
		db:    db,
		appid: appId,
		c:     c,
	}
}

// 增
func (s *AppUser) Create(info *dbm.DB_AppUser) (row int64, err error) {
	//创建会自动重新赋值info.AppId为新插入的数据AppId
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(s.appid)).Create(info)
	return tx.RowsAffected, tx.Error
}

// 业务逻辑操作尽量不用appUserId,容易造成混乱,请使用uid
func (s *AppUser) Info(id int) (info dbm.DB_AppUser, err error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *AppUser) InfoKey(绑定信息 string) (info dbm.DB_AppUser, err error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("`Key` = ?", 绑定信息).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *AppUser) InfoUid(Uid int) (info dbm.DB_AppUser, err error) {

	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Uid = ?", Uid).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *AppUser) Infos(where map[string]interface{}) (info []dbm.DB_AppUser, err error) {
	info = make([]dbm.DB_AppUser, 0)
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(s.appid)).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *AppUser) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
func (s *AppUser) Update2(where map[string]interface{}, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(s.appid)).Where(where).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
func (s *AppUser) UpdateUid(Uid int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Uid = ?", Uid).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// Id点数增减 可能减少到0以下 ,增加无限制
func (s *AppUser) Id点数增减_批量(Id []int, 增减值 int64, is增加 bool) (err error) {
	//因为无符号 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	if len(Id) == 0 {
		return errors.New("Id数组不能为空")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return nil
	}
	sql := "VipTime - ?"
	if is增加 {
		sql = "VipTime + ?"
	}
	err = s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Id IN ?", Id).Update("VipTime", gorm.Expr(sql, 增减值)).Error
	return err

}

// User或卡号取Id 按AppId和用户名/卡号取App用户Id
func (s *AppUser) User或卡号取Id(AppId int, user string) int {
	var id int

	if s.appid == AppId {
		//同应用直接查询
	} else {
		s.appid = AppId
	}

	// 判断应用类型需要查询AppInfo, 这里直接使用SQL避免依赖
	// 如果是卡号模式(3,4), 从db_Ka取Uid; 否则从db_User取Uid
	appInfo := AppInfo{db: s.db, c: s.c}
	if appInfo.App是否为卡号(AppId) {
		// 执行合并后的SQL语句
		s.db.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_Ka` WHERE `Name` = ?) LIMIT 1", user).Scan(&id)
	} else {
		// 执行合并后的SQL语句
		s.db.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_User` WHERE `User` = ?) LIMIT 1", user).Scan(&id)
	}

	return id
}

// User或卡号取Uid 按AppId和用户名/卡号取Uid
func (s *AppUser) User或卡号取Uid(AppId int, user string) int {
	var id int

	appInfo := AppInfo{db: s.db, c: s.c}
	if appInfo.App是否为卡号(AppId) {
		// 执行合并后的SQL语句
		s.db.Raw("SELECT `Id` FROM `db_Ka` WHERE `Name` = ? LIMIT 1", user).Scan(&id)
	} else {
		// 执行合并后的SQL语句
		s.db.Raw("SELECT `Id` FROM `db_User` WHERE `User` = ? LIMIT 1", user).Scan(&id)
	}

	return id
}

// K卡号取Id 按AppId和卡号取App用户Id
func (s *AppUser) K卡号取Id(AppId int, user string) int {
	var id = 0
	s.db.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_Ka` WHERE `Name` = ?) LIMIT 1", user).Scan(&id)
	return id
}

// User取Id 按AppId和用户名取App用户Id
func (s *AppUser) User取Id(AppId int, user string) int {
	var id = 0

	// 执行合并后的SQL语句
	s.db.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_User` WHERE `User` = ?) LIMIT 1", user).Scan(&id)

	return id
}

// Id取Uid 按AppId和Id取Uid
func (s *AppUser) Id取Uid(AppId, id int) int {
	var Uid = 0
	s.db.Raw("SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Id` =  ? LIMIT 1", id).Scan(&Uid)
	return Uid
}

// Id取Uid_批量 按AppId和Id数组取Uid数组
func (s *AppUser) Id取Uid_批量(AppId int, id []int) []int {
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

// Id取User 按AppId和Id取用户名/卡号
func (s *AppUser) Id取User(AppId int, id int) string {
	var 用户名 string
	appInfo := AppInfo{db: s.db, c: s.c}
	if appInfo.App是否为卡号(AppId) {
		s.db.Raw("SELECT `Name` FROM `db_Ka` WHERE Id = (SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE Id = ?  LIMIT 1) LIMIT 1", id).Scan(&用户名)
	} else {
		// 执行合并后的SQL语句
		s.db.Raw("SELECT `User` FROM `db_User` WHERE Id = (SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE Id = ?  LIMIT 1) LIMIT 1", id).Scan(&用户名)
	}

	return 用户名
}

// Uid取User 按AppId和Uid取用户名/卡号
func (s *AppUser) Uid取User(AppId int, Uid int) string {
	var 用户名 string
	appInfo := AppInfo{db: s.db, c: s.c}
	if appInfo.App是否为卡号(AppId) {
		_ = s.db.Model(dbm.DB_Ka{}).Select("Name").Where("Id=?", Uid).First(&用户名)
	} else {
		//从db_User表取用户名
		_ = s.db.Model(dbm.DB_User{}).Select("User").Where("Id=?", Uid).First(&用户名)
	}
	return 用户名
}

// Uid取备注 按AppId和Uid取备注
func (s *AppUser) Uid取备注(AppId int, Uid int) string {
	var 备注 string
	if AppId < 10000 { //屏蔽掉管理平台代理平台等
		return ""
	}
	_ = s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Note").Where("Uid=?", Uid).First(&备注).Error

	return 备注
}

// Uid是否存在 按AppId和Uid是否存在
func (s *AppUser) Uid是否存在(AppId int, Uid int) bool {
	var Count int64
	_ = s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("UId=?", Uid).First(&Count)
	return Count != 0
}

// Id是否存在 按AppId和Id是否存在
func (s *AppUser) Id是否存在(AppId int, Id int) bool {
	var Count int64
	_ = s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("Id=?", Id).Take(&Count)
	return Count != 0
}

// Id取详情 按AppId和Id取详情
func (s *AppUser) Id取详情(AppId int, Id int) (dbm.DB_AppUser, error) {
	var App用户信息 dbm.DB_AppUser
	err := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(int(AppId))).Where("Id=?", Id).First(&App用户信息).Error
	return App用户信息, err
}

// Uid取详情 按AppId和Uid取详情
func (s *AppUser) Uid取详情(AppId int, Uid int) (dbm.DB_AppUser, bool) {
	var App用户信息 dbm.DB_AppUser
	err := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid=?", Uid).First(&App用户信息).Error
	return App用户信息, err == nil
}

// Uid取Id 按AppId和Uid取Id
func (s *AppUser) Uid取Id(AppId int, Uid int) int {
	var App用户ID = 0
	_ = s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Id").Where("Uid=?", Uid).First(&App用户ID).Error
	return App用户ID
}

// Get用户总数 按AppId取用户总数
func (s *AppUser) Get用户总数(AppId int) int {
	var 局_总数 int64
	err := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Count(&局_总数).Error
	if err != nil {
		return 0
	}
	return int(局_总数)
}

// Get用户会员和非会员数量 按AppId取会员和非会员数量
func (s *AppUser) Get用户会员和非会员数量(AppId int) (会员, 非会员 int64) {
	appInfo := AppInfo{db: s.db, c: s.c}
	if appInfo.App是否为计点(AppId) {
		s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Where("VipTime>0").Count(&会员)
		s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Where("VipTime<=0").Count(&非会员)
	} else {
		s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("VipTime>?", time.Now().Unix()).Count(&会员)
		s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("VipTime<=?", time.Now().Unix()).Count(&非会员)
	}

	return 会员, 非会员
}

// B绑定信息是否存在 按AppId和绑定信息是否存在
func (s *AppUser) B绑定信息是否存在(AppId int, 绑定信息 string) bool {
	if 绑定信息 == "" {
		return true
	}
	var Count int64
	_ = s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("`Key` = ?", 绑定信息).Take(&Count)
	return Count != 0
}

// Set绑定信息 设置绑定信息
func (s *AppUser) Set绑定信息(AppId, 用户Uid int, 绑定信息 string) error {
	err := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ? ", 用户Uid).Update("Key", 绑定信息).Error
	if err != nil {
		return err
	}
	return nil
}

// Q取绑定信息 按AppId和用户Uid取绑定信息
func (s *AppUser) Q取绑定信息(AppId, 用户Uid int) string {
	var 绑定信息 = ""
	err := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ? ", 用户Uid).Select("Key").Take(&绑定信息).Error
	if err != nil {
		return ""
	}
	return 绑定信息
}

// Ser用户类型VipTime 设置用户类型和Vip时间
func (s *AppUser) Ser用户类型VipTime(AppId, 用户Uid, 用户类型Id int, VipTime int64) error {
	err := s.db.Model(dbm.DB_AppUser{}).
		Table("db_AppUser_"+strconv.Itoa(AppId)).
		Where("Uid = ? ", 用户Uid).
		Updates(map[string]interface{}{
			"UserClassId": 用户类型Id,
			"VipTime":     VipTime,
		}).Error
	if err != nil {
		return err
	}
	return nil
}

// Id积分增减_批量 批量增减积分(可能减少到0以下)
func (s *AppUser) Id积分增减_批量(AppId int, Id []int, 增减值 float64, is增加 bool) error {
	//因为float64 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	if len(Id) == 0 {
		return errors.New("用户id数组不能为空")
	}
	if 增减值 < 0 {
		return errors.New("增减值不能小于等于0")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return nil
	}

	sql := "VipNumber + ?"
	if !is增加 {
		sql = "VipNumber - ?"
	}
	err := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("VipNumber", gorm.Expr(sql, 增减值)).Error
	return err
}

// X修改用户类型_批量 批量修改用户类型
func (s *AppUser) X修改用户类型_批量(AppId int, Id []int, UserClassId int) (int64, error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("UserClassId", UserClassId)
	return tx.RowsAffected, tx.Error
}

// X修改用户绑定信息_批量 批量修改用户绑定信息
func (s *AppUser) X修改用户绑定信息_批量(AppId int, Id []int, Key string) (int64, error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("Key", Key)
	return tx.RowsAffected, tx.Error
}

// X修改软件用户备注_批量 批量修改软件用户备注
func (s *AppUser) X修改软件用户备注_批量(AppId int, Id []int, Note string) (int64, error) {
	tx := s.db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("Note", Note)
	return tx.RowsAffected, tx.Error
}
