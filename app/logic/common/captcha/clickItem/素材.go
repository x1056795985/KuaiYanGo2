package clickItem

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/png"
	"sync"
)

//go:embed icon/*.png
var 集_图标文件 embed.FS

var (
	集_素材一次 sync.Once
	// 集_素材 懒加载后的图标素材集。
	集_素材 结构_点击素材
	// 集_素材错误 素材加载错误，进程内只记录首次。
	集_素材错误 error
)

// 结构_点击素材 验证码图标素材集：提示图、空白格与91张图标原图。
type 结构_点击素材 struct {
	提示图 image.Image
	空白格 image.Image
	原图   [91]image.Image
}

// 素材_取 懒加载并返回图标素材，进程内只加载一次。
func 素材_取() (结构_点击素材, error) {
	集_素材一次.Do(素材_加载)
	return 集_素材, 集_素材错误
}

// 素材_加载 解码全部内嵌图标资源。
func 素材_加载() {
	集_素材.提示图, 集_素材错误 = 素材_解码图("icon/请依次点击.png")
	if 集_素材错误 != nil {
		return
	}
	集_素材.空白格, 集_素材错误 = 素材_解码图("icon/0.png")
	if 集_素材错误 != nil {
		return
	}
	for 局_id := 1; 局_id <= 90; 局_id++ {
		集_素材.原图[局_id], 集_素材错误 = 素材_解码图(fmt.Sprintf("icon/%d.png", 局_id))
		if 集_素材错误 != nil {
			return
		}
	}
}

// 素材_解码图 从embed文件系统解码单张PNG。
func 素材_解码图(名称 string) (image.Image, error) {
	局_数据, 局_错误 := 集_图标文件.ReadFile(名称)
	if 局_错误 != nil {
		return nil, 局_错误
	}
	局_图, 局_错误 := png.Decode(bytes.NewReader(局_数据))
	if 局_错误 != nil {
		return nil, fmt.Errorf("解码验证码素材 %s: %w", 名称, 局_错误)
	}
	return 局_图, nil
}
