package Captcha

import (
	. "EFunc/utils"
	"bytes"
	"embed"
	"encoding/base64"
	draw2 "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"server/global"
	"strconv"
	"strings"
	"time"
)

// 校验验证码  就是多一步 校验,实际没区别
func Captcha_Verify点选(id, Value string, 是否删除 bool) bool {
	// 记录选中的4个图片的位置
	var 局_数组_被选择坐标 []image.Rectangle
	// 从缓存中获取被选中的图片坐标
	局_缓存值, ok := global.H缓存.Get(CAPTCHA + id)
	if !ok { // 如果缓存中不存在，直接返回假
		return false
	}
	// 类型断言，将 interface{} 转换为 []image.Rectangle
	if 局_数组_被选择坐标, ok = 局_缓存值.([]image.Rectangle); !ok { // 如果类型断言失败，返回假
		return false
	}
	// 解析用户点击的坐标
	点击坐标 := strings.Split(Value, ",")
	if len(点击坐标) != 4 {
		return false
	}

	// 验证每个点击的坐标是否在对应的图片区域内
	for i, 坐标 := range 点击坐标 {
		xy := strings.Split(坐标, "|")
		if len(xy) != 2 {
			return false
		}

		x, err := strconv.Atoi(xy[0])
		if err != nil {
			return false
		}

		y, err := strconv.Atoi(xy[1])
		if err != nil {
			return false
		}

		// 检查点击的坐标是否在对应的图片区域内
		if !image.Pt(x, y).In(局_数组_被选择坐标[i]) {
			return false
		}
	}

	// 如果验证通过且需要删除缓存，则删除
	if 是否删除 {
		global.H缓存.Delete(CAPTCHA + id)
	}

	return true
}

//go:embed icon/*
var icon embed.FS //每个png图片都是48*48的图片 编号 1.png ~ 96.png  和 请依次点击.png

// Captcha_取点选验证码 生成点选验证码图片

func Captcha_取点选验证码(难度 int) (验证码id, Base64验证码图片 string, err error) {

	局_id := W文本_取随机字符串(18)
	局_横 := 7
	局_纵 := 3
	if 难度 >= 局_横*局_纵-4 { //最大不能超过数量,还需要减掉右下角位置,用于显示刷新按钮
		难度 = 局_横*局_纵 - 4 - 1
	}

	局_数组_所有图片名 := W文本_取随机数字数组(1, 90, 4+难度)
	局_数组_被选择 := S数组_取随机成员(局_数组_所有图片名, 4)

	// 创建一背景图片
	background := image.NewRGBA(image.Rect(0, 0, 局_横*48, 局_纵*48+48)) //纵+48是因为要留出顶部一行
	//左上角读取 请依次点击.png 并画到背景图片左上角
	画图片(background, "请依次点击", image.Rect(0, 0, 150, 48))
	//将 局_数组_被选择 四个图片 画在后面
	for i, 局_图片 := range 局_数组_被选择 {
		画图片(background, 局_图片, image.Rect(i*48+150, 0, i*48+150+48, 48))
	}
	//在y48位置画一条像素为1 横线
	draw.Draw(background, image.Rect(0, 47, background.Bounds().Dx(), 48), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{0, 0}, draw.Over)
	// 记录选中的4个图片的位置
	局_数组_被选择坐标 := make([]image.Rectangle, 4)
	//打乱 局_数组_所有图片名 的顺序
	rand.Shuffle(len(局_数组_所有图片名), func(i, j int) {
		局_数组_所有图片名[i], 局_数组_所有图片名[j] = 局_数组_所有图片名[j], 局_数组_所有图片名[i]
	})

	//补充空白图片到图片尾部
	for i := range 局_横 * 局_纵 {
		if len(局_数组_所有图片名) < 局_横*局_纵 {
			//单数加到数组头,双数加到尾部
			if i%2 == 0 {
				局_数组_所有图片名 = append(局_数组_所有图片名, "0")
			} else {
				局_数组_所有图片名 = append([]string{"0"}, 局_数组_所有图片名...)
			}
		} else {
			break
		}
	}

	// 将24个图片按照8x3矩阵排列
	var imgData []byte
	for i := 0; i < 局_横*局_纵; i++ {
		imgData, err = icon.ReadFile("icon/" + 局_数组_所有图片名[i] + ".png")
		img, err2 := png.Decode(bytes.NewReader(imgData))
		if err2 != nil {
			//fmt.Println("读取文件时出错:"+局_数组_所有图片名[i], err2)
			return "", "", err2
		}

		// 对图片进行 S 型扭曲
		img = T图片_扭曲(img, 10.0, 0.05) // 调整 amplitude 和 frequency 参数
		// 对图片进行随机角度旋转随机旋转角度（-90~90度）
		局_随机角度 := Int64到Float64(int64(rand.Intn(90) - 45))
		img = T图片_旋转(img, 局_随机角度)

		x := (i % 局_横) * 48
		y := (i/局_横)*48 + 48
		rect := image.Rect(x, y, x+48, y+48)
		draw.Draw(background, rect, img, image.Point{0, 0}, draw.Over)

		// 记录选中的4个图片的位置
		for j, selected := range 局_数组_被选择 {
			if selected == 局_数组_所有图片名[i] {
				局_数组_被选择坐标[j] = rect
			}
		}
	}

	// 将img png图片字节数组转换为base64编码
	var buf bytes.Buffer
	err = png.Encode(&buf, background)
	if err != nil {
		return "", "", err
	}

	//将 positions 转换成文本
	global.H缓存.Set(CAPTCHA+局_id, 局_数组_被选择坐标, time.Minute*5) //[(0,48)-(48,96) (48,48)-(96,96) (96,48)-(144,96) (144,48)-(192,96)]
	//fmt.Print(局_数组_被选择坐标)
	base64Img := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	return 局_id, base64Img, nil
}

func 画图片(background *image.RGBA, 图片名称 string, 坐标 image.Rectangle) {
	imgData, _ := icon.ReadFile("icon/" + 图片名称 + ".png")
	img, _ := png.Decode(bytes.NewReader(imgData))
	draw.Draw(background, 坐标, img, image.Point{0, 0}, draw.Over)
}

// T图片_旋转 将图片以图片中心旋转指定角度（以度为单位）
func T图片_旋转(src image.Image, angle float64) image.Image {
	// 将角度转换为弧度
	radians := angle * (math.Pi / 180.0)

	// 获取图片的宽高
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算旋转后的新图片的宽高
	newWidth := int(math.Ceil(math.Abs(float64(width)*math.Cos(radians)) + math.Abs(float64(height)*math.Sin(radians))))
	newHeight := int(math.Ceil(math.Abs(float64(height)*math.Cos(radians)) + math.Abs(float64(width)*math.Sin(radians))))

	// 创建一个新的图片
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 计算旋转中心点
	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0

	// 创建仿射变换矩阵
	affine := f64.Aff3{}
	affine[0] = math.Cos(radians)
	affine[1] = -math.Sin(radians)
	affine[2] = centerX - (affine[0]*centerX + affine[1]*centerY)
	affine[3] = math.Sin(radians)
	affine[4] = math.Cos(radians)
	affine[5] = centerY - (affine[3]*centerX + affine[4]*centerY)

	// 使用仿射变换进行旋转
	draw2.BiLinear.Transform(dst, affine, src, src.Bounds(), draw.Over, nil)

	return dst
}

// T图片_扭曲 对图片进行 S 型扭曲
func T图片_扭曲(src image.Image, amplitude, frequency float64) image.Image {
	// 获取原图的尺寸
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建一个新的 RGBA 图片
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 遍历目标图片的每个像素
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算 S 型扭曲的偏移量
			offset := int(amplitude * math.Sin(float64(y)*frequency))

			// 计算原图的像素坐标
			srcX := x + offset
			srcY := y

			// 确保原图坐标在范围内
			if srcX >= 0 && srcX < width && srcY >= 0 && srcY < height {
				dst.Set(x, y, src.At(srcX, srcY))
			} else {
				// 如果超出范围，设置为透明
				dst.Set(x, y, color.Transparent)
			}
		}
	}

	return dst
}
