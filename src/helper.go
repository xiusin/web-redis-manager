package src

import (
  "sync"

  "github.com/astaxie/beego/logs"
  "github.com/gomodule/redigo/redis"
)

const DEBUG  = true


type redisClient struct {
  conn   redis.Conn
  locker sync.Mutex
}

type ResponseData struct {
  Status int64       `json:"status"`
  Msg    string      `json:"msg"`
  Data   interface{} `json:"data"`
}

type connection struct {
  ID    int64  `json:"id"`
  Title string `json:"title"`
  Ip    string `json:"ip"`
  Port  int    `json:"port"`
  Auth  string `json:"auth"`
}

var (
  CacheDir        string
  totalConnection = 0
  connectionList  []connection
  jsonFile        string
)

func init()  {
  connectionList = []connection{}
  GetCacheDir(DEBUG)
  log := logs.NewLogger(10)
  _ = log.SetLogger("file", `{"filename":"`+CacheDir+`/rdm-error.log"}`)
}
