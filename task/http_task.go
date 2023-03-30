package task

import (
	"cloud-export/conf"
	"cloud-export/model"
	"cloud-export/model/excel"
	"cloud-export/model/helper"
	"cloud-export/server"
	"encoding/json"
	"fmt"
	"github.com/laydong/toolpkg"
	"github.com/laydong/toolpkg/logx"
	"github.com/laydong/toolpkg/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"runtime"
	"time"
)

type SourceHTTP struct {
	URL    string                 `json:"url"`
	Method string                 `json:"method"`
	Header map[string]string      `json:"header"`
	Param  map[string]interface{} `json:"param"`
}

type HttpWorker struct {
	Tasks  *server.Mq
	req    *server.SourceHTTP
	taskCh chan *server.ExportTask
}

func (w *HttpWorker) Run(pool int) {
	// 单消费端 多任务执行
	w.Tasks = &server.Mq{Key: server.TaskHttpKey}
	w.req = server.NewSourceHTTP()
	// 缓冲区越大，程序宕机后丢消息越多
	w.taskCh = make(chan *server.ExportTask, 20)
	log.Print("httpWorker pool=", pool)
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
func (w *HttpWorker) startWorker(pool int) {
	for i := 0; i < pool; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {

					log.Print("[runtime] err ,recoverd", "", fmt.Errorf("%v", err).Error())

					ctx := toolpkg.GetNewGinContext()
					logx.ErrorF(ctx, "异常："+fmt.Sprintf("%v", err))

					buf := make([]byte, 64<<10)
					runtime.Stack(buf, false)
					track := []byte{}
					for _, i := range buf {
						if i == 0 {
							break
						}
						track = append(track, i)
					}
					logx.ErrorF(ctx, "调用："+string(track))
				}
			}()
			for {
				w.work()
			}
		}()
	}
}

// work 处理单个任务
func (w *HttpWorker) work() {
	currTask := <-w.taskCh
	taskID := currTask.TaskID
	//st := carbon.Now()
	ctx := toolpkg.GetNewGinContext()
	if taskID == "" {
		logx.ErrorF(ctx, "TaskNotFund hash_key: %s", taskID)
		return
	}
	// 1. 数据库中查询任务详情
	models := new(model.ExportLogModel)
	expLog, err := models.QueryByHashKey(ctx, taskID)
	if err != nil {
		logx.ErrorF(ctx, "TaskNotFund hash_key: %s", taskID)
		return
	}
	if expLog.Id == 0 {
		logx.ErrorF(ctx, "TaskNotFund hash_key: %s", taskID)
		return
	}
	// 任务取消
	if expLog.Status == model.Status_cancle {
		return
	}
	requestParam := SourceHTTP{}
	err = json.Unmarshal([]byte(expLog.Param), &requestParam)
	if err != nil {
		expLog.FailReason = "参数解析失败：" + err.Error()
		expLog.Status = model.Status_fail
		expLog.UpdatedAt = utils.TimeFrom(time.Now())
		err = models.UpDataTask(ctx, &expLog)
		if err != nil {
			logx.ErrorF(ctx, "更新数据失败："+err.Error())
		}
		w.req.Notify(ctx, expLog.Callback, taskID)
		return
	}

	// 2. 获取数据源的数据 -> 3. 写入excel
	baseParam := &server.HttpParam{
		Page:   1,
		Url:    requestParam.URL,
		Method: requestParam.Method,
		Header: requestParam.Header,
		Param:  requestParam.Param,
	}
	// 带上此次任务ID 方便日志追踪
	if baseParam.Header == nil {
		baseParam.Header = make(map[string]string)
	}
	baseParam.Header["xt-export-taskId"] = taskID
	// 获取第一页，获取分页信息
	_, totalPage, lists, err := w.req.FirstPage(ctx, baseParam)
	page := 1
	log.Printf("[%s] 抓取到(%d/%d)页\n", taskID, page, totalPage)
	logx.InfoF(ctx, fmt.Sprintf("抓取到(%d/%d)页, taskid: %s, page: %d, totalPage: %d", page, totalPage, taskID, page, totalPage))
	if err != nil || lists == "" || totalPage == 0 {
		reason := "获取数据源失败：" + err.Error()
		logx.ErrorF(ctx, "获取数据源失败, taskID: %s, err: %s", taskID, err.Error())
		expLog.FailReason = reason
		expLog.Status = model.Status_fail
		expLog.UpdatedAt = utils.TimeFrom(time.Now())
		err = models.UpDataTask(ctx, &expLog)
		if err != nil {
			logx.ErrorF(ctx, "更新数据失败："+err.Error())
		}
		w.req.Notify(ctx, expLog.Callback, taskID)
		return
	}

	excelTmpPath := conf.App.Upload.FileUrl // excel 临时文件目录
	if excelTmpPath == "" {
		excelTmpPath = "/home/outexcel"
	}
	err = helper.TouchDir(excelTmpPath)
	if err != nil {
		logx.ErrorF(ctx, "文件夹创建失败："+err.Error())
		return
	}
	maxlines := viper.GetInt("storage.excel_maxlines") // excel 最大行数
	if maxlines == 0 {
		maxlines = 5000
	}
	conn := viper.GetInt("storage.http_req_conn") // http并发最大请求数
	if conn == 0 {
		conn = 5
	}
	logx.InfoF(ctx, fmt.Sprintf("导出配置：outexcel_tmp %s, excel_maxlines %d, http_req_conn %d", excelTmpPath, maxlines, conn))
	filename := expLog.Title + "-%d." + expLog.ExtType
	excelw := excel.NewExcelRecorderPage(path.Join(excelTmpPath, taskID, filename), maxlines)
	p := excelw.WritePagpenate(excel.Pos{X: 1, Y: 1}, lists, "", true)
	page += 1
	var end bool

	logx.InfoF(ctx, "开始采集数据")
	for { // 获取剩下页
		params := make([]server.HttpParam, 0, conn)
		for j := 0; j < conn; j++ {
			if page > totalPage {
				end = true
				break
			}
			baseParam.Page = page
			params = append(params, *baseParam)
			log.Printf("[%s] 开始抓取(%d/%d)页\n", taskID, page, totalPage)
			logx.InfoF(ctx, "开始抓取(%d/%d)页,taskId: %s, page: %s, totalPage: %s", taskID, page, totalPage)
			page += 1
		}
		listdata, err := w.req.BatchRequest(ctx, params...)
		if err != nil {
			reason := "获取数据源失败：" + err.Error()
			logx.ErrorF(ctx, "获取数据源失败, taskId: %s, err: %s", taskID, err.Error())
			expLog.FailReason = reason
			expLog.Status = model.Status_fail
			expLog.UpdatedAt = utils.TimeFrom(time.Now())
			err = models.UpDataTask(ctx, &expLog)
			if err != nil {
				logx.ErrorF(ctx, "更新数据失败："+err.Error())
			}
			w.req.Notify(ctx, expLog.Callback, taskID)
		}
		if len(listdata) == 0 {
			reason := "获取数据源失败：未获取到数据"
			logx.ErrorF(ctx, "获取数据源失败", taskID, "未获取到数据")
			expLog.FailReason = reason
			expLog.Status = model.Status_fail
			expLog.UpdatedAt = utils.TimeFrom(time.Now())
			err = models.UpDataTask(ctx, &expLog)
			if err != nil {
				logx.ErrorF(ctx, "更新数据失败："+err.Error())
			}
			w.req.Notify(ctx, expLog.Callback, taskID)
		}
		// 有序写入excel
		for _, alist := range listdata {
			p = excelw.WritePagpenate(p, alist, "", false)
		}
		if end {
			break
		}
	}
	logx.InfoF(ctx, "数据采集完成")

	if err = excelw.Save(); err != nil {
		logx.ErrorF(ctx, "Excel 文件保存失败："+err.Error())
	}
	logx.InfoF(ctx, "Excel 文件保存成功")

	// 4. 压缩文件夹 并删除源文件
	zipFilePath := path.Join(excelTmpPath, taskID+".zip")
	taskDir := path.Join(excelTmpPath, taskID)
	helper.FolderZip(taskDir, zipFilePath)
	if err = os.RemoveAll(taskDir); err != nil {
		logx.ErrorF(ctx, "临时文件夹删除失败："+err.Error())
	}
	logx.InfoF(ctx, "Excel 文件已打包，临时文件删除成功")

	//// 5. 上传云 OSS -> 删除本地文件
	//objname, err := aoss.PutExportFile(zipFilePath)
	//if err != nil {
	//	reason := "上传阿里云oss失败：" + err.Error()
	//	expLog.FailReason = reason
	//	expLog.Status = model.Status_fail
	//	expLog.UpdatedAt = global.TimeFrom(time.Now())
	//	err = models.UpDataTask(ctx, &expLog)
	//	if err != nil {
	//		glogs.ErrorF(ctx, "更新数据失败："+err.Error())
	//	}
	//	return
	//}
	//glogs.InfoF(ctx, "Excel 打包文件上传阿里云成功")

	//if err := os.RemoveAll(zipFilePath); err != nil {
	//	glogs.ErrorF(ctx, "删除临时压缩包失败："+err.Error())
	//}
	//glogs.InfoF(ctx, "Excel 打包文件删除成功")

	// 6. 修改任务状态
	expLog.FileUrl = zipFilePath
	expLog.Status = model.Status_succ
	expLog.UpdatedAt = utils.TimeFrom(time.Now())
	err = models.UpDataTask(ctx, &expLog)
	if err != nil {
		logx.ErrorF(ctx, "更新数据失败："+err.Error())
	}
	w.req.Notify(ctx, expLog.Callback, taskID)
	//dt := carbon.Now().DiffInSecondsWithAbs(st)
	//log.Printf("[%s] 任务完成 耗时%ds", taskID, dt)
	logx.InfoF(ctx, "任务完成,taskId: %s, 耗时%ds", taskID)
}
