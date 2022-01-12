# redis_manager #

redis的web管理客户, 支持同时管理多个redis实例, 慢日志, 服务器信息, 配置信息, CLI模式.

> 演示案例: https://rdm.xiusin.cn
# 特性 #

- basicauth
- 支持redis常用数据类型管理: `list`, `string`, `hashmap`, `set`, `sorted set`等.  
- CLI提醒模式
- 慢日志查询打印
- `channel` 订阅发布

# 示意图 #

## 入口页面 ##

![./images/1-min.png](./images/1-min.png)

## 连接实例 ##

![./images/2-min.png](./images/2-min.png)

## 管理键 ##

![./images/3-min.png](./images/3-min.png)

## 操作值 ##

![./images/4-min.png](./images/4-min.png)

## 配置信息 ##

![./images/5-min.png](./images/5-min.png)

## 服务器信息 ##

![./images/6-min.png](./images/6-min.png)

## 慢日志 ##

![./images/7-min.png](./images/7-min.png)

## CLI管理 ##

![./images/8-min.png](./images/8-min.png)
