package ninjacoin

import (
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"testing"
)

func TestBEAMBlockScanner_GetCurrentBlock(t *testing.T) {
	b, err := tw.Blockscanner.GetCurrentBlock()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("block = %+v", b)
}

func TestBEAMBlockScanner_GetCurrentBlockHeader(t *testing.T) {
	wm := testNewWalletManager()
	header, err := wm.GetBlockScanner().GetCurrentBlockHeader()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("header: %v", header)
}

func TestGetBlockHeight(t *testing.T) {
	wm := testNewWalletManager()
	height, err := wm.Blockscanner.GetCurrentBlock()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("height: %v", height)
}

func TestBEAMBlockScanner_GetBlockByHeight(t *testing.T) {
	wm := testNewWalletManager()
	block, err := wm.Blockscanner.GetBlockByHeight(67213)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("block: %v", block)
}

func TestBEAMBlockScanner_GetBlockByHash(t *testing.T) {
	//wm := testNewWalletManager()
	//block, err := wm.Blockscanner.GetBlockByHash("7oVvXaXdb8nF47ffzPJ3VKt57KfZgCBzDpGwpBojmr63")
	//if err != nil {
	//	t.Errorf("unexpected error: %v", err)
	//	return
	//}
	//log.Infof("block: %v", block)
}

func TestBEAMBlockScanner_GetTransaction(t *testing.T) {
	//wm := testNewWalletManager()
	//tx, err := wm.Blockscanner.GetTransaction("de2f7fef9d9948809c88117b385a7c30")
	//if err != nil {
	//	t.Errorf("unexpected error: %v", err)
	//	return
	//}
	//log.Infof("tx: %+v", tx)
}

func TestBEAMBlockScanner_ExtractTransactionData(t *testing.T) {

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanTargetFunc := func(target openwallet.ScanTarget) (string, bool) {
		if target.Address == "GAECXQFHEMVMYJ7UUHL6NXJVZAUGYRIGW6STKUNT3QUIQRPN2ULJXTQ7" {
			return "sender", true
		} else if target.Address == "GCUH7EXP2H4KW7KL2Z3NGTNM3U6L23F35UOOKCAS7M2EYTU7SWBC2DAW" {
			return "recipient", true
		}
		return "", false
	}

	txs, err := tw.Blockscanner.ExtractTransactionData("9ixa2UdT8MbuqxCWQvGWDPXN2vDBzHoj58udYwpUWedk", scanTargetFunc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	for key, datas := range txs {
		log.Notice("key:", key)
		for _, data := range datas {
			for i, input := range data.TxInputs {
				log.Infof("data.TxInputs[%d]: %+v", i, input)
			}

			for i, output := range data.TxOutputs {
				log.Infof("data.TxOutputs[%d]: %+v", i, output)
			}

			log.Infof("data.Transaction: %+v", data.Transaction)
		}
	}
}
