package Ser_Admin

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"server/global"
	DB "server/structs/db"
	. "server/utils"
	"strconv"
)

func Id是否存在(Id int) bool {

	var Count int64
	result := global.GVA_DB.Model(DB.DB_Admin{}).Select("1").Where("Id=?", Id).Take(&Count)
	return result.Error == nil
}
func Id置新密码(Id int, NewPassWord string) error {
	if Id == 0 {
		return errors.New("id不能为0")
	}

	err := global.GVA_DB.Model(DB.DB_Admin{}).Where("Id = ?", Id).Updates(map[string]interface{}{"PassWord": Md5String(NewPassWord)}).Error
	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Id置新密码失败:%v,%v,%v", Id, NewPassWord, err.Error()))
		return errors.New("修改密码失败")
	}
	return nil

}
func Id取User(Id int) string {
	if Id == 0 {
		return ""
	}
	var 用户名 string
	global.GVA_DB.Model(DB.DB_Admin{}).Select("User").Where("Id=?", Id).Take(&用户名)
	return 用户名
}

func Id余额增减(Id int, 增减值 float64, is增加 bool) (新余额 float64, err error) {
	//return Id余额增减2(Id, 增减值, is增加)
	if Id == 0 {
		return 0, errors.New("管理员不存在")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return Id取余额(Id), nil
	}

	if is增加 {
		err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(DB.DB_Admin{}).Where("Id = ?", Id).Update("RMB", gorm.Expr("RMB + ?", 增减值)).Error
			if err != nil {
				global.GVA_LOG.Error(strconv.Itoa(Id) + "管理员Id余额增加失败:" + err.Error())
				return err
			}

			err = tx.Model(DB.DB_Admin{}).Select("Rmb").Where("Id=?", Id).First(&新余额).Error
			return err
		})
		return
	}

	//这里就是减少,需要开启事务保证
	db := global.GVA_DB
	tx := db.Begin() //开启事务

	// 减少余额
	sql := "UPDATE db_Admin SET RMB = RMB - ? WHERE Id = ?"
	tx.Exec(sql, 增减值, Id)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "管理员Id余额减少失败:" + tx.Error.Error())
		return 0, errors.New("余额减少失败查看服务器日志检查原因")
	}

	// 查询新余额
	sql = "SELECT RMB FROM db_Admin WHERE Id = ?"
	tx = tx.Raw(sql, Id).Scan(&新余额)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "管理员Id查询余额失败:" + tx.Error.Error())
		return 0, errors.New("查询余额失败查看服务器日志检查原因")
	}

	if 新余额 < 0 {
		// 余额不足,回滚并返回   表必须InnoDB引擎才可以,否则会真实发生扣余额,
		tx.Rollback()
		return 0, errors.New("管理员余额不足,缺少:" + utils.Float64到文本(utils.Float64取绝对值(新余额), 2))
	} else {
		tx.Commit() //操作完成提交事务
		return 新余额, nil
	}
}
func Id取余额(Id int) (余额 float64) {
	db := *global.GVA_DB
	_ = db.Model(DB.DB_Admin{}).Select("Rmb").Where("Id=?", Id).Take(&余额).Error
	return
}
func User用户名取id(用户名 string) int {
	if 用户名 == "" {
		return 0
	}

	var Id int
	db := *global.GVA_DB
	db.Model(DB.DB_Admin{}).Select("Id").Where("User=?", 用户名).Take(&Id)
	return Id
}
func Id取详情(Id int) (用户详情 DB.DB_Admin, ok bool) {
	err := global.GVA_DB.Model(DB.DB_Admin{}).Where("Id=?", Id).Take(&用户详情).Error
	return 用户详情, err == nil
}
