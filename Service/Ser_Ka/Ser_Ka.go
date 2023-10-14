package Ser_Ka

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"server/Service/Ser_Agent"
	"server/Service/Ser_AgentInventory"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/Service/Ser_UserClass"
	"server/global"
	DB "server/structs/db"
	"server/utils"
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
func Ka批量创建(卡信息切片 []DB.DB_Ka, 卡类id int, 制卡人账号 string, 管理员备注 string, 代理备注 string, 有效期时间戳 int64) error {

	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(卡类id)
	if err != nil { //估计是卡类不存在
		return err
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for i := range 卡信息切片 {
			if 卡信息切片[i].Name == "" {
				for I := 0; I < 10; I++ {
					卡信息切片[i].Name = KaClass详细信息.Prefix
					卡信息切片[i].Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix)-1, KaClass详细信息.KaStringType)
					卡信息切片[i].Name += 生成校验字符(卡信息切片[i].Name)
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
				if !Ka校验卡号(卡信息切片[i].Name) {
					return errors.New("卡号:" + 卡信息切片[i].Name + "不符合校验规则,仅可指定系统生成后删除的卡号")
				}
				var Count int64
				err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
				if Count == 1 {
					return errors.New("卡号:" + 卡信息切片[i].Name + "已存在无法使用")
				}
			}

			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterUser = 制卡人账号
			卡信息切片[i].RegisterTime = int(time.Now().Unix())
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
func Ka代理批量购买(卡信息切片 []DB.DB_Ka, 卡类id, 购卡人Id int, 代理备注 string, 有效期时间戳 int64, ip string) error {

	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(卡类id)
	if err != nil { //估计是卡类不存在
		return err
	}
	局_购卡人信息, ok := Ser_User.Id取详情(购卡人Id)
	if !ok {
		return errors.New("用户不存在")
	}
	局_总计金额 := utils.Float64乘int64(KaClass详细信息.AgentMoney, int64(len(卡信息切片)))
	if 局_购卡人信息.Rmb < 局_总计金额 { //先检查一遍,节约事务性能

		return errors.New("余额不足")

	}
	var 新余额 float64
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {

		// 减少余额
		err = tx.Exec("UPDATE db_User SET RMB = RMB - ? WHERE Id = ?", 局_总计金额, 局_购卡人信息.Id).Error
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
					卡信息切片[i].Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix)-1, KaClass详细信息.KaStringType)
					卡信息切片[i].Name += 生成校验字符(卡信息切片[i].Name)
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
				if !Ka校验卡号(卡信息切片[i].Name) {
					return errors.New("卡号:" + 卡信息切片[i].Name + "不符合校验规则,仅可指定系统生成后删除的卡号")
				}
				var Count int64
				err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
				if Count == 1 {
					return errors.New("卡号:" + 卡信息切片[i].Name + "已存在无法使用")
				}
			}
			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterUser = 局_购卡人信息.User
			卡信息切片[i].RegisterTime = int(time.Now().Unix())
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
	go Ser_Log.Log_写余额日志(局_购卡人信息.User, ip, 局_文本, utils.Float64取负值(局_总计金额))
	局_文本 = fmt.Sprintf("新制卡号:[%s -> %s],同时间批次({{卡号索引}}/%d)", Ser_AppInfo.App取AppName(卡信息切片[0].AppId), Ser_KaClass.Id取Name(卡信息切片[0].KaClassId), len(卡信息切片))
	go Ser_Log.Log_写卡号操作日志(局_购卡人信息.User, ip, 局_文本, 数组_卡号, 1, Ser_Agent.Q取Id代理级别(局_购卡人信息.Id))

	//开始分利润
	var 下级信息 DB.DB_User
	下级信息 = 局_购卡人信息
	局_下级分成百分比 := 0
	for {

		局_百分比分成 := utils.Float64除int64(int64(下级信息.AgentDiscount-局_下级分成百分比), 100, 2)

		局_分成金额 := utils.Float64乘Float64(局_总计金额, 局_百分比分成)
		if 局_分成金额 > 0 {
			新余额, err2 := Ser_User.Id余额增减(下级信息.Id, 局_分成金额, true)
			if err2 != nil {
				//,一般不会出现,除非用户不存在
				global.GVA_LOG.Error(fmt.Sprintf("代理制卡分成余额增加失败:%s,代理ID:%d,金额¥%v,卡号ID:%s", err2.Error(), 下级信息.Id, 局_分成金额, 局_ID列表))
			} else {
				str := fmt.Sprintf("下级代理:%s,制卡ID{%s},分成:¥%s (¥%s*(%d%%-%d%%)),|新余额≈%s", 局_购卡人信息.User, 局_ID列表, utils.Float64到文本(局_分成金额, 2), utils.Float64到文本(局_总计金额, 2), 下级信息.AgentDiscount, 局_下级分成百分比, utils.Float64到文本(新余额, 2))
				if 下级信息.Id == 局_购卡人信息.Id {
					str = fmt.Sprintf("代理制卡ID{%s},自消费分成:¥%s (¥%s*(%d%%-%d%%)),|新余额≈%s", 局_ID列表, utils.Float64到文本(局_分成金额, 2), utils.Float64到文本(局_总计金额, 2), 下级信息.AgentDiscount, 局_下级分成百分比, utils.Float64到文本(新余额, 2))
				}
				Ser_Log.Log_写余额日志(下级信息.User, ip, str, 局_分成金额)
			}
		}

		if 下级信息.UPAgentId <= 0 {
			//上级是管理员了 跳出循环
			break
		}

		局_下级分成百分比 = 下级信息.AgentDiscount
		下级信息, ok = Ser_User.Id取详情(下级信息.UPAgentId)
		if !ok {
			//,一般不会出现,除非用户不存在
			global.GVA_LOG.Error(fmt.Sprintf("代理制卡分成,取上级代理ID:%d详情失败,下级代理ID:%d,卡号ID:%s", 下级信息.UPAgentId, 下级信息.Id, 局_ID列表))
			break
		}

	}

	return nil
}

// Ka代理批量购买 切片可以直接传址 所以放切片  卡信息切片[:]
// 有效期 0=9999999999 无限制
func Ka代理批量库存购买(卡信息切片 []DB.DB_Ka, 库存Id, 制卡数量, 购卡人Id int, 代理备注 string, ip string) error {
	if 制卡数量 <= 0 {
		return errors.New("生成数量必须大于0")
	}
	if 制卡数量 > 500 {
		return errors.New("生成数量每批最大500")
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
					卡信息切片[i].Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix)-1, KaClass详细信息.KaStringType)
					卡信息切片[i].Name += 生成校验字符(卡信息切片[i].Name)
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
				if !Ka校验卡号(卡信息切片[i].Name) {
					return errors.New("卡号:" + 卡信息切片[i].Name + "不符合校验规则,仅可指定系统生成后删除的卡号")
				}
				var Count int64
				err = tx.Select("1").Model(DB.DB_Ka{}).Where("Name=?", 卡信息切片[i].Name).Scan(&Count).Error
				if Count == 1 {
					return errors.New("卡号:" + 卡信息切片[i].Name + "已存在无法使用")
				}
			}
			卡信息切片[i].AppId = KaClass详细信息.AppId
			卡信息切片[i].KaClassId = KaClass详细信息.Id
			卡信息切片[i].Status = 1
			卡信息切片[i].RegisterUser = 局_购卡人User
			卡信息切片[i].RegisterTime = int(time.Now().Unix())
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
	局_文本 := fmt.Sprintf("制卡库存Id:%d,应用:%s,卡类:%s,同时间批次({{卡号索引}}/%d)", 局_库存详情.Id, Ser_AppInfo.App取AppName(卡信息切片[0].AppId), Ser_KaClass.Id取Name(卡信息切片[0].KaClassId), len(卡信息切片))
	go Ser_Log.Log_写卡号操作日志(局_购卡人User, ip, 局_文本, 数组_卡号, 1, Ser_Agent.Q取Id代理级别(购卡人Id))
	return nil
}
func Q取总数() int64 {

	var 局_总数 int64
	_ = global.GVA_DB.Model(DB.DB_Ka{}).Count(&局_总数).Error
	return 局_总数
}

// 有效期 0=9999999999 无限制
func Ka单卡创建(卡类id int, 制卡人账号 string, 管理员备注 string, 代理备注 string, 有效期时间戳 int64) (卡信息切片 DB.DB_Ka, err error) {

	KaClass详细信息, err := Ser_KaClass.KaClass取详细信息(卡类id)
	if err != nil { //估计是卡类不存在
		return 卡信息切片, err
	}

	for I := 0; I < 10; I++ {
		卡信息切片.Name = KaClass详细信息.Prefix
		卡信息切片.Name += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix)-1, KaClass详细信息.KaStringType)
		卡信息切片.Name += 生成校验字符(卡信息切片.Name)
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
	卡信息切片.RegisterUser = 制卡人账号
	卡信息切片.RegisterTime = int(time.Now().Unix())
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
	卡信息切片.InviteUser = ""
	卡信息切片.EndTime = 9999999999
	if 有效期时间戳 != 0 {
		卡信息切片.EndTime = 有效期时间戳
	}
	return 卡信息切片, global.GVA_DB.Model(DB.DB_Ka{}).Create(&卡信息切片).Error
}

func 生成校验字符(str string) string {

	ArrInt := []byte(str)
	Int := 0
	for _, 值 := range ArrInt {
		Int += int(值)
	}
	Int = Int % len(str)

	return string(str[Int])
}
func Ka校验卡号(str string) bool {
	if len(str) < 2 {
		return false
	}
	局_待校验文本 := str[0 : len(str)-1]
	局_校验值 := string(str[len(str)-1])

	return 生成校验字符(局_待校验文本) == 局_校验值
}

func 生成随机字符串(lenNum int, 类型 int) string {

	var CHARS []string
	switch 类型 {
	case 2:
		CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	case 3:
		CHARS = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	default:
		CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
			"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	}

	str := strings.Builder{}
	length := len(CHARS)
	for i := 0; i < lenNum; i++ {
		str.WriteString(CHARS[rand.Intn(length)])
	}
	return str.String()
}

func Ka修改状态(id []int, status int) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Update("Status", status).Error
}

// 代理权限不在这里校验,在api接口校验
func Ka更换卡号(id, 代理Id int, ip string) error {
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
		局_新卡号 += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix)-1, KaClass详细信息.KaStringType)
		局_新卡号 += 生成校验字符(局_新卡号)
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
		Ser_Log.Log_写卡号操作日志(代理User, ip, 局_log, []string{局_卡号详情.Name}, 3, Ser_Agent.Q取Id代理级别(代理Id))
	}
	return err
}
func Ka修改已用次数加一(id []int) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Update("Num", gorm.Expr("Num+1 , UserTime=CONCAT(UserTime,?)", strconv.Itoa(int(time.Now().Unix()))+",")).Error
}

func Ka修改管理员备注(id []int, AdminNote string) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Update("AdminNote", AdminNote).Error
}

func Ka修改代理备注(代理User string, id []int, AgentNote string) error {
	return global.GVA_DB.Model(DB.DB_Ka{}).Where("Id IN ? ", id).Where("RegisterUser = ? ", 代理User).Update("AgentNote", AgentNote).Error
}

// 已用充值卡将相应的卡使用者和推荐人强行扣回充值卡面值,可能扣成负数
func K卡号追回(ID int) (提示 string, 错误 error) {

	卡号详情卡号, err := Id取详情(ID)
	if err != nil {
		return "", err
	}

	if 卡号详情卡号.Num == 0 {
		return "", errors.New("卡号未使用")
	}
	局_is卡号 := Ser_AppInfo.App是否为卡号(卡号详情卡号.AppId)
	局_is计点 := Ser_AppInfo.App是否为计点(卡号详情卡号.AppId)
	局_点数OR时间 := "时间"
	if 局_is计点 {
		局_点数OR时间 = "点数"
	}
	已用用户数组 := utils.W文本_分割文本(卡号详情卡号.User, ",")
	// 0 无需追回,1成功 2失败
	局_追回时间点数结果 := make(map[string]int, len(已用用户数组))
	局_追回积分结果 := make(map[string]int, len(已用用户数组))
	局_追回余额结果 := make(map[string]int, len(已用用户数组))
	局_追回推荐人时间点数结果 := make(map[string]int, len(已用用户数组))

	if 卡号详情卡号.User == "" {
		return "", errors.New("无已充值用户,但有使用次数,可能手动修改使用次数导致的")
	}
	for _, 值 := range 已用用户数组 {
		if 值 == "" {
			continue //如果值为空,到循环尾
		}
		局_应用用户ID := 0
		if 局_is卡号 {
			局_应用用户ID = Ser_AppUser.K卡号取Id(卡号详情卡号.AppId, 值)
		} else {
			局_应用用户ID = Ser_AppUser.User取Id(卡号详情卡号.AppId, 值)
		}

		if 局_应用用户ID == 0 {
			return "", errors.New("应用用户:" + 值 + "不存在")
		}

		//防sb客户放负值 这样操作负值也可以追回
		if 卡号详情卡号.VipTime != 0 {
			err = Ser_AppUser.Id点数增减_批量(卡号详情卡号.AppId, []int{局_应用用户ID}, 卡号详情卡号.VipTime, false)
			if err == nil {
				局_追回时间点数结果[值] = 1
				if 局_is计点 {
					go Ser_Log.Log_写积分点数时间日志(值, "127.0.0.1", "追回卡号:"+卡号详情卡号.Name+",减少用户点数", float64(-卡号详情卡号.VipTime), 卡号详情卡号.AppId, 2)
				} else {
					go Ser_Log.Log_写积分点数时间日志(值, "127.0.0.1", "追回卡号:"+卡号详情卡号.Name+",减少用户时间", float64(-卡号详情卡号.VipTime), 卡号详情卡号.AppId, 3)

				}
			} else {
				局_追回时间点数结果[值] = 2
			}
		}
		if 卡号详情卡号.VipNumber != 0 {
			err = Ser_AppUser.Id积分增减_批量(卡号详情卡号.AppId, []int{局_应用用户ID}, 卡号详情卡号.VipNumber, false)
			go Ser_Log.Log_写积分点数时间日志(值, "127.0.0.1", "追回卡号:"+卡号详情卡号.Name+",减少用户积分", utils.Float64取负值(卡号详情卡号.VipNumber), 卡号详情卡号.AppId, 1)
			if err == nil {
				局_追回积分结果[值] = 1
			} else {
				局_追回积分结果[值] = 2
			}
		}
		if !局_is卡号 && 卡号详情卡号.RMb != 0 {
			err = Ser_User.Id余额增减_批量([]int{Ser_User.User用户名取id(值)}, 卡号详情卡号.RMb, false)
			if err == nil {
				go Ser_Log.Log_写余额日志(值, "127.0.0.1", "追回卡号:"+卡号详情卡号.Name+",减少用户余额", utils.Float64取负值(卡号详情卡号.RMb))
				局_追回余额结果[值] = 1
			} else {
				局_追回余额结果[值] = 2
			}
		}
	}
	//追回 推荐人
	已用推荐人数组 := utils.W文本_分割文本(卡号详情卡号.InviteUser, ",")
	for _, 值 := range 已用推荐人数组 {
		局_应用用户ID := 0
		if 局_is卡号 {
			局_应用用户ID = Ser_AppUser.K卡号取Id(卡号详情卡号.AppId, 值)
		} else {
			局_应用用户ID = Ser_AppUser.User取Id(卡号详情卡号.AppId, 值)
		}

		if 局_应用用户ID == 0 {
			continue
		}
		//防sb客户放负值 这样操作负值也可以追回
		if 卡号详情卡号.VipTime != 0 {
			err = Ser_AppUser.Id点数增减_批量(卡号详情卡号.AppId, []int{局_应用用户ID}, 卡号详情卡号.InviteCount, false)
			if err == nil {
				if 局_is计点 {
					go Ser_Log.Log_写积分点数时间日志(值, "127.0.0.1", "追回卡号:"+卡号详情卡号.Name+",减少用户点数", float64(-卡号详情卡号.VipTime), 卡号详情卡号.AppId, 2)
				} else {
					go Ser_Log.Log_写积分点数时间日志(值, "127.0.0.1", "追回卡号:"+卡号详情卡号.Name+",减少用户时间", float64(-卡号详情卡号.VipTime), 卡号详情卡号.AppId, 3)

				}
				局_追回推荐人时间点数结果[值] = 1
			} else {
				局_追回推荐人时间点数结果[值] = 2
			}
		}
	}
	log := ""
	局_成功 := ""
	局_失败 := ""
	for 值 := range 局_追回时间点数结果 {
		if 局_追回时间点数结果[值] == 1 {
			局_成功 += 值 + ","
		} else if 局_追回时间点数结果[值] == 2 {
			局_失败 += 值 + ","
		}
	}
	if 局_成功 != "" {
		log += "追回" + 局_点数OR时间 + "成功(" + 局_成功 + "),"
	}
	if 局_失败 != "" {
		log += "追回" + 局_点数OR时间 + "失败(" + 局_成功 + "),"
	}

	局_成功 = ""
	局_失败 = ""
	for 值 := range 局_追回积分结果 {
		if 局_追回积分结果[值] == 1 {
			局_成功 += 值 + ","
		} else if 局_追回积分结果[值] == 2 {
			局_失败 += 值 + ","
		}
	}
	if 局_成功 != "" {
		log += "追回积分成功(" + 局_成功 + "),"
	}
	if 局_失败 != "" {
		log += "追回积分失败(" + 局_成功 + "),"
	}

	局_成功 = ""
	局_失败 = ""
	for 值 := range 局_追回余额结果 {
		if 局_追回余额结果[值] == 1 {
			局_成功 += 值 + ","
		} else if 局_追回余额结果[值] == 2 {
			局_失败 += 值 + ","
		}
	}
	if 局_成功 != "" {
		log += "追回余额成功(" + 局_成功 + "),"
	}
	if 局_失败 != "" {
		log += "追回余额失败(" + 局_成功 + "),"
	}

	局_成功 = ""
	局_失败 = ""
	for 值 := range 局_追回推荐人时间点数结果 {
		if 局_追回推荐人时间点数结果[值] == 1 {
			局_成功 += 值 + ","
		} else if 局_追回推荐人时间点数结果[值] == 2 {
			局_失败 += 值 + ","
		}
	}
	if 局_成功 != "" {
		log += "追回推荐人" + 局_点数OR时间 + "成功(" + 局_成功 + "),"
	}
	if 局_失败 != "" {
		log += "追回推荐人" + 局_点数OR时间 + "失败(" + 局_成功 + "),"
	}

	//重置卡并冻结,删除信息
	global.GVA_DB.Model(DB.DB_Ka{}).Where("Id = ? ", 卡号详情卡号.Id).Updates(
		map[string]interface{}{
			"Status":     2,
			"User":       "",
			"Num":        0,
			"InviteUser": "",
			"UserTime":   "",
			"AdminNote":  卡号详情卡号.AdminNote + log,
		})

	return log, nil
}

type 结构请求_批量修改状态 struct {
	Id     []int `json:"Id"`     //用户id数组
	Status int   `json:"Status"` //1 解冻 2冻结
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
func Ka卡号取制卡人(卡号 string) string {
	var 制卡人 string
	global.GVA_DB.Model(DB.DB_Ka{}).Select("RegisterUser").Where("Name=?", 卡号).First(&制卡人)
	return 制卡人
}
func Id取制卡人(Id int) string {
	var 制卡人 string
	global.GVA_DB.Model(DB.DB_Ka{}).Select("RegisterUser").Where("Id=?", Id).First(&制卡人)
	return 制卡人
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

func K卡号充值_已废弃(来源AppId int, 卡号, 充值用户, 推荐人 string) (用户充值结果, 推荐人充值结果 error) {
	//因为无事务且为直接赋值字段,并发会有问题,已废弃仅供参考
	if !Ka校验卡号(卡号) { //节约数据库性能,
		return errors.New("卡号不存在"), nil
	}

	局_卡信息, err := Ka卡号取详情(卡号)
	if err != nil {
		return errors.New("卡号不存在"), nil
	}

	if 局_卡信息.Status == 2 {
		return errors.New("卡号已冻结,无法充值"), nil
	}
	if 局_卡信息.Num >= 局_卡信息.NumMax {
		//开启事务前检测一次,过滤降低数据库压力
		return errors.New("卡号已经使用到最大次数"), nil
	}
	if 来源AppId != 0 && 来源AppId != 局_卡信息.AppId {
		return errors.New("不是本应用卡号"), nil
	}

	if 充值用户 == 推荐人 {
		return errors.New("充值用户和推荐人不能相同"), nil
	}

	if 局_卡信息.KaType == 2 && utils.W文本_是否包含关键字(局_卡信息.User, 充值用户+",") {
		return errors.New("已使用本卡号充值过了,请勿重复充值"), nil
	}
	////1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	if Ser_AppInfo.App是否为卡号(局_卡信息.AppId) {
		return errors.New("卡号登录直接登录使用即可,无需充值"), nil
	}
	局_用户信息, ok := Ser_User.User取详情(充值用户)
	if !ok {
		return errors.New("用户不存在"), nil
	}
	if 推荐人 != "" && Ser_User.User用户名取id(推荐人) == 0 {
		return errors.New("推荐人账号不存在"), nil
	}

	if 局_用户信息.Status == 2 {
		return errors.New("用户已冻结,无法充值"), nil
	}
	//检测用户分组是否相同 不相同处理
	局_App用户, ok := Ser_AppUser.Uid取详情(局_卡信息.AppId, 局_用户信息.Id)
	if !ok {
		return errors.New("未注册应用,请先操作登录一次"), nil
	}
	if 局_卡信息.UserClassId == 局_App用户.UserClassId || 局_App用户.UserClassId == 0 {
		//分类相同,或用户为未分类 不处理
	} else {
		if 局_卡信息.NoUserClass == 2 {
			return errors.New("用户类型不同无法充值."), nil
		}
	}
	//到这里基本就都没问题了,开启事务,增加卡使用次数,更新用户信息就可以了
	// 开启事务
	tx := global.GVA_DB.Begin()
	//在事务中执行数据库操作，使用的是tx变量，不是db。

	//已用次数+1
	//RowsAffected用于返回sql执行后影响的行数
	m := map[string]interface{}{}
	m["Num"] = gorm.Expr("Num + 1")
	m["User"] = gorm.Expr("CONCAT(User,?)", 充值用户+",")
	m["UserTime"] = gorm.Expr("CONCAT(UserTime,?)", strconv.Itoa(int(time.Now().Unix()))+",")
	if 推荐人 != "" {
		m["InviteUser"] = gorm.Expr("CONCAT(InviteUser,?)", 推荐人+",")
	}

	rowsAffected := tx.Model(DB.DB_Ka{}).
		Where("Name = ?", 卡号).Where("Num<NumMax").Updates(&m).RowsAffected
	if rowsAffected == 0 {
		tx.Rollback() //失败回滚事务
		return errors.New("卡号已经使用到最大次数"), nil
	}
	//卡库存减少成功,开始增加客户数据 ,重新读取App用户信息,防止并发数据错误
	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Uid=?", 局_用户信息.Id).First(&局_App用户).Error
	if err != nil {
		tx.Rollback() //失败回滚事务
		return errors.New("未注册应用???感觉不可能,之前读取过,请联系管理员"), nil
	}
	//处理新信息
	局_App用户.VipNumber += 局_卡信息.VipNumber //积分不会变直接处理即可
	局_App用户.MaxOnline = 局_卡信息.MaxOnline  //最大在线数直接赋值处理即可
	局_现行时间戳 := time.Now().Unix()

	if 局_App用户.UserClassId == 局_卡信息.UserClassId {
		//分类相同,正常处理时间或点数
		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) || 局_App用户.VipTime > 局_现行时间戳 {
			//如果为计点 或 时间大于现在时间直接加就行了
			局_App用户.VipTime += 局_卡信息.VipTime
		} else {
			//如果为计时 已经过期很久了,直接现行时间戳加卡时间
			局_App用户.VipTime = 局_现行时间戳 + 局_卡信息.VipTime
		}

	} else {
		//用户类型不同, 根据权重处理

		局_旧用户类型权重 := Ser_UserClass.Get权重(局_App用户.UserClassId)
		局_新用户类型权重 := Ser_UserClass.Get权重(局_卡信息.UserClassId)

		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) {
			局_增减时间点数 := 局_App用户.VipTime * 局_旧用户类型权重 / 局_新用户类型权重 //转换结果值
			局_App用户.VipTime = 局_增减时间点数 + 局_卡信息.VipTime          //转回后再增加新类型 值
		} else {
			if 局_App用户.VipTime < 局_现行时间戳 {
				//已经过期了直接赋值新类型 现行时间+新时间就可以了
				局_App用户.VipTime = 局_现行时间戳 + 局_卡信息.VipTime
			} else {
				局_App用户.VipTime = 局_App用户.VipTime - 局_现行时间戳                 //先计算还剩多长时间
				局_增减时间点数 := 局_App用户.VipTime * 局_旧用户类型权重 / 局_新用户类型权重         //剩余时间 权重转换转换结果值
				局_App用户.VipTime = int64(局_现行时间戳) + 局_增减时间点数 + 局_卡信息.VipTime // 现在时间 + 旧权重转换后的新权重时间+卡增减时间
			}
		}
		局_App用户.UserClassId = 局_卡信息.UserClassId //最后更换类型,防止前面用到卡类id,计算权重转换类型错误
	}

	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Select("UserClassId", "VipTime", "VipNumber", "MaxOnline").Where("Id = ?", 局_App用户.Id).Updates(局_App用户).Error
	if err != nil {
		//有报错
		tx.Rollback() //失败回滚事务
		return errors.New("充值失败,重试"), nil
	}
	//用户的充值成功 提交事务
	tx.Commit()

	if 局_卡信息.InviteCount == 0 {
		return nil, nil //没有推广人直接返回成功就好了
	}

	//开始处理推广人用户
	局_推荐人Uid := Ser_User.User用户名取id(推荐人)

	if 局_推荐人Uid == 0 {
		return nil, errors.New("推荐人不存在") //没有推广人直接返回成功就好了
	}

	局_推荐人信息, ok := Ser_AppUser.Uid取详情(局_卡信息.AppId, 局_推荐人Uid)
	if !ok {
		return nil, errors.New("推荐人未使用本应用") //没有推广人直接返回成功就好了
	}

	if 局_推荐人信息.UserClassId == 局_卡信息.UserClassId {
		//分类相同,正常处理时间或点数
		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) || 局_推荐人信息.VipTime > 局_现行时间戳 {
			//如果为计点 或 时间大于现在时间直接加就行了
			局_推荐人信息.VipTime += 局_卡信息.VipTime
		} else {
			//如果为计时 已经过期很久了,直接现行时间戳加卡时间
			局_推荐人信息.VipTime = 局_现行时间戳 + 局_卡信息.VipTime
		}

	} else {
		//用户类型不同, 根据权重处理

		局_推荐人用户类型权重 := Ser_UserClass.Get权重(局_推荐人信息.UserClassId)
		局_新用户类型权重 := Ser_UserClass.Get权重(局_卡信息.UserClassId)

		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) {
			局_增减时间点数 := 局_卡信息.InviteCount * 局_新用户类型权重 / 局_推荐人用户类型权重 //转换结果值
			局_推荐人信息.VipTime = 局_增减时间点数 + 局_卡信息.InviteCount          //转回后再增加卡 推荐人值
		} else {
			if 局_推荐人信息.VipTime < 局_现行时间戳 {
				//已经过期了 现行时间+新时间就可以了
				局_推荐人信息.VipTime = 局_现行时间戳 + 局_卡信息.InviteCount
			} else {
				局_推荐人信息.VipTime = 局_推荐人信息.VipTime - 局_现行时间戳             //先计算还剩多长时间
				局_增减时间点数 := 局_卡信息.InviteCount * 局_新用户类型权重 / 局_推荐人用户类型权重 // 推荐人加点 权重转换
				局_推荐人信息.VipTime += 局_增减时间点数                             // 原值 + 推荐人加点
			}
		}
	}

	err = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Id = ?", 局_推荐人信息.Id).Update("VipTime=?", 局_推荐人信息.VipTime).Error
	if err != nil {
		return nil, err //没有推广人直接返回成功就好了
	}
	return nil, err //没有推广人直接返回成功就好了
}
func K卡号充值_事务(来源AppId int, 卡号, 充值用户, 推荐人, 来源IP string) (用户充值结果, 推荐人充值结果 error) {
	//已优化,事务处理,数据库内直接加减乘除计算字段值,可以并发,不出错
	if len(卡号) <= 2 || !Ka校验卡号(卡号) { //节约数据库性能,
		return errors.New("卡号不存在"), nil
	}

	局_卡信息, err := Ka卡号取详情(卡号)
	if err != nil {
		return errors.New("卡号不存在"), nil
	}

	if 局_卡信息.Status == 2 {
		return errors.New("卡号已冻结,无法充值"), nil
	}
	if 局_卡信息.Num >= 局_卡信息.NumMax {
		//开启事务前检测一次,过滤降低数据库压力
		return errors.New("卡号已经使用到最大次数"), nil
	}
	if 局_卡信息.EndTime != 0 && 局_卡信息.EndTime < time.Now().Unix() {
		//开启事务前检测一次,过滤降低数据库压力
		return errors.New("卡号已过有效期"), nil
	}
	if 来源AppId != 0 && 来源AppId != 局_卡信息.AppId {
		return errors.New("不是本应用卡号"), nil
	}

	if 充值用户 == 推荐人 {
		return errors.New("充值用户和推荐人不能相同"), nil
	}

	if 局_卡信息.KaType == 2 && utils.W文本_是否包含关键字(局_卡信息.User, 充值用户+",") {
		return errors.New("账号已使用本卡号充值过了,请勿重复充值"), nil
	}
	////1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	//2023-9-6 用户新需求 卡号充值卡号
	/*	if Ser_AppInfo.App是否为卡号(局_卡信息.AppId) {
		return errors.New("卡号登录直接登录使用即可,无需充值"), nil
	}*/
	局_is卡号 := Ser_AppInfo.App是否为卡号(局_卡信息.AppId)
	局_充值用户Uid := 0
	if !局_is卡号 {
		局_用户信息, ok := Ser_User.User取详情(充值用户)
		if !ok {
			return errors.New("用户不存在"), nil
		}
		if 局_用户信息.Status == 2 {
			return errors.New("用户已冻结,无法充值"), nil
		}
		局_充值用户Uid = 局_用户信息.Id
	} else {
		局_充值用户Uid = Ka卡号取id(局_卡信息.AppId, 充值用户)
	}
	局_推荐人用户Uid := 0
	if 推荐人 != "" {
		if 局_is卡号 {
			局_推荐人用户Uid = Ka卡号取id(局_卡信息.AppId, 推荐人)
			if 局_推荐人用户Uid == 0 {
				return errors.New("推荐人卡号不存在"), nil
			}
		} else {
			if Ser_User.User用户名取id(推荐人) == 0 {
				return errors.New("推荐人账号不存在"), nil
			}
		}
	}

	//检测用户分组是否相同 不相同处理
	局_App用户, ok := Ser_AppUser.Uid取详情(局_卡信息.AppId, 局_充值用户Uid)
	if !ok {
		return errors.New("用户未注册,请先操作登录一次应用:" + Ser_AppInfo.App取AppName(局_卡信息.AppId)), nil
	}
	if 局_卡信息.UserClassId == 局_App用户.UserClassId || 局_App用户.UserClassId == 0 {
		//分类相同,或用户为未分类 不处理
	} else {
		if 局_卡信息.NoUserClass == 2 {
			return errors.New("用户类型不同无法充值."), nil
		}
	}
	//到这里基本就都没问题了,开启事务,增加卡使用次数,更新用户信息就可以了
	// 开启事务
	tx := global.GVA_DB.Begin()
	//在事务中执行数据库操作，使用的是tx变量，不是db。

	//已用次数+1
	//RowsAffected用于返回sql执行后影响的行数
	m := map[string]interface{}{}
	m["Num"] = gorm.Expr("Num + 1")
	m["User"] = gorm.Expr("CONCAT(User,?)", 充值用户+",")
	m["UserTime"] = gorm.Expr("CONCAT(UserTime,?)", strconv.Itoa(int(time.Now().Unix()))+",")
	m["InviteUser"] = gorm.Expr("CONCAT(InviteUser,?)", 推荐人+",") //空推荐人也增加, 这样才能和用户充值顺序对应

	rowsAffected := tx.Model(DB.DB_Ka{}).
		Where("Name = ?", 卡号).Where("Num<NumMax").Updates(&m).RowsAffected
	if rowsAffected == 0 {
		tx.Rollback() //失败回滚事务
		return errors.New("卡号已经使用到最大次数"), nil
	}

	//卡库存减少成功,开始增加客户数据 ,重新读取App用户信息,防止并发数据错误
	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Uid=?", 局_App用户.Uid).First(&局_App用户).Error
	if err != nil {
		tx.Rollback() //失败回滚事务
		return errors.New("未注册应用???感觉不可能,之前读取过,请联系管理员"), nil
	}

	//处理新信息
	客户expr := map[string]interface{}{}
	客户expr["VipNumber"] = gorm.Expr("VipNumber + ?", 局_卡信息.VipNumber) //积分不会变直接增加即可
	客户expr["MaxOnline"] = 局_卡信息.MaxOnline                             //最大在线数直接赋值处理即可

	局_现行时间戳 := time.Now().Unix()

	if 局_App用户.UserClassId == 局_卡信息.UserClassId {
		//分类相同,正常处理时间或点数
		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) || 局_App用户.VipTime > 局_现行时间戳 {
			//如果为计点 或 时间大于现在时间直接加就行了
			客户expr["VipTime"] = gorm.Expr("VipTime + ?", 局_卡信息.VipTime)
		} else {
			//如果为计时 已经过期很久了,直接现行时间戳加卡时间
			客户expr["VipTime"] = 局_现行时间戳 + 局_卡信息.VipTime
		}

	} else {
		//用户类型不同, 根据权重处理
		局_旧用户类型权重 := Ser_UserClass.Get权重(局_App用户.UserClassId)
		局_新用户类型权重 := Ser_UserClass.Get权重(局_卡信息.UserClassId)

		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) {
			//转换结果值,转后再增加新类型 值
			客户expr["VipTime"] = gorm.Expr("VipTime * ? / ? +?", 局_旧用户类型权重, 局_新用户类型权重, 局_卡信息.VipTime)
		} else {
			if 局_App用户.VipTime < 局_现行时间戳 {
				//已经过期了直接赋值新类型 现行时间+新时间就可以了
				客户expr["VipTime"] = 局_现行时间戳 + 局_卡信息.VipTime
			} else {
				//先计算还剩多长时间,剩余时间权重转换转换结果值,+现在时间+卡增减时间
				客户expr["VipTime"] = gorm.Expr("(VipTime-?) * ? / ? +?", 局_现行时间戳, 局_旧用户类型权重, 局_新用户类型权重, 局_现行时间戳+局_卡信息.VipTime)
			}
		}
		//最后更换类型,防止前面用到卡类id,计算权重转换类型错误
		客户expr["UserClassId"] = 局_卡信息.UserClassId
	}

	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Id = ?", 局_App用户.Id).Updates(&客户expr).Error
	if err != nil {
		//有报错
		tx.Rollback() //失败回滚事务
		global.GVA_LOG.Error("充值失败,回滚事务,报错信息:" + err.Error())
		return errors.New("充值失败,重试"), nil
	}
	if !局_is卡号 && 局_卡信息.RMb > 0 {
		err = tx.Model(DB.DB_User{}).Where("Id = ?", 局_App用户.Uid).Update("RMB", gorm.Expr("RMB + ?", 局_卡信息.RMb)).Error
		if err != nil {
			tx.Rollback() //失败回滚事务
			global.GVA_LOG.Error("充值余额时失败,回滚事务,报错信息:" + err.Error())
			return errors.New("充值卡号余额时失败,请重试"), nil
		}
		var 局_新余额 float64
		_ = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", 局_App用户.Uid).First(&局_新余额).Error
		go Ser_Log.Log_写余额日志(充值用户, 来源IP, "使用卡号"+局_卡信息.Name+"充值余额|新余额≈"+utils.Float64到文本(局_新余额, 2), 局_卡信息.RMb)
	}

	//用户的充值成功 提交事务
	tx.Commit()

	if 局_卡信息.InviteCount == 0 {
		return nil, nil //没有推广人直接返回成功就好了
	}

	//开始处理推广人用户
	局_推荐人Uid := Ser_User.User用户名取id(推荐人)
	if 局_推荐人Uid == 0 {
		return nil, errors.New("推荐人不存在") //没有推广人直接返回成功就好了
	}

	局_推荐人信息, ok := Ser_AppUser.Uid取详情(局_卡信息.AppId, 局_推荐人Uid)
	if !ok {
		return nil, errors.New("推荐人未使用本应用") //没有推广人直接返回成功就好了
	}

	//处理新信息
	推荐人expr := map[string]interface{}{}

	if 局_推荐人信息.UserClassId == 局_卡信息.UserClassId {
		//分类相同,正常处理时间或点数
		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) || 局_推荐人信息.VipTime > 局_现行时间戳 {
			//如果为计点 或 时间大于现在时间直接加就行了
			推荐人expr["VipTime"] = gorm.Expr("VipTime + ?", 局_卡信息.InviteCount)
		} else {
			//如果为计时 已经过期很久了,直接现行时间戳加卡时间
			推荐人expr["VipTime"] = 局_现行时间戳 + 局_卡信息.InviteCount
		}

	} else {
		//用户类型不同, 根据权重处理

		局_推荐人用户类型权重 := Ser_UserClass.Get权重(局_推荐人信息.UserClassId)
		局_新用户类型权重 := Ser_UserClass.Get权重(局_卡信息.UserClassId)

		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) {
			//计算推荐人用户类型的实际点数,
			//这里有点绕,比如增加1秒,但是这个新用户类型权重为2, 荐人类型权重为1
			//那么实际新用户类型是推荐人类型的两倍,转换到推荐人类型值应该为2
			//所以 卡秒数+新用户类型权重=2,在除以推荐人用户类型权重=2
			局_增减时间点数 := 局_卡信息.InviteCount * 局_新用户类型权重 / 局_推荐人用户类型权重
			推荐人expr["VipTime"] = gorm.Expr("VipTime +?", 局_增减时间点数)
		} else {
			if 局_推荐人信息.VipTime < 局_现行时间戳 {
				//已经过期了 现行时间+新时间就可以了
				推荐人expr["VipTime"] = 局_现行时间戳 + 局_卡信息.InviteCount
			} else {
				局_增减时间点数 := 局_卡信息.InviteCount * 局_新用户类型权重 / 局_推荐人用户类型权重
				//原来的值+推荐人增加点数权重转换结果就好了
				推荐人expr["VipTime"] = gorm.Expr("VipTime+?", 局_增减时间点数)
			}
		}
	}

	err = global.GVA_DB.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Id = ?", 局_推荐人信息.Id).Updates(&推荐人expr).Error
	if err != nil {
		return nil, err //没有推广人直接返回成功就好了
	}
	return nil, err //没有推广人直接返回成功就好了
}
func K卡类直冲_事务(卡类ID, 软件用户id int, 来源IP string) error {
	//已优化,事务处理,数据库内直接加减乘除计算字段值,可以并发,不出错

	局_卡信息, err := Ser_KaClass.KaClass取详细信息(卡类ID)
	if err != nil {
		return errors.New("卡类不存在")
	}
	局_App用户, err := Ser_AppUser.Id取详情(局_卡信息.AppId, 软件用户id)
	if err != nil {
		return errors.New("软件用户不存在")
	}
	局_is卡号 := Ser_AppInfo.App是否为卡号(局_卡信息.AppId)
	//检测用户分组是否相同 不相同处理
	if 局_卡信息.UserClassId == 局_App用户.UserClassId || 局_App用户.UserClassId == 0 {
		//分类相同,或用户为未分类 不处理
	} else {
		if 局_卡信息.NoUserClass == 2 {
			return errors.New("用户类型不同无法充值.")
		}
	}
	//到这里基本就都没问题了,开启事务,增加卡使用次数,更新用户信息就可以了
	// 开启事务
	tx := global.GVA_DB.Begin()
	//在事务中执行数据库操作，使用的是tx变量，不是db。

	//卡库存减少成功,开始增加客户数据 ,重新读取App用户信息,防止并发数据错误
	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Uid=?", 局_App用户.Uid).First(&局_App用户).Error
	if err != nil {
		tx.Rollback() //失败回滚事务
		return errors.New("未注册应用???感觉不可能,之前读取过,请联系管理员")
	}

	//处理新信息
	客户expr := map[string]interface{}{}
	客户expr["VipNumber"] = gorm.Expr("VipNumber + ?", 局_卡信息.VipNumber) //积分不会变直接增加即可
	客户expr["MaxOnline"] = 局_卡信息.MaxOnline                             //最大在线数直接赋值处理即可

	局_现行时间戳 := time.Now().Unix()

	if 局_App用户.UserClassId == 局_卡信息.UserClassId {
		//分类相同,正常处理时间或点数
		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) || 局_App用户.VipTime > 局_现行时间戳 {
			//如果为计点 或 时间大于现在时间直接加就行了
			客户expr["VipTime"] = gorm.Expr("VipTime + ?", 局_卡信息.VipTime)
		} else {
			//如果为计时 已经过期很久了,直接现行时间戳加卡时间
			客户expr["VipTime"] = 局_现行时间戳 + 局_卡信息.VipTime
		}

	} else {
		//用户类型不同, 根据权重处理
		局_旧用户类型权重 := Ser_UserClass.Get权重(局_App用户.UserClassId)
		局_新用户类型权重 := Ser_UserClass.Get权重(局_卡信息.UserClassId)

		if Ser_AppInfo.App是否为计点(局_卡信息.AppId) {
			//转换结果值,转后再增加新类型 值
			客户expr["VipTime"] = gorm.Expr("VipTime * ? / ? +?", 局_旧用户类型权重, 局_新用户类型权重, 局_卡信息.VipTime)
		} else {
			if 局_App用户.VipTime < 局_现行时间戳 {
				//已经过期了直接赋值新类型 现行时间+新时间就可以了
				客户expr["VipTime"] = 局_现行时间戳 + 局_卡信息.VipTime
			} else {
				//先计算还剩多长时间,剩余时间权重转换转换结果值,+现在时间+卡增减时间
				客户expr["VipTime"] = gorm.Expr("(VipTime-?) * ? / ? +?", 局_现行时间戳, 局_旧用户类型权重, 局_新用户类型权重, 局_现行时间戳+局_卡信息.VipTime)
			}
		}
		//最后更换类型,防止前面用到卡类id,计算权重转换类型错误
		客户expr["UserClassId"] = 局_卡信息.UserClassId
	}

	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_卡信息.AppId)).Where("Id = ?", 局_App用户.Id).Updates(&客户expr).Error
	if err != nil {
		//有报错
		tx.Rollback() //失败回滚事务
		global.GVA_LOG.Error("充值失败,回滚事务,报错信息:" + err.Error())
		return errors.New("充值失败,重试")
	}
	if !局_is卡号 || 局_卡信息.RMb > 0 {
		err = tx.Model(DB.DB_User{}).Where("Id = ?", 局_App用户.Uid).Update("RMB", gorm.Expr("RMB + ?", 局_卡信息.RMb)).Error
		if err != nil {
			tx.Rollback() //失败回滚事务
			global.GVA_LOG.Error("卡类直冲余额时失败,回滚事务,报错信息:" + err.Error())
			return errors.New("卡类直冲余额时失败,请重试")
		}
		var 局_新余额 float64
		_ = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", 局_App用户.Uid).First(&局_新余额).Error
		go Ser_Log.Log_写余额日志(Ser_AppUser.Id取User(局_卡信息.AppId, 局_App用户.Uid), 来源IP, "购卡直冲应用ID:"+strconv.Itoa(局_卡信息.AppId)+"卡类Id:"+strconv.Itoa(局_卡信息.Id)+"充值余额|新余额≈"+utils.Float64到文本(局_新余额, 2), 局_卡信息.RMb)
	}
	//用户的充值成功 提交事务
	tx.Commit()
	return nil

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
func S删除耗尽次数卡号(AppId int) (影响行数 int64, err error) {
	db := global.GVA_DB.Model(DB.DB_Ka{})
	影响行数 = db.Where("Num = NumMax ").Where("AppId= ? ", AppId).Delete("").RowsAffected
	return 影响行数, err
}
