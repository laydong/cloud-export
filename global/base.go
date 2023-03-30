package global

import (
	"cloud-export/conf"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/laydong/toolpkg/db"
	"gorm.io/gorm"
)

var DB *gorm.DB
var Rdb *redis.Client

func GetDB(c *gin.Context, dbNmae ...string) *gorm.DB {
	key := conf.App.DBConf.DbName
	if len(dbNmae) > 0 {
		key = dbNmae[0]
	}
	if key == "" {
		key = "grom_cxt"
	}
	return DB.Set(key, c).WithContext(c)
}
func InitApp() (err error) {
	DB, err = db.InitDB(conf.App.DBConf.Dsn)
	if err != nil {
		return
	}
	Rdb, err = db.InitRdb(conf.App.RDConf.Addr, conf.App.RDConf.Password, conf.App.RDConf.DB)
	if err != nil {
		return
	}
	return
}
