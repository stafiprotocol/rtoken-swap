package bsc

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stafiprotocol/chainbridge/utils/crypto/secp256k1"
)

type PoolClient struct {
	ethClient     *ethclient.Client
	batchTransfer *BatchTransfer
	kp            *secp256k1.Keypair
	fromAddress   common.Address
	maxGasPrice   int64 //gwei
	chainId       int64
}

func NewPoolClient(ethApi, batchTransferAddress string, kp *secp256k1.Keypair, maxGasPrice, chainId int64) (*PoolClient, error) {
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
		maxGasPrice:   maxGasPrice,
		chainId:       chainId,
		fromAddress:   kp.CommonAddress(),
	}
	return &pool, nil
}

func (p *PoolClient) GetEthClient() *ethclient.Client {
	return p.ethClient
}
func (p *PoolClient) GetFromAddress() common.Address {
	return p.fromAddress
}

func (p *PoolClient) GetBatchTransfer() *BatchTransfer {
	return p.batchTransfer
}

func (p *PoolClient) GetTransactionOpts() (*bind.TransactOpts, error) {
	suggestGasPrice, err := p.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	if suggestGasPrice.Cmp(big.NewInt(p.maxGasPrice*1e9)) > 0 {
		suggestGasPrice = big.NewInt(p.maxGasPrice * 1e9)
	}

	opts := bind.TransactOpts{
		From: p.fromAddress,
		Signer: func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return signTx(tx, p.kp.PrivateKey(), big.NewInt(p.chainId))
		},
		GasPrice: suggestGasPrice,
		Context:  context.Background(),
	}

	return &opts, nil
}

func (p *PoolClient) GetCallOpts() *bind.CallOpts {
	callOpts := bind.CallOpts{
		Pending:     false,
		From:        p.fromAddress,
		BlockNumber: nil,
		Context:     context.Background(),
	}
	return &callOpts
}

func signTx(rawTx *types.Transaction, privateKey *ecdsa.PrivateKey, chainId *big.Int) (signedTx *types.Transaction, err error) {
	// Sign the transaction and verify the sender to avoid hardware fault surprises
	signedTx, err = types.SignTx(rawTx, types.NewEIP155Signer(chainId), privateKey)
	return
}
