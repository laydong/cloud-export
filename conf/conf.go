package conf

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var App *Config

func InitApp(path string) error {
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.WatchConfig()
	initConf()
	viper.OnConfigChange(func(_ fsnotify.Event) {
		initConf()
	})
	return nil
}

func initConf() {
	App = &Config{
		AppConf: AppConf{
			AppName:    viper.GetString("app.app_name"),
			AppMode:    viper.GetString("app.app_mode"),
			Port:       viper.GetString("app.port"),
			HttpListen: viper.GetString("app.http_listen"),
			Url:        viper.GetString("app.url"),
			IsSso:      viper.GetBool("app.is_sso"),
			Params:     viper.GetBool("app.params"),
			Logger:     viper.GetString("app.logger"),
			Version:    viper.GetString("app.version"),
		},
		Jwt: Jwt{
			SigningKey:  viper.GetString("jwt.signing_key"),
			ExpiresTime: viper.GetInt64("jwt.expires_time"),
			BufferTime:  viper.GetInt64("jwt.buffer_time"),
		},
		DBConf: DBConf{
			Dsn:             viper.GetString("mysql.dsn"),
			MaxIdleConn:     viper.GetInt("mysql.max_idle_conn"),
			MaxOpenConn:     viper.GetInt("mysql.max_open_conn"),
			ConnMaxLifetime: viper.GetInt("mysql.conn_max_lifetime"),
			DbName:          viper.GetString("mysql.db_name"),
			Prefix:          viper.GetString("mysql.prefix"),
		},
		MGConf: MGConf{
			Dsn:             viper.GetString("mongodb.dsn"),
			ConnTimeOut:     viper.GetInt("mongodb.conn_time_out"),
			ConnMaxPoolSize: viper.GetInt("mongodb.conn_max_pool_size"),
		},
		RDConf: RDConf{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		TaskPool: TaskPool{
			HttpWorker: viper.GetInt("task_pool.http_worker"),
			RowWorker:  viper.GetInt("task_pool.row_worker"),
		},
		Upload: Upload{
			FileUrl: viper.GetString("upload.file_url"),
		},
	}
}

type Config struct {
	AppConf  AppConf
	Jwt      Jwt
	DBConf   DBConf
	MGConf   MGConf
	RDConf   RDConf
	TaskPool TaskPool
	Upload   Upload
}

type AppConf struct {
	AppName    string `json:"app_name"`
	AppMode    string `json:"app_mode"`
	Port       string `json:"port"`
	HttpListen string `json:"http_listen"`
	Url        string `json:"url"`
	IsSso      bool   `json:"is_sso"`
	Params     bool   `json:"params"`
	Logger     string `json:"logger"`
	Version    string `json:"version"`
}

type Jwt struct {
	SigningKey  string `json:"signing_key"`
	ExpiresTime int64  `json:"expires_time"`
	BufferTime  int64  `json:"buffer_time"`
}

type DBConf struct {
	MaxIdleConn     int    `json:"max_idle_conn"`
	MaxOpenConn     int    `json:"max_open_conn"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	Dsn             string `json:"dsn"`
	DbName          string `json:"db_name"`
	Prefix          string `json:"prefix"`
}

type MGConf struct {
	ConnTimeOut     int    `json:"conn_time_out"`
	ConnMaxPoolSize int    `json:"conn_max_pool_size"`
	Dsn             string `json:"dsn"`
}

type RDConf struct {
	DB       int    `json:"db"`       // redis的哪个数据库
	Addr     string `json:"addr"`     // 服务器地址:端口
	Password string `json:"password"` // 密码
}

type TaskPool struct {
	HttpWorker int `json:"http_worker"`
	RowWorker  int `json:"row_worker"`
}

type Upload struct {
	FileUrl string `json:"file_url"`
}
