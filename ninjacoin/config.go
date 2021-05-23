package ninjacoin

import (
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/common/file"
	"path/filepath"
	"strings"
	"time"
)

const (
	//币种
	Symbol    = "NINJA"
	Decimal   = 5
	CurveType = owcrypt.ECC_CURVE_ED25519

	//交易单发送超时时限
	DefaultTxSendingTimeout = 5 * time.Minute
)

const (
	//交易单状态
	//Pending (0)     - initial state, a transaction is created, but not sent nowhere
	//InProgress (1)  - "Waiting for Sender/Waiting for Receiver" - to indicate that sender or receiver should come online to initiate the transaction
	//Canceled (2)    - "Cancelled" (by Sender, due to Rollback)
	//Completed (3)   - a transaction is completed
	//Failed (4)      - failed for some reason
	//Registering (5) - a transaction is taken care by the blockchain, some miner needs to PoW and to add it to a block, the block should be added to the blockchain

	TxStatusPending     = 0
	TxStatusInProgress  = 1
	TxStatusCanceled    = 2
	TxStatusCompleted   = 3
	TxStatusFailed      = 4
	TxStatusRegistering = 5
)

type WalletConfig struct {

	//币种
	Symbol string
	//配置文件路径
	configFilePath string
	//配置文件名
	configFileName string
	//区块链数据文件
	BlockchainFile string
	//本地数据库文件路径
	dbPath string
	//默认配置内容
	DefaultConfig string
	//曲线类型
	CurveType uint32
	//固定手续费
	fixfees string
	// 远程服务
	remoteserver string
	//是否开启协商密码通信
	enablekeyagreement bool
	//是否支持ssl：https，wss等
	enablessl bool
	//网络请求超时，单位：秒
	requesttimeout int
	//钱包API
	walletapi string
	//浏览器API
	blockNodeApi string
	//连接方式
	connecttype string
	//信任节点
	trustnodeid string
	//是否作为服务端
	enableserver bool
	//是否输出LogDebugg日志
	logdebug bool
	//通信证书私钥
	cert string
	//汇总地址
	summaryaddress string
	//汇总阈值
	summarythreshold string
	//汇总时间周期
	summaryperiod string
	//日志路径
	logdir string
	//交易单发送超时
	txsendingtimeout time.Duration
	//钱包wallet.db备份目录
	walletdatabackupdir string
	//钱包wallet.db绝对路径
	walletdatafile string
}

func NewConfig(symbol string) *WalletConfig {

	c := WalletConfig{}

	//币种
	c.Symbol = symbol
	c.CurveType = CurveType

	//区块链数据
	//blockchainDir = filepath.Join("data", strings.ToLower(Symbol), "blockchain")
	//配置文件路径
	c.configFilePath = filepath.Join("conf")
	//配置文件名
	c.configFileName = c.Symbol + ".ini"
	//区块链数据文件
	c.BlockchainFile = "blockchain.db"
	//本地数据库文件路径
	c.dbPath = filepath.Join("data", strings.ToLower(c.Symbol), "db")

	//创建目录
	file.MkdirAll(c.dbPath)

	return &c
}
