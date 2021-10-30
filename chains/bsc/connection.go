package bsc

import (
	"errors"
	"fmt"
	"github.com/stafiprotocol/chainbridge/utils/crypto/secp256k1"
	"github.com/stafiprotocol/chainbridge/utils/keystore"
	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/shared/bsc"
	"time"

	"github.com/ChainSafe/log15"
)

const (
	BlockRetryLimit    = 100
	BlockRetryInterval = time.Second * 6
	BlockConfirmNumber = 10
)

type Connection struct {
	url        string
	symbol     core.RSymbol
	poolClient *bsc.PoolClient
	log        log15.Logger
	stop       <-chan int
}

func NewConnection(cfg *core.ChainConfig, log log15.Logger, stop <-chan int) (*Connection, error) {

	subAccount, ok := cfg.Opts[config.SubAccountKey].(string)
	if !ok || len(subAccount) == 0 {
		return nil, fmt.Errorf("no sub account")
	}
	bscEndpoint, ok := cfg.Opts[config.BscEndpointKey].(string)
	if !ok || len(bscEndpoint) == 0 {
		return nil, fmt.Errorf("bsc endpoint")
	}
	batchTransferAddress, ok := cfg.Opts[config.BatchTransferAddressKey].(string)
	if !ok || len(batchTransferAddress) == 0 {
		return nil, fmt.Errorf("no batchTransfer")
	}

	chainId, ok := cfg.Opts[config.ChainIdKey].(int64)
	if !ok || chainId == 0 {
		return nil, errors.New("config must has chainId")
	}

	maxGasPrice, ok := cfg.Opts[config.MaxGasPriceKey].(int64)
	if !ok || maxGasPrice == 0 {
		return nil, errors.New("config must has maxGasPrice")
	}

	gasLimit, ok := cfg.Opts[config.GasLimitKey].(int64)
	if !ok || gasLimit == 0 {
		return nil, errors.New("config must has gasLimit")
	}

	fmt.Printf("Will open bsc wallet from <%s>. \nPlease ", cfg.KeystorePath)
	kpI, err := keystore.KeypairFromAddress(subAccount, keystore.EthChain, cfg.KeystorePath, cfg.Insecure)
	if err != nil {
		return nil, err
	}
	kp, _ := kpI.(*secp256k1.Keypair)
	poolClient, err := bsc.NewPoolClient(bscEndpoint, batchTransferAddress, kp, maxGasPrice, gasLimit, chainId)
	if err != nil {
		return nil, err
	}
	return &Connection{
		url:        cfg.Endpoint,
		symbol:     cfg.Symbol,
		log:        log,
		stop:       stop,
		poolClient: poolClient,
	}, nil
}

func (c *Connection) GetPoolClient() *bsc.PoolClient {
	return c.poolClient
}
