package ninjacoin

import (
	"fmt"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		to      string
		amount  string
		txFrom  []string
		txTo    []string
		fixFees *big.Int
	)

	for k, v := range rawTx.To {
		to = k
		amount = v
	}

	amountDec, _ := decimal.NewFromString(amount)
	amountDec = amountDec.Shift(decoder.wm.Decimal())

	//取一个地址作为发送
	addresses, err := decoder.wm.walletClient.GetAddressList()
	if err != nil {
		return err
	}

	if addresses == nil || len(addresses) == 0 {
		return openwallet.Errorf(openwallet.ErrAccountNotAddress, "wallet address is not created")
	}

	from := addresses[0]

	if len(rawTx.FeeRate) > 0 {
		fixFees = common.StringNumToBigIntWithExp(rawTx.FeeRate, decoder.wm.Decimal())
		rawTx.Fees = rawTx.FeeRate
	} else {
		fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.fixfees, decoder.wm.Decimal())
		rawTx.FeeRate = decoder.wm.Config.fixfees
		rawTx.Fees = decoder.wm.Config.fixfees
	}

	if fixFees.Cmp(big.NewInt(0)) <= 0 {
		return openwallet.Errorf(openwallet.ErrUnknownException, "fee is lower than 0")
	}

	//_, err := decoder.wm.walletClient.GetWalletStatus()
	//if err != nil {
	//	return err
	//}

	sendAmount := uint64(amountDec.IntPart())

	//判断钱包余额是否足够
	//if walletStatus.Available < sendAmount+fixFees.Uint64() {
	//	return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "wallet available balance is not enough")
	//}
	if 0 < sendAmount+fixFees.Uint64() {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "wallet available balance is not enough")
	}
	txFrom = []string{fmt.Sprintf("%s:%s", from, amount)}
	txTo = []string{fmt.Sprintf("%s:%s", to, amount)}

	rawTx.FeeRate = decoder.wm.Config.fixfees
	rawTx.Fees = decoder.wm.Config.fixfees
	rawTx.IsBuilt = true
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil

}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	return nil
}

//
//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	rawTx.IsCompleted = true

	return nil
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	var (
		to      string
		amount  string
		fixFees *big.Int
	)

	for k, v := range rawTx.To {
		to = k
		amount = v
	}

	amountDec, _ := decimal.NewFromString(amount)
	amountDec = amountDec.Shift(decoder.wm.Decimal())

	//取一个地址作为发送
	addresses, err := decoder.wm.walletClient.GetAddressList()
	if err != nil {
		return nil, err
	}

	if addresses == nil || len(addresses) == 0 {
		return nil, fmt.Errorf("wallet address is not created")
	}

	from := addresses[0]

	if len(rawTx.FeeRate) > 0 {
		fixFees = common.StringNumToBigIntWithExp(rawTx.FeeRate, decoder.wm.Decimal())
		rawTx.Fees = rawTx.FeeRate
	} else {
		fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.fixfees, decoder.wm.Decimal())
		rawTx.FeeRate = decoder.wm.Config.fixfees
		rawTx.Fees = decoder.wm.Config.fixfees
	}

	if fixFees.Cmp(big.NewInt(0)) <= 0 {
		return nil, openwallet.Errorf(openwallet.ErrUnknownException, "fee is lower than 0")
	}

	//walletStatus, err := decoder.wm.walletClient.GetWalletStatus()
	if err != nil {
		return nil, err
	}

	sendAmount := uint64(amountDec.IntPart())

	//判断钱包余额是否足够
	//if walletStatus.Available < sendAmount+fixFees.Uint64() {
	//	return nil, openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "wallet available balance is not enough")
	//}

	txid, err := decoder.wm.walletClient.SendTransaction(from, to, sendAmount, fixFees.Uint64(), "")
	if err != nil {
		return nil, err
	}

	decoder.wm.Log.Infof("Transaction [%s] submitted to the network successfully.", txid)

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	txFrom := []string{fmt.Sprintf("%s:%s", from, amount)}
	txTo := []string{fmt.Sprintf("%s:%s", to, amount)}

	decimals := decoder.wm.Decimal()

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       txFrom,
		To:         txTo,
		Amount:     amount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		Fees:       rawTx.Fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return decoder.wm.Config.fixfees, "TX", nil
}

//CreateSummaryRawTransaction 创建汇总交易
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	var (
		rawTxWithErrArray []*openwallet.RawTransactionWithError
		rawTxArray        = make([]*openwallet.RawTransaction, 0)
		err               error
	)
	rawTxWithErrArray, err = decoder.CreateSummaryRawTransactionWithError(wrapper, sumRawTx)
	if err != nil {
		return nil, err
	}
	for _, rawTxWithErr := range rawTxWithErrArray {
		if rawTxWithErr.Error != nil {
			continue
		}
		rawTxArray = append(rawTxArray, rawTxWithErr.RawTx)
	}
	return rawTxArray, nil
}

//CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		decimals        = decoder.wm.Decimal()
		rawTxArray      = make([]*openwallet.RawTransactionWithError, 0)
		minTransfer     = common.StringNumToBigIntWithExp(sumRawTx.MinTransfer, decimals)
		retainedBalance = common.StringNumToBigIntWithExp(sumRawTx.RetainedBalance, decimals)
		fixFees         = big.NewInt(0)
	)

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	if len(sumRawTx.FeeRate) > 0 {
		fixFees = common.StringNumToBigIntWithExp(sumRawTx.FeeRate, decoder.wm.Decimal())
	} else {
		fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.fixfees, decoder.wm.Decimal())
		sumRawTx.FeeRate = decoder.wm.Config.fixfees
	}

	if fixFees.Cmp(big.NewInt(0)) <= 0 {
		return nil, openwallet.Errorf(openwallet.ErrUnknownException, "fee is lower than 0")
	}

	//walletStatus, err := decoder.wm.walletClient.GetWalletStatus()
	//if err != nil {
	//	return nil, err
	//}

	//检查余额是否超过最低转账
	addrBalance_BI := new(big.Int)
	addrBalance_BI.SetUint64(0)
	addrBalance := common.IntToDecimals(int64(0), decoder.wm.Decimal())

	if addrBalance_BI.Cmp(minTransfer) < 0 || addrBalance_BI.Cmp(big.NewInt(0)) <= 0 {
		return rawTxArray, nil
	}
	//计算汇总数量 = 余额 - 保留余额
	sumAmount_BI := new(big.Int)
	sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

	//减去手续费
	sumAmount_BI.Sub(sumAmount_BI, fixFees)
	if sumAmount_BI.Cmp(big.NewInt(0)) <= 0 {
		return rawTxArray, nil
	}

	sumAmount := common.BigIntToDecimals(sumAmount_BI, decimals)
	feesAmount := common.BigIntToDecimals(fixFees, decimals)

	decoder.wm.Log.Debugf("balance: %v", addrBalance.String())
	decoder.wm.Log.Debugf("fees: %v", feesAmount)
	decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

	//创建一笔交易单
	rawTx := &openwallet.RawTransaction{
		Coin:    sumRawTx.Coin,
		Account: sumRawTx.Account,
		To: map[string]string{
			sumRawTx.SummaryAddress: sumAmount.StringFixed(decoder.wm.Decimal()),
		},
		Required: 1,
	}

	createTxErr := decoder.CreateRawTransaction(wrapper, rawTx)
	rawTxWithErr := &openwallet.RawTransactionWithError{
		RawTx: rawTx,
		Error: openwallet.ConvertError(createTxErr),
	}

	//创建成功，添加到队列
	rawTxArray = append(rawTxArray, rawTxWithErr)

	return rawTxArray, nil
}
