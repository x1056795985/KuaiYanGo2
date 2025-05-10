package ka

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/log"
	"server/new/app/logic/common/userClass"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
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
		卡类详情     dbm.DB_KaClass
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
				if info.app用户详情.UserClassId > 0 {
					if 局_旧用户类型权重, err = service.NewUserClass(c, tx).Info(info.app用户详情.UserClassId); err != nil {
						return errors.Join(err, errors.New("读取旧用户类型权重失败"))
					}
				} else {
					局_旧用户类型权重.Weight = 1
				}

				if info.卡类详情.UserClassId > 0 {
					if 局_新用户类型权重, err = service.NewUserClass(c, tx).Info(info.卡类详情.UserClassId); err != nil {
						return errors.Join(err, errors.New("读取新用户类型权重失败"))
					}
				} else {
					局_新用户类型权重.Weight = 1
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
		卡类信息 dbm.DB_KaClass
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
	卡信息切片.RegisterTime = time.Now().Unix()
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
	卡信息切片.UseTime = 0
	卡信息切片.InviteUser = ""
	卡信息切片.EndTime = 9999999999
	if 有效期时间戳 != 0 {
		卡信息切片.EndTime = 有效期时间戳
	}
	return 卡信息切片, tx.Model(DB.DB_Ka{}).Create(&卡信息切片).Error
}

// 来源AppId int, 卡号, 充值用户, 推荐人, 来源IP string)
func (j *ka) K卡号充值_事务(c *gin.Context, 来源AppId int, 卡号, 充值用户, 推荐人 string) (err error) {
	var info struct {
		卡号详情     DB.DB_Ka
		app用户详情  DB.DB_AppUser
		user用户详情 DB.DB_User
		ka用户详情   DB.DB_Ka
		app详情    DB.DB_AppInfo
		is卡号     bool
		is计点     bool
		ip       string

		app用户详情_推荐人  DB.DB_AppUser
		user用户详情_推荐人 DB.DB_User
		ka用户详情_推荐人   DB.DB_Ka
		用户类型_推荐人     DB.DB_UserClass
		新用户类型_推荐人    DB.DB_UserClass

		logMoney     []DB.DB_LogMoney     //余额日志
		logVipNumber []DB.DB_LogVipNumber //积分,点数日志

	}

	if c == nil || c.Request == nil {
		info.ip = "未知" // 处理 `gin.Context` 为 `nil` 的情况
	} else {
		info.ip = c.ClientIP() //有可能报错 如果c 不是外部请求过来的
	}
	//第一个查询不用tx 直接用全局即可,后面事务的才用tx
	db := *global.GVA_DB
	S_Ka := service.NewKa(c, &db)
	if info.卡号详情, err = S_Ka.InfoKa(卡号); err != nil {
		err = errors.New("卡号不存在")
		return
	}
	if info.卡号详情.Status == 2 {
		return errors.New("卡号已冻结,无法充值")
	}
	if info.卡号详情.Num >= info.卡号详情.NumMax {
		//开启事务前检测一次,过滤降低数据库压力
		return errors.New("卡号已经使用到最大次数")
	}
	if info.卡号详情.EndTime != 0 && info.卡号详情.EndTime < time.Now().Unix() {
		//开启事务前检测一次,过滤降低数据库压力
		return errors.New("卡号已过有效期")
	}
	if 来源AppId != 0 && 来源AppId != info.卡号详情.AppId {
		return errors.New("不是本应用卡号")
	}
	if 充值用户 == 推荐人 {
		return errors.New("充值用户和推荐人不能相同")
	}
	if info.卡号详情.KaType == 2 && W文本_是否包含关键字(info.卡号详情.User, 充值用户+",") {
		return errors.New("账号已使用本卡号充值过了,请勿重复充值")
	}
	if info.app详情, err = service.NewAppInfo(c, &db).Info(info.卡号详情.AppId); err != nil {
		err = errors.New("应用不存在")
		return
	}
	info.is卡号 = S三元(info.app详情.AppType == 3 || info.app详情.AppType == 4, true, false)
	info.is计点 = S三元(info.app详情.AppType == 2 || info.app详情.AppType == 4, true, false)

	if !info.is卡号 {
		info.user用户详情, err = service.NewUser(c, &db).InfoName(充值用户)
		if err != nil {
			return errors.New("用户不存在")
		}
		if info.ka用户详情.Status == 2 {
			return errors.New("用户已冻结,无法充值")
		}
		info.app用户详情, err = service.NewAppUser(c, &db, info.卡号详情.AppId).InfoUid(info.user用户详情.Id)
	} else {
		info.ka用户详情, err = service.NewKa(c, &db).InfoKa(充值用户)
		if err != nil {
			return errors.New("用户不存在")
		}
		if info.ka用户详情.Status == 2 {
			return errors.New("用户已冻结,无法充值")
		}
		info.app用户详情, err = service.NewAppUser(c, &db, info.卡号详情.AppId).InfoUid(info.ka用户详情.Id)
	}
	if err != nil {
		return errors.New("用户未登录过本应用,请先操作登录")
	}
	if info.app用户详情.Status == 2 {
		return errors.New("app用户已冻结,无法充值")
	}

	if 推荐人 != "" {
		if !info.is卡号 {
			info.user用户详情_推荐人, err = service.NewUser(c, &db).InfoName(推荐人)
			if err != nil {
				return errors.New("推荐人用户不存在")
			}
			if info.user用户详情_推荐人.Status == 2 {
				return errors.New("推荐人用户已冻结,无法充值")
			}
			info.app用户详情_推荐人, err = service.NewAppUser(c, &db, info.卡号详情.AppId).InfoUid(info.user用户详情_推荐人.Id)
		} else {
			info.ka用户详情_推荐人, err = service.NewKa(c, &db).InfoKa(推荐人)
			if err != nil {
				return errors.New("推荐人用户不存在")
			}
			if info.ka用户详情_推荐人.Status == 2 {
				return errors.New("推荐人用户已冻结,无法充值")
			}
			info.app用户详情_推荐人, err = service.NewAppUser(c, &db, info.卡号详情.AppId).InfoUid(info.ka用户详情_推荐人.Id)
		}
		if err != nil {
			return errors.New("推荐人用户未登录过本应用,请先操作登录")
		}
		if info.app用户详情_推荐人.Status == 2 {
			return errors.New("推荐人app用户已冻结,无法充值")
		}
	}
	//检测用户分组是否相同 不相同处理
	if info.卡号详情.UserClassId == info.app用户详情.UserClassId || info.app用户详情.UserClassId == 0 {
		//分类相同,或用户为未分类 不处理
	} else {
		if info.卡号详情.NoUserClass == 2 {
			return errors.New("用户类型不同无法充值.")
		}
	}
	//到这里基本就都没问题了,开启事务,增加卡使用次数,更新用户信息就可以了
	// 开启事务,检测上层是否有事务,如果有直接使用,没有就创建一个
	var db2 *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		db2 = tempObj.(*gorm.DB)
	} else {
		db = *global.GVA_DB
		db2 = &db
	}
	//在事务中执行数据库操作，使用的是tx变量，不是db。
	err = db2.Transaction(func(tx *gorm.DB) error {
		//加锁重新读卡信息
		err = tx.Model(DB.DB_Ka{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", info.卡号详情.Id).First(&info.卡号详情).Error
		if err != nil {
			return errors.Join(err, errors.New("卡号不存在"))
		}
		if info.卡号详情.Num >= info.卡号详情.NumMax {
			return errors.Join(err, errors.New("卡号已经使用到最大次数"))
		}
		//已用次数+1
		//RowsAffected用于返回sql执行后影响的行数
		m := map[string]interface{}{}
		m["Num"] = gorm.Expr("Num + 1")
		m["User"] = gorm.Expr("CONCAT(User,?)", 充值用户+",")
		m["UserTime"] = gorm.Expr("CONCAT(UserTime,?)", strconv.Itoa(int(time.Now().Unix()))+",")
		m["UseTime"] = time.Now().Unix()
		m["InviteUser"] = gorm.Expr("CONCAT(InviteUser,?)", 推荐人+",") //空推荐人也增加, 这样才能和用户充值顺序对应
		rowsAffected := tx.Model(DB.DB_Ka{}).
			Where("Id = ?", info.卡号详情.Id).Updates(&m).RowsAffected
		if rowsAffected == 0 {
			return errors.New("卡号不存在或已经使用到最大次数")
		}
		//卡库存减少成功,开始增加客户数据 ,重新加锁读取App用户信息,防止并发数据错误
		err = tx.Model(DB.DB_AppUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).Where("Uid=?", info.app用户详情.Uid).First(&info.app用户详情).Error
		if err != nil {
			return errors.Join(err, errors.New("未注册应用???感觉不可能,之前读取过,请联系管理员"))
		}
		//处理新信息
		客户expr := map[string]interface{}{}
		if info.卡号详情.VipNumber > 0 {
			客户expr["VipNumber"] = gorm.Expr("VipNumber + ?", info.卡号详情.VipNumber) //积分不会变直接增加即可
			info.logVipNumber = append(info.logVipNumber, DB.DB_LogVipNumber{
				AppId: info.卡号详情.AppId,
				Count: info.卡号详情.VipNumber,
				Ip:    info.ip,
				Note:  "应用ID:" + strconv.Itoa(info.卡号详情.AppId) + "卡号Id:" + strconv.Itoa(info.卡号详情.Id) + "充值积分",
				Time:  time.Now().Unix(),
				Type:  1,
				User:  充值用户,
			})
		}

		if info.卡号详情.MaxOnline > 0 {
			客户expr["MaxOnline"] = info.卡号详情.MaxOnline //最大在线数直接赋值处理即可
		}

		//卡号充值时,如果充值的用户没有归属代理,则自动设置该卡号的制卡人为归属代理
		if info.app用户详情.AgentUid == 0 {
			局_制卡人uid := Ser_User.User用户名取id(info.卡号详情.RegisterUser)
			if 局_制卡人uid >= 0 {
				客户expr["AgentUid"] = 局_制卡人uid
			}
		}

		局_现行时间戳 := time.Now().Unix()
		if info.卡号详情.VipTime != 0 { //只有时间增减不为0的时候设置的用户分类才有效
			if info.app用户详情.UserClassId == info.卡号详情.UserClassId {
				//分类相同,正常处理时间或点数
				if Ser_AppInfo.App是否为计点(info.卡号详情.AppId) || info.app用户详情.VipTime > 局_现行时间戳 {
					//如果为计点 或 时间大于现在时间直接加就行了
					客户expr["VipTime"] = gorm.Expr("VipTime + ?", info.卡号详情.VipTime)
				} else {
					//如果为计时 已经过期很久了,直接现行时间戳加卡时间
					客户expr["VipTime"] = 局_现行时间戳 + info.卡号详情.VipTime
				}
			} else {
				//用户类型不同, 根据权重处理
				var 局_旧用户类型权重, 局_新用户类型权重 DB.DB_UserClass
				if info.app用户详情.UserClassId > 0 {
					if 局_旧用户类型权重, err = service.NewUserClass(c, tx).Info(info.app用户详情.UserClassId); err != nil {
						return errors.Join(err, errors.New("读取旧用户类型权重失败"))
					}
				} else {
					局_旧用户类型权重.Weight = 1
				}

				if info.卡号详情.UserClassId > 0 {
					if 局_新用户类型权重, err = service.NewUserClass(c, tx).Info(info.卡号详情.UserClassId); err != nil {
						return errors.Join(err, errors.New("读取新用户类型权重失败"))
					}
				} else {
					局_新用户类型权重.Weight = 1
				}

				if info.is计点 {
					//转换结果值,转后再增加新类型 值
					客户expr["VipTime"] = gorm.Expr("VipTime * ? / ? +?", 局_旧用户类型权重.Weight, 局_新用户类型权重.Weight, info.卡号详情.VipTime)
					info.logVipNumber = append(info.logVipNumber, DB.DB_LogVipNumber{
						User:  info.user用户详情.User,
						Ip:    info.ip,
						Count: info.卡号详情.VipNumber,
						Note:  "应用ID:" + strconv.Itoa(info.卡号详情.AppId) + "卡号Id:" + strconv.Itoa(info.卡号详情.Id) + "充值点数",
					})
				} else {
					if info.app用户详情.VipTime < 局_现行时间戳 {
						//已经过期了直接赋值新类型 现行时间+新时间就可以了
						客户expr["VipTime"] = 局_现行时间戳 + info.卡号详情.VipTime
					} else {
						//先计算还剩多长时间,剩余时间权重转换转换结果值,+现在时间+卡增减时间
						客户expr["VipTime"] = gorm.Expr("(VipTime-?) * ? / ? +?", 局_现行时间戳, 局_旧用户类型权重.Weight, 局_新用户类型权重.Weight, 局_现行时间戳+info.卡号详情.VipTime)
					}
				}
				//最后更换类型,防止前面用到卡类id,计算权重转换类型错误
				客户expr["UserClassId"] = info.卡号详情.UserClassId
			}
		}
		//更新客户数据
		err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).Where("Id = ?", info.app用户详情.Id).Updates(&客户expr).Error
		if err != nil {
			return errors.Join(err, errors.New("充值失败,重试"))
		}
		//处理账号的RMB增减
		if !info.is卡号 && info.卡号详情.RMb > 0 {
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id=?", info.app用户详情.Uid).First(&info.user用户详情).Error
			if err != nil {
				return errors.Join(err, errors.New("用户账号不存在"))
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", info.app用户详情.Uid).Update("RMB", gorm.Expr("RMB + ?", info.卡号详情.RMb)).Error
			if err != nil {
				return errors.Join(err, errors.New("充值余额时失败,请重试"))
			}
			var 局_新余额 float64
			err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", info.user用户详情.Id).First(&局_新余额).Error
			if err != nil {
				return errors.Join(err, errors.New("充值后读取新余额失败"))
			}
			//日志仅写到上下文内,由实际业务处理是否写入日志和修改备注信息
			info.logMoney = append(info.logMoney, DB.DB_LogMoney{
				User:  info.user用户详情.User,
				Ip:    info.ip,
				Count: info.卡号详情.RMb,
				Note:  "应用ID:" + strconv.Itoa(info.卡号详情.AppId) + "卡号Id:" + strconv.Itoa(info.卡号详情.Id) + "充值余额|新余额≈" + Float64到文本(局_新余额, 2),
			})
		}
		//开始处理推荐人
		if info.卡号详情.InviteCount > 0 && info.app用户详情_推荐人.Id != 0 {
			//处理新信息
			推荐人expr := map[string]interface{}{}
			if info.app用户详情_推荐人.UserClassId == info.卡号详情.UserClassId {
				//分类相同,正常处理时间或点数
				if info.is计点 || info.app用户详情_推荐人.VipTime > 局_现行时间戳 {
					//如果为计点 或 时间大于现在时间直接加就行了
					推荐人expr["VipTime"] = gorm.Expr("VipTime + ?", info.卡号详情.InviteCount)
				} else {
					//如果为计时 已经过期很久了,直接现行时间戳加卡时间
					推荐人expr["VipTime"] = 局_现行时间戳 + info.卡号详情.InviteCount
				}

			} else {
				//用户类型不同, 根据权重处理
				if info.用户类型_推荐人, err = service.NewUserClass(c, tx).Info(info.app用户详情_推荐人.UserClassId); err != nil {
					return errors.New("读取推荐人用户类型详情失败,可能已不存在id:" + strconv.Itoa(info.app用户详情_推荐人.UserClassId))
				}
				if info.新用户类型_推荐人, err = service.NewUserClass(c, tx).Info(info.卡号详情.UserClassId); err != nil {
					return errors.New("读取卡号详情UserClassId失败,可能已不存在id:" + strconv.Itoa(info.卡号详情.UserClassId))
				}

				if info.is计点 {
					//计算推荐人用户类型的实际点数,
					//这里有点绕,比如增加1秒,但是这个新用户类型权重为2, 荐人类型权重为1
					//那么实际新用户类型是推荐人类型的两倍,转换到推荐人类型值应该为2
					//所以 卡秒数+新用户类型权重=2,在除以推荐人用户类型权重=2
					局_增减时间点数 := info.卡号详情.InviteCount * info.新用户类型_推荐人.Weight / info.用户类型_推荐人.Weight
					推荐人expr["VipTime"] = gorm.Expr("VipTime +?", 局_增减时间点数)
					info.logVipNumber = append(info.logVipNumber, DB.DB_LogVipNumber{
						User:  info.user用户详情.User,
						Ip:    info.ip,
						Count: Int64到Float64(局_增减时间点数),
						Note:  "应用ID:" + strconv.Itoa(info.卡号详情.AppId) + "用户:" + 充值用户 + ",使用充值卡号Id:" + strconv.Itoa(info.卡号详情.Id) + ",获得推荐人增加点数",
					})

				} else {
					局_增减时间点数 := info.卡号详情.InviteCount * info.新用户类型_推荐人.Weight / info.用户类型_推荐人.Weight
					if info.app用户详情_推荐人.VipTime < 局_现行时间戳 {
						//已经过期了 现行时间+新时间就可以了
						推荐人expr["VipTime"] = 局_现行时间戳 + 局_增减时间点数
					} else {
						//原来的值+推荐人增加点数权重转换结果就好了
						推荐人expr["VipTime"] = gorm.Expr("VipTime+?", 局_增减时间点数)
					}
				}
			}
			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).Where("Uid = ?", info.app用户详情_推荐人.Uid).Updates(&推荐人expr).Error
			if err != nil {
				return errors.Join(err, errors.New("推荐人更新失败"))
			}
		}

		return nil
	})
	//写到上下文,备用
	c.Set("info.卡类详情", info.卡号详情)
	c.Set("info.app用户详情", info.app用户详情)
	c.Set("info.user用户详情", info.user用户详情)
	c.Set("info.app详情", info.app详情)
	if len(info.logMoney) > 0 {
		//最后写出日志
		if err = log.L_log.S输出日志(c, info.logMoney); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
	}

	if len(info.logVipNumber) > 0 {
		if err = log.L_log.S输出日志(c, info.logVipNumber); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
	}

	return err
}

// 已用充值卡将相应的卡使用者和推荐人强行扣回充值卡面值,可能扣成负数
func (j *ka) K卡号追回(c *gin.Context, Id int, 操作人 string) (err error) {
	var info struct {
		卡号详情         DB.DB_Ka
		卡号应用         DB.DB_AppInfo
		is卡号         bool
		is计点         bool
		vipTime名称    string
		已充值用户        []DB.DB_AppUser
		操作人详情        DB.DB_User
		LogMoney     []DB.DB_LogMoney
		LogVipNumber []DB.DB_LogVipNumber
	}
	// 开启事务,检测上层是否有事务,如果有直接使用,没有就创建一个
	var db *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		db = tempObj.(*gorm.DB)
	} else {
		db2 := *global.GVA_DB
		db = &db2
	}
	if info.卡号详情, err = service.NewKa(c, db).Info(Id); err != nil {
		err = errors.Join(err, errors.New("卡号不存在"))
		return
	}
	if info.卡号详情.Num == 0 {
		return errors.New("卡号未使用")
	}

	//防sb客户放负值
	if info.卡号详情.VipTime < 0 || info.卡号详情.VipNumber < 0 || info.卡号详情.RMb < 0 {
		return errors.New("追回的卡号,充值的时间点数,积分,rmb,不能为负数,请手动处理")
	}

	if info.卡号应用, err = service.NewAppInfo(c, db).Info(info.卡号详情.AppId); err != nil {
		err = errors.Join(err, errors.New("应用不存在")) //概率较小,但是有可能, 比如制卡使用后把应用删除了,然后代理追回卡号
		return
	}
	info.is卡号 = info.卡号应用.AppType == 3 || info.卡号应用.AppType == 4
	info.is计点 = info.卡号应用.AppType == 2 || info.卡号应用.AppType == 4

	if info.is卡号 && info.卡号详情.User == "" {
		return errors.New("卡号模式应用,已转换软件用户的卡号,无法追回")
	}

	if info.卡号详情.User == "" {
		return errors.New("无已充值用户,但有使用次数,可能手动修改")
	}

	info.vipTime名称 = S三元(info.is计点, "点数", "时间")

	已用用户数组 := W文本_分割文本(info.卡号详情.User, ",")
	已用推荐人数组 := W文本_分割文本(info.卡号详情.InviteUser, ",")

	err = db.Transaction(func(tx *gorm.DB) error {
		//tx = tx.Debug()
		for 索引, 值 := range 已用用户数组 {
			if 值 == "" {
				continue //如果值为空,到循环尾
			}

			var 临时软件用户info DB.DB_AppUser
			var 临时账号info DB.DB_User
			var 临时卡号info DB.DB_Ka

			if info.is卡号 {
				tempTx := tx.Model(DB.DB_Ka{}).Clauses(clause.Locking{Strength: "UPDATE"}).First(&临时卡号info, "Name = ?", 值)
				if tempTx.Error != nil {
					return errors.Join(tempTx.Error, errors.New(值+"已充卡号不存在"))
				}
			} else {
				tempTx := tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).First(&临时账号info, "User = ?", 值)
				if tempTx.Error != nil {
					return errors.Join(tempTx.Error, errors.New("已充账号不存在"))
				}
			}
			tempTx := tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&临时软件用户info, "Uid = ?", S三元(info.is卡号, 临时卡号info.Id, 临时账号info.Id))
			if tempTx.Error != nil {
				return errors.Join(tempTx.Error, errors.New(值+"已充软件用户info不存在"))
			}
			临时软件用户info.VipNumber = Float64减float64(临时软件用户info.VipNumber, info.卡号详情.VipNumber, 2)

			局_临时整数, err2 := userClass.L_userClass.J计算权重值2(c, info.卡号详情.UserClassId, 临时软件用户info.UserClassId, info.卡号详情.VipTime)
			if err2 != nil {
				return errors.Join(tempTx.Error, errors.New(值+"计算权重值失败"))
			}
			临时软件用户info.VipTime -= 局_临时整数
			update := map[string]interface{}{
				"VipTime":   临时软件用户info.VipTime,
				"VipNumber": 临时软件用户info.VipNumber,
			}
			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).Where("Id = ?", 临时软件用户info.Id).
				Updates(&update).Error
			if err != nil {
				return errors.Join(err, errors.New("修改uid:"+strconv.Itoa(临时软件用户info.Uid)+"软件用户信息失败"))
			}
			if info.卡号详情.VipTime != 0 {
				info.LogVipNumber = append(info.LogVipNumber, DB.DB_LogVipNumber{
					User:  S三元(info.is卡号, 临时卡号info.Name, 临时账号info.User),
					AppId: info.卡号详情.AppId,
					Type:  S三元(info.is计点, constant.Log_type_点数, constant.Log_type_时间),
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Int64到Float64(info.卡号详情.VipTime),
					Note:  fmt.Sprintf(操作人+"操作追回id:%d,卡号:%s,扣除卡号充值"+info.vipTime名称+",用户类型id:卡号%d->用户%d", info.卡号详情.Id, info.卡号详情.Name, info.卡号详情.UserClassId, 临时软件用户info.UserClassId),
				})
			}
			if info.卡号详情.VipNumber != 0 {
				info.LogVipNumber = append(info.LogVipNumber, DB.DB_LogVipNumber{
					User:  S三元(info.is卡号, 临时卡号info.Name, 临时账号info.User),
					AppId: info.卡号详情.AppId,
					Type:  constant.Log_type_积分,
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Float64取负值(info.卡号详情.VipNumber),
					Note:  fmt.Sprintf(操作人+"操作追回id:%d,卡号:%s,扣除卡号充值积分|新积分≈%s", info.卡号详情.Id, info.卡号详情.Name, Float64到文本(临时软件用户info.VipNumber, 2)),
				})
			}

			if info.is卡号 {
				//	卡号没什么可变更的
			} else {
				//余额只需要改RMB
				if info.卡号详情.RMb != 0 {
					临时账号info.Rmb = Float64减float64(临时账号info.Rmb, info.卡号详情.RMb, 2)
					tempTx = tx.Model(DB.DB_User{}).Where("Id = ?", 临时账号info.Id).
						Update("Rmb", 临时账号info.Rmb)
					if tempTx.Error != nil {
						return errors.Join(tempTx.Error, errors.New("修改id:"+strconv.Itoa(临时账号info.Id)+"用户Rmb信息失败"))
					}
					info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
						User:  临时账号info.User,
						Time:  time.Now().Unix(),
						Ip:    c.ClientIP(),
						Count: Float64取负值(info.卡号详情.RMb),
						Note:  fmt.Sprintf(操作人+"操作追回id:%d,卡号:%s,扣除卡号充值余额%s|新余额≈%s", info.卡号详情.Id, info.卡号详情.Name, Float64到文本(info.卡号详情.RMb, 2), Float64到文本(临时账号info.Rmb, 2)),
					})
				}
			}

			//开始处理推荐人增加的时间点数
			if 索引 >= len(已用推荐人数组) {
				continue
			}
			值 = 已用推荐人数组[索引]
			if 值 == "" {
				continue
			}
			// First 传入的如果带主键id会自动增加这个条件
			临时账号info = DB.DB_User{}
			临时卡号info = DB.DB_Ka{}
			临时软件用户info = DB.DB_AppUser{}
			if info.is卡号 {
				tempTx = tx.Model(DB.DB_Ka{}).Clauses(clause.Locking{Strength: "UPDATE"}).First(&临时卡号info, "Name = ?", 值)
				if tempTx.Error != nil {
					return errors.Join(tempTx.Error, errors.New("推荐人["+值+"]卡号不存在"))
				}
			} else {
				tempTx = tx.Debug().Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).First(&临时账号info, "User = ?", 值)
				if tempTx.Error != nil {
					return errors.Join(tempTx.Error, errors.New("推荐人["+值+"]账号不存在"))
				}
			}
			tempTx = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&临时软件用户info, "Uid = ?", S三元(info.is卡号, 临时卡号info.Id, 临时账号info.Id))
			if tempTx.Error != nil {
				return errors.Join(tempTx.Error, errors.New(值+"推荐人软件用户info不存在"))
			}
			局_临时整数, err2 = userClass.L_userClass.J计算权重值2(c, info.卡号详情.UserClassId, 临时软件用户info.UserClassId, info.卡号详情.InviteCount)
			if err2 != nil {
				return errors.Join(tempTx.Error, errors.New(值+"计算推荐人权重值失败"))
			}
			临时软件用户info.VipTime -= 局_临时整数
			update = map[string]interface{}{
				"VipTime": 临时软件用户info.VipTime,
			}
			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(info.卡号详情.AppId)).Where("Id = ?", 临时软件用户info.Id).
				Updates(&update).Error
			if err != nil {
				return errors.Join(err, errors.New("修改推荐人uid:"+strconv.Itoa(临时软件用户info.Uid)+"软件用户信息失败"))
			}
			if info.卡号详情.VipTime != 0 {
				info.LogVipNumber = append(info.LogVipNumber, DB.DB_LogVipNumber{
					User:  S三元(info.is卡号, 临时卡号info.Name, 临时账号info.User),
					AppId: info.卡号详情.AppId,
					Type:  S三元(info.is计点, constant.Log_type_点数, constant.Log_type_时间),
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Int64到Float64(info.卡号详情.VipTime),
					Note:  fmt.Sprintf(操作人+"操作追回id:%d,卡号:%s,扣除卡号推荐人充值"+S三元(info.is计点, "点数", "时间"), info.卡号详情.Id, info.卡号详情.Name),
				})
			}

		}
		//重置卡并冻结,删除信息
		err = tx.Model(DB.DB_Ka{}).Where("Id = ? ", info.卡号详情.Id).Updates(
			map[string]interface{}{
				"Status":     2,
				"User":       "",
				"Num":        0,
				"InviteUser": "",
				"UserTime":   "",
				"UseTime":    0,
				"AdminNote":  info.卡号详情.AdminNote + "已被追回,历史充值用户:" + info.卡号详情.User + ",推荐人用户:" + info.卡号详情.InviteUser,
			}).Error
		return err
	})

	if err == nil {
		//如果是上层传递的,就返回日志,否则直接输出
		if _, ok := c.Get("tx"); ok {
			c.Set("info.LogVipNumber", info.LogVipNumber)
			c.Set("info.LogMoney", info.LogMoney)
		} else {
			if err = log.L_log.S输出日志(c, info.LogMoney); err != nil {
				global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
			}
			if err = log.L_log.S输出日志(c, info.LogVipNumber); err != nil {
				global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
			}
		}
	}

	return err
}

// 删除已耗尽次数卡号
func (j *ka) S删除耗尽次数卡号(c *gin.Context, AppId int) (数量 int64, err error) {
	var info struct {
		卡号详情  []DB.DB_Ka
		删除id  []int
		app详情 DB.DB_AppInfo
		is卡号  bool
		is计点  bool
		ip    string
		LogKa []DB.DB_LogKa //卡号
	}
	info.ip = "未知" // 处理 `gin.Context` 为 `nil` 的情况
	if c != nil {
		info.ip = c.ClientIP()
	}
	db := *global.GVA_DB
	db.Where("Num = NumMax ").Where("AppId= ? ", AppId).Find(&info.卡号详情)
	if info.卡号详情 == nil || len(info.卡号详情) == 0 {
		return 0, nil
	}
	if info.app详情, err = service.NewAppInfo(c, &db).Info(AppId); err != nil {
		err = errors.New("应用不存在")
		return
	}
	info.is卡号 = S三元(info.app详情.AppType == 3 || info.app详情.AppType == 4, true, false)
	info.is计点 = S三元(info.app详情.AppType == 2 || info.app详情.AppType == 4, true, false)

	if !info.is卡号 {
		for _, v := range info.卡号详情 {
			info.删除id = append(info.删除id, v.Id)
		}
	} else {
		局_已过期卡号id := []int{}
		for _, v := range info.卡号详情 {
			局_已过期卡号id = append(局_已过期卡号id, v.Id)
		}
		局_vipTime := time.Now().Unix()
		if info.is计点 {
			局_vipTime = 0
		}
		//这些卡号Uid 且未过期的
		局_排除Id := []int{}
		db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).
			Where("Uid IN ?", 局_已过期卡号id).
			Where("VipTime > ?", 局_vipTime).
			Select("Uid").
			Find(&局_排除Id)
		for _, v := range 局_已过期卡号id {
			if !S数组_整数是否存在(局_排除Id, v) {
				info.删除id = append(info.删除id, v)
			}
		}

	}
	if len(info.删除id) == 0 {
		return 0, nil
	}

	数量, err = service.NewKa(c, &db).Delete(info.删除id)
	if err == nil && 数量 > 0 {
		局_临时数组 := make([]DB.DB_Ka, 0, len(info.卡号详情))
		for v := range len(info.卡号详情) {
			if S数组_整数是否存在(info.删除id, info.卡号详情[v].Id) {
				局_临时数组 = append(局_临时数组, info.卡号详情[v])
			}
		}
		局_时间 := time.Now().Unix()
		for i, v := range 局_临时数组 {
			局_文本 := fmt.Sprintf("批量删除耗尽卡号批次id:%d(%d/%d),卡号id:%d,卡类id:%d", 局_时间, i+1, len(局_临时数组), v.Id, v.KaClassId)
			info.LogKa = append(info.LogKa, DB.DB_LogKa{
				User:     c.GetString("User"),
				UserType: constant.Log_卡操作用户_管理员,
				Ka:       v.Name,
				KaType:   constant.Log_卡操作_删,
				Time:     局_时间,
				Ip:       info.ip,
				Note:     局_文本,
			})
		}
		//最后写出日志
		if err = log.L_log.S输出日志(c, info.LogKa); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
	}

	return
}
