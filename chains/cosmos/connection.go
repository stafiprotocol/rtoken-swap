package cosmos

import (
	"errors"
	"fmt"
	"os"
	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/shared/cosmos"
	"rtoken-swap/shared/cosmos/rpc"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	BlockRetryLimit    = 50
	BlockRetryInterval = time.Second * 6
	BlockConfirmNumber = 10
)

type Connection struct {
	url             string
	symbol          core.RSymbol
	currentHeight   int64
	threshold       int
	poolClient      *cosmos.PoolClient
	multisigAddress types.AccAddress
	log             log15.Logger
	stop            <-chan int
}

func NewConnection(cfg *core.ChainConfig, log log15.Logger, stop <-chan int) (*Connection, error) {

	multisigAccouont, ok := cfg.Opts[config.MultisigAccountKey].(string)
	if !ok || len(multisigAccouont) == 0 {
		return nil, fmt.Errorf("no multisig account")
	}

	subAccount, ok := cfg.Opts[config.SubAccountKey].(string)
	if !ok || len(multisigAccouont) == 0 {
		return nil, fmt.Errorf("no sub account")
	}

	chainId, ok := cfg.Opts[config.ChainIdKey].(string)
	if !ok || len(chainId) == 0 {
		return nil, errors.New("config must has chainId")
	}
	denom, ok := cfg.Opts[config.DenomKey].(string)
	if !ok || len(chainId) == 0 {
		return nil, errors.New("config must has denom")
	}
	gasPrice, ok := cfg.Opts[config.GasPriceKey].(string)
	if !ok || len(gasPrice) == 0 {
		return nil, errors.New("config must has gasPrice")
	}
	thresholdStr, ok := cfg.Opts[config.ThresholdKey].(string)
	if !ok {
		return nil, errors.New("config must has threshold")
	}
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Will open cosmos wallet from <%s>. \nPlease ", cfg.KeystorePath)
	key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, cfg.KeystorePath, os.Stdin)
	if err != nil {
		return nil, err
	}
	multisigKey, err := key.Key(multisigAccouont)
	if err != nil {
		return nil, err
	}
	multisigAddr := multisigKey.GetAddress()

	client, err := rpc.NewClient(key, chainId, multisigAccouont, gasPrice, denom, cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	poolClient := cosmos.NewPoolClient(log, client, subAccount, 0)

	return &Connection{
		url:             cfg.Endpoint,
		symbol:          cfg.Symbol,
		log:             log,
		stop:            stop,
		multisigAddress: multisigAddr,
		poolClient:      poolClient,
		threshold:       threshold,
	}, nil
}

func (c *Connection) GetTx(poolClient *cosmos.PoolClient, txHash string) (*types.TxResponse, error) {
	var txRes *types.TxResponse
	var err error
	retryTx := 0
	for {
		if retryTx >= BlockRetryLimit {
			return nil, errors.New("QueryTxByHash reach retry limit")
		}
		txRes, err = poolClient.GetRpcClient().QueryTxByHash(txHash)
		if err != nil {
			c.log.Warn(fmt.Sprintf("QueryTxByHash err: %s ,will retry queryTx after %f second", err, BlockRetryInterval.Seconds()))
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		}
		currentHeight := atomic.LoadInt64(&c.currentHeight)
		if txRes.Height+BlockConfirmNumber > currentHeight {
			c.log.Warn(fmt.Sprintf("confirm number is smaller than %d ,will retry queryTx after %f second", BlockConfirmNumber, BlockRetryInterval.Seconds()))
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		} else {
			break
		}

	}
	return txRes, nil
}

func (c *Connection) GetBlock(poolClient *cosmos.PoolClient, height int64) (*ctypes.ResultBlock, error) {
	var blockRes *ctypes.ResultBlock
	var err error
	retryTx := 0
	for {
		if retryTx >= BlockRetryLimit {
			return nil, errors.New("QueryBlock reach retry limit")
		}
		blockRes, err = poolClient.GetRpcClient().QueryBlock(height)
		if err != nil {
			c.log.Warn(fmt.Sprintf("QueryBlock err: %s ,will retry queryBlock after %f second", err, BlockRetryInterval.Seconds()))
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		}
		break
	}
	return blockRes, nil
}

func (c *Connection) GetPoolClient() *cosmos.PoolClient {
	return c.poolClient
}

func (c *Connection) GetMultisigAddress() types.AccAddress {
	return c.multisigAddress
}
