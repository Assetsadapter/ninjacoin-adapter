/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package ninjacoin

import (
	"fmt"
	"github.com/blocktree/openwallet/openwallet"
)

//SaveLocalBlockHead 记录区块高度和hash到本地
func (bs *NinjaBlockScanner) SaveLocalBlockHead(blockHeight uint64, blockHash string) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	header := &openwallet.BlockHeader{
		Hash:   blockHash,
		Height: blockHeight,
		Fork:   false,
		Symbol: bs.wm.Symbol(),
	}

	bs.wm.Log.Std.Info("block scanner SaveLocalBlockHead: %v", header)

	return bs.BlockchainDAI.SaveCurrentBlockHead(header)
}

//GetLocalBlockHead 获取本地记录的区块高度和hash
func (bs *NinjaBlockScanner) GetLocalBlockHead() (uint64, string, error) {

	if bs.BlockchainDAI == nil {
		return 0, "", fmt.Errorf("Blockchain DAI is not setup ")
	}

	header, err := bs.BlockchainDAI.GetCurrentBlockHead(bs.wm.Symbol())
	if err != nil {
		return 0, "", err
	}

	return header.Height, header.Hash, nil
}

//SaveLocalBlock 记录本地新区块
func (bs *NinjaBlockScanner) SaveLocalBlock(blockHeader *openwallet.BlockHeader) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	header := &openwallet.BlockHeader{
		Hash:              blockHeader.Hash,
		Previousblockhash: blockHeader.Previousblockhash,
		Height:            blockHeader.Height,
		Time:              0,
		Symbol:            bs.wm.Symbol(),
	}

	//bs.wm.Log.Std.Info("block scanner SaveLocalBlock: %v", header)

	return bs.BlockchainDAI.SaveLocalBlockHead(header)
}

//GetLocalBlock 获取本地区块数据
func (bs *NinjaBlockScanner) GetLocalBlock(height uint64) (*openwallet.BlockHeader, error) {

	if bs.BlockchainDAI == nil {
		return nil, fmt.Errorf("Blockchain DAI is not setup ")
	}

	header, err := bs.BlockchainDAI.GetLocalBlockHeadByHeight(height, bs.wm.Symbol())
	if err != nil {
		return nil, err
	}

	block := &openwallet.BlockHeader{
		Hash:              header.Hash,
		Height:            header.Height,
		Previousblockhash: header.Previousblockhash,
	}

	bs.wm.Log.Std.Info("block scanner GetLocalBlock: %v", block)

	return block, nil
}

//SaveUnscanRecord 保存交易记录到钱包数据库
func (bs *NinjaBlockScanner) SaveUnscanRecord(record *openwallet.UnscanRecord) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.SaveUnscanRecord(record)
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (bs *NinjaBlockScanner) DeleteUnscanRecord(height uint64) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.DeleteUnscanRecordByHeight(height, bs.wm.Symbol())
}

func (bs *NinjaBlockScanner) GetUnscanRecords() ([]*openwallet.UnscanRecord, error) {

	if bs.BlockchainDAI == nil {
		return nil, fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.GetUnscanRecords(bs.wm.Symbol())
}
