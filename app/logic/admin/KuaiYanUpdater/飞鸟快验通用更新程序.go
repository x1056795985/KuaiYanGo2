package KuaiYanUpdater

import (
	"EFunc/utils"
	"crypto/md5"
	json2 "encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/imroc/req/v3"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/valyala/fastjson"
	"gorm.io/gorm"

	"server/app/global"
)

var J_系统更新状态 = 0 //0未更新 1 更新中 2更新失败 3下载成功
var J_系统更新提示 = ""
var J_任务列表 []更新文件列表
var 集_运行目录 = ""

func K快验系统开始更新(更新文件文本 string, 更新成功后处理程序 func(执行程序本地路径 string)) {
	if J_系统更新状态 != 0 {
		fmt.Println("已经在更新程序了,不要再调用了")
		return
	}
	J_系统更新状态 = 1
	集_运行目录 = utils.C程序_取运行目录()
	/*	if runtime.GOOS == "windows" {
		集_运行目录 = "."
	}*/
	局_json, err := fastjson.Parse(更新文件文本)
	if err != nil {
		J_系统更新提示 = "更新失败,请重试"
		J_系统更新状态 = 2
		return
	}
	var 执行程序路径 = ""
	局_文件 := 局_json.GetArray("data")
	J_任务列表 = make([]更新文件列表, len(局_文件))
	for 索引 := range 局_文件 {
		/*		{
				"WenJianMin":"文件名.exe",
				"md5":"e10adc3949ba59abbe56e057f20f883e(小写文件md5可选,有就校验,空就只校验文件名)",
				"Lujing":"\/(下载本地相对路径)",
				"size":"12345",
				"url":"https:\/\/www.baidu.com\/文件名.exe(下载路径)",
				"YunXing":"1(值为更新完成后会运行这个文件,只能有一个文件值为1)"
			}*/
		局_临时文件名 := string(局_文件[索引].GetStringBytes("WenJianMin"))
		if 局_临时文件名 == "" {
			局_临时文件名 = string(局_文件[索引].GetStringBytes("url"))
			局_临时文件名 = filepath.Base(局_临时文件名) //取出路径文件名
		}

		var 局_本地路径 string
		局_本地路径 = 集_运行目录
		if string(局_文件[索引].GetStringBytes("Lujing")) == "" {
			局_本地路径 += "/"
		} else {
			局_本地路径 += string(局_文件[索引].GetStringBytes("Lujing"))
		}
		局_本地路径 += 局_临时文件名
		if utils.W文件_是否存在(集_运行目录 + string(局_文件[索引].GetStringBytes("Lujing"))) {
			_ = utils.M目录_创建(集_运行目录 + string(局_文件[索引].GetStringBytes("Lujing")))
		}
		J_任务列表[索引] = 更新文件列表{
			本地文件名:  局_本地路径,
			远程下载地址: string(局_文件[索引].GetStringBytes("url")),
			更新结束后是否需要自动执行该文件: string(局_文件[索引].GetStringBytes("YunXing")) == "1",
			是否已下载: false,
		}
		if J_任务列表[索引].更新结束后是否需要自动执行该文件 {
			执行程序路径 = J_任务列表[索引].本地文件名
		}
		J_系统更新提示 = "正在读取并校验更新文件，请稍候....."

		局_临时文件MD5 := strings.ToUpper(string(局_文件[索引].GetStringBytes("md5")))
		if 局_临时文件MD5 != "" { // 有md5 就校验,没有就文件名校验
			局_本地文件MD5 := ""
			data := utils.W文件_读入文件(局_本地路径) //切片
			if data != nil {
				has := md5.Sum(data)
				局_本地文件MD5 = strings.ToUpper(fmt.Sprintf("%x", has)) //将[]byte转成16进制

				if 局_临时文件MD5 == 局_本地文件MD5 {
					J_任务列表[索引].是否已下载 = true //文件已经存在直接跳过
					continue                //到循环尾
				}
			}
		} else if utils.W文件_是否存在(局_本地路径) { //不推荐文件名,可能会出现不准确的情况
			J_任务列表[索引].是否已下载 = true //文件已经存在直接跳过
		}

	}

	if len(J_任务列表) == 0 {
		goto 标签_更新成功

	}
	//开始下载列表
	for 索引 := range J_任务列表 {
		if J_任务列表[索引].是否已下载 { //不需要下载
			continue //到循环尾
		}

		callback := func(info req.DownloadInfo) {
			if info.Response.Response != nil {
				J_系统更新提示 = fmt.Sprintf("下载更新中:%v/%v,已下载: %.2f%%\n", 索引+1, len(J_任务列表), float64(info.DownloadedSize)/float64(info.Response.ContentLength)*100.0)
				fmt.Printf(J_系统更新提示)
			}
		}

		client := req.C().EnableInsecureSkipVerify() //.DevMode()

		transport := client.GetTransport()
		transport.WrapRoundTripFunc(func(rt http.RoundTripper) req.HttpRoundTripFunc {
			return func(req *http.Request) (resp *http.Response, err error) {
				// before request
				// ...
				req.Header.Add("x-cos-traffic-limit", "10485760") //限速1280kb/s

				//	权限只读id和key
				secretID := utils.D到文本(utils.B编码_BASE64解码("QUtJREdOR3RIVFI5Y3BuV3pEQ3ZQZGNMcDRhcnRnRGFrZUpp"))
				secretKey := utils.D到文本(utils.B编码_BASE64解码("Q0F0TmJhSm4xMGpEU1NDZ3Z1ZThOTThldnhqR1haTHM="))
				startTime := time.Unix(time.Now().Unix()-3600, 0)
				endTime := time.Unix(time.Now().Unix()+36000, 0) //有效期 1小时
				authTime := &cos.AuthTime{
					SignStartTime: startTime,
					SignEndTime:   endTime,
					KeyStartTime:  startTime,
					KeyEndTime:    endTime,
				}

				cos.AddAuthorizationHeader(secretID, secretKey, "", req, authTime)
				resp, err = rt.RoundTrip(req)
				// after response
				// ...
				return
			}
		})

		_, err1 := client.R().
			SetOutputFile(J_任务列表[索引].本地文件名).
			SetDownloadCallback(callback).
			Get(J_任务列表[索引].远程下载地址)

		if err1 != nil {
			J_系统更新提示 = "文件:" + J_任务列表[索引].远程下载地址 + ",下载失败," + err1.Error()
			J_系统更新状态 = 2
			fmt.Println(J_系统更新提示)
			J_任务列表 = make([]更新文件列表, 0)
			return
		}

	}
标签_更新成功:
	J_系统更新提示 = "下载成功"

	//J_任务列表 = make([]更新文件列表, 0) // 全部先下载成功 更新成功的不要清除,不然无法判断是否更新了
	J_系统更新状态 = 3
	if 更新成功后处理程序 != nil {
		J_系统更新提示 = "等待重启"
		更新成功后处理程序(执行程序路径)
	}

}

type 更新文件列表 struct {
	本地文件名            string
	远程下载地址           string
	是否已下载            bool
	更新结束后是否需要自动执行该文件 bool //注意，该数据类型中此成员为真只允许有一个。建议设为主程序。
}

// 解决启动后宝塔显示未运行的情况 不完美,但是不需要gcc
func B宝塔_修改项目信息pid() {
	var files []string
	root := "/var/tmp/gopids" // 请将此处替换为你的目录路径
	f, err := os.Open(root)
	if err != nil {
		//fmt.Println("打开目录失败:", err)
		return
	}
	defer f.Close()

	files2, err := f.Readdir(-1)
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return
	}
	for _, file := range files2 {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".pid") {
			fmt.Println(filepath.Join(root, file.Name()))
			files = append(files, filepath.Join(root, file.Name()))
		}
	}

	if len(files) == 1 {
		//  /var/tmp/gopids  修改这个可以修改宝塔检测飞鸟的pid值,用来显示是否运行中
		pid := os.Getpid()
		err = os.WriteFile(files[0], []byte(strconv.Itoa(pid)), 0644)
		if err != nil {
			global.GVA_LOG.Println(fmt.Sprintf("写出pid失败:%v", err.Error()))
			return
		}
		global.GVA_LOG.Println(fmt.Sprintf("写出pid成功:%v", pid))
		return
	}
	global.GVA_LOG.Println(fmt.Sprintf("扫描pid文件信息:%v", files))
	return
}

type 宝塔_项目信息 struct {
	Id            int    `json:"id" gorm:"column:id"`
	Name          string `json:"name" gorm:"column:name;comment:名称"`
	Path          string `json:"path" gorm:"column:path;comment:路径"`
	ProjectConfig string `json:"projectConfig" gorm:"column:project_config;comment:状态"`
}

type 宝塔_项目配置 struct {
	SslPath      string   `json:"ssl_path"`
	ProjectName  string   `json:"project_name"`
	ProjectExe   string   `json:"project_exe"`
	BindExtranet int      `json:"bind_extranet"`
	Domains      []string `json:"domains"`
	ProjectCmd   string   `json:"project_cmd"`
	IsPowerOn    int      `json:"is_power_on"`
	RunUser      string   `json:"run_user"`
	Port         int      `json:"port"`
	ProjectPath  string   `json:"project_path"`
	LogPath      string   `json:"log_path"`
}

// B宝塔_修改项目信息 使用纯 Go SQLite 驱动更新宝塔 Go 项目信息。
func B宝塔_修改项目信息() {
	局_数据库, 局_错误 := gorm.Open(sqlite.Open("/www/server/panel/data/default.db"), &gorm.Config{})
	if 局_错误 != nil {
		global.GVA_LOG.Println(局_错误.Error())
		return
	}
	局_Pid := os.Getpid()
	局_执行文件路径, 局_错误 := filepath.Abs(os.Args[0])
	if 局_错误 != nil {
		global.GVA_LOG.Println(fmt.Sprintf("获取执行文件路径失败:%v", 局_错误.Error()))
		return
	}
	global.GVA_LOG.Println("执行文件路径:" + 局_执行文件路径)
	global.GVA_LOG.Println(fmt.Sprintf("pid:%v", 局_Pid))

	局_项目信息, 局_错误 := 宝塔_读取Go项目信息(局_数据库)
	if 局_错误 != nil {
		global.GVA_LOG.Println(局_错误.Error())
		return
	}
	fmt.Printf("项目信息:%v", 局_项目信息)
	if 局_执行文件路径 == 局_项目信息.Path { //无变化,不改动
		return
	}

	//  /var/tmp/gopids  修改这个可以修改宝塔检测飞鸟的pid值,用来显示是否运行中
	局_错误 = os.WriteFile("/var/tmp/gopids/"+局_项目信息.Name+".pid", []byte(strconv.Itoa(局_Pid)), 0644)
	if 局_错误 != nil {
		global.GVA_LOG.Println(fmt.Sprintf("写出pid失败:%v", 局_错误.Error()))
		return
	}

	局_错误 = 宝塔_更新项目数据库(局_数据库, 局_项目信息, 局_执行文件路径)
	if 局_错误 != nil {
		global.GVA_LOG.Println(局_错误.Error())
		return
	}
	// 删除脚本就可以,启动时会自动再创建
	局_错误 = utils.W文件_删除("/www/server/go_project/vhost/scripts/" + 局_项目信息.Name + ".sh")
	if 局_错误 != nil {
		global.GVA_LOG.Println("脚本删除失败:" + 局_错误.Error())
		return
	}
	global.GVA_LOG.Println("处理成功")
}

func 宝塔_读取Go项目信息(database *gorm.DB) (宝塔_项目信息, error) {
	var 局_项目信息 宝塔_项目信息
	局_SQL := `SELECT id, name, path, project_config FROM sites WHERE project_type = 'Go'`
	局_错误 := database.Raw(局_SQL).First(&局_项目信息).Error
	return 局_项目信息, 局_错误
}

func 宝塔_更新项目数据库(database *gorm.DB, project 宝塔_项目信息, executionPath string) error {
	var 局_项目配置 宝塔_项目配置
	if 局_错误 := json2.Unmarshal([]byte(project.ProjectConfig), &局_项目配置); 局_错误 != nil {
		return fmt.Errorf("解析项目配置失败: %w", 局_错误)
	}
	局_项目配置.ProjectExe = executionPath
	局_项目配置.ProjectCmd = executionPath

	局_配置JSON, 局_错误 := json2.Marshal(局_项目配置)
	if 局_错误 != nil {
		return fmt.Errorf("编码项目配置失败: %w", 局_错误)
	}
	局_SQL := `UPDATE sites SET path = ?, project_config = ? WHERE id = ?`
	return database.Exec(局_SQL, executionPath, 局_配置JSON, project.Id).Error
}
