package src

import (
  "fmt"
  "github.com/asticode/go-astilog"
)

const DEBUG = false

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

func init() {
  connectionList = []connection{}
  GetCacheDir(DEBUG)
  fmt.Println("baseDir", CacheDir)
  astilog.SetLogger(astilog.New(astilog.Configuration{
    AppName:  "RedisManager",
    Filename: CacheDir + `/rdm.log`,
    Verbose:  DEBUG,
  }))
  astilog.FlagConfig()
  astilog.Infof("baseDir: %s", CacheDir)
}
