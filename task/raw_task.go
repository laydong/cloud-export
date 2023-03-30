package task

import (
	"cloud-export/conf"
	"cloud-export/model"
	"cloud-export/model/excel"
	"cloud-export/model/helper"
	"cloud-export/server"
	"fmt"
	"github.com/laydong/toolpkg"
	"github.com/laydong/toolpkg/logx"
	"github.com/laydong/toolpkg/utils"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type RawWorker struct {
	Tasks  *server.Mq
	taskCh chan *server.ExportTask
	req    *server.SourceHTTP
}

func (w *RawWorker) Run(pool int) {
	// 单消费端 多任务执行
	w.Tasks = &server.Mq{Key: server.TaskRawKey}
	// 缓冲区越大，程序宕机后丢消息越多
	w.taskCh = make(chan *server.ExportTask, 20)
	log.Print("RawWorker pool=", pool)
	// 启动工作协程
	w.startWorker(pool)

	// 监听队列
	go func() {
		w.Tasks.BPop(func(s string) {
			atask := &server.ExportTask{}
			atask.Build(s)
			// 丢入缓冲区
			w.taskCh <- atask
		})
	}()
}

// startWorker 启动工作协程
func (w *RawWorker) startWorker(pool int) {
	for i := 0; i < pool; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Print("[runtime] err ,recoverd", "", fmt.Errorf("%v", err).Error())
				}
			}()
			for {
				w.work()
			}
		}()
	}
}

func (w *RawWorker) work() {
	currTask := <-w.taskCh
	taskID := currTask.TaskID
	ctx := toolpkg.GetNewGinContext()
	// 1. 数据库中查询任务详情
	models := new(model.ExportLogModel)
	expLog, err := models.QueryByHashKey(ctx, taskID)
	if err != nil {
		logx.ErrorF(ctx, "TaskNotFund hash_key: %s", taskID)
		return
	}
	// 任务取消
	if expLog.Status == model.Status_cancle {
		return
	}
	// 2. 拿到json数据 -> 3. 写入excel
	paramDir := conf.App.Upload.RowFile
	excelTmpPath := conf.App.Upload.FileUrl
	err = helper.TouchDir(excelTmpPath)
	if err != nil {
		logx.ErrorF(ctx, "文件夹创建失败："+err.Error())
		return
	}
	filename := expLog.Title + "." + expLog.ExtType
	paramFilePath := path.Join(paramDir, taskID+".json")
	lists, err := ioutil.ReadFile(paramFilePath)
	if err != nil {
		expLog.FailReason = "请求参数json文件读取失败：" + err.Error()
		expLog.Status = model.Status_fail
		expLog.UpdatedAt = utils.TimeFrom(time.Now())
		err = models.UpDataTask(ctx, &expLog)
		if err != nil {
			logx.ErrorF(ctx, "更新数据失败："+err.Error())
		}
		return
	}
	excelw := excel.NewExcelRecorder(path.Join(excelTmpPath, taskID, filename))
	excelw.JsonListWrite(excel.Pos{X: 1, Y: 1}, string(lists), true)
	excelw.Save()

	// 4. 压缩文件夹 并删除源文件
	zipFilePath := path.Join(excelTmpPath, taskID+".zip")
	taskDir := path.Join(excelTmpPath, taskID)
	helper.FolderZip(taskDir, zipFilePath)
	logx.InfoF(ctx, "remove Files dir: %s, path: %s, err1: %v, err2: %v", taskDir, paramFilePath, os.RemoveAll(taskDir), os.Remove(paramFilePath))

	os.RemoveAll(zipFilePath)

	// 6. 修改任务状态，写文件
	expLog.FileUrl = zipFilePath
	expLog.Status = model.Status_succ
	expLog.UpdatedAt = utils.TimeFrom(time.Now())
	err = models.UpDataTask(ctx, &expLog)
	if err != nil {
		logx.ErrorF(ctx, "更新数据失败："+err.Error())
	}
	//回调通知
	w.req.Notify(ctx, expLog.Callback, taskID)
	log.Print(taskID, "任务完成")
}
