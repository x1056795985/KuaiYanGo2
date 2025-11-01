package Ser_User

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/Service/Ser_Admin"
	"server/Service/Ser_Log"
	"server/global"

	DB "server/structs/db"
	. "server/utils"
	"strconv"
	"time"
	"unicode/utf8"
)

func UserId是否存在(id int) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.DB_User{}).Select("1").Where("Id=?", id).Take(&Count)
	return result.Error == nil
}

func User用户名取id(用户名 string) int {
	if 用户名 == "" {
		return 0
	}

	var Id = 0
	_ = global.GVA_DB.Model(DB.DB_User{}).Select("Id").Where("User=?", 用户名).Take(&Id)
	return Id
}

// 负数会取管理员表的信息
func Id取User(Id int) string {
	if Id == 0 {
		return ""
	}
	var 用户名 string
	if Id < 0 {
		global.GVA_DB.Model(DB.DB_Admin{}).Select("User").Where("Id=?", -Id).Scan(&用户名)
		return 用户名
	}

	err := global.GVA_DB.Debug().Model(DB.DB_User{}).Select("User").Where("Id=?", Id).Scan(&用户名).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	return 用户名
}

// 取用户表的信息_批量,仅限用户表
func Id取User_批量(Id []int) map[int]string {
	if len(Id) == 0 {
		return map[int]string{}
	}
	var 用户名 []DB.DB_User
	global.GVA_DB.Model(DB.DB_User{}).Select("Id,User").Where("Id IN ?", Id).Find(&用户名)
	var 局_返回 = make(map[int]string, len(用户名))

	for 索引, _ := range 用户名 {
		局_返回[用户名[索引].Id] = 用户名[索引].User
	}

	return 局_返回
}

// 负数会取管理员表的信息
func Id取状态(Id int) int {
	if Id == 0 {
		return 1
	}
	var Status int
	if Id < 0 {
		global.GVA_DB.Model(DB.DB_Admin{}).Select("Status").Where("Id=?", -Id).First(&Status)
		return Status
	}

	global.GVA_DB.Model(DB.DB_User{}).Select("Status").Where("Id=?", Id).First(&Status)
	return Status
}

func User余额增减(用户名 string, 增减值 float64, is增加 bool) (float64, error) {
	return Id余额增减(User用户名取id(用户名), 增减值, is增加)
}

// Id余额增减 可能减少到0以下 ,增加无限制
func Id余额增减_批量(Id []int, 增减值 float64, is增加 bool) error {
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

	sql := "RMB + ?"
	if !is增加 {
		sql = "RMB - ?"
	}
	err := global.GVA_DB.Model(DB.DB_User{}).Where("Id IN ?", Id).Update("RMB", gorm.Expr(sql, 增减值)).Error
	return err
}

func Id余额增减(Id int, 增减值 float64, is增加 bool) (新余额 float64, err error) {
	//return Id余额增减2(Id, 增减值, is增加)
	if Id == 0 {
		return 0, errors.New("用户不存在")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return Id取余额(Id), nil
	}

	if is增加 {
		err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
			err = tx.Model(DB.DB_User{}).Where("Id = ?", Id).Update("RMB", gorm.Expr("RMB + ?", 增减值)).Error
			if err != nil {
				global.GVA_LOG.Error(strconv.Itoa(Id) + "Id余额增加失败:" + err.Error())
				return err
			}
			err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id=?", Id).First(&新余额).Error
			return err
		})
		return
	}

	//这里就是减少,需要开启事务保证
	db := global.GVA_DB
	tx := db.Begin() //开启事务

	// 减少余额
	sql := "UPDATE db_User SET RMB = RMB - ? WHERE Id = ?"
	tx.Exec(sql, 增减值, Id)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "Id余额减少失败:" + tx.Error.Error())
		return 0, errors.New("余额减少失败查看服务器日志检查原因")
	}

	// 查询新余额
	sql = "SELECT RMB FROM db_User WHERE Id = ?"
	tx = tx.Raw(sql, Id).Scan(&新余额)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "Id查询余额失败:" + tx.Error.Error())
		return 0, errors.New("查询余额失败查看服务器日志检查原因")
	}

	if 新余额 < 0 {
		// 余额不足,回滚并返回   表必须InnoDB引擎才可以,否则会真实发生扣余额,
		tx.Rollback()
		return 0, errors.New("用户余额不足,缺少:" + utils.Float64到文本(utils.Float64取绝对值(新余额), 2))
	} else {
		tx.Commit() //操作完成提交事务
		return 新余额, nil
	}

}
func Id余额转账(Id, 目标id int, 增减值 float64, ip string) (源新余额, 目标新余额 float64, err error) {
	//return Id余额增减2(Id, 增减值, is增加)
	if Id == 0 || 目标id == 0 {
		return 源新余额, 目标新余额, errors.New("用户不存在")
	}
	if 增减值 <= 0 {
		return 源新余额, 目标新余额, errors.New("金额必须大于0")
	}
	var 源用户详情 DB.DB_User

	var ok bool
	//首次查询,无锁先判断一次
	if 源用户详情, ok = Id取详情(Id); !ok {
		return 源新余额, 目标新余额, errors.New("源用户不存在")
	}

	if 源用户详情.Rmb < 增减值 {
		return 源新余额, 目标新余额, errors.New("余额不足")
	}
	var 目标用户详情 DB.DB_User
	if 目标用户详情, ok = Id取详情(目标id); !ok {
		return 源新余额, 目标新余额, errors.New("目标用户不存在")
	}

	//这里就是转账了,需要开启事务保证
	db := global.GVA_DB
	tx := db.Begin() //开启事务

	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id=?", Id).First(&源用户详情).Error //加锁再查一次

	if err != nil || 源用户详情.Rmb < 增减值 {
		// 余额不足,回滚并返回   表必须InnoDB引擎才可以,否则会真实发生扣余额,
		tx.Rollback()
		return 源新余额, 目标新余额, errors.New("余额不足")
	}
	// 减少余额
	sql := "UPDATE db_User SET RMB = RMB - ? WHERE Id = ?"
	tx.Exec(sql, 增减值, Id)

	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "Id余额减少失败:" + tx.Error.Error())
		return 源新余额, 目标新余额, errors.New("余额减少失败查看服务器日志检查原因")
	}

	// 查询新余额
	sql = "SELECT RMB FROM db_User WHERE Id = ?"
	tx = tx.Raw(sql, Id).Scan(&源新余额)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "Id查询余额失败:" + tx.Error.Error())
		return 源新余额, 目标新余额, errors.New("查询余额失败查看服务器日志检查原因")
	}

	if 源新余额 < 0 {
		// 余额不足,回滚并返回   表必须InnoDB引擎才可以,否则会真实发生扣余额,
		tx.Rollback()
		return 源新余额, 目标新余额, errors.New("用户余额不足,缺少:" + utils.Float64到文本(utils.Float64取绝对值(源新余额), 2))
	}
	//目标账号
	// 增加余额
	sql = "UPDATE db_User SET RMB = RMB + ? WHERE Id = ?"
	tx.Exec(sql, 增减值, 目标id)

	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "目标Id余额增加失败:" + tx.Error.Error())
		return 源新余额, 目标新余额, errors.New("余额增加失败查看服务器日志检查原因")
	}
	// 查询新余额
	sql = "SELECT RMB FROM db_User WHERE Id = ?"
	tx = tx.Raw(sql, 目标id).Scan(&目标新余额)
	if tx.Error != nil {
		tx.Rollback()
		global.GVA_LOG.Error(strconv.Itoa(Id) + "目标id查询余额失败:" + tx.Error.Error())
		return 源新余额, 目标新余额, errors.New("目标id查询余额失败查看服务器日志检查原因")
	}
	tx.Commit() //操作完成提交事务

	Ser_Log.Log_写余额日志(源用户详情.User, ip, "转账给:"+目标用户详情.User+"|新余额≈"+utils.Float64到文本(源新余额, 2), 增减值)
	// 源用户详情.User   20240422 因为可能泄漏上级代理用户名,所以不在直接显示
	Ser_Log.Log_写余额日志(目标用户详情.User, ip, "收到来自上级的转账.|新余额≈"+utils.Float64到文本(目标新余额, 2), 增减值)

	return 源新余额, 目标新余额, nil

}
func User取详情(User string) (用户详情 DB.DB_User, ok bool) {
	err := global.GVA_DB.Model(DB.DB_User{}).Where("User=?", User).First(&用户详情).Error
	return 用户详情, err == nil
}

func Id取详情(Id int) (用户详情 DB.DB_User, ok bool) {
	err := global.GVA_DB.Model(DB.DB_User{}).Where("Id=?", Id).First(&用户详情).Error
	return 用户详情, err == nil
}

func Id取详情_数组(Id []int) ([]DB.DB_User, error) {
	var 局_用户详情 = make([]DB.DB_User, 0, len(Id))
	if len(Id) == 0 {
		return 局_用户详情, nil
	}
	err := global.GVA_DB.Model(DB.DB_User{}).Where("Id IN ?", Id).Find(&局_用户详情).Error
	return 局_用户详情, err
}
func Id取余额(Id int) (余额 float64) {
	_ = global.GVA_DB.Model(DB.DB_User{}).Select("Rmb").Where("Id=?", Id).First(&余额).Error
	return
}
func Id置最后登录AppId(Id, AppId int, Ip string) {
	if Id == 0 {
		return
	}
	err := global.GVA_DB.Model(DB.DB_User{}).Where("Id = ?", Id).Updates(map[string]interface{}{"LoginAppid": AppId, "LoginIp": Ip, "LoginTime": time.Now().Unix()}).Error

	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Id置最后登录AppId失败ID:%v,AppId,%v,Ip,%v,%v", Id, AppId, Ip, err.Error()))
	}
	return

}
func Id置QQ邮箱手机号(Id int, QQ, 邮箱, 手机号 string) error {
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

	err := global.GVA_DB.Model(DB.DB_User{}).Where("Id = ?", Id).Updates(&局data).Error

	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Id置QQ邮箱手机号失败ID:%v,%v,%v,%v,%v", Id, QQ, 邮箱, 手机号, err.Error()))
		return err
	}
	return nil

}

// New用户信息
func New用户信息(User, PassWord, SuperPassWord, Qq, Email, Phone, Ip, 备注 string, UPAgentId int, AgentDiscount int, Rmb float64, RealNameAttestation string) (DB.DB_User, error) {
	var 局_User DB.DB_User
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
	//不用校验 任意填写
	/*	if Email != "" && !utils.Z正则_校验email(Email, &msg) {
			return errors.New("email邮箱格式不正确")
		}
	*/
	/*	if SuperPassWord == PassWord {
		return errors.New("超级密码不能和密码相同")
	}*/

	/*	if !utils.Z正则_校验密码(SuperPassWord, &msg) {
		return errors.New("超级密码" + msg)
	}*/
	if User用户名取id(User) != 0 || Ser_Admin.User用户名取id(User) != 0 {
		return 局_User, errors.New("用户名已存在")
	}

	局_User.Id = 0
	局_User.User = User
	局_User.Qq = Qq
	局_User.Email = Email
	局_User.Phone = Phone
	局_User.PassWord = BcryptHash(PassWord)
	局_User.SuperPassWord = BcryptHash(SuperPassWord)
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
	err := global.GVA_DB.Model(DB.DB_User{}).Where("User = ?", 局_User.User).Count(&count).Error
	// 没查到数据
	if count != 0 {
		return 局_User, errors.New("用户已存在")
	}

	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(DB.DB_User{}).Create(&局_User).Error
		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "New用户信息非预计错误:"+err.Error(), Ip)
			return errors.New("添加失败")
		}
		if 局_User.UPAgentId == 0 {
			return nil
		}
		//有上级代理信息,添加代理关系
		err = tx.Create(&DB.Db_Agent_Level{Uid: 局_User.Id, UPAgentId: 局_User.UPAgentId, Level: 1}).Error
		if err != nil {
			return err
		}
		上级代理ID := 局_User.UPAgentId //是上级的代理信息
		for i := 0; 上级代理ID > 0; i++ {
			var 上级代理的一级代理信息 DB.Db_Agent_Level
			//查询代理上一级代理的信息
			err = tx.Where("Uid = ?", 上级代理ID).Where("Level = 1").First(&上级代理的一级代理信息).Error
			if err != nil {
				return err
			}
			//几级代理循环几次 查询上级代理ID的一级代理
			上级代理ID = 上级代理的一级代理信息.UPAgentId
			err = tx.Create(&DB.Db_Agent_Level{Uid: 局_User.Id, UPAgentId: 上级代理ID, Level: i + 2}).Error
			if err != nil {
				return err
			}
		}

		return nil
	})

	return 局_User, nil
}

// 0 非代理,1 一级代理 2 二级代理 3 三级代理 本包专用, 方式环形导包
func 取Id代理级别(用户ID int) int {
	var Count int64 = 0
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Where("Uid=?", 用户ID).Count(&Count)
	return int(Count)
}
func Id置新密码(Id int, NewPassWord string) error {
	if Id == 0 {
		return errors.New("Id不能为0")
	}

	err := global.GVA_DB.Model(DB.DB_User{}).Where("Id = ?", Id).Updates(map[string]interface{}{"PassWord": Md5String(NewPassWord)}).Error

	if err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("Id置新密码失败:%v,%v,%v", Id, NewPassWord, err.Error()))
		return errors.New("修改密码失败")
	}
	return nil

}

func Q取总数() int64 {
	var 局_总数 int64
	_ = global.GVA_DB.Model(DB.DB_User{}).Count(&局_总数).Error
	return 局_总数
}
func Id取上级代理ID(Id int) int {
	if Id == 0 {
		return 0
	}
	var 上级代理ID int
	global.GVA_DB.Model(DB.DB_User{}).Select("UPAgentId").Where("Id=?", Id).First(&上级代理ID)
	return 上级代理ID
}

func Id取下级代理分成最高(Id int) int {
	if Id == 0 {
		return 0
	}
	var 上级代理ID = 0
	global.GVA_DB.Model(DB.DB_User{}).Select(" Max(AgentDiscount) ").Where("UPAgentId=?", Id).First(&上级代理ID)
	//如果没有下级代理,报错,直接返回0
	return 上级代理ID
}
