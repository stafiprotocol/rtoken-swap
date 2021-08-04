// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package stafi

import (
	"fmt"
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
	blockstore   blockstore.Blockstorer
	conn         *Connection
	router       chains.Router
	log          log15.Logger
	stop         <-chan int
	sysErr       chan<- error
	lastEraBlock uint64
}

// Frequency of polling for a new block
var (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 20
)

func NewListener(name string, symbol, care core.RSymbol, opts map[string]interface{}, startBlock uint64, bs blockstore.Blockstorer, conn *Connection, log log15.Logger, stop <-chan int, sysErr chan<- error) *listener {
	return &listener{
		name:         name,
		symbol:       symbol,
		care:         care,
		startBlock:   startBlock,
		blockstore:   bs,
		conn:         conn,
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
				l.log.Info("processTransInfos", "info", transInfos)
				err := l.processTransInfos(transInfos)
				if err != nil {
					panic(err)
				}
			}
			if nextDealBlock%100 == 0 {
				l.log.Info("next deal block", "blocknumber", nextDealBlock)
			}

			nextDealBlock++
			retry = BlockRetryLimit
		}
	}
}

func (l *listener) processTransInfos(infos *submodel.TransInfoList) error {

	switch l.care {
	case core.RATOM:
		//check transinfo is not deal
		for _, transInfo := range infos.List {
			if transInfo.IsDeal {
				return fmt.Errorf("transInfo must all not deal, symbol: %s block: %d", l.care, infos.Block)
			}
		}
		msg := &core.Message{Destination: l.care, Reason: core.NewTransInfos, Content: infos}
		return l.submitWriteMessage(msg)
	case core.RDOT, core.RKSM:
		needDeal := false
		for i, transInfo := range infos.List {
			if !transInfo.IsDeal {
				needDeal = true
				infoSingle := submodel.TransInfoSingle{
					Block: infos.Block,
					Index: uint32(i),
					Info:  transInfo,
				}
				msg := &core.Message{Destination: l.care, Reason: core.NewTransInfoSingle, Content: &infoSingle}
				err := l.submitWriteMessage(msg)
				if err != nil {
					return fmt.Errorf("submitWriteMessage %s", err)
				}
				//todo wait until pre is deal ,then continue to deal next
			}
		}
		//if no need to deal, it should't happend here
		if !needDeal {
			return fmt.Errorf("transInfoList must has info that need deal, symbol:%s block:%d", l.care, infos.Block)
		}
		return nil
	default:
		return fmt.Errorf("unsupport care symbol: %s", l.care)
	}
}

// submitMessage send msg to other chain
func (l *listener) submitWriteMessage(m *core.Message) error {
	m.Source = l.symbol
	if m.Destination == "" {
		m.Destination = l.care
	}
	return l.router.SendWriteMesage(m)
}

func (l *listener) blockDelay() uint64 {
	switch l.symbol {
	case core.RFIS:
		return 5
	default:
		return 0
	}
}
