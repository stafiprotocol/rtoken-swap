package substrate_test

import (
	"testing"
	"time"

	"rtoken-swap/shared/substrate"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

var logger = log15.Root()

// endpoint := "wss://polkadot-rpc3.stafi.io"
var endpoint = "wss://kusama-rpc.polkadot.io"

// endpoint ="wss://mainnet-rpc.stafi.io"

func TestUpdateMetaData(t *testing.T) {
	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/kusama.json", substrate.AddressTypeAccountId, nil, logger)
	if err != nil {
		t.Fatal(err)
	}
	head, err := sc.GetFinalizedHead()

	if err != nil {
		t.Fatal(err)
	}
	err = sc.UpdateMeta(head.Hex())
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetEvents(t *testing.T) {
	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/kusama.json", substrate.AddressTypeAccountId, nil, logger)
	if err != nil {
		t.Fatal(err)
	}
	finalNumber, err := sc.GetFinalizedBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("finalNumber:", finalNumber)
	finalNumber = 9866422
	for i := finalNumber; i > 0; i++ {
		t.Log("now deal number", i)
		var hash types.Hash
		for {

			hashStr, err := sc.GetBlockHash(i)
			if err != nil {
				t.Log(err)
				time.Sleep(time.Second * 1)
				continue
			}

			hash, err = types.NewHashFromHexString(hashStr)
			if err != nil {
				t.Log(err)
				time.Sleep(time.Second * 1)
				continue
			}
			break
		}

		number, err := sc.GetBlockNumber(hash)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("number:", number)

		extrinsics, err := sc.GetExtrinsics(hash.Hex())
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range extrinsics {
			t.Log(n.Address, n.CallModuleName, n.CallName, n.Params)
		}

		event, err := sc.GetChainEvents(hash.Hex())
		if err != nil {
			t.Fatal(err)
		}

		for _, e := range event {
			t.Log(e.EventId, e.ModuleId, e.Params)
		}
	}
}
