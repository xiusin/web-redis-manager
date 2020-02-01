# redis_manager #
自用redis管理客户端,  基于[go-astilectron](https://github.com/asticode/go-astilectron.git) 开发构建, 代码比较简陋. 后期时间充足了会重构代码

## BUG ##
1. http模式pubsub接口兼容
2. client链接数过多
3. 字段视图转JSON时候数据错误
4. 选中key无法高亮

## TODO ##
- [ ] ~~键搜索, tree组件无法提供, 暂时不添加功能~~
- [ ] 集群模式
- [ ] Electron Cpu占用率太高
- [ ] windows 不关闭Electron进程
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
|      golang/redigo      |
+-----+----------^--------+
      |          |
+-----v----------+-------+
|     redis/cluster      |
+------------------------+
```
