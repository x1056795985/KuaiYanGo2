package clickItem

import (
	"crypto/rand"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"math"
	mathrand "math/rand"

	drawx "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// X渲染_新建随机源 用加密随机字节作为种子创建渲染专用伪随机源：
// 种子不可预测，且生成速度远快于逐次调用加密随机数。
func X渲染_新建随机源() (*mathrand.Rand, error) {
	var 局_种子 [8]byte
	if _, 局_错误 := rand.Read(局_种子[:]); 局_错误 != nil {
		return nil, 局_错误
	}
	return mathrand.New(mathrand.NewSource(int64(binary.BigEndian.Uint64(局_种子[:])))), nil
}

// 渲染_新建画布 创建一张固定尺寸的验证码画布。
func 渲染_新建画布() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, B布_画布宽, B布_画布高))
}

// 渲染_画提示栏 在顶部提示栏绘制"请依次点击"文字与目标图标缩略图。
// 缩略图做轻量色彩抖动与噪点，阻止提示图与网格图直接逐像素比对。
func 渲染_画提示栏(画布 *image.RGBA, 目标id []int, 随机 *mathrand.Rand) {
	draw.Draw(画布, image.Rect(0, 0, 布_提示栏宽, 格_边长), 集_素材.提示图, image.Point{}, draw.Src)
	for 局_i, 局_id := range 目标id {
		局_x := 布_提示栏宽 + 局_i*格_边长
		局_缩略图 := 图标_色彩抖动(集_素材.原图[局_id], 随机, 14, 0.16, 0.12, 3)
		draw.Draw(画布, image.Rect(局_x, 0, 局_x+格_边长, 格_边长), 局_缩略图, image.Point{}, draw.Over)
	}
	draw.Draw(画布, image.Rect(0, 格_边长-1, B布_画布宽, 格_边长), image.NewUniform(color.Black), image.Point{}, draw.Src)
}

// 渲染_画网格 将布局中的图标逐格绘制到画布网格区。
// 每格铺设随机浅色底；图标执行随机渲染管线(色彩抖动+双向扭曲+旋转缩放)；
// 遮挡概率大于0时，随机用其他图标的缩放剪影盖住格子一角，
// 选择性破坏图标的判别性特征，人类可靠格式塔补全识别。
func 渲染_画网格(画布 *image.RGBA, 布局 布局_结果, 随机 *mathrand.Rand, 遮挡概率 int) {
	for 局_序, 局_id := range 布局.网格 {
		局_x := 局_序 % 格_列数 * 格_边长
		局_y := 局_序/格_列数*格_边长 + 格_边长
		局_格子 := image.Rect(局_x, 局_y, 局_x+格_边长, 局_y+格_边长)
		if 局_id == 0 {
			draw.Draw(画布, 局_格子, 集_素材.空白格, image.Point{}, draw.Over)
			continue
		}
		图标_填充格子底色(画布, 局_格子, 随机)
		draw.Draw(画布, 局_格子, 图标_随机渲染(集_素材.原图[局_id], 随机), image.Point{}, draw.Over)
		if 遮挡概率 > 0 && 随机.Intn(100) < 遮挡概率 {
			画_遮挡物(画布, 局_格子, 随机)
		}
	}
}

// 画_遮挡物 在格子的随机一角叠放一张缩小后的随机图标，
// 遮住目标图标的一部分，制造"缺失脑补"识别条件。
func 画_遮挡物(画布 *image.RGBA, 格子 image.Rectangle, 随机 *mathrand.Rand) {
	局_遮挡id := 1 + 随机.Intn(90)
	局_边长 := 格_边长*45/100 + 随机.Intn(格_边长*20/100)
	局_偏移集合 := [][2]int{
		{0, 0},
		{格_边长 - 局_边长, 0},
		{0, 格_边长 - 局_边长},
		{格_边长 - 局_边长, 格_边长 - 局_边长},
		{格_边长/2 - 局_边长/2, 格_边长/2 - 局_边长/2},
	}
	局_偏移 := 局_偏移集合[随机.Intn(len(局_偏移集合))]
	局_区域 := image.Rect(格子.Min.X+局_偏移[0], 格子.Min.Y+局_偏移[1], 0, 0)
	局_区域.Max = 局_区域.Min.Add(image.Pt(局_边长, 局_边长))
	局_遮挡图 := 图标_随机渲染(集_素材.原图[局_遮挡id], 随机)
	drawx.NearestNeighbor.Scale(画布, 局_区域, 局_遮挡图, 局_遮挡图.Bounds(), draw.Over, nil)
}

// 渲染_加装饰 在画布上叠加干扰曲线与整图噪点，
// 保证每次渲染的字节级不同，杜绝离线哈希比对。
func 渲染_加装饰(画布 *image.RGBA, 随机 *mathrand.Rand) {
	//画_干扰曲线(画布, 随机)
	画_整图噪点(画布, 随机, 2)
}

// 画_聚光罩 将除高亮格子与提示栏外的区域压暗，
// 用于动态题的"依次闪现"效果，顺序信息只存在于帧序列中。
func 画_聚光罩(帧 *image.RGBA, 高亮 image.Rectangle) {
	局_罩 := image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 148})
	for 局_序 := 0; 局_序 < 格_列数*格_行数; 局_序++ {
		局_x := 局_序 % 格_列数 * 格_边长
		局_y := 局_序/格_列数*格_边长 + 格_边长
		局_格子 := image.Rect(局_x, 局_y, 局_x+格_边长, 局_y+格_边长)
		if 局_格子 == 高亮 {
			continue
		}
		draw.Draw(帧, 局_格子, 局_罩, image.Point{}, draw.Over)
	}
}

// 图标_随机渲染 对原图执行整套运行时随机变换：
// 色彩抖动 -> 双向正弦扭曲 -> 随机角度与缩放旋转。
// 全部参数在每次请求时随机抽取，攻击者无法建立有限的模板库。
func 图标_随机渲染(原图 image.Image, 随机 *mathrand.Rand) image.Image {
	局_抖动图 := 图标_色彩抖动(原图, 随机, 30, 0.3, 0.22, 5)
	局_扭曲图 := 图标_随机扭曲(局_抖动图, 随机)
	return 图标_旋转缩放(局_扭曲图, (随机.Float64()*2-1)*42, 0.8+随机.Float64()*0.35)
}

// 图标_色彩抖动 在 HSV 空间对整图做随机色相/饱和度/明度偏移，
// 并对每个不透明像素叠加随机噪点。人眼识别图标主要依赖形状，
// 颜色扰动几乎不影响人工识别，但可瓦解基于颜色特征与像素哈希的匹配。
func 图标_色彩抖动(原图 image.Image, 随机 *mathrand.Rand, 色相幅度, 饱和幅度, 明度幅度 float64, 噪点 int) *image.RGBA {
	局_色相偏移 := (随机.Float64()*2 - 1) * 色相幅度 * math.Pi / 180
	局_饱和倍率 := 1 + (随机.Float64()*2-1)*饱和幅度
	局_明度倍率 := 1 + (随机.Float64()*2-1)*明度幅度
	局_噪点模 := uint32(噪点*2 + 1)
	局_边界 := 原图.Bounds()
	局_输出 := image.NewRGBA(image.Rect(0, 0, 局_边界.Dx(), 局_边界.Dy()))
	for 局_y := 0; 局_y < 局_边界.Dy(); 局_y++ {
		for 局_x := 0; 局_x < 局_边界.Dx(); 局_x++ {
			局_r, 局_g, 局_b, 局_a := 原图.At(局_边界.Min.X+局_x, 局_边界.Min.Y+局_y).RGBA()
			if 局_a == 0 {
				continue
			}
			局_透明度 := int(局_a >> 8)
			局_色相, 局_饱和, 局_明度 := 色_转HSV(uint8(局_r*255/局_a>>8), uint8(局_g*255/局_a>>8), uint8(局_b*255/局_a>>8))
			局_色相 = math.Mod(局_色相+局_色相偏移+2*math.Pi, 2*math.Pi)
			局_饱和 = 色_限01(局_饱和 * 局_饱和倍率)
			局_明度 = 色_限01(局_明度 * 局_明度倍率)
			局_红, 局_绿, 局_蓝 := 色_转RGB(局_色相, 局_饱和, 局_明度)
			局_随机数 := 随机.Uint32()
			局_红 = uint8(色_限8位(int(局_红) + int(局_随机数%局_噪点模) - 噪点))
			局_绿 = uint8(色_限8位(int(局_绿) + int(局_随机数>>8%局_噪点模) - 噪点))
			局_蓝 = uint8(色_限8位(int(局_蓝) + int(局_随机数>>16%局_噪点模) - 噪点))
			局_输出.SetRGBA(局_x, 局_y, color.RGBA{
				uint8(int(局_红) * 局_透明度 / 255),
				uint8(int(局_绿) * 局_透明度 / 255),
				uint8(int(局_蓝) * 局_透明度 / 255),
				uint8(局_透明度),
			})
		}
	}
	return 局_输出
}

// 图标_随机扭曲 用随机幅度、频率、相位的双向正弦扰动重新采样图像，
// 产生非仿射的局部形变，破坏特征点匹配所依赖的几何一致性。
func 图标_随机扭曲(原图 image.Image, 随机 *mathrand.Rand) *image.RGBA {
	局_x幅度 := 4 + 随机.Float64()*7
	局_x频率 := 0.04 + 随机.Float64()*0.06
	局_x相位 := 随机.Float64() * 2 * math.Pi
	局_y幅度 := 1.5 + 随机.Float64()*3.5
	局_y频率 := 0.05 + 随机.Float64()*0.07
	局_y相位 := 随机.Float64() * 2 * math.Pi
	局_边界 := 原图.Bounds()
	局_宽, 局_高 := 局_边界.Dx(), 局_边界.Dy()
	局_输出 := image.NewRGBA(image.Rect(0, 0, 局_宽, 局_高))
	for 局_y := 0; 局_y < 局_高; 局_y++ {
		for 局_x := 0; 局_x < 局_宽; 局_x++ {
			局_源x := 局_x + int(局_x幅度*math.Sin(float64(局_y)*局_x频率+局_x相位))
			局_源y := 局_y + int(局_y幅度*math.Sin(float64(局_x)*局_y频率+局_y相位))
			if 局_源x >= 0 && 局_源x < 局_宽 && 局_源y >= 0 && 局_源y < 局_高 {
				局_输出.Set(局_x, 局_y, 原图.At(局_边界.Min.X+局_源x, 局_边界.Min.Y+局_源y))
			}
		}
	}
	return 局_输出
}

// 图标_旋转缩放 按随机角度与缩放比做双线性插值变换，
// 连续参数令模板穷举失效。
func 图标_旋转缩放(原图 image.Image, 角度, 缩放 float64) image.Image {
	局_弧度 := 角度 * math.Pi / 180
	局_中心 := float64(格_边长) / 2
	局_余弦 := math.Cos(局_弧度) * 缩放
	局_正弦 := math.Sin(局_弧度) * 缩放
	局_矩阵 := f64.Aff3{
		局_余弦, -局_正弦, 局_中心 - 局_余弦*局_中心 + 局_正弦*局_中心,
		局_正弦, 局_余弦, 局_中心 - 局_正弦*局_中心 - 局_余弦*局_中心,
	}
	局_输出 := image.NewRGBA(image.Rect(0, 0, 格_边长, 格_边长))
	drawx.BiLinear.Transform(局_输出, 局_矩阵, 原图, 原图.Bounds(), draw.Over, nil)
	return 局_输出
}

// 图标_填充格子底色 为格子铺设随机的浅色底，消除纯白背景带来的零成本分割。
func 图标_填充格子底色(画布 *image.RGBA, 格子 image.Rectangle, 随机 *mathrand.Rand) {
	局_基准 := 226 + 随机.Intn(30)
	draw.Draw(画布, 格子, image.NewUniform(color.RGBA{
		uint8(色_限8位(局_基准 - 7 + 随机.Intn(15))),
		uint8(色_限8位(局_基准 - 7 + 随机.Intn(15))),
		uint8(色_限8位(局_基准 - 7 + 随机.Intn(15))),
		255,
	}), image.Point{}, draw.Src)
}

// 画_干扰曲线 在网格区域叠放多条半透明随机贝塞尔曲线，
// 干扰轮廓提取与直线扫描，人眼几乎不受影响。
func 画_干扰曲线(画布 *image.RGBA, 随机 *mathrand.Rand) {
	局_边界 := 画布.Bounds()
	局_顶部 := 格_边长 + 6
	局_高度 := 局_边界.Dy() - 局_顶部
	if 局_高度 <= 4 {
		return
	}
	for 局_序号, 局_总数 := 0, 3+随机.Intn(3); 局_序号 < 局_总数; 局_序号++ {
		局_起点 := image.Pt(随机.Intn(局_边界.Dx()), 局_顶部+随机.Intn(局_高度))
		局_终点 := image.Pt(随机.Intn(局_边界.Dx()), 局_顶部+随机.Intn(局_高度))
		局_控制点 := image.Pt(随机.Intn(局_边界.Dx()), 局_顶部+随机.Intn(局_高度))
		局_画笔 := image.NewUniform(color.RGBA{
			uint8(40 + 随机.Intn(170)),
			uint8(40 + 随机.Intn(170)),
			uint8(40 + 随机.Intn(170)),
			uint8(30 + 随机.Intn(60)),
		})
		局_步数 := 局_边界.Dx() / 2
		for 局_步 := 0; 局_步 <= 局_步数; 局_步++ {
			局_比例 := float64(局_步) / float64(局_步数)
			局_余 := 1 - 局_比例
			局_x := 局_余*局_余*float64(局_起点.X) + 2*局_余*局_比例*float64(局_控制点.X) + 局_比例*局_比例*float64(局_终点.X)
			局_y := 局_余*局_余*float64(局_起点.Y) + 2*局_余*局_比例*float64(局_控制点.Y) + 局_比例*局_比例*float64(局_终点.Y)
			draw.Draw(画布, image.Rect(int(局_x), int(局_y), int(局_x)+2, int(局_y)+2), 局_画笔, image.Point{}, draw.Over)
		}
	}
}

// 画_整图噪点 对整幅图像的不透明像素叠加轻微随机噪点，
// 保证同一图标每次渲染的输出字节级不同，杜绝离线哈希比对。
func 画_整图噪点(画布 *image.RGBA, 随机 *mathrand.Rand, 幅度 int) {
	局_模 := uint32(幅度*2 + 1)
	for 局_下标 := 0; 局_下标+3 < len(画布.Pix); 局_下标 += 4 {
		if 画布.Pix[局_下标+3] == 0 {
			continue
		}
		局_随机数 := 随机.Uint32()
		画布.Pix[局_下标] = byte(色_限8位(int(画布.Pix[局_下标]) + int(局_随机数%局_模) - 幅度))
		画布.Pix[局_下标+1] = byte(色_限8位(int(画布.Pix[局_下标+1]) + int(局_随机数>>8%局_模) - 幅度))
		画布.Pix[局_下标+2] = byte(色_限8位(int(画布.Pix[局_下标+2]) + int(局_随机数>>16%局_模) - 幅度))
	}
}

// 色_转HSV 将 8 位 RGB 转为弧度制 HSV（色相范围 [0,2π)）。
func 色_转HSV(红, 绿, 蓝 uint8) (float64, float64, float64) {
	局_红, 局_绿, 局_蓝 := float64(红)/255, float64(绿)/255, float64(蓝)/255
	局_最大 := math.Max(局_红, math.Max(局_绿, 局_蓝))
	局_最小 := math.Min(局_红, math.Min(局_绿, 局_蓝))
	局_差 := 局_最大 - 局_最小
	if 局_最大 == 0 {
		return 0, 0, 0
	}
	局_饱和 := 局_差 / 局_最大
	if 局_差 == 0 {
		return 0, 局_饱和, 局_最大
	}
	var 局_色相 float64
	switch 局_最大 {
	case 局_红:
		局_色相 = math.Mod((局_绿-局_蓝)/局_差, 6)
	case 局_绿:
		局_色相 = (局_蓝-局_红)/局_差 + 2
	default:
		局_色相 = (局_红-局_绿)/局_差 + 4
	}
	局_色相 *= math.Pi / 3
	if 局_色相 < 0 {
		局_色相 += 2 * math.Pi
	}
	return 局_色相, 局_饱和, 局_最大
}

// 色_转RGB 将弧度制 HSV 转回 8 位 RGB。
func 色_转RGB(色相, 饱和, 明度 float64) (uint8, uint8, uint8) {
	局_色度 := 明度 * 饱和
	局_段 := math.Mod(色相, 2*math.Pi) / (math.Pi / 3)
	局_次色 := 局_色度 * (1 - math.Abs(math.Mod(局_段, 2)-1))
	var 局_红, 局_绿, 局_蓝 float64
	switch int(局_段) {
	case 0:
		局_红, 局_绿, 局_蓝 = 局_色度, 局_次色, 0
	case 1:
		局_红, 局_绿, 局_蓝 = 局_次色, 局_色度, 0
	case 2:
		局_红, 局_绿, 局_蓝 = 0, 局_色度, 局_次色
	case 3:
		局_红, 局_绿, 局_蓝 = 0, 局_次色, 局_色度
	case 4:
		局_红, 局_绿, 局_蓝 = 局_次色, 0, 局_色度
	default:
		局_红, 局_绿, 局_蓝 = 局_色度, 0, 局_次色
	}
	局_补偿 := 明度 - 局_色度
	return uint8(math.Round((局_红 + 局_补偿) * 255)),
		uint8(math.Round((局_绿 + 局_补偿) * 255)),
		uint8(math.Round((局_蓝 + 局_补偿) * 255))
}

// 色_限01 将浮点色值钳制到 [0,1]。
func 色_限01(值 float64) float64 {
	if 值 < 0 {
		return 0
	}
	if 值 > 1 {
		return 1
	}
	return 值
}

// 色_限8位 将整数色值钳制到 [0,255]。
func 色_限8位(值 int) int {
	if 值 < 0 {
		return 0
	}
	if 值 > 255 {
		return 255
	}
	return 值
}
