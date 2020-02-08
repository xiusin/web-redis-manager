# redis_manager #
redis管理客户端,  基于[go-astilectron](https://github.com/asticode/go-astilectron.git) 开发构建, 代码比较简陋.

## BUG ##
1. http模式pubsub接口兼容, 调整订阅与发布顺序, 消除第一次发送消息失败的情况
2. 选中tree节点时无法高亮
3. 刷新value概率锁定按钮loading

## 伪redis-cli功能 ##
可以使用`help`查看可用的命令. 部分还原redis-cli的响应值.

## TODO ##
- [ ] ~~键搜索, tree组件无法提供, 暂时不添加功能~~
- [ ] 集群模式
- [ ] Electron Cpu占用率太高, windows 不关闭Electron进程
- [ ] redis-cli功能新增切换连接情况功能.(保持select 数据库的状态)
- [ ] 替换类库为(https://github.com/zserge/lorca), 精简文件大小与依赖
- [ ] 切换db清空cli的输入内容(暴露全局命令行对象)
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
