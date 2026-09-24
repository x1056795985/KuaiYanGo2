package captcha

import (
	"os"
	"path/filepath"
	"testing"

	"server/app/logic/common/captcha/clickItem"
)

// TestSampleOutput 生成全部题型的样例图片到临时目录，供人工检查视觉效果。
// 运行: go test ./app/logic/common/captcha/ -run TestSampleOutput -v
func TestSampleOutput(t *testing.T) {
	局_原存储 := VerificationCodes
	VerificationCodes = NewStore(newMemoryCache())
	t.Cleanup(func() { VerificationCodes = 局_原存储 })

	for _, 局_项 := range 集_点击题型表 {
		局_随机, 局_错误 := clickItem.X渲染_新建随机源()
		if 局_错误 != nil {
			t.Fatal(局_错误)
		}
		局_挑战, 局_错误 := 局_项.题型.T题_生成(8, 局_随机)
		if 局_错误 != nil {
			t.Fatal(局_错误)
		}
		局_前缀, 局_数据, 局_错误 := 编码_挑战图(局_挑战)
		if 局_错误 != nil {
			t.Fatal(局_错误)
		}
		局_扩展名 := map[string]string{图_PNG前缀: "png", 图_JPEG前缀: "jpg", 图_GIF前缀: "gif"}[局_前缀]
		局_路径 := filepath.Join(os.TempDir(), "captcha_"+局_项.题型.Q题型_名称()+"."+局_扩展名)
		if 局_错误 = os.WriteFile(局_路径, 局_数据, 0644); 局_错误 != nil {
			t.Fatal(局_错误)
		}
		t.Logf("[%s] %s (%d KB)", 局_项.题型.Q题型_名称(), 局_路径, len(局_数据)/1024)
	}
}
