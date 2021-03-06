package ninjacoin

import (
	"fmt"
	"net/http"

	"github.com/Assetsadapter/ninjacoin-adapter/ninjacoin/walletapi"
	"github.com/blocktree/openwallet/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

// A Client is a Bitcoin RPC BlockClient. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type ApiClient struct {
	WalletAPI, BlockAPI string
	Debug               bool
	BlockClient         *req.Req
	WalletClient        *walletapi.WalletAPI
}

func NewWalletClient(blockAPIUrl, walletIpAddr, walletIpPort, walletRpcPassword string, walletObj walletapi.Wallet, debug bool) *ApiClient {

	//walletAPI = strings.TrimSuffix(walletAPI, "/")
	//explorerAPI = strings.TrimSuffix(explorerAPI, "/")
	c := ApiClient{
		WalletAPI: "http://" + walletIpAddr + ":" + walletIpPort,
		//BlockAPI: "http://127.0.0.1:8547",
		BlockAPI: blockAPIUrl,
		Debug:    debug,
	}
	c.BlockClient = req.New()

	//wallet :=  walletapi.Wallet{
	//	Filename:   "mywallet_001.wallet",
	//	Password:   "123456",
	//	DaemonHost: "27.0.0.1",
	//	DaemonPort: 11801,
	//}

	wapi := walletapi.InitWalletAPI(walletRpcPassword, walletIpAddr, walletIpPort)
	wapi.OpenWallet(&walletObj)
	c.WalletClient = wapi
	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (c *ApiClient) call(method string, request interface{}) (*gjson.Result, error) {

	var (
		body = make(map[string]interface{}, 0)
	)

	if c.BlockClient == nil {
		return nil, fmt.Errorf("API url is not setup. ")
	}

	authHeader := req.Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	//json-rpc
	body["jsonrpc"] = "2.0"
	body["id"] = 1
	body["method"] = method
	body["params"] = request

	if c.Debug {
		log.Std.Info("Start Request API...")
	}

	r, err := c.BlockClient.Post(c.WalletAPI, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = c.isError(r)
	if err != nil {
		return nil, err
	}

	result := resp.Get("result")

	return &result, nil
}

// GET
func (c *ApiClient) get(path string) (*gjson.Result, error) {

	if c.BlockClient == nil {
		return nil, fmt.Errorf("API url is not setup. ")
	}

	if c.Debug {
		log.Std.Info("Start Request API...")
	}

	path = c.BlockAPI + "/" + path

	r, err := c.BlockClient.Get(path)

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = c.isError(r)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

//isError ????????????
func (c *ApiClient) isError(r *req.Resp) error {

	if r.Response().StatusCode != http.StatusOK {
		message := r.Response().Status
		status := r.Response().StatusCode
		return fmt.Errorf("[%d]%s", status, message)
	}

	result := gjson.ParseBytes(r.Bytes())

	if result.Get("error").IsObject() {

		return fmt.Errorf("[%d]%s",
			result.Get("error.code").Int(),
			result.Get("error.message").String())

	}

	return nil

}

//CreateAddress
func (c *ApiClient) CreateAddress() (string, error) {

	addressMap, err := c.WalletClient.CreateAddress()
	if err != nil {
		return "", err
	}
	return addressMap["address"], nil
}

// CreateBatchAddress ??????????????????
// @count ??????????????????
// @workerSize ????????????????????????20??????
func (c *ApiClient) CreateBatchAddress(count, workerSize uint64) ([]string, error) {

	var (
		quit         = make(chan struct{})
		done         = uint64(0) //????????????
		failed       = uint64(0)
		shouldDone   = count //?????????????????????
		addressArr   = make([]string, 0)
		workPermitCH = make(chan struct{}, workerSize) //????????????
		producer     = make(chan AddressCreateResult)  //????????????
		worker       = make(chan AddressCreateResult)  //????????????
	)

	defer func() {
		close(workPermitCH)
		close(producer)
		close(worker)
	}()

	if count == 0 {
		return nil, fmt.Errorf("create address count is zero")
	}

	//????????????
	consumeWork := func(result chan AddressCreateResult) {
		//?????????????????????
		for gets := range result {

			if gets.Success {
				addressArr = append(addressArr, gets.Address)
			} else {
				failed++ //?????????????????????
			}

			//????????????????????????
			done++
			if done == shouldDone {
				//bs.wm.Log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
				close(quit) //????????????????????????????????????nil
			}
		}
	}

	//????????????
	produceWork := func(eCount uint64, eProducer chan AddressCreateResult) {
		for i := uint64(0); i < uint64(eCount); i++ {
			workPermitCH <- struct{}{}
			go func(end chan struct{}, mProducer chan<- AddressCreateResult) {

				//????????????
				addr, createErr := c.CreateAddress()
				result := AddressCreateResult{
					Success: true,
					Address: addr,
					Err:     createErr,
				}
				mProducer <- result
				//??????
				<-end

			}(workPermitCH, eProducer)
		}
	}

	//????????????????????????
	go consumeWork(worker)

	//????????????????????????
	go produceWork(count, producer)

	//??????????????????????????????
	batchCreateAddressRuntime(producer, worker, quit)

	if failed > 0 {
		log.Debugf("create address failed: %d", failed)
	}

	return addressArr, nil
}

//batchCreateAddressRuntime ?????????
func batchCreateAddressRuntime(producer chan AddressCreateResult, worker chan AddressCreateResult, quit chan struct{}) {

	var (
		values = make([]AddressCreateResult, 0)
	)

	for {

		var activeWorker chan<- AddressCreateResult
		var activeValue AddressCreateResult

		//???????????????????????????????????????????????????????????????
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//?????????????????????????????????????????????????????????
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//??????
			return
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}

}

//GetAddressList
func (c *ApiClient) GetAddressList() ([]string, error) {

	request := map[string]interface{}{
		"own": true,
	}

	r, err := c.call("addr_list", request)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0)
	if r.IsArray() {
		for _, a := range r.Array() {
			own := a.Get("own").Bool()
			expired := a.Get("expired").Bool()
			commet := a.Get("comment").String()
			if own && expired == false && "self" == commet {
				addrs = append(addrs, a.Get("address").String())
			}

		}
	}

	return addrs, nil
}

//SendTransaction
func (c *ApiClient) SendTransaction(to, amount, paymentId string) (string, error) {

	sendAmount := convertFromAmount(amount, Decimal)

	txId, err := c.WalletClient.SendTransactionBasic(to, paymentId, sendAmount)
	if err != nil {
		return "", err
	}
	return txId, nil
}

//GetBlockByHeight
func (c *ApiClient) GetBlockByHeight(height uint64) (*Block, error) {
	path := fmt.Sprintf("block/%d", height)
	r, err := c.get(path)
	if err != nil {
		return nil, err
	}
	block := NewBlock(r)

	txs, err := c.WalletClient.GetTransactionsInRange(height, height+1)
	if err != nil {
		return nil, err
	}
	block.Txs = *txs
	log.Infof("startHeight=%d endHeight=%d tx_len=%d", height, height+1, len(block.Txs))
	return block, nil
}

//GetBlockByHash
func (c *ApiClient) GetBlockByHash(hash string) (*Block, error) {
	path := fmt.Sprintf("block?hash=%s", hash)
	r, err := c.get(path)
	if err != nil {
		return nil, err
	}
	block := NewBlock(r)
	return block, nil
}

//GetTransaction
func (c *ApiClient) GetTransaction(txid string) (*Transaction, error) {
	request := map[string]interface{}{
		"txId": txid,
	}

	r, err := c.call("tx_status", request)
	if err != nil {
		return nil, err
	}
	return NewTransaction(r), nil
}

//GetTransactionsByHeight
func (c *ApiClient) GetTransactionsByHeight(height uint64) ([]*Transaction, error) {
	request := map[string]interface{}{
		"filter": map[string]interface{}{
			//"status": 4,
			"height": height,
		},
		//"skip":  0,
		//"count": 10,
	}

	r, err := c.call("tx_list", request)
	if err != nil {
		return nil, err
	}

	txs := make([]*Transaction, 0)
	if r.IsArray() {
		for _, obj := range r.Array() {
			tx := NewTransaction(&obj)
			txs = append(txs, tx)

		}
	}

	return txs, nil
}

//GetWalletStatus
func (c *ApiClient) GetWalletStatus() (*WalletStatus, error) {

	r, err := c.get("block/last")
	if err != nil {
		return nil, err
	}

	return NewWalletStatus(r), nil
}

//??????????????????
func (c *ApiClient) GetAddressBalance(address string) (string, error) {

	r, err := c.WalletClient.GetAddressBalance(address)
	if err != nil {
		return "", err
	}
	return convertToAmount(r.Unlocked, Decimal), nil

}

//????????????
func (c *ApiClient) ValidateAddress(address string) bool {

	_, err := c.WalletClient.ValidateAddress(address)
	if err != nil {
		return false
	}
	return true
}

//GetWalletStatus
func (c *ApiClient) optimizeWallet() (string, error) {

	txId, err := c.WalletClient.SendFusionBasic()
	if err != nil {
		return "", err
	}

	return txId, nil
}
