package monitoring

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"server/app/global"
)

// 响应_MySQL诊断 MySQL 诊断结果
type 响应_MySQL诊断 struct {
	S是否成功   bool             `json:"success"`
	C采集时间   string           `json:"collectedAt"`
	C错误      string           `json:"error"`
	J进程列表   []响应_MySQL进程  `json:"processList"`
	M慢日志文件 string           `json:"slowQueryLogFile"`
	M慢日志状态 string           `json:"slowQueryLogStatus"`
	C长查询时间 string           `json:"longQueryTime"`
	Z摘要排行   []响应_MySQL摘要  `json:"statementSummary"`
	J进程总数   int              `json:"processCount"`
	S睡眠进程数 int              `json:"sleepProcessCount"`
	H活跃进程数 int              `json:"activeProcessCount"`
	Z最长睡眠秒 int64            `json:"maxSleepSeconds"`
	S说明      []string         `json:"notes"`
}

// 响应_MySQL进程 SHOW FULL PROCESSLIST 单行
type 响应_MySQL进程 struct {
	Id      int64  `json:"id"`
	User    string `json:"user"`
	Host    string `json:"host"`
	Db      string `json:"db"`
	Command string `json:"command"`
	Time    int64  `json:"time"`
	State   string `json:"state"`
	Info    string `json:"info"`
	Level   string `json:"level"` // normal/warning/danger 用于前端着色
}

// 响应_MySQL摘要 events_statements_summary_by_digest 单行
type 响应_MySQL摘要 struct {
	SchemaName       string  `json:"schemaName"`
	DigestText       string  `json:"digestText"`
	CountStar        uint64  `json:"countStar"`
	SumTimerMs       float64 `json:"sumTimerMs"`
	MinTimerMs       float64 `json:"minTimerMs"`
	AvgTimerMs       float64 `json:"avgTimerMs"`
	MaxTimerMs       float64 `json:"maxTimerMs"`
	SumLockMs        float64 `json:"sumLockMs"`
	SumErrors        uint64  `json:"sumErrors"`
	SumWarnings      uint64  `json:"sumWarnings"`
	SumRowsAffected  uint64  `json:"sumRowsAffected"`
	SumRowsSent      uint64  `json:"sumRowsSent"`
	SumRowsExamined  uint64  `json:"sumRowsExamined"`
	SumTmpDiskTables uint64  `json:"sumTmpDiskTables"`
	SumTmpTables     uint64  `json:"sumTmpTables"`
	SumFullJoin      uint64  `json:"sumFullJoin"`
	SumFullRangeJoin uint64  `json:"sumFullRangeJoin"`
	SumSelectRange   uint64  `json:"sumSelectRange"`
	SumSelectScan    uint64  `json:"sumSelectScan"`
	SumNoIndexUsed   uint64  `json:"sumNoIndexUsed"`
	FirstSeen        string  `json:"firstSeen"`
	LastSeen         string  `json:"lastSeen"`
	Level            string  `json:"level"`  // normal/warning/danger
	Advice           string  `json:"advice"` // 人话解释
}

// QMySQL诊断 执行三步 MySQL 诊断：
//  1. SHOW FULL PROCESSLIST
//  2. 开启慢日志 + 查询慢日志文件路径
//  3. 查询 performance_schema.events_statements_summary_by_digest 取耗时 Top 10
func (j *监控器) QMySQL诊断() (*响应_MySQL诊断, error) {
	局_结果 := &响应_MySQL诊断{
		C采集时间: time.Now().Format(time.RFC3339),
	}

	if global.GVA_DB == nil {
		局_结果.C错误 = "未连接数据库"
		局_结果.S说明 = []string{"后端未启用数据库连接，无法执行 MySQL 诊断。"}
		return 局_结果, nil
	}

	局_db := *global.GVA_DB

	// 1. SHOW FULL PROCESSLIST
	局_进程列表, 局_进程错误 := j.查询进程列表(&局_db)
	if 局_进程错误 != nil {
		局_结果.C错误 = fmt.Sprintf("查询 SHOW FULL PROCESSLIST 失败:%s", 局_进程错误.Error())
		return 局_结果, 局_进程错误
	}
	局_结果.J进程列表 = 局_进程列表
	局_结果.J进程总数 = len(局_进程列表)
	for _, 局_进程 := range 局_进程列表 {
		if strings.EqualFold(局_进程.Command, "Sleep") {
			局_结果.S睡眠进程数++
			if 局_进程.Time > 局_结果.Z最长睡眠秒 {
				局_结果.Z最长睡眠秒 = 局_进程.Time
			}
		} else {
			局_结果.H活跃进程数++
		}
	}

	// 2. 开启慢查询日志 + 读取慢日志文件路径
	局_慢日志文件, 局_慢日志状态, 局_长查询时间, 局_慢日志错误 := j.配置慢查询日志(&局_db)
	if 局_慢日志错误 != nil {
		// 不阻断后续流程，仅记录错误
		if 局_结果.C错误 == "" {
			局_结果.C错误 = fmt.Sprintf("配置慢查询日志失败:%s", 局_慢日志错误.Error())
		}
	} else {
		局_结果.M慢日志文件 = 局_慢日志文件
		局_结果.M慢日志状态 = 局_慢日志状态
		局_结果.C长查询时间 = 局_长查询时间
	}

	// 3. 查询 SQL 摘要 Top 10
	局_摘要列表, 局_摘要错误 := j.查询SQL摘要(&局_db)
	if 局_摘要错误 != nil {
		if 局_结果.C错误 == "" {
			局_结果.C错误 = fmt.Sprintf("查询 SQL 摘要失败:%s", 局_摘要错误.Error())
		}
	} else {
		局_结果.Z摘要排行 = 局_摘要列表
	}

	局_结果.S是否成功 = 局_进程错误 == nil && 局_慢日志错误 == nil && 局_摘要错误 == nil
	局_结果.S说明 = 构建MySQL诊断说明(局_结果)
	return 局_结果, nil
}

// 查询进程列表 执行 SHOW FULL PROCESSLIST
func (j *监控器) 查询进程列表(数据库 *gorm.DB) ([]响应_MySQL进程, error) {
	局_行集, 局_错误 := 数据库.Raw("SHOW FULL PROCESSLIST").Rows()
	if 局_错误 != nil {
		return nil, 局_错误
	}
	defer 局_行集.Close()

	局_结果 := make([]响应_MySQL进程, 0, 16)
	for 局_行集.Next() {
		var 局_行 struct {
			Id      int64
			User    string
			Host    string
			Db      *string
			Command string
			Time    int64
			State   *string
			Info    *string
		}
		if 局_扫描错误 := 数据库.ScanRows(局_行集, &局_行); 局_扫描错误 != nil {
			continue
		}
		局_进程 := 响应_MySQL进程{
			Id:      局_行.Id,
			User:    局_行.User,
			Host:    局_行.Host,
			Db:      取可空字符串(局_行.Db),
			Command: 局_行.Command,
			Time:    局_行.Time,
			State:   取可空字符串(局_行.State),
			Info:    取可空字符串(局_行.Info),
		}
		局_进程.Level = 评估进程级别(局_进程)
		局_结果 = append(局_结果, 局_进程)
	}

	// 按 Time 倒序，Sleep 排后面，活跃进程排前面
	sort.SliceStable(局_结果, func(i, k int) bool {
		if 局_结果[i].Command == "Sleep" && 局_结果[k].Command != "Sleep" {
			return false
		}
		if 局_结果[i].Command != "Sleep" && 局_结果[k].Command == "Sleep" {
			return true
		}
		return 局_结果[i].Time > 局_结果[k].Time
	})

	return 局_结果, nil
}

// 配置慢查询日志 开启慢日志、设置阈值、读取慢日志文件路径
func (j *监控器) 配置慢查询日志(数据库 *gorm.DB) (慢日志文件 string, 慢日志状态 string, 长查询时间 string, 错误 error) {
	// 开启慢查询日志
	if 局_错误 := 数据库.Exec("SET GLOBAL slow_query_log = 'ON'").Error; 局_错误 != nil {
		return "", "", "", 局_错误
	}
	// 设置慢查询阈值为 1 秒
	if 局_错误 := 数据库.Exec("SET GLOBAL long_query_time = 1").Error; 局_错误 != nil {
		return "", "", "", 局_错误
	}

	// 查询慢日志文件路径
	var 局_变量 struct {
		VariableName string `gorm:"column:Variable_name"`
		Value        string `gorm:"column:Value"`
	}
	if 局_错误 := 数据库.Raw("SHOW VARIABLES LIKE 'slow_query_log_file'").Scan(&局_变量).Error; 局_错误 != nil {
		return "", "", "", 局_错误
	}
	慢日志文件 = 局_变量.Value

	// 查询慢日志状态
	var 局_状态变量 struct {
		VariableName string `gorm:"column:Variable_name"`
		Value        string `gorm:"column:Value"`
	}
	if 局_错误 := 数据库.Raw("SHOW VARIABLES LIKE 'slow_query_log'").Scan(&局_状态变量).Error; 局_错误 != nil {
		return 慢日志文件, "", "", 局_错误
	}
	慢日志状态 = 局_状态变量.Value

	// 查询长查询时间
	var 局_时间变量 struct {
		VariableName string `gorm:"column:Variable_name"`
		Value        string `gorm:"column:Value"`
	}
	if 局_错误 := 数据库.Raw("SHOW VARIABLES LIKE 'long_query_time'").Scan(&局_时间变量).Error; 局_错误 != nil {
		return 慢日志文件, 慢日志状态, "", 局_错误
	}
	长查询时间 = 局_时间变量.Value

	return 慢日志文件, 慢日志状态, 长查询时间, nil
}

// 查询SQL摘要 查询 performance_schema.events_statements_summary_by_digest 取 Top 10
func (j *监控器) 查询SQL摘要(数据库 *gorm.DB) ([]响应_MySQL摘要, error) {
	局_行集, 局_错误 := 数据库.Raw(`SELECT
		SCHEMA_NAME, DIGEST_TEXT, COUNT_STAR,
		SUM_TIMER_WAIT, MIN_TIMER_WAIT, AVG_TIMER_WAIT, MAX_TIMER_WAIT,
		SUM_LOCK_TIME, SUM_ERRORS, SUM_WARNINGS,
		SUM_ROWS_AFFECTED, SUM_ROWS_SENT, SUM_ROWS_EXAMINED,
		SUM_CREATED_TMP_DISK_TABLES, SUM_CREATED_TMP_TABLES,
		SUM_SELECT_FULL_JOIN, SUM_SELECT_FULL_RANGE_JOIN,
		SUM_SELECT_RANGE, SUM_SELECT_SCAN,
		SUM_NO_INDEX_USED, FIRST_SEEN, LAST_SEEN
	FROM performance_schema.events_statements_summary_by_digest
	ORDER BY SUM_TIMER_WAIT DESC
	LIMIT 10`).Rows()
	if 局_错误 != nil {
		return nil, 局_错误
	}
	defer 局_行集.Close()

	局_结果 := make([]响应_MySQL摘要, 0, 10)
	for 局_行集.Next() {
		var 局_行 struct {
			SCHEMA_NAME            *string
			DIGEST_TEXT            *string
			COUNT_STAR             uint64
			SUM_TIMER_WAIT         uint64
			MIN_TIMER_WAIT         uint64
			AVG_TIMER_WAIT         uint64
			MAX_TIMER_WAIT         uint64
			SUM_LOCK_TIME          uint64
			SUM_ERRORS             uint64
			SUM_WARNINGS           uint64
			SUM_ROWS_AFFECTED      uint64
			SUM_ROWS_SENT          uint64
			SUM_ROWS_EXAMINED      uint64
			SUM_CREATED_TMP_DISK_TABLES uint64
			SUM_CREATED_TMP_TABLES uint64
			SUM_SELECT_FULL_JOIN   uint64
			SUM_SELECT_FULL_RANGE_JOIN uint64
			SUM_SELECT_RANGE       uint64
			SUM_SELECT_SCAN        uint64
			SUM_NO_INDEX_USED      uint64
			FIRST_SEEN             *string
			LAST_SEEN              *string
		}
		if 局_扫描错误 := 数据库.ScanRows(局_行集, &局_行); 局_扫描错误 != nil {
			continue
		}
		// MySQL performance_schema 计时单位为皮秒(ps)，1ms = 1,000,000 ps
		局_摘要 := 响应_MySQL摘要{
			SchemaName:       取可空字符串(局_行.SCHEMA_NAME),
			DigestText:       截断文本(取可空字符串(局_行.DIGEST_TEXT), 500),
			CountStar:        局_行.COUNT_STAR,
			SumTimerMs:       皮秒转毫秒(局_行.SUM_TIMER_WAIT),
			MinTimerMs:       皮秒转毫秒(局_行.MIN_TIMER_WAIT),
			AvgTimerMs:       皮秒转毫秒(局_行.AVG_TIMER_WAIT),
			MaxTimerMs:       皮秒转毫秒(局_行.MAX_TIMER_WAIT),
			SumLockMs:        皮秒转毫秒(局_行.SUM_LOCK_TIME),
			SumErrors:        局_行.SUM_ERRORS,
			SumWarnings:      局_行.SUM_WARNINGS,
			SumRowsAffected:  局_行.SUM_ROWS_AFFECTED,
			SumRowsSent:      局_行.SUM_ROWS_SENT,
			SumRowsExamined:  局_行.SUM_ROWS_EXAMINED,
			SumTmpDiskTables: 局_行.SUM_CREATED_TMP_DISK_TABLES,
			SumTmpTables:     局_行.SUM_CREATED_TMP_TABLES,
			SumFullJoin:      局_行.SUM_SELECT_FULL_JOIN,
			SumFullRangeJoin: 局_行.SUM_SELECT_FULL_RANGE_JOIN,
			SumSelectRange:   局_行.SUM_SELECT_RANGE,
			SumSelectScan:    局_行.SUM_SELECT_SCAN,
			SumNoIndexUsed:    局_行.SUM_NO_INDEX_USED,
			FirstSeen:        取可空字符串(局_行.FIRST_SEEN),
			LastSeen:         取可空字符串(局_行.LAST_SEEN),
		}
		局_摘要.Level = 评估摘要级别(局_摘要)
		局_摘要.Advice = 构建摘要建议(局_摘要)
		局_结果 = append(局_结果, 局_摘要)
	}

	return 局_结果, nil
}

// 评估进程级别 根据 Command/Time/State 判断风险
func 评估进程级别(进程 响应_MySQL进程) string {
	if 进程.Command == "Query" && 进程.Time >= 10 {
		return "danger"
	}
	if 进程.Command == "Query" && 进程.Time >= 3 {
		return "warning"
	}
	if strings.Contains(strings.ToLower(进程.State), "sending data") ||
		strings.Contains(strings.ToLower(进程.State), "copying to tmp table") ||
		strings.Contains(strings.ToLower(进程.State), "sorting result") {
		if 进程.Time >= 5 {
			return "danger"
		}
		return "warning"
	}
	if 进程.Command == "Sleep" && 进程.Time >= 30 {
		return "warning"
	}
	return "normal"
}

// 评估摘要级别 根据耗时与扫描行数判断风险
func 评估摘要级别(摘要 响应_MySQL摘要) string {
	// 超过 10 秒的累计耗时直接标红
	if 摘要.SumTimerMs >= 10000 {
		return "danger"
	}
	// 单次最大超过 1 秒
	if 摘要.MaxTimerMs >= 1000 {
		return "danger"
	}
	// 用到磁盘临时表
	if 摘要.SumTmpDiskTables > 0 {
		return "warning"
	}
	// 没有使用索引
	if 摘要.SumNoIndexUsed > 0 {
		return "warning"
	}
	// 扫描行数远大于发送行数（低效查询）
	if 摘要.SumRowsExamined > 0 && 摘要.SumRowsSent > 0 {
		if 摘要.SumRowsExamined/摘要.SumRowsSent >= 100 {
			return "warning"
		}
	}
	return "normal"
}

// 构建摘要建议 给单条 SQL 摘要生成人话解释
func 构建摘要建议(摘要 响应_MySQL摘要) string {
	var 局_说明 []string
	if 摘要.SumErrors > 0 {
		局_说明 = append(局_说明, fmt.Sprintf("出现 %d 次错误", 摘要.SumErrors))
	}
	if 摘要.SumTmpDiskTables > 0 {
		局_说明 = append(局_说明, fmt.Sprintf("生成 %d 次磁盘临时表，说明结果集过大或排序/分组未走索引", 摘要.SumTmpDiskTables))
	}
	if 摘要.SumFullJoin > 0 {
		局_说明 = append(局_说明, fmt.Sprintf("出现 %d 次全表连接(join 无索引)", 摘要.SumFullJoin))
	}
	if 摘要.SumNoIndexUsed > 0 {
		局_说明 = append(局_说明, fmt.Sprintf("有 %d 次未使用索引", 摘要.SumNoIndexUsed))
	}
	if 摘要.SumRowsExamined > 0 && 摘要.SumRowsSent > 0 && 摘要.SumRowsExamined/摘要.SumRowsSent >= 100 {
		局_说明 = append(局_说明, fmt.Sprintf("扫描 %d 行仅发送 %d 行，扫描/发送比超过 100 倍，建议检查 WHERE 或 LIMIT", 摘要.SumRowsExamined, 摘要.SumRowsSent))
	}
	if 摘要.MaxTimerMs >= 1000 {
		局_说明 = append(局_说明, fmt.Sprintf("单次最大耗时 %.2f ms，已超过 1 秒", 摘要.MaxTimerMs))
	}
	if len(局_说明) == 0 {
		return "当前指标正常，暂无需要优先关注的问题。"
	}
	return strings.Join(局_说明, "；")
}

// 构建MySQL诊断说明 根据诊断结果生成顶部说明列表
func 构建MySQL诊断说明(结果 *响应_MySQL诊断) []string {
	局_说明 := make([]string, 0, 6)
	局_说明 = append(局_说明, fmt.Sprintf("进程总数 %d，其中 Sleep %d、活跃 %d。", 结果.J进程总数, 结果.S睡眠进程数, 结果.H活跃进程数))
	if 结果.H活跃进程数 > 0 {
		局_说明 = append(局_说明, "活跃进程表示当前有 SQL 正在执行，重点看 Command=Query 且 Time 较大的行。")
	} else {
		局_说明 = append(局_说明, "当前没有正在执行的 SQL，所有连接都处于 Sleep 状态。")
	}
	if 结果.M慢日志状态 != "" {
		局_说明 = append(局_说明, fmt.Sprintf("慢查询日志已开启(状态:%s)，阈值 %s 秒，日志文件：%s。", 结果.M慢日志状态, 结果.C长查询时间, 结果.M慢日志文件))
		局_说明 = append(局_说明, "在服务器上执行 `tail -f "+结果.M慢日志文件+"` 可以实时观察慢 SQL。")
	} else {
		局_说明 = append(局_说明, "未能成功读取慢日志配置，可能是当前账号无 SUPER 权限。")
	}
	if len(结果.Z摘要排行) > 0 {
		局_说明 = append(局_说明, "下表按累计耗时排序展示 Top 10 SQL 模板，红色表示高风险(单次超过 1 秒或累计超过 10 秒)。")
		局_说明 = append(局_说明, "计时单位已从皮秒换算为毫秒；扫描行数/发送行数比值过大说明查询效率低。")
	} else {
		局_说明 = append(局_说明, "performance_schema 未采集到 SQL 摘要，可能是当前库没有产生查询。")
	}
	return 局_说明
}

// 皮秒转毫秒 MySQL performance_schema 计时单位为皮秒(ps)
func 皮秒转毫秒(皮秒 uint64) float64 {
	return 保留两位小数(float64(皮秒) / 1_000_000)
}

// 取可空字符串 处理 *string 可空字段
func 取可空字符串(值 *string) string {
	if 值 == nil {
		return ""
	}
	return *值
}
