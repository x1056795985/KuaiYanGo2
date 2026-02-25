package Ser_Ka

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_AgentInventory"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/kaClassUpPrice"

	dbm "server/new/app/models/db"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

func KaId是否存在(Appid int, id int) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.DB_Ka{}).Select("1").Where("Id=?", id).Where("AppId=?", Appid).First(&Count)
	return result.Error == nil

}

// Ka批量创建 切片可以直接传址 所以放切片  卡信息切片[:]
// 有效期 0=9999999999 无限制
func Ka批量创建(卡信息切片 []DB.DB_Ka, 卡类id, 制卡人id int, 制卡人账号 string, 管理员备注 string, 代理备注 string, 有效期时间戳 int64) error {
	if len(卡信息切片) >= 2621 { //65535 / 25 ≈ 2621.4。所以一次最多只能插入2621条记录
		return errors.New("每批次最大数量不能超过2621")
	}

	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(卡类id)
	if err != nil { //估计是卡类不存在
		return err
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for i := range 卡信息切片 {
			if 卡信息切片[i].Name == "" {
				for I := 0; I < 10; I++ {
					卡信息切片[i].Name = KaClass详细信息.Prefix
					卡信息切片[i].Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)

					var Count int64
					err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
					if Count == 0 {
						break
					}
					if I == 9 {
						return errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
					}
				}
			} else {

				var Count int64
				err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
				if Count == 1 {
					return errors.New("卡号:" + 卡信息切片[i].Name + "已存在无法使用")
				}
			}

			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterId = 制卡人id
			卡信息切片[i].RegisterUser = 制卡人账号
			卡信息切片[i].RegisterTime = time.Now().Unix()
			卡信息切片[i].AdminNote = 管理员备注
			卡信息切片[i].AgentNote = 代理备注
			卡信息切片[i].VipTime = KaClass详细信息.VipTime
			卡信息切片[i].InviteCount = KaClass详细信息.InviteCount
			卡信息切片[i].RMb = KaClass详细信息.RMb
			卡信息切片[i].VipNumber = KaClass详细信息.VipNumber
			卡信息切片[i].Money = KaClass详细信息.Money
			卡信息切片[i].AgentMoney = KaClass详细信息.AgentMoney
			卡信息切片[i].UserClassId = KaClass详细信息.UserClassId
			卡信息切片[i].NoUserClass = KaClass详细信息.NoUserClass
			卡信息切片[i].KaType = KaClass详细信息.KaType
			卡信息切片[i].MaxOnline = KaClass详细信息.MaxOnline
			卡信息切片[i].Num = 0
			卡信息切片[i].NumMax = KaClass详细信息.Num
			卡信息切片[i].User = ""
			卡信息切片[i].UserTime = ""
			卡信息切片[i].UseTime = 0
			卡信息切片[i].InviteUser = ""
			卡信息切片[i].EndTime = 9999999999
			if 有效期时间戳 != 0 {
				卡信息切片[i].EndTime = 有效期时间戳
			}
		}
		err = tx.Model(DB.DB_Ka{}).Create(&卡信息切片).Error

		return err
	})

}

// Ka代理批量购买 切片可以直接传址 所以放切片  卡信息切片[:]
// 有效期 0=9999999999 无限制
func Ka代理批量购买(c *gin.Context, 卡信息切片 []DB.DB_Ka, 卡类id, 购卡人Id int, 代理备注 string, 有效期时间戳 int64, ip string) error {
	var 局_价格组成 struct {
		总卡类价格 float64

		总调价  float64 //这个是已经*数量的
		调价详情 []dbm.DB_KaClassUpPrice
		购买数量 int64

		总付款金额 float64
	}
	if len(卡信息切片) >= 2621 { //65535 / 25 ≈ 2621.4。所以一次最多只能插入2621条记录
		return errors.New("每批次最大数量不能超过2621")
	}
	局_价格组成.购买数量 = int64(len(卡信息切片))

	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(卡类id)
	if err != nil { //估计是卡类不存在
		return err
	}
	局_价格组成.总卡类价格 = utils.Float64乘int64(KaClass详细信息.AgentMoney, 局_价格组成.购买数量)

	局_购卡人信息, ok := Ser_User.Id取详情(购卡人Id)
	if !ok {
		return errors.New("用户不存在")
	}

	局_价格组成.总调价, 局_价格组成.调价详情, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 卡类id, 局_购卡人信息.UPAgentId)
	if err != nil {
		return err
	}

	局_价格组成.总调价 = utils.Float64乘int64(局_价格组成.总调价, 局_价格组成.购买数量)
	局_价格组成.总付款金额 = utils.Float64加float64(局_价格组成.总调价, 局_价格组成.总卡类价格, 2)

	if 局_购卡人信息.Rmb < 局_价格组成.总付款金额 { //先检查一遍,节约事务性能
		return fmt.Errorf("余额不足 (当前余额:%.2f < 需支付:%.2f)", 局_购卡人信息.Rmb, 局_价格组成.总付款金额)
	}

	if 局_价格组成.总付款金额 < 0 {
		return errors.New("卡类代理价格异常")
	}

	var 新余额 float64
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {

		// 减少余额
		err = tx.Exec("UPDATE db_User SET RMB = RMB - ? WHERE Id = ?", 局_价格组成.总付款金额, 局_购卡人信息.Id).Error
		if err != nil {
			global.GVA_LOG.Error(strconv.Itoa(局_购卡人信息.Id) + "Id余额减少失败:" + err.Error())
			return errors.New("余额减少失败查看服务器日志检查原因")
		}

		// 查询新余额
		err = tx.Raw("SELECT RMB FROM db_User WHERE Id = ?", 局_购卡人信息.Id).Scan(&新余额).Error
		if err != nil {
			global.GVA_LOG.Error(strconv.Itoa(局_购卡人信息.Id) + "Id查询余额失败:" + err.Error())
			return errors.New("查询余额失败查看服务器日志检查原因")
		}

		if 新余额 < 0 {
			// 余额不足,回滚并返回   表必须InnoDB引擎才可以,否则会真实发生扣余额,
			return errors.New("用户余额不足,缺少:" + utils.Float64到文本(utils.Float64取绝对值(新余额), 2))
		}
		//扣款成功,开始制卡
		for i := range 卡信息切片 {
			if 卡信息切片[i].Name == "" {
				for I := 0; I < 10; I++ {
					卡信息切片[i].Name = KaClass详细信息.Prefix
					卡信息切片[i].Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
					var Count int64
					err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
					if Count == 0 {
						break
					}
					if I == 9 {
						return errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
					}
				}
			} else {

				var Count int64
				err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
				if Count == 1 {
					return errors.New("卡号:" + 卡信息切片[i].Name + "已存在无法使用")
				}
			}
			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterId = 局_购卡人信息.Id
			卡信息切片[i].RegisterUser = 局_购卡人信息.User
			卡信息切片[i].RegisterTime = time.Now().Unix()
			卡信息切片[i].AdminNote = ""
			卡信息切片[i].AgentNote = 代理备注
			卡信息切片[i].VipTime = KaClass详细信息.VipTime
			卡信息切片[i].InviteCount = KaClass详细信息.InviteCount
			卡信息切片[i].RMb = KaClass详细信息.RMb
			卡信息切片[i].VipNumber = KaClass详细信息.VipNumber
			卡信息切片[i].Money = KaClass详细信息.Money
			卡信息切片[i].AgentMoney = KaClass详细信息.AgentMoney
			卡信息切片[i].UserClassId = KaClass详细信息.UserClassId
			卡信息切片[i].NoUserClass = KaClass详细信息.NoUserClass
			卡信息切片[i].KaType = KaClass详细信息.KaType
			卡信息切片[i].MaxOnline = KaClass详细信息.MaxOnline
			卡信息切片[i].Num = 0
			卡信息切片[i].NumMax = KaClass详细信息.Num
			卡信息切片[i].User = ""
			卡信息切片[i].UserTime = ""
			卡信息切片[i].UseTime = 0
			卡信息切片[i].InviteUser = ""
			卡信息切片[i].EndTime = 9999999999
			if 有效期时间戳 != 0 {
				卡信息切片[i].EndTime = 有效期时间戳
			}
		}

		err = tx.Create(&卡信息切片).Error //不知道为什么 不生效,插入的是 全null的记录 2023/8/16 找到原因了是因为上面把tx 重新赋值了导致的,

		//err = global.GVA_DB.Create(&卡信息切片).Error //这样也可以, 成功上边的也使用,不成功就回退
		return err
	})
	if err != nil {
		//制卡失败直接返回
		return err
	}

	数组_卡号 := make([]string, 0, len(卡信息切片))
	var builder strings.Builder
	for i := 0; i < len(卡信息切片); i++ {
		数组_卡号 = append(数组_卡号, 卡信息切片[i].Name)
		builder.WriteString(strconv.Itoa(卡信息切片[i].Id))
		builder.WriteString(",")
	}
	局_ID列表 := builder.String()
	局_文本 := fmt.Sprintf("代理购卡[%s -> %s],卡号ID{%s},|新余额≈%s", Ser_AppInfo.App取AppName(KaClass详细信息.AppId), KaClass详细信息.Name, 局_ID列表, utils.Float64到文本(新余额, 2))
	Ser_Log.Log_写余额日志(局_购卡人信息.User, ip, 局_文本, utils.Float64取负值(局_价格组成.总付款金额))
	局_文本 = fmt.Sprintf("新制卡号:[%s -> %s],批次id:{{批次id}}({{卡号索引}}/%d)", Ser_AppInfo.App取AppName(卡信息切片[0].AppId), Ser_KaClass.Id取Name(卡信息切片[0].KaClassId), len(卡信息切片))
	Ser_Log.Log_写卡号操作日志(局_购卡人信息.User, ip, 局_文本, 数组_卡号, 1, agentLevel.L_agentLevel.Q取Id代理级别(c, 局_购卡人信息.Id))

	//开始分利润 20240202 mark处理重构以后改事务
	//先分成 代理调价信息的价格 然后再计算百分比的价格
	if 局_价格组成.总调价 > 0 {
		局_日志前缀 := fmt.Sprintf("下级代理:%s,制卡ID{%s}", 局_购卡人信息.User, 局_ID列表)
		err = agent.L_agent.Z执行调价信息分成(c, 局_价格组成.调价详情, 局_价格组成.购买数量, 局_日志前缀)
		if err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("Z执行调价信息分成失败:", err.Error()))
		}
	}
	//然后再计算百分比的价格
	代理分成数据, err2 := agent.L_agent.D代理分成计算(c, 局_购卡人信息.Id, 局_价格组成.总卡类价格)
	if err2 != nil {
		global.GVA_LOG.Error(fmt.Sprintf("代理制卡分成计算失败:%s,代理ID:%d,金额¥%v,卡号ID:%s", err2.Error(), 局_购卡人信息.UPAgentId, 局_价格组成.总卡类价格, 局_ID列表))
		return err2
	}
	if len(代理分成数据) >= 0 {
		局_日志前缀 := fmt.Sprintf("下级代理:%s,制卡ID{%s},", 局_购卡人信息.User, 局_ID列表)
		err = agent.L_agent.Z执行百分比代理分成(c, 代理分成数据, 局_价格组成.总卡类价格, 局_日志前缀, 局_价格组成.总调价 == 0)
		if err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("Z执行百分比代理分成:%s", err.Error()))

		}
	}
	// 分成结束==============
	return nil
}

// Ka代理批量购买 切片可以直接传址 所以放切片  卡信息切片[:]
// 有效期 0=9999999999 无限制
func Ka代理批量库存购买(c *gin.Context, 卡信息切片 []DB.DB_Ka, 库存Id, 制卡数量, 购卡人Id int, 代理备注 string, ip string) error {
	if 制卡数量 <= 0 {
		return errors.New("生成数量必须大于0")
	}
	if 制卡数量 > 2621 {
		return errors.New("生成数量每批最大2621")
	}

	局_库存详情, ok := Ser_AgentInventory.Id取详情(库存Id)

	if !ok {
		return errors.New("库存ID不存在")
	}

	if 局_库存详情.Uid != 购卡人Id {
		return errors.New("只能使用归属自己的库存制卡")
	}

	if 局_库存详情.NumMax-局_库存详情.Num < 制卡数量 {
		return errors.New("库存剩余可制卡次数不足")
	}
	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(局_库存详情.KaClassId)
	if err != nil {
		return errors.New("库存所属卡类id不存在,可能已被管理员删除")
	}
	局_购卡人User := Ser_User.Id取User(购卡人Id)
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 减少库存
		err = tx.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 局_库存详情.Id).Update("Num", gorm.Expr("Num + ?", 制卡数量)).Error
		if err != nil {
			return err
		}
		var 剩余库存 int
		// 查询新余额
		err = tx.Model(DB.Db_Agent_库存卡包{}).Select("NumMax-Num").Where("Id = ?", 局_库存详情.Id).Take(&剩余库存).Error
		if err != nil {
			return err
		}

		if 剩余库存 < 0 {
			// 余额不足,回滚并返回   表必须InnoDB引擎才可以,否则会真实发生扣余额,
			return errors.New("库存可用次数不足,缺少次数:" + strconv.Itoa(-剩余库存))
		}

		//扣库存成功,开始制卡
		for i := range 卡信息切片 {
			if 卡信息切片[i].Name == "" {
				for I := 0; I < 10; I++ {
					卡信息切片[i].Name = KaClass详细信息.Prefix
					卡信息切片[i].Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)

					var Count int64
					err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
					if Count == 0 {
						break
					}
					if I == 9 {
						return errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
					}
				}
			} else {

				var Count int64
				err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
				if Count == 1 {
					return errors.New("卡号:" + 卡信息切片[i].Name + "已存在无法使用")
				}
			}
			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterId = 购卡人Id
			卡信息切片[i].RegisterUser = 局_购卡人User
			卡信息切片[i].RegisterTime = time.Now().Unix()
			卡信息切片[i].AdminNote = ""
			卡信息切片[i].AgentNote = 代理备注
			卡信息切片[i].VipTime = KaClass详细信息.VipTime
			卡信息切片[i].InviteCount = KaClass详细信息.InviteCount
			卡信息切片[i].RMb = KaClass详细信息.RMb
			卡信息切片[i].VipNumber = KaClass详细信息.VipNumber
			卡信息切片[i].Money = KaClass详细信息.Money
			卡信息切片[i].AgentMoney = KaClass详细信息.AgentMoney
			卡信息切片[i].UserClassId = KaClass详细信息.UserClassId
			卡信息切片[i].NoUserClass = KaClass详细信息.NoUserClass
			卡信息切片[i].KaType = KaClass详细信息.KaType
			卡信息切片[i].MaxOnline = KaClass详细信息.MaxOnline
			卡信息切片[i].Num = 0
			卡信息切片[i].NumMax = KaClass详细信息.Num
			卡信息切片[i].User = ""
			卡信息切片[i].UserTime = ""
			卡信息切片[i].UseTime = 0
			卡信息切片[i].InviteUser = ""
			卡信息切片[i].EndTime = 局_库存详情.EndTime
		}

		err = tx.Create(&卡信息切片).Error //不知道为什么 不生效,插入的是 全null的记录
		//err = global.GVA_DB.Create(&卡信息切片).Error //这样也可以, 成功上边的也使用,不成功就回退
		return err
	})
	if err != nil {
		//制卡失败直接返回
		return err
	}

	数组_卡号 := make([]string, 0, len(卡信息切片))
	for i := 0; i < len(卡信息切片); i++ {
		数组_卡号 = append(数组_卡号, 卡信息切片[i].Name)
	}
	局_文本 := fmt.Sprintf("制卡库存Id:%d,应用:%s,卡类:%s,批次id:{{批次id}}({{卡号索引}}/%d)", 局_库存详情.Id, Ser_AppInfo.App取AppName(卡信息切片[0].AppId), Ser_KaClass.Id取Name(卡信息切片[0].KaClassId), len(卡信息切片))
	go Ser_Log.Log_写卡号操作日志(局_购卡人User, ip, 局_文本, 数组_卡号, 1, agentLevel.L_agentLevel.Q取Id代理级别(c, 购卡人Id))
	return nil
}
func Q取总数() int64 {

	var 局_总数 int64
	_ = global.GVA_DB.Model(DB.DB_Ka{}).Count(&局_总数).Error
	return 局_总数
}

// 有效期 0=9999999999 无限制
func Ka单卡创建(卡类id, 制卡人ID int, 制卡人账号 string, 管理员备注 string, 代理备注 string, 有效期时间戳 int64) (卡信息切片 DB.DB_Ka, err error) {

	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(卡类id)
	if err != nil { //估计是卡类不存在
		return 卡信息切片, err
	}

	for I := 0; I < 10; I++ {
		卡信息切片.Name = KaClass详细信息.Prefix
		卡信息切片.Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)

		if !Ka卡号是否存在(卡信息切片.Name) {
			break
		}
		if I == 9 {
			return 卡信息切片, errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
		}
	}

	卡信息切片.AppId = KaClass详细信息.AppId
	卡信息切片.KaClassId = KaClass详细信息.Id
	卡信息切片.Status = 1
	卡信息切片.RegisterId = 制卡人ID
	卡信息切片.RegisterUser = 制卡人账号
	卡信息切片.RegisterTime = time.Now().Unix()
	卡信息切片.AdminNote = 管理员备注
	卡信息切片.AgentNote = 代理备注
	卡信息切片.VipTime = KaClass详细信息.VipTime
	卡信息切片.InviteCount = KaClass详细信息.InviteCount
	卡信息切片.RMb = KaClass详细信息.RMb
	卡信息切片.VipNumber = KaClass详细信息.VipNumber
	卡信息切片.Money = KaClass详细信息.Money
	卡信息切片.AgentMoney = KaClass详细信息.AgentMoney
	卡信息切片.UserClassId = KaClass详细信息.UserClassId
	卡信息切片.NoUserClass = KaClass详细信息.NoUserClass
	卡信息切片.KaType = KaClass详细信息.KaType
	卡信息切片.MaxOnline = KaClass详细信息.MaxOnline
	卡信息切片.Num = 0
	卡信息切片.NumMax = KaClass详细信息.Num
	卡信息切片.User = ""
	卡信息切片.UserTime = ""
	卡信息切片.UseTime = 0
	卡信息切片.InviteUser = ""
	卡信息切片.EndTime = 9999999999
	if 有效期时间戳 != 0 {
		卡信息切片.EndTime = 有效期时间戳
	}
	return 卡信息切片, global.GVA_DB.Model(DB.DB_Ka{}).Create(&卡信息切片).Error
}

func Ka修改状态(id []int, status int) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Update("Status", status).Error
}

// 卡号模式关联 软件用户,同时冻结解冻,用户模式不关联  冻结会注销在线
func Ka修改状态_同步卡号模式软件用户(id []int, status int) error {
	局_db := global.GVA_DB
	局_sql := `SELECT DISTINCT AppId  FROM db_App_Info  WHERE AppId IN (SELECT DISTINCT AppId  FROM db_Ka  WHERE Id IN ?) AND AppType IN (3,4)`

	var 局数组_卡号Appid []int
	局_db = 局_db.Raw(局_sql, id).Scan(&局数组_卡号Appid)
	//如果卡号id数组内没有卡号类型应用id,直接执行就可以,
	if len(局数组_卡号Appid) == 0 {
		return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Update("Status", status).Error
	}
	//开启事务执行
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		//先获取所有卡号的AppId
		局_ka := make([]DB.DB_Ka, 0, len(id))
		err := tx.Model(DB.DB_Ka{}).Select("Id,AppId").Where("Id IN ?", id).Scan(&局_ka).Error
		if err != nil {
			return err
		}

		//先处理用户类型AppId的卡号
		局_map := make(map[int][]int, len(id)+1) //给非卡号模式的id分配一个Appid位置
		for _, 值 := range 局_ka {
			局_最终AppId := 值.AppId
			//不是卡号模式的 卡Id 直接赋值AppId 1   一会一起处理
			if !utils.S数组_整数是否存在(局数组_卡号Appid, 局_最终AppId) {
				局_最终AppId = 1
			}

			//判断是否存在键,不存在创建内存空间
			if _, ok := 局_map[局_最终AppId]; ok {
				局_map[局_最终AppId] = make([]int, len(id))
			}
			//把 卡id追加到 appid的键值内
			局_map[局_最终AppId] = append(局_map[局_最终AppId], 值.Id)
		}

		for _, 值 := range 局_ka {
			局_最终AppId := 值.AppId
			//不是卡号模式的 卡Id 直接赋值AppId 1   一会一起处理
			if !utils.S数组_整数是否存在(局数组_卡号Appid, 局_最终AppId) {
				局_最终AppId = 1
			}

			//判断是否存在键,不存在创建内存空间
			if _, ok := 局_map[局_最终AppId]; ok {
				局_map[局_最终AppId] = make([]int, len(id))
			}
			//把 卡id追加到 appid的键值内
			局_map[局_最终AppId] = append(局_map[局_最终AppId], 值.Id)
		}
		// 通appid 的卡id 合并完毕,开始冻结解冻

		for AppId := range 局_map {
			err = tx.Model(DB.DB_Ka{}).Where("Id IN ? ", 局_map[AppId]).Update("Status", status).Error
			if err != nil {
				return err //出错就返回并回滚
			}

			//同步冻结软件用户AppId  用户模式的卡号ID不用冻结
			if AppId >= 10000 {
				err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).Where("Uid IN ? ", 局_map[AppId]).Update("Status", status).Error
				if err != nil {
					return err //出错就返回并回滚
				}
			}

			//如果是冻结同时注销在线的uid
			if status == 2 {
				_ = Ser_LinkUser.Set批量注销Uid数组(局_map[AppId], AppId, Ser_LinkUser.Z注销_管理员手动注销)
			}

		}
		return nil //处理完毕 返回
	})

}

// 代理权限不在这里校验,在api接口校验
func Ka更换卡号(c *gin.Context, id, 代理Id int, ip string) error {
	局_卡号详情, err := Id取详情(id)
	if err != nil {
		return errors.New("卡号ID不存在")
	}
	if 局_卡号详情.Num != 0 {
		return errors.New("卡号已使用无法更换")
	}

	if Ser_AppInfo.App是否为卡号(局_卡号详情.AppId) {
		return errors.New("应用为卡号登录模式,无法更改卡号")
	}
	代理User := Ser_User.Id取User(代理Id)
	if 局_卡号详情.RegisterUser != 代理User {
		return errors.New("只有自己的卡才可以更换卡号")
	}

	if 局_卡号详情.Status != 1 {
		return errors.New("卡号已冻结,暂不可更换卡号")
	}
	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(局_卡号详情.KaClassId)
	if err != nil {
		return errors.New("卡号对应卡类ID不存在,可能已删除")
	}

	var 局_新卡号 = ""
	for I := 0; I < 10; I++ {
		局_新卡号 = KaClass详细信息.Prefix
		局_新卡号 += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)

		if !Ka卡号是否存在(局_新卡号) {
			break
		}
		if I == 9 {
			return errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
		}
	}

	err = global.GVA_DB.Model(DB.DB_Ka{}).Where("Id = ? ", id).Update("Name", 局_新卡号).Error
	if err == nil {
		局_log := fmt.Sprintf("操作更换卡号:  %s  ->  %s", 局_卡号详情.Name, 局_新卡号)
		Ser_Log.Log_写卡号操作日志(代理User, ip, 局_log, []string{局_卡号详情.Name}, 3, agentLevel.L_agentLevel.Q取Id代理级别(c, 代理Id))
	}
	return err
}

func Ka修改已用次数加一(id []int) error {
	now := time.Now().Unix()
	return global.GVA_DB.Model(DB.DB_Ka{}).
		Where("Id IN ?", id).
		Updates(map[string]interface{}{
			"Num":      gorm.Expr("Num + 1"),
			"UserTime": gorm.Expr("CONCAT(UserTime, ?)", strconv.Itoa(int(now))+","),
			"UseTime":  now,
		}).Error
}

func Ka修改管理员备注(id []int, AdminNote string) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Update("AdminNote", AdminNote).Error
}

func Ka修改代理备注(代理User string, id []int, AgentNote string) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Where("RegisterUser = ? ", 代理User).Update("AgentNote", AgentNote).Error
}

func Ka卡号是否存在(卡号 string) bool {
	var Count int64
	//只判断卡号不判断应用, 不然后期卡号读取很麻烦,因为有可能不知道应用信息   卡号所有应用都不可以重复
	_ = global.GVA_DB.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡号).First(&Count)
	return Count != 0
}

func Ka卡号取id(Appid int, 卡号 string) int {
	var Id int
	global.GVA_DB.Model(DB.DB_Ka{}).Select("Id").Where("Name=?", 卡号).Where("AppId=?", Appid).First(&Id)
	return Id
}

func Id取制卡人(Id int) string {
	var 制卡人 string
	global.GVA_DB.Model(DB.DB_Ka{}).Select("RegisterUser").Where("Id=?", Id).First(&制卡人)
	return 制卡人
}
func Id检测制卡人(Id []int, 制卡人 string) bool {
	var 实际制卡人 []string
	global.GVA_DB.Model(DB.DB_Ka{}).Distinct("RegisterUser").Where("Id IN ?", Id).Find(&实际制卡人)
	if len(实际制卡人) == 1 && 制卡人 == 实际制卡人[0] {
		return true
	}

	return false
}
func Id取卡号(Id int) string {
	var 卡号 string
	global.GVA_DB.Model(DB.DB_Ka{}).Select("Name").Where("Id=?", Id).First(&卡号)
	return 卡号
}
func Ka卡号取详情(卡号 string) (卡号详情卡号 DB.DB_Ka, ok error) {
	err := global.GVA_DB.Model(DB.DB_Ka{}).Where("Name=?", 卡号).First(&卡号详情卡号).Error
	return 卡号详情卡号, err
}
func Id取详情(Id int) (卡号详情卡号 DB.DB_Ka, err error) {
	err = global.GVA_DB.Model(DB.DB_Ka{}).Where("Id=?", Id).First(&卡号详情卡号).Error
	return 卡号详情卡号, err
}
func Ka取已购卡列表(制卡人账号 string, 页数, 数量 int) (卡号详情卡号 []DB.DB_Ka, ok error) {
	err := global.GVA_DB.Model(DB.DB_Ka{}).Order("Id DESC").Where("RegisterUser=?", 制卡人账号).Limit(数量).Offset((页数 - 1) * 数量).Find(&卡号详情卡号).Error
	return 卡号详情卡号, err
}

func Get卡号总数(AppId int) int {
	var 局_总数 int64
	err := global.GVA_DB.Model(DB.DB_Ka{}).Where("AppId=?", AppId).Count(&局_总数).Error
	if err != nil {
		return 0
	}
	return int(局_总数)
}
func Get卡类卡号总数(ClassId int) int {
	var 局_总数 int64
	err := global.GVA_DB.Model(DB.DB_Ka{}).Where("KaClassId=?", ClassId).Count(&局_总数).Error
	if err != nil {
		return 0
	}
	return int(局_总数)
}
func Get应用已用和未用数量(AppId int) (已用, 可用 int64) {
	global.GVA_DB.Model(DB.DB_Ka{}).Where("AppId=?", AppId).Where("Num=NumMax").Count(&已用)
	global.GVA_DB.Model(DB.DB_Ka{}).Where("AppId=?", AppId).Where("Num<NumMax").Count(&可用)
	return
}
func Get卡类已用和未用数量(卡类Id int) (已用, 可用 int64) {
	global.GVA_DB.Model(DB.DB_Ka{}).Where("KaClassId=?", 卡类Id).Where("Num=NumMax").Count(&已用)
	global.GVA_DB.Model(DB.DB_Ka{}).Where("KaClassId=?", 卡类Id).Where("Num<NumMax").Count(&可用)
	return
}
