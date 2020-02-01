package src

import (
  "encoding/gob"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/asticode/go-astilectron"
  "github.com/asticode/go-astilog"
  "os"
  "reflect"
  "regexp"
  "strconv"
  "strings"
  "time"

  "github.com/gomodule/redigo/redis"
)

var Window *astilectron.Window
var pubSubs = map[string]bool{}

type slowLog struct {
  UsedTime string  `json:"used_time"`
  Command  string `json:"command"`
  Time     string `json:"time"`
}

var loc, _ = time.LoadLocation("PRC")

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

  logs, err := redis.Values(client.Do("SLOWLOG", "GET"))
  if err != nil {
    return JSON(ResponseData{5000, "读取慢日志失败:" + err.Error(), nil})
  }
  //1：每个慢查询条目的唯一的递增标识符。
  //2：处理记录命令的unix时间戳。
  //3：命令执行所需的总时间，以微秒为单位。
  //4：组成该命令的参数的数组。
  var structLogs []slowLog
  for _, log := range logs {
    var sl slowLog
    for k, val := range log.([]interface{}) {
      if k == 1 {
        sl.Time = time.Unix(val.(int64), 0).In(loc).Format("2006-01-02 15:04:05")
      } else if k == 2 {
        sl.UsedTime = strconv.Itoa(int(val.(int64))) + "μs"
      } else if k == 3 {
        sl.Command = strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%s", val), "["), "]")
      }
    }
    structLogs = append(structLogs, sl)
  }

  return JSON(ResponseData{200, "保存成功", map[string]interface{}{
    "data":     d,
    "config":   c,
    "slowLogs": structLogs,
  }})
}

func RedisManagerConnectionTest(data map[string]interface{}) string {
  var config connection
  config.Ip = data["ip"].(string)
  config.Title = data["title"].(string)
  config.Port = int(data["port"].(float64))
  config.Auth = data["auth"].(string)

  client, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", config.Ip, strconv.Itoa(config.Port)))
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
  var config connection
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

func RedisManagerConnectionList(_ map[string]interface{}) string {
  err := readConfigJSON()
  if err != nil {
    return JSON(ResponseData{5000, "获取列表失败:" + err.Error(), nil})
  }
  return JSON(ResponseData{200, "获取列表成功", connectionList})
}

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
        fmt.Println("subscribe channel", channel)
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

    if ok, _ := pubSubs[fmt.Sprintf("channel-%d", id)]; !ok {
      pubSubs[fmt.Sprintf("channel-%d", id)] = true
      go func() {
        defer func(id int) {
          pubSubs[fmt.Sprintf("channel-%d", id)] = false
        }(id)
        pubsub := redis.PubSubConn{Conn: client}
        if err := pubsub.PSubscribe("*"); err != nil {
          panic(err)
        }
        for {
          message := pubsub.Receive()
          switch v := message.(type) {
          case redis.Message: //单个订阅subscribe
            _ = Window.SendMessage(map[string]string{
              "data":    string(v.Data),
              "id":      strconv.Itoa(id),
              "channel": v.Channel,
              "time":    time.Now().In(loc).Format("15:04:05"),
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
    return JSON(ResponseData{5000, "获取数据错误", fmt.Sprintf("(error) %s", err)})
  }
  if val == nil {
    return JSON(ResponseData{200, "成功", `(nil)`})
  }
  fmt.Println(reflect.TypeOf(val).String())
  switch val.(type) {
  case []byte, string:
    res, _ := redis.String(val, nil)
    return JSON(ResponseData{200, "成功", `"` + res + `"`})

  case int64, int, int32:
    res, _ := redis.Int64(val, nil)
    return JSON(ResponseData{200, "成功", fmt.Sprintf("(integer) %d", res)})

  case []interface{}:
    res, err := redis.StringMap(val, nil)
    if err != nil {
      return JSON(ResponseData{5000, "成功", err.Error()})
    }
    var strs []string
    var i int
    switch strings.ToUpper(commands[0]) {
    case "HGETALL", "SMEMBERS", "BLPOP", "LRANGE":
      for k, v := range res {
        i++
        if ok, _ := regexp.MatchString("^d+$", k); ok {
          strs = append(strs, strconv.Itoa(i)+") "+k)
        } else {
          strs = append(strs, strconv.Itoa(i)+`) "`+k+`"`)
        }
        i++
        strs = append(strs, strconv.Itoa(i)+") \""+v+"\"")
      }

    default:
      strs = append(strs, fmt.Sprintf("%#v", res))
    }
    return JSON(ResponseData{200, "成功", strings.Join(strs, "<br/>")})

  default:
    return JSON(ResponseData{200, "成功", val})
  }
}

func RedisManagerRemoveConnection(data map[string]interface{}) string {
  var configs []connection
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
      var ok int
      ok, err = redis.Int(client.Do("SADD", key, data["data"]))
      if ok == 0 {
        return JSON(ResponseData{5000, "添加失败", nil})
      }
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
      rowkey := int(data["rowkey"].(float64))
      _, err = client.Do("LSET", key, rowkey, data["data"])
    case "set":
      rowkey := data["rowkey"].(string)
      _, err = client.Do("SREM", key, rowkey)
      _, err = client.Do("SADD", key, data["data"])
    case "zset":
      score := int(data["score"].(float64))
      rowkey := data["rowkey"].(string)
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
  f, err := os.OpenFile(ConnectionFile, os.O_RDWR, os.ModePerm)
  if err != nil && os.IsNotExist(err) {
    return nil
  }
  defer f.Close()
  decoder := gob.NewDecoder(f)
  err = decoder.Decode(&connectionList)
  if err != nil {
    return err
  }
  if len(connectionList) > 0 {
    last := connectionList[len(connectionList)-1]
    totalConnection = int(last.ID) + 1
  }
  return nil
}

func RedisManagerGetCommandList(_ map[string]interface{}) string {
  commandList := `EXISTS:::测试给定的 key 是否存在
DEL:::删除给定的 key
EXPIRE:::设置键超时时间
GET:::获取给定的key
GETSET:::设置指定 key 的值，并返回 key 的旧值。
HDEL:::删除哈希表 key 中的一个或多个指定字段，不存在的字段将被忽略
HEXISTS:::查看哈希表的指定字段是否存在。
HGET:::返回哈希表中指定字段的值。
HKEYS:::获取哈希表中的所有域（field）。
HVALS:::返回哈希表所有域(field)的值。
HGETALL:::返回哈希表中，所有的字段和值。在返回值里，紧跟每个字段名(field name)之后是字段的值(value)，所以返回值的长度是哈希表大小的两倍。
HLEN:::获取哈希表中字段的数量。
HMGET:::哈希表中，一个或多个给定字段的值。如果指定的字段不存在于哈希表，那么返回一个 nil 值。
HMSET:::同时将多个 field-value (字段-值)对设置到哈希表中。如果哈希表不存在，会创建一个空哈希表，并执行 HMSET 操作。此命令会覆盖哈希表中已存在的字段。
HSET:::为哈希表中的字段赋值 。如果哈希表不存在，一个新的哈希表被创建并进行 HSET 操作。如果字段已经存在于哈希表中，旧值将被覆盖。
INCR:::将 key 中储存的数字值增一。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCR 操作。
INCRBY:::将 key 中储存的数字加上指定的增量值。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCRBY 命令。
DECR:::将 key 中储存的数字值减一。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 DECR 操作。
DECRBY:::将 key 所储存的值减去指定的减量值。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 DECRBY 操作。
KEYS:::查找所有符合给定模式 pattern 的 key 。。
LINDEX:::通过索引获取列表中的元素。你也可以使用负数下标，以 -1 表示列表的最后一个元素， -2 表示列表的倒数第二个元素，以此类推
LLEN:::返回列表的长度。 如果列表 key 不存在，则 key 被解释为一个空列表，返回 0 。 如果 key 不是列表类型，返回一个错误。
LPOP:::移除并返回列表的第一个元素。
LPUSH:::Redis Lpush 命令将一个或多个值插入到列表头部。 如果 key 不存在，一个空列表会被创建并执行 LPUSH 操作。
LRANGE:::返回列表中指定区间内的元素，区间以偏移量 START 和 END 指定。其中 0 表示列表的第一个元素， 1 表示列表的第二个元素
LREM:::根据参数 COUNT 的值，移除列表中与参数 VALUE 相等的元素。
LSET:::通过索引来设置元素的值。
BLPOP:::命令移出并获取列表的第一个元素， 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
LTRIM:::对一个列表进行修剪(trim)，就是说，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除。
MGET:::返回所有(一个或多个)给定 key 的值。 如果给定的 key 里面，有某个 key 不存在，那么这个 key 返回特殊值 nil 。
MSET:::同时设置一个或多个 key-value 对。
MSETNX:::所有给定 key 都不存在时，同时设置一个或多个 key-value 对。
PEXPIRE:::和 EXPIRE 命令的作用类似，但是它以毫秒为单位设置 key 的生存时间，而不像 EXPIRE 命令那样，以秒为单位。
RENAME:::用于修改 key 的名称 。
RENAMENX:::在新的 key 不存在时修改 key 的名称 。
RPOP:::移除列表的最后一个元素，返回值为移除的元素。
RPOPLPUSH:::移除列表的最后一个元素，并将该元素添加到另一个列表并返回。
RPUSH:::将一个或多个值插入到列表的尾部(最右边)。
SADD:::将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。
SCARD:::返回集合中元素的数量。
SDIFF:::返回给定集合之间的差集。不存在的集合 key 将视为空集。
SDIFFSTORE:::将给定集合之间的差集存储在指定的集合中。如果指定的集合 key 已存在，则会被覆盖。
SET:::用于设置给定 key 的值。如果 key 已经存储其他值， SET 就覆写旧值，且无视类型。
SETEX:::为指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值。
SINTER:::返回给定所有给定集合的交集。 不存在的集合 key 被视为空集。 当给定集合当中有一个空集时，结果也为空集(根据集合运算定律)。
SINTERSTORE:::将给定集合之间的交集存储在指定的集合中。如果指定的集合已经存在，则将其覆盖。
SISMEMBER:::判断成员元素是否是集合的成员。
SMEMBERS:::返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
SMOVE:::将指定成员 member 元素从 source 集合移动到 destination 集合。
SORT:::返回或保存给定列表、集合、有序集合 key 中经过排序的元素。
SPOP:::移除集合中的指定 key 的一个或多个随机元素，移除后会返回移除的元素。
SRANDMEMBER:::返回集合中的一个随机元素。
SREM:::移除集合中的一个或多个成员元素，不存在的成员元素会被忽略。
SUNION:::返回给定集合的并集。不存在的集合 key 被视为空集。
SUNIONSTORE:::将给定集合的并集存储在指定的集合 destination 中。如果 destination 已经存在，则将其覆盖。
TTL:::以秒为单位返回 key 的剩余过期时间。
TYPE:::返回 key 所储存的值的类型。
ZADD:::将一个或多个成员元素及其分数值加入到有序集当中。
ZCARD:::计算集合中元素的数量。
ZCOUNT:::计算有序集合中指定分数区间的成员数量。
ZINCRBY:::对有序集合中指定成员的分数加上增量 increment 可以通过传递一个负数值 increment ，让分数减去相应的值
ZRANGE:::返回有序集中，指定区间内的成员。
ZRANGEBYSCORE:::返回有序集合中指定分数区间的成员列表。
ZRANK:::返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。
ZREM:::移除有序集中的一个或多个成员，不存在的成员将被忽略。
ZREMRANGEBYSCORE:::移除有序集中，指定分数（score）区间内的所有成员。
ZREVRANGE:::返回有序集中，指定区间内的成员。
ZSCORE:::返回有序集中，成员的分数值。 如果成员元素不是有序集 key 的成员，或 key 不存在，返回 nil 。`

  return JSON(ResponseData{200, "成功", commandList})
}

func writeConfigJSON() error {
  //os.O_TRUNC 清空内容
  f, err := os.OpenFile(ConnectionFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
  if err != nil {
    return err
  }
  defer f.Close()
  encoder := gob.NewEncoder(f)
  err = encoder.Encode(&connectionList)
  if err != nil {
    return err
  }
  totalConnection = len(connectionList)
  return nil
}
