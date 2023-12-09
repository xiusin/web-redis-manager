package handler

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xiusin/logger"

	"github.com/gomodule/redigo/redis"
)

var pubSubs = map[string]bool{}
var cliClients = cliConns{conns: map[string]redis.Conn{}}

type slowLog struct {
	UsedTime string `json:"used_time"`
	Command  string `json:"command"`
	Time     string `json:"time"`
}

type ConnectionListItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type cliConns struct {
	sync.Mutex
	conns map[string]redis.Conn
}

var loc, _ = time.LoadLocation("PRC")

func RedisManagerGetInfo(data RequestData) string {
	client, _ := getRedisClient(data, false, false)
	defer client.Close()
	d, err := redis.String(client.Do("INFO"))
	ThrowIf(err)

	c, err := redis.Strings(client.Do("CONFIG", "GET", "*"))
	ThrowIf(err)

	logs, err := redis.Values(client.Do("SLOWLOG", "GET", 50))
	ThrowIf(err)

	structLogs := []slowLog{}
	for _, log := range logs {
		var sl slowLog
		for k, val := range log.([]interface{}) {
			if k == 1 {
				sl.Time = time.Unix(val.(int64), 0).In(loc).Format("2006-01-02 15:04:05")
			} else if k == 2 {
				sl.UsedTime = strconv.Itoa(int(val.(int64)))
			} else if k == 3 {
				sl.Command = strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%s", val), "["), "]")
			}
		}
		structLogs = append(structLogs, sl)
	}

	return JSON(ResponseData{SuccessCode, "保存成功", RequestData{"data": d, "config": c, "slowLogs": structLogs}})
}

func RedisManagerConnectionTest(data RequestData) string {
	var config connection
	config.Ip = data["ip"].(string)
	config.Title = data["title"].(string)
	config.Port = data["port"].(string)
	config.Auth = data["auth"].(string)

	client, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", config.Ip, config.Port))
	ThrowIf(err)
	defer client.Close()
	if config.Auth != "" {
		_, err := client.Do("AUTH", config.Auth)
		ThrowIf(err)
	} else {
		_, err := client.Do("PING")
		ThrowIf(err)
	}
	return JSON(ResponseData{SuccessCode, "连接成功", nil})
}

func RedisManagerConfigSave(data RequestData) string {
	var config connection
	config.Ip = data["ip"].(string)
	config.Title = data["title"].(string)
	config.Port = data["port"].(string)
	config.Auth = data["auth"].(string)
	config.Readonly, _ = strconv.ParseBool(data["readonly"].(string))
	totalConnection = totalConnection + 1
	config.ID = int64(totalConnection)
	ThrowIf(config.Title == "", "名称不能为空")

	for _, conn := range connectionList {
		if conn.Ip == config.Ip && conn.Port == config.Port {
			return JSON(ResponseData{FailedCode, "已经存在相同的连接, 名称为: " + conn.Title, nil})
		}
	}
	connectionList = append(connectionList, config)
	ThrowIf(writeConfigJSON())
	return JSON(ResponseData{SuccessCode, "保存成功", config})
}

func RedisManagerConnectionList(_ RequestData) string {
	ThrowIf(readConfigJSON())
	var conns = []ConnectionListItem{}
	for _, conn := range connectionList {
		conns = append(conns, ConnectionListItem{ID: conn.ID, Title: conn.Title})
	}
	return JSON(ResponseData{SuccessCode, "获取列表成功", conns})
}

func getFromInterfaceOrFloat64ToInt(id interface{}) int {
	switch id := id.(type) {
	case float64:
		return int(id)
	case string:
		idInfo, _ := strconv.Atoi(id)
		return idInfo
	default:
		panic(fmt.Errorf("invalid type: %T", id))
	}
}

func RedisManagerRenameKey(data RequestData) string {
	client, key := getRedisClient(data, true, true)
	defer client.Close()

	newKey := data["newKey"].(string)
	ThrowIf(len(newKey) == 0, "new key is empty")

	resp, _ := client.Do("EXISTS", newKey)
	ThrowIf(resp.(int64) != 0, "new key already exists")

	_, err := client.Do("RENAME", key, newKey)
	ThrowIf(err)
	return JSON(ResponseData{SuccessCode, "重命名成功", nil})
}

func RedisManagerMoveKey(data RequestData) string {
	client, key := getRedisClient(data, true, true)
	defer client.Close()

	_, err := client.Do("MOVE", key, data["todb"].(string))
	ThrowIf(err)

	return JSON(ResponseData{SuccessCode, "移动成功", nil})
}

func RedisPubSub(data RequestData) string {
	wsIntf := data["ws"]
	channelPrefix := "channel-"
	var ws *websocket.Conn
	if wsIntf != nil {
		channelPrefix = "ws-channel-"
		ws = wsIntf.(*websocket.Conn)
	} else {
		ws = nil
	}
	client, _ := getRedisClient(data, false, false)
	id := getFromInterfaceOrFloat64ToInt(data["id"])
	channels, err := redis.Strings(client.Do("PUBSUB", "channels"))
	ThrowIf(err)
	ok := pubSubs[fmt.Sprintf("%s%d", channelPrefix, id)]

	// 检查订阅所有通道
	if (ws != nil) && !ok {
		pubSubs[fmt.Sprintf("%s%d", channelPrefix, id)] = true
		go func() {
			defer func(id int) {
				pubSubs[fmt.Sprintf("%s%d", channelPrefix, id)] = false
			}(id)
			pubsub := redis.PubSubConn{Conn: client}
			if err := pubsub.PSubscribe("*"); err != nil {
				logger.Warning(err)
				return
			}
			for {
				message, err := client.Receive()
				if err == nil {
					fmt.Println(string(message.([]byte)), err)
					switch v := message.(type) {
					case redis.Message: //单个订阅subscribe
						retData := map[string]string{
							"data":    string(v.Data),
							"id":      strconv.Itoa(id),
							"channel": v.Channel,
							"time":    time.Now().In(loc).Format("15:04:05"),
						}
						if ws != nil {
							resultValue, _ := json.Marshal(&retData)
							if err := ws.WriteMessage(websocket.TextMessage, resultValue); err != nil {
								logger.Warning(err)
								return
							}
						}
					case error:
						logger.Warning(v)
						return
					default:
					}
				} else {
					fmt.Println(err)
				}

			}
		}()
	}

	// 获取所有通道列表, 如果通道还没订阅那么就开启订阅协程
	if channel, ok := data["channel"]; ok {
		msg := data["msg"]
		ThrowIf(msg == "" || channel == "", "发布内容失败")
		var flag bool // 先查看是否有消费者订阅频道
		for _, ch := range channels {
			if ch == channel {
				flag = true
				break
			}
		}
		if !flag {
			go func() {
				client, _ := getRedisClient(data, false, false)
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
		ThrowIf(err)
		return JSON(ResponseData{SuccessCode, "发布内容成功", nil})
	}

	return JSON(ResponseData{SuccessCode, SuccessMsg, channels})
}

func RedisManagerCommand(data RequestData) string {
	cliClients.Lock()
	defer cliClients.Unlock()
	id := strconv.Itoa(getFromInterfaceOrFloat64ToInt(data["id"]))
	conn, ok := cliClients.conns[id]
	if !ok || conn == nil || conn.Err() != nil {
		conn, _ = getRedisClient(data, false, false)
		cliClients.conns[id] = conn
	}
	command, ok := data["command"]
	ThrowIf(!ok, "command empty!")

	var commands []interface{}

	err := json.Unmarshal([]byte(command.(string)), &commands)
	ThrowIf(err)

	var flags []interface{}
	for _, v := range commands[1:] {
		rightfulParam := strings.Replace(v.(string), "\"", "\\\"", -1)
		rightfulParam = strings.Replace(rightfulParam, "'", "\\'", -1)
		flags = append(flags, rightfulParam)
	}
	val, err := conn.Do(commands[0].(string), flags...)
	if err != nil {
		return JSON(ResponseData{SuccessCode, SuccessMsg, fmt.Sprintf("(error) %s", err)})
	}
	if val == nil {
		return JSON(ResponseData{SuccessCode, SuccessMsg, `(nil)`})
	}
	switch val := val.(type) {
	case []byte, string:
		res, _ := redis.String(val, nil)
		return JSON(ResponseData{SuccessCode, SuccessMsg, res})

	case int64, int, int32:
		res, _ := redis.Int64(val, nil)
		return JSON(ResponseData{SuccessCode, SuccessMsg, fmt.Sprintf("(integer) %d", res)})

	case []interface{}:
		var ret = ""
		parseInterfaces(val, &ret)
		return JSON(ResponseData{SuccessCode, SuccessMsg, ret})
	default:
		return JSON(ResponseData{SuccessCode, SuccessMsg, val})
	}
}

func parseInterfaces(val []interface{}, target *string) {
	var strs = []string{}
	if len(val) == 0 {
		*target = "(empty list or set)"
		return
	}
	var i int
	strList, err := redis.ByteSlices(val, nil)
	if err == nil {
		for _, v := range strList {
			fmt.Println(reflect.TypeOf(v))
			i++
			strs = append(strs, fmt.Sprintf("%d) \"%s\"", i, string(v)))
		}
		*target = strings.Join(strs, "<br/>")
		return
	}
	res, err := redis.StringMap(val, nil)
	if err == nil {
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
		*target = strings.Join(strs, "<br/>")
		return
	}
	deepLoopInterfaces(val, &strs, 0, "")
	*target = strings.Join(strs, "<br/>")
}

func deepLoopInterfaces(val []interface{}, strs *[]string, level int, prefix string) {
	if level == 10 {
		return
	}
	for k, v := range val {
		item := ""
		if k == 0 && prefix != "" {
			item += prefix
		}
		item += strings.Repeat("　", level)
		switch v := v.(type) {
		case []byte:
			*strs = append(*strs, fmt.Sprintf("%s%d)　%s", item, k+1, string(v)))
		case string:
			*strs = append(*strs, fmt.Sprintf("%s%d)　%s", item, k+1, v))
		case int, int8, int32, int64:
			*strs = append(*strs, fmt.Sprintf("%s%d)　%d", item, k+1, v))
		case []interface{}:
			exp, _ := regexp.Compile("　+")
			item = exp.ReplaceAllString(item, "　")
			deepLoopInterfaces(v, strs, level+1, fmt.Sprintf("%s%d)　", item, k+1))
		default:
			*strs = append(*strs, fmt.Sprintf("%s,%+v", reflect.TypeOf(v), v))
		}
	}
}

func RedisManagerRemoveConnection(data RequestData) string {
	var configs []connection
	id := int64(getFromInterfaceOrFloat64ToInt(data["id"]))
	if id == 0 {
		return JSON(ResponseData{SuccessCode, FailedMsg, nil})
	}
	for _, v := range connectionList {
		if v.ID != id {
			configs = append(configs, v)
		}
	}
	connectionList = configs
	ThrowIf(writeConfigJSON())
	return JSON(ResponseData{SuccessCode, SuccessMsg, nil})
}

var redisPools = map[int64]*redis.Pool{}

func GetServerCfg(data RequestData) (connection, error) {
	var config = connection{}
	if len(connectionList) == 0 {
		_ = readConfigJSON()
	}
	id := getFromInterfaceOrFloat64ToInt(data["id"])
	config.ID = int64(id)
	for _, v := range connectionList {
		if v.ID == config.ID {
			config = v
			break
		}
	}
	ThrowIf(config.Title == "", "no connection")
	return config, nil
}

func getRedisClient(data RequestData, getSelectedIndexClient bool, getKey bool) (redis.Conn, string) {
	var pool *redis.Pool
	var ok bool
	config, err := GetServerCfg(data)
	ThrowIf(err)

	if pool, ok = redisPools[config.ID]; !ok {
		pool = &redis.Pool{
			Dial: func() (conn redis.Conn, err error) {
				conn, err = redis.Dial("tcp", config.Ip+":"+config.Port)
				ThrowIf(err)
				if config.Auth != "" {
					if _, err := conn.Do("AUTH", config.Auth); err != nil {
						conn.Close()
						panic(err)
					}
				}
				conn.Do("CLIENT", "SETNAME", fmt.Sprintf("RDM:(%d):CLIENT(%d)", config.ID, rand.Intn(19999)))
				return conn, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) >= 3*time.Minute {
					c.Do("PING")
				}
				return nil
			},
			MaxIdle: 3, MaxActive: 0, IdleTimeout: time.Minute * 10, Wait: true,
		}
		redisPools[config.ID] = pool
	}
	client := pool.Get()
	var index = 0
	if getSelectedIndexClient {
		index = getFromInterfaceOrFloat64ToInt(data["index"])
	}
	_, _ = client.Do("SELECT", index)
	var key string
	if getKey {
		key = data["key"].(string)
		ThrowIf(key == "", "please select the key to operate")
	} else {
		key = ""
	}
	return client, key
}

func RedisManagerConnectionServer(data RequestData) string {
	client, _ := getRedisClient(data, false, false)
	defer client.Close()
	var err error
	action := strings.Trim(data["action"].(string), " ")
	switch action {
	case "get_value":
		index := getFromInterfaceOrFloat64ToInt(data["index"])
		_, err = client.Do("SELECT", index)
		ThrowIf(err)
		key := data["key"].(string)
		ThrowIf(key == "", FailedMsg)
		typeStr, _ := redis.String(client.Do("TYPE", key))
		if typeStr == "none" {
			return JSON(ResponseData{5001, FailedMsg, nil})
		}
		ttl, _ := redis.Int64(client.Do("TTL", key))
		size := 1000
		switch typeStr {
		case "list": // 读取总长度
			llen, _ := redis.Int64(client.Do("LLEN", key))
			val, err := redis.Strings(client.Do("LRANGE", key, 0, size))
			ThrowIf(err)
			return JSON(ResponseData{SuccessCode, "读取所有key成功", RequestData{"type": typeStr, "data": val, "ttl": ttl, "count": llen, "size": 50, "current": 1}})
		case "set":
			llen, _ := redis.Int64(client.Do("SCARD", key))
			repl, err := client.Do("SSCAN", key, 0, "COUNT", size)
			ThrowIf(err)
			keys, err := redis.Strings(repl.([]interface{})[1], nil)
			ThrowIf(err)
			return JSON(ResponseData{SuccessCode, "读取所有key成功", RequestData{"type": typeStr, "data": keys, "ttl": ttl, "count": llen, "size": 50, "current": 1}})
		case "stream":
			llen, _ := redis.Int64(client.Do("XLEN", key))
			val, err := client.Do("XRANGE", key, "-", "+", "COUNT", size)
			ThrowIf(err)
			vds := val.([]interface{})
			var retData []map[string][]string
			for _, v := range vds {
				item := map[string][]string{}
				v := v.([]interface{})
				xid := string(v[0].([]byte))
				item[xid] = []string{}
				fv := v[1].([]interface{})
				for _, v := range fv {
					item[xid] = append(item[xid], string(v.([]byte)))
				}
				retData = append(retData, item)
			}
			return JSON(ResponseData{SuccessCode, "读取所有key成功", RequestData{"type": typeStr, "data": retData, "ttl": ttl, "count": llen, "size": 50, "current": 1}})
		case "zset":
			llen, _ := redis.Int64(client.Do("ZCARD", key))
			repl, err := client.Do("ZSCAN", key, 0, "COUNT", size)
			ThrowIf(err)
			values, err := redis.Strings(repl.([]interface{})[1], nil)
			ThrowIf(err)

			var retData []map[string]string
			for i, v := range values {
				if i%2 == 1 {
					retData = append(retData, map[string]string{"value": values[i-1], "score": v})
				}
			}
			return JSON(ResponseData{SuccessCode, "读取所有key成功", RequestData{"type": typeStr, "data": retData, "ttl": ttl, "count": llen, "size": 50, "current": 1}})
		case "string":
			val, err := redis.String(client.Do("GET", key))
			ThrowIf(err)
			return JSON(ResponseData{SuccessCode, "读取所有key成功", RequestData{"type": typeStr, "data": val, "ttl": ttl}})
		case "hash":
			llen, _ := redis.Int64(client.Do("HLEN", key))
			repl, err := client.Do("HSCAN", key, 0, "COUNT", size)
			ThrowIf(err)
			keys, err := redis.StringMap(repl.([]interface{})[1], nil)
			ThrowIf(err)
			return JSON(ResponseData{SuccessCode, "读取所有key成功", RequestData{"type": typeStr, "data": keys, "ttl": ttl, "count": llen, "size": 50, "current": 1}})
		}
	case "dblist":
		var dbs []int
		for i := 0; i < 20; i++ {
			if _, err := client.Do("SELECT", i); err != nil {
				break
			}
			total, _ := redis.Int(client.Do("DBSIZE"))
			dbs = append(dbs, total)
		}
		return JSON(ResponseData{SuccessCode, "连接数据库成功", dbs})
	case "select_db":
		index := getFromInterfaceOrFloat64ToInt(data["index"])
		_, _ = client.Do("SELECT", index) //选择数据库
		var nextCur = "0"
		var resKeys = map[string][]string{}

		filter := data["filter"].(string)
		if filter == "" {
			filter = "*"
		}

		repl, err := client.Do("SCAN", nextCur, "MATCH", filter, "COUNT", 200)
		if err != nil {
			return JSON(ResponseData{FailedCode, err.Error(), nil})
		}
		nextCur = string(repl.([]interface{})[0].([]byte))
		keys, err := redis.Strings(repl.([]interface{})[1], nil)
		if err != nil {
			return JSON(ResponseData{FailedCode, "错误,无法解析SCAN返回值", nil})
		}
		for _, v := range keys {
			resKeys[v] = append(resKeys[v], v)
		}

		return JSON(ResponseData{SuccessCode, "读取所有key成功:" + string(nextCur), resKeys})
	}
	return JSON(ResponseData{FailedCode, "错误,无法解析到动作:" + action, nil})
}

func RedisManagerRemoveKey(data RequestData) string {
	client, key := getRedisClient(data, true, true)
	defer client.Close()
	_, err := client.Do("UNLINK", key) // UNLINK (异步) 替代 DEL (同步)
	ThrowIf(err)
	return JSON(ResponseData{SuccessCode, "删除成功", nil})
}

func RedisManagerFlushDB(data RequestData) string {
	client, _ := getRedisClient(data, true, false)
	defer client.Close()
	_, err := client.Do("FLUSHDB")
	ThrowIf(err)
	return JSON(ResponseData{SuccessCode, "Successfully cleared the database", nil})
}

func RedisManagerRemoveRow(data RequestData) string {
	client, key := getRedisClient(data, true, true)
	defer client.Close()
	var err error
	valType := data["type"].(string)
	ThrowIf(valType == "", "Unable to parse data type")
	switch valType {
	case "list":
		_, err = client.Do("LREM", key, 1, data["data"])
	case "set":
		_, err = client.Do("SREM", key, data["data"])
	case "zset":
		_, err = client.Do("ZREM", key, data["data"].(string))
	case "hash":
		_, err = client.Do("HDEL", key, data["data"])
	case "stream":
		_, err = client.Do("XDEL", key, data["data"])
	}
	ThrowIf(err)
	return JSON(ResponseData{SuccessCode, "Successfully deleted", nil})
}

func RedisManagerUpdateKey(data RequestData) string {
	client, key := getRedisClient(data, true, true)
	defer client.Close()
	var err error
	action := data["action"].(string)
	var extraData = map[string]interface{}{}
	switch action {
	case "ttl": //更新ttl时间
		ttl := getFromInterfaceOrFloat64ToInt(data["ttl"])
		if ttl < -1 {
			_, err = client.Do("PERSIST", key)
		} else {
			_, err = client.Do("EXPIRE", key, ttl)
		}
		ThrowIf(err)
	case "value": //更新value
		valType := data["type"].(string)
		ThrowIf(valType == "", "Unable to parse data type")
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
			ThrowIf(rowkey == "", "parameter error")
			_, err = client.Do("HSET", key, rowkey, data["data"])
		}
	case "addrow": // 添加新的列
		valType := data["type"].(string)
		ThrowIf(valType == "", "Unable to parse data type")
		switch valType {
		case "list":
			_, err = client.Do("RPUSH", key, data["data"])
		case "set":
			var ok int
			ok, err = redis.Int(client.Do("SADD", key, data["data"]))
			ThrowIf(ok == 0, "Adding failed, data already exists")
		case "stream":
			newId := data["score"].(string)
			if len(newId) == 0 {
				newId = "*"
			}
			params := []interface{}{key, newId}
			fvs := strings.Split(data["data"].(string), "\n")
			if len(fvs)%2 == 1 {
				return JSON(ResponseData{FailedCode, "stream值必须成对设置", nil})
			}
			for _, fv := range fvs {
				params = append(params, fv)
			}
			newId, err = redis.String(client.Do("XADD", params...))
			extraData["id"] = newId
		case "zset":
			score := getFromInterfaceOrFloat64ToInt(data["rowkey"])
			_, err = client.Do("ZADD", key, score, data["data"])
		case "hash":
			rowkey := data["rowkey"].(string)
			if rowkey == "" {
				return JSON(ResponseData{FailedCode, "参数错误", nil})
			}
			_, err = client.Do("HSET", key, rowkey, data["data"])
		}
	case "updateRowValue":
		valType := data["type"].(string)
		ThrowIf(valType == "", "Unable to parse data type")
		switch strings.ToLower(valType) {
		case "list":
			_, ok := data["rowkey"]
			if !ok {
				return JSON(ResponseData{FailedCode, "请选择要编辑的数据", nil})
			}
			rowkey := getFromInterfaceOrFloat64ToInt(data["rowkey"])
			_, err = client.Do("LSET", key, rowkey, data["data"])
		case "set":
			rowkey, ok := data["rowkey"].(string)
			if !ok {
				return JSON(ResponseData{FailedCode, "请选择要编辑的数据", nil})
			}
			_, _ = client.Do("SREM", key, rowkey)

			_, err = client.Do("SADD", key, data["data"])
		case "zset":
			rowkey := data["rowkey"].(string)
			score := getFromInterfaceOrFloat64ToInt(data["score"])
			_, _ = client.Do("ZREM", key, rowkey)
			_, err = client.Do("ZADD", key, score, data["data"])
		case "stream":
			return JSON(ResponseData{FailedCode, "不支持修改Steam内容", nil})
		case "hash":
			hashKey := data["rowkey"].(string)
			_, err = client.Do("HSET", key, hashKey, data["data"])
		}
	default:
		return JSON(ResponseData{FailedCode, "Unable to parse action", nil})
	}
	ThrowIf(err)
	return JSON(ResponseData{SuccessCode, "Operation successful", extraData})
}

func RedisManagerAddKey(data RequestData) string {
	client, key := getRedisClient(data, true, true)
	var err error
	defer client.Close()
	valType := data["type"].(string)
	ThrowIf(valType == "", "Unable to parse data type")
	switch valType {
	case "list":
		_, err = client.Do("LPUSH", key, data["data"].(string))
	case "set":
		_, err = client.Do("SADD", key, data["data"].(string))
	case "zset":
		score := getFromInterfaceOrFloat64ToInt(data["rowKey"])
		_, err = client.Do("ZADD", key, score, data["data"].(string))
	case "string":
		_, err = client.Do("SET", key, data["data"].(string))
	case "stream":
		newId := data["rowKey"].(string)
		if len(newId) == 0 {
			newId = "*"
		}
		params := []interface{}{key, newId}
		fvs := strings.Split(data["data"].(string), "\n")
		if len(fvs)%2 == 1 {
			return JSON(ResponseData{FailedCode, "stream值必须成对设置", nil})
		}
		for _, fv := range fvs {
			params = append(params, fv)
		}
		_, err = redis.String(client.Do("XADD", params...))
	case "hash":
		rowkey := data["rowKey"].(string)
		ThrowIf(rowkey == "", "参数错误")
		_, err = client.Do("HSET", key, rowkey, data["data"].(string))
	}
	ThrowIf(err)
	return JSON(ResponseData{SuccessCode, "操作成功", nil})
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

func RedisManagerGetCommandList(_ RequestData) string {
	return JSON(ResponseData{SuccessCode, "成功", map[string]interface{}{
		"helpers": FromRedisSourceCommandHelper(),
	}})
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
