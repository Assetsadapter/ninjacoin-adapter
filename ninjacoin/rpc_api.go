package ninjacoin

import (
	"fmt"
	"github.com/Assetsadapter/ninjacoin-adapter/ninjacoin/walletapi"
	"github.com/blocktree/openwallet/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"net/http"
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

//isError 是否报错
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

// CreateBatchAddress 批量创建地址
// @count 连续创建数量
// @workerSize 并行线程数。建议20条。
func (c *ApiClient) CreateBatchAddress(count, workerSize uint64) ([]string, error) {

	var (
		quit         = make(chan struct{})
		done         = uint64(0) //完成标记
		failed       = uint64(0)
		shouldDone   = count //需要完成的总数
		addressArr   = make([]string, 0)
		workPermitCH = make(chan struct{}, workerSize) //工作令牌
		producer     = make(chan AddressCreateResult)  //生产通道
		worker       = make(chan AddressCreateResult)  //消费通道
	)

	defer func() {
		close(workPermitCH)
		close(producer)
		close(worker)
	}()

	if count == 0 {
		return nil, fmt.Errorf("create address count is zero")
	}

	//消费工作
	consumeWork := func(result chan AddressCreateResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {
				addressArr = append(addressArr, gets.Address)
			} else {
				failed++ //标记生成失败数
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				//bs.wm.Log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//生产工作
	produceWork := func(eCount uint64, eProducer chan AddressCreateResult) {
		for i := uint64(0); i < uint64(eCount); i++ {
			workPermitCH <- struct{}{}
			go func(end chan struct{}, mProducer chan<- AddressCreateResult) {

				//生成地址
				addr, createErr := c.CreateAddress()
				result := AddressCreateResult{
					Success: true,
					Address: addr,
					Err:     createErr,
				}
				mProducer <- result
				//释放
				<-end

			}(workPermitCH, eProducer)
		}
	}

	//独立线程运行消费
	go consumeWork(worker)

	//独立线程运行生产
	go produceWork(count, producer)

	//以下使用生产消费模式
	batchCreateAddressRuntime(producer, worker, quit)

	if failed > 0 {
		log.Debugf("create address failed: %d", failed)
	}

	return addressArr, nil
}

//batchCreateAddressRuntime 运行时
func batchCreateAddressRuntime(producer chan AddressCreateResult, worker chan AddressCreateResult, quit chan struct{}) {

	var (
		values = make([]AddressCreateResult, 0)
	)

	for {

		var activeWorker chan<- AddressCreateResult
		var activeValue AddressCreateResult

		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
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
func (c *ApiClient) SendTransaction(from, to string, value, fee uint64, comment string) (string, error) {

	request := map[string]interface{}{
		"value":   value,
		"fee":     fee,
		"from":    from,
		"address": to,
		"comment": comment,
	}

	r, err := c.call("tx_send", request)
	if err != nil {
		return "", err
	}
	return r.Get("txId").String(), nil
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

//获取地址余额
func (c *ApiClient) GetAddressBalance(address string) (string, error) {

	r, err := c.WalletClient.GetAddressBalance(address)
	if err != nil {
		return "", err
	}
	return convertToAmount(r.Unlocked, Decimal), nil

}

//CancelTx 取消交易
func (c *ApiClient) ValidateAddress(address string) (bool, error) {
	request := map[string]interface{}{
		"address": address,
	}

	r, err := c.call("validate_address", request)
	if err != nil {
		return false, err
	}

	return r.Get("is_valid").Bool(), nil
}
