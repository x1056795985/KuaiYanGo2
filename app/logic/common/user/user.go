package user

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/logic/common/log"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	"server/app/utils"
	"strconv"
	"time"
	"unicode/utf8"
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

// Id余额增减_批量 批量增减用户余额(事务操作)
// ids: 用户Id数组, 增减值: 金额, is增加: true=增加 false=减少
// 减少时不校验余额是否充足(允许负数), 适合批量扣费场景
func (j *user) Id余额增减_批量(c *gin.Context, Ids []int, 增减值 float64, is增加 bool) (err error) {
	if len(Ids) == 0 {
		return errors.New("用户id数组不能为空")
	}
	if 增减值 == 0 {
		return nil
	}

	db := *global.GVA_DB
	sql := "RMB + ?"
	if !is增加 {
		sql = "RMB - ?"
	}

	//分批处理,避免IN子句过长
	for i := 0; i < len(Ids); i += 5000 {
		end := i + 5000
		if end > len(Ids) {
			end = len(Ids)
		}
		err = db.Model(dbm.DB_User{}).Where("Id IN ?", Ids[i:end]).Update("RMB", gorm.Expr(sql, 增减值)).Error
		if err != nil {
			global.GVA_LOG.Println(fmt.Sprintf("Id余额增减_批量失败:%v", err.Error()))
			return
		}
	}
	return
}

// Id余额转账 用户余额转账(事务操作,保证余额不为负)
// 从Id从扣款,转入ToId,转账金额必须大于0
func (j *user) Id余额转账(c *gin.Context, FromId, ToId int, 转账金额 float64) (err error) {
	if FromId == 0 || ToId == 0 {
		return errors.New("用户不存在")
	}
	if FromId == ToId {
		return errors.New("不能给自己转账")
	}
	if 转账金额 <= 0 {
		return errors.New("转账金额必须大于0")
	}

	db := *global.GVA_DB
	err = db.Transaction(func(tx *gorm.DB) error {
		//扣款
		err := tx.Model(dbm.DB_User{}).Where("Id = ?", FromId).Update("RMB", gorm.Expr("RMB - ?", 转账金额)).Error
		if err != nil {
			return err
		}
		//查扣款后余额
		var 局_新余额 float64
		err = tx.Model(dbm.DB_User{}).Select("Rmb").Where("Id=?", FromId).First(&局_新余额).Error
		if err != nil {
			return err
		}
		if 局_新余额 < 0 {
			return errors.New("余额不足,缺少:" + Float64到文本(Float64取绝对值(局_新余额), 2))
		}
		//加款
		err = tx.Model(dbm.DB_User{}).Where("Id = ?", ToId).Update("RMB", gorm.Expr("RMB + ?", 转账金额)).Error
		return err
	})
	return
}

// New用户信息 创建新用户(多表事务操作: DB_User + Db_Agent_Level)
func (j *user) New用户信息(c *gin.Context, User, PassWord, SuperPassWord, Qq, Email, Phone, Ip, 备注 string, UPAgentId int, AgentDiscount int, Rmb float64, RealNameAttestation string) (dbm.DB_User, error) {
	var 局_User dbm.DB_User
	msg := ""
	局_最短长度 := 6
	if UPAgentId != 0 {
		局_最短长度 = 2
	}
	if utf8.RuneCountInString(User) < 局_最短长度 || utf8.RuneCountInString(User) > 18 {
		return 局_User, errors.New("用户名长度必须大于" + strconv.Itoa(局_最短长度) + "小于18")
	}

	if UPAgentId != 0 {
		if !Z正则_校验代理用户名(User, &msg) {
			return 局_User, errors.New("用户名" + msg)
		}
	} else {
		if !Z正则_校验用户名(User, &msg) {
			return 局_User, errors.New("用户名" + msg)
		}
	}

	if !Z正则_校验密码(PassWord, &msg) {
		return 局_User, errors.New("密码" + msg)
	}

	db := *global.GVA_DB
	var 局_管理结果 dbm.DB_Admin
	局_管理结果, _ = service.NewAdmin(c, &db).Info2(map[string]interface{}{"User": User})

	sUser := service.NewUser(c, &db)
	if sUser.User用户名取id(User) != 0 || 局_管理结果.Id != 0 {
		return 局_User, errors.New("用户名已存在")
	}

	局_User.Id = 0
	局_User.User = User
	局_User.Qq = Qq
	局_User.Email = Email
	局_User.Phone = Phone
	局_User.PassWord = utils.BcryptHash(PassWord)
	局_User.SuperPassWord = utils.BcryptHash(SuperPassWord)
	局_User.Status = 1
	局_User.RegisterIp = Ip
	局_User.RegisterTime = time.Now().Unix()
	局_User.UPAgentId = UPAgentId
	局_User.RealNameAttestation = RealNameAttestation
	局_User.AgentDiscount = AgentDiscount
	局_User.LoginTime = 0
	局_User.LoginAppid = 0
	局_User.LoginIp = ""
	局_User.Note = 备注
	局_User.Rmb = Rmb

	var count int64
	err := db.Model(dbm.DB_User{}).Where("User = ?", 局_User.User).Count(&count).Error
	if count != 0 {
		return 局_User, errors.New("用户已存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(dbm.DB_User{}).Create(&局_User).Error
		if err != nil {
			go log.L_log.Log_写用户消息(log.Log用户消息类型_系统执行错误, constant.APPID_管理平台, "系统", "系统", global.X系统信息.B版本号当前, "New用户信息非预计错误:"+err.Error(), Ip)
			return errors.New("添加失败")
		}
		if 局_User.UPAgentId == 0 {
			return nil
		}
		//有上级代理信息,添加代理关系
		err = tx.Create(&dbm.Db_Agent_Level{Uid: 局_User.Id, UPAgentId: 局_User.UPAgentId, Level: 1}).Error
		if err != nil {
			return err
		}
		上级代理ID := 局_User.UPAgentId
		for i := 0; 上级代理ID > 0; i++ {
			var 上级代理的一级代理信息 dbm.Db_Agent_Level
			err = tx.Where("Uid = ?", 上级代理ID).Where("Level = 1").First(&上级代理的一级代理信息).Error
			if err != nil {
				return err
			}
			上级代理ID = 上级代理的一级代理信息.UPAgentId
			err = tx.Create(&dbm.Db_Agent_Level{Uid: 局_User.Id, UPAgentId: 上级代理ID, Level: i + 2}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})

	return 局_User, err
}
