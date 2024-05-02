package ka

import (
	. "EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/Service/Ser_AppInfo"
	"server/global"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

var L_ka ka

func init() {
	L_ka = ka{}

}

type ka struct {
}

func (j *ka) K卡类直冲_事务(c *gin.Context, 卡类ID, 软件用户Uid int) (err error) {
	//已优化,事务处理,数据库内直接加减乘除计算字段值,可以并发,不出错
	var info struct {
		卡类详情     DB.DB_KaClass
		app用户详情  DB.DB_AppUser
		user用户详情 DB.DB_User
		app详情    DB.DB_AppInfo
		is卡号     bool
		is计点     bool
	}
	//第一个查询不用tx 直接用全局即可,后面事务的才用tx
	db := *global.GVA_DB
	S_KaClass := service.NewKaClass(c, &db)
	if info.卡类详情, err = S_KaClass.Info(卡类ID); err != nil {
		err = errors.New("卡类不存在")
		return
	}

	if info.app用户详情, err = service.NewAppUser(c, &db, info.卡类详情.AppId).InfoUid(软件用户Uid); err != nil {
		err = errors.New("软件用户不存在")
		return
	}
	if info.app详情, err = service.NewAppInfo(c, &db).Info(info.卡类详情.AppId); err != nil {
		err = errors.New("应用不存在")
		return
	}
	info.is卡号 = S三元(info.app详情.AppType == 3 || info.app详情.AppType == 4, true, false)
	info.is计点 = S三元(info.app详情.AppType == 2 || info.app详情.AppType == 4, true, false)

	//检测用户分组是否相同 不相同处理
	if info.卡类详情.UserClassId == info.app用户详情.UserClassId || info.app用户详情.UserClassId == 0 {
		//分类相同,或用户为未分类 不处理
	} else {
		if info.卡类详情.NoUserClass == 2 {
			return errors.New("用户类型不同无法充值.")
		}
	}
	//到这里基本就都没问题了,开启事务,增加卡使用次数,更新用户信息就可以了
	// 开启事务,检测上层是否有事务,如果有直接使用,没有就创建一个
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db = *global.GVA_DB
		tx = &db
	}
	//在事务中执行数据库操作，使用的是tx变量，不是db。
	err = tx.Transaction(func(tx *gorm.DB) error {
		//卡库存减少成功,开始增加客户数据 ,重新加锁读取App用户信息,防止并发数据错误
		err = tx.Model(DB.DB_AppUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Table("db_AppUser_"+strconv.Itoa(info.卡类详情.AppId)).Where("Uid=?", info.app用户详情.Uid).First(&info.app用户详情).Error
		if err != nil {
			return errors.Join(err, errors.New("未注册应用???感觉不可能,之前读取过,请联系管理员"))
		}
		//处理新信息
		客户expr := map[string]interface{}{}
		客户expr["VipNumber"] = gorm.Expr("VipNumber + ?", info.卡类详情.VipNumber) //积分不会变直接增加即可
		if info.卡类详情.MaxOnline > 0 {
			客户expr["MaxOnline"] = info.卡类详情.MaxOnline //最大在线数直接赋值处理即可
		}
		局_现行时间戳 := time.Now().Unix()
		if info.卡类详情.VipTime != 0 { //只有时间增减不为0的时候设置的用户分类才有效
			if info.app用户详情.UserClassId == info.卡类详情.UserClassId {
				//分类相同,正常处理时间或点数
				if Ser_AppInfo.App是否为计点(info.卡类详情.AppId) || info.app用户详情.VipTime > 局_现行时间戳 {
					//如果为计点 或 时间大于现在时间直接加就行了
					客户expr["VipTime"] = gorm.Expr("VipTime + ?", info.卡类详情.VipTime)
				} else {
					//如果为计时 已经过期很久了,直接现行时间戳加卡时间
					客户expr["VipTime"] = 局_现行时间戳 + info.卡类详情.VipTime
				}
			} else {
				//用户类型不同, 根据权重处理
				var 局_旧用户类型权重, 局_新用户类型权重 DB.DB_UserClass
				if 局_旧用户类型权重, err = service.NewUserClass(c, tx).Info(info.app用户详情.UserClassId); err != nil {
					return errors.Join(err, errors.New("读取旧用户类型权重失败"))
				}
				if 局_新用户类型权重, err = service.NewUserClass(c, tx).Info(info.卡类详情.UserClassId); err != nil {
					return errors.Join(err, errors.New("读取新用户类型权重失败"))
				}

				if info.is计点 {
					//转换结果值,转后再增加新类型 值
					客户expr["VipTime"] = gorm.Expr("VipTime * ? / ? +?", 局_旧用户类型权重.Weight, 局_新用户类型权重.Weight, info.卡类详情.VipTime)
				} else {
					if info.app用户详情.VipTime < 局_现行时间戳 {
						//已经过期了直接赋值新类型 现行时间+新时间就可以了
						客户expr["VipTime"] = 局_现行时间戳 + info.卡类详情.VipTime
					} else {
						//先计算还剩多长时间,剩余时间权重转换转换结果值,+现在时间+卡增减时间
						客户expr["VipTime"] = gorm.Expr("(VipTime-?) * ? / ? +?", 局_现行时间戳, 局_旧用户类型权重.Weight, 局_新用户类型权重.Weight, 局_现行时间戳+info.卡类详情.VipTime)
					}
				}
				//最后更换类型,防止前面用到卡类id,计算权重转换类型错误
				客户expr["UserClassId"] = info.卡类详情.UserClassId
			}
		}
		//更新客户数据
		err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡类详情.AppId)).Where("Id = ?", info.app用户详情.Id).Updates(&客户expr).Error
		if err != nil {
			return errors.Join(err, errors.New("充值失败,重试"))
		}
		//处理账号的RMB增减
		if !info.is卡号 && info.卡类详情.RMb > 0 {
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id=?", info.app用户详情.Uid).First(&info.user用户详情).Error
			if err != nil {
				return errors.Join(err, errors.New("用户账号不存在"))
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", info.app用户详情.Uid).Update("RMB", gorm.Expr("RMB + ?", info.卡类详情.RMb)).Error
			if err != nil {
				return errors.Join(err, errors.New("充值余额时失败,请重试"))
			}
			var 局_新余额 float64
			err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", info.user用户详情.Id).First(&局_新余额).Error
			if err != nil {
				return errors.Join(err, errors.New("充值后读取新余额失败"))
			}
			//日志仅写到上下文内,由实际业务处理是否写入日志和修改备注信息
			c.Set("logMoney", DB.DB_LogMoney{
				User:  info.user用户详情.User,
				Ip:    c.ClientIP(),
				Count: info.卡类详情.RMb,
				Note:  "应用ID:" + strconv.Itoa(info.卡类详情.AppId) + "卡类Id:" + strconv.Itoa(info.卡类详情.Id) + "充值余额|新余额≈" + Float64到文本(局_新余额, 2),
			})
		}
		return nil
	})
	//写到上下文,备用
	c.Set("info.卡类详情", info.卡类详情)
	c.Set("info.app用户详情", info.app用户详情)
	c.Set("info.user用户详情", info.user用户详情)
	c.Set("info.app详情", info.app详情)
	return err
}

// 有效期 0=9999999999 无限制
func (j *ka) Ka单卡创建(c *gin.Context, 卡类id int, 制卡人账号 string, 管理员备注 string, 代理备注 string, 有效期时间戳 int64) (卡信息切片 DB.DB_Ka, err error) {
	var info struct {
		卡类信息 DB.DB_KaClass
	}

	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}

	if info.卡类信息, err = service.NewKaClass(c, tx).Info(卡类id); err != nil { //估计是卡类不存在
		return 卡信息切片, err
	}

	for I := 0; I < 10; I++ {
		卡信息切片.Name = info.卡类信息.Prefix
		卡信息切片.Name += 生成随机字符串(info.卡类信息.KaLength-len(info.卡类信息.Prefix)-1, info.卡类信息.KaStringType)
		卡信息切片.Name += 生成校验字符(卡信息切片.Name)
		_, err2 := service.NewKa(c, tx).Info2(gin.H{"Name": 卡信息切片.Name})
		if err2 != nil { //如果有错误,说明没这卡号,可以使用
			break
		}
		if I == 9 {
			return 卡信息切片, errors.New("创建失败,连续10次没有随机到不重复卡号,请尝试删除无用卡号,再重新制卡")
		}
	}

	卡信息切片.AppId = info.卡类信息.AppId
	卡信息切片.KaClassId = info.卡类信息.Id
	卡信息切片.Status = 1
	卡信息切片.RegisterUser = 制卡人账号
	卡信息切片.RegisterTime = int(time.Now().Unix())
	卡信息切片.AdminNote = 管理员备注
	卡信息切片.AgentNote = 代理备注
	卡信息切片.VipTime = info.卡类信息.VipTime
	卡信息切片.InviteCount = info.卡类信息.InviteCount
	卡信息切片.RMb = info.卡类信息.RMb
	卡信息切片.VipNumber = info.卡类信息.VipNumber
	卡信息切片.Money = info.卡类信息.Money
	卡信息切片.AgentMoney = info.卡类信息.AgentMoney
	卡信息切片.UserClassId = info.卡类信息.UserClassId
	卡信息切片.NoUserClass = info.卡类信息.NoUserClass
	卡信息切片.KaType = info.卡类信息.KaType
	卡信息切片.MaxOnline = info.卡类信息.MaxOnline
	卡信息切片.Num = 0
	卡信息切片.NumMax = info.卡类信息.Num
	卡信息切片.User = ""
	卡信息切片.UserTime = ""
	卡信息切片.InviteUser = ""
	卡信息切片.EndTime = 9999999999
	if 有效期时间戳 != 0 {
		卡信息切片.EndTime = 有效期时间戳
	}
	return 卡信息切片, tx.Model(DB.DB_Ka{}).Create(&卡信息切片).Error
}
