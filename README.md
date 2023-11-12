# Web Redis Manager
A modern Redis management tool that works on the web or desktop.
Web Redis Manager is a comprehensive web-based management client for Redis databases. It provides robust features and tools to simplify the process of managing and maintaining your Redis instances.

<p align="center">
  <img src="./frontend/static/redis.svg" />
</p>

####  Manage Multiple Redis Instances

With Web Redis Manager, you can manage multiple Redis instances simultaneously. This makes it easy to monitor, maintain, and manage all your Redis instances from one central location.

#### Slow Log Monitoring
Our tool provides slow log monitoring to help you identify and address performance issues. By monitoring your slow logs, you can identify queries that are taking a long time to execute and take steps to optimize them.

#### Server Information
Web Redis Manager provides detailed server information at your fingertips. You can easily view and analyze key server metrics to ensure your Redis instances are running optimally.

#### Configuration Management
Our tool makes it easy to view and modify your Redis configuration settings. Whether you need to adjust memory usage, set replication settings, or tweak other configuration options, you can do it all from our user-friendly interface.

#### CLI Mode
For power users who prefer working from the command line, we offer a CLI mode. This gives you the flexibility to manage your Redis instances in a way that suits your workflow.

### Publish/Subscribe Mode

Web Redis Manager supports the publish/subscribe messaging paradigm, allowing for real-time message communication between publishers and subscribers.

### Performance Chart Monitoring

Our tool offers performance chart monitoring, providing you with visual insights into your Redis instances' performance. This feature makes it easier to track and optimize the performance of your Redis databases.

### Modern Design and Interface

Our clean, user-friendly interface makes it easy to manage your Redis databases. You'll have all the tools you need at your fingertips.

### Web and Desktop Availability

Whether you prefer working in a web interface or a standalone desktop application, we've got you covered. Our tool works seamlessly on both platforms.

### Comprehensive Redis Management Capabilities

- **Data Visualization**: Easily view and navigate your data in a visual format.
- **Data Editing**: Modify your data directly within the tool.
- **Performance Monitoring**: Keep track of your Redis database's performance to identify and address issues promptly.
- **Security Features**: We offer features like secure password protection to help keep your data safe.

## Installation

```shell
git clone --depth=1 https://github.com/xiusin/web-redis-manager.git
cd web-redis-manager
git checkout develop

yarn # install
yarn build # build

cd server

go mod tidy # sync deps

go build -o rdm.exe # compile windows
go build -o rdm # *nix

# setup
./rdm.exe

# basic auth setup (For password authorization login on the web)
./rdm.exe --username=admin --password=123456
```

## Screenshots ##

### setup ###

![./images/1-min.png](./images/1-min.png)

### connection ###

![./images/2-min.png](./images/2-min.png)

### key list ###

![./images/3-min.png](./images/3-min.png)

### value ###

![./images/4-min.png](./images/4-min.png)

### configure ###

![./images/5-min.png](./images/5-min.png)

### server info ###

![./images/6-min.png](./images/6-min.png)

### slow log ###

![./images/7-min.png](./images/7-min.png)

### cli mode ###

![./images/8-min.png](./images/8-min.png)
