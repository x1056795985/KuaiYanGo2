package ka

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/global"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/kaClassUpPrice"
	"server/new/app/logic/common/log"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"strconv"
	"strings"
	"time"
)

// Ka代理批量库存购买 使用库存卡包批量购买卡号
func (j *ka) Ka代理批量库存购买(c *gin.Context, 卡信息切片 []dbm.DB_Ka, 库存Id, 制卡数量, 购卡人Id int, 代理备注 string, ip string) error {
	局_db := *global.GVA_DB
	if 制卡数量 <= 0 {
		return errors.New("生成数量必须大于0")
	}
	if 制卡数量 > 2621 {
		return errors.New("生成数量每批最大2621")
	}
	局_库存详情, ok := service.NewAgentInventory(c, &局_db).Id取详情(库存Id)
	if !ok {
		return errors.New("库存ID不存在")
	}
	if 局_库存详情.Uid != 购卡人Id {
		return errors.New("只能使用归属自己的库存制卡")
	}
	if 局_库存详情.NumMax-局_库存详情.Num < 制卡数量 {
		return errors.New("库存剩余可制卡次数不足")
	}
	KaClass详细信息, err := service.NewKaClass(c, &局_db).KaClass取详细信息(局_库存详情.KaClassId)
	if err != nil {
		return errors.New("库存所属卡类id不存在,可能已被管理员删除")
	}
	局_购卡人User := service.NewUser(c, &局_db).Id取User(购卡人Id)

	type 卡号记录 struct {
		Index   int
		Name    string
		AutoGen bool
	}
	var 待检查卡号 []卡号记录
	for i := range 卡信息切片 {
		if 卡信息切片[i].Name == "" {
			name := KaClass详细信息.Prefix + 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
			卡信息切片[i].Name = name
			待检查卡号 = append(待检查卡号, 卡号记录{Index: i, Name: name, AutoGen: true})
		} else {
			待检查卡号 = append(待检查卡号, 卡号记录{Index: i, Name: 卡信息切片[i].Name, AutoGen: false})
		}
	}
	nameSet := make(map[string]struct{})
	for _, rec := range 待检查卡号 {
		if _, ok := nameSet[rec.Name]; ok {
			return errors.New("本批次内卡号重复: " + rec.Name)
		}
		nameSet[rec.Name] = struct{}{}
	}

	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(dbm.Db_Agent_库存卡包{}).Where("Id = ?", 局_库存详情.Id).Update("Num", gorm.Expr("Num + ?", 制卡数量)).Error
		if err != nil {
			return err
		}
		var 剩余库存 int
		err = tx.Model(dbm.Db_Agent_库存卡包{}).Select("NumMax-Num").Where("Id = ?", 局_库存详情.Id).Take(&剩余库存).Error
		if err != nil {
			return err
		}
		if 剩余库存 < 0 {
			return errors.New("库存可用次数不足,缺少次数:" + strconv.Itoa(-剩余库存))
		}

		maxRetries := 10
		for retry := 0; retry <= maxRetries; retry++ {
			nameList := make([]string, len(待检查卡号))
			for i, rec := range 待检查卡号 {
				nameList[i] = rec.Name
			}
			var existingNames []string
			err = tx.Model(dbm.DB_Ka{}).Select("Name").Where("Name IN (?)", nameList).Find(&existingNames).Error
			if err != nil {
				return err
			}
			existingSet := make(map[string]bool)
			for _, n := range existingNames {
				existingSet[n] = true
			}
			var conflictIndices []int
			for i, rec := range 待检查卡号 {
				if existingSet[rec.Name] {
					if rec.AutoGen {
						conflictIndices = append(conflictIndices, i)
					} else {
						return errors.New("卡号:" + rec.Name + "已存在无法使用")
					}
				}
			}
			if len(conflictIndices) == 0 {
				break
			}
			if retry == maxRetries {
				return errors.New("创建失败,连续多次随机到重复卡号,请尝试删除无用卡号,再重新制卡")
			}
			for _, idx := range conflictIndices {
				rec := &待检查卡号[idx]
				newName := KaClass详细信息.Prefix + 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
				rec.Name = newName
				卡信息切片[rec.Index].Name = newName
			}
			nameSet = make(map[string]struct{})
			for _, rec := range 待检查卡号 {
				if _, ok := nameSet[rec.Name]; ok {
					return errors.New("本批次内卡号重复: " + rec.Name)
				}
				nameSet[rec.Name] = struct{}{}
			}
		}

		for i := range 卡信息切片 {
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
		err = tx.Model(dbm.DB_Ka{}).Create(&卡信息切片).Error
		return err
	})
	if err != nil {
		return err
	}
	数组_卡号 := make([]string, 0, len(卡信息切片))
	for i := 0; i < len(卡信息切片); i++ {
		数组_卡号 = append(数组_卡号, 卡信息切片[i].Name)
	}
	局_文本 := fmt.Sprintf("制卡库存Id:%d,应用:%s,卡类:%s,批次id:{{批次id}}({{卡号索引}}/%d)", 局_库存详情.Id, service.NewAppInfo(c, &局_db).App取AppName(卡信息切片[0].AppId), service.NewKaClass(c, &局_db).Id取Name(卡信息切片[0].KaClassId), len(卡信息切片))
	go log.L_log.Log_写卡号操作日志(局_购卡人User, ip, 局_文本, 数组_卡号, 1, agentLevel.L_agentLevel.Q取Id代理级别(c, 购卡人Id))
	return nil
}

// Ka代理批量购买 代理批量购买卡号
func (j *ka) Ka代理批量购买(c *gin.Context, 卡信息切片 []dbm.DB_Ka, 卡类id, 购卡人Id int, 代理备注 string, 有效期时间戳 int64, ip string) error {
	局_db := *global.GVA_DB
	var 局_价格组成 struct {
		总卡类价格 float64
		总调价    float64
		调价详情  []dbm.DB_KaClassUpPrice
		购买数量  int64
		总付款金额 float64
	}
	if len(卡信息切片) >= 2621 {
		return errors.New("每批次最大数量不能超过2621")
	}
	局_价格组成.购买数量 = int64(len(卡信息切片))

	KaClass详细信息, err := service.NewKaClass(c, &局_db).KaClass取详细信息(卡类id)
	if err != nil {
		return err
	}
	局_价格组成.总卡类价格 = Float64乘int64(KaClass详细信息.AgentMoney, 局_价格组成.购买数量)

	局_购卡人信息, ok := service.NewUser(c, &局_db).Id取详情(购卡人Id)
	if !ok {
		return errors.New("用户不存在")
	}

	局_价格组成.总调价, 局_价格组成.调价详情, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 卡类id, 局_购卡人信息.UPAgentId)
	if err != nil {
		return err
	}
	局_价格组成.总调价 = Float64乘int64(局_价格组成.总调价, 局_价格组成.购买数量)
	局_价格组成.总付款金额 = Float64加float64(局_价格组成.总调价, 局_价格组成.总卡类价格, 2)

	if 局_购卡人信息.Rmb < 局_价格组成.总付款金额 {
		return fmt.Errorf("余额不足 (当前余额:%.2f < 需支付:%.2f)", 局_购卡人信息.Rmb, 局_价格组成.总付款金额)
	}
	if 局_价格组成.总付款金额 < 0 {
		return errors.New("卡类代理价格异常")
	}

	var 新余额 float64
	type 卡号记录 struct {
		Index   int
		Name    string
		AutoGen bool
	}
	var 待检查卡号 []卡号记录
	for i := range 卡信息切片 {
		if 卡信息切片[i].Name == "" {
			name := KaClass详细信息.Prefix + 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
			卡信息切片[i].Name = name
			待检查卡号 = append(待检查卡号, 卡号记录{Index: i, Name: name, AutoGen: true})
		} else {
			待检查卡号 = append(待检查卡号, 卡号记录{Index: i, Name: 卡信息切片[i].Name, AutoGen: false})
		}
	}
	nameSet := make(map[string]struct{})
	for _, rec := range 待检查卡号 {
		if _, ok := nameSet[rec.Name]; ok {
			return errors.New("本批次内卡号重复: " + rec.Name)
		}
		nameSet[rec.Name] = struct{}{}
	}

	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Exec("UPDATE db_User SET RMB = RMB - ? WHERE Id = ?", 局_价格组成.总付款金额, 局_购卡人信息.Id).Error
		if err != nil {
			global.GVA_LOG.Println(strconv.Itoa(局_购卡人信息.Id) + "Id余额减少失败:" + err.Error())
			return errors.New("余额减少失败查看服务器日志检查原因")
		}
		err = tx.Raw("SELECT RMB FROM db_User WHERE Id = ?", 局_购卡人信息.Id).Scan(&新余额).Error
		if err != nil {
			global.GVA_LOG.Println(strconv.Itoa(局_购卡人信息.Id) + "Id查询余额失败:" + err.Error())
			return errors.New("查询余额失败查看服务器日志检查原因")
		}
		if 新余额 < 0 {
			return errors.New("用户余额不足,缺少:" + Float64到文本(Float64取绝对值(新余额), 2))
		}

		maxRetries := 10
		for retry := 0; retry <= maxRetries; retry++ {
			nameList := make([]string, len(待检查卡号))
			for i, rec := range 待检查卡号 {
				nameList[i] = rec.Name
			}
			var existingNames []string
			err = tx.Model(dbm.DB_Ka{}).Select("Name").Where("Name IN (?)", nameList).Find(&existingNames).Error
			if err != nil {
				return err
			}
			existingSet := make(map[string]bool)
			for _, n := range existingNames {
				existingSet[n] = true
			}
			var conflictIndices []int
			for i, rec := range 待检查卡号 {
				if existingSet[rec.Name] {
					if rec.AutoGen {
						conflictIndices = append(conflictIndices, i)
					} else {
						return errors.New("卡号:" + rec.Name + "已存在无法使用")
					}
				}
			}
			if len(conflictIndices) == 0 {
				break
			}
			if retry == maxRetries {
				return errors.New("创建失败,连续多次随机到重复卡号,请尝试删除无用卡号,再重新制卡")
			}
			for _, idx := range conflictIndices {
				rec := &待检查卡号[idx]
				newName := KaClass详细信息.Prefix + 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
				rec.Name = newName
				卡信息切片[rec.Index].Name = newName
			}
			nameSet = make(map[string]struct{})
			for _, rec := range 待检查卡号 {
				if _, ok := nameSet[rec.Name]; ok {
					return errors.New("本批次内卡号重复: " + rec.Name)
				}
				nameSet[rec.Name] = struct{}{}
			}
		}

		for i := range 卡信息切片 {
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
		err = tx.Model(dbm.DB_Ka{}).Create(&卡信息切片).Error
		return err
	})
	if err != nil {
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
	局_文本 := fmt.Sprintf("代理购卡[%s -> %s],卡号ID{%s},|新余额≈%s", service.NewAppInfo(c, &局_db).App取AppName(KaClass详细信息.AppId), KaClass详细信息.Name, 局_ID列表, Float64到文本(新余额, 2))
	log.L_log.Log_写余额日志(局_购卡人信息.User, ip, 局_文本, Float64取负值(局_价格组成.总付款金额))
	局_文本 = fmt.Sprintf("新制卡号:[%s -> %s],批次id:{{批次id}}({{卡号索引}}/%d)", service.NewAppInfo(c, &局_db).App取AppName(卡信息切片[0].AppId), service.NewKaClass(c, &局_db).Id取Name(卡信息切片[0].KaClassId), len(卡信息切片))
	log.L_log.Log_写卡号操作日志(局_购卡人信息.User, ip, 局_文本, 数组_卡号, 1, agentLevel.L_agentLevel.Q取Id代理级别(c, 局_购卡人信息.Id))

	return nil
}

// Ka更换卡号 更换卡号
func (j *ka) Ka更换卡号(c *gin.Context, id, 代理Id int, ip string) error {
	局_db := *global.GVA_DB
	局_卡号详情, err := service.NewKa(c, &局_db).Id取详情(id)
	if err != nil {
		return errors.New("卡号ID不存在")
	}
	if 局_卡号详情.Num != 0 {
		return errors.New("卡号已使用无法更换")
	}
	if service.NewAppInfo(c, &局_db).App是否为卡号(局_卡号详情.AppId) {
		return errors.New("应用为卡号登录模式,无法更改卡号")
	}
	代理User := service.NewUser(c, &局_db).Id取User(代理Id)
	if 局_卡号详情.RegisterUser != 代理User {
		return errors.New("只有自己的卡才可以更换卡号")
	}
	if 局_卡号详情.Status != 1 {
		return errors.New("卡号已冻结,暂不可更换卡号")
	}
	KaClass详细信息, err := service.NewKaClass(c, &局_db).KaClass取详细信息(局_卡号详情.KaClassId)
	if err != nil {
		return errors.New("卡号对应卡类ID不存在,可能已删除")
	}

	var 局_新卡号 = ""
	for I := 0; I < 10; I++ {
		局_新卡号 = KaClass详细信息.Prefix
		局_新卡号 += 生成随机字符串(KaClass详细信息.KaLength-len(KaClass详细信息.Prefix), KaClass详细信息.KaStringType)
		if !j.Ka卡号是否存在(局_新卡号) {
			break
		}
		if I == 9 {
			return errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
		}
	}

	err = global.GVA_DB.Model(dbm.DB_Ka{}).Where("Id = ? ", id).Update("Name", 局_新卡号).Error
	if err == nil {
		局_log := fmt.Sprintf("操作更换卡号:  %s  ->  %s", 局_卡号详情.Name, 局_新卡号)
		log.L_log.Log_写卡号操作日志(代理User, ip, 局_log, []string{局_卡号详情.Name}, 3, agentLevel.L_agentLevel.Q取Id代理级别(c, 代理Id))
	}
	return err
}

// Ka卡号是否存在 卡号是否存在
func (j *ka) Ka卡号是否存在(卡号 string) bool {
	var Count int64
	_ = global.GVA_DB.Select("1").Model(dbm.DB_Ka{}).Where("Name=?", 卡号).First(&Count)
	return Count != 0
}
