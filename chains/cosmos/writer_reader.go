package cosmos

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
	"rtoken-swap/utils"

	"github.com/ChainSafe/log15"
	substrateTypes "github.com/stafiprotocol/go-substrate-rpc-client/types"
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
			w.log.Info("cosmos msgHandler stop")
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
	client := poolClient.GetRpcClient()
	multisigAddr := w.conn.GetMultisigAddress()
	threshold := w.conn.threshold
	unSignedTx, outPuts, err := GetTransferUnsignedTxV2(client, multisigAddr, transInfoList.List, w.log)
	if err != nil {
		w.log.Error("GetTransferUnsignedTx failed", "err", err)
		return false
	}
	w.log.Info("new transInfo", "transinfo", outPuts)
	//use current seq
	seq, err := client.GetSequence(0, multisigAddr)
	if err != nil {
		w.log.Error("GetSequence failed",
			"err", err)
		return false
	}

	sigBts, err := client.SignMultiSigRawTxWithSeq(seq, unSignedTx, poolClient.GetSubKeyName())
	if err != nil {
		w.log.Error("processNewTransInfos SignMultiSigRawTx failed",
			"unsignedTx", string(unSignedTx),
			"err", err)
		return false
	}

	proposalId := GetTransferProposalId(utils.BlakeTwo256(unSignedTx))
	proposalIdHexStr := hex.EncodeToString(proposalId)

	param := submodel.SubmitSignatureParams{
		Symbol:     w.conn.symbol,
		Block:      substrateTypes.NewU64(transInfoList.Block),
		ProposalId: substrateTypes.NewBytes(proposalId),
		Signature:  substrateTypes.NewBytes(sigBts),
	}
	result := &core.Message{Source: core.RATOM, Destination: core.RFIS, Reason: core.SubmitSignature, Content: &param}
	subSignatureOk := w.submitWriteMessage(result)
	if !subSignatureOk {
		w.log.Error("processNewTransInfos SignMultiSigRawTx failed",
			"unsignedTx", string(unSignedTx),
			"proposalId", proposalIdHexStr,
			"err", err)
		return false
	}

	w.log.Info("processNewTransInfos submitSignature",
		"unsignedTx", string(unSignedTx),
		"proposalId", proposalIdHexStr)
	var sigs [][]byte
	for {
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
		sigs, err = w.getSubmitSignature(w.conn.symbol, transInfoList.Block, proposalId)
		if err != nil {
			w.log.Warn("getSubmitSignature failed", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		if len(sigs) < int(threshold) {
			w.log.Warn("getSubmitSignature sigs not enough yet", "num", len(sigs), "need", threshold)
			time.Sleep(BlockRetryInterval)
			continue
		}
		break
	}

	txHash, txBts, err := client.AssembleMultiSigTx(unSignedTx, sigs, uint32(threshold))
	if err != nil {
		w.log.Error("processSignatureEnoughEvt AssembleMultiSigTx failed",
			"unsignedTx", hex.EncodeToString(unSignedTx),
			"signatures", bytesArrayToStr(sigs),
			"threshold", threshold,
			"err", err)
		return false
	}

	return w.checkAndSend(poolClient, txHash, txBts, transInfoList.Block)
}

func (w *writer) getSubmitSignature(symbol core.RSymbol, block uint64, proposalId []byte) ([][]byte, error) {
	getSigsParam := submodel.GetSignaturesParam{
		Symbol:     symbol,
		Block:      block,
		ProposalId: proposalId,
		Signatures: make(chan []substrateTypes.Bytes, 1),
	}
	m := &core.Message{Source: core.RATOM, Destination: core.RFIS, Reason: core.GetSignatures, Content: &getSigsParam}
	subOk := w.submitReadMessage(m)
	if !subOk {
		return nil, fmt.Errorf("submitMessage err")
	}

	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		return nil, fmt.Errorf("time out")
	case sigs := <-getSigsParam.Signatures:
		ret := make([][]byte, 0)
		for _, sig := range sigs {
			ret = append(ret, []byte(sig))
		}
		return ret, nil
	}
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
