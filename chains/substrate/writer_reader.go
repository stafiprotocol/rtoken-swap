// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package substrate

import (
	"fmt"
	"rtoken-swap/chains"
	"rtoken-swap/core"

	"github.com/ChainSafe/log15"
)

const msgLimit = 1024

type writer struct {
	symbol          core.RSymbol
	conn            *Connection
	router          chains.Router
	msgChan         chan *core.Message
	log             log15.Logger
	sysErr          chan<- error
	currentChainEra uint32
	stop            <-chan int
}

func NewReaderWriter(symbol core.RSymbol, opts map[string]interface{}, conn *Connection, log log15.Logger, sysErr chan<- error, stop <-chan int) *writer {

	return &writer{
		symbol:          symbol,
		conn:            conn,
		log:             log,
		sysErr:          sysErr,
		msgChan:         make(chan *core.Message, msgLimit),
		currentChainEra: 0,
		stop:            stop,
	}
}

func (w *writer) setRouter(r chains.Router) {
	w.router = r
}

func (w *writer) Start() error {
	go w.msgHandler()
	return nil
}

func (w *writer) msgHandler() error {
out:
	for {
		select {
		case <-w.stop:
			w.log.Info("msgHandler stop")
			break out
		case msg := <-w.msgChan:
			w.resolveWriteMessage(msg)
		}
	}
	return nil
}

func (w *writer) QueueWriteMessage(m *core.Message) {
	w.msgChan <- m
}

func (w *writer) ResolveReadMessage(m *core.Message) {
}

func (w *writer) resolveWriteMessage(m *core.Message) (processOk bool) {
	defer func() {
		if !processOk {
			panic(fmt.Sprintf("resolveMessage process failed. %+v", m))
		}
	}()

	switch m.Reason {
	case core.NewTransInfos:

	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return true
	}
	return
}

func (w *writer) printContentError(m *core.Message) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason)
}
