package captcha

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"sync"

	"server/app/logic/common/captcha/clickItem"
)

const (
	// 图_PNG前缀 静态PNG图片的data URL前缀。
	图_PNG前缀 = "data:image/png;base64,"
	// 图_JPEG前缀 静态JPEG图片的data URL前缀。
	图_JPEG前缀 = "data:image/jpeg;base64,"
	// 图_GIF前缀 动态GIF图片的data URL前缀。
	图_GIF前缀 = "data:image/gif;base64,"
	// 图_默认帧延迟 挑战未声明帧延迟时的兜底值(1/100秒)。
	图_默认帧延迟 = 10
)

var (
	// 集_缓冲池 图片编码缓冲区的复用池。
	集_缓冲池 = sync.Pool{New: func() any { return new(bytes.Buffer) }}
	// 集_PNG编码器 快速压缩并复用编码缓冲的PNG编码器。
	集_PNG编码器 = png.Encoder{
		CompressionLevel: png.BestSpeed,
		BufferPool:       &集_PNG编码缓冲池,
	}
	// 集_PNG编码缓冲池 集_PNG编码器 的配套缓冲池。
	集_PNG编码缓冲池 PNG编码缓冲池
	// 集_量化色板 GIF量化用的256色调色板(216网页安全色+40级灰阶)，
	// 对平涂色图标还原度良好，无需引入外部调色板依赖。
	集_量化色板 = 色板_生成()
)

// PNG编码缓冲池 实现 png.EncoderBufferPool 接口，
// Get/Put 方法名由接口约束，不可改名。
type PNG编码缓冲池 struct {
	池 sync.Pool
}

func (j *PNG编码缓冲池) Get() *png.EncoderBuffer {
	if 局_值 := j.池.Get(); 局_值 != nil {
		return 局_值.(*png.EncoderBuffer)
	}
	return new(png.EncoderBuffer)
}

func (j *PNG编码缓冲池) Put(buffer *png.EncoderBuffer) {
	j.池.Put(buffer)
}

// 色板_生成 构造216色网页安全色立方体并补充40级灰阶。
func 色板_生成() color.Palette {
	var 局_色板 color.Palette
	局_梯度 := [6]uint8{0, 51, 102, 153, 204, 255}
	for 局_r := range 局_梯度 {
		for 局_g := range 局_梯度 {
			for 局_b := range 局_梯度 {
				局_色板 = append(局_色板, color.RGBA{R: 局_梯度[局_r], G: 局_梯度[局_g], B: 局_梯度[局_b], A: 255})
			}
		}
	}
	for 局_i := 0; 局_i < 40; 局_i++ {
		局_灰 := uint8(局_i * 255 / 39)
		局_色板 = append(局_色板, color.RGBA{R: 局_灰, G: 局_灰, B: 局_灰, A: 255})
	}
	return 局_色板
}

// 编码_挑战图 将挑战编码为图片字节并返回data URL前缀：
// 单帧按挑战声明的格式编码为PNG或JPEG，多帧量化后编码为循环GIF。
// 返回的字节切片为缓冲区副本，不受缓冲池回收影响。
func 编码_挑战图(挑战 *clickItem.T题_挑战) (string, []byte, error) {
	局_缓冲 := 集_缓冲池.Get().(*bytes.Buffer)
	局_缓冲.Reset()
	defer 集_缓冲池.Put(局_缓冲)

	var (
		局_前缀 string
		局_错误 error
	)
	if len(挑战.Z帧) == 1 {
		局_前缀, 局_错误 = 编码_单帧(局_缓冲, 挑战)
	} else {
		局_前缀, 局_错误 = 编码_动图(局_缓冲, 挑战)
	}
	return 局_前缀, append([]byte(nil), 局_缓冲.Bytes()...), 局_错误
}

// 编码_单帧 将静态挑战帧编码为JPEG或PNG。
func 编码_单帧(缓冲 *bytes.Buffer, 挑战 *clickItem.T题_挑战) (string, error) {
	if 挑战.G格式 == "jpeg" {
		return 图_JPEG前缀, jpeg.Encode(缓冲, 挑战.Z帧[0], &jpeg.Options{Quality: 88})
	}
	return 图_PNG前缀, 集_PNG编码器.Encode(缓冲, 挑战.Z帧[0])
}

// 编码_动图 将多帧挑战量化编码为无限循环GIF。
func 编码_动图(缓冲 *bytes.Buffer, 挑战 *clickItem.T题_挑战) (string, error) {
	局_动画 := &gif.GIF{LoopCount: 0}
	for 局_i, 局_帧 := range 挑战.Z帧 {
		局_调色帧 := image.NewPaletted(局_帧.Bounds(), 集_量化色板)
		draw.Draw(局_调色帧, 局_帧.Bounds(), 局_帧, image.Point{}, draw.Src)
		局_延迟 := 图_默认帧延迟
		if 局_i < len(挑战.Y延迟) {
			局_延迟 = 挑战.Y延迟[局_i]
		}
		局_动画.Image = append(局_动画.Image, 局_调色帧)
		局_动画.Delay = append(局_动画.Delay, 局_延迟)
	}
	return 图_GIF前缀, gif.EncodeAll(缓冲, 局_动画)
}
