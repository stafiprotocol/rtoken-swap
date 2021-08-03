package substrate_test

import (
	"fmt"
	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/models/submodel"
	"rtoken-swap/shared/substrate"
	"testing"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/chainbridge/utils/keystore"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
	"github.com/stretchr/testify/assert"
)

var (
	AliceKey     = keystore.TestKeyRing.SubstrateKeys[keystore.AliceKey].AsKeyringPair()
	From         = "31yavGB5CVb8EwpqKQaS9XY7JZcfbK6QpWPn5kkweHVpqcov"
	LessPolka    = "1334v66HrtqQndbugYxX9m56V6222m97LbavB4KAMmqgjsas"
	From1        = "31d96Cq9idWQqPq3Ch5BFY84zrThVE3r98M7vG4xYaSWHwsX"
	From2        = "1TgYb5x8xjsZRyL5bwvxUoAWBn36psr4viSMHbRXA8bkB2h"
	Wen          = "1swvN162p1siDjm63UhhWoa59bpPZTSNKGVcbCwHUYkfRRW"
	Jun          = "33RQ73d9XfPTaE2SV7dzdhQQ17YaeMQ4yzhzAQhhFVenxMuJ"
	KeystorePath = "/Users/fwj/Go/stafi/rtoken-relay/keys"
)

var (
	tlog = log15.Root()
)

const (
	stafiTypesFile  = "/Users/tpkeeper/gowork/stafi/rtoken-swap/network/stafi.json"
	polkaTypesFile  = "/Users/fwj/Go/stafi/rtoken-relay/network/polkadot.json"
	kusamaTypesFile = "/Users/fwj/Go/stafi/rtoken-relay/network/kusama.json"
)

var sc *substrate.SarpcClient
var gc *substrate.GsrpcClient

func init() {
	var err error
	sc, err = substrate.NewSarpcClient(substrate.ChainTypeStafi, "ws://127.0.0.1:9944", stafiTypesFile, tlog)
	if err != nil {
		panic(err)
	}
	stop := make(chan int)
	gc, err = substrate.NewGsrpcClient("ws://127.0.0.1:9944", substrate.AddressTypeMultiAddress, AliceKey, tlog, stop)
	if err != nil {
		panic(err)
	}
}

func TestGetLatestDealBLock(t *testing.T) {
	symBz, err := types.EncodeToBytes(core.RATOM)
	if err != nil {
		t.Fatal(err)
	}

	var block uint64
	exists, err := gc.QueryStorage(config.RDexnSwapModuleId, config.StorageLatestDealBLock, symBz, nil, &block)
	if err != nil {
		t.Fatal(fmt.Errorf("storge %s", err))
	}
	t.Log(exists, block)
}

func TestGetTransInfo(t *testing.T) {
	key := submodel.TransInfoKey{
		Symbol: core.RATOM,
		Block:  2947,
	}
	keyBz, err := types.EncodeToBytes(key)
	if err != nil {
		panic(err)
	}

	var transInfos []submodel.TransInfo

	exists, err := gc.QueryStorage(config.RDexnSwapModuleId, config.StorageTransInfos, keyBz, nil, &transInfos)
	if err != nil {
		t.Fatal(fmt.Errorf("storge %s", err))
	}
	t.Log(exists, transInfos)
}

func TestConnection_GetEvents(t *testing.T) {
	_, err := sc.GetEvents(7111716)
	assert.NoError(t, err)

	for i := 7111000; i < 7112000; i++ {
		evts, err := sc.GetEvents(uint64(i))
		assert.NoError(t, err)
		for _, evt := range evts {
			t.Log("eventId", evt.EventId)
			t.Log("moduleId", evt.ModuleId)
			t.Log("params", evt.Params)
		}
	}
}

func TestConnection_GetExtrinsics(t *testing.T) {

	for i := 7111000; i < 7112000; i++ {
		bh, err := sc.GetBlockHash(uint64(i))
		assert.NoError(t, err)
		exts, err := sc.GetExtrinsics(bh)
		assert.NoError(t, err)
		for _, ext := range exts {
			t.Log("\n=============", i)
			t.Log(ext.ExtrinsicHash)
			t.Log(ext.Address)
			t.Log(ext.CallModuleName)
			t.Log(ext.CallName)
			t.Log(ext.Params)
		}

	}
}

func TestConnection_PaymentQueryInfo(t *testing.T) {
	info, err := sc.GetPaymentQueryInfo("0x74b58b2d6fbd6319e4cc7927f1b789d48fc2629437cf9373bf8224934b831f58")
	assert.NoError(t, err)
	t.Log(info)
}
