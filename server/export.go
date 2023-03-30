package server

import (
	"cloud-export/model"
	"cloud-export/model/request"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/laydong/toolpkg/logx"
)

// 获取参数哈希
func paramHash(v interface{}) (hashKey string, err error) {
	paramBt, err := json.Marshal(v)
	if err != nil {
		return
	}
	hash := md5.Sum(paramBt)
	hashKey = fmt.Sprintf("%x", hash)
	return
}

//HandelSHttp 请求投递队列
func HandelSHttp(c *gin.Context, param *request.ExpSHttpParam) (ret interface{}, err error) {
	// 1. 参数哈希
	hashKey, err := paramHash(param)
	if err != nil {
		logx.ErrorF(c, "param hash err: %s", err.Error())
		return
	}
	ret = map[string]string{"hash_key": hashKey}

	models := new(model.ExportLogModel)
	// 2. 查询任务是否存在，不存在则记录 存在直接返回
	res, _ := models.QueryByHashKey(c, hashKey)
	if res.Id > 0 {
		logx.ErrorF(c, "任务已经存在 hash_key: %s", hashKey)
		return
	}
	expLog := &model.ExportLogModel{
		HashKey:    hashKey,
		Title:      param.Title,
		ExtType:    param.EXTType,
		SourceType: model.StypeHttp,
		Status:     model.Status_pending,
		Callback:   param.CallBack,
	}
	source, err := json.Marshal(param.SourceHTTP)
	if err != nil {
		logx.ErrorF(c, "json.Marshal err: %s", err.Error())
		return
	}
	expLog.Param = string(source)
	err = models.CreateTask(c, expLog)
	if err != nil {
		logx.ErrorF(c, "exportLog insert err: %s", err.Error())
		return
	}
	// 3. 准备参数丢任务队列中
	httpQ := &Mq{
		Key: TaskHttpKey,
	}
	httpQ.Push(c, &ExportTask{
		TaskID: hashKey,
	})
	return
}
