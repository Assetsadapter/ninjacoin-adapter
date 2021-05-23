package ninjacoin

import (
	"fmt"
	"path/filepath"

	"github.com/astaxie/beego/logs"
	"github.com/blocktree/openwallet/common/file"
	"github.com/blocktree/openwallet/log"
)

//SetupLog 配置日志
func (wm *WalletManager) SetupLog(logDir, logFile string, debug bool) {

	//记录日志
	logLevel := log.LevelInformational
	if debug {
		logLevel = log.LevelDebug
	}

	if len(logDir) > 0 {
		file.MkdirAll(logDir)
		logFile := filepath.Join(logDir, logFile)
		logConfig := fmt.Sprintf(`{"filename":"%s","level":%d,"daily":true,"maxdays":7,"maxsize":0}`, logFile, logLevel)
		//log.Println(logConfig)
		wm.Log.SetLogger(logs.AdapterFile, logConfig)
		wm.Log.SetLogger(logs.AdapterConsole, logConfig)
		log.SetLevel(logLevel)
	} else {
		wm.Log.SetLevel(logLevel)
		log.SetLevel(logLevel)
	}
}
