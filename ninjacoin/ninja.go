package ninjacoin

import (
	"github.com/Assetsadapter/ninjacoin-adapter/ninjacoin/walletapi"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/common/file"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/openwallet/owtp"
	"time"
)

//CurveType 曲线类型
func (wm *WalletManager) CurveType() uint32 {
	return wm.Config.CurveType
}

//FullName 币种全名
func (wm *WalletManager) FullName() string {
	return "beam"
}

//Symbol 币种标识
func (wm *WalletManager) Symbol() string {
	return wm.Config.Symbol
}

//Decimal 小数位精度
func (wm *WalletManager) Decimal() int32 {
	return Decimal
}

//BalanceModelType 余额模型类别
func (wm *WalletManager) BalanceModelType() openwallet.BalanceModelType {
	return openwallet.BalanceModelTypeAddress
}

//GetAddressDecode 地址解析器
func (wm *WalletManager) GetAddressDecode() openwallet.AddressDecoder {
	return wm.Decoder
}

//GetTransactionDecoder 交易单解析器
func (wm *WalletManager) GetTransactionDecoder() openwallet.TransactionDecoder {
	return wm.TxDecoder
}

//GetBlockScanner 获取区块链
func (wm *WalletManager) GetBlockScanner() openwallet.BlockScanner {

	return wm.Blockscanner
}

//LoadAssetsConfig 加载外部配置
func (wm *WalletManager) LoadAssetsConfig(c config.Configer) error {

	var (
		err error
	)

	wm.Config.blockNodeApi = c.String("blockNodeApi")
	wm.Config.remoteserver = c.String("remoteserver")
	wm.Config.enableserver, _ = c.Bool("enableserver")
	wm.Config.fixfees = c.String("fixfees")
	wm.Config.connecttype = c.String("connecttype")
	wm.Config.enablekeyagreement, _ = c.Bool("enablekeyagreement")
	wm.Config.enablessl, _ = c.Bool("enablessl")
	wm.Config.requesttimeout, _ = c.Int("requesttimeout")
	wm.Config.trustnodeid = c.String("trustnodeid")
	wm.Config.cert = c.String("cert")
	wm.Config.logdebug, _ = c.Bool("logdebug")
	wm.Config.logdir = c.String("logdir")
	wm.Config.summaryaddress = c.String("summaryaddress")
	wm.Config.summarythreshold = c.String("summarythreshold")
	wm.Config.summaryperiod = c.String("summaryperiod")

	wm.Config.walletdatafile = c.String("walletdatafile")
	wm.Config.walletdatabackupdir = c.String("walletdatabackupdir")

	txsendingtimeout := c.String("txsendingtimeout")
	if len(txsendingtimeout) == 0 {
		wm.Config.txsendingtimeout = DefaultTxSendingTimeout
	} else {
		wm.Config.txsendingtimeout, err = time.ParseDuration(txsendingtimeout)
		if err != nil {
			return err
		}
	}

	if wm.Config.enableserver {
		walletIpAddr := c.String("walletIpAddr")
		walletIpPort := c.String("walletIpPort")
		walletRpcPassword := c.String("walletRpcPassword")
		walletFileName := c.String("walletFileName")
		walletFilePasswd := c.String("walletFilePasswd")
		walletFileDaemonHost := c.String("walletFileDaemonHost")
		walletFileDaemonPort, _ := c.Int("walletFileDaemonPort")
		walletObj := walletapi.Wallet{
			Filename:   walletFileName,
			Password:   walletFilePasswd,
			DaemonHost: walletFileDaemonHost,
			DaemonPort: walletFileDaemonPort,
		}
		wm.walletClient = NewWalletClient(wm.Config.blockNodeApi, walletIpAddr, walletIpPort, walletRpcPassword, walletObj, wm.Config.logdebug)
		wm.server, err = NewServer(wm)
		if err != nil {
			return err
		}
		wm.server.Listen()
	} else {
		wm.client, err = NewClient(wm)
		if err != nil {
			return err
		}
	}

	//建立日志文件夹
	file.MkdirAll(wm.Config.logdir)
	file.MkdirAll(wm.Config.walletdatabackupdir)

	logfile := ""
	if wm.Config.enableserver {
		logfile = "beam-server.log"
	} else {
		logfile = "beam-BlockClient.log"
	}

	//设置日志文件
	wm.SetupLog(wm.Config.logdir, logfile, wm.Config.logdebug)
	owtp.Debug = wm.Config.logdebug

	return nil
}

//InitAssetsConfig 初始化默认配置
func (wm *WalletManager) InitAssetsConfig() (config.Configer, error) {
	return config.NewConfigData("ini", []byte(wm.Config.DefaultConfig))
}

//GetAssetsLogger 获取资产账户日志工具
func (wm *WalletManager) GetAssetsLogger() *log.OWLogger {
	return wm.Log
}

//GetSmartContractDecoder 获取智能合约解析器
func (wm *WalletManager) GetSmartContractDecoder() openwallet.SmartContractDecoder {
	return wm.ContractDecoder
}
