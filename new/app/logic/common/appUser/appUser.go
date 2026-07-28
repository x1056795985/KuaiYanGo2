package appUser

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/new/app/global"
	"server/new/app/logic/common/log"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"strconv"
	"time"
)

// Id点数增减 应用内某用户点数增减(事务操作,减少时校验是否充足)
// AppId: 应用Id, Id: AppUser.Id, 增减值: 点数, is增加: true=增加 false=减少
// 减少时若为时间模式(非计点),会自动减去当前时间戳后判断剩余时间是否充足
func (j *appUser) Id点数增减(c *gin.Context, AppId, Id int, 增减值 int64, is增加 bool) error {
	//因为无符号 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	if Id == 0 {
		return errors.New("用户不存在")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return nil
	}
	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	if is增加 {
		//增加直接处理就可以了,不用事务
		err := db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipTime", gorm.Expr("VipTime + ?", 增减值)).Error
		if err != nil {
			global.GVA_LOG.Println(strconv.Itoa(int(Id)) + "Id点数增加失败:" + err.Error())
			return err
		}
		return nil
	}
	//这里就是减少,需要开启事务保证
	tx := db.Begin() //开启事务
	var 局_点数 int64

	tx.Raw(fmt.Sprintf(`SELECT VipTime FROM db_AppUser_%d WHERE Id = %d  LIMIT 1`, AppId, Id)).Scan(&局_点数)
	//读取旧的数值

	局_AppInfo, _ := service.NewAppInfo(c, db).Info(AppId)
	局_计点 := 局_AppInfo.AppType == 2 || 局_AppInfo.AppType == 4
	if !局_计点 {
		// 如果不是计点方式 减去当前时间戳 为真实剩余时间
		局_点数 -= time.Now().Unix()
	}

	if 局_点数 < 增减值 {
		// 局_点数或时间不足,回滚并返回
		tx.Rollback()
		if 局_计点 {
			return errors.New("点数不足")
		} else {
			return errors.New("剩余时间不足")
		}

	}

	err := tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipTime", gorm.Expr("VipTime - ?", 增减值)).Error
	if err != nil {
		tx.Rollback() //出错回滚
		global.GVA_LOG.Println(strconv.Itoa(int(Id)) + "Id点数减少失败:" + err.Error())
		return errors.New("点数减少失败查看服务器日志检查原因")
	}
	tx.Commit() //操作完成提交事务
	return nil
}

var L_appUser appUser

func init() {
	L_appUser = appUser{}

}

type appUser struct {
}

func (j *appUser) Uid积分减少(c *gin.Context, AppId, Uid int, 减少值 float64, 唯一标识 string, 唯一有效期 int64) error {
	if Uid == 0 {
		return errors.New("用户不存在")
	}
	if 减少值 <= 0 {
		return errors.New("增减值不能小于等于0")
	}
	局_唯一文本 := ""

	if 唯一标识 != "" { //如果有唯一标识,就先查一下,如果存在就返回错误
		局_唯一文本 = strconv.Itoa(Uid) + "_" + 唯一标识
		_, ok := global.H缓存.Get(局_唯一文本)
		if ok {
			return errors.New("唯一标识重复")
		}
	}

	db := global.GVA_DB
	//这里就是减少,需要开启事务保证
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		if 唯一标识 != "" {
			//先读取判断是否存在 如果不存在则插入一个
			var 局_唯一标识 dbm.DB_UniqueNumLog
			err = tx.Model(dbm.DB_UniqueNumLog{}).Table(dbm.DB_UniqueNumLog{}.TableName()+"_"+strconv.Itoa(AppId)).Clauses(clause.Locking{Strength: "UPDATE"}).Where("ItemKey = ?", 局_唯一文本).First(&局_唯一标识).Error
			if err == nil { //如果存在则判断 判断是否过期 如果没过期返回失败,如果过期则更新
				if 局_唯一标识.EndTime > time.Now().Unix() {
					err = errors.New("唯一标识重复")
				} else {
					局_唯一标识.EndTime = time.Now().Unix() + 唯一有效期
					_, err = service.NewUniqueNumLog(c, tx, AppId).Update(局_唯一标识.Id, map[string]interface{}{"EndTime": 局_唯一标识.EndTime})
					if err != nil { //如果更新失败了?? 感觉不太可能吧,
						global.GVA_LOG.Println(strconv.Itoa(Uid) + "Uid积分唯一标识更新失败:" + err.Error())
						return errors.New("唯一标识重复")
					}
				}

			} else {
				局_唯一标识 = dbm.DB_UniqueNumLog{
					ItemKey:    局_唯一文本,
					CreateTime: time.Now().Unix(),
					EndTime:    time.Now().Unix() + 唯一有效期,
				}
				_, err = service.NewUniqueNumLog(c, tx, AppId).Create(&局_唯一标识)

			}

			if err != nil { //插入失败,就是唯一标识重复了 这个是兜底
				return errors.New("唯一标识重复")
			}
		}

		err = tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ?", Uid).Update("VipNumber", gorm.Expr("VipNumber - ?", 减少值)).Error
		if err != nil {
			global.GVA_LOG.Println(strconv.Itoa(Uid) + "Uid积分减少失败:" + err.Error())
			return errors.New("积分减少失败查看服务器日志检查原因")
		}
		var 局_积分 float64
		var sql = fmt.Sprintf(`SELECT VipNumber FROM db_AppUser_%d WHERE Uid = %d  LIMIT 1`, AppId, Uid)

		if err = tx.Raw(sql).Scan(&局_积分).Error; err != nil {
			return err
		}
		//读取新的数值
		if 局_积分 < 0 {
			// 局_积分不足,回滚并返回
			return errors.New("积分不足")
		}
		return nil
	})

	if err == nil {
		//缓存唯一标识 使其短时间内无需重复查库
		if 唯一标识 != "" {
			global.H缓存.Set(局_唯一文本, 1, time.Second*time.Duration(唯一有效期))
		}
	}

	return err
}

func (j *appUser) Z置状态_同步卡号修改(c *gin.Context, AppId int, id []int, Status int) (err error) {

	var 表名_AppUser = "db_AppUser_" + strconv.Itoa(AppId)
	var info struct {
		AppInfo dbm.DB_AppInfo
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
		tx = db
	}

	info.AppInfo, err = service.NewAppInfo(c, tx).Info(AppId)
	// 卡号模式的   处理同步ka冻结
	err = tx.Transaction(func(tx2 *gorm.DB) error {
		//先修改软件用户
		err = tx2.Table(表名_AppUser).Where("Id IN ? ", id).Update("Status", Status).Error
		if err != nil {
			return err
		}
		if info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4 {
			// 子查询获取所有软件用户的Uid 在修改卡号
			err = tx.Debug().Model(&dbm.DB_Ka{}).Where("Id IN (?)", tx.Table(表名_AppUser).Select("Uid").Where("Id IN (?)", id)).Update("Status", Status).Error
		}
		return err
	})

	return
}

// Id积分增减 单个用户积分增减(事务操作,减少时校验是否充足)
// AppId: 应用Id, Id: AppUser.Id, 增减值: 积分, is增加: true=增加 false=减少
// 减少无法减少到0以下,增加无限制
func (j *appUser) Id积分增减(c *gin.Context, AppId, Id int, 增减值 float64, is增加 bool) error {
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

	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	if is增加 {
		//增加直接处理就可以了,不用事务
		err := db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipNumber", gorm.Expr("VipNumber + ?", 增减值)).Error
		if err != nil {
			global.GVA_LOG.Println(strconv.Itoa(Id) + "Id积分增加失败:" + err.Error())
			return err
		}
		return nil
	}
	//这里就是减少,需要开启事务保证
	tx := db.Begin() //开启事务

	err := tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id = ?", Id).Update("VipNumber", gorm.Expr("VipNumber - ?", 增减值)).Error
	if err != nil {
		tx.Rollback()
		global.GVA_LOG.Println(strconv.Itoa(Id) + "Id积分减少失败:" + err.Error())
		return errors.New("积分减少失败查看服务器日志检查原因")
	}
	var 局_积分 float64
	var sql = fmt.Sprintf(`SELECT VipNumber FROM db_AppUser_%d WHERE Id = %d  LIMIT 1`, AppId, Id)
	tx.Raw(sql).Scan(&局_积分)

	//读取新的数值
	if 局_积分 < 0 {
		// 局_积分不足,回滚并返回
		tx.Rollback()
		return errors.New("积分不足")
	}

	tx.Commit() //操作完成提交事务
	return nil
}

// New用户信息 创建新软件用户(单表操作但涉及业务逻辑,放logic层)
func (j *appUser) New用户信息(c *gin.Context, AppId int, Uid int, 绑定信息 string, 最大在线数量 int, VipTime int64, VipNumber float64, UserClassId int, Note string) error {
	var 局_AppUser dbm.DB_AppUser

	局_AppUser.Id = 0
	局_AppUser.Uid = Uid
	局_AppUser.Status = 1
	局_AppUser.Key = 绑定信息
	局_AppUser.VipTime = VipTime
	局_AppUser.VipNumber = VipNumber
	局_AppUser.Note = Note
	局_AppUser.MaxOnline = 最大在线数量
	局_AppUser.UserClassId = UserClassId
	局_AppUser.RegisterTime = time.Now().Unix()
	局_AppUser.AgentUid = 0 //不在这里赋值,单独处理

	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	_, err := service.NewAppUser(c, db, AppId).Create(&局_AppUser)
	return err
}

// S删除VipTime小于等于X 删除VipTime<=X的软件用户
func (j *appUser) S删除VipTime小于等于X(c *gin.Context, AppId int, VipTime int64) (影响行数 int64, err error) {
	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	影响行数 = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Where("VipTime <= ? ", VipTime).Delete("").RowsAffected
	return 影响行数, err
}

// S删除VipTime小于等于X且删除卡号 删除VipTime<=X的软件用户及其卡号(多表事务操作)
func (j *appUser) S删除VipTime小于等于X且删除卡号(c *gin.Context, AppId int, VipTime int64, Ip string) (id int64, err error) {
	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	sAppInfo := service.NewAppInfo(c, db)
	if !sAppInfo.App是否为卡号(AppId) {
		return 0, errors.New("仅限卡号类型应用使用")
	}

	var ids []int64
	err = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Select("Uid").Where("VipTime <= ? ", VipTime).Find(&ids).Error
	if err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	// 分批查询卡号名称，避免占位符超限
	var KaNames []string
	for i := 0; i < len(ids); i += 5000 {
		end := i + 5000
		if end > len(ids) {
			end = len(ids)
		}
		var batch []string
		err = db.Model(dbm.DB_Ka{}).Select("Name").Where("id IN ? ", ids[i:end]).Find(&batch).Error
		if err != nil {
			return
		}
		KaNames = append(KaNames, batch...)
	}
	id = int64(len(ids))
	err = db.Transaction(func(tx *gorm.DB) error {
		// 分批删除AppUser，避免占位符超限
		for i := 0; i < len(ids); i += 5000 {
			end := i + 5000
			if end > len(ids) {
				end = len(ids)
			}
			err = tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid IN ?", ids[i:end]).Delete("").Error
			if err != nil {
				return errors.New("删除应用用户失败:" + err.Error())
			}
		}
		// 分批删除Ka，避免占位符超限
		for i := 0; i < len(ids); i += 5000 {
			end := i + 5000
			if end > len(ids) {
				end = len(ids)
			}
			err = tx.Model(dbm.DB_Ka{}).Where("AppId = ?", AppId).Where("id IN ?", ids[i:end]).Delete("").Error
			if err != nil {
				return errors.New("删除应用卡号失败:" + err.Error())
			}
		}
		return nil
	})

	if err == nil {
		局_文本 := fmt.Sprintf("删除VipTime小于等于%d且删除卡号:{{卡号}},批次id:{{批次id}}({{卡号索引}}/%d)", VipTime, id)
		go log.L_log.Log_写卡号操作日志(c.GetString("User"), Ip, 局_文本, KaNames, 4, 4)
	}

	return
}

// S删除卡号不存在的软件用户 删除卡号不存在的软件用户(多表操作)
func (j *appUser) S删除卡号不存在的软件用户(c *gin.Context, AppId int) (id int64, err error) {
	sAppInfo := service.NewAppInfo(c, global.GVA_DB)
	if !sAppInfo.App是否为卡号(AppId) {
		return 0, errors.New("仅限卡号类型应用使用")
	}

	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	var ids []int
	//获取全部uid 就是卡号id
	err = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId)).Select("Uid").Find(&ids).Error
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	var KaId []int
	err = db.Model(dbm.DB_Ka{}).Select("Id").Where("AppId = ? ", AppId).Scan(&KaId).Error
	if err != nil {
		return 0, err
	}

	Uids := S数组_整数取差集(KaId, ids)
	if len(Uids) == 0 {
		return 0, nil
	}
	// 分批删除，避免占位符超限
	var total int64
	for i := 0; i < len(Uids); i += 5000 {
		end := i + 5000
		if end > len(Uids) {
			end = len(Uids)
		}
		tx := db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid IN ? ", Uids[i:end]).Delete("")
		if tx.Error != nil {
			return total, tx.Error
		}
		total += tx.RowsAffected
	}
	return total, nil
}

// P批量_全部用户增减时间或点数 批量增减时间或点数(复杂查询)
func (j *appUser) P批量_全部用户增减时间或点数(c *gin.Context, AppId int, Number int64, 账号状态 int, 用户或卡号前缀 string, 注册时间开始, 注册时间结束 int, UserClassId []int) (影响行数 int64, err error) {
	sAppInfo := service.NewAppInfo(c, global.GVA_DB)
	if AppId < 10000 || !sAppInfo.AppId是否存在(AppId) {
		return 0, errors.New("AppId不存在")
	}

	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId) + " ai").Select("ai.Id")

	局_is计点 := sAppInfo.App是否为计点(AppId)
	局_is卡号 := sAppInfo.App是否为卡号(AppId)
	if 用户或卡号前缀 != "" {
		if 局_is卡号 {
			db = db.Joins("LEFT JOIN db_Ka ka ON ai.Uid = ka.Id").Where("ka.AppId = ?", AppId).Where("ka.Name like ?", 用户或卡号前缀+"%")
		} else {
			db = db.Joins("LEFT JOIN db_User ON ai.Uid = db_User.Id").Model(dbm.DB_User{}).Where("User like ?", 用户或卡号前缀+"%")
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
			db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", 局_id数组).Where("VipTime < ?", time.Now().Unix()).Update("VipTime", time.Now().Unix())
		}
		影响行数 = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", 局_id数组).Update("VipTime", gorm.Expr("VipTime + ?", Number)).RowsAffected
		var 局_id数组文本 string
		for _, num := range 局_id数组 {
			局_id数组文本 += strconv.Itoa(num) + ","
		}
		局_id数组文本 = fmt.Sprintf("管理员进行了批量维护时间点数,AppId:%d,软件用户ID[%s],操作类型增减指定值,修改值:%d", AppId, 局_id数组文本, Number)
		global.GVA_LOG.Println(局_id数组文本)
	}

	return 影响行数, err
}

// P批量_全部用户修改为指定时间或点数 批量修改为指定时间或点数(复杂查询)
func (j *appUser) P批量_全部用户修改为指定时间或点数(c *gin.Context, AppId int, Number int64, 账号状态 int, 用户或卡号前缀 string, 注册时间开始, 注册时间结束 int) (影响行数 int64, err error) {
	sAppInfo := service.NewAppInfo(c, global.GVA_DB)
	if AppId < 10000 || !sAppInfo.AppId是否存在(AppId) {
		return 0, errors.New("AppId不存在")
	}

	db := global.GVA_DB.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId))
	db = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(AppId) + " ai").Select("ai.Id")

	局_is计点 := sAppInfo.App是否为计点(AppId)
	局_is卡号 := sAppInfo.App是否为卡号(AppId)
	if 用户或卡号前缀 != "" {
		if 局_is卡号 {
			db = db.Joins("LEFT JOIN db_Ka ka ON ai.Uid = ka.Id").Where("ka.AppId = ?", AppId).Where("ka.Name like ?", 用户或卡号前缀+"%")
		} else {
			db = db.Joins("LEFT JOIN db_User ON ai.Uid = db_User.Id").Model(dbm.DB_User{}).Where("User like ?", 用户或卡号前缀+"%")
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
		影响行数 = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Id IN ?", 局_id数组).Update("VipTime", Number).RowsAffected
		var 局_id数组文本 string
		for _, num := range 局_id数组 {
			局_id数组文本 += strconv.Itoa(num) + ","
		}
		局_id数组文本 = fmt.Sprintf("管理员进行了批量维护时间点数,AppId:%d,软件用户ID[%s],操作类型修改指定值,修改值:%d", AppId, 局_id数组文本, Number)
		global.GVA_LOG.Println(局_id数组文本)
	}

	return 影响行数, err
}
