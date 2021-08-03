package cosmos

import (
	"rtoken-swap/shared/cosmos/rpc"
	"sync"

	"github.com/ChainSafe/log15"
)

//one pool address with one poolClient
type PoolClient struct {
	eraBlockNumber int64
	log            log15.Logger
	rpcClient      *rpc.Client
	subKeyName     string //subKey is one of pubKeys of multiSig pool address,subKeyName is subKey`s name in keyring
	mtx            sync.RWMutex
}

func NewPoolClient(log log15.Logger, rpcClient *rpc.Client, subKey string, eraBLockNumber int64) *PoolClient {
	return &PoolClient{
		eraBlockNumber: eraBLockNumber,
		log:            log,
		rpcClient:      rpcClient,
		subKeyName:     subKey}
}

func (pc *PoolClient) GetRpcClient() *rpc.Client {
	return pc.rpcClient
}

func (pc *PoolClient) GetSubKeyName() string {
	return pc.subKeyName
}

func (pc *PoolClient) GetHeightByEra(era uint32) int64 {
	return int64(era) * pc.eraBlockNumber
}

func (pc *PoolClient) GetCurrentEra() (int64, uint32, error) {
	height, err := pc.GetRpcClient().GetCurrentBLockHeight()
	if err != nil {
		return 0, 0, err
	}
	if pc.eraBlockNumber == 0 {
		panic("eraBlockNumber is zero")
	}
	era := uint32(height / pc.eraBlockNumber)
	return height, era, nil
}
