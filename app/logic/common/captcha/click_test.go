package captcha

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strconv"
	"strings"
	"testing"

	"server/app/logic/common/captcha/clickItem"
)

// 解码_按前缀 根据data URL的MIME前缀选择解码器解码图片。
func 解码_按前缀(t *testing.T, dataURL string) image.Image {
	t.Helper()
	if !strings.HasPrefix(dataURL, "data:image/") {
		t.Fatalf("图片缺少data URL前缀: %s", dataURL[:32])
	}
	局_载荷, 局_错误 := base64.StdEncoding.DecodeString(dataURL[strings.Index(dataURL, ",")+1:])
	if 局_错误 != nil {
		t.Fatalf("图片不是有效base64: %v", 局_错误)
	}
	switch {
	case strings.HasPrefix(dataURL, 图_GIF前缀):
		局_图, 局_错误 := gif.Decode(bytes.NewReader(局_载荷))
		if 局_错误 != nil {
			t.Fatalf("图片不是有效GIF: %v", 局_错误)
		}
		return 局_图
	case strings.HasPrefix(dataURL, 图_JPEG前缀):
		局_图, 局_错误 := jpeg.Decode(bytes.NewReader(局_载荷))
		if 局_错误 != nil {
			t.Fatalf("图片不是有效JPEG: %v", 局_错误)
		}
		return 局_图
	default:
		局_图, 局_错误 := png.Decode(bytes.NewReader(局_载荷))
		if 局_错误 != nil {
			t.Fatalf("图片不是有效PNG: %v", 局_错误)
		}
		return 局_图
	}
}

func TestGenerateAndVerifyClick(t *testing.T) {
	局_原存储 := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	t.Cleanup(func() { VerificationCodes = 局_原存储 })

	id, dataURL, err := GenerateClick(-1)
	if err != nil {
		t.Fatalf("GenerateClick() error = %v", err)
	}
	if len(id) != 18 {
		t.Fatalf("challenge id length = %d, want 18", len(id))
	}
	局_图片 := 解码_按前缀(t, dataURL)
	if got, want := 局_图片.Bounds(), image.Rect(0, 0, clickItem.B布_画布宽, clickItem.B布_画布高); got != want {
		t.Fatalf("image bounds = %v, want %v", got, want)
	}

	raw, ok := VerificationCodes.backend().Get(cachePrefix + id)
	if !ok {
		t.Fatal("challenge was not cached")
	}
	targets, ok := raw.([]image.Rectangle)
	if !ok || len(targets) != clickItem.T题_点击数 {
		t.Fatalf("cached targets = %#v", raw)
	}
	局_答案段 := make([]string, len(targets))
	for i, target := range targets {
		局_答案段[i] = strconv.Itoa(target.Min.X+1) + "|" + strconv.Itoa(target.Min.Y+1)
	}
	局_答案 := strings.Join(局_答案段, ",")
	if !VerifyClick(id, 局_答案, true) {
		t.Fatal("correct click answer failed")
	}
	if VerifyClick(id, 局_答案, true) {
		t.Fatal("consumed click answer passed twice")
	}
}

func TestVerifyClickRejectsMalformedAnswerWithoutConsuming(t *testing.T) {
	局_原存储 := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	t.Cleanup(func() { VerificationCodes = 局_原存储 })

	VerificationCodes.set("id", []image.Rectangle{
		image.Rect(0, 0, 48, 48),
		image.Rect(48, 0, 96, 48),
		image.Rect(96, 0, 144, 48),
		image.Rect(144, 0, 192, 48),
	})
	if VerifyClick("id", "bad", true) {
		t.Fatal("malformed click answer passed")
	}
	if !VerifyClick("id", "1|1,49|1,97|1,145|1", true) {
		t.Fatal("valid answer failed after malformed attempt")
	}
}

// TestGenerateClickAllMembers 遍历全部注册成员，
// 保证每一种题型都能生成合法挑战、编码为可解码图片并通过点击校验全链路。
func TestGenerateClickAllMembers(t *testing.T) {
	局_原存储 := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	t.Cleanup(func() { VerificationCodes = 局_原存储 })

	for _, 局_项 := range 集_点击题型表 {
		局_名称 := 局_项.题型.Q题型_名称()
		局_随机, 局_错误 := clickItem.X渲染_新建随机源()
		if 局_错误 != nil {
			t.Fatalf("[%s] 新建随机源失败: %v", 局_名称, 局_错误)
		}
		局_挑战, 局_错误 := 局_项.题型.T题_生成(8, 局_随机)
		if 局_错误 != nil {
			t.Fatalf("[%s] 生成失败: %v", 局_名称, 局_错误)
		}
		if len(局_挑战.Z帧) == 0 {
			t.Fatalf("[%s] 未生成任何帧", 局_名称)
		}
		for _, 局_帧 := range 局_挑战.Z帧 {
			if got, want := 局_帧.Bounds(), image.Rect(0, 0, clickItem.B布_画布宽, clickItem.B布_画布高); got != want {
				t.Fatalf("[%s] 帧尺寸 = %v, want %v", 局_名称, got, want)
			}
		}
		if len(局_挑战.M目标) != clickItem.T题_点击数 {
			t.Fatalf("[%s] 目标数量 = %d", 局_名称, len(局_挑战.M目标))
		}

		局_前缀, 局_数据, 局_错误 := 编码_挑战图(局_挑战)
		if 局_错误 != nil {
			t.Fatalf("[%s] 编码失败: %v", 局_名称, 局_错误)
		}
		局_图片 := 解码_按前缀(t, 局_前缀+base64.StdEncoding.EncodeToString(局_数据))
		if 局_图片.Bounds().Dx() != clickItem.B布_画布宽 {
			t.Fatalf("[%s] 编码后图片宽度异常", 局_名称)
		}
		if len(局_挑战.Z帧) > 1 && 局_前缀 != 图_GIF前缀 {
			t.Fatalf("[%s] 多帧题未编码为GIF", 局_名称)
		}
	}
}

func BenchmarkGenerateClick(b *testing.B) {
	局_原存储 := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	b.Cleanup(func() { VerificationCodes = 局_原存储 })
	_, _, _ = GenerateClick(8)
	b.ResetTimer()
	for range b.N {
		if _, _, err := GenerateClick(8); err != nil {
			b.Fatal(err)
		}
	}
}
