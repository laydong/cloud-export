package router

import (
	"cloud-export/handler"
	"github.com/gin-gonic/gin"
)

func ApiRouter(r *gin.Engine) {
	v := r.Group("/api/")
	{
		v.POST("http", handler.ExportSHttp)
		v.POST("raw", handler.ExportSRaw)
		v.POST("detail", handler.ExportDetail)
		//v.POST("test", handler.Test)
	}
}
