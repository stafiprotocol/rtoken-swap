package bsc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stafiprotocol/chainbridge/utils/crypto/secp256k1"
)

type PoolClient struct {
	ethClient     *ethclient.Client
	batchTransfer *BatchTransfer
	kp            *secp256k1.Keypair
}

func NewPoolClient(ethApi, batchTransferAddress string, kp *secp256k1.Keypair) (*PoolClient, error) {
	ethClient, err := ethclient.Dial(ethApi)
	if err != nil {
		return nil, err
	}

	batchTransfer, err := NewBatchTransfer(common.HexToAddress(batchTransferAddress), ethClient)
	if err != nil {
		return nil, err
	}
	pool := PoolClient{
		ethClient:     ethClient,
		batchTransfer: batchTransfer,
		kp:            kp,
	}
	return &pool, nil
}

func (p *PoolClient) GetEthClient() *ethclient.Client {
	return p.ethClient
}

func (p *PoolClient) GetBatchTransfer() *BatchTransfer {
	return p.batchTransfer
}
