package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/models/common"
	appUtils "server/new/app/utils"
	DB "server/structs/db"
	serverUtils "server/utils"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestC处理响应数据_Abort后输出响应(t *testing.T) {
	gin.SetMode(gin.TestMode)
	局_业务已执行 := false
	路由 := gin.New()
	路由.Use(Z测试上下文(&common.Q请求_上下文{AppInfo: DB.DB_AppInfo{CryptoType: 1}}))
	路由.Use(C处理响应数据())
	路由.Use(func(c *gin.Context) {
		response.FailMsg(c, 12345, "blocked")
		c.Abort()
	})
	路由.GET("/", func(c *gin.Context) {
		局_业务已执行 = true
	})

	记录器 := httptest.NewRecorder()
	路由.ServeHTTP(记录器, httptest.NewRequest(http.MethodGet, "/", nil))

	if 局_业务已执行 {
		t.Fatal("Abort 后仍执行了业务 handler")
	}
	var 响应 map[string]interface{}
	if err := json.Unmarshal(记录器.Body.Bytes(), &响应); err != nil {
		t.Fatalf("响应不是有效 JSON: %v, body=%q", err, 记录器.Body.String())
	}
	if 响应["Status"] != float64(12345) || 响应["Msg"] != "blocked" {
		t.Fatalf("响应内容错误: %s", 记录器.Body.String())
	}
}

func TestC处理响应数据_AES加密和明文回退(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantAES bool
	}{
		{name: "有效密钥", key: "123456789012345678901234", wantAES: true},
		{name: "无效密钥回退明文", key: "short", wantAES: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			路由 := gin.New()
			路由.Use(Z测试上下文(&common.Q请求_上下文{
				AppInfo:      DB.DB_AppInfo{CryptoType: 2},
				CryptoKeyAes: tt.key,
			}))
			路由.Use(C处理响应数据())
			路由.GET("/", func(c *gin.Context) {
				response.FailMsg(c, 12345, "result")
			})

			记录器 := httptest.NewRecorder()
			路由.ServeHTTP(记录器, httptest.NewRequest(http.MethodGet, "/", nil))
			响应明文 := 记录器.Body.String()

			if tt.wantAES {
				var 加密包 请求响应_加密包
				if err := json.Unmarshal(记录器.Body.Bytes(), &加密包); err != nil {
					t.Fatalf("加密响应解析失败: %v", err)
				}
				密文, err := base64.StdEncoding.DecodeString(加密包.A密文)
				if err != nil {
					t.Fatalf("密文 base64 解码失败: %v", err)
				}
				响应明文 = serverUtils.Aes解密_cbc192(密文, tt.key)
			}

			var 响应 map[string]interface{}
			if err := json.Unmarshal([]byte(响应明文), &响应); err != nil {
				t.Fatalf("响应明文不是有效 JSON: %v, body=%q", err, 响应明文)
			}
			if 响应["Status"] != float64(12345) || 响应["Msg"] != "result" {
				t.Fatalf("响应内容错误: %s", 响应明文)
			}
		})
	}
}

func Z测试上下文(ctx *common.Q请求_上下文) gin.HandlerFunc {
	return func(c *gin.Context) {
		appUtils.Z置上下文(c, ctx)
		c.Next()
	}
}
