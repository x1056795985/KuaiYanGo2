package Ser_AppUser

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
	"strconv"
	"time"
)

func User或卡号取Id(AppId int, user string) int {
	var id int

	if Ser_AppInfo.App是否为卡号(AppId) {
		// 执行合并后的SQL语句
		global.GVA_DB.Raw("SELECT `Id` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Uid` = (SELECT `Id` FROM `db_Ka` WHERE `Name` = ?) LIMIT 1", user).Scan(&id)

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
func Id取Uid_批量(AppId int, id []int) []int {
	var Uid []int
	global.GVA_DB.Raw("SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Id` IN  ? ", id).Scan(&Uid)
	return Uid
}
func Id取User(AppId int, id int) string {
	var 用户名 string
	if Ser_AppInfo.App是否为卡号(AppId) {
		global.GVA_DB.Raw("SELECT `Name` FROM `db_Ka` WHERE Id = (SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE Id = ?  LIMIT 1) LIMIT 1", id).Scan(&用户名)

	} else {
		// 执行合并后的SQL语句
		global.GVA_DB.Raw("SELECT `User` FROM `db_User` WHERE Id = (SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE Id = ?  LIMIT 1) LIMIT 1", id).Scan(&用户名)
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

func Uid取User(AppId int, Uid int) string {
	var 用户名 string
	if Ser_AppInfo.App是否为卡号(AppId) {
		/*		var 卡号 string
				return 卡号
				用户名 = Ser_Ka.Id取卡号(Uid)   //这个有循环导入报错,待解决
		*/
		_ = global.GVA_DB.Model(DB.DB_Ka{}).Select("Name").Where("Id=?", Uid).First(&用户名)

	} else {
		用户名 = Ser_User.Id取User(Uid)
	}
	return 用户名
}

func Uid取备注(AppId int, Uid int) string {
	var 备注 string
	if AppId < 10000 { //屏蔽掉管理平台代理平台等
		return ""
	}
	_ = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Note").Where("Uid=?", Uid).First(&备注).Error

	return 备注
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
func New用户信息(AppId int, Uid int, 绑定信息 string, 最大在线数量 int, VipTime int64, VipNumber float64, UserClassId int, Note string, AgentUid int) error {
	var 局_AppUser DB.DB_AppUser

	局_AppUser.Id = 0
	局_AppUser.Uid = Uid
	局_AppUser.Status = 1
	局_AppUser.Key = 绑定信息
	局_AppUser.VipTime = VipTime
	局_AppUser.VipNumber = VipNumber
	局_AppUser.Note = Note
	局_AppUser.MaxOnline = 最大在线数量
	局_AppUser.UserClassId = UserClassId
	局_AppUser.RegisterTime = int(time.Now().Unix())
	局_AppUser.AgentUid = AgentUid

	err := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Create(&局_AppUser).Error
	return err
}

func B绑定信息是否存在(AppId int, 绑定信息 string) bool {
	if 绑定信息 == "" {
		return true
	}
	var Count int64
	_ = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("1").Where("`Key` = ?", 绑定信息).Take(&Count)
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
	db := *global.GVA_DB
	db = *db.Debug()
	err := db.Model(DB.DB_AppUser{}).
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

	局_计点 := Ser_AppInfo.App是否为计点(AppId)
	if !局_计点 {
		// 如果不是计点方式 减去当前时间戳 为真实剩余时间
		局_点数 -= time.Now().Unix()
	}

	if 局_点数 < 增减值 {
		// 局_点数或时间不足,回滚并返回
		db.Rollback()
		if 局_计点 {
			return errors.New("点数不足")
		} else {
			return errors.New("剩余时间不足")
		}

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
func S删除VipTime小于等于X(AppId int, VipTime int64) (影响行数 int64, err error) {
	db := global.GVA_DB.Model(DB.DB_AppUser{})
	影响行数 = db.Table("db_AppUser_"+strconv.Itoa(AppId)).Where("VipTime <= ? ", VipTime).Delete("").RowsAffected
	return 影响行数, err
}
func S删除VipTime小于等于X且删除卡号(c *gin.Context, AppId int, VipTime int64, Ip string) (id int64, err error) {
	if !Ser_AppInfo.App是否为卡号(AppId) {
		return 0, errors.New("仅限卡号类型应用使用")
	}

	db := global.GVA_DB.Model(DB.DB_AppUser{})
	var ids []int64

	err = db.Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Uid").Where("VipTime <= ? ", VipTime).Find(&ids).Error

	if err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	var KaNames []string
	err = db.Model(DB.DB_Ka{}).Select("Name").Where("Uid IN ? ", ids).Find(&KaNames).Error
	id = int64(len(ids))
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid IN ?", ids).Delete("").Error
		if err != nil {
			return errors.New("删除应用用户失败:" + err.Error())
		}

		err = tx.Model(DB.DB_Ka{}).Where("AppId = ?", AppId).Where("id IN ?", ids).Delete("").Error
		if err != nil {
			return errors.New("删除应用卡号失败:" + err.Error())
		}
		return nil
	})

	if err == nil {
		局_文本 := fmt.Sprintf("删除VipTime小于等于%d且删除卡号:{{卡号}},批次id:{{批次id}}({{卡号索引}}/%d)", VipTime, id)
		go Ser_Log.Log_写卡号操作日志(c.GetString("User"), Ip, 局_文本, KaNames, 4, 4)
	}

	return
}
func S删除卡号不存在的软件用户(c *gin.Context, AppId int) (id int64, err error) {
	if !Ser_AppInfo.App是否为卡号(AppId) {
		return 0, errors.New("仅限卡号类型应用使用")
	}

	db := *global.GVA_DB.Model(DB.DB_AppUser{})
	var ids []int
	//获取全部uid 就是卡号id
	err = db.Table("db_AppUser_" + strconv.Itoa(AppId)).Select("Uid").Find(&ids).Error

	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	var KaId []int
	db = *global.GVA_DB.Model(DB.DB_Ka{})
	err = db.Select("Id").Where("AppId = ? ", AppId).Scan(&KaId).Error
	if err != nil {
		return 0, err
	}

	Uids := utils.S数组_整数取差集(KaId, ids)
	if len(Uids) == 0 {
		return 0, nil
	}
	db = *global.GVA_DB.Model(DB.DB_AppUser{})
	db2 := db.Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid IN ? ", Uids).Delete("")
	return db2.RowsAffected, db2.Error
}

func Z置状态_同步卡号修改(AppId int, id []int, Status int) error {
	var 表名_AppUser = "db_AppUser_" + strconv.Itoa(AppId)
	if !Ser_AppInfo.App是否为卡号(AppId) {
		return global.GVA_DB.Table(表名_AppUser).Where("Id IN ? ", id).Update("Status", Status).Error
	}
	// 卡号模式的   处理同步ka冻结
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		//先修改软件用户
		err := tx.Table(表名_AppUser).Where("Id IN ? ", id).Update("Status", Status).Error
		if err != nil {
			return err
		}
		// 子查询获取所有软件用户的Uid 在修改卡号
		err = tx.Debug().Model(&DB.DB_Ka{}).Where("Id IN (?)", tx.Table(表名_AppUser).Select("Uid").Where("Id IN (?)", id)).Update("Status", Status).Error

		//err = tx.Debug().Model(DB.DB_Ka{}).Where("Id IN ? ", tx.Exec("SELECT Uid  FROM ?  WHERE Id IN ?", 表名_AppUser, id)).Update("Status", Status).Error
		return err
	})
}
func P批量_全部用户增减时间或点数(AppId int, Number int64, 账号状态 int, 用户或卡号前缀 string, 注册时间开始, 注册时间结束 int, UserClassId []int) (影响行数 int64, err error) {

	if AppId < 10000 || !Ser_AppInfo.AppId是否存在(AppId) {
		return 0, errors.New("AppId不存在")
	}

	db := global.GVA_DB.Debug().Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId) + " ai").Select("ai.Id")

	局_is计点 := Ser_AppInfo.App是否为计点(AppId)
	局_is卡号 := Ser_AppInfo.App是否为卡号(AppId)
	if 用户或卡号前缀 != "" {
		if 局_is卡号 {
			db = db.Joins("LEFT JOIN db_Ka ka ON ai.Uid = ka.Id").Where("ka.AppId = ?", AppId).Where("ka.Name like ?", 用户或卡号前缀+"%")
		} else {
			db = db.Joins("LEFT JOIN db_User ON ai.Uid = db_User.Id").Model(DB.DB_User{}).Where("User like ?", 用户或卡号前缀+"%")
		}
	}

	switch 账号状态 {
	default:
		return 0, errors.New("账号状态错误")
	case 1: //全部

	case 2: //已过期 点数为0
		if 局_is计点 {
			db = db.Where("ai.VipTime = 0 ")
		} else {
			db = db.Where("ai.VipTime < ? ", time.Now().Unix())
		}

	case 3: //未过期
		if 局_is计点 {
			db = db.Where("ai.VipTime >0 ")
		} else {
			db = db.Where("ai.VipTime > ? ", time.Now().Unix())
		}
	}
	if 注册时间开始 > 0 {
		db = db.Where("ai.RegisterTime > ?", 注册时间开始)
	}
	if 注册时间结束 > 0 {
		db = db.Where("ai.RegisterTime < ?", 注册时间结束)
	}
	if len(UserClassId) > 0 {
		db = db.Where("ai.UserClassId IN ?", UserClassId)
	}

	var 局_id数组 []int
	db.Find(&局_id数组)
	if len(局_id数组) > 0 {
		//如果是增加时间 Number 先给过期的修改为当前时间戳
		if Number > 0 {
			global.GVA_DB.Debug().Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", 局_id数组).Where("VipTime < ?", time.Now().Unix()).Update("VipTime", time.Now().Unix())
		}
		影响行数 = global.GVA_DB.Debug().Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", 局_id数组).Update("VipTime", gorm.Expr("VipTime + ?", Number)).RowsAffected
		var 局_id数组文本 string
		for _, num := range 局_id数组 {
			局_id数组文本 += strconv.Itoa(num) + ","
		}
		局_id数组文本 = fmt.Sprintf("管理员进行了批量维护时间点数,AppId:%d,软件用户ID[%s],操作类型增减指定值,修改值:%d", AppId, 局_id数组文本, Number)
		global.GVA_LOG.Log(1, 局_id数组文本)
	}

	return 影响行数, err
}

func P批量_全部用户修改为指定时间或点数(AppId int, Number int64, 账号状态 int, 用户或卡号前缀 string, 注册时间开始, 注册时间结束 int) (影响行数 int64, err error) {

	if AppId < 10000 || !Ser_AppInfo.AppId是否存在(AppId) {
		return 0, errors.New("AppId不存在")
	}

	db := global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId) + " ai").Select("ai.Id")

	局_is计点 := Ser_AppInfo.App是否为计点(AppId)
	局_is卡号 := Ser_AppInfo.App是否为卡号(AppId)
	if 用户或卡号前缀 != "" {
		if 局_is卡号 {
			db = db.Joins("LEFT JOIN db_Ka ka ON ai.Uid = ka.Id").Where("ka.AppId = ?", AppId).Where("ka.Name like ?", 用户或卡号前缀+"%")
		} else {
			db = db.Joins("LEFT JOIN db_User ON ai.Uid = db_User.Id").Model(DB.DB_User{}).Where("User like ?", 用户或卡号前缀+"%")
		}
	}

	switch 账号状态 {
	default:
		return 0, errors.New("账号状态错误")
	case 1: //全部

	case 2: //已过期 点数为0
		if 局_is计点 {
			db = db.Where("ai.VipTime = 0 ")
		} else {
			db = db.Where("ai.VipTime < ? ", time.Now().Unix())
		}

	case 3: //未过期
		if 局_is计点 {
			db = db.Where("ai.VipTime >0 ")
		} else {
			db = db.Where("ai.VipTime > ? ", time.Now().Unix())
		}
	}
	if 注册时间开始 > 0 {
		db = db.Where("ai.RegisterTime > ?", 注册时间开始)
	}
	if 注册时间结束 > 0 {
		db = db.Where("ai.RegisterTime < ?", 注册时间结束)
	}

	var 局_id数组 []int
	db.Find(&局_id数组)
	if len(局_id数组) > 0 {
		影响行数 = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", 局_id数组).Update("VipTime", Number).RowsAffected
		var 局_id数组文本 string
		for _, num := range 局_id数组 {
			局_id数组文本 += strconv.Itoa(num) + ","
		}
		局_id数组文本 = fmt.Sprintf("管理员进行了批量维护时间点数,AppId:%d,软件用户ID[%s],操作类型修改指定值,修改值:%d", AppId, 局_id数组文本, Number)
		global.GVA_LOG.Log(1, 局_id数组文本)
	}

	return 影响行数, err
}

// Id点数增减 可能减少到0以下 ,增加无限制
func X修改用户类型_批量(AppId int, Id []int, UserClassId int) (int64, error) {
	//因为无符号 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	db := *global.GVA_DB
	db2 := db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", Id).Update("UserClassId", UserClassId)
	return db2.RowsAffected, db2.Error
}
