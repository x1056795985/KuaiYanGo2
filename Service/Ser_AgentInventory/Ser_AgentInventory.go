package Ser_AgentInventory

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/kaClassUpPrice"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

// 只有管理员才会调用
// 库存卡包创建人ID 负数为管理员
// 资源包来源ID 直接购买为0 管理员下发为-1
func New(c *gin.Context, 归属Uid, KaClassId, NumMax, 库存卡包创建人ID, 资源包来源ID int, 有效期 int64, 备注 string) (DB.Db_Agent_库存卡包, error) {

	if !Ser_KaClass.KaClassId是否存在(KaClassId) {
		return DB.Db_Agent_库存卡包{}, errors.New("卡类ID不存在")
	}
	if agentLevel.L_agentLevel.Q取Id代理级别(c, 归属Uid) == 0 {
		return DB.Db_Agent_库存卡包{}, errors.New("代理ID不存在")
	}
	if NumMax <= 0 {
		return DB.Db_Agent_库存卡包{}, errors.New("库存卡包可使用次数必须大于0")
	}

	if 库存卡包创建人ID > 0 {
		return DB.Db_Agent_库存卡包{}, errors.New("只有管理员或开发者可以直接创建库存")
	}

	库存卡包 := DB.Db_Agent_库存卡包{
		Uid:            归属Uid,
		KaClassId:      KaClassId,
		Num:            0,
		NumMax:         NumMax,
		RegisterUserId: 库存卡包创建人ID,
		EndTime:        有效期,
		Note:           备注,
		SourceID:       资源包来源ID,
		SourceUid:      库存卡包创建人ID,
		StartTime:      time.Now().Unix(),
	}
	if 有效期 == 0 {
		库存卡包.EndTime = 9999999999
	}
	err := global.GVA_DB.Create(&库存卡包).Error
	return 库存卡包, err
}

func New代理购买(c *gin.Context, 归属Uid, KaClassId, NumMax int, 有效期 int64, 备注, ip string) (DB.Db_Agent_库存卡包, error) {

	局_卡类详情, err := Ser_KaClass.KaClass取详细信息(KaClassId)
	if err != nil {
		return DB.Db_Agent_库存卡包{}, errors.New("卡类ID不存在")
	}
	if 局_卡类详情.AgentMoney < 0 { //0元也可以购买
		return DB.Db_Agent_库存卡包{}, errors.New("卡类代理价格为负数,不可购买,请联系管理员")
	}
	if agentLevel.L_agentLevel.Q取Id代理级别(c, 归属Uid) == 0 {
		return DB.Db_Agent_库存卡包{}, errors.New("代理ID不存在")
	}
	if NumMax <= 0 {
		return DB.Db_Agent_库存卡包{}, errors.New("库存卡包可使用次数必须大于0")
	}

	可制卡号, _ := agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 归属Uid)
	if !S数组_整数是否存在(可制卡号, KaClassId) {
		return DB.Db_Agent_库存卡包{}, errors.New("权限不足,无法创建卡类ID:" + strconv.Itoa(KaClassId) + "的库存卡包,请联系上级代理授权该卡类")
	}
	库存卡包 := DB.Db_Agent_库存卡包{
		Uid:            归属Uid,
		KaClassId:      KaClassId,
		Num:            0,
		NumMax:         NumMax,
		RegisterUserId: 归属Uid,
		EndTime:        有效期,
		Note:           备注,
		SourceID:       0,
		SourceUid:      归属Uid,
		StartTime:      time.Now().Unix(),
	}
	if 有效期 == 0 {
		库存卡包.EndTime = 9999999999
	}

	var 局_价格组成 struct {
		总调价  float64
		调价详情 []dbm.DB_KaClassUpPrice
		购买数量 int64

		卡类金额 float64
		付款金额 float64
	}
	db := *global.GVA_DB
	var 局_代理详情 DB.DB_User
	局_代理详情, err = service.NewUser(c, &db).Info(归属Uid)
	if err != nil {
		return DB.Db_Agent_库存卡包{}, errors.Join(err, errors.New("取代理详情失败"))
	}

	局_价格组成.总调价, 局_价格组成.调价详情, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类详情.Id, 局_代理详情.UPAgentId)
	if err != nil {
		return DB.Db_Agent_库存卡包{}, errors.Join(err, errors.New("计算代理调价失败"))
	}
	局_价格组成.购买数量 = int64(NumMax)
	局_价格组成.总调价 = Float64乘int64(局_价格组成.总调价, 局_价格组成.购买数量)
	局_价格组成.卡类金额 = Float64乘int64(局_卡类详情.AgentMoney, 局_价格组成.购买数量)
	局_价格组成.付款金额 = Float64加float64(局_价格组成.卡类金额, 局_价格组成.总调价, 2)

	var 局_新余额 float64
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(DB.DB_User{}).Where("Id = ?", 归属Uid).Update("RMB", gorm.Expr("RMB - ?", 局_价格组成.付款金额)).Error
		if err != nil {
			return err
		}
		err = tx.Model(DB.DB_User{}).Select("RMB").Where("Id = ?", 归属Uid).Take(&局_新余额).Error
		if err != nil {
			return err
		}
		if 局_新余额 < 0 {
			return errors.New("余额不足,缺少:" + Float64到文本(局_新余额, 2))
		}
		//扣款成功,创建库存

		err = tx.Create(&库存卡包).Error
		return err
	})

	if err == nil {
		局_log := fmt.Sprintf("购买库存包ID:%d,代理价格(%v)*库存数量(%d)|新余额≈%v", 库存卡包.Id, 局_卡类详情.AgentMoney, NumMax, Float64到文本(局_新余额, 2))
		go Ser_Log.Log_写余额日志(Ser_User.Id取User(归属Uid), ip, 局_log, Float64取负值(局_价格组成.付款金额))
	} else {
		return 库存卡包, err
	}
	//代理分成 		//开始分利润 20240202 mark处理重构以后改事务
	//先分成 代理调价信息的价格
	if 局_价格组成.总调价 > 0 {
		局_日志前缀 := fmt.Sprintf("代理:%s,购买库存包ID{%d}", 局_代理详情.User, 库存卡包.Id)
		err = agent.L_agent.Z执行调价信息分成(c, 局_价格组成.调价详情, 局_价格组成.购买数量, 局_日志前缀)
		if err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("Z执行调价信息分成失败:", err.Error()))
		}
	}
	if 局_价格组成.卡类金额 > 0 {
		//然后再计算百分比的价格
		代理分成数据, err3 := agent.L_agent.D代理分成计算(c, 局_代理详情.Id, 局_价格组成.卡类金额)
		if err3 == nil {
			局_日志前缀 := fmt.Sprintf("代理:%s,购买库存包ID{%d}", 局_代理详情.User, 库存卡包.Id)
			err = agent.L_agent.Z执行百分比代理分成(c, 代理分成数据, 局_价格组成.卡类金额, 局_日志前缀)
			if err != nil {
				global.GVA_LOG.Error(fmt.Sprintf("Z执行百分比代理分成:%s", err.Error()))
			}
		}
	}
	// 分成结束==============

	return 库存卡包, err
}

func K库存发送(c *gin.Context, 原库存ID, 新代理Uid, 转出数量 int, 转出备注, IP string) error {

	原库存详情, ok := Id取详情(原库存ID)
	if !ok {
		return errors.New("库存ID不存在")
	}
	if 转出数量 <= 0 {
		return errors.New("转出数量必须大于0")
	}
	if 原库存详情.Uid != Ser_User.Id取上级代理ID(新代理Uid) {
		return errors.New("只能转出给自己下级代理")
	}

	if 原库存详情.NumMax-原库存详情.Num < 转出数量 {
		return errors.New("库存卡包可使用次数不足" + strconv.Itoa(转出数量))
	}

	if Ser_User.Id取状态(新代理Uid) == 2 {
		return errors.New("接收库存用户已冻结,不可发送")
	}

	err返回 := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 原库存ID).Update("Num", gorm.Expr("Num + ?", 转出数量)).Error
		if err != nil {
			return err
		}

		err = tx.Model(DB.Db_Agent_库存卡包{}).Where("Id=?", 原库存ID).First(&原库存详情).Error
		if err != nil {
			return err
		}
		if 原库存详情.Num > 原库存详情.NumMax {
			return errors.New("库存卡包可使用次数不足" + strconv.Itoa(转出数量))
		}
		//原库存,扣除成功,开始创建新库存ID

		新库存卡包 := DB.Db_Agent_库存卡包{
			Uid:            新代理Uid,
			KaClassId:      原库存详情.KaClassId,
			Num:            0,
			NumMax:         转出数量,
			RegisterUserId: 原库存详情.RegisterUserId,
			EndTime:        原库存详情.EndTime,
			Note:           转出备注,
			SourceID:       原库存详情.Id,
			SourceUid:      Id取归属Uid(原库存详情.Id),
			StartTime:      time.Now().Unix(),
		}
		err = tx.Create(&新库存卡包).Error
		if err == nil {
			var User1, User2 string
			User1 = Ser_User.Id取User(原库存详情.Uid)
			User2 = Ser_User.Id取User(新代理Uid)
			User1角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, 原库存详情.Uid)
			if User1角色 == 0 {
				User1角色 = 4
			}
			User2角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, 新代理Uid)
			if User2角色 == 0 {
				User2角色 = 4
			}
			Ser_Log.Log_写库存转移日志(原库存详情.Id, 转出数量, 1, User1, User1角色, User2, User2角色, IP, "发送到新库存ID:"+strconv.Itoa(新库存卡包.Id)+转出备注)
			Ser_Log.Log_写库存转移日志(新库存卡包.Id, 转出数量, 2, User2, User2角色, User1, User1角色, IP, "接收新库存,来自库存ID:"+strconv.Itoa(原库存详情.Id))
		}
		return err
	})
	return err返回
}
func K库存撤回(c *gin.Context, 操作UId, 原库存ID, 撤回数量 int, 备注, IP string) error {
	原库存详情, ok := Id取详情(原库存ID)
	if !ok {
		return errors.New("库存ID不存在")
	}
	if 撤回数量 <= 0 {
		return errors.New("撤回数量必须大于0")
	}

	//除非uid负数的管理员 否则,只有代理转出的库存,可以撤回
	if 操作UId > 0 && Id取归属Uid(原库存详情.SourceID) != 操作UId {
		//管理员可以撤回所有,代理只能撤回自己转出的库存
		return errors.New("只能撤回自己转出的库存")
	}

	if 原库存详情.NumMax-原库存详情.Num < 撤回数量 {
		return errors.New("库存卡包可使用次数不足" + strconv.Itoa(撤回数量))
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 原库存ID).Update("Num", gorm.Expr("Num + ?", 撤回数量)).Error
		if err != nil {
			return err
		}

		err = tx.Model(DB.Db_Agent_库存卡包{}).Where("Id=?", 原库存ID).First(&原库存详情).Error
		if err != nil {
			return err
		}
		if 原库存详情.Num > 原库存详情.NumMax {
			return errors.New("库存卡包可使用次数不足" + strconv.Itoa(撤回数量))
		}
		//下级库存,已使用次数增加成功,开始减少上级库存已使用次数
		局_来源库存Id是否存在 := Id是否存在(原库存详情.SourceID)
		if 原库存详情.SourceID < 0 || !局_来源库存Id是否存在 {
			//如果是管理员直接成功,不用实际撤回
			err = nil
		} else {
			err = tx.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 原库存详情.SourceID).Update("Num", gorm.Expr("Num - ?", 撤回数量)).Error
		}
		if err == nil {

			var User1, User2 string

			User1 = Ser_User.Id取User(原库存详情.Uid)

			局_User2Id := 操作UId
			if 原库存详情.SourceID > 0 {
				//如果来源id小于0 那么就是管理员,Uid就是管理员Uid
				局_User2Id = Id取归属Uid(原库存详情.SourceID)
			}
			User2 = Ser_User.Id取User(局_User2Id)

			User1角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, 原库存详情.Uid)
			if User1角色 == 0 {
				User1角色 = 4
			}
			User2角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, 局_User2Id)
			if User2角色 == 0 {
				User2角色 = 4
			}

			局_msg := fmt.Sprintf("被撤回库存到上级库存ID:%d,原因:%s", 原库存详情.SourceID, 备注)
			if 原库存详情.SourceID < 0 || 局_来源库存Id是否存在 {
				局_msg = fmt.Sprintf("被撤回库存到上级库存ID:%d(负数管理员库存),原因:%s", 原库存详情.SourceID, 备注)
			} else if !局_来源库存Id是否存在 {
				局_msg = fmt.Sprintf("被撤回库存到上级库存ID:%d(已删除),原因:%s", 原库存详情.SourceID, 备注)
			}

			Ser_Log.Log_写库存转移日志(原库存详情.Id, 撤回数量, 1, User1, User1角色, User2, User2角色, IP, 局_msg)
			if 原库存详情.SourceID > 0 {
				//写转入日志
				Ser_Log.Log_写库存转移日志(原库存详情.SourceID, 撤回数量, 2, User2, User2角色, User1, User1角色, IP, "来自已撤回下级库存ID:"+strconv.Itoa(原库存详情.Id)+",原因:"+备注)
			} else {
				//如果是管理员不用写日志,因为也没转入到管理员库存ID
			}

		}

		return err
	})
}
func Id取详情(Id int) (库存卡包详情 DB.Db_Agent_库存卡包, ok bool) {
	err := global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Where("Id=?", Id).First(&库存卡包详情).Error
	return 库存卡包详情, err == nil
}

func Id取归属Uid(Id int) int {
	if Id == 0 {
		return 0
	}

	Uid := 0
	_ = global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Select("Uid").Where("Id=?", Id).First(&Uid).Error
	return Uid
}
func Id是否存在(Id int) bool {
	var Count int64
	result := global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Select("1").Where("Id=?", Id).First(&Count)
	return result.Error == nil
}

func K库存延期(库存ID, 代理Uid, 延期秒数 int) error {

	原库存详情, ok := Id取详情(库存ID)
	if !ok {
		return errors.New("库存ID不存在")
	}

	if 代理Uid != 原库存详情.RegisterUserId {
		return errors.New("只能库存原始购买人可修改过期时间")
	}
	if 原库存详情.EndTime == 9999999999 {
		return errors.New("有效期无限制不可修改")
	}
	var err error
	if 延期秒数 > 9999999999 {
		err = global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 库存ID).Update("EndTime", 9999999999).Error
	} else {
		err = global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 库存ID).Update("EndTime", gorm.Expr("EndTime + ?", 延期秒数)).Error
	}

	return err
}

func K库存修改备注(库存ID, 代理Uid int, 新备注 string) error {

	原库存详情, ok := Id取详情(库存ID)
	if !ok {
		return errors.New("库存ID不存在")
	}

	if 代理Uid != 原库存详情.Uid {
		return errors.New("只能修改归属自己的库存")
	}

	err := global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Where("Id = ?", 库存ID).Update("Note", 新备注).Error

	return err
}
