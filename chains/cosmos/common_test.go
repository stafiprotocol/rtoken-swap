package cosmos_test

import (
	"encoding/hex"
	"rtoken-swap/chains/cosmos"
	"testing"

	substrateTypes "github.com/stafiprotocol/go-substrate-rpc-client/types"
)

func TestGetBondUnBondProposalId(t *testing.T) {
	bts := cosmos.GetTransferProposalId(substrateTypes.NewHash([]byte{2, 2}))
	t.Log(hex.EncodeToString(bts))
}
