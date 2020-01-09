package src

import (
  "encoding/json"
  "errors"
  "fmt"
  "github.com/asticode/go-astilectron"
  "github.com/asticode/go-astilog"
  "io"
  "io/ioutil"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "time"

  "github.com/gomodule/redigo/redis"
)

var Window *astilectron.Window

func GetCacheDir(debug bool) string {
  if "" != CacheDir {
    return CacheDir
  }
  var workingDir string
  if debug {
    workingDir, _ = os.Getwd()
  } else {
    workingDir, _ = os.Executable()
    workingDir = filepath.Dir(workingDir)
  }
  CacheDir = workingDir
  jsonFile = CacheDir + "/rdm-connections.json"
  return CacheDir
}

func RedisManagerGetInfo(data map[string]interface{}) string {
  client, _, _ := getRedisClient(data, false, false)
  defer client.Close()
  d, err := redis.String(client.Do("INFO"))
  if err != nil {
    return JSON(ResponseData{5000, "读取服务器信息失败:" + err.Error(), nil})
  }

  c, err := redis.Strings(client.Do("CONFIG", "GET", "*"))
  if err != nil {
    return JSON(ResponseData{5000, "读取配置文件失败:" + err.Error(), nil})
  }
  return JSON(ResponseData{200, "保存成功", map[string]interface{}{
    "data":   d,
    "config": c,
  }})
}

func RedisManagerConnectionTest(data map[string]interface{}) string {
  config := connection{}
  config.Ip = data["ip"].(string)
  config.Title = data["title"].(string)
  config.Port = int(data["port"].(float64))
  config.Auth = data["auth"].(string)

  client, err := redis.Dial("tcp", config.Ip+":"+strconv.Itoa(config.Port))
  if err != nil {
    return JSON(ResponseData{5000, "Connect to redis error" + err.Error(), err.Error()})
  }
  defer client.Close()
  if config.Auth != "" {
    _, err := client.Do("AUTH", config.Auth)
    if err != nil {
      return JSON(ResponseData{5000, "auth: " + err.Error(), err.Error()})
    }
  }
  return JSON(ResponseData{200, "连接成功", nil})
}

func RedisManagerConfigSave(data map[string]interface{}) string {
  config := connection{}
  config.Ip = data["ip"].(string)
  config.Title = data["title"].(string)
  config.Port = int(data["port"].(float64))
  config.Auth = data["auth"].(string)
  totalConnection = totalConnection + 1
  config.ID = int64(totalConnection)
  connectionList = append(connectionList, config)
  err := writeConfigJSON()
  if err != nil {
    return JSON(ResponseData{5000, "保存成功:" + err.Error(), nil})
  }
  return JSON(ResponseData{200, "保存成功", config})
}

func RedisManagerConnectionList(data map[string]interface{}) string {
  err := readConfigJSON()
  if err != nil {
    return JSON(ResponseData{5000, "获取列表失败:" + err.Error(), nil})
  }
  return JSON(ResponseData{200, "获取列表成功", connectionList})
}

var pubsubs = map[string]bool{}

func RedisPubSub(data map[string]interface{}) string {
  client, _, err := getRedisClient(data, false, false)
  id := int(data["id"].(float64))
  if err != nil {
    return JSON(ResponseData{5000, "连接错误" + err.Error(), nil})
  }
  channels, err := redis.Strings(client.Do("PUBSUB", "channels"))
  if err != nil {
    return JSON(ResponseData{5000, "获取订阅列表失败", err.Error()})
  }

  if channel, ok := data["channel"]; ok {
    msg := data["msg"]
    if msg == "" || channel == "" {
      return JSON(ResponseData{5000, "发布内容失败", nil})
    }
    // 先查看是否有消费者订阅频道
    var flag bool
    for _, ch := range channels {
      if ch == channel {
        flag = true
        break
      }
    }
    if !flag {
      // 订阅该渠道
      go func() {
        client, _, _ := getRedisClient(data, false, false)
        defer client.Close()
        pubsub := redis.PubSubConn{Conn: client}
        if err := pubsub.Subscribe(channel); err != nil {
          panic(err)
        }
        for range time.Tick(time.Second * 10) {
          pubsub.ReceiveWithTimeout(time.Second)
        }
      }()
    }
    _, err := client.Do("PUBLISH", channel, msg)
    if err != nil {
      return JSON(ResponseData{5000, "发布内容失败", err.Error()})
    } else {
      return JSON(ResponseData{200, "发布内容成功", nil})
    }
  } else {

    if ok, _ := pubsubs[fmt.Sprintf("channel-%d", id)]; !ok {
      pubsubs[fmt.Sprintf("channel-%d", id)] = true
      go func() {
        defer func(id int) {
          pubsubs[fmt.Sprintf("channel-%d", id)] = false
        }(id)
        pubsub := redis.PubSubConn{Conn: client}
        if err := pubsub.PSubscribe("*"); err != nil {
          panic(err)
        }
        for {
          message := pubsub.Receive()
          switch v := message.(type) {
          case redis.Message: //单个订阅subscribe
            Window.SendMessage(map[string]string{
              "data":    string(v.Data),
              "id":      strconv.Itoa(id),
              "channel": v.Channel,
              "time":    time.Now().In(loc).Format("13:04:05"),
            }, func(m *astilectron.EventMessage) {
              astilog.Debugf("received %s", m)
            })
          case error:
            panic(v)
          default:
          }
        }
      }()
    }
    return JSON(ResponseData{200, "获取列表成功", channels})
  }
}

var loc, _ = time.LoadLocation("PRC")

func RedisManagerCommand(data map[string]interface{}) string {
  client, _, err := getRedisClient(data, true, false)
  if err != nil {
    return JSON(ResponseData{5000, "连接错误: " + err.Error(), nil})
  }

  command, ok := data["command"]
  if !ok {
    return JSON(ResponseData{5000, "命令错误", nil})
  }
  commands := strings.Split(command.(string), " ")
  var flags []interface{}

  for _, v := range commands[1:] {
    flags = append(flags, v)
  }

  val, err := client.Do(commands[0], flags...)
  if err != nil {
    return JSON(ResponseData{5000, "获取数据错误", err.Error()})
  }

  switch val.(type) {
  case []byte:
    res, _ := redis.String(val, nil)
    return JSON(ResponseData{200, "成功", res})
  case []interface{}:
    res, err := redis.StringMap(val, nil)
    if err != nil {
      return JSON(ResponseData{200, "成功", err.Error()})

    }
    var strs []string
    var i int
    switch commands[0] {
    case "HGETALL":
      for k, v := range res {
        i++
        strs = append(strs, strconv.Itoa(i)+"): "+k)
        i++
        strs = append(strs, strconv.Itoa(i)+"): "+v)
      }
    case "LRANGE":
      for k, v := range res {
        i++
        strs = append(strs, strconv.Itoa(i)+"): "+k)
        i++
        strs = append(strs, strconv.Itoa(i)+"): "+v)
      }
    case "SMEMBERS":
      for k, v := range res {
        i++
        strs = append(strs, strconv.Itoa(i)+"): "+k)
        i++
        strs = append(strs, strconv.Itoa(i)+"): "+v)
      }
    default:
      strs = append(strs, fmt.Sprintf("%#v", res))
    }
    return JSON(ResponseData{200, "成功", strings.Join(strs, "<br/>")})

  default:
    return JSON(ResponseData{200, "成功", fmt.Sprintf("%#v", val)})
  }
}

func RedisManagerRemoveConnection(data map[string]interface{}) string {
  var configs = []connection{}
  id := int64(data["id"].(float64))
  if id == 0 {
    return JSON(ResponseData{200, "参数失败", nil})
  }
  for _, v := range connectionList {
    if v.ID != id {
      configs = append(configs, v)
    }
  }
  connectionList = configs
  err := writeConfigJSON()
  if err != nil {
    return JSON(ResponseData{5000, "删除失败:" + err.Error(), nil})
  }
  return JSON(ResponseData{200, "删除成功", nil})
}

func getRedisClient(data map[string]interface{}, getSelectedIndexClient bool, getKey bool) (redis.Conn, string, error) {
  var config connection
  id := int(data["id"].(float64))
  if id == 0 {
    return nil, "", errors.New("参数错误")
  }
  config.ID = int64(id)
  var client redis.Conn
  var err error
  for _, v := range connectionList {
    if v.ID == config.ID {
      config = v
      break
    }
  }
  if config.Title == "" {
    return nil, "", errors.New("没有该连接项")
  }
  client, err = redis.Dial("tcp", config.Ip+":"+strconv.Itoa(config.Port))
  if err != nil {
    return nil, "", err
  }
  if config.Auth != "" {
    _, err = client.Do("AUTH", config.Auth)
    if err != nil {
      return nil, "", err
    }
  }

  if getSelectedIndexClient {
    index := int(data["index"].(float64))
    _, _ = client.Do("SELECT", index) //选择数据库
  }
  var key string
  if getKey {
    key = data["key"].(string)
    if key == "" {
      return nil, key, errors.New("请选择要操作的key")
    }
  } else {
    key = ""
  }
  return client, key, nil
}

func RedisManagerConnectionServer(data map[string]interface{}) string {
  client, _, err := getRedisClient(data, false, false)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  action := strings.Trim(data["action"].(string), " ")
  switch action {
  case "get_value":
    index := int(data["index"].(float64))
    _, err = client.Do("SELECT", index) //选择数据库
    if err != nil {
      return JSON(ResponseData{5000, "选择数据库失败", nil})
    }
    key := data["key"].(string)
    if key == "" {
      return JSON(ResponseData{5000, "请选择key", nil})
    }
    typeStr, _ := redis.String(client.Do("TYPE", key))
    if typeStr == "none" {
      return JSON(ResponseData{5001, "缓存不存在或已过期", nil})
    }
    ttl, _ := redis.Int64(client.Do("TTL", key))
    switch typeStr {
    case "list":
      val, err := redis.Strings(client.Do("LRANGE", key, 0, 1000))
      if err != nil {
        return JSON(ResponseData{5000, "读取数据错误", err.Error()})
      } else {
        return JSON(ResponseData{200, "读取所有key成功", map[string]interface{}{
          "type": typeStr,
          "data": val,
          "ttl":  ttl,
        }})
      }
    case "set":
      val, err := redis.Strings(client.Do("SMEMBERS", key))
      if err != nil {
        return JSON(ResponseData{5000, "读取数据错误", err.Error()})
      } else {
        return JSON(ResponseData{200, "读取所有key成功", map[string]interface{}{
          "type": typeStr,
          "data": val,
          "ttl":  ttl,
        }})
      }
    case "zset":
      val, err := redis.StringMap(client.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
      if err != nil {
        return JSON(ResponseData{5000, "读取数据错误", err.Error()})
      } else {
        var retData []map[string]string
        for k, v := range val {
          retData = append(retData, map[string]string{"value": k, "score": v})
        }
        return JSON(ResponseData{200, "读取所有key成功", map[string]interface{}{
          "type": typeStr,
          "data": retData,
          "ttl":  ttl,
        }})
      }
    case "string":
      val, err := redis.String(client.Do("GET", key))
      if err != nil {
        return JSON(ResponseData{5000, "读取数据错误", err.Error()})
      } else {
        return JSON(ResponseData{200, "读取所有key成功", map[string]interface{}{
          "type": typeStr,
          "data": val,
          "ttl":  ttl,
        }})
      }
    case "hash":
      val, err := redis.StringMap(client.Do("HGETALL", key))
      if err != nil {
        return JSON(ResponseData{5000, "读取数据错误", err.Error()})
      } else {
        return JSON(ResponseData{200, "读取所有key成功", map[string]interface{}{
          "type": typeStr,
          "data": val,
          "ttl":  ttl,
        }})
      }
    }
  case "dblist":
    //读取数据库列表
    var dbs []int
    for i := 0; i < 20; i++ {
      _, err := client.Do("SELECT", i)
      if err != nil {
        break
      }
      //读取数据总量
      total, _ := redis.Int(client.Do("DBSIZE"))
      dbs = append(dbs, total)
    }
    return JSON(ResponseData{200, "连接数据库成功", dbs})
  case "select_db":
    index := int(data["index"].(float64))
    _, _ = client.Do("SELECT", index) //选择数据库
    //todo 这里要优化
    keys, err := redis.Strings(client.Do("KEYS", "*"))
    if err != nil {
      return JSON(ResponseData{5000, "读取数据错误", err.Error()})
    }
    var reskeys = map[string][]string{}
    for _, v := range keys {
      //strs := strings.Split(v, ":")
      //fmt.Println(strs)
      //if len(strs) > 1 {
      //	//_, ok := reskeys[strs[0]]
      //	reskeys[strs[0]] = append(reskeys[strs[0]], strings.Join(strs[1:], ":"))
      //} else {
      reskeys[v] = append(reskeys[v], v)
      //}
    }
    return JSON(ResponseData{200, "读取所有key成功", reskeys})
  }
  return JSON(ResponseData{5000, "错误,无法解析到动作:" + action, nil})
}

func RedisManagerRemoveKey(data map[string]interface{}) string {
  client, key, err := getRedisClient(data, true, true)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  _, err = client.Do("DEL", key)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  return JSON(ResponseData{200, "删除成功", nil})
}

func RedisManagerFlushDB(data map[string]interface{}) string {
  client, _, err := getRedisClient(data, true, false)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  _, err = client.Do("FLUSHDB")
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  return JSON(ResponseData{200, "清空数据库成功", nil})
}

func RedisManagerRemoveRow(data map[string]interface{}) string {
  client, key, err := getRedisClient(data, true, true)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  valType := data["type"].(string)
  if valType == "" {
    return JSON(ResponseData{5000, "无法解析数据类型", nil})
  }
  switch valType {
  case "list":
    _, err = client.Do("LREM", key, 1, data["data"])
  case "set":
    _, err = client.Do("SREM", key, data["data"])
  case "zset":
    _, err = client.Do("ZREM", key, data["data"].(string))
  case "hash":
    _, err = client.Do("HDEL", key, data["data"])
  }
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  return JSON(ResponseData{200, "删除成功", nil})
}

func RedisManagerUpdateKey(data map[string]interface{}) string {
  client, key, err := getRedisClient(data, true, true)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  action := data["action"].(string)
  switch action {
  case "ttl": //更新ttl时间
    ttl := int(data["ttl"].(float64))
    _, err = client.Do("EXPIRE", key, ttl)
    if err != nil {
      return JSON(ResponseData{5000, "操作失败", nil})
    }
  case "value": //更新value
    valType := data["type"].(string)
    if valType == "" {
      return JSON(ResponseData{5000, "无法解析数据类型", nil})
    }
    switch valType {
    case "list":
      _, err = client.Do("LPUSH", key, data["data"])
    case "set":
      _, err = client.Do("SADD", key, data["data"])
    case "zset":
      _, err = client.Do("ZADD", key, data["rowkey"], data["data"])
    case "string":
      _, err = client.Do("SET", key, data["data"])
    case "hash":
      rowkey := data["rowkey"].(string)
      if rowkey == "" {
        return JSON(ResponseData{5000, "参数错误", nil})
      }
      _, err = client.Do("HSET", key, rowkey, data["data"])
    }
  case "addrow": // 添加新的列
    valType := data["type"].(string)
    if valType == "" {
      return JSON(ResponseData{5000, "无法解析数据类型", nil})
    }
    switch valType {
    case "list":
      _, err = client.Do("RPUSH", key, data["data"])
    case "set":
      _, err = client.Do("SADD", key, data["data"])
    case "zset":
      score := int(data["rowkey"].(float64))
      _, err = client.Do("ZADD", key, score, data["data"])
    case "hash":
      rowkey := data["rowkey"].(string)
      if rowkey == "" {
        return JSON(ResponseData{5000, "参数错误", nil})
      }
      _, err = client.Do("HSET", key, rowkey, data["data"])
    }
  case "updateRowValue":
    valType := data["type"].(string)
    if valType == "" {
      return JSON(ResponseData{5000, "无法解析数据类型", nil})
    }

    switch valType {
    case "list":
      rowkey, _ := strconv.Atoi(data["rowkey"].(string))
      _, err = client.Do("LSET", key, rowkey, data["data"])
    case "set":
      rowkey := data["rowkey"].(string)
      _, err = client.Do("SREM", key, rowkey)
      _, err = client.Do("SADD", key, data["data"])
    case "zset":
      score := int(data["score"].(float64))
      rowkey := data["rowkey"].(string)
      //fmt.Println("rowkey", rowkey)
      _, err = client.Do("ZREM", rowkey)
      _, err = client.Do("ZADD", rowkey, score, data["data"])
    case "hash":
      hashKey := data["rowkey"].(string)
      _, err = client.Do("HSET", key, hashKey, data["data"])
    }
  default:
    return JSON(ResponseData{5000, "无法解析动作", nil})
  }
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  return JSON(ResponseData{200, "操作成功", nil})
}

func RedisManagerAddKey(data map[string]interface{}) string {
  client, key, err := getRedisClient(data, true, true)
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }

  valType := data["type"].(string)
  if valType == "" {
    return JSON(ResponseData{5000, "无法解析数据类型", nil})
  }
  switch valType {
  case "list":
    _, err = client.Do("LPUSH", key, data["data"].(string))
  case "set":
    _, err = client.Do("SADD", key, data["data"].(string))
  case "zset":
    score := int(data["rowKey"].(float64))
    _, err = client.Do("ZADD", key, score, data["data"].(string))
  case "string":
    _, err = client.Do("SET", key, data["data"].(string))
  case "hash":
    rowkey := data["rowKey"].(string)
    if rowkey == "" {
      return JSON(ResponseData{5000, "参数错误", nil})
    }
    _, err = client.Do("HSET", key, rowkey, data["data"].(string))
  }
  if err != nil {
    return JSON(ResponseData{5000, err.Error(), nil})
  }
  return JSON(ResponseData{200, "操作成功", nil})
}

func JSON(data ResponseData) string {
  b, _ := json.Marshal(data)
  return string(b)
}

func readConfigJSON() error {
  f, err := os.OpenFile(jsonFile, os.O_RDWR, os.ModePerm)
  if err != nil && os.IsNotExist(err) {
    return nil
  }
  defer f.Close()
  s, err := ioutil.ReadAll(f)
  if err != nil {
    return err
  }
  err = json.Unmarshal(s, &connectionList)
  if err != nil {
    return err
  }
  if len(connectionList) > 0 {
    last := connectionList[len(connectionList)-1]
    totalConnection = int(last.ID) + 1
  }
  return nil
}

func writeConfigJSON() error {
  //os.O_TRUNC 清空内容
  f, err := os.OpenFile(jsonFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
  if err != nil {
    return err
  }
  defer f.Close()
  str, err := json.Marshal(&connectionList)
  if err != nil {
    return err
  }
  _, err = io.WriteString(f, string(str))
  if err != nil {
    return err
  }
  totalConnection = len(connectionList)
  return nil
}
