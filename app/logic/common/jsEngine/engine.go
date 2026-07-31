package jsEngine

import (
	. "EFunc/utils"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"

	"server/app/global"
	"server/app/logic/common/cycleNot"
	"server/app/logic/common/rmbPay"
	dbm "server/app/models/db"
	"server/app/utils/Qqwry"
)

type 脚本引擎_运行时绑定 struct {
	名称 string
	值  any
}

var (
	集_模块导入正则  = regexp.MustCompile(`\n@?import\s+['"](.*?)['"]`)
	集_模块下载客户端 = req.C().EnableInsecureSkipVerify().SetTimeout(15 * time.Second)
	集_静态绑定    []脚本引擎_运行时绑定
)

func init() {
	集_静态绑定 = []脚本引擎_运行时绑定{
		{"$程序_延时", 脚本引擎_延时},
		{"$api_用户Id取详情", 脚本引擎_用户Id取详情},
		{"$api_卡号Id取详情", 脚本引擎_卡号Id取详情},
		{"$api_取软件用户详情", 脚本引擎_取软件用户详情},
		{"$api_在线注销", 脚本引擎_在线注销},
		{"$api_用户Id增减余额", 脚本引擎_用户Id增减余额},
		{"$api_用户Id增减积分", 脚本引擎_用户Id增减积分},
		{"$api_用户Id增减时间点数", 脚本引擎_用户Id增减时间点数},
		{"$api_读公共变量", 脚本引擎_读公共变量},
		{"$api_置公共变量", 脚本引擎_置公共变量},
		{"$api_网页访问_GET", 脚本引擎_网页访问Get},
		{"$api_网页访问_POST", 脚本引擎_网页访问Post},
		{"$api_置动态标记", 脚本引擎_置动态标记},
		{"$api_执行SQL查询", 脚本引擎_执行SQL查询},
		{"$api_执行SQL功能", 脚本引擎_执行SQL功能},
		{"$api_任务池_任务创建", 脚本引擎_任务池任务创建},
		{"$api_任务池_任务查询", 脚本引擎_任务池任务查询},
		{"$api_短信发送", 脚本引擎_短信发送},
		{"$api_用户名或卡号取uid", 脚本引擎_用户名或卡号取Uid},
		{"$api_取用户云配置", 脚本引擎_取用户云配置},
		{"$api_置用户云配置", 脚本引擎_置用户云配置},
		{"$api_取缓存", 脚本引擎_取缓存},
		{"$api_置缓存", 脚本引擎_置缓存},
		{"$api_置黑名单", 脚本引擎_置黑名单},
		{"$api_置软件用户状态", 脚本引擎_置软件用户状态},
		{"$api_任务池Uuid添加到队列", 脚本引擎_任务池Uuid添加到队列},
		{"$api_任务池_取队列长度", 脚本引擎_任务池取队列长度},
		{"$api_Jwt生成", 脚本引擎_Jwt生成},
		{"$api_云存储_取外链", 脚本引擎_云存储取外链},
		{"$api_云存储_取文件上传授权", 脚本引擎_云存储取文件上传授权},
		{"$api_云存储_文件信息", 脚本引擎_云存储_文件信息},
		{"$api_ws_发送消息", 脚本引擎_WebSocket发送消息},
		{"$api_ws_发送消息_批量", 脚本引擎_WebSocket批量发送消息},
		{"$api_ws_筛选id", 脚本引擎_WebSocket筛选Id},
		{"$api_编码_BASE64编码", B编码_BASE64编码},
		{"$api_编码_BASE64解码", B编码_BASE64解码},
		{"$api_字节集_十六进制到字节集", Z字节集_十六进制到字节集},
		{"$api_字节集_字节集到十六进制", Z字节集_字节集到十六进制},
		{"$api_文本_取文本右边", W文本_取文本右边},
		{"$api_文本_取文本左边", W文本_取文本左边},
		{"$api_文本_取出中间文本", W文本_取出中间文本},
		{"$api_文本_子文本替换", W文本_子文本替换},
		{"$api_时间_取现行时间戳", S时间_取现行时间戳},
		{"$api_时间_取现行时间戳13", S时间_取现行时间戳13},
		{"$api_生成二维码并转base64", rmbPay.L_rmbPay.S生成二维码并转base64},
		{"$api_VMP计算授权码", 脚本引擎_VMP计算授权码},
		{"$api_ip查地区", Qqwry.Ip查信息2},
		{"$api_定制批量注册", 脚本引擎_定制批量注册},
		{"$api_定制批量充值", 脚本引擎_定制批量充值},
		{"$api_定制批量取账号信息", 脚本引擎_定制批量取账号信息},
	}
	cycleNot.J脚本引擎_设置初始化函数(J脚本引擎_初始化用户)
}

// J脚本引擎_初始化用户 创建隔离的脚本运行时并注册宿主 API。
func J脚本引擎_初始化用户(c *gin.Context, appInfo *dbm.DB_AppInfo, online *dbm.DB_LinksToken, publicJS *dbm.DB_PublicJs) *goja.Runtime {
	if c == nil {
		c = 脚本引擎_后台上下文()
	}
	if appInfo == nil {
		appInfo = &dbm.DB_AppInfo{}
	}
	if online == nil {
		online = &dbm.DB_LinksToken{}
	}
	if publicJS == nil {
		publicJS = &dbm.DB_PublicJs{}
	}

	局_运行时 := goja.New()
	_ = 局_运行时.Set("$用户在线信息", online)
	_ = 局_运行时.Set("$应用信息", map[string]any{
		"AppId": appInfo.AppId, "AppName": appInfo.AppName,
		"Status": appInfo.Status, "VipData": appInfo.VipData,
	})
	局_控制台 := 局_运行时.NewObject()
	_ = 局_控制台.Set("log", 脚本引擎_控制台日志)
	_ = 局_运行时.Set("console", 局_控制台)
	for _, 局_绑定 := range 集_静态绑定 {
		_ = 局_运行时.Set(局_绑定.名称, 局_绑定.值)
	}
	_ = 局_运行时.Set("$Request", 脚本引擎_请求快照(c))

	publicJS.Value = 脚本引擎_展开模块导入(publicJS.Value)
	return 局_运行时
}

func 脚本引擎_请求快照(c *gin.Context) map[string]any {
	局_空快照 := map[string]any{"Url": map[string]any{}, "Header": []string{}, "Form": map[string]any{}, "Host": "", "Body": "", "Method": ""}
	if c == nil || c.Request == nil {
		return 局_空快照
	}
	局_请求头 := make([]string, 0, len(c.Request.Header))
	for 局_名称, 局_值数组 := range c.Request.Header {
		局_请求头 = append(局_请求头, 局_名称+": "+strings.Join(局_值数组, ", "))
	}
	局_表单 := make(map[string]any, len(c.Request.Form))
	for 局_名称, 局_值数组 := range c.Request.Form {
		if len(局_值数组) > 0 {
			局_表单[局_名称] = 局_值数组[0]
		}
	}
	return map[string]any{
		"Url": c.Request.URL, "Header": 局_请求头, "Form": 局_表单,
		"Host": c.Request.Host, "Body": c.Request.Body, "Method": c.Request.Method,
	}
}

func 脚本引擎_展开模块导入(source string) string {
	if !strings.Contains(source, "import '") && !strings.Contains(source, `import "`) {
		return source
	}
	局_匹配数组 := 集_模块导入正则.FindAllStringSubmatch("\n"+source, -1)
	for _, 局_匹配 := range 局_匹配数组 {
		if len(局_匹配) < 2 {
			continue
		}
		局_模块名 := 局_匹配[1]
		局_本地路径, 局_是否远程, 局_有效 := 脚本引擎_解析模块路径(局_模块名)
		if !局_有效 {
			continue
		}
		if 局_是否远程 {
			局_强制刷新 := strings.HasPrefix(strings.TrimSpace(局_匹配[0]), "@")
			if 局_强制刷新 || !W文件_是否存在(局_本地路径) {
				if 局_响应, 局_错误 := 集_模块下载客户端.R().Get(局_模块名); 局_错误 == nil && 局_响应.StatusCode >= 200 && 局_响应.StatusCode < 300 {
					_ = M目录_创建(W文件_取父目录(局_本地路径))
					_ = W文件_写到文件(局_本地路径, 局_响应.Bytes())
				}
			}
		}
		if W文件_是否存在(局_本地路径) {
			source = strings.Replace(source, strings.TrimSpace(局_匹配[0]), W文件_读入文本(局_本地路径), 1)
		}
	}
	return source
}

func 脚本引擎_解析模块路径(moduleName string) (string, bool, bool) {
	局_基础路径 := filepath.Clean(filepath.Join(global.GVA_CONFIG.Q取运行目录, "云函数"))
	if strings.HasPrefix(moduleName, "http://") || strings.HasPrefix(moduleName, "https://") {
		局_摘要 := sha256.Sum256([]byte(moduleName))
		return filepath.Join(局_基础路径, "lib", hex.EncodeToString(局_摘要[:])+".js"), true, true
	}
	局_候选路径 := filepath.Clean(filepath.Join(局_基础路径, filepath.FromSlash(moduleName)))
	局_相对路径, 局_错误 := filepath.Rel(局_基础路径, 局_候选路径)
	if 局_错误 != nil || 局_相对路径 == ".." || strings.HasPrefix(局_相对路径, ".."+string(filepath.Separator)) {
		return "", false, false
	}
	return 局_候选路径, false, true
}

func 脚本引擎_控制台日志(call goja.FunctionCall) goja.Value {
	局_值 := call.Argument(0)
	global.GVA_LOG.Println(局_值.String())
	return 局_值
}
