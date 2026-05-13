package middleware

import (
	"github.com/gin-gonic/gin"

	"tgbot/internal/conf"
)

func CheckInstalledAfter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if conf.Security.InstallLock {
			c.Redirect(302, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
