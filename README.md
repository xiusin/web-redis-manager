# redis_manager #
redis管理客户端,  基于GO语言+Electron开发, 支持常见的数据结构的CRUD, TTL管理, 服务器信息显示, 慢日志查询.
还原`redis-cli`客户端(不依赖本地安装redis服务), 并且还原大部分响应结果,复杂的命令建议从cli, 支持发布订阅模式.

## TODO ##
- [ ] ~~键搜索, tree组件无法提供, 暂时不添加功能~~
- [ ] 替换类库为(https://github.com/zserge/lorca), 精简文件大小与依赖
- [ ] 切换连接清空cli的输入内容(暴露全局命令行对象)
