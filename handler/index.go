package handler

import (
	"cloud-export/model/request"
	"cloud-export/server"
	"cloud-export/utils"
	"github.com/gin-gonic/gin"
)

// ExportSHttp http接口导出excel
func ExportSHttp(c *gin.Context) {
	param := request.ExpSHttpParam{}
	if err := c.ShouldBind(&param); err != nil {
		utils.FailWithMessage(c, err.Error())

		return
	}
	//param.Timestamp = time.Now().Unix()
	//param.EXTType = "xlsx"
	//param.Title = "测试导出"
	//param.CallBack = ""
	//param.SourceHTTP = request.SourceHTTP{
	//	URL:    "http://127.0.0.1/api/test",
	//	Method: "post",
	//	Param: map[string]interface{}{
	//		"per_page": 10,
	//	},
	//}
	data, err := server.HandelSHttp(c, &param)
	if err != nil {
		utils.FailWithMessage(c, err.Error())
		return
	}
	utils.OkWithData(c, data)
}

// ExportSRaw 数据导出excel
func ExportSRaw(c *gin.Context) {
	param := request.ExpSRawParam{}
	if err := c.ShouldBind(&param); err != nil {
		utils.FailWithMessage(c, err.Error())
		return
	}
	data, err := server.HandelSRaw(c, &param)
	if err != nil {
		utils.FailWithMessage(c, err.Error())
		return
	}
	utils.OkWithData(c, data)
}

func ExportDetail(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		utils.FailWithMessage(c, "key 不能为空")
		return
	}
	data, err := server.Detail(c, key)
	if err != nil {
		utils.FailWithMessage(c, err.Error())
		return
	}
	utils.OkWithData(c, data)
}

//// ExportSRaw 源数据导出excel
//func ExportSRaw(c *gin.Context) {
//	param := valid.ExpSRawParam{}
//	err := valid.BindAndCheck(c, &param)
//	if err != nil {
//		r.Fail(c, err)
//		return
//	}
//	exportServ := new(page.ExportServ)
//	data, err := exportServ.HandelSRaw(c, &param)
//	if err != nil {
//		r.Fail(c, err)
//		return
//	}
//	r.Succ(c, data)
//
//}
//
//func ExportDetail(c *gin.Context) {
//	key := c.Query("key")
//	if key == "" {
//		r.Fail(c, exception.ParamInValid("key 不能为空"))
//		return
//	}
//
//	data, err := new(page.ExportServ).Detail(c, key)
//	if err != nil {
//		r.Fail(c, err)
//		return
//	}
//	r.Succ(c, data)
//}
//
//func ExportHistory(c *gin.Context) {
//	userInfo := &valid.EmployeeInfo{}
//	// 从header中解析员工ID
//	userInfo.Parse(c)
//	param := valid.ExpLogHistory{
//		Uid: userInfo.ID,
//	}
//	data, err := new(page.ExportServ).History(c, &param)
//	if err != nil {
//		r.Fail(c, err)
//		return
//	}
//	r.Succ(c, data)
//}

//func Test(c *gin.Context) {
//	param := PageList{}
//	if err := c.ShouldBind(&param); err != nil {
//		utils.FailWithMessage(c, err.Error())
//		return
//	}
//	var data []Date
//	for i := 1; i <= param.PerPage; i++ {
//		if i == 0 {
//			data = append(data, Date{
//				ID:   "ID",
//				Name: "名字",
//				Sex:  "性别",
//			})
//		}
//		data = append(data, Date{
//			ID:   strconv.Itoa(i + (param.Page-1)*param.PerPage),
//			Name: "名字" + strconv.Itoa(i),
//			Sex:  strconv.Itoa(i),
//		})
//	}
//	utils.OkWithData(c, CutPage(50, param.Page, param.PerPage, len(data), data))
//}
//
//type Date struct {
//	ID   string `json:"id"`
//	Name string `json:"name"`
//	Sex  string `json:"sex"`
//}
//
//type PageData struct {
//	Meta struct {
//		Pagination struct {
//			Count       int   `json:"count"`
//			CurrentPage int   `json:"current_page"`
//			PerPage     int   `json:"per_page"`
//			Total       int64 `json:"total"`
//			TotalPages  int   `json:"total_pages"`
//		} `json:"pagination"`
//	} `json:"meta"`
//	Data interface{} `json:"data"`
//}
//
//func CutPage(total int64, page, size, count int, data interface{}) (pager PageData) {
//	pager.Meta.Pagination.Total = total
//	pager.Meta.Pagination.TotalPages = int(int64(math.Ceil(float64(total) / float64(size))))
//	pager.Meta.Pagination.CurrentPage = page
//	pager.Meta.Pagination.PerPage = size
//	pager.Meta.Pagination.Count = count
//	pager.Data = data
//	return
//}
//
//type PageList struct {
//	Page    int `json:"page,default=1" form:"page,default=1" uri:"page,default=1" binding:"required"`
//	PerPage int `json:"per_page,default=10" form:"per_page,default=10" uri:"per_page,default=10" binding:"required"`
//}
