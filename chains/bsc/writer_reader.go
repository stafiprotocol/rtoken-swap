package bsc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ChainSafe/log15"
	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
)

const msgLimit = 4096

//write to cosmos
type writer struct {
	conn    *Connection
	router  chains.Router
	log     log15.Logger
	msgChan chan *core.Message
	sysErr  chan<- error
	stop    chan int
}

func NewWriter(conn *Connection, log log15.Logger, sysErr chan<- error) *writer {
	return &writer{
		conn:    conn,
		log:     log,
		sysErr:  sysErr,
		msgChan: make(chan *core.Message, msgLimit),
	}
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
			w.log.Info("bsc msgHandler stop")
			break out
		case msg := <-w.msgChan:
			w.resolveWriteMessage(msg)
		}
	}
	return nil
}

func (w *writer) setRouter(r chains.Router) {
	w.router = r
}

func (w *writer) QueueWriteMessage(m *core.Message) {
	w.msgChan <- m
}

func (w *writer) ResolveReadMessage(m *core.Message) {
}

//resolve msg from other chains
func (w *writer) resolveWriteMessage(m *core.Message) (processOk bool) {
	defer func() {
		if !processOk {
			panic(fmt.Sprintf("resolveMessage process failed. %+v", m))
		}
	}()

	switch m.Reason {
	case core.NewTransInfos:
		return w.processNewTransInfos(m)
	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return true
	}
}

func (w *writer) processNewTransInfos(m *core.Message) bool {
	transInfoList, ok := m.Content.(*submodel.TransInfoList)
	if !ok {
		w.printContentError(m, errors.New("msg cast to traninfo not ok"))
		return false
	}
	if transInfoList.DestSymbol != core.RATOM {
		w.printContentError(m, errors.New("traninfo dest symbol != RATOM"))
		return false
	}
	// check block is deal
	isDeal, err := w.checkDeal(transInfoList.Block)
	if err != nil {
		w.log.Error("checkDeal failed", "err", err)
		return false
	}
	if isDeal {
		w.log.Info("block has deal ", "block", transInfoList.Block)
		return true
	}
	w.conn.GetPoolClient()
	return true

}

func (w *writer) getLatestDealBLock(symbol core.RSymbol) (uint64, error) {
	getLatestDealBlockParam := submodel.GetLatestDealBLockParam{
		Symbol: symbol,
		Block:  make(chan uint64, 1),
	}
	m := &core.Message{Source: core.RATOM, Destination: core.RFIS, Reason: core.GetLatestDealBLock, Content: &getLatestDealBlockParam}
	subOk := w.submitReadMessage(m)
	if !subOk {
		return 0, fmt.Errorf("submitMessage err")
	}

	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		return 0, fmt.Errorf("time out")
	case block := <-getLatestDealBlockParam.Block:
		return block, nil
	}
}

func (w *writer) checkDeal(block uint64) (bool, error) {
	var latestDealBlock uint64
	var err error
	retry := 0
	for {
		if retry > BlockRetryLimit {
			return false, fmt.Errorf("getLatestDealBLock reach retry limit")
		}
		latestDealBlock, err = w.getLatestDealBLock(w.conn.symbol)
		if err != nil {
			w.log.Warn("getLatestDealBLock failed", "err", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}
	return latestDealBlock >= block, nil
}

func bytesArrayToStr(bts [][]byte) string {
	ret := ""
	for _, b := range bts {
		ret += " | "
		ret += hex.EncodeToString(b)
	}
	return ret
}
