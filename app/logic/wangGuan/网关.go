package wangGuan

// 网关转发逻辑层:鉴权缓存、JWT签发缓存、网关配置缓存、透明反向代理(HTTP+WS)
// 原则:哑管道。不读业务body、不加解密、不记日志、不限流
// 设计文档:server2/网关功能设计方案.md

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgrijalva/jwt-go" //nolint
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/models/constant"
	"server/app/models/dbm"
)

// JWT 有效期秒数
const 常量_JWT有效期 = 60

// Token 信息缓存秒数
const 常量_Token缓存秒 = 30

// 无效 Token 负缓存秒数(防暴力枚举打库)
const 常量_负缓存秒 = 10

// 集_网关缓存 网关配置+专属传输层缓存,key=网关Id
var 集_网关缓存 atomic.Value // map[int]结构_网关缓存项

// 集_缓存加载锁 防止并发重复加载网关表
var 集_缓存加载锁 sync.Mutex

// 集_Token缓存 有效Token信息缓存,key=Token原文
var 集_Token缓存 sync.Map

// 集_Token负缓存 无效Token缓存,key=Token原文
var 集_Token负缓存 sync.Map

// 集_JWT缓存 已签发JWT复用缓存,key=网关Id|tokenId|ip
var 集_JWT缓存 sync.Map

type 结构_网关缓存项 struct {
	配置  dbm.DB_Gateway
	传输层 *http.Transport
}

type 结构_Token缓存项 struct {
	信息    dbm.DB_LinksToken
	软件AppId int            //真实软件应用AppId(软件令牌=LoginAppid;Web用户中心令牌=AppIdEx,拼分表/查应用/签JWT统一用它)
	用户信息  dbm.DB_AppUser //软件用户信息(签JWT用),游客/不存在时为零值
	应用类型  int            //1/2账密型 3/4卡号型(签JWT用)
	账号或卡号 string         //账密型=账号,卡号型=卡号(签JWT用)
	余额    float64        //账密型才有值(签JWT用)
	到期    int64
}

type 结构_JWT缓存项 struct {
	值  string
	到期 int64
	密钥 string //签发时用的Secret,与当前配置不符则重签(密钥轮换自动失效)
}

type 网关 struct{}

// L_网关 逻辑层单例
var L_网关 = 网关{}

// G刷新网关缓存 全量加载网关表到内存,增删改后由控制器调用(写时失效)
func G刷新网关缓存() error {
	db := *global.GVA_DB
	var 局_列表 []dbm.DB_Gateway
	if err := db.Model(dbm.DB_Gateway{}).Find(&局_列表).Error; err != nil {
		return err
	}
	局_新缓存 := make(map[int]结构_网关缓存项, len(局_列表))
	for _, v := range 局_列表 {
		局_新缓存[v.Id] = 结构_网关缓存项{
			配置:  v,
			传输层: Q新建传输层(v),
		}
	}
	集_网关缓存.Store(局_新缓存)
	return nil
}

// Q新建传输层 每网关独立传输层(响应头超时按网关配置),连接池参数是反代性能关键
func Q新建传输层(网关配置 dbm.DB_Gateway) *http.Transport {
	局_超时秒 := 网关配置.TimeoutSec
	if 局_超时秒 <= 0 {
		局_超时秒 = 60
	}
	return &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   100, //默认值仅2,高并发会疯狂重建TCP
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: time.Second,
		ResponseHeaderTimeout: time.Duration(局_超时秒) * time.Second, //只限制到响应头,流式body不受限
	}
}

// Q取网关 取网关配置(带缓存,未加载时懒加载)
func (j *网关) Q取网关(id int) (dbm.DB_Gateway, bool) {
	if 局_项, ok := j.Q取缓存项(id); ok {
		return 局_项.配置, true
	}
	return dbm.DB_Gateway{}, false
}

// Q取传输层 取网关专属传输层
func (j *网关) Q取传输层(id int) *http.Transport {
	if 局_项, ok := j.Q取缓存项(id); ok {
		return 局_项.传输层
	}
	return Q新建传输层(dbm.DB_Gateway{})
}

// Q取缓存项 取缓存条目,缓存未加载时加载一次
func (j *网关) Q取缓存项(id int) (结构_网关缓存项, bool) {
	if 局_值 := 集_网关缓存.Load(); 局_值 != nil {
		局_项, ok := 局_值.(map[int]结构_网关缓存项)[id]
		return 局_项, ok
	}
	集_缓存加载锁.Lock()
	if 集_网关缓存.Load() == nil {
		_ = G刷新网关缓存()
	}
	集_缓存加载锁.Unlock()
	if 局_值 := 集_网关缓存.Load(); 局_值 != nil {
		局_项, ok := 局_值.(map[int]结构_网关缓存项)[id]
		return 局_项, ok
	}
	return 结构_网关缓存项{}, false
}

// Q鉴权Token 校验客户端Token,返回在线信息与软件用户信息。无效/注销/非软件应用Token返回错误
// LoginAppid>=10000 才放行(软件应用);Web用户中心(10)同样放行(2026-10-06:手机端经H5用户中心登录,其令牌即此类型,Uid与软件令牌同源);
// 管理员(1)/代理(2)/WebApi(3)/WS(11)一律拒绝
func (j *网关) Q鉴权Token(token string) (结构_Token缓存项, error) {
	var 局_空 结构_Token缓存项
	if len(token) < 8 {
		return 局_空, errors.New("token无效")
	}
	//有效缓存
	if 局_值, ok := 集_Token缓存.Load(token); ok {
		局_项 := 局_值.(结构_Token缓存项)
		if time.Now().Unix() < 局_项.到期 {
			return 局_项, nil
		}
		集_Token缓存.Delete(token)
	}
	//负缓存
	if 局_值, ok := 集_Token负缓存.Load(token); ok {
		if time.Now().Unix() < 局_值.(int64) {
			return 局_空, errors.New("token无效")
		}
		集_Token负缓存.Delete(token)
	}

	db := *global.GVA_DB
	var 局_在线信息 dbm.DB_LinksToken
	err := db.Model(dbm.DB_LinksToken{}).Where("Token = ?", token).First(&局_在线信息).Error
	if err != nil || 局_在线信息.Status != 1 {
		集_Token负缓存.Store(token, time.Now().Unix()+常量_负缓存秒)
		return 局_空, errors.New("token无效或已注销")
	}
	if 局_在线信息.LoginAppid < 10000 && 局_在线信息.LoginAppid != constant.APPID_Web用户中心 {
		集_Token负缓存.Store(token, time.Now().Unix()+常量_负缓存秒)
		return 局_空, errors.New("该令牌不允许使用网关")
	}

	//软件身份AppId:软件令牌=LoginAppid;Web用户中心令牌(10)是虚拟应用无软件用户表,
	//真实应用AppId在AppIdEx(登录时写入),必须用它拼分表/查应用,否则会查询不存在的db_AppUser_10
	局_软件AppId := 局_在线信息.LoginAppid
	if 局_软件AppId == constant.APPID_Web用户中心 {
		局_软件AppId = 局_在线信息.AppIdEx
	}

	//软件用户信息按应用分表 db_AppUser_{AppId},游客/未注册时为零值不影响签发
	var 局_用户信息 dbm.DB_AppUser
	_ = db.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_软件AppId)).
		Where("Uid = ?", 局_在线信息.Uid).First(&局_用户信息).Error

	//应用类型与账号/卡号/余额(账密型才有余额)
	局_缓存项 := 结构_Token缓存项{信息: 局_在线信息, 软件AppId: 局_软件AppId, 用户信息: 局_用户信息, 到期: time.Now().Unix() + 常量_Token缓存秒}
	var 局_应用信息 dbm.DB_AppInfo
	if 局_错误 := db.Model(dbm.DB_AppInfo{}).Where("AppId = ?", 局_软件AppId).First(&局_应用信息).Error; 局_错误 == nil {
		局_缓存项.应用类型 = 局_应用信息.AppType
		局_缓存项.账号或卡号 = 局_在线信息.User
		if 局_应用信息.AppType == 1 || 局_应用信息.AppType == 2 { //账密型:余额在 db_User.Rmb
			var 局_用户 dbm.DB_User
			if 局_错误 := db.Model(dbm.DB_User{}).Where("Id = ?", 局_在线信息.Uid).First(&局_用户).Error; 局_错误 == nil {
				局_缓存项.余额 = 局_用户.Rmb
			}
		}
	}

	集_Token缓存.Store(token, 局_缓存项)
	return 局_缓存项, nil
}

// Q签发JWT 用网关独立密钥签发HS256轻量JWT,同一token+ip在有效期内复用同一串
// claims含软件用户信息: user账号或卡号/key绑定信息/vipTime到期或剩余点数/vipNumber积分备用/userClassId用户分类
// 账密型应用(AppType 1/2)额外携带 rmb 余额
func (j *网关) Q签发JWT(网关配置 dbm.DB_Gateway, 鉴权信息 结构_Token缓存项, ip string) (string, error) {
	在线信息 := 鉴权信息.信息
	用户信息 := 鉴权信息.用户信息
	局_缓存键 := strconv.Itoa(网关配置.Id) + "|" + strconv.Itoa(在线信息.Id) + "|" + ip
	局_现在 := time.Now().Unix()
	if 局_值, ok := 集_JWT缓存.Load(局_缓存键); ok {
		局_项 := 局_值.(结构_JWT缓存项)
		//密钥已轮换则立即重签
		if 局_项.密钥 == 网关配置.Secret && 局_现在 < 局_项.到期 {
			return 局_项.值, nil
		}
		集_JWT缓存.Delete(局_缓存键)
	}

	局_声明 := jwt.MapClaims{
		"uid":         在线信息.Uid,
		"appId":       鉴权信息.软件AppId, //真实软件应用AppId(Web用户中心令牌≠LoginAppid)
		"tokenId":     在线信息.Id,
		"user":        鉴权信息.账号或卡号,     //账密型=账号,卡号型=卡号
		"key":         用户信息.Key,       //绑定信息
		"vipTime":     用户信息.VipTime,   //到期时间或剩余点数
		"vipNumber":   用户信息.VipNumber, //积分单独备用
		"userClassId": 用户信息.UserClassId,
		"ip":          ip,
		"iat":         局_现在,
		"exp":         局_现在 + 常量_JWT有效期,
	}
	if 鉴权信息.应用类型 == 1 || 鉴权信息.应用类型 == 2 { //账密型才携带余额
		局_声明["rmb"] = 鉴权信息.余额
	}
	局_令牌, err := jwt.NewWithClaims(jwt.SigningMethodHS256, 局_声明).SignedString([]byte(网关配置.Secret))
	if err != nil {
		return "", errors.New("JWT签发失败:" + err.Error())
	}
	集_JWT缓存.Store(局_缓存键, 结构_JWT缓存项{值: 局_令牌, 到期: 局_现在 + 常量_JWT有效期 - 5, 密钥: 网关配置.Secret})
	return 局_令牌, nil
}

// Q取真实ip 取客户端真实IP。优先X-Real-Ip(nginx的$remote_addr,客户端伪造不了),
// 其次X-Forwarded-For最右一跳(由前置代理追加),最后直连地址
func (j *网关) Q取真实ip(c *gin.Context) string {
	if 局_v := strings.TrimSpace(c.GetHeader("X-Real-Ip")); 局_v != "" {
		return 局_v
	}
	if 局_v := strings.TrimSpace(c.GetHeader("X-Forwarded-For")); 局_v != "" {
		局_数组 := strings.Split(局_v, ",")
		return strings.TrimSpace(局_数组[len(局_数组)-1])
	}
	局_host, _, 局_err := net.SplitHostPort(c.Request.RemoteAddr)
	if 局_err != nil {
		return c.Request.RemoteAddr
	}
	return 局_host
}

// Q执行转发 透明反向代理(HTTP与WebSocket同一入口,ReverseProxy原生支持WS升级透传)
// 转发热路径:0查库+0日志+0body解析,仅做头替换
func (j *网关) Q执行转发(c *gin.Context, 网关配置 dbm.DB_Gateway, jwt文本, ip string) {
	局_目标, 局_err := url.Parse(网关配置.Url)
	if 局_err != nil || 局_目标.Host == "" || (局_目标.Scheme != "http" && 局_目标.Scheme != "https") {
		Q响应网关错误(c.Writer, http.StatusBadGateway, "网关Url配置无效")
		return
	}
	局_内网前缀 := 局_目标.Scheme + "://" + 局_目标.Host

	局_代理 := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			//URL改写:/wangGuan前缀 → 业务根
			局_转发路径 := strings.TrimPrefix(pr.In.URL.Path, "/wangGuan")
			if 局_转发路径 == "" {
				局_转发路径 = "/"
			}
			pr.Out.URL.Scheme = 局_目标.Scheme
			pr.Out.URL.Host = 局_目标.Host
			pr.Out.URL.Path = 局_转发路径
			pr.Out.URL.RawPath = ""
			pr.Out.Host = 局_目标.Host //Host头必须改为业务地址

			//头处理:剔除客户端凭证与可伪造头,注入网关身份头(无条件替换,防伪造)
			局_头 := pr.Out.Header
			局_头.Del("Token")
			局_头.Del("WangGuanId")
			局_头.Del("X-Forwarded-For")
			局_头.Del("X-Real-Ip")
			局_头.Del("X-Forwarded-Host")
			局_头.Del("X-Forwarded-Proto")
			局_头.Set("x-wg-jwt", jwt文本)
			局_头.Set("x-wg-ip", ip)
			//hop-by-hop头由ReverseProxy自动剥离(Upgrade/Connection对WS请求自动保留)
		},
		Transport:     j.Q取传输层(网关配置.Id),
		FlushInterval: -1, //立即flush,WS/SSE/流式必需
		ModifyResponse: func(resp *http.Response) error {
			//Location改写:业务返回3xx时把内网地址替换回/wangGuan,防泄露
			for _, 局_头名 := range []string{"Location", "Content-Location"} {
				if 局_值 := resp.Header.Get(局_头名); 局_值 != "" {
					局_新值 := strings.ReplaceAll(局_值, 局_内网前缀, "/wangGuan")
					if 局_新值 != 局_值 {
						resp.Header.Set(局_头名, 局_新值)
					}
				}
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			//上游不可达/超时:空body+状态码+x-wg-error头
			if errors.Is(err, context.DeadlineExceeded) || is超时错误(err) {
				Q响应网关错误(w, http.StatusGatewayTimeout, "业务服务响应超时")
				return
			}
			Q响应网关错误(w, http.StatusBadGateway, "业务服务连接失败")
		},
	}

	局_代理.ServeHTTP(c.Writer, c.Request)
}

// Q响应网关错误 空body+状态码+x-wg-error头
// header值只允许可见ASCII,中文做百分号编码(客户端用decodeURIComponent还原)
func Q响应网关错误(w http.ResponseWriter, 状态码 int, 信息 string) {
	w.Header().Set("x-wg-error", url.PathEscape(信息))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(状态码)
}

// is超时错误 判断是否超时类错误
func is超时错误(err error) bool {
	if 局_超时, ok := err.(net.Error); ok {
		return 局_超时.Timeout()
	}
	return strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "Timeout exceeded")
}

// Q生成密钥 32字节随机数转64位hex,每个网关独立,严禁共用
func (j *网关) Q生成密钥() (string, error) {
	局_字节 := make([]byte, 32)
	if _, err := rand.Read(局_字节); err != nil {
		return "", err
	}
	return hex.EncodeToString(局_字节), nil
}
