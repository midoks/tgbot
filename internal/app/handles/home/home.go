package home

import (
	// "fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"tgbot/internal/app/common"
	// "tgbot/internal/op"
)

func Index(c *gin.Context) {
	data := common.FrontendCommonVer(c)
	c.HTML(http.StatusOK, "home/index.tmpl", data)
}

func NotFound(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "404.tmpl", data)
}
