package openwtester

import (
	"github.com/Assetsadapter/ninjacoin-adapter/ninjacoin"
	"github.com/astaxie/beego/config"
	"path/filepath"
	"time"
)

var (
	serverNode *ninjacoin.WalletManager
	clientNode *ninjacoin.WalletManager
)

func init() {

	//serverNode = testNewWalletManager("server.ini")
	//clientNode = testNewWalletManager("client.ini")
	clientNode = testNewWalletManager("server.ini")
	time.Sleep(1 * time.Second)
}

func testNewWalletManager(conf string) *ninjacoin.WalletManager {
	wm := ninjacoin.NewWalletManager()

	//读取配置
	absFile := filepath.Join("conf", conf)
	//log.Debug("absFile:", absFile)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return nil
	}
	wm.LoadAssetsConfig(c)
	return wm
}
