package solana

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"rtoken-swap/chains"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"

	"github.com/ChainSafe/log15"
	solClient "github.com/stafiprotocol/solana-go-sdk/client"
	solCommon "github.com/stafiprotocol/solana-go-sdk/common"
	"github.com/stafiprotocol/solana-go-sdk/multisigprog"
	"github.com/stafiprotocol/solana-go-sdk/sysprog"
	solTypes "github.com/stafiprotocol/solana-go-sdk/types"
)

const (
	msgLimit              = 4096
	multisigSendMaxNumber = 5
)

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
		return w.processNewTransInfoList(m)
	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return true
	}
}

func (w *writer) processNewTransInfoList(m *core.Message) bool {
	transInfoList, ok := m.Content.(*submodel.TransInfoList)
	if !ok {
		w.printContentError(m, errors.New("msg cast to transInfoList not ok"))
		return false
	}
	if transInfoList.DestSymbol != core.RSOL {
		w.log.Error("transinfo dest symbol != w.symbol", "destsymbol", transInfoList.DestSymbol, "w.symbol", w.conn.symbol)
		return false
	}
	w.log.Info("processNewTransInfoList", "transInfo", transInfoList)

	if len(transInfoList.List) == 0 {
		w.log.Error("processNewTransInfoList list len is zero")
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

	//sort outPuts for the same rawTx from different relayer
	sort.SliceStable(transInfoList.List, func(i, j int) bool {
		return bytes.Compare(transInfoList.List[i].Account[:], transInfoList.List[j].Account[:]) < 0
	})

	poolClient, err := w.conn.GetPoolClient()
	if err != nil {
		w.log.Error("processWithdrawReportedEvent failed",
			"error", err)
		return false
	}
	poolAddrBase58Str := poolClient.MultisignerPubkey.ToBase58()
	rpcClient := poolClient.GetRpcClient()

	for i := 0; i <= len(transInfoList.List)/multisigSendMaxNumber; i++ {
		multisigTxAccountPubkey, multisigTxAccountSeed := GetMultisigTxAccountPubkeyForTransfer(
			poolClient.MultisigTxBaseAccountPubkey,
			poolClient.MultisigProgramId,
			uint32(transInfoList.Block),
			i)

		transferInstructions := make([]solTypes.Instruction, 0)
		programIds := make([]solCommon.PublicKey, 0)
		accountMetas := make([][]solTypes.AccountMeta, 0)
		txDatas := make([][]byte, 0)

		for j := 0; j < multisigSendMaxNumber; j++ {
			index := i*multisigSendMaxNumber + j
			//check overflow
			if index > len(transInfoList.List)-1 {
				break
			}
			receive := transInfoList.List[index]
			to := solCommon.PublicKeyFromBytes(receive.Account[:])
			value := receive.Value.Int
			transferInstruction := sysprog.Transfer(poolClient.MultisignerPubkey, to, value.Uint64())
			transferInstructions = append(transferInstructions, transferInstruction)

			programIds = append(programIds, transferInstruction.ProgramID)
			accountMetas = append(accountMetas, transferInstruction.Accounts)
			txDatas = append(txDatas, transferInstruction.Data)
			w.log.Info("will transfer to ", "index ", index, " addr ", to.ToBase58(), " value ", value.Int64())
		}
		remainingAccounts := multisigprog.GetRemainAccounts(transferInstructions)

		if poolClient.HasBaseAccountAuth {
			if poolClient.MultisigTxBaseAccount == nil {
				w.log.Error("MultisigTxBaseAccount privkey not exist", "MultisigTxBaseAccount", poolClient.MultisigTxBaseAccountPubkey)
				return false
			}

			_, err = rpcClient.GetMultisigTxAccountInfo(context.Background(), multisigTxAccountPubkey.ToBase58())
			if err != nil && err == solClient.ErrAccountNotFound {
				sendOk := w.createMultisigTxAccount(rpcClient, poolClient, poolAddrBase58Str, programIds, accountMetas, txDatas,
					multisigTxAccountPubkey, multisigTxAccountSeed, "processWithdrawReportedEvent")
				if !sendOk {
					return false
				}
			}

			if err != nil && err != solClient.ErrAccountNotFound {
				w.log.Error("processWithdrawReportedEvent GetMultisigTxAccountInfo err",
					"pool  address", poolAddrBase58Str,
					"multisig tx account address", multisigTxAccountPubkey.ToBase58(),
					"err", err)
				return false
			}
		}
		//check multisig tx account is created
		create := w.waitingForMultisigTxCreate(rpcClient, poolAddrBase58Str, multisigTxAccountPubkey.ToBase58(), "processWithdrawReportedEvent")
		if !create {
			return false
		}
		w.log.Info("processWithdrawReportedEvent multisigTxAccount has create", "multisigTxAccount", multisigTxAccountPubkey.ToBase58())

		valid := w.CheckMultisigTx(rpcClient, multisigTxAccountPubkey, programIds, accountMetas, txDatas)
		if !valid {
			w.log.Info("processWithdrawReportedEvent CheckMultisigTx failed", "multisigTxAccount", multisigTxAccountPubkey.ToBase58())
			return false
		}

		//if has exe just continue
		isExe := w.IsMultisigTxExe(rpcClient, multisigTxAccountPubkey)
		if isExe {
			w.log.Info("processWithdrawReportedEvent multisigTxAccount has execute", "multisigTxAccount", multisigTxAccountPubkey.ToBase58())
			continue
		}
		//approve tx
		send := w.approveMultisigTx(rpcClient, poolClient, poolAddrBase58Str, multisigTxAccountPubkey, remainingAccounts, "processWithdrawReportedEvent")
		if !send {
			return false
		}

		//check multisig exe result
		exe := w.waitingForMultisigTxExe(rpcClient, poolAddrBase58Str, multisigTxAccountPubkey.ToBase58(), "processWithdrawReportedEvent")
		if !exe {
			return false
		}
		w.log.Info("processWithdrawReportedEvent multisigTxAccount has execute", "multisigTxAccount", multisigTxAccountPubkey.ToBase58(), "block", transInfoList.Block)
	}

	w.log.Info("check exe success", "block", transInfoList.Block, "transInfoList", transInfoListToStr(transInfoList))
	report := submodel.TransResultWithBlock{
		Symbol: core.RSOL,
		Block:  transInfoList.Block,
	}
	return w.reportTransResultWithBlock(core.RSOL, core.FIS, &report)
}

func (w *writer) reportTransResultWithBlock(source, dest core.RSymbol, m *submodel.TransResultWithBlock) bool {
	msg := &core.Message{Source: source, Destination: dest, Reason: core.ReportTransResultWithBlock, Content: m}
	return w.submitWriteMessage(msg)
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
