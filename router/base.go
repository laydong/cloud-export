package router

import (
	"cloud-export/conf"
	"github.com/gin-gonic/gin"
	"github.com/laydong/toolpkg/middleware"
	"net/http"
)

func Routers() *gin.Engine {
	var Router = gin.Default()
	Router.StaticFS(conf.App.Upload.FileUrl, http.Dir("."+conf.App.Upload.FileUrl)) // 为用户头像和文件提供静态地址
	// 跨域
	Router.Use(middleware.Cors()) // 如需跨域可以打开
	Router.NoRoute(middleware.NotRouter())
	Router.NoMethod(middleware.NoMethodHandle())
	Router.Use(middleware.MiddlewareApiLog)
	ApiRouter(Router)
	return Router
}
