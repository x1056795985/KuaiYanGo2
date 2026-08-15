package monitoring

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm/logger"

	"server/app/global"
)

var (
	零_SQL单引号正则  = regexp.MustCompile(`'[^']*'`)
	零_SQL双引号正则  = regexp.MustCompile(`"[^"]*"`)
	零_SQL数字正则   = regexp.MustCompile(`\b\d+\b`)
	零_SQL十六进制正则 = regexp.MustCompile(`0x[0-9a-fA-F]+`)
)

type 监控_gorm日志器 struct {
	基础日志器  logger.Interface
	监控实例   *监控器
	慢SQL阈值 time.Duration
}

func C初始化Gorm日志器(基础日志器 logger.Interface, 慢SQL阈值 time.Duration) logger.Interface {
	if 慢SQL阈值 <= 0 {
		慢SQL阈值 = 零_默认慢SQL毫秒 * time.Millisecond
	}
	return &监控_gorm日志器{
		基础日志器:  基础日志器,
		监控实例:   Q监控,
		慢SQL阈值: 慢SQL阈值,
	}
}

func (j *监控_gorm日志器) LogMode(日志级别 logger.LogLevel) logger.Interface {
	return &监控_gorm日志器{
		基础日志器:  j.基础日志器.LogMode(日志级别),
		监控实例:   j.监控实例,
		慢SQL阈值: j.慢SQL阈值,
	}
}

func (j *监控_gorm日志器) Info(ctx context.Context, msg string, data ...interface{}) {
	j.基础日志器.Info(ctx, msg, data...)
}

func (j *监控_gorm日志器) Warn(ctx context.Context, msg string, data ...interface{}) {
	j.基础日志器.Warn(ctx, msg, data...)
}

func (j *监控_gorm日志器) Error(ctx context.Context, msg string, data ...interface{}) {
	j.基础日志器.Error(ctx, msg, data...)
}

func (j *监控_gorm日志器) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	局_耗时 := time.Since(begin)
	局_sql语句, 局_影响行数 := fc()
	j.监控实例.记录SQL(局_耗时, 局_影响行数, 局_sql语句, err, j.慢SQL阈值)
	j.基础日志器.Trace(ctx, begin, func() (string, int64) {
		return 局_sql语句, 局_影响行数
	}, err)
}

func (j *监控器) 记录SQL(耗时 time.Duration, 影响行数 int64, sql语句 string, 错误 error, 慢SQL阈值 time.Duration) {
	局_记录时间 := time.Now().Format(time.RFC3339)
	局_sql语句 := 清洗SQL(sql语句)
	局_耗时毫秒 := 保留两位小数(时长转毫秒(耗时))

	if 耗时 >= 慢SQL阈值 {
		j.互斥锁.Lock()
		j.慢SQL = 追加有界切片(j.慢SQL, 监控_慢SQL记录{
			S时间:    局_记录时间,
			H耗时毫秒:  局_耗时毫秒,
			Y影响行数:  影响行数,
			SSQL语句: 局_sql语句,
		}, 零_最大慢SQL数量)
		j.互斥锁.Unlock()
	}

	if 错误 == nil || errors.Is(错误, logger.ErrRecordNotFound) {
		return
	}

	j.互斥锁.Lock()
	j.数据库错误 = 追加有界切片(j.数据库错误, 监控_数据库错误记录{
		S时间:    局_记录时间,
		H耗时毫秒:  局_耗时毫秒,
		Y影响行数:  影响行数,
		S死锁:    是否死锁错误(错误),
		C错误:    错误.Error(),
		SSQL语句: 局_sql语句,
	}, 零_最大数据库错误数量)
	j.互斥锁.Unlock()
}

func (j *监控器) 取数据库总览() 响应_数据库总览 {
	j.互斥锁.RLock()
	局_慢SQL列表 := append([]监控_慢SQL记录(nil), j.慢SQL...)
	局_数据库错误列表 := append([]监控_数据库错误记录(nil), j.数据库错误...)
	j.互斥锁.RUnlock()

	for 局_左, 局_右 := 0, len(局_慢SQL列表)-1; 局_左 < 局_右; 局_左, 局_右 = 局_左+1, 局_右-1 {
		局_慢SQL列表[局_左], 局_慢SQL列表[局_右] = 局_慢SQL列表[局_右], 局_慢SQL列表[局_左]
	}
	for 局_左, 局_右 := 0, len(局_数据库错误列表)-1; 局_左 < 局_右; 局_左, 局_右 = 局_左+1, 局_右-1 {
		局_数据库错误列表[局_左], 局_数据库错误列表[局_右] = 局_数据库错误列表[局_右], 局_数据库错误列表[局_左]
	}

	局_结果 := 响应_数据库总览{
		S是否启用:     global.GVA_DB != nil,
		M慢SQL阈值毫秒: 零_默认慢SQL毫秒,
		M慢SQL列表:   局_慢SQL列表,
		C错误列表:     局_数据库错误列表,
		M模板排行:     构建SQL模板排行(局_慢SQL列表, 局_数据库错误列表),
	}

	if global.GVA_DB == nil {
		return 局_结果
	}

	局_db := *global.GVA_DB
	局_sqlDB, 局_错误 := (&局_db).DB()
	if 局_错误 != nil {
		return 局_结果
	}

	局_结果.L连接池 = 映射数据库连接池(局_sqlDB.Stats())
	return 局_结果
}

func 映射数据库连接池(连接池 sql.DBStats) 响应_数据库连接池 {
	return 响应_数据库连接池{
		D打开连接数:     连接池.OpenConnections,
		Z正在使用数:     连接池.InUse,
		K空闲连接数:     连接池.Idle,
		D等待次数:      连接池.WaitCount,
		D等待毫秒:      保留两位小数(时长转毫秒(连接池.WaitDuration)),
		Z最大空闲关闭数:   连接池.MaxIdleClosed,
		Z最大空闲时长关闭数: 连接池.MaxIdleTimeClosed,
		Z最大生命周期关闭数: 连接池.MaxLifetimeClosed,
	}
}

func 是否死锁错误(错误 error) bool {
	局_错误文本 := strings.ToLower(错误.Error())
	return strings.Contains(局_错误文本, "deadlock") ||
		strings.Contains(局_错误文本, "error 1213") ||
		strings.Contains(局_错误文本, "lock wait timeout") ||
		strings.Contains(局_错误文本, "error 1205")
}

func 清洗SQL(sql语句 string) string {
	局_sql语句 := strings.TrimSpace(sql语句)
	局_sql语句 = strings.ReplaceAll(局_sql语句, "\r", " ")
	局_sql语句 = strings.ReplaceAll(局_sql语句, "\n", " ")
	for strings.Contains(局_sql语句, "  ") {
		局_sql语句 = strings.ReplaceAll(局_sql语句, "  ", " ")
	}
	return 局_sql语句
}

func 构建SQL模板排行(慢SQL列表 []监控_慢SQL记录, 数据库错误列表 []监控_数据库错误记录) []响应_SQL模板聚合 {
	type SQL模板聚合中间 struct {
		模板   string
		次数   int
		总毫秒  float64
		最大毫秒 float64
		死锁次数 int
		最近时间 string
	}

	局_映射 := make(map[string]*SQL模板聚合中间)
	取聚合项 := func(模板 string) *SQL模板聚合中间 {
		if 模板 == "" {
			模板 = "-"
		}
		局_项, 局_存在 := 局_映射[模板]
		if 局_存在 {
			return 局_项
		}
		局_项 = &SQL模板聚合中间{模板: 模板}
		局_映射[模板] = 局_项
		return 局_项
	}

	for _, 局_记录 := range 慢SQL列表 {
		局_模板 := 归一化SQL模板(局_记录.SSQL语句)
		局_项 := 取聚合项(局_模板)
		局_项.次数++
		局_项.总毫秒 += 局_记录.H耗时毫秒
		if 局_记录.H耗时毫秒 > 局_项.最大毫秒 {
			局_项.最大毫秒 = 局_记录.H耗时毫秒
		}
		if 局_记录.S时间 > 局_项.最近时间 {
			局_项.最近时间 = 局_记录.S时间
		}
	}

	for _, 局_记录 := range 数据库错误列表 {
		局_模板 := 归一化SQL模板(局_记录.SSQL语句)
		局_项 := 取聚合项(局_模板)
		局_项.次数++
		局_项.总毫秒 += 局_记录.H耗时毫秒
		if 局_记录.H耗时毫秒 > 局_项.最大毫秒 {
			局_项.最大毫秒 = 局_记录.H耗时毫秒
		}
		if 局_记录.S死锁 {
			局_项.死锁次数++
		}
		if 局_记录.S时间 > 局_项.最近时间 {
			局_项.最近时间 = 局_记录.S时间
		}
	}

	局_结果 := make([]响应_SQL模板聚合, 0, len(局_映射))
	for _, 局_项 := range 局_映射 {
		局_平均毫秒 := 0.0
		if 局_项.次数 > 0 {
			局_平均毫秒 = 保留两位小数(局_项.总毫秒 / float64(局_项.次数))
		}
		局_结果 = append(局_结果, 响应_SQL模板聚合{
			M模板:   截断文本(局_项.模板, 240),
			C次数:   局_项.次数,
			P平均毫秒: 局_平均毫秒,
			Z最大毫秒: 保留两位小数(局_项.最大毫秒),
			S死锁次数: 局_项.死锁次数,
			Z最近时间: 局_项.最近时间,
		})
	}

	sort.Slice(局_结果, func(i, k int) bool {
		if 局_结果[i].S死锁次数 == 局_结果[k].S死锁次数 {
			if 局_结果[i].Z最大毫秒 == 局_结果[k].Z最大毫秒 {
				return 局_结果[i].C次数 > 局_结果[k].C次数
			}
			return 局_结果[i].Z最大毫秒 > 局_结果[k].Z最大毫秒
		}
		return 局_结果[i].S死锁次数 > 局_结果[k].S死锁次数
	})

	return 局_结果
}

func 归一化SQL模板(sql语句 string) string {
	局_模板 := 清洗SQL(sql语句)
	局_模板 = 零_SQL十六进制正则.ReplaceAllString(局_模板, "?")
	局_模板 = 零_SQL单引号正则.ReplaceAllString(局_模板, "?")
	局_模板 = 零_SQL双引号正则.ReplaceAllString(局_模板, "?")
	局_模板 = 零_SQL数字正则.ReplaceAllString(局_模板, "?")
	局_模板 = strings.ToUpper(局_模板)
	return 局_模板
}
