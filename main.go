package main

import (
	"cloud-export/conf"
	"cloud-export/global"
	"cloud-export/router"
	"cloud-export/task"
	"github.com/gin-gonic/gin"
	"github.com/laydong/toolpkg"
)

var path = "./conf/app.toml"

func main() {
	err := conf.InitApp(path)
	if err != nil {
		panic(err)
	}
	toolpkg.InitLog(toolpkg.AppConf{
		AppName: conf.App.AppConf.AppName,
		AppMode: conf.App.AppConf.AppMode,
	})
	err = global.InitApp()
	if err != nil {
		panic(err)
	}
	//开启处理协程
	go new(task.HttpWorker).Run(conf.App.TaskPool.HttpWorker)
	// 初始化操作
	gin.SetMode(conf.App.AppConf.AppMode)
	router.Routers().Run(conf.App.AppConf.Port)
}
