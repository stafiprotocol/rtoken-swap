// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package substrate

import (
	"fmt"
	"time"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/chainbridge/utils/blockstore"
	"rtoken-swap/chains"
	"rtoken-swap/config"
	"rtoken-swap/core"
)

type listener struct {
	name          string
	symbol        core.RSymbol
	startBlock    uint64
	blockstore    blockstore.Blockstorer
	conn          *Connection
	subscriptions map[eventName]eventHandler // Handlers for specific events
	router        chains.Router
	log           log15.Logger
	stop          <-chan int
	sysErr        chan<- error
	lastEraBlock  uint64
}

// Frequency of polling for a new block
var (
	OneBlockTime              = 5 * time.Second
	BlockRetryInterval        = time.Second * 1
	BlockRetryLimit           = 5
	BlockIntervalToProcessEra = uint64(10)
	EventRetryLimit           = 10
	EventRetryInterval        = 100 * time.Millisecond
)

func NewListener(name string, symbol core.RSymbol, opts map[string]interface{}, startBlock uint64, bs blockstore.Blockstorer, conn *Connection, log log15.Logger, stop <-chan int, sysErr chan<- error) *listener {

	return &listener{
		name:          name,
		symbol:        symbol,
		startBlock:    startBlock,
		blockstore:    bs,
		conn:          conn,
		subscriptions: make(map[eventName]eventHandler),
		log:           log,
		stop:          stop,
		sysErr:        sysErr,
		lastEraBlock:  0,
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

	for _, sub := range Subscriptions {
		err := l.registerEventHandler(sub.name, sub.handler)
		if err != nil {
			return err
		}
	}

	go func() {
		err = l.pollBlocks()
		if err != nil {
			l.log.Error("Polling blocks failed", "err", err)
		}
	}()

	return nil
}

// registerEventHandler enables a handler for a given event. This cannot be used after Start is called.
func (l *listener) registerEventHandler(name eventName, handler eventHandler) error {
	if l.subscriptions[name] != nil {
		return fmt.Errorf("event %s already registered", name)
	}
	l.subscriptions[name] = handler
	return nil
}

// pollBlocks will poll for the latest block and proceed to parse the associated events as it sees new blocks.
// Polling begins at the block defined in `l.startBlock`. Failed attempts to fetch the latest block or parse
// a block will be retried up to BlockRetryLimit times before returning with an error.
func (l *listener) pollBlocks() error {
	var currentBlock = l.startBlock
	var retry = BlockRetryLimit
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

			// Sleep if the block we want comes after the most recently finalized block
			if currentBlock+l.blockDelay() > finalBlk {
				if currentBlock%100 == 0 {
					l.log.Trace("Block not yet finalized", "target", currentBlock, "finalBlk", finalBlk)
				}
				time.Sleep(OneBlockTime)
				continue
			}

			err = l.processEvents(currentBlock)
			if err != nil {
				l.log.Error("Failed to process events in block", "block", currentBlock, "err", err)
				retry--
				continue
			}

			currentBlock++
			retry = BlockRetryLimit
		}
	}
}

// processEvents fetches a block and parses out the events, calling Listener.handleEvents()
func (l *listener) processEvents(blockNum uint64) error {
	if blockNum%100 == 0 {
		l.log.Debug("processEvents", "blockNum", blockNum)
	}
	evts, err := l.conn.GetEvents(blockNum)
	if err != nil {
		for i := 0; i < EventRetryLimit; i++ {
			time.Sleep(EventRetryInterval)
			evts, err = l.conn.GetEvents(blockNum)
			if err == nil {
				break
			}
		}
		if err != nil {
			return err
		}
	}

	for _, evt := range evts {
		switch {
		case evt.ModuleId == config.MultisigModuleId && evt.EventId == config.NewMultisigEventId:
			l.log.Trace("Handling NewMultisigEvent", "block", blockNum)
			flow, err := l.processNewMultisigEvt(evt)
			if err != nil {
				if err.Error() == ErrMultiEnd.Error() {
					l.log.Info("listener received an ended NewMultisig event, ignored")
					continue
				}
				return err
			}
			if l.subscriptions[NewMultisig] != nil {
				l.submitWriteMessage(l.subscriptions[NewMultisig](flow))
			}
		case evt.ModuleId == config.MultisigModuleId && evt.EventId == config.MultisigExecutedEventId:
			l.log.Trace("Handling MultisigExecutedEvent", "block", blockNum)
			flow, err := l.processMultisigExecutedEvt(evt)
			if err != nil {
				return err
			}
			if l.subscriptions[MultisigExecuted] != nil {
				l.submitWriteMessage(l.subscriptions[MultisigExecuted](flow))
			}
		}
	}

	return nil
}

// submitMessage inserts the chainId into the msg and sends it to the router
func (l *listener) submitWriteMessage(m *core.Message, err error) {
	if err != nil {
		l.log.Error("Critical error before sending message", "err", err)
		return
	}
	m.Source = l.symbol
	if m.Destination == "" {
		m.Destination = m.Source
	}
	err = l.router.SendWriteMesage(m)
	if err != nil {
		l.log.Error("failed to send message", "err", err)
	}
}

func (l *listener) blockDelay() uint64 {
	switch l.symbol {
	case core.RFIS:
		return 5
	default:
		return 0
	}
}
