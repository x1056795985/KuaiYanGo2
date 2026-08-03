package appInfo

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"regexp"
	"server/app/logic/common/cloudStorage"
	dbm "server/app/models/db"
	"strings"
)

// 处理应用下载更新地址中的模板变量
func App下载更新地址变量处理(DB_AppInfo dbm.DB_AppInfo) string {
	局_新文本 := DB_AppInfo.UrlDownload

	局_新文本 = strings.Replace(局_新文本, "{{AppName}}", DB_AppInfo.AppName, -1)

	if strings.Index(局_新文本, "{{AppVer}}") != -1 && DB_AppInfo.AppVer != "" {
		局_可用版本 := W文本_分割文本(DB_AppInfo.AppVer, "\n")
		if len(局_可用版本) > 0 {
			局_新文本 = strings.Replace(局_新文本, "{{AppVer}}", 局_可用版本[0], -1)
		}
	}

	//{{(.*?)\((.*?)\)}}  正则匹配指令,  子匹配1为指令名 子匹配2为参数
	if strings.Index(局_新文本, "{{") != -1 { //判断是否还有变量
		re := regexp.MustCompile(`{{(.*?)\((.*?)\)}}`)
		result := re.FindAllStringSubmatch(局_新文本, -1)
		for i, _ := range result {
			局_完整文本 := result[i][0]
			局_指令名 := result[i][1]
			局_参数 := W文本_分割文本(result[i][2], ",")
			switch 局_指令名 {
			case "云存储_取外链":
				if len(局_参数) == 2 {
					下载地址, err := cloudStorage.L_云存储.Q取外链地址(&gin.Context{}, strings.Trim(局_参数[0], "'"), gconv.Int64(局_参数[1]))
					if err == nil {
						局_新文本 = strings.Replace(局_新文本, 局_完整文本, 下载地址, -1)
					}
				}
			case "云存储_取ETag":
				if len(局_参数) == 1 {
					ETag, err := cloudStorage.L_云存储.Q取ETag(&gin.Context{}, strings.Trim(局_参数[0], "'"))
					if err == nil {
						局_新文本 = strings.Replace(局_新文本, 局_完整文本, ETag, -1)
					}
				}
			}
		}
	}

	return 局_新文本
}
