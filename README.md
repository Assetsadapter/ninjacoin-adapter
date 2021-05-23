# beam-adapter

beam暂时无法适配openwallet钱包体系，无法实现离线生产地址，离线交易签名。
因为beam的Mimblewimble协议需要钱包一直保持在线，才能正常地接收和发送资产。
为了实现企业级别钱包解决方案，我们划分两个应用场景。
- 用户托管钱包场景。就是企业托管了注册用户的钱包，为每一个用户分配接收地址，这个钱包不对外开放。在一台装有beam钱包的服务器上，运行openw-beam钱包服务，监听用户的充值，并定时进行汇总。
- 财务系统提币钱包场景。企业不会直接从用户托管钱包中做转账，而是在一个独立的热钱包中驻留日常业务的需要的资产。财务系统集成beam-adapter，可发送指令向openw-beam钱包服务创建新地址，可查询用户托管钱包的新充值记录。

## 官方资料

### 官网

https://www.beam.mw/

### 接口文档

#### Wallet API

https://github.com/BeamMW/beam/wiki/Beam-wallet-protocol-API

#### Explorer API

https://github.com/BeamMW/beam/wiki/Beam-Node-Explorer-API

### 浏览器

https://explorer.beam.mw/

### 测试币领取

https://bitmate.ch/

## openw-beam使用说明

openw-beam是一个运行在beam用户托管钱包服务器的后台程序，为财务系统的热钱包提供创建用户地址，获取交易记录，获取钱包余额，执行定时汇总等功能。

### 编译程序

#### 安装xgo（支持跨平台编译C代码）

```shell

# 依赖工具
$ make clean deps

# 本地系统编译
$ make clean build

# 跨平台编译wmd，更多可选平台可修改Makefile的$TARGETS变量
$ make clean openw-beam


```

### 服务端配置文件

beam钱包安装成功后通过下面命令检查rpc接口是否正常

```shell

# check wallet api
curl -d '{"jsonrpc":"2.0","id":1,"method":"wallet_status"}' -H "Content-Type: application/json" -X POST http://127.0.0.1:20021/api/wallet

# check explorer api
curl http://127.0.0.1:20022/status

```



在用户托管钱包的服务器配置，运行openw-beam后台服务，开放接口给财务系统的热钱包使用，样例如下：

```ini

# Beam Wallet RPC API, beam钱包API
walletapi = "http://192.168.1.123:12345/api/wallet"

# Beam explore API, beam钱包浏览器API
explorerapi = "http://192.168.1.123:12346"

# beam-adapter Remote Server, openw-beam服务的固定IP或域名
remoteserver = ":20888"

# True: Run for server, False: Run for client, 作为服务端启动
enableserver = true

# Fix Transaction Fess, 最低手续费
fixfees = "0.000001"

# Node Connect Type, 连接方式：ws: websocket
connecttype = "ws"

# Enable key agreement on local node communicate with client server, 开启协商密码
enablekeyagreement = true

# log debug info, 是否打印debug日志
logdebug = false

# Log file path, 日志目录
logdir = "./logs/"

# trust node id, 服务端让授信的客户端连接
trustnodeid = "11111"

# summary address 汇总地址
summaryaddress = "111111"

# summary threshold 汇总阈值
summarythreshold = "0.001"

# Wallet Summary Period,  汇总周期
summaryperiod = "30s"

# Transaction sending timeout, 如果接受方钱包不在线，交易会一直处于发送中状态，需要设置一个超时时间，超时取消发送中的交易
# Such as "30s", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
txsendingtimeout = "5m"

# Backup wallet.db directory, 备份wallet data文件，每完成一次汇总，都会备份wallet.db到这个目录
walletdatabackupdir = "./backup/"

# beam wallet.db Absolute Path, beam wallet.db文件绝对路径
walletdatafile = "/data/beam/openw-beam/wallet.db"
```

在用户托管钱包的服务器运行beam-walle

```shell

# 加载配置server.ini，运行walletserver后台服务
$ ./openw-beam -c=server.ini walletserver

```

### 客户端配置文件

在财务系统的钱包服务器配置，财务系统集成beam-adapter，通过AssetsAdapter接口加载如下配置：

```ini

# Beam Wallet RPC API, beam钱包API
walletapi = "http://192.168.1.123:12345/api/wallet"

# Beam explore API, beam钱包浏览器API
explorerapi = "http://192.168.1.123:12346"

# beam-adapter Remote Server, openw-beam服务的固定IP或域名
remoteserver = "127.0.0.1:20888"

# True: Run for server, False: Run for client, 作为客户端启动
enableserver = false

# Fix Transaction Fess, 最低手续费
fixfees = "0.000001"

# Node Connect Type, 连接方式：ws: websocket
connecttype = "ws"

# Enable key agreement on local node communicate with client server, 开启协商密码
enablekeyagreement = true

# Enable https or wss, 如果服务器有SSL证书，可开启SSL，
enablessl = false

# Network request timeout, unit: second, 请求连接超时限制
requesttimeout = 120

# log debug info, 是否打印debug日志
logdebug = false

# Log file path, 日志目录
logdir = "./logs/"

# Generate Node, 客户端证书私钥
cert = "1111"

# Transaction sending timeout, 如果接受方钱包不在线，交易会一直处于发送中状态，需要设置一个超时时间，超时取消发送中的交易
# Such as "30s", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
txsendingtimeout = "5m"


```

系统集成beam-adapter/beam包功能

```go

    //创建beam钱包管理对象
	clientNode := beam.NewWalletManager()
	
	//加载client.ini配置文件
	c, err := config.NewConfig("ini", "client.ini")
	if err != nil {
		return nil
	}
	clientNode.LoadAssetsConfig(c)
	
	//向远程服务，创建用户托管钱包的地址
	addrs, err := clientNode.CreateRemoteWalletAddress(100, 10)
	if err != nil {
        return
	}
	
	//获取本地钱包（热钱包）余额
	balanceLocal, err := clientNode.GetLocalWalletBalance()

    //获取用户充值钱包余额
    balanceRemote, err := clientNode.GetRemoteWalletBalance()
    	
	//发起转账交易
    rawTx := &openwallet.RawTransaction{
        To: map[string]string{
            "3b769e29f6e2fc59fb7d1cd88fa03bd0777318b83d0e5111941992ad5efbe670d31": "0.0000001",
        },
        FeeRate: "",
    }

    txdecoder := clientNode.TxDecoder
    tx, err := txdecoder.SubmitRawTransaction(nil, rawTx)
    
    //启动区块链扫描器
    scanner := clientNode.GetBlockScanner()
	scanner.Run()
	
```

### 注意事项

`钱包数据备份`

由于beam无法适配openwallet钱包体系，所以地址私钥等都托管在beam钱包上。
钱包管理员在安装beam钱包后，需要备份好助记词和密码，定时备份wallet.db。

`绑定信任节点进行通信`

为了满足用户充值钱包与提现热钱包的安全通信。OWTP可绑定固定的节点进行通信。
客户端配置文件中的`cert`字段，可通过`openw-cli`的`genkeychain`命令生成通信私钥，
把`PRIVATE KEY`填到`cert`字段。把`NODE ID`填到服务端配置文件的`trustnodeid`字段。
