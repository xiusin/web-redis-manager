# redis_manager #
redis管理客户端,  基于[go-astilectron](https://github.com/asticode/go-astilectron.git) 开发构建, 代码比较简陋.

## BUG ##
1. http模式pubsub接口兼容, 调整订阅与发布顺序, 消除第一次发送消息失败的情况
2. macos下打开发布订阅模式软件崩溃
3. 字段视图转JSON时候数据错误
4. 选中key无法高亮

## 伪redis-cli功能 ##
可以使用`help`查看可用的命令. 部分还原redis-cli的响应值.

## TODO ##
- [ ] ~~键搜索, tree组件无法提供, 暂时不添加功能~~
- [ ] 集群模式
- [ ] Electron Cpu占用率太高, windows 不关闭Electron进程
- [ ] redis-cli功能新增切换连接情况功能.(保持select 数据库的状态)
- [ ] 替换类库为(https://github.com/zserge/lorca), 精简文件大小与依赖

## 原理图 ##
```
+-------------------------+
|      js/tcp/iview       |
+--------^---+------------+
         |   |
+--------+---v------------+
|  electron/astilectron   |
+-------+-----^-----------+
        |     |
+-------v-----+-----------+
|     golang/redigo       |
+-----+----------^--------+
      |          |
+-----v----------+-------+
|     redis/cluster      |
+------------------------+
```
