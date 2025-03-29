package appUser

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

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
						global.GVA_LOG.Error(strconv.Itoa(Uid) + "Uid积分唯一标识更新失败:" + err.Error())
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

		err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid = ?", Uid).Update("VipNumber", gorm.Expr("VipNumber - ?", 减少值)).Error
		if err != nil {
			global.GVA_LOG.Error(strconv.Itoa(Uid) + "Uid积分减少失败:" + err.Error())
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
		AppInfo DB.DB_AppInfo
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
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
			err = tx.Debug().Model(&DB.DB_Ka{}).Where("Id IN (?)", tx.Table(表名_AppUser).Select("Uid").Where("Id IN (?)", id)).Update("Status", Status).Error
		}
		return err
	})

	return
}
