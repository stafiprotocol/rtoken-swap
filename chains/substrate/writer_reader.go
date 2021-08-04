// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package substrate

import (
	"fmt"
	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
	"sync"

	"github.com/ChainSafe/log15"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const msgLimit = 1024

type writer struct {
	symbol          core.RSymbol
	conn            *Connection
	router          chains.Router
	eventMtx        sync.RWMutex
	newMulTicsMtx   sync.RWMutex
	events          map[string]*submodel.MultiEventFlow
	newMultics      map[string]*submodel.EventNewMultisig
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
		events:          make(map[string]*submodel.MultiEventFlow),
		newMultics:      make(map[string]*submodel.EventNewMultisig),
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
	case core.NewTransInfoSingle:

	case core.NewMultisig:
		return w.processNewMultisig(m)
	case core.MultisigExecuted:
		return w.processMultisigExecuted(m)
	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return true
	}
	return
}

func (w *writer) printContentError(m *core.Message) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason)
}

func (w *writer) processNewMultisig(m *core.Message) bool {
	flow, ok := m.Content.(*submodel.EventNewMultisig)
	if !ok {
		w.printContentError(m)
		return false
	}

	// w.conn.Address()
	// _, ok = w.getBondedPools(hexutil.Encode(flow.ID[:]))
	// if !ok {
	// 	w.log.Info("received a newMultisig event which the ID is not in the bondedPools, ignored")
	// 	return true
	// }

	w.setNewMultics(flow.CallHashStr, flow)

	evt, ok := w.getEvents(flow.CallHashStr)
	if !ok {
		w.log.Info("receive a newMultisig, wait for more flow data", "callHash", flow.CallHashStr)
		return true
	}

	identify := hexutil.Encode(evt.Key.PublicKey)
	for _, apv := range flow.Approvals {
		if identify == hexutil.Encode(apv[:]) {
			w.log.Info("receive a newMultisig which has already approved, will ignore", "callHash", flow.CallHashStr)
			return true
		}
	}

	for _, call := range evt.OpaqueCalls {
		if call.CallHash == flow.CallHashStr {
			call.TimePoint = flow.TimePoint
			delete(evt.NewMulCallHashs, flow.CallHashStr)
		}
	}

	if len(evt.NewMulCallHashs) != 0 {
		w.log.Info("processNewMultisig wait for more callhash", "eventId", evt.EventId)
		return true
	}

	err := w.conn.AsMulti(evt)
	if err != nil {
		w.log.Error("AsMulti error", "err", err, "callHash", flow.CallHashStr)
		return false
	}

	w.log.Info("AsMulti success", "callHash", flow.CallHashStr)
	return true
}

func (w *writer) processMultisigExecuted(m *core.Message) bool {
	flow, ok := m.Content.(*submodel.EventMultisigExecuted)
	if !ok {
		w.printContentError(m)
		return false
	}

	// _, ok = w.getBondedPools(hexutil.Encode(flow.ID[:]))
	// if !ok {
	// 	w.log.Info("received a multisigExecuted event which the ID is not in the bondedPools, ignored")
	// 	return true
	// }

	evt, ok := w.getEvents(flow.CallHashStr)
	if !ok {
		w.log.Info("receive a multisigExecuted but no evt found")
		return true
	}

	delete(evt.MulExeCallHashs, flow.CallHashStr)
	if len(evt.MulExeCallHashs) != 0 {
		w.log.Info("processMultisigExecuted wait for more callhash", "eventId", evt.EventId)
		return true
	}
	w.deleteEvents(flow.CallHashStr)
	w.deleteNewMultics(flow.CallHashStr)
	message := submodel.TransResultWithIndex{
		Symbol: w.symbol,
		Block:  0,
		Index:  0,
	}
	return w.reportTransResultWithIndex(w.symbol, &message)
}

func (w *writer) getEvents(key string) (*submodel.MultiEventFlow, bool) {
	w.eventMtx.RLock()
	defer w.eventMtx.RUnlock()
	value, exist := w.events[key]
	return value, exist
}

func (w *writer) setEvents(key string, value *submodel.MultiEventFlow) {
	w.eventMtx.Lock()
	defer w.eventMtx.Unlock()
	w.events[key] = value
}

func (w *writer) deleteEvents(key string) {
	w.eventMtx.Lock()
	defer w.eventMtx.Unlock()
	delete(w.events, key)
}

func (w *writer) getNewMultics(key string) (*submodel.EventNewMultisig, bool) {
	w.newMulTicsMtx.RLock()
	defer w.newMulTicsMtx.RUnlock()
	value, exist := w.newMultics[key]
	return value, exist
}

func (w *writer) setNewMultics(key string, value *submodel.EventNewMultisig) {
	w.newMulTicsMtx.Lock()
	defer w.newMulTicsMtx.Unlock()
	w.newMultics[key] = value
}

func (w *writer) deleteNewMultics(key string) {
	w.newMulTicsMtx.Lock()
	defer w.newMulTicsMtx.Unlock()
	delete(w.newMultics, key)
}

func (w *writer) reportTransResultWithIndex(source core.RSymbol, m *submodel.TransResultWithIndex) bool {
	msg := &core.Message{Source: source, Destination: core.RFIS, Reason: core.ReportTransResultWithIndex, Content: m}
	return w.submitWriteMessage(msg)
}

// submitMessage inserts the chainId into the msg and sends it to the router
func (w *writer) submitWriteMessage(m *core.Message) bool {
	err := w.router.SendWriteMesage(m)
	if err != nil {
		w.log.Error("failed to send message", "err", err)
		return false
	}
	return true
}
