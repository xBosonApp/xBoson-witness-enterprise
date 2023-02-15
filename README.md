# xBoson 见证者节点 - 企业版

连接到 xBoson 运算核心见证并保存数字签名.

该项目由 [上海竹呗信息技术有限公司](https://xboson.net/) 提供技术支持.

## 构建

go1.20.1 或更高版本编译

```sh
# 可选: 编译静态文件为程序资源, 发布时不依赖 www 目录;
# 开发时删除 `web/resource_www.go` 文件;
# nodejs > version 6
node web/build

go build -o build
```

## 启动:

`witness [-c config_file]`

参数:

`-c` 指定启动配置文件, 默认 `witness-config.json`

## 说明:

* 配置文件中含有私钥, 请妥善保管, 一旦遗失或泄漏可能引起不必要的经济损失.
* 首次启动, 指定的配置文件如果不存在, 则会创建新的配置文件, 包括公钥和私钥, 数据库目录.
* 首次启动时会询问本机有效 ip, http 本机监听端口, 连接平台的主机地址和端口.
* 若本机首选 ip 地址有变动, 将被程序检测到, 启动时需要再次选择本机 ip.
* 本程序按照 `区块链见证者节点接入方法` 开发, 如有变动不另行通知.


## 其他

* [xBoson平台运算核心](https://github.com/yanmingsohu/xBoson-core)