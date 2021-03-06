package submodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"rtoken-swap/core"
	"rtoken-swap/utils"

	"github.com/ethereum/go-ethereum/common/hexutil"
	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/utiles"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

var (
	ValueNotStringError      = errors.New("value not string")
	ValueNotMapError         = errors.New("value not map")
	ValueNotU32              = errors.New("value not u32")
	ValueNotStringSliceError = errors.New("value not string slice")
)


func EventNewMultisigData(evt *ChainEvent) (*EventNewMultisig, error) {
	if len(evt.Params) != 3 {
		return nil, fmt.Errorf("EventNewMultisigData params number not right: %d, expected: 3", len(evt.Params))
	}
	who, err := parseAccountId(evt.Params[0].Value)
	if err != nil {
		return nil, fmt.Errorf("EventNewMultisig params[0] -> who error: %s", err)
	}

	id, err := parseAccountId(evt.Params[1].Value)
	if err != nil {
		return nil, fmt.Errorf("EventNewMultisig params[1] -> id error: %s", err)
	}

	hash, err := parseHash(evt.Params[2].Value)
	if err != nil {
		return nil, fmt.Errorf("EventNewMultisig params[2] -> hash error: %s", err)
	}

	return &EventNewMultisig{
		Who:      who,
		ID:       id,
		CallHash: hash,
	}, nil
}

func EventMultisigExecutedData(evt *ChainEvent) (*EventMultisigExecuted, error) {
	if len(evt.Params) != 5 {
		return nil, fmt.Errorf("EventMultisigExecuted params number not right: %d, expected: 5", len(evt.Params))
	}

	approving, err := parseAccountId(evt.Params[0].Value)
	if err != nil {
		return nil, fmt.Errorf("EventMultisigExecuted params[0] -> approving error: %s", err)
	}

	tp, err := parseTimePoint(evt.Params[1].Value)
	if err != nil {
		return nil, fmt.Errorf("EventMultisigExecuted params[1] -> timepoint error: %s", err)
	}

	id, err := parseAccountId(evt.Params[2].Value)
	if err != nil {
		return nil, fmt.Errorf("EventMultisigExecuted params[2] -> id error: %s", err)
	}

	hash, err := parseHash(evt.Params[3].Value)
	if err != nil {
		return nil, fmt.Errorf("EventMultisigExecuted params[3] -> hash error: %s", err)
	}

	ok, err := parseDispatchResult(evt.Params[4].Value)
	if err != nil {
		return nil, fmt.Errorf("EventMultisigExecuted params[4] -> dispatchresult error: %s", err)
	}

	return &EventMultisigExecuted{
		Who:       approving,
		TimePoint: tp,
		ID:        id,
		CallHash:  hash,
		Result:    ok,
	}, nil
}

func parseRsymbol(value interface{}) (core.RSymbol, error) {
	sym, ok := value.(string)
	if !ok {
		return core.RSymbol(""), ValueNotStringError
	}

	return core.RSymbol(sym), nil
}

func parseEra(param scalecodec.EventParam) (*Era, error) {
	bz, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	era := new(Era)
	err = json.Unmarshal(bz, era)
	if err != nil {
		return nil, err
	}

	return era, nil
}

func parseBytes(value interface{}) ([]byte, error) {
	val, ok := value.(string)
	if !ok {
		return nil, ValueNotStringError
	}

	bz, err := hexutil.Decode(utiles.AddHex(val))
	if err != nil {
		return nil, err
	}

	return bz, nil
}

func parseVecBytes(value interface{}) ([]types.Bytes, error) {
	vals, ok := value.([]interface{})
	if !ok {
		return nil, ValueNotStringSliceError
	}
	result := make([]types.Bytes, 0)
	for _, val := range vals {
		bz, err := parseBytes(val)
		if err != nil {
			return nil, err
		}

		result = append(result, bz)
	}

	return result, nil
}

func parseBigint(value interface{}) (*big.Int, error) {
	val, ok := value.(string)
	if !ok {
		return nil, ValueNotStringError
	}

	i, ok := utils.StringToBigint(val)
	if !ok {
		return nil, fmt.Errorf("string to bigint error: %s", val)
	}

	return i, nil
}

func parseAccountId(value interface{}) (types.AccountID, error) {
	val, ok := value.(string)
	if !ok {
		return types.NewAccountID([]byte{}), ValueNotStringError
	}
	ac, err := hexutil.Decode(utiles.AddHex(val))
	if err != nil {
		return types.NewAccountID([]byte{}), err
	}

	return types.NewAccountID(ac), nil
}

func parseHash(value interface{}) (types.Hash, error) {
	val, ok := value.(string)
	if !ok {
		return types.NewHash([]byte{}), ValueNotStringError
	}

	hash, err := types.NewHashFromHexString(utiles.AddHex(val))
	if err != nil {
		return types.NewHash([]byte{}), err
	}

	return hash, err
}

func parseTimePoint(value interface{}) (types.TimePoint, error) {
	bz, err := json.Marshal(value)
	if err != nil {
		return types.TimePoint{}, err
	}

	var tp types.TimePoint
	err = json.Unmarshal(bz, &tp)
	if err != nil {
		return types.TimePoint{}, err
	}

	return tp, nil
}

func parseDispatchResult(value interface{}) (bool, error) {
	result, ok := value.(map[string]interface{})
	if !ok {
		return false, ValueNotMapError
	}
	_, ok = result["Ok"]
	return ok, nil
}
