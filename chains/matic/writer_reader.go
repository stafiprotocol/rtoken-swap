package matic

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"

	"github.com/ChainSafe/log15"
	"github.com/ethereum/go-ethereum/common"
)

const msgLimit = 4096

var baseBig = big.NewInt(1)

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
			w.log.Info("matic msgHandler stop")
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
	if transInfoList.DestSymbol != core.RMATIC {
		w.printContentError(m, errors.New("traninfo dest symbol != RBNB"))
		return false
	}
	w.log.Info("processNewTransInfos", "transInfo", transInfoList)
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
	poolClient := w.conn.GetPoolClient()
	batchTransfer := poolClient.GetBatchTransfer()
	txOpts, err := poolClient.GetTransactionOpts()
	if err != nil {
		return false
	}
	callOpts := poolClient.GetCallOpts()

	ethClient := poolClient.GetEthClient()
	if err != nil {
		w.log.Error("poolClient.GetTransactionOpts failed", "err", err)
		return false
	}
	tos := make([]common.Address, 0)
	values := make([]*big.Int, 0)
	for _, l := range transInfoList.List {
		tos = append(tos, common.BytesToAddress(l.Receiver))
		values = append(values, new(big.Int).Mul(l.Value.Int, baseBig))
	}
	block := big.NewInt(int64(transInfoList.Block))
	tx, err := batchTransfer.BatchTransfer(txOpts, block, tos, values)
	//todo check already exe err
	if err != nil {
		w.log.Error("batchTransfer.BatchTransfer failed", "err", err)
		return false
	}
	w.log.Info("send batchTransfer", "gasPrice", tx.GasPrice().String(), "nonce", tx.Nonce(), "txHash", tx.Hash(), "gas", tx.Gas())
	//check is confirmed
	retry := 0
	for {
		if retry > BlockRetryLimit {
			w.log.Error("check BatchTransfer tx reach retry", "tx", tx.Hash(), "fromAddress", poolClient.GetFromAddress().String())
			return false
		}
		_, isPending, err := ethClient.TransactionByHash(context.Background(), tx.Hash())
		if err == nil && !isPending {
			break
		} else {
			w.log.Warn("check BatchTransfer tx failed ,watting...", " isPending ", isPending, " err ", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
	}

	phash := GetProposalHash(block, tos, values)
	retry = 0
	for {
		if retry > BlockRetryLimit {
			w.log.Error("check proposal reach retry", "proposal", hex.EncodeToString(phash[:]), "fromAddress", poolClient.GetFromAddress().String())
			return false
		}

		proposal, err := batchTransfer.Proposals(callOpts, phash)
		if err != nil {
			w.log.Warn("check proposal failed ,watting...", " err ", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		if proposal.Status != 2 {
			w.log.Warn("check proposal not exe yet ,watting...", "status ", proposal.Status)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	w.log.Info("check proposal exe success", "block", transInfoList.Block, "transInfoList", transInfoListToStr(transInfoList))
	report := submodel.TransResultWithBlock{
		Symbol: core.RMATIC,
		Block:  transInfoList.Block,
	}
	return w.reportTransResultWithBlock(core.RMATIC, core.FIS, &report)
}

func (w *writer) getLatestDealBLock(symbol core.RSymbol) (uint64, error) {
	getLatestDealBlockParam := submodel.GetLatestDealBLockParam{
		Symbol: symbol,
		Block:  make(chan uint64, 1),
	}
	m := &core.Message{Source: core.RMATIC, Destination: core.FIS, Reason: core.GetLatestDealBLock, Content: &getLatestDealBlockParam}
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

func transInfoListToStr(transInfoList *submodel.TransInfoList) string {
	ret := ""
	for _, b := range transInfoList.List {
		line := fmt.Sprintf("account: %s reciever: %s value: %s\n",
			hex.EncodeToString(b.Account[:]), hex.EncodeToString(b.Receiver), b.Value.String())
		ret += line
	}
	return ret
}
