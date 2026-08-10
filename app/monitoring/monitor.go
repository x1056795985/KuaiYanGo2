package monitoring

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	rpprof "runtime/pprof"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	gprofile "github.com/google/pprof/profile"
	gprocess "github.com/shirou/gopsutil/v3/process"

	"server/app/utils"
)

const (
	零_最大近期耗时数量  = 120
	零_最大慢请求数量   = 80
	零_最大恐慌数量    = 30
	零_最大慢SQL数量  = 80
	零_最大数据库错误数量 = 50
	零_画像文本最大字节  = 128 * 1024
	零_默认慢请求毫秒   = 1200
	零_默认慢SQL毫秒  = 200
	零_block画像速率 = 1_000_000
	零_mutex画像比例 = 5
	零_最大进程排行数量  = 10
	零_路由趋势采样秒数  = 10
	零_最大路由趋势点数  = 360
	零_最大路由趋势数量  = 8
	零_告警高CPU阈值  = 85
	零_告警高内存阈值   = 85
	零_告警高磁盘阈值   = 90
	零_告警高P95毫秒  = 1500
	零_告警活动请求阈值  = 20
)

var 零_进程CPU采样等待 = 800 * time.Millisecond

var Q监控 = C初始化监控器()

type 监控器 struct {
	启动时间    time.Time
	下一个请求Id atomic.Uint64

	互斥锁       sync.RWMutex
	路由表       map[string]*监控_路由统计
	活动请求      map[uint64]*监控_活动请求
	慢请求       []监控_请求事件
	恐慌事件      []监控_恐慌事件
	慢SQL      []监控_慢SQL记录
	数据库错误     []监控_数据库错误记录
	路由趋势历史    map[string][]监控_路由趋势点
	路由趋势基线    map[string]监控_路由趋势基线
	block画像速率 int
	mutex画像比例 int

	cpu互斥锁 sync.Mutex

	正在CPU画像   bool
	最近CPU画像   []byte
	最近CPU画像时间 time.Time
	最近CPU画像秒数 int
	最近CPU热点   []响应_CPU热点
}

type 监控_路由统计 struct {
	互斥锁 sync.Mutex

	方法 string
	路由 string

	次数     uint64
	错误次数   uint64
	恐慌次数   uint64
	总耗时纳秒  int64
	最大耗时纳秒 int64
	最近耗时纳秒 int64
	最近状态码  int
	最近访问时间 time.Time
	进行中数量  int64

	近期耗时列表 [零_最大近期耗时数量]int64
	近期索引   int
	近期数量   int
}

type 监控_活动请求 struct {
	请求Id  uint64
	方法    string
	路由    string
	路径    string
	查询    string
	客户端IP string
	开始时间  time.Time
}

type 监控_路由趋势点 struct {
	C采样时间 time.Time
	C次数   uint64
	C错误次数 uint64
	P平均毫秒 float64
	P95毫秒 float64
}

type 监控_路由趋势基线 struct {
	F方法    string
	L路由    string
	C次数    uint64
	C错误次数  uint64
	Z总耗时纳秒 int64
	P95毫秒  float64
}

type 监控_请求事件 struct {
	I请求Id  uint64  `json:"id"`
	F方法    string  `json:"method"`
	L路由    string  `json:"route"`
	J路径    string  `json:"path"`
	C查询    string  `json:"query"`
	K客户端IP string  `json:"clientIp"`
	Z状态码   int     `json:"status"`
	H耗时毫秒  float64 `json:"durationMs"`
	K开始时间  string  `json:"startedAt"`
	J结束时间  string  `json:"finishedAt"`
}

type 监控_恐慌事件 struct {
	F方法    string `json:"method"`
	L路由    string `json:"route"`
	J路径    string `json:"path"`
	C查询    string `json:"query"`
	K客户端IP string `json:"clientIp"`
	C错误    string `json:"error"`
	Z栈     string `json:"stack"`
	C创建时间  string `json:"createdAt"`
}

type 响应_路由指标 struct {
	F方法     string  `json:"method"`
	L路由     string  `json:"route"`
	C次数     uint64  `json:"count"`
	C错误次数   uint64  `json:"errorCount"`
	K恐慌次数   uint64  `json:"panicCount"`
	J进行中数量  int64   `json:"inFlight"`
	P平均毫秒   float64 `json:"avgMs"`
	P95毫秒   float64 `json:"p95Ms"`
	Z最大毫秒   float64 `json:"maxMs"`
	Z最近毫秒   float64 `json:"lastMs"`
	Z总毫秒    float64 `json:"totalMs"`
	Z最近状态码  int     `json:"lastStatus"`
	Z最近访问时间 string  `json:"lastSeenAt"`
}

type 响应_活动请求快照 struct {
	I请求Id   uint64  `json:"id"`
	F方法     string  `json:"method"`
	L路由     string  `json:"route"`
	J路径     string  `json:"path"`
	C查询     string  `json:"query"`
	K客户端IP  string  `json:"clientIp"`
	K开始时间   string  `json:"startedAt"`
	D当前耗时毫秒 float64 `json:"currentDurationMs"`
}

type 响应_运行时快照 struct {
	K开始时间       string  `json:"startedAt"`
	Y运行秒数       int64   `json:"uptimeSeconds"`
	G协程数        int     `json:"goroutines"`
	GGomaxprocs int     `json:"goMaxProcs"`
	NNumCpu     int     `json:"numCpu"`
	NNumCgoCall int64   `json:"numCgoCall"`
	T线程创建数      int     `json:"threadCreateCount"`
	A已分配MB      float64 `json:"allocMb"`
	T总分配MB      float64 `json:"totalAllocMb"`
	S系统MB       float64 `json:"sysMb"`
	D堆分配MB      float64 `json:"heapAllocMb"`
	D堆使用MB      float64 `json:"heapInuseMb"`
	D堆空闲MB      float64 `json:"heapIdleMb"`
	D堆释放MB      float64 `json:"heapReleasedMb"`
	D堆对象数       uint64  `json:"heapObjects"`
	Z栈使用MB      float64 `json:"stackInuseMb"`
	X下次GCMB     float64 `json:"nextGcMb"`
	Z最近GC暂停毫秒   float64 `json:"lastGcPauseMs"`
	Z总暂停毫秒      float64 `json:"pauseTotalMs"`
	GGC次数       uint32  `json:"numGc"`
	GGCCpu占比    float64 `json:"gcCpuFraction"`
	BBlock已开启   bool    `json:"blockProfileEnabled"`
	BBlock速率    int     `json:"blockProfileRate"`
	MMutex已开启   bool    `json:"mutexProfileEnabled"`
	MMutex比例    int     `json:"mutexProfileFraction"`
	CCPU画像进行中   bool    `json:"cpuProfileRunning"`
	Z最近CPU画像时间  string  `json:"lastCpuProfileAt"`
	Z最近CPU画像秒数  int     `json:"lastCpuProfileSeconds"`
}

type 响应_画像信息 struct {
	M名称   string `json:"name"`
	S数量   int    `json:"count"`
	S手动开关 bool   `json:"manualToggle"`
	Y已开启  bool   `json:"enabled"`
	M描述   string `json:"description"`
}

type 响应_Pprof状态 struct {
	H画像列表    []响应_画像信息  `json:"profiles"`
	Z最近CPU热点 []响应_CPU热点 `json:"lastCpuTop"`
}

type 响应_数据库连接池 struct {
	D打开连接数     int     `json:"openConnections"`
	Z正在使用数     int     `json:"inUse"`
	K空闲连接数     int     `json:"idle"`
	D等待次数      int64   `json:"waitCount"`
	D等待毫秒      float64 `json:"waitDurationMs"`
	Z最大空闲关闭数   int64   `json:"maxIdleClosed"`
	Z最大空闲时长关闭数 int64   `json:"maxIdleTimeClosed"`
	Z最大生命周期关闭数 int64   `json:"maxLifetimeClosed"`
}

type 监控_慢SQL记录 struct {
	S时间    string  `json:"at"`
	H耗时毫秒  float64 `json:"durationMs"`
	Y影响行数  int64   `json:"rowsAffected"`
	SSQL语句 string  `json:"sql"`
}

type 监控_数据库错误记录 struct {
	S时间    string  `json:"at"`
	H耗时毫秒  float64 `json:"durationMs"`
	Y影响行数  int64   `json:"rowsAffected"`
	S死锁    bool    `json:"deadlock"`
	C错误    string  `json:"error"`
	SSQL语句 string  `json:"sql"`
}

type 响应_数据库总览 struct {
	S是否启用     bool         `json:"enabled"`
	M慢SQL阈值毫秒 float64      `json:"slowThresholdMs"`
	L连接池      响应_数据库连接池    `json:"pool"`
	M慢SQL列表   []监控_慢SQL记录  `json:"slowSqls"`
	C错误列表     []监控_数据库错误记录 `json:"errors"`
	M模板排行     []响应_SQL模板聚合 `json:"sqlTemplates"`
}

type 响应_监控总览 struct {
	F服务      *utils.Server `json:"server"`
	Y运行时     响应_运行时快照      `json:"runtime"`
	S数据库     响应_数据库总览      `json:"database"`
	L路由列表    []响应_路由指标     `json:"routes"`
	L路由趋势列表  []响应_路由趋势     `json:"routeTrends"`
	H活动请求    []响应_活动请求快照   `json:"activeRequests"`
	M慢请求列表   []监控_请求事件     `json:"slowRequests"`
	K恐慌事件列表  []监控_恐慌事件     `json:"panicEvents"`
	G告警列表    []响应_告警项      `json:"alerts"`
	PPprof状态 响应_Pprof状态    `json:"pprof"`
	S说明      []string      `json:"notes"`
}

type 响应_画像文本 struct {
	M名称   string `json:"name"`
	D调试等级 int    `json:"debug"`
	S是否截断 bool   `json:"truncated"`
	C采集时间 string `json:"collectedAt"`
	W文本   string `json:"text"`
}

type 响应_CPU热点 struct {
	H函数名  string  `json:"function"`
	P平耗毫秒 float64 `json:"flatMs"`
	P平耗占比 float64 `json:"flatPercent"`
	L累计毫秒 float64 `json:"cumulativeMs"`
	L累计占比 float64 `json:"cumulativePercent"`
}

type 响应_CPU画像结果 struct {
	S持续秒数 int        `json:"durationSeconds"`
	C采集时间 string     `json:"capturedAt"`
	R热点列表 []响应_CPU热点 `json:"top"`
}

type 响应_进程指标 struct {
	J进程ID  int32   `json:"pid"`
	M名称    string  `json:"name"`
	C命令行   string  `json:"command"`
	CCPU占比 float64 `json:"cpuPercent"`
	N内存MB  float64 `json:"memoryMb"`
	N内存占比  float64 `json:"memoryPercent"`
	X线程数   int32   `json:"threadCount"`
	Z状态    string  `json:"status"`
	Q启动时间  string  `json:"startedAt"`
	Y运行秒数  int64   `json:"uptimeSeconds"`
}

type 响应_进程排行 struct {
	C采集时间 string    `json:"collectedAt"`
	J进程列表 []响应_进程指标 `json:"processes"`
}

type 响应_路由趋势点 struct {
	S时间   string  `json:"at"`
	C次数   uint64  `json:"count"`
	C错误次数 uint64  `json:"errorCount"`
	P平均毫秒 float64 `json:"avgMs"`
	P95毫秒 float64 `json:"p95Ms"`
}

type 响应_路由趋势 struct {
	F方法  string     `json:"method"`
	L路由  string     `json:"route"`
	D点列表 []响应_路由趋势点 `json:"points"`
}

type 响应_告警项 struct {
	J级别 string `json:"level"`
	B标题 string `json:"title"`
	X详情 string `json:"detail"`
}

type 响应_SQL模板聚合 struct {
	M模板   string  `json:"template"`
	C次数   int     `json:"count"`
	P平均毫秒 float64 `json:"avgMs"`
	Z最大毫秒 float64 `json:"maxMs"`
	S死锁次数 int     `json:"deadlockCount"`
	Z最近时间 string  `json:"lastAt"`
}

type 进程_初始采样 struct {
	J进程    *gprocess.Process
	CCPU总秒 float64
}

func C初始化监控器() *监控器 {
	局_监控器 := &监控器{
		启动时间:   time.Now(),
		路由表:    make(map[string]*监控_路由统计),
		活动请求:   make(map[uint64]*监控_活动请求),
		路由趋势历史: make(map[string][]监控_路由趋势点),
		路由趋势基线: make(map[string]监控_路由趋势基线),
	}
	go 局_监控器.路由趋势采样循环()
	return 局_监控器
}

func (j *监控器) Q监控中间件() gin.HandlerFunc {
	return func(c *gin.Context) {
		局_开始时间 := time.Now()
		局_请求Id := j.下一个请求Id.Add(1)
		局_路由 := 归一化路由(c.FullPath(), c.Request.URL.Path)

		局_统计 := j.取路由统计(c.Request.Method, 局_路由)
		局_统计.增加进行中(1)

		j.加入活动请求(&监控_活动请求{
			请求Id:  局_请求Id,
			方法:    c.Request.Method,
			路由:    局_路由,
			路径:    c.Request.URL.Path,
			查询:    c.Request.URL.RawQuery,
			客户端IP: c.ClientIP(),
			开始时间:  局_开始时间,
		})

		defer func() {
			局_最终路由 := 归一化路由(c.FullPath(), 局_路由)
			局_耗时 := time.Since(局_开始时间)
			if 局_恢复值 := recover(); 局_恢复值 != nil {
				局_统计.完成(局_耗时, 500, true, true)
				j.移除活动请求(局_请求Id)
				j.追加慢请求(监控_请求事件{
					I请求Id:  局_请求Id,
					F方法:    c.Request.Method,
					L路由:    局_最终路由,
					J路径:    c.Request.URL.Path,
					C查询:    c.Request.URL.RawQuery,
					K客户端IP: c.ClientIP(),
					Z状态码:   500,
					H耗时毫秒:  时长转毫秒(局_耗时),
					K开始时间:  局_开始时间.Format(time.RFC3339),
					J结束时间:  time.Now().Format(time.RFC3339),
				}, true)
				panic(局_恢复值)
			}

			局_状态码 := c.Writer.Status()
			局_是否错误 := 局_状态码 >= 500 || len(c.Errors) > 0
			局_统计.完成(局_耗时, 局_状态码, 局_是否错误, false)
			j.移除活动请求(局_请求Id)
			j.追加慢请求(监控_请求事件{
				I请求Id:  局_请求Id,
				F方法:    c.Request.Method,
				L路由:    局_最终路由,
				J路径:    c.Request.URL.Path,
				C查询:    c.Request.URL.RawQuery,
				K客户端IP: c.ClientIP(),
				Z状态码:   局_状态码,
				H耗时毫秒:  时长转毫秒(局_耗时),
				K开始时间:  局_开始时间.Format(time.RFC3339),
				J结束时间:  time.Now().Format(time.RFC3339),
			}, false)
		}()

		c.Next()
	}
}

func (j *监控器) J记录恐慌(c *gin.Context, 恢复值 interface{}, 栈信息 string) {
	if 恢复值 == nil {
		return
	}

	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()

	j.恐慌事件 = 追加有界切片(j.恐慌事件, 监控_恐慌事件{
		F方法:    安全方法(c),
		L路由:    归一化路由(c.FullPath(), 安全路径(c)),
		J路径:    安全路径(c),
		C查询:    安全查询(c),
		K客户端IP: 安全客户端IP(c),
		C错误:    fmt.Sprint(恢复值),
		Z栈:     栈信息,
		C创建时间:  time.Now().Format(time.RFC3339),
	}, 零_最大恐慌数量)
}

func (j *监控器) Q监控总览() (*响应_监控总览, error) {
	局_服务, 局_错误 := 采集服务信息()
	if 局_错误 != nil {
		return nil, 局_错误
	}

	局_运行时快照 := j.取运行时快照()
	局_数据库快照 := j.取数据库总览()

	j.互斥锁.RLock()
	局_路由统计列表 := make([]*监控_路由统计, 0, len(j.路由表))
	for _, 局_统计 := range j.路由表 {
		局_路由统计列表 = append(局_路由统计列表, 局_统计)
	}

	局_活动请求列表 := make([]*监控_活动请求, 0, len(j.活动请求))
	for _, 局_请求 := range j.活动请求 {
		局_活动请求列表 = append(局_活动请求列表, 局_请求)
	}

	局_慢请求列表 := append([]监控_请求事件(nil), j.慢请求...)
	局_恐慌事件列表 := append([]监控_恐慌事件(nil), j.恐慌事件...)
	局_最近CPU热点 := append([]响应_CPU热点(nil), j.最近CPU热点...)
	局_block画像速率 := j.block画像速率
	局_mutex画像比例 := j.mutex画像比例
	j.互斥锁.RUnlock()

	局_路由指标列表 := make([]响应_路由指标, 0, len(局_路由统计列表))
	for _, 局_统计 := range 局_路由统计列表 {
		局_路由指标列表 = append(局_路由指标列表, 局_统计.转快照())
	}
	sort.Slice(局_路由指标列表, func(i, k int) bool {
		if 局_路由指标列表[i].Z总毫秒 == 局_路由指标列表[k].Z总毫秒 {
			return 局_路由指标列表[i].C次数 > 局_路由指标列表[k].C次数
		}
		return 局_路由指标列表[i].Z总毫秒 > 局_路由指标列表[k].Z总毫秒
	})

	局_当前时间 := time.Now()
	局_活动请求快照 := make([]响应_活动请求快照, 0, len(局_活动请求列表))
	for _, 局_请求 := range 局_活动请求列表 {
		局_活动请求快照 = append(局_活动请求快照, 响应_活动请求快照{
			I请求Id:   局_请求.请求Id,
			F方法:     局_请求.方法,
			L路由:     局_请求.路由,
			J路径:     局_请求.路径,
			C查询:     局_请求.查询,
			K客户端IP:  局_请求.客户端IP,
			K开始时间:   局_请求.开始时间.Format(time.RFC3339),
			D当前耗时毫秒: 时长转毫秒(局_当前时间.Sub(局_请求.开始时间)),
		})
	}
	sort.Slice(局_活动请求快照, func(i, k int) bool {
		return 局_活动请求快照[i].D当前耗时毫秒 > 局_活动请求快照[k].D当前耗时毫秒
	})

	sort.Slice(局_慢请求列表, func(i, k int) bool {
		return 局_慢请求列表[i].H耗时毫秒 > 局_慢请求列表[k].H耗时毫秒
	})
	sort.Slice(局_恐慌事件列表, func(i, k int) bool {
		return 局_恐慌事件列表[i].C创建时间 > 局_恐慌事件列表[k].C创建时间
	})

	局_路由趋势列表 := j.取路由趋势列表(局_路由指标列表)
	局_告警列表 := 构建告警列表(局_服务, 局_运行时快照, 局_数据库快照, 局_路由指标列表, 局_活动请求快照, 局_慢请求列表)

	return &响应_监控总览{
		F服务:     局_服务,
		Y运行时:    局_运行时快照,
		S数据库:    局_数据库快照,
		L路由列表:   局_路由指标列表,
		L路由趋势列表: 局_路由趋势列表,
		H活动请求:   局_活动请求快照,
		M慢请求列表:  局_慢请求列表,
		K恐慌事件列表: 局_恐慌事件列表,
		G告警列表:   局_告警列表,
		PPprof状态: 响应_Pprof状态{
			H画像列表: []响应_画像信息{
				构造画像信息("goroutine", false, true, "协程栈，适合排查卡死和阻塞"),
				构造画像信息("heap", false, true, "当前堆内存快照"),
				构造画像信息("allocs", false, true, "历史分配视角"),
				构造画像信息("threadcreate", false, true, "系统线程创建情况"),
				构造画像信息("block", true, 局_block画像速率 > 0, "阻塞采样，需要手动开启"),
				构造画像信息("mutex", true, 局_mutex画像比例 > 0, "锁竞争采样，需要手动开启"),
			},
			Z最近CPU热点: 局_最近CPU热点,
		},
		S说明: []string{
			"路由排行反映的是请求耗时热点，不等同于精确 CPU 或内存归因。",
			"CPU 热点建议在复现问题时手动抓取 10 到 30 秒 CPU Profile。",
			"block 和 mutex profile 默认关闭，避免长期增加生产环境采样开销。",
		},
	}, nil
}

func (j *监控器) 路由趋势采样循环() {
	j.采样路由趋势()

	局_定时器 := time.NewTicker(零_路由趋势采样秒数 * time.Second)
	defer 局_定时器.Stop()

	for range 局_定时器.C {
		j.采样路由趋势()
	}
}

func (j *监控器) 采样路由趋势() {
	j.互斥锁.RLock()
	局_路由统计列表 := make([]*监控_路由统计, 0, len(j.路由表))
	for _, 局_统计 := range j.路由表 {
		局_路由统计列表 = append(局_路由统计列表, 局_统计)
	}
	j.互斥锁.RUnlock()

	局_当前时间 := time.Now()
	局_当前快照映射 := make(map[string]监控_路由趋势基线, len(局_路由统计列表))
	for _, 局_统计 := range 局_路由统计列表 {
		局_快照 := 局_统计.取趋势基线()
		局_当前快照映射[局_快照.F方法+" "+局_快照.L路由] = 局_快照
	}

	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()

	for 局_键, 局_当前快照 := range 局_当前快照映射 {
		局_上次快照, 局_存在 := j.路由趋势基线[局_键]
		if 局_存在 && 局_当前快照.C次数 >= 局_上次快照.C次数 && 局_当前快照.C错误次数 >= 局_上次快照.C错误次数 && 局_当前快照.Z总耗时纳秒 >= 局_上次快照.Z总耗时纳秒 {
			局_新增次数 := 局_当前快照.C次数 - 局_上次快照.C次数
			局_新增错误次数 := 局_当前快照.C错误次数 - 局_上次快照.C错误次数
			局_新增总耗时纳秒 := 局_当前快照.Z总耗时纳秒 - 局_上次快照.Z总耗时纳秒
			局_平均毫秒 := 0.0
			if 局_新增次数 > 0 {
				局_平均毫秒 = 保留两位小数(时长转毫秒(time.Duration(局_新增总耗时纳秒 / int64(局_新增次数))))
			}
			j.路由趋势历史[局_键] = 追加有界切片(j.路由趋势历史[局_键], 监控_路由趋势点{
				C采样时间: 局_当前时间,
				C次数:   局_新增次数,
				C错误次数: 局_新增错误次数,
				P平均毫秒: 局_平均毫秒,
				P95毫秒: 局_当前快照.P95毫秒,
			}, 零_最大路由趋势点数)
		}

		j.路由趋势基线[局_键] = 局_当前快照
	}
}

func (j *监控器) 取路由趋势列表(路由列表 []响应_路由指标) []响应_路由趋势 {
	if len(路由列表) == 0 {
		return []响应_路由趋势{}
	}

	局_上限 := 零_最大路由趋势数量
	if len(路由列表) < 局_上限 {
		局_上限 = len(路由列表)
	}

	j.互斥锁.RLock()
	defer j.互斥锁.RUnlock()

	局_结果 := make([]响应_路由趋势, 0, 局_上限)
	for _, 局_路由 := range 路由列表[:局_上限] {
		局_键 := 局_路由.F方法 + " " + 局_路由.L路由
		局_历史 := append([]监控_路由趋势点(nil), j.路由趋势历史[局_键]...)
		局_点列表 := make([]响应_路由趋势点, 0, len(局_历史))
		for _, 局_点 := range 局_历史 {
			局_点列表 = append(局_点列表, 响应_路由趋势点{
				S时间:   局_点.C采样时间.Format(time.RFC3339),
				C次数:   局_点.C次数,
				C错误次数: 局_点.C错误次数,
				P平均毫秒: 局_点.P平均毫秒,
				P95毫秒: 局_点.P95毫秒,
			})
		}
		局_结果 = append(局_结果, 响应_路由趋势{
			F方法:  局_路由.F方法,
			L路由:  局_路由.L路由,
			D点列表: 局_点列表,
		})
	}
	return 局_结果
}

func (j *监控器) G更新Pprof设置(开启Block bool, block速率 int, 开启Mutex bool, mutex比例 int) 响应_运行时快照 {
	if !开启Block {
		block速率 = 0
	} else if block速率 <= 0 {
		block速率 = 零_block画像速率
	}
	runtime.SetBlockProfileRate(block速率)

	if !开启Mutex {
		mutex比例 = 0
	} else if mutex比例 <= 0 {
		mutex比例 = 零_mutex画像比例
	}
	runtime.SetMutexProfileFraction(mutex比例)

	j.互斥锁.Lock()
	j.block画像速率 = block速率
	j.mutex画像比例 = mutex比例
	j.互斥锁.Unlock()

	return j.取运行时快照()
}

func (j *监控器) Q画像文本(名称 string, 调试等级 int, 采集前GC bool) (*响应_画像文本, error) {
	if !是否支持文本画像(名称) {
		return nil, fmt.Errorf("unsupported pprof profile: %s", 名称)
	}
	if 采集前GC && (名称 == "heap" || 名称 == "allocs") {
		runtime.GC()
	}
	if 调试等级 <= 0 {
		调试等级 = 1
	}

	局_画像 := rpprof.Lookup(名称)
	if 局_画像 == nil {
		return nil, fmt.Errorf("profile not found: %s", 名称)
	}

	var 局_缓冲 bytes.Buffer
	if 局_错误 := 局_画像.WriteTo(&局_缓冲, 调试等级); 局_错误 != nil {
		return nil, 局_错误
	}

	局_文本 := 局_缓冲.String()
	局_是否截断 := false
	if len(局_文本) > 零_画像文本最大字节 {
		局_文本 = 局_文本[:零_画像文本最大字节]
		局_是否截断 = true
	}

	return &响应_画像文本{
		M名称:   名称,
		D调试等级: 调试等级,
		S是否截断: 局_是否截断,
		C采集时间: time.Now().Format(time.RFC3339),
		W文本:   局_文本,
	}, nil
}

func (j *监控器) Q下载画像(名称 string, 采集前GC bool) ([]byte, string, error) {
	if !是否支持下载画像(名称) {
		return nil, "", fmt.Errorf("unsupported download profile: %s", 名称)
	}
	if 采集前GC && (名称 == "heap" || 名称 == "allocs") {
		runtime.GC()
	}

	局_画像 := rpprof.Lookup(名称)
	if 局_画像 == nil {
		return nil, "", fmt.Errorf("profile not found: %s", 名称)
	}

	var 局_缓冲 bytes.Buffer
	if 局_错误 := 局_画像.WriteTo(&局_缓冲, 0); 局_错误 != nil {
		return nil, "", 局_错误
	}

	局_文件名 := fmt.Sprintf("%s-%s.pprof", 名称, time.Now().Format("20060102-150405"))
	return 局_缓冲.Bytes(), 局_文件名, nil
}

func (j *监控器) C抓取CPU画像(秒数 int) (*响应_CPU画像结果, error) {
	if 秒数 <= 0 {
		秒数 = 10
	}
	if 秒数 > 60 {
		秒数 = 60
	}

	j.cpu互斥锁.Lock()
	defer j.cpu互斥锁.Unlock()

	j.互斥锁.Lock()
	if j.正在CPU画像 {
		j.互斥锁.Unlock()
		return nil, errors.New("cpu profile is already running")
	}
	j.正在CPU画像 = true
	j.互斥锁.Unlock()

	var 局_缓冲 bytes.Buffer
	局_开始时间 := time.Now()
	if 局_错误 := rpprof.StartCPUProfile(&局_缓冲); 局_错误 != nil {
		j.互斥锁.Lock()
		j.正在CPU画像 = false
		j.互斥锁.Unlock()
		return nil, 局_错误
	}

	time.Sleep(time.Duration(秒数) * time.Second)
	rpprof.StopCPUProfile()

	局_热点列表, 局_错误 := 汇总CPU画像(局_缓冲.Bytes())
	if 局_错误 != nil {
		j.互斥锁.Lock()
		j.正在CPU画像 = false
		j.互斥锁.Unlock()
		return nil, 局_错误
	}

	j.互斥锁.Lock()
	j.正在CPU画像 = false
	j.最近CPU画像 = append([]byte(nil), 局_缓冲.Bytes()...)
	j.最近CPU画像时间 = 局_开始时间
	j.最近CPU画像秒数 = 秒数
	j.最近CPU热点 = 局_热点列表
	j.互斥锁.Unlock()

	return &响应_CPU画像结果{
		S持续秒数: 秒数,
		C采集时间: 局_开始时间.Format(time.RFC3339),
		R热点列表: 局_热点列表,
	}, nil
}

func (j *监控器) Q最近CPU画像() ([]byte, string, error) {
	j.互斥锁.RLock()
	defer j.互斥锁.RUnlock()

	if len(j.最近CPU画像) == 0 {
		return nil, "", errors.New("no cpu profile captured yet")
	}

	局_文件名 := fmt.Sprintf("cpu-%s-%ds.pprof", j.最近CPU画像时间.Format("20060102-150405"), j.最近CPU画像秒数)
	return append([]byte(nil), j.最近CPU画像...), 局_文件名, nil
}

func (j *监控器) Q进程排行() (*响应_进程排行, error) {
	局_进程列表, 局_错误 := gprocess.Processes()
	if 局_错误 != nil {
		return nil, 局_错误
	}

	局_当前进程ID := int32(os.Getpid())
	局_初始采样映射 := make(map[int32]进程_初始采样, len(局_进程列表))
	for _, 局_进程 := range 局_进程列表 {
		if 局_进程 == nil || 局_进程.Pid == 局_当前进程ID {
			continue
		}
		局_CPU总秒, 局_可用 := 取进程CPU总秒(局_进程)
		if !局_可用 {
			continue
		}
		局_初始采样映射[局_进程.Pid] = 进程_初始采样{
			J进程:    局_进程,
			CCPU总秒: 局_CPU总秒,
		}
	}

	局_开始时间 := time.Now()
	time.Sleep(零_进程CPU采样等待)
	局_经过秒数 := time.Since(局_开始时间).Seconds()
	if 局_经过秒数 <= 0 {
		局_经过秒数 = float64(零_进程CPU采样等待) / float64(time.Second)
	}

	局_结果列表 := make([]响应_进程指标, 0, len(局_初始采样映射))
	for _, 局_采样 := range 局_初始采样映射 {
		局_指标, 局_可用 := 构造进程指标(局_采样, 局_经过秒数)
		if !局_可用 {
			continue
		}
		局_结果列表 = append(局_结果列表, 局_指标)
	}

	sort.Slice(局_结果列表, func(i, k int) bool {
		if 局_结果列表[i].CCPU占比 == 局_结果列表[k].CCPU占比 {
			if 局_结果列表[i].N内存MB == 局_结果列表[k].N内存MB {
				return 局_结果列表[i].J进程ID < 局_结果列表[k].J进程ID
			}
			return 局_结果列表[i].N内存MB > 局_结果列表[k].N内存MB
		}
		return 局_结果列表[i].CCPU占比 > 局_结果列表[k].CCPU占比
	})

	if len(局_结果列表) > 零_最大进程排行数量 {
		局_结果列表 = 局_结果列表[:零_最大进程排行数量]
	}

	return &响应_进程排行{
		C采集时间: time.Now().Format(time.RFC3339),
		J进程列表: 局_结果列表,
	}, nil
}

func (j *监控器) 加入活动请求(请求 *监控_活动请求) {
	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()
	j.活动请求[请求.请求Id] = 请求
}

func (j *监控器) 移除活动请求(请求Id uint64) {
	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()
	delete(j.活动请求, 请求Id)
}

func (j *监控器) 追加慢请求(请求事件 监控_请求事件, 强制记录 bool) {
	if !强制记录 && 请求事件.H耗时毫秒 < 零_默认慢请求毫秒 {
		return
	}

	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()
	j.慢请求 = 追加有界切片(j.慢请求, 请求事件, 零_最大慢请求数量)
}

func (j *监控器) 取路由统计(方法 string, 路由 string) *监控_路由统计 {
	局_键 := 方法 + " " + 路由

	j.互斥锁.RLock()
	局_统计, 局_存在 := j.路由表[局_键]
	j.互斥锁.RUnlock()
	if 局_存在 {
		return 局_统计
	}

	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()
	if 局_统计, 局_存在 = j.路由表[局_键]; 局_存在 {
		return 局_统计
	}

	局_统计 = &监控_路由统计{
		方法: 方法,
		路由: 路由,
	}
	j.路由表[局_键] = 局_统计
	return 局_统计
}

func (j *监控器) 取运行时快照() 响应_运行时快照 {
	var 局_内存 runtime.MemStats
	runtime.ReadMemStats(&局_内存)

	j.互斥锁.RLock()
	局_block画像速率 := j.block画像速率
	局_mutex画像比例 := j.mutex画像比例
	局_cpu画像进行中 := j.正在CPU画像
	局_最近CPU画像时间 := j.最近CPU画像时间
	局_最近CPU画像秒数 := j.最近CPU画像秒数
	j.互斥锁.RUnlock()

	局_最近GC暂停毫秒 := 0.0
	if 局_内存.NumGC > 0 {
		局_最近GC暂停毫秒 = float64(局_内存.PauseNs[(局_内存.NumGC-1)%uint32(len(局_内存.PauseNs))]) / float64(time.Millisecond)
	}

	return 响应_运行时快照{
		K开始时间:       j.启动时间.Format(time.RFC3339),
		Y运行秒数:       int64(time.Since(j.启动时间).Seconds()),
		G协程数:        runtime.NumGoroutine(),
		GGomaxprocs: runtime.GOMAXPROCS(0),
		NNumCpu:     runtime.NumCPU(),
		NNumCgoCall: runtime.NumCgoCall(),
		T线程创建数:      取画像数量("threadcreate"),
		A已分配MB:      字节转MB(局_内存.Alloc),
		T总分配MB:      字节转MB(局_内存.TotalAlloc),
		S系统MB:       字节转MB(局_内存.Sys),
		D堆分配MB:      字节转MB(局_内存.HeapAlloc),
		D堆使用MB:      字节转MB(局_内存.HeapInuse),
		D堆空闲MB:      字节转MB(局_内存.HeapIdle),
		D堆释放MB:      字节转MB(局_内存.HeapReleased),
		D堆对象数:       局_内存.HeapObjects,
		Z栈使用MB:      字节转MB(局_内存.StackInuse),
		X下次GCMB:     字节转MB(局_内存.NextGC),
		Z最近GC暂停毫秒:   局_最近GC暂停毫秒,
		Z总暂停毫秒:      float64(局_内存.PauseTotalNs) / float64(time.Millisecond),
		GGC次数:       局_内存.NumGC,
		GGCCpu占比:    局_内存.GCCPUFraction,
		BBlock已开启:   局_block画像速率 > 0,
		BBlock速率:    局_block画像速率,
		MMutex已开启:   局_mutex画像比例 > 0,
		MMutex比例:    局_mutex画像比例,
		CCPU画像进行中:   局_cpu画像进行中,
		Z最近CPU画像时间:  格式化时间(局_最近CPU画像时间),
		Z最近CPU画像秒数:  局_最近CPU画像秒数,
	}
}

func (j *监控_路由统计) 增加进行中(增量 int64) {
	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()
	j.进行中数量 += 增量
}

func (j *监控_路由统计) 完成(耗时 time.Duration, 状态码 int, 是否错误 bool, 是否恐慌 bool) {
	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()

	局_耗时纳秒 := 耗时.Nanoseconds()
	j.次数++
	j.总耗时纳秒 += 局_耗时纳秒
	j.最近耗时纳秒 = 局_耗时纳秒
	j.最近状态码 = 状态码
	j.最近访问时间 = time.Now()
	if 局_耗时纳秒 > j.最大耗时纳秒 {
		j.最大耗时纳秒 = 局_耗时纳秒
	}
	if 是否错误 {
		j.错误次数++
	}
	if 是否恐慌 {
		j.恐慌次数++
	}
	j.进行中数量--

	j.近期耗时列表[j.近期索引] = 局_耗时纳秒
	j.近期索引 = (j.近期索引 + 1) % 零_最大近期耗时数量
	if j.近期数量 < 零_最大近期耗时数量 {
		j.近期数量++
	}
}

func (j *监控_路由统计) 转快照() 响应_路由指标 {
	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()

	局_平均毫秒 := 0.0
	if j.次数 > 0 {
		局_平均毫秒 = 时长转毫秒(time.Duration(j.总耗时纳秒 / int64(j.次数)))
	}

	局_近期耗时 := make([]int64, 0, j.近期数量)
	for 局_索引 := 0; 局_索引 < j.近期数量; 局_索引++ {
		局_近期耗时 = append(局_近期耗时, j.近期耗时列表[局_索引])
	}
	sort.Slice(局_近期耗时, func(i, k int) bool { return 局_近期耗时[i] < 局_近期耗时[k] })

	局_p95毫秒 := 0.0
	if len(局_近期耗时) > 0 {
		局_索引 := int(float64(len(局_近期耗时)-1) * 0.95)
		局_p95毫秒 = 时长转毫秒(time.Duration(局_近期耗时[局_索引]))
	}

	return 响应_路由指标{
		F方法:     j.方法,
		L路由:     j.路由,
		C次数:     j.次数,
		C错误次数:   j.错误次数,
		K恐慌次数:   j.恐慌次数,
		J进行中数量:  j.进行中数量,
		P平均毫秒:   保留两位小数(局_平均毫秒),
		P95毫秒:   保留两位小数(局_p95毫秒),
		Z最大毫秒:   保留两位小数(时长转毫秒(time.Duration(j.最大耗时纳秒))),
		Z最近毫秒:   保留两位小数(时长转毫秒(time.Duration(j.最近耗时纳秒))),
		Z总毫秒:    保留两位小数(时长转毫秒(time.Duration(j.总耗时纳秒))),
		Z最近状态码:  j.最近状态码,
		Z最近访问时间: 格式化时间(j.最近访问时间),
	}
}

func (j *监控_路由统计) 取趋势基线() 监控_路由趋势基线 {
	j.互斥锁.Lock()
	defer j.互斥锁.Unlock()

	局_p95毫秒 := 0.0
	if j.近期数量 > 0 {
		局_近期耗时 := make([]int64, 0, j.近期数量)
		for 局_索引 := 0; 局_索引 < j.近期数量; 局_索引++ {
			局_近期耗时 = append(局_近期耗时, j.近期耗时列表[局_索引])
		}
		sort.Slice(局_近期耗时, func(i, k int) bool { return 局_近期耗时[i] < 局_近期耗时[k] })
		局_位置 := int(float64(len(局_近期耗时)-1) * 0.95)
		局_p95毫秒 = 保留两位小数(时长转毫秒(time.Duration(局_近期耗时[局_位置])))
	}

	return 监控_路由趋势基线{
		F方法:    j.方法,
		L路由:    j.路由,
		C次数:    j.次数,
		C错误次数:  j.错误次数,
		Z总耗时纳秒: j.总耗时纳秒,
		P95毫秒:  局_p95毫秒,
	}
}

func 采集服务信息() (*utils.Server, error) {
	var (
		局_服务   utils.Server
		局_错误列表 []string
	)

	局_服务.Os = utils.InitOS()
	if 局_cpu信息, 局_错误 := utils.InitCPU(); 局_错误 == nil {
		局_服务.Cpu = 局_cpu信息
	} else {
		局_错误列表 = append(局_错误列表, 局_错误.Error())
	}
	if 局_内存信息, 局_错误 := utils.InitRAM(); 局_错误 == nil {
		局_服务.Ram = 局_内存信息
	} else {
		局_错误列表 = append(局_错误列表, 局_错误.Error())
	}
	if 局_磁盘信息, 局_错误 := utils.InitDisk(); 局_错误 == nil {
		局_服务.Disk = 局_磁盘信息
	} else {
		局_错误列表 = append(局_错误列表, 局_错误.Error())
	}

	if len(局_错误列表) > 0 {
		return &局_服务, errors.New(strings.Join(局_错误列表, "; "))
	}
	return &局_服务, nil
}

func 构建告警列表(服务 *utils.Server, 运行时 响应_运行时快照, 数据库 响应_数据库总览, 路由列表 []响应_路由指标, 活动请求 []响应_活动请求快照, 慢请求列表 []监控_请求事件) []响应_告警项 {
	局_结果 := make([]响应_告警项, 0, 8)
	if 服务 != nil {
		局_CPU峰值 := 0.0
		for _, 局_值 := range 服务.Cpu.Cpus {
			if 局_值 > 局_CPU峰值 {
				局_CPU峰值 = 局_值
			}
		}
		if 局_CPU峰值 >= 零_告警高CPU阈值 {
			局_结果 = append(局_结果, 响应_告警项{
				J级别: "error",
				B标题: "CPU 压力较高",
				X详情: fmt.Sprintf("当前最忙核心约 %.0f%%，建议结合“进程排行”和“CPU 抓样”定位热点。", 局_CPU峰值),
			})
		}
		if 服务.Ram.UsedPercent >= 零_告警高内存阈值 {
			局_结果 = append(局_结果, 响应_告警项{
				J级别: "warning",
				B标题: "宿主机内存偏高",
				X详情: fmt.Sprintf("当前宿主机内存占用约 %d%%，建议继续查看 heap / allocs。", 服务.Ram.UsedPercent),
			})
		}
		if 服务.Disk.UsedPercent >= 零_告警高磁盘阈值 {
			局_结果 = append(局_结果, 响应_告警项{
				J级别: "warning",
				B标题: "磁盘空间偏紧",
				X详情: fmt.Sprintf("当前磁盘占用约 %d%%，日志或临时文件可能需要清理。", 服务.Disk.UsedPercent),
			})
		}
	}

	if len(活动请求) >= 零_告警活动请求阈值 {
		局_结果 = append(局_结果, 响应_告警项{
			J级别: "warning",
			B标题: "活动请求较多",
			X详情: fmt.Sprintf("当前有 %d 个请求正在执行，建议优先关注活动请求和慢请求面板。", len(活动请求)),
		})
	}

	if len(数据库.C错误列表) > 0 && 数据库.C错误列表[0].S死锁 {
		局_结果 = append(局_结果, 响应_告警项{
			J级别: "error",
			B标题: "检测到数据库死锁/锁等待",
			X详情: fmt.Sprintf("最近一次数据库异常发生在 %s，可直接查看下方“最近 SQL 错误 / 死锁”。", 数据库.C错误列表[0].S时间),
		})
	}

	if len(数据库.M慢SQL列表) >= 3 {
		局_结果 = append(局_结果, 响应_告警项{
			J级别: "warning",
			B标题: "近期慢 SQL 较多",
			X详情: fmt.Sprintf("最近缓存中有 %d 条慢 SQL，建议先看 SQL 模板排行。", len(数据库.M慢SQL列表)),
		})
	}

	for _, 局_路由 := range 路由列表 {
		if 局_路由.C次数 >= 5 && 局_路由.P95毫秒 >= 零_告警高P95毫秒 {
			局_结果 = append(局_结果, 响应_告警项{
				J级别: "warning",
				B标题: "存在高延迟路由",
				X详情: fmt.Sprintf("%s %s 的 P95 约 %.0fms，建议结合趋势图和慢请求继续排查。", 局_路由.F方法, 局_路由.L路由, 局_路由.P95毫秒),
			})
			break
		}
	}

	if len(慢请求列表) >= 5 {
		局_结果 = append(局_结果, 响应_告警项{
			J级别: "info",
			B标题: "慢请求缓存较多",
			X详情: fmt.Sprintf("当前慢请求缓存里有 %d 条记录，可先筛查是否集中在同一路由。", len(慢请求列表)),
		})
	}

	if len(局_结果) == 0 {
		局_结果 = append(局_结果, 响应_告警项{
			J级别: "success",
			B标题: "当前未发现明显异常",
			X详情: "核心资源、数据库和路由指标暂未命中预设告警阈值。",
		})
	}

	return 局_结果
}

func 取进程CPU总秒(进程 *gprocess.Process) (float64, bool) {
	if 进程 == nil {
		return 0, false
	}

	局_CPU时间, 局_错误 := 进程.Times()
	if 局_错误 != nil || 局_CPU时间 == nil {
		return 0, false
	}
	return 局_CPU时间.User + 局_CPU时间.System, true
}

func 构造进程指标(采样 进程_初始采样, 经过秒数 float64) (响应_进程指标, bool) {
	if 采样.J进程 == nil || 经过秒数 <= 0 {
		return 响应_进程指标{}, false
	}

	局_当前CPU总秒, 局_可用 := 取进程CPU总秒(采样.J进程)
	if !局_可用 {
		return 响应_进程指标{}, false
	}

	局_CPU占比 := 保留两位小数((局_当前CPU总秒 - 采样.CCPU总秒) * 100 / 经过秒数)
	if 局_CPU占比 < 0 {
		局_CPU占比 = 0
	}

	局_名称, _ := 采样.J进程.Name()
	局_命令行, _ := 采样.J进程.Cmdline()
	局_内存信息, _ := 采样.J进程.MemoryInfo()
	局_内存占比, _ := 采样.J进程.MemoryPercent()
	局_线程数, _ := 采样.J进程.NumThreads()
	局_状态列表, _ := 采样.J进程.Status()
	局_启动毫秒, _ := 采样.J进程.CreateTime()

	局_内存MB := 0.0
	if 局_内存信息 != nil {
		局_内存MB = 字节转MB(局_内存信息.RSS)
	}

	局_启动时间 := ""
	局_运行秒数 := int64(0)
	if 局_启动毫秒 > 0 {
		局_启动时间值 := time.UnixMilli(局_启动毫秒)
		局_启动时间 = 局_启动时间值.Format(time.RFC3339)
		局_运行秒数 = int64(time.Since(局_启动时间值).Seconds())
		if 局_运行秒数 < 0 {
			局_运行秒数 = 0
		}
	}

	局_状态 := strings.Join(局_状态列表, ", ")
	if 局_状态 == "" {
		局_状态 = "-"
	}
	if 局_名称 == "" {
		局_名称 = fmt.Sprintf("PID %d", 采样.J进程.Pid)
	}

	return 响应_进程指标{
		J进程ID:  采样.J进程.Pid,
		M名称:    局_名称,
		C命令行:   截断文本(局_命令行, 180),
		CCPU占比: 局_CPU占比,
		N内存MB:  保留两位小数(局_内存MB),
		N内存占比:  保留两位小数(float64(局_内存占比)),
		X线程数:   局_线程数,
		Z状态:    局_状态,
		Q启动时间:  局_启动时间,
		Y运行秒数:  局_运行秒数,
	}, true
}

func 汇总CPU画像(画像数据 []byte) ([]响应_CPU热点, error) {
	局_画像, 局_错误 := gprofile.ParseData(画像数据)
	if 局_错误 != nil {
		return nil, 局_错误
	}

	局_采样索引 := 0
	for 局_索引, 局_采样类型 := range 局_画像.SampleType {
		if 局_采样类型 != nil && 局_采样类型.Unit == "nanoseconds" {
			局_采样索引 = 局_索引
			break
		}
	}

	局_平耗映射 := make(map[string]int64)
	局_累计映射 := make(map[string]int64)
	局_总量 := int64(0)

	for _, 局_样本 := range 局_画像.Sample {
		if 局_采样索引 >= len(局_样本.Value) {
			continue
		}
		局_值 := 局_样本.Value[局_采样索引]
		if 局_值 <= 0 {
			continue
		}
		局_总量 += 局_值

		局_已统计 := make(map[string]struct{})
		for 局_索引, 局_位置 := range 局_样本.Location {
			局_函数名 := 取位置函数名(局_位置)
			if 局_函数名 == "" {
				continue
			}
			if 局_索引 == 0 {
				局_平耗映射[局_函数名] += 局_值
			}
			if _, 局_存在 := 局_已统计[局_函数名]; !局_存在 {
				局_累计映射[局_函数名] += 局_值
				局_已统计[局_函数名] = struct{}{}
			}
		}
	}

	if 局_总量 == 0 {
		return []响应_CPU热点{}, nil
	}

	局_函数名列表 := make([]string, 0, len(局_累计映射))
	for 局_函数名 := range 局_累计映射 {
		局_函数名列表 = append(局_函数名列表, 局_函数名)
	}
	sort.Slice(局_函数名列表, func(i, k int) bool {
		if 局_平耗映射[局_函数名列表[i]] == 局_平耗映射[局_函数名列表[k]] {
			return 局_累计映射[局_函数名列表[i]] > 局_累计映射[局_函数名列表[k]]
		}
		return 局_平耗映射[局_函数名列表[i]] > 局_平耗映射[局_函数名列表[k]]
	})

	局_上限 := 15
	if len(局_函数名列表) < 局_上限 {
		局_上限 = len(局_函数名列表)
	}

	局_结果 := make([]响应_CPU热点, 0, 局_上限)
	for _, 局_函数名 := range 局_函数名列表[:局_上限] {
		局_结果 = append(局_结果, 响应_CPU热点{
			H函数名:  局_函数名,
			P平耗毫秒: 保留两位小数(float64(局_平耗映射[局_函数名]) / float64(time.Millisecond)),
			P平耗占比: 保留两位小数(float64(局_平耗映射[局_函数名]) * 100 / float64(局_总量)),
			L累计毫秒: 保留两位小数(float64(局_累计映射[局_函数名]) / float64(time.Millisecond)),
			L累计占比: 保留两位小数(float64(局_累计映射[局_函数名]) * 100 / float64(局_总量)),
		})
	}

	return 局_结果, nil
}

func 取位置函数名(位置 *gprofile.Location) string {
	if 位置 == nil {
		return ""
	}
	for _, 局_行 := range 位置.Line {
		if 局_行.Function != nil && 局_行.Function.Name != "" {
			return 局_行.Function.Name
		}
	}
	return ""
}

func 取画像数量(名称 string) int {
	局_画像 := rpprof.Lookup(名称)
	if 局_画像 == nil {
		return 0
	}
	return 局_画像.Count()
}

func 构造画像信息(名称 string, 手动开关 bool, 已开启 bool, 描述 string) 响应_画像信息 {
	return 响应_画像信息{
		M名称:   名称,
		S数量:   取画像数量(名称),
		S手动开关: 手动开关,
		Y已开启:  已开启,
		M描述:   描述,
	}
}

func 是否支持文本画像(名称 string) bool {
	switch 名称 {
	case "goroutine", "heap", "allocs", "threadcreate", "block", "mutex":
		return true
	default:
		return false
	}
}

func 是否支持下载画像(名称 string) bool {
	return 是否支持文本画像(名称)
}

func 归一化路由(路由 string, 兜底 string) string {
	if 路由 != "" {
		return 路由
	}
	if 兜底 == "" {
		return "/"
	}
	return 兜底
}

func 时长转毫秒(耗时 time.Duration) float64 {
	return float64(耗时) / float64(time.Millisecond)
}

func 字节转MB(值 uint64) float64 {
	return float64(值) / 1024 / 1024
}

func 保留两位小数(值 float64) float64 {
	return float64(int(值*100)) / 100
}

func 格式化时间(时间值 time.Time) string {
	if 时间值.IsZero() {
		return ""
	}
	return 时间值.Format(time.RFC3339)
}

func 截断文本(文本 string, 最大长度 int) string {
	if 最大长度 <= 0 || len(文本) <= 最大长度 {
		return 文本
	}
	return 文本[:最大长度] + "..."
}

func 安全方法(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return c.Request.Method
}

func 安全路径(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Path
}

func 安全查询(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.RawQuery
}

func 安全客户端IP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.ClientIP()
}

func 追加有界切片[T any](切片 []T, 项 T, 最大数量 int) []T {
	if 最大数量 <= 0 {
		return 切片
	}
	if len(切片) < 最大数量 {
		return append(切片, 项)
	}
	copy(切片, 切片[1:])
	切片[len(切片)-1] = 项
	return 切片
}
