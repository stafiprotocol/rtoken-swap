// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package substrate

import (
	"fmt"
	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"

	"github.com/ChainSafe/log15"
	subTypes "github.com/stafiprotocol/go-substrate-rpc-client/types"
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
	go w.resolveReadMessage(m)
}

func (w *writer) resolveWriteMessage(m *core.Message) (processOk bool) {
	defer func() {
		if !processOk {
			panic(fmt.Sprintf("resolveMessage process failed. %+v", m))
		}
	}()

	switch m.Reason {
	case core.ReportTransResultWithBlock:
		processOk = w.processReportResultWithBlock(m)
	case core.ReportTransResultWithIndex:
		processOk = w.processReportResultWithIndex(m)
	case core.SubmitSignature:
		processOk = w.processSubmitSignature(m)
	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return true
	}
	return
}

func (w *writer) resolveReadMessage(m *core.Message) bool {
	switch m.Reason {
	case core.GetLatestDealBLock:
		w.processGetLatestDealBlock(m)
	case core.GetSignatures:
		w.processGetSignatures(m)
	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return false
	}
	return false
}

func (w *writer) processGetLatestDealBlock(m *core.Message) bool {
	getLatestBlockParam, ok := m.Content.(*submodel.GetLatestDealBLockParam)
	if !ok {
		w.printContentError(m)
		return false
	}

	dealBlock, err := w.conn.GetLatestDealBlock(getLatestBlockParam.Symbol)
	if err != nil {
		w.log.Warn("GetLatestDealBlock failed", "err", err)
		if err == ErrNotExist {
			getLatestBlockParam.Block <- 0
			return true
		}
	}
	getLatestBlockParam.Block <- dealBlock
	return true
}

func (w *writer) processGetSignatures(m *core.Message) bool {
	getSignatureParam, ok := m.Content.(*submodel.GetSignaturesParam)
	if !ok {
		w.printContentError(m)
		return false
	}

	signatures, err := w.conn.GetSignature(getSignatureParam.Symbol, getSignatureParam.Block, getSignatureParam.ProposalId)
	if err != nil {
		w.log.Warn("GetSignature failed", "err", err)
		if err == ErrNotExist {
			getSignatureParam.Signatures <- []subTypes.Bytes{}
			return true
		}
	}
	getSignatureParam.Signatures <- signatures
	return true
}

func (w *writer) processSubmitSignature(m *core.Message) bool {
	param, ok := m.Content.(*submodel.SubmitSignatureParams)
	if !ok {
		w.printContentError(m)
		return false
	}
	result := w.conn.submitSignature(param)
	w.log.Info("submitSignature", "symbol", m.Source, "result", result)
	return result
}

func (w *writer) processReportResultWithBlock(m *core.Message) bool {
	transResult, ok := m.Content.(*submodel.TransResultWithBlock)
	if !ok {
		w.printContentError(m)
		return false
	}
	return w.conn.reportTransResultWithBlock(transResult.Symbol, transResult.Block)
}

func (w *writer) processReportResultWithIndex(m *core.Message) bool {
	transResult, ok := m.Content.(*submodel.TransResultWithIndex)
	if !ok {
		w.printContentError(m)
		return false
	}
	return w.conn.reportTransResultWithIndex(transResult.Symbol, transResult.Block, transResult.Index)
}

func (w *writer) printContentError(m *core.Message) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason)
}
