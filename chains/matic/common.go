package matic

import (
	"errors"
	"math/big"

	"rtoken-swap/core"
	"rtoken-swap/models/submodel"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

var ErrNoOutPuts = errors.New("outputs length is zero")

func (w *writer) printContentError(m *core.Message, err error) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason, "err", err)
}

// submitMessage inserts the chainId into the msg and sends it to the router
func (w *writer) submitWriteMessage(m *core.Message) bool {
	err := w.router.SendWriteMesage(m)
	if err != nil {
		w.log.Error("failed to SendWriteMesage", "err", err)
		return false
	}
	return true
}

func (w *writer) submitReadMessage(m *core.Message) bool {
	err := w.router.SendReadMesage(m)
	if err != nil {
		w.log.Error("failed to SendReadMesage", "err", err)
		return false
	}
	return true
}

func (w *writer) reportTransResultWithBlock(source, dest core.RSymbol, m *submodel.TransResultWithBlock) bool {
	msg := &core.Message{Source: source, Destination: dest, Reason: core.ReportTransResultWithBlock, Content: m}
	return w.submitWriteMessage(msg)
}

// {
//     "internalType": "uint256",
//     "name": "_block",
//     "type": "uint256"
//   },
//   {
//     "internalType": "address[]",
//     "name": "_tos",
//     "type": "address[]"
//   },
//   {
//     "internalType": "uint256[]",
//     "name": "_values",
//     "type": "uint256[]"
//   }

func GetProposalHash(block *big.Int, tos []common.Address, values []*big.Int) [32]byte {
	uint256Ty, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		panic(err)
	}
	addressListTy, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		panic(err)
	}
	uint256ListTy, err := abi.NewType("uint256[]", "uint256[]", nil)
	if err != nil {
		panic(err)
	}
	var proposalArguments = abi.Arguments{
		{
			Type: uint256Ty,
		},
		{
			Type: addressListTy,
		},
		{
			Type: uint256ListTy,
		},
	}

	bytes, err := proposalArguments.Pack(block, tos, values)
	if err != nil {
		panic(err)
	}

	var retBuf []byte
	hash := sha3.NewLegacyKeccak256()
	hash.Write(bytes)
	retBuf = hash.Sum(retBuf)

	var ret [32]byte
	copy(ret[:], retBuf)
	return ret
}
