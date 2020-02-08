package src

type ResponseData struct {
  Status int64       `json:"status"`
  Msg    string      `json:"msg"`
  Data   interface{} `json:"data"`
}

type connection struct {
  ID    int64  `json:"id"`
  Title string `json:"title"`
  Ip    string `json:"ip"`
  Port  string `json:"port"`
  Auth  string `json:"auth"`
}

var (
  totalConnection = 0
  connectionList  []connection
  ConnectionFile  string
  SecretKey string
)
