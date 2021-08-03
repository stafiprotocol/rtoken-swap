// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package stafi

import (
	"errors"
	"fmt"
	"time"

	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
	"rtoken-swap/shared/substrate"

	"github.com/ChainSafe/log15"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/stafiprotocol/chainbridge/utils/crypto/sr25519"
	"github.com/stafiprotocol/chainbridge/utils/keystore"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

var ErrNotExist = fmt.Errorf("not exist in storage")

type Connection struct {
	url    string
	symbol core.RSymbol
	sc     *substrate.SarpcClient
	gc     *substrate.GsrpcClient
	log    log15.Logger
	stop   <-chan int
}

var (
	ErrTargetNotExist  = errors.New("ErrTargetNotExist")
	BlockInterval      = 6 * time.Second
	WaitUntilFinalized = 10 * BlockInterval

	WsRetryLimit    = 240
	WsRetryInterval = 500 * time.Millisecond
)

func NewConnection(cfg *core.ChainConfig, log log15.Logger, stop <-chan int) (*Connection, error) {
	log.Info("NewConnection", "KeystorePath", cfg.KeystorePath, "Endpoint", cfg.Endpoint, "typesPath", cfg.Opts["typesPath"])

	typesPath := cfg.Opts[config.TypesPathKey]
	path, ok := typesPath.(string)
	if !ok {
		return nil, errors.New("no typesPath")
	}

	adType := cfg.Opts[config.AddressTypeKey]
	addressType, ok := adType.(string)
	if !ok {
		return nil, errors.New("addressType not ok")
	}

	subAccountInterface := cfg.Opts[config.SubAccountKey]
	subAccount, ok := subAccountInterface.(string)
	if !ok {
		return nil, errors.New("addressType not ok")
	}

	sc, err := substrate.NewSarpcClient(cfg.Name, cfg.Endpoint, path, log)
	if err != nil {
		return nil, err
	}

	kp, err := keystore.KeypairFromAddress(subAccount, keystore.SubChain, cfg.KeystorePath, cfg.Insecure)
	if err != nil {
		return nil, fmt.Errorf("keypairFromAddress err: %s", err)
	}
	krp := kp.(*sr25519.Keypair).AsKeyringPair()
	gc, err := substrate.NewGsrpcClient(cfg.Endpoint, addressType, krp, log, stop)
	if err != nil {
		return nil, fmt.Errorf("substrate.NewGsrpcClient err %s", err)
	}

	return &Connection{
		url:    cfg.Endpoint,
		symbol: cfg.Symbol,
		log:    log,
		stop:   stop,
		sc:     sc,
		gc:     gc,
	}, nil
}

func (c *Connection) GetBlockNumber(hash types.Hash) (uint64, error) {
	return c.gc.GetBlockNumber(hash)
}

func (c *Connection) LatestBlockNumber() (uint64, error) {
	return c.gc.GetLatestBlockNumber()
}

func (c *Connection) FinalizedBlockNumber() (uint64, error) {
	return c.gc.GetFinalizedBlockNumber()
}

func (c *Connection) Address() string {
	return c.gc.Address()
}

func (c *Connection) GetEvents(blockNum uint64) ([]*submodel.ChainEvent, error) {
	return c.sc.GetEvents(blockNum)
}

// queryStorage performs a storage lookup. Arguments may be nil, result must be a pointer.
func (c *Connection) QueryStorage(prefix, method string, arg1, arg2 []byte, result interface{}) (bool, error) {
	return c.gc.QueryStorage(prefix, method, arg1, arg2, result)
}

func (c *Connection) GetExtrinsics(blockhash string) ([]*submodel.Transaction, error) {
	return c.sc.GetExtrinsics(blockhash)
}

func (c *Connection) LatestMetadata() (*types.Metadata, error) {
	return c.gc.GetLatestMetadata()
}

func (c *Connection) FreeBalance(who []byte) (types.U128, error) {
	return c.gc.FreeBalance(who)
}

func (c *Connection) ExistentialDeposit() (types.U128, error) {
	return c.gc.ExistentialDeposit()
}

func (c *Connection) GetLatestDealBlock(sym core.RSymbol) (uint64, error) {
	symBz, err := types.EncodeToBytes(sym)
	if err != nil {
		return 0, err
	}

	var block uint64
	exists, err := c.QueryStorage(config.RDexnSwapModuleId, config.StorageLatestDealBLock, symBz, nil, &block)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	return block, nil
}

func (c *Connection) GetTransInfos(sym core.RSymbol, blockNumber uint64) (*submodel.TransInfoList, error) {
	key := submodel.TransInfoKey{
		Symbol: sym,
		Block:  blockNumber,
	}
	keyBz, err := types.EncodeToBytes(key)
	if err != nil {
		return nil, err
	}

	var transInfos []submodel.TransInfo
	exists, err := c.QueryStorage(config.RDexnSwapModuleId, config.StorageTransInfos, keyBz, nil, &transInfos)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrNotExist
	}
	ret := submodel.TransInfoList{
		Block: blockNumber,
		List:  transInfos,
	}

	return &ret, nil
}

func (c *Connection) GetSignature(symbol core.RSymbol, block uint64, proposalId []byte) ([]types.Bytes, error) {
	symBz, err := types.EncodeToBytes(symbol)
	if err != nil {
		return nil, err
	}

	sigkey := submodel.GetSignaturesKey{
		Block:      block,
		ProposalId: types.NewBytes(proposalId),
	}
	skBz, err := types.EncodeToBytes(sigkey)
	if err != nil {
		return nil, err
	}

	var sigs []types.Bytes
	exist, err := c.QueryStorage(config.RDexnSignaturesModuleId, config.StorageSignatures, symBz, skBz, &sigs)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, ErrNotExist
	}

	return sigs, nil
}

func (c *Connection) TransferCall(recipient []byte, value types.UCompact) (*submodel.MultiOpaqueCall, error) {
	return c.gc.TransferCall(recipient, value)
}

func (c *Connection) PaymentQueryInfo(ext string) (info *rpc.PaymentQueryInfo, err error) {
	for i := 0; i < WsRetryLimit; i++ {
		info, err = c.sc.GetPaymentQueryInfo(ext)
		if err == nil {
			return
		}

		time.Sleep(WsRetryInterval)
	}

	return
}

func (c *Connection) AsMulti(flow *submodel.MultiEventFlow) error {
	for i := 0; i < BlockRetryLimit; i++ {
		err := c.asMulti(flow)
		if err != nil {
			c.log.Warn("asmulti err will retry after 10 s", "err", err)
			time.Sleep(BlockInterval)
			continue
		} else {
			return nil
		}
	}

	return fmt.Errorf("asmulti reach limit symbol %s", flow.Symbol)
}

func (c *Connection) asMulti(flow *submodel.MultiEventFlow) error {
	gc := c.gc
	if gc == nil {
		panic(fmt.Sprintf("key disappear: %s, symbol: %s", hexutil.Encode(flow.Key.PublicKey), c.symbol))
	}

	l := len(flow.OpaqueCalls)
	if l == 1 {
		moc := flow.OpaqueCalls[0]
		ext, err := gc.NewUnsignedExtrinsic(config.MethodAsMulti, flow.Threshold, flow.Others, moc.TimePoint, moc.Opaque, false, flow.PaymentInfo.Weight)
		if err != nil {
			return err
		}

		return gc.SignAndSubmitTx(ext)
	}

	calls := make([]types.Call, 0)
	for _, oc := range flow.OpaqueCalls {
		ext, err := c.gc.NewUnsignedExtrinsic(config.MethodAsMulti, flow.Threshold, flow.Others, oc.TimePoint, oc.Opaque, false, flow.PaymentInfo.Weight)
		if err != nil {
			return err
		}

		if xt, ok := ext.(*types.Extrinsic); ok {
			calls = append(calls, xt.Method)
		} else if xt, ok := ext.(*types.ExtrinsicMulti); ok {
			calls = append(calls, xt.Method)
		}
	}

	ext, err := gc.NewUnsignedExtrinsic(config.MethodBatch, calls)
	if err != nil {
		return err
	}

	return gc.SignAndSubmitTx(ext)
}

func (c *Connection) submitSignature(param *submodel.SubmitSignatureParams) bool {
	for i := 0; i < BlockRetryLimit; i++ {
		c.log.Info("submitSignature on chain...")
		ext, err := c.gc.NewUnsignedExtrinsic(config.SubmitSignatures, param.Symbol,
			param.Block, param.ProposalId, param.Signature)
		if err != nil {
			c.log.Warn("submitSignature error will retry", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		err = c.gc.SignAndSubmitTx(ext)
		if err != nil {
			if err.Error() == ErrorTerminated.Error() {
				c.log.Error("submitSignature  met TerminatedError")
				return false
			}
			c.log.Warn("submitSignature error will retry", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		return true
	}
	return true
}

func (c *Connection) reportTransResultWithBlock(symbol core.RSymbol, block uint64) bool {
	for i := 0; i < BlockRetryLimit; i++ {
		c.log.Info("reportTransResultWithBlock on chain...")
		ext, err := c.gc.NewUnsignedExtrinsic(config.ReportTransResultWithBlock, symbol, block)
		if err != nil {
			c.log.Warn("reportTransResultWithBlock error will retry", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		err = c.gc.SignAndSubmitTx(ext)
		if err != nil {
			if err.Error() == ErrorTerminated.Error() {
				c.log.Error("reportTransResultWithBlock  met TerminatedError")
				return false
			}
			c.log.Warn("reportTransResultWithBlock error will retry", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		return true
	}
	return true
}

func (c *Connection) reportTransResultWithIndex(symbol core.RSymbol, block uint64, index uint32) bool {
	for i := 0; i < BlockRetryLimit; i++ {
		c.log.Info("reportTransResultWithIndex on chain...")
		ext, err := c.gc.NewUnsignedExtrinsic(config.ReportTransResultWithIndex, symbol, block, index)
		if err != nil {
			c.log.Warn("reportTransResultWithIndex error will retry", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		err = c.gc.SignAndSubmitTx(ext)
		if err != nil {
			if err.Error() == ErrorTerminated.Error() {
				c.log.Error("reportTransResultWithIndex  met TerminatedError")
				return false
			}
			c.log.Warn("reportTransResultWithIndex error will retry", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		return true
	}
	return true
}
