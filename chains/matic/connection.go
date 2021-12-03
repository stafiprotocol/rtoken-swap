package matic

import (
	"errors"
	"fmt"
	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/shared/matic"
	"time"

	"github.com/stafiprotocol/chainbridge/utils/crypto/secp256k1"
	"github.com/stafiprotocol/chainbridge/utils/keystore"

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
	poolClient *matic.PoolClient
	log        log15.Logger
	stop       <-chan int
	threshold  int
}

func NewConnection(cfg *core.ChainConfig, log log15.Logger, stop <-chan int) (*Connection, error) {

	subAccount, ok := cfg.Opts[config.SubAccountKey].(string)
	if !ok || len(subAccount) == 0 {
		return nil, fmt.Errorf("no sub account")
	}

	batchTransferAddress, ok := cfg.Opts[config.BatchTransferAddressKey].(string)
	if !ok || len(batchTransferAddress) == 0 {
		return nil, fmt.Errorf("no batchTransfer")
	}
	threshold, ok := cfg.Opts[config.ThresholdKey].(float64)
	if !ok || threshold == 0 {
		return nil, errors.New("config must has threshold")
	}

	maxGasPrice, ok := cfg.Opts[config.MaxGasPriceKey].(float64)
	if !ok || maxGasPrice == 0 {
		return nil, errors.New("config must has maxGasPrice")
	}

	fmt.Printf("Will open matic wallet from <%s>. \nPlease ", cfg.KeystorePath)
	kpI, err := keystore.KeypairFromAddress(subAccount, keystore.EthChain, cfg.KeystorePath, cfg.Insecure)
	if err != nil {
		return nil, err
	}
	kp, _ := kpI.(*secp256k1.Keypair)
	poolClient, err := matic.NewPoolClient(cfg.Endpoint, batchTransferAddress, kp, int64(maxGasPrice))
	if err != nil {
		return nil, err
	}
	return &Connection{
		url:        cfg.Endpoint,
		symbol:     cfg.Symbol,
		log:        log,
		stop:       stop,
		poolClient: poolClient,
		threshold:  int(threshold),
	}, nil
}

func (c *Connection) GetPoolClient() *matic.PoolClient {
	return c.poolClient
}
