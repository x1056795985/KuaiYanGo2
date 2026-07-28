package user

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/global"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"strconv"
)

var L_user user

func init() {
	L_user = user{}
}

type user struct {
}

// Id余额增减 用户余额增减(事务操作,保证余额不为负)
// is增加: true=增加 false=减少(减少时会校验余额是否足够)
// 返回: 新余额, 错误信息
func (j *user) Id余额增减(c *gin.Context, Id int, 增减值 float64, is增加 bool) (新余额 float64, err error) {
	if Id == 0 {
		return 0, errors.New("用户不存在")
	}
	if 增减值 == 0 {
		//增减0 直接成功,查询当前余额返回
		db := *global.GVA_DB
		局_User, e := service.NewUser(c, &db).Info(Id)
		if e != nil {
			return 0, e
		}
		return 局_User.Rmb, nil
	}

	db := *global.GVA_DB

	if is增加 {
		err = db.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(dbm.DB_User{}).Where("Id = ?", Id).Update("RMB", gorm.Expr("RMB + ?", 增减值)).Error
			if err != nil {
				global.GVA_LOG.Println(strconv.Itoa(Id) + "Id余额增加失败:" + err.Error())
				return err
			}
			err = tx.Model(dbm.DB_User{}).Select("Rmb").Where("Id=?", Id).First(&新余额).Error
			return err
		})
		return
	}

	//这里就是减少,需要开启事务保证
	tx := db.Begin()

	// 减少余额
	sql := "UPDATE db_User SET RMB = RMB - ? WHERE Id = ?"
	tx.Exec(sql, 增减值, Id)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Println(strconv.Itoa(Id) + "Id余额减少失败:" + tx.Error.Error())
		return 0, errors.New("余额减少失败查看服务器日志检查原因")
	}

	// 查询新余额
	sql = "SELECT RMB FROM db_User WHERE Id = ?"
	tx = tx.Raw(sql, Id).Scan(&新余额)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Println(strconv.Itoa(Id) + "Id查询余额失败:" + tx.Error.Error())
		return 0, errors.New("查询余额失败查看服务器日志检查原因")
	}

	if 新余额 < 0 {
		// 余额不足,回滚并返回
		tx.Rollback()
		return 0, errors.New("用户余额不足,缺少:" + Float64到文本(Float64取绝对值(新余额), 2))
	}

	tx.Commit()
	return 新余额, nil
}

// String 格式化余额日志文本
func (j *user) S格式化余额日志(操作描述 string, 新余额 float64) string {
	return fmt.Sprintf("%s|新余额%v", 操作描述, 新余额)
}
