package Ser_AppUser

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"server/Service/Ser_AppInfo"
	"server/global"
	DB "server/structs/db"
	"strconv"
	"time"
)

func User或卡号取Id(AppId int, user string) int {
	var id int

	if Ser_AppInfo.App是否为卡号(AppId) {
		// 执行合并后的SQL语句
		global.GVA_DB.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_Ka` WHERE `User` = ?) LIMIT 1", user).Scan(&id)

	} else {
		// 执行合并后的SQL语句
		global.GVA_DB.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_User` WHERE `User` = ?) LIMIT 1", user).Scan(&id)

	}

	return id
}

func K卡号取Id(AppId int, user string) int {
	var id = 0
	global.GVA_DB.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_Ka` WHERE `Name` = ?) LIMIT 1", user).Scan(&id)
	return id
}
func User取Id(AppId int, user string) int {
	var id = 0

	// 执行合并后的SQL语句
	global.GVA_DB.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_User` WHERE `User` = ?) LIMIT 1", user).Scan(&id)

	return id
}
func Id取Uid(AppId, id int) int {
	var Uid = 0
	global.GVA_DB.Raw("SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Id` =  ? LIMIT 1", id).Scan(&Uid)
	return Uid
}

func Id取User(AppId int, id int) string {
	var 用户名 string
	if Ser_AppInfo.App是否为卡号(AppId) {
		// 执行合并后的SQL语句
		global.GVA_DB.Raw("SELECT `User` FROM `db_User` WHERE Id = (SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE Id = ?  LIMIT 1) LIMIT 1", id).Scan(&用户名)
	} else {
		global.GVA_DB.Raw("SELECT `Name` FROM `db_Ka` WHERE Id = (SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE Id = ?  LIMIT 1) LIMIT 1", id).Scan(&用户名)
	}

	return 用户名
	//下边是屎
	/*	var Uid = 0
		global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Uid").Where("Id=?", id).First(&Uid)
		if Uid == 0 {
			return ""
		}
		var 用户名 = ""
		global.GVA_DB.Model(DB.DB_User{}).Select("User").Where("Id=?", Uid).First(&用户名)

		return 用户名*/
}
func Uid是否存在(AppId int, Uid int) bool {
	var Count int64
	_ = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("UId=?", Uid).First(&Count)
	return Count != 0

}

func Id是否存在(AppId int, Id int) bool {
	var Count int64
	_ = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("Id=?", Id).Take(&Count)
	return Count != 0

}
func Id取详情(AppId int, Id int) (DB.DB_AppUser, error) {
	var App用户信息 DB.DB_AppUser
	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(int(AppId))).Where("Id=?", Id).First(&App用户信息).Error
	return App用户信息, err
}
func Uid取详情(AppId int, Uid int) (DB.DB_AppUser, bool) {
	var App用户信息 DB.DB_AppUser
	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid=?", Uid).First(&App用户信息).Error
	return App用户信息, err == nil
}
func Uid取Id(AppId int, Uid int) int {
	var App用户ID = 0
	_ = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Id").Where("Uid=?", Uid).First(&App用户ID).Error
	return App用户ID
}
func Get用户总数(AppId int) int {
	var 局_总数 int64
	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Count(&局_总数).Error
	if err != nil {
		return 0
	}
	return int(局_总数)
}
func Get用户会员和非会员数量(AppId int) (会员, 非会员 int64) {
	if Ser_AppInfo.App是否为计点(AppId) {
		global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Where("VipTime>0").Count(&会员)
		global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Where("VipTime<=0").Count(&非会员)
	} else {
		global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("VipTime>?", time.Now().Unix()).Count(&会员)
		global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("VipTime<=?", time.Now().Unix()).Count(&非会员)
	}

	return 会员, 非会员
}

// New用户信息
func New用户信息(AppId int, Uid int, 绑定信息 string, 最大在线数量 int, VipTime int64, VipNumber float64) error {
	var 局_AppUser DB.DB_AppUser

	局_AppUser.Id = 0
	局_AppUser.Uid = Uid
	局_AppUser.Status = 1
	局_AppUser.Key = 绑定信息
	局_AppUser.VipTime = VipTime
	局_AppUser.VipNumber = VipNumber
	局_AppUser.Note = ""
	局_AppUser.MaxOnline = 最大在线数量
	局_AppUser.UserClassId = 0
	局_AppUser.RegisterTime = int(time.Now().Unix())

	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Create(&局_AppUser).Error
	return err
}

func B绑定信息是否存在(AppId int, 绑定信息 string) bool {
	if 绑定信息 == "" {
		return true
	}
	var Count int64
	_ = global.GVA_DB.Debug().Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("`Key` = ?", 绑定信息).Take(&Count)
	return Count != 0
}
func Set绑定信息(AppId, 用户Uid int, 绑定信息 string) error {

	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ? ", 用户Uid).Update("Key", 绑定信息).Error
	if err != nil {
		return err
	}
	return nil
}
func Q取绑定信息(AppId, 用户Uid int) string {
	var 绑定信息 = ""
	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ? ", 用户Uid).Select("Key").Take(&绑定信息).Error
	if err != nil {
		return ""
	}
	return 绑定信息
}
func Ser用户类型Vip时间(AppId, 用户Uid, 用户类型Id int, VipTime int64) error {

	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ? ", 用户Uid).Updates(map[string]interface{}{"UserClassId": 用户类型Id, "VipTime": VipTime}).Error
	if err != nil {
		return err
	}
	return nil
}

// Id积分增减 可能减少到0以下 ,增加无限制
func Id积分增减_批量(AppId int, Id []int, 增减值 float64, is增加 bool) error {
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
	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("VipNumber", gorm.Expr(sql, 增减值)).Error
	return err
}

// Id积分增减 减少无法减少到0以下 ,增加无限制
func Id积分增减(AppId, Id int, 增减值 float64, is增加 bool) error {
	//因为float64 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	if Id == 0 {
		return errors.New("用户不存在")
	}
	if 增减值 <= 0 {
		return errors.New("增减值不能小于等于0")
	}

	if 增减值 == 0 {
		//增减0 直接成功
		return nil
	}

	if is增加 {
		//增加直接处理就可以了,不用事务
		err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipNumber", gorm.Expr("VipNumber + ?", 增减值)).Error
		if err != nil {
			global.GVA_LOG.Error(strconv.Itoa(int(Id)) + "Id积分增加失败:" + err.Error())
			return err
		}
		return nil
	}
	//这里就是减少,需要开启事务保证
	db := global.GVA_DB.Begin() //开启事务

	err := db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipNumber", gorm.Expr("VipNumber - ?", 增减值)).Error
	if err != nil {
		db.Rollback() //出错回滚
		global.GVA_LOG.Error(strconv.Itoa(Id) + "Id积分减少失败:" + err.Error())
		return errors.New("积分减少失败查看服务器日志检查原因")
	}
	var 局_积分 float64
	var sql = fmt.Sprintf(`SELECT VipNumber FROM db_AppUser_%d WHERE Id = %d  LIMIT 1`, AppId, Id)
	db.Raw(sql).Scan(&局_积分)

	//读取新的数值
	if 局_积分 < 0 {
		// 局_积分不足,回滚并返回
		db.Rollback()
		return errors.New("积分不足")
	}

	db.Commit() //操作完成提交事务
	return nil
}

// Id点数增减 减少无法减少到0以下 ,增加无限制
func Id点数增减(AppId, Id int, 增减值 int64, is增加 bool) error {
	//因为无符号 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	if Id == 0 {
		return errors.New("用户不存在")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return nil
	}

	if is增加 {
		//增加直接处理就可以了,不用事务
		err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipTime", gorm.Expr("VipTime + ?", 增减值)).Error
		if err != nil {
			global.GVA_LOG.Error(strconv.Itoa(int(Id)) + "Id点数增加失败:" + err.Error())
			return err
		}
		return nil
	}
	//这里就是减少,需要开启事务保证
	db := global.GVA_DB.Begin() //开启事务
	var 局_点数 int64

	//db.Table("db_AppUser_"+strconv.Itoa(AppId)).Select("VipTime").Where("Id = ?", Id).First(&局_点数)   //这种方式会有警告没有模型
	db.Raw(fmt.Sprintf(`SELECT VipTime FROM db_AppUser_%d WHERE Id = %d  LIMIT 1`, AppId, Id)).Scan(&局_点数)
	//读取旧的数值

	if !Ser_AppInfo.App是否为计点(AppId) {
		// 如果不是计点方式 减去当前时间戳 为真实剩余时间
		局_点数 -= time.Now().Unix()
	}

	if 局_点数 < 增减值 {
		// 局_点数或时间不足,回滚并返回
		db.Rollback()
		return errors.New("点数不足")
	}

	err := db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipTime", gorm.Expr("VipTime - ?", 增减值)).Error
	if err != nil {
		db.Rollback() //出错回滚
		global.GVA_LOG.Error(strconv.Itoa(int(Id)) + "Id点数减少失败:" + err.Error())
		return errors.New("点数减少失败查看服务器日志检查原因")
	}
	db.Commit() //操作完成提交事务
	return nil
}

// Id点数增减 可能减少到0以下 ,增加无限制
func Id点数增减_批量(AppId int, Id []int, 增减值 int64, is增加 bool) error {
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
	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("VipTime", gorm.Expr(sql, 增减值)).Error
	return err

}
