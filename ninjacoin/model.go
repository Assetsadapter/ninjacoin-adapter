package ninjacoin

import (
	"github.com/Assetsadapter/ninjacoin-adapter/ninjacoin/walletapi"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/tidwall/gjson"
)

type Block struct {
	Hash          string
	PrevBlockHash string
	Height        uint64
	Txs           []walletapi.Transaction
}

func NewBlock(result *gjson.Result) *Block {
	obj := Block{}
	//解析json
	obj.Hash = result.Get("hash").String()
	obj.PrevBlockHash = result.Get("prevHash").String()
	obj.Height = result.Get("height").Uint()

	return &obj
}

//BlockHeader 区块链头
func (b *Block) BlockHeader(symbol string) *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	//obj.Confirmations = b.Confirmations
	//obj.Merkleroot = b.TransactionMerkleRoot
	obj.Previousblockhash = b.PrevBlockHash
	obj.Height = b.Height
	obj.Time = uint64(obj.Time)
	obj.Symbol = symbol

	return &obj
}

type Transaction struct {
	Comment       string
	CreateTime    int64
	Fee           uint64
	TxID          string
	Value         uint64
	Kernel        string
	Receiver      string
	Sender        string
	Income        bool
	Status        int64
	StatusString  string
	Confirmations uint64
	BlockHeight   uint64
	BlockHash     string

	/*
			{
		        "comment": "",
		        "confirmations": 5,
		        "create_time": 1559873867,
		        "fee": 1,
		        "height": 221412,
		        "income": true,
		        "kernel": "cf3634952569171015fe08b949ed617692a30947747fb576e826d6f48a1b8035",
		        "receiver": "21aff5eb4da2591321ac12bb280ac69ea39a33472166c600ec122cf3381b6c9e772",
		        "sender": "22d090004ab6de7e62d0d3829e0164d05cc065404ebc9874d181dc070d54237bbd8",
		        "status": 3,
		        "status_string": "received",
		        "txId": "72f8f349f9244b11b0e6471250ca68a1",
		        "value": 10000
		    }
	*/
}

func NewTransaction(result *gjson.Result) *Transaction {
	obj := Transaction{}
	obj.Comment = result.Get("comment").String()
	obj.CreateTime = result.Get("create_time").Int()
	obj.Fee = result.Get("fee").Uint()
	obj.Income = result.Get("income").Bool()
	obj.Kernel = result.Get("kernel").String()
	obj.Receiver = result.Get("receiver").String()
	obj.Sender = result.Get("sender").String()
	obj.Status = result.Get("status").Int()
	obj.StatusString = result.Get("status_string").String()
	obj.TxID = result.Get("txId").String()
	obj.Value = result.Get("value").Uint()
	obj.Confirmations = result.Get("confirmations").Uint()
	obj.BlockHeight = result.Get("height").Uint()

	return &obj
}

type TrustNodeInfo struct {
	NodeID      string `json:"nodeID"` //@required 节点ID
	NodeName    string `json:"nodeName"`
	ConnectType string `json:"connectType"`
}

type WalletStatus struct {
	CurrentHeight uint64
	CurrentHash   string
}

func NewWalletStatus(result *gjson.Result) *WalletStatus {
	obj := WalletStatus{}
	obj.CurrentHeight = result.Get("height").Uint()
	obj.CurrentHash = result.Get("hash").String()
	return &obj
}

type AddressCreateResult struct {
	Success bool
	Err     error
	Address string
}
