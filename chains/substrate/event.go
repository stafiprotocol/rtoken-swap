package substrate

import (
	"errors"
	"rtoken-swap/config"
	"rtoken-swap/models/submodel"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var ErrMultiEnd = errors.New("multiEnd")

func (l *listener) processNewMultisigEvt(evt *submodel.ChainEvent) (*submodel.EventNewMultisig, error) {
	data, err := submodel.EventNewMultisigData(evt)
	if err != nil {
		return nil, err
	}

	mul := new(submodel.Multisig)
	exist, err := l.conn.QueryStorage(config.MultisigModuleId, config.StorageMultisigs, data.ID[:], data.CallHash[:], mul)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, ErrMultiEnd
	}

	data.TimePoint = submodel.NewOptionTimePoint(mul.When)
	data.Approvals = mul.Approvals
	data.CallHashStr = hexutil.Encode(data.CallHash[:])
	return data, nil
}

func (l *listener) processMultisigExecutedEvt(evt *submodel.ChainEvent) (*submodel.EventMultisigExecuted, error) {
	data, err := submodel.EventMultisigExecutedData(evt)
	if err != nil {
		return nil, err
	}
	data.CallHashStr = hexutil.Encode(data.CallHash[:])
	return data, nil
}
