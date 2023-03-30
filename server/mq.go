package server

import (
	"cloud-export/global"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/laydong/toolpkg/logx"
	"log"
)

type Mq struct {
	Key string
}

type MqMsg interface {
	String() string
	Build(string) error
}

func (mq *Mq) Push(ctx *gin.Context, msg MqMsg) {
	res := global.Rdb.LPush(global.Rdb.Context(), mq.Key, msg.String())
	if err := res.Err(); err != nil {
		logx.ErrorF(ctx, "PushQueue err: %s", err.Error())
	}
}

// 消费者，常驻内存
func (mq *Mq) BPop(hander func(string)) {
	for {
		// 阻塞式监听该key
		res := global.Rdb.BRPop(global.Rdb.Context(), 0, mq.Key)
		err := res.Err()
		if err == nil {
			hander(res.Val()[1])
		}
		if err == redis.Nil {
			log.Print("queueIsEmpty")
		}
	}
}

type ExportTask struct {
	TaskID string
}

func (t *ExportTask) String() string {
	jsonstr, _ := json.Marshal(t)
	return string(jsonstr)
}

func (t *ExportTask) Build(jsonStr string) (err error) {
	return json.Unmarshal([]byte(jsonStr), t)
}
