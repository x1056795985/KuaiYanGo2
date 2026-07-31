package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	db2 "server/app/models/db"
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

// New用户信息 创建新用户(含代理关系,事务操作)
// 多表事务: User表 + Db_Agent_Level 代理关系表
func (j *user) New用户信息(c *gin.Context, User, PassWord, SuperPassWord, Qq, Email, Phone, Ip, 备注 string, UPAgentId int, AgentDiscount int, Rmb float64, RealNameAttestation string) (db2.DB_User, error) {
	var 局_User db2.DB_User
	msg := ""
	局_最短长度 := 6
	if UPAgentId != 0 {
		局_最短长度 = 2
	}
	if utf8.RuneCountInString(User) < 局_最短长度 || utf8.RuneCountInString(User) > 18 {
		return 局_User, errors.New("用户名长度必须大于" + strconv.Itoa(局_最短长度) + "小于18")
	}

	if UPAgentId != 0 {
		if !utils.Z正则_校验代理用户名(User, &msg) {
			return 局_User, errors.New("用户名" + msg)
		}
	} else {
		if !utils.Z正则_校验用户名(User, &msg) {
			return 局_User, errors.New("用户名" + msg)
		}
	}

	if !utils.Z正则_校验密码(PassWord, &msg) {
		return 局_User, errors.New("密码" + msg)
	}

	db := *global.GVA_DB

	// 检查用户名是否已存在(用户表+管理员表)
	_, err := service.NewUser(c, &db).InfoName(User)
	if err == nil {
		return 局_User, errors.New("用户名已存在")
	}
	_, err = service.NewAdmin(c, &db).InfoName(User)
	if err == nil {
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

	// 再次检查用户是否存在(并发安全)
	var count int64
	err = db.Model(db2.DB_User{}).Where("User = ?", 局_User.User).Count(&count).Error
	if err != nil {
		return 局_User, err
	}
	if count != 0 {
		return 局_User, errors.New("用户已存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		_, err = service.NewUser(c, tx).Create(&局_User)
		if err != nil {
			global.GVA_LOG.Println("New用户信息创建失败:" + err.Error())
			return errors.New("添加失败")
		}
		if 局_User.UPAgentId == 0 {
			return nil
		}
		// 有上级代理信息,添加代理关系
		err = tx.Create(&db2.Db_Agent_Level{Uid: 局_User.Id, UPAgentId: 局_User.UPAgentId, Level: 1}).Error
		if err != nil {
			return err
		}
		上级代理ID := 局_User.UPAgentId
		for i := 0; 上级代理ID > 0; i++ {
			var 上级代理的一级代理信息 db2.Db_Agent_Level
			err = tx.Where("Uid = ?", 上级代理ID).Where("Level = 1").First(&上级代理的一级代理信息).Error
			if err != nil {
				return err
			}
			上级代理ID = 上级代理的一级代理信息.UPAgentId
			err = tx.Create(&db2.Db_Agent_Level{Uid: 局_User.Id, UPAgentId: 上级代理ID, Level: i + 2}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})

	return 局_User, err
}
