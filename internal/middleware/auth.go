package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthRequired 检查管理员是否已登录
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		authenticated := session.Get("authenticated")

		if authenticated != true {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
