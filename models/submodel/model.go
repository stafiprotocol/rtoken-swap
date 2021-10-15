package submodel

import (
	"rtoken-swap/core"

	scale "github.com/itering/scale.go"
	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/stafiprotocol/go-substrate-rpc-client/signature"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

type MultiEventFlow struct {
	EventId         string
	Symbol          core.RSymbol
	EventData       interface{}
	Block           uint64
	Index           uint32
	Threshold       uint16
	SubAccounts     []types.Bytes
	Key             *signature.KeyringPair
	Others          []types.AccountID
	OpaqueCalls     []*MultiOpaqueCall
	PaymentInfo     *rpc.PaymentQueryInfo
	NewMulCallHashs map[string]bool
	MulExeCallHashs map[string]bool
}

type EventNewMultisig struct {
	Who, ID     types.AccountID
	CallHash    types.Hash
	CallHashStr string
	TimePoint   *OptionTimePoint
	Approvals   []types.AccountID
}

type Multisig struct {
	When      types.TimePoint
	Deposit   types.U128
	Depositor types.AccountID
	Approvals []types.AccountID
}

type EventMultisigExecuted struct {
	Who, ID     types.AccountID
	TimePoint   types.TimePoint
	CallHash    types.Hash
	CallHashStr string
	Result      bool
}

type MultiCallParam struct {
	TimePoint *OptionTimePoint
	Opaque    []byte
	Extrinsic string
	CallHash  string
}

type Receive struct {
	Recipient []byte
	Value     types.UCompact
}

type Era struct {
	Type  string `json:"type"`
	Value uint32 `json:"value"`
}

type ChainEvent struct {
	ModuleId string             `json:"module_id" `
	EventId  string             `json:"event_id" `
	Params   []scale.EventParam `json:"params"`
}

type MultiOpaqueCall struct {
	Extrinsic string
	Opaque    []byte
	CallHash  string
	TimePoint *OptionTimePoint
}

type Transaction struct {
	ExtrinsicHash  string
	CallModuleName string
	CallName       string
	Address        interface{}
	Params         []scale.ExtrinsicParam
}

type TransInfoSingle struct {
	Block      uint64
	Index      uint32
	DestSymbol core.RSymbol
	Info       TransInfo
}

type TransInfoList struct {
	Block      uint64
	DestSymbol core.RSymbol
	List       []TransInfo
}

type TransInfoKey struct {
	Symbol core.RSymbol
	Block  uint64
}

type TransInfo struct {
	Account  types.AccountID
	Receiver []byte
	Value    types.U128
	IsDeal   bool `json:"is_deal"`
}

type TransResultWithBlock struct {
	Symbol core.RSymbol
	Block  uint64
}

type TransResultWithIndex struct {
	Symbol core.RSymbol
	Block  uint64
	Index  uint32
}

type GetLatestDealBLockParam struct {
	Symbol core.RSymbol
	Block  chan uint64
}

type GetSignaturesParam struct {
	Symbol     core.RSymbol
	Block      uint64
	ProposalId []byte
	Signatures chan []types.Bytes
}

type GetSignaturesKey struct {
	Block      uint64
	ProposalId types.Bytes
}

type SubmitSignatureParams struct {
	Symbol     core.RSymbol
	Block      types.U64
	ProposalId types.Bytes
	Signature  types.Bytes
}
