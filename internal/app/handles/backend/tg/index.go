package tg

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tgbot/internal/app/common"
)

func Home(c *gin.Context) {
	data := common.CommonVer(c)

	c.HTML(http.StatusOK, "backend/tg/index.tmpl", data)
}

func Add(c *gin.Context) {
	data := common.CommonVer(c)

	c.HTML(http.StatusOK, "backend/tg/add.tmpl", data)
}
