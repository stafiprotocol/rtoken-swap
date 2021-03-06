// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package stafix

import (
	"fmt"
	"math/big"
	"time"

	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/chainbridge/utils/blockstore"
)

type listener struct {
	name         string
	symbol       core.RSymbol
	care         core.RSymbol
	startBlock   uint64
	conn         *Connection
	blockstore   blockstore.Blockstorer
	router       chains.Router
	log          log15.Logger
	stop         <-chan int
	sysErr       chan<- error
	lastEraBlock uint64
}

// Frequency of polling for a new block
var (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 100
)

func NewListener(name string, symbol, care core.RSymbol, opts map[string]interface{}, startBlock uint64,
	conn *Connection, bs blockstore.Blockstorer, log log15.Logger, stop <-chan int, sysErr chan<- error) *listener {
	return &listener{
		name:         name,
		symbol:       symbol,
		care:         care,
		startBlock:   startBlock,
		conn:         conn,
		blockstore:   bs,
		log:          log,
		stop:         stop,
		sysErr:       sysErr,
		lastEraBlock: 0,
	}
}

func (l *listener) setRouter(r chains.Router) {
	l.router = r
}

// Start creates the initial subscription for all events
func (l *listener) start() error {
	latestBlk, err := l.conn.LatestBlockNumber()
	if err != nil {
		return err
	}

	if latestBlk < l.startBlock {
		return fmt.Errorf("starting block (%d) is greater than latest known block (%d)", l.startBlock, latestBlk)
	}

	go func() {
		err = l.pollBlocks()
		if err != nil {
			l.log.Error("Polling blocks failed", "err", err)
		}
	}()

	return nil
}

func (l *listener) pollBlocks() error {
	var nextDealBlock = l.startBlock
	var retry = BlockRetryLimit
	latestDealBlock, err := l.conn.GetLatestDealBlock(l.care)
	if err != nil {
		panic(err)
	}
	if latestDealBlock+1 > l.startBlock {
		nextDealBlock = latestDealBlock + 1
	}

	for {
		select {
		case <-l.stop:
			return ErrorTerminated
		default:
			// No more retries, goto next block
			if retry == 0 {
				l.sysErr <- fmt.Errorf("event polling retries exceeded: %s", l.symbol)
				return nil
			}

			finalBlk, err := l.conn.FinalizedBlockNumber()
			if err != nil {
				l.log.Error("Failed to fetch latest blockNumber", "err", err)
				retry--
				time.Sleep(BlockRetryInterval)
				continue
			}
			latestBlockFlag := finalBlk - l.blockDelay()

			//wait if next deal block is too new
			if nextDealBlock > latestBlockFlag {
				time.Sleep(BlockInterval)
				continue
			}
			//check ransinfo in next deal block
			transInfos, err := l.conn.GetTransInfos(l.care, nextDealBlock)
			//retry if err
			if err != nil && err != ErrNotExist {
				l.log.Error("Failed to fetch get transinfo ", "err", err, "block", nextDealBlock)
				retry--
				time.Sleep(BlockRetryInterval)
				continue
			}
			//process if has some transInfos
			if err == nil && len(transInfos.List) > 0 {
				l.log.Info("get new transInfos", "trans block", transInfos.Block, "dest", transInfos.DestSymbol)
				err := l.processTransInfos(transInfos)
				if err != nil {
					l.sysErr <- fmt.Errorf("processTransInfos failed: %s", err)
					return err
				}
			}
			if nextDealBlock%100 == 0 {
				l.log.Info("next deal block", "blocknumber", nextDealBlock)
			}
			// Write to blockstore which is already dealed
			err = l.blockstore.StoreBlock(new(big.Int).SetUint64(nextDealBlock))
			if err != nil {
				l.log.Error("Failed to write to blockstore", "err", err)
			}

			nextDealBlock++
			retry = BlockRetryLimit
		}
	}
}

func (l *listener) processTransInfos(infos *submodel.TransInfoList) error {

	switch infos.DestSymbol {
	case core.RATOM, core.RBNB, core.RMATIC, core.RSOL:
		//check transinfo all not dealed or all dealed
		allDeal := true
		for _, transInfo := range infos.List {
			if !transInfo.IsDeal {
				allDeal = false
				break
			}
		}

		//skip is all deal
		if allDeal {
			l.log.Warn("transInfoList allready dealed will skip: %+v", infos)
			return nil
		}

		msg := &core.Message{Destination: infos.DestSymbol, Reason: core.NewTransInfos, Content: infos}
		err := l.submitWriteMessage(msg)
		if err != nil {
			return fmt.Errorf("submitWriteMessage %s", err)
		}
		// wait until it is deal ,then continue to deal next
		retry := 0
		for {
			if retry > BlockRetryLimit {
				return fmt.Errorf("l.conn.TransInfoIsDeal reach retry limit, dest symbol:%s,block:%d,index:%d",
					infos.DestSymbol, infos.Block, 0)
			}
			isDeal, err := l.conn.TransInfoIsDeal(infos.DestSymbol, infos.Block, 0)
			if err == nil && isDeal {
				l.log.Info("TransInfo has deal", "symbol", infos.DestSymbol, "block", infos.Block, "index", 0)
				break
			}
			l.log.Warn("TransInfo still not deal, will wait...", "symbol", infos.DestSymbol, "block", infos.Block, "index", 0)
			retry++
			time.Sleep(BlockRetryInterval)
		}
		return nil
	case core.RDOT, core.RKSM, core.RFIS:
		for i, transInfo := range infos.List {
			//only transfer not dealed
			if transInfo.IsDeal {
				l.log.Warn("transInfo allready dealed will skip: %+v", transInfo)
				continue
			} else {
				infoSingle := submodel.TransInfoSingle{
					Block:      infos.Block,
					Index:      uint32(i),
					DestSymbol: infos.DestSymbol,
					Info:       transInfo,
				}
				msg := &core.Message{Destination: infos.DestSymbol, Reason: core.NewTransInfoSingle, Content: &infoSingle}
				err := l.submitWriteMessage(msg)
				if err != nil {
					return fmt.Errorf("submitWriteMessage %s", err)
				}
				// wait until it is deal ,then continue to deal next
				retry := 0
				for {
					if retry > BlockRetryLimit {
						return fmt.Errorf("l.conn.TransInfoIsDeal reach retry limit, dest symbol:%s,block:%d,index:%d",
							infos.DestSymbol, infos.Block, i)
					}
					isDeal, err := l.conn.TransInfoIsDeal(infos.DestSymbol, infos.Block, i)
					if err == nil && isDeal {
						l.log.Info("TransInfoSingle has deal", "symbol", infos.DestSymbol, "block", infos.Block, "index", i)
						break
					}
					l.log.Warn("TransInfoSingle still not deal, will wait...", "symbol", infos.DestSymbol, "block", infos.Block, "index", i)
					retry++
					time.Sleep(BlockRetryInterval)
				}
			}
		}

		return nil
	default:
		return fmt.Errorf("unsupport care symbol: %s", l.care)
	}
}

// submitMessage send msg to other chain
func (l *listener) submitWriteMessage(m *core.Message) error {
	m.Source = l.symbol
	if len(m.Destination) == 0 {
		m.Destination = l.care
	}
	return l.router.SendWriteMesage(m)
}

func (l *listener) blockDelay() uint64 {
	switch l.symbol {
	case core.FIS:
		return 1
	default:
		return 0
	}
}
