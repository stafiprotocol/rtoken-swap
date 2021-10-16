package substrate

import (
	"errors"
)

const (
	ChainTypeStafi    = "stafi"
	ChainTypePolkadot = "polkadot"

	AddressTypeAccountId    = "AccountId"
	AddressTypeMultiAddress = "MultiAddress"
)

var (
	ErrBondEqualToUnbond = errors.New("BondEqualToUnbondError")
)
