# redis_manager #
自用redis管理客户端,  基于[go-astilectron](https://github.com/asticode/go-astilectron.git) 开发构建, 代码比较简陋. 后期时间充足了会重构代码

# 使用yarn #
提醒 `@vue/app`无法找到预设, 需要到对应模块下的package.json删除babel下的预设

# 部分截图 #
![](images/image1.png)
![](images/image2.png)
![](images/image3.png)
![](images/image4.png)

## BUG ##
3. 字段视图转JSON时候数据错误
4. 选中key无法高亮

## TODO ##
- [ ] ~~键搜索, tree组件无法提供, 暂时不添加功能~~
- [ ] 集群模式
- [ ] cli模式异步返回值问题
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
