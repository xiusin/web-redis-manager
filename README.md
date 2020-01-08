# redis_manager #

自用redis管理客户端,  基于[go-astilectron](https://github.com/asticode/go-astilectron.git) 开发构建, 代码比较简陋. 没打算重构

# 部分截图 #
![](images/image1.png)

![](images/image2.png)

![](images/image3.png)

![](images/image4.png)


## BUG ##
1. set的排序问题
2. zset读取排序问题(由于返回的是map所以乱序)
3. 字段视图转JSON时候数据错误
4. 选中key无法高亮


## TODO ##
- [ ] 键搜索
- [ ] 集群模式
- [ ] 配置信息以及服务器信息打印
