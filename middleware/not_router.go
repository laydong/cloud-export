package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NotRouter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			map[string]string{
				"code":       "400",
				"data":       "",
				"msg":        "当前请求不存在",
				"request_id": c.GetString("request_id"),
			})
	}
}
