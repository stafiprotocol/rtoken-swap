package cosmos

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
	"rtoken-swap/shared/cosmos"
	"rtoken-swap/shared/cosmos/rpc"
	"rtoken-swap/utils"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/types"
	errType "github.com/cosmos/cosmos-sdk/types/errors"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	substrateTypes "github.com/stafiprotocol/go-substrate-rpc-client/types"
)

var ErrNoOutPuts = errors.New("outputs length is zero")

func GetTransferProposalId(txHash [32]byte) []byte {
	proposalId := make([]byte, 32)
	copy(proposalId, txHash[:])
	return proposalId
}

func ParseTransferProposalId(content []byte) (shotId substrateTypes.Hash, err error) {
	if len(content) != 32 {
		err = errors.New("cont length is not right")
		return
	}
	shotId = substrateTypes.NewHash(content)
	return
}

func GetValidatorUpdateProposalId(content []byte) []byte {
	hash := utils.BlakeTwo256(content)
	return hash[:]
}

func GetTransferUnsignedTxV2(client *rpc.Client, multisigAddress types.AccAddress, receives []submodel.TransInfo,
	logger log15.Logger) ([]byte, []xBankTypes.Output, error) {

	outPuts := make([]xBankTypes.Output, 0)
	for _, receive := range receives {
		hexAccountStr := hex.EncodeToString(receive.Receiver[:20])
		addr, err := types.AccAddressFromHex(hexAccountStr)
		if err != nil {
			logger.Error("GetTransferUnsignedTx AccAddressFromHex failed", "hexAccount", hexAccountStr, "err", err)
			continue
		}
		valueBigInt := receive.Value.Int
		out := xBankTypes.Output{
			Address: addr.String(),
			Coins:   types.NewCoins(types.NewCoin(client.GetDenom(), types.NewIntFromBigInt(valueBigInt))),
		}
		outPuts = append(outPuts, out)
	}

	//len should not be 0
	if len(outPuts) == 0 {
		return nil, nil, ErrNoOutPuts
	}

	//sort outPuts for the same rawTx from different relayer
	sort.SliceStable(outPuts, func(i, j int) bool {
		return bytes.Compare([]byte(outPuts[i].Address), []byte(outPuts[j].Address)) < 0
	})

	txBts, err := client.GenMultiSigRawBatchTransferTx(multisigAddress, outPuts)
	if err != nil {
		return nil, nil, ErrNoOutPuts
	}
	return txBts, outPuts, nil
}

func (w *writer) printContentError(m *core.Message, err error) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason, "err", err)
}

// submitMessage inserts the chainId into the msg and sends it to the router
func (w *writer) submitWriteMessage(m *core.Message) bool {
	err := w.router.SendWriteMesage(m)
	if err != nil {
		w.log.Error("failed to SendWriteMesage", "err", err)
		return false
	}
	return true
}

func (w *writer) submitReadMessage(m *core.Message) bool {
	err := w.router.SendReadMesage(m)
	if err != nil {
		w.log.Error("failed to SendReadMesage", "err", err)
		return false
	}
	return true
}

func (w *writer) reportTransResultWithBlock(source, dest core.RSymbol, m *submodel.TransResultWithBlock) bool {
	msg := &core.Message{Source: source, Destination: dest, Reason: core.ReportTransResultWithBlock, Content: m}
	return w.submitWriteMessage(msg)
}

func (w *writer) checkAndSend(poolClient *cosmos.PoolClient, txHash, txBts []byte, block uint64) bool {
	retry := 0
	txHashHexStr := hex.EncodeToString(txHash)
	client := poolClient.GetRpcClient()

	for {
		if retry > BlockRetryLimit {
			w.log.Error("checkAndSend broadcast tx reach retry limit",
				"pool hex address", w.conn.GetMultisigAddress())
			break
		}
		//check on chain
		res, err := client.QueryTxByHash(txHashHexStr)
		if err != nil || res.Empty() || res.Code != 0 {
			w.log.Warn(fmt.Sprintf(
				"checkAndSend QueryTxByHash failed. will rebroadcast after %f second",
				BlockRetryInterval.Seconds()),
				"tx hash", txHashHexStr,
				"err or res.empty", err)

			//broadcast if not on chain
			_, err = client.BroadcastTx(txBts)
			if err != nil && err != errType.ErrTxInMempoolCache {
				w.log.Warn("checkAndSend BroadcastTx failed  will retry",
					"err", err)
			}
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}

		w.log.Info("checkAndSend success",
			"txHash", txHashHexStr)
		report := submodel.TransResultWithBlock{
			Symbol: core.RATOM,
			Block:  block,
		}
		return w.reportTransResultWithBlock(core.RATOM, core.RFISX, &report)

	}
	return false
}
