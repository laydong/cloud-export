package model

import (
	"cloud-export/global"
	"github.com/gin-gonic/gin"
	"github.com/laydong/toolpkg/utils"
	"gorm.io/gorm"
)

type ExportLogModel struct {
	Id         int            `gorm:"column:id;primary_key;auto_increment" json:"id"`
	HashKey    string         `gorm:"column:hash_key" json:"hash_key"`                           //参数哈希
	Title      string         `gorm:"column:title" json:"title"`                                 //导出标题
	ExtType    string         `gorm:"column:ext_type" json:"ext_type"`                           //导出类型(文件后缀)
	SourceType string         `gorm:"column:source_type" json:"source_type"`                     //数据源类型
	Param      string         `gorm:"column:param" json:"param"`                                 //请求参数（json）
	FileUrl    string         `gorm:"column:file_url" json:"file_url"`                           //文件地址
	Callback   string         `gorm:"column:callback" json:"callback"`                           //回调地址
	Status     int            `gorm:"column:status" json:"status"`                               //状态：1处理中 2导出成功 3导出失败 4导出取消
	FailReason string         `gorm:"column:fail_reason" json:"fail_reason"`                     //失败理由
	CreatedAt  utils.Time     `gorm:"column:created_at;NOT NULL;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt  utils.Time     `gorm:"column:updated_at;NOT NULL;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`                               // 删除时间
}

// 导出详情
type ExportLogDetail struct {
	Log ExportLogModel `json:"log"`
	//File    ExportFile `json:"file"`
	DownUrl string `json:"down_url"`
}

const (
	StypeHttp      = "http"
	StypeSql       = "sql"
	StypeRaw       = "raw"
	Status_pending = 1 // 处理中
	Status_succ    = 2 // 导出成功
	Status_fail    = 3 // 导出失败
	Status_cancle  = 4 // 导出取消
)

func (e *ExportLogModel) TableName() string {
	return "b_export_log"
}

func (e *ExportLogModel) QueryByHashKey(c *gin.Context, hashKey string) (data ExportLogModel, err error) {
	err = global.DB.Model(ExportLogModel{}).Where("hash_key", hashKey).Find(&data).Error
	return
}

func (e *ExportLogModel) CreateTask(c *gin.Context, expLog *ExportLogModel) (err error) {
	err = global.DB.Model(ExportLogModel{}).Create(&expLog).Error
	return
}

func (e *ExportLogModel) UpDataTask(c *gin.Context, expLog *ExportLogModel) (err error) {
	err = global.DB.Model(ExportLogModel{}).Where("id", expLog.Id).Updates(&expLog).Error
	return
}
