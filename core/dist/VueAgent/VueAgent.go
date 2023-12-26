package VueAgent

import (
	_ "embed"
	"github.com/gin-gonic/gin"
	"net/http"
)

//go:embed index.html
var Html []byte

type HtmlHandler struct{}

func NewHtmlHandler() *HtmlHandler {
	return &HtmlHandler{}
}

func (h *HtmlHandler) RedirectIndex(c *gin.Context) {
	c.Redirect(http.StatusFound, "/index.html")
	return
}

func (h *HtmlHandler) Index(c *gin.Context) {
	c.Header("content-type", "text/html;charset=utf-8")
	c.String(200, string(Html))
	return
}
