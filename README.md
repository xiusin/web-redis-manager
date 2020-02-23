# redis_manager #
redis管理客户端,  基于GO语言+Electron开发, 支持常见的数据结构的CRUD, TTL管理, 服务器信息显示, 慢日志查询.
还原`redis-cli`客户端(不依赖本地安装redis服务), 并且还原大部分响应结果,复杂的命令建议从cli, 支持发布订阅模式.


## BUG ##
1. ~~(暂不支持)http模式pubsub接口兼容, 调整订阅与发布顺序, 消除第一次发送消息失败的情况~~
2. 选中tree节点时无法高亮
5. 偶现Electron Cpu占用率太高, windows 不关闭Electron进程(没有办法则切换到lorca开发)

## TODO ##
- [ ] ~~键搜索, tree组件无法提供, 暂时不添加功能~~
- [ ] 集群模式
- [ ] 替换类库为(https://github.com/zserge/lorca), 精简文件大小与依赖
- [ ] 切换db清空cli的输入内容(暴露全局命令行对象)
- [ ] 使用SCAN读取key列表.
