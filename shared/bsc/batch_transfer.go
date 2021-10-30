// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bsc

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// BatchTransferABI is the input ABI used to generate the binding from.
const BatchTransferABI = "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"initialSubAccounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"initialThreshold\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"_proposals\",\"outputs\":[{\"internalType\":\"enumBatchTransfer.ProposalStatus\",\"name\":\"_status\",\"type\":\"uint8\"},{\"internalType\":\"uint40\",\"name\":\"_yesVotes\",\"type\":\"uint40\"},{\"internalType\":\"uint8\",\"name\":\"_yesVotesTotal\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_threshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"subAccount\",\"type\":\"address\"}],\"name\":\"addSubAccount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_block\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"_tos\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_values\",\"type\":\"uint256[]\"}],\"name\":\"batchTransfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newThreshold\",\"type\":\"uint256\"}],\"name\":\"changeThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"subAccount\",\"type\":\"address\"}],\"name\":\"getSubAccountIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"subAccount\",\"type\":\"address\"}],\"name\":\"removeSubAccount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// BatchTransfer is an auto generated Go binding around an Ethereum contract.
type BatchTransfer struct {
	BatchTransferCaller     // Read-only binding to the contract
	BatchTransferTransactor // Write-only binding to the contract
	BatchTransferFilterer   // Log filterer for contract events
}

// BatchTransferCaller is an auto generated read-only Go binding around an Ethereum contract.
type BatchTransferCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchTransferTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BatchTransferTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchTransferFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BatchTransferFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchTransferSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BatchTransferSession struct {
	Contract     *BatchTransfer    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BatchTransferCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BatchTransferCallerSession struct {
	Contract *BatchTransferCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// BatchTransferTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BatchTransferTransactorSession struct {
	Contract     *BatchTransferTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// BatchTransferRaw is an auto generated low-level Go binding around an Ethereum contract.
type BatchTransferRaw struct {
	Contract *BatchTransfer // Generic contract binding to access the raw methods on
}

// BatchTransferCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BatchTransferCallerRaw struct {
	Contract *BatchTransferCaller // Generic read-only contract binding to access the raw methods on
}

// BatchTransferTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BatchTransferTransactorRaw struct {
	Contract *BatchTransferTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBatchTransfer creates a new instance of BatchTransfer, bound to a specific deployed contract.
func NewBatchTransfer(address common.Address, backend bind.ContractBackend) (*BatchTransfer, error) {
	contract, err := bindBatchTransfer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BatchTransfer{BatchTransferCaller: BatchTransferCaller{contract: contract}, BatchTransferTransactor: BatchTransferTransactor{contract: contract}, BatchTransferFilterer: BatchTransferFilterer{contract: contract}}, nil
}

// NewBatchTransferCaller creates a new read-only instance of BatchTransfer, bound to a specific deployed contract.
func NewBatchTransferCaller(address common.Address, caller bind.ContractCaller) (*BatchTransferCaller, error) {
	contract, err := bindBatchTransfer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BatchTransferCaller{contract: contract}, nil
}

// NewBatchTransferTransactor creates a new write-only instance of BatchTransfer, bound to a specific deployed contract.
func NewBatchTransferTransactor(address common.Address, transactor bind.ContractTransactor) (*BatchTransferTransactor, error) {
	contract, err := bindBatchTransfer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BatchTransferTransactor{contract: contract}, nil
}

// NewBatchTransferFilterer creates a new log filterer instance of BatchTransfer, bound to a specific deployed contract.
func NewBatchTransferFilterer(address common.Address, filterer bind.ContractFilterer) (*BatchTransferFilterer, error) {
	contract, err := bindBatchTransfer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BatchTransferFilterer{contract: contract}, nil
}

// bindBatchTransfer binds a generic wrapper to an already deployed contract.
func bindBatchTransfer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BatchTransferABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchTransfer *BatchTransferRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchTransfer.Contract.BatchTransferCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchTransfer *BatchTransferRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchTransfer.Contract.BatchTransferTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchTransfer *BatchTransferRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchTransfer.Contract.BatchTransferTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchTransfer *BatchTransferCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchTransfer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchTransfer *BatchTransferTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchTransfer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchTransfer *BatchTransferTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchTransfer.Contract.contract.Transact(opts, method, params...)
}

// Proposals is a free data retrieval call binding the contract method 0xf2e2af94.
//
// Solidity: function _proposals(bytes32 ) view returns(uint8 _status, uint40 _yesVotes, uint8 _yesVotesTotal)
func (_BatchTransfer *BatchTransferCaller) Proposals(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Status        uint8
	YesVotes      *big.Int
	YesVotesTotal uint8
}, error) {
	var out []interface{}
	err := _BatchTransfer.contract.Call(opts, &out, "_proposals", arg0)

	outstruct := new(struct {
		Status        uint8
		YesVotes      *big.Int
		YesVotesTotal uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Status = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.YesVotes = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.YesVotesTotal = *abi.ConvertType(out[2], new(uint8)).(*uint8)

	return *outstruct, err

}

// Proposals is a free data retrieval call binding the contract method 0xf2e2af94.
//
// Solidity: function _proposals(bytes32 ) view returns(uint8 _status, uint40 _yesVotes, uint8 _yesVotesTotal)
func (_BatchTransfer *BatchTransferSession) Proposals(arg0 [32]byte) (struct {
	Status        uint8
	YesVotes      *big.Int
	YesVotesTotal uint8
}, error) {
	return _BatchTransfer.Contract.Proposals(&_BatchTransfer.CallOpts, arg0)
}

// Proposals is a free data retrieval call binding the contract method 0xf2e2af94.
//
// Solidity: function _proposals(bytes32 ) view returns(uint8 _status, uint40 _yesVotes, uint8 _yesVotesTotal)
func (_BatchTransfer *BatchTransferCallerSession) Proposals(arg0 [32]byte) (struct {
	Status        uint8
	YesVotes      *big.Int
	YesVotesTotal uint8
}, error) {
	return _BatchTransfer.Contract.Proposals(&_BatchTransfer.CallOpts, arg0)
}

// Threshold is a free data retrieval call binding the contract method 0x7f3c8160.
//
// Solidity: function _threshold() view returns(uint8)
func (_BatchTransfer *BatchTransferCaller) Threshold(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _BatchTransfer.contract.Call(opts, &out, "_threshold")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Threshold is a free data retrieval call binding the contract method 0x7f3c8160.
//
// Solidity: function _threshold() view returns(uint8)
func (_BatchTransfer *BatchTransferSession) Threshold() (uint8, error) {
	return _BatchTransfer.Contract.Threshold(&_BatchTransfer.CallOpts)
}

// Threshold is a free data retrieval call binding the contract method 0x7f3c8160.
//
// Solidity: function _threshold() view returns(uint8)
func (_BatchTransfer *BatchTransferCallerSession) Threshold() (uint8, error) {
	return _BatchTransfer.Contract.Threshold(&_BatchTransfer.CallOpts)
}

// GetSubAccountIndex is a free data retrieval call binding the contract method 0x763f8680.
//
// Solidity: function getSubAccountIndex(address subAccount) view returns(uint256)
func (_BatchTransfer *BatchTransferCaller) GetSubAccountIndex(opts *bind.CallOpts, subAccount common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BatchTransfer.contract.Call(opts, &out, "getSubAccountIndex", subAccount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSubAccountIndex is a free data retrieval call binding the contract method 0x763f8680.
//
// Solidity: function getSubAccountIndex(address subAccount) view returns(uint256)
func (_BatchTransfer *BatchTransferSession) GetSubAccountIndex(subAccount common.Address) (*big.Int, error) {
	return _BatchTransfer.Contract.GetSubAccountIndex(&_BatchTransfer.CallOpts, subAccount)
}

// GetSubAccountIndex is a free data retrieval call binding the contract method 0x763f8680.
//
// Solidity: function getSubAccountIndex(address subAccount) view returns(uint256)
func (_BatchTransfer *BatchTransferCallerSession) GetSubAccountIndex(subAccount common.Address) (*big.Int, error) {
	return _BatchTransfer.Contract.GetSubAccountIndex(&_BatchTransfer.CallOpts, subAccount)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchTransfer *BatchTransferCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchTransfer.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchTransfer *BatchTransferSession) Owner() (common.Address, error) {
	return _BatchTransfer.Contract.Owner(&_BatchTransfer.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchTransfer *BatchTransferCallerSession) Owner() (common.Address, error) {
	return _BatchTransfer.Contract.Owner(&_BatchTransfer.CallOpts)
}

// AddSubAccount is a paid mutator transaction binding the contract method 0xb52e1eff.
//
// Solidity: function addSubAccount(address subAccount) returns()
func (_BatchTransfer *BatchTransferTransactor) AddSubAccount(opts *bind.TransactOpts, subAccount common.Address) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "addSubAccount", subAccount)
}

// AddSubAccount is a paid mutator transaction binding the contract method 0xb52e1eff.
//
// Solidity: function addSubAccount(address subAccount) returns()
func (_BatchTransfer *BatchTransferSession) AddSubAccount(subAccount common.Address) (*types.Transaction, error) {
	return _BatchTransfer.Contract.AddSubAccount(&_BatchTransfer.TransactOpts, subAccount)
}

// AddSubAccount is a paid mutator transaction binding the contract method 0xb52e1eff.
//
// Solidity: function addSubAccount(address subAccount) returns()
func (_BatchTransfer *BatchTransferTransactorSession) AddSubAccount(subAccount common.Address) (*types.Transaction, error) {
	return _BatchTransfer.Contract.AddSubAccount(&_BatchTransfer.TransactOpts, subAccount)
}

// BatchTransfer is a paid mutator transaction binding the contract method 0x7b5cbf50.
//
// Solidity: function batchTransfer(uint256 _block, address[] _tos, uint256[] _values) returns()
func (_BatchTransfer *BatchTransferTransactor) BatchTransfer(opts *bind.TransactOpts, _block *big.Int, _tos []common.Address, _values []*big.Int) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "batchTransfer", _block, _tos, _values)
}

// BatchTransfer is a paid mutator transaction binding the contract method 0x7b5cbf50.
//
// Solidity: function batchTransfer(uint256 _block, address[] _tos, uint256[] _values) returns()
func (_BatchTransfer *BatchTransferSession) BatchTransfer(_block *big.Int, _tos []common.Address, _values []*big.Int) (*types.Transaction, error) {
	return _BatchTransfer.Contract.BatchTransfer(&_BatchTransfer.TransactOpts, _block, _tos, _values)
}

// BatchTransfer is a paid mutator transaction binding the contract method 0x7b5cbf50.
//
// Solidity: function batchTransfer(uint256 _block, address[] _tos, uint256[] _values) returns()
func (_BatchTransfer *BatchTransferTransactorSession) BatchTransfer(_block *big.Int, _tos []common.Address, _values []*big.Int) (*types.Transaction, error) {
	return _BatchTransfer.Contract.BatchTransfer(&_BatchTransfer.TransactOpts, _block, _tos, _values)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 newThreshold) returns()
func (_BatchTransfer *BatchTransferTransactor) ChangeThreshold(opts *bind.TransactOpts, newThreshold *big.Int) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "changeThreshold", newThreshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 newThreshold) returns()
func (_BatchTransfer *BatchTransferSession) ChangeThreshold(newThreshold *big.Int) (*types.Transaction, error) {
	return _BatchTransfer.Contract.ChangeThreshold(&_BatchTransfer.TransactOpts, newThreshold)
}

// ChangeThreshold is a paid mutator transaction binding the contract method 0x694e80c3.
//
// Solidity: function changeThreshold(uint256 newThreshold) returns()
func (_BatchTransfer *BatchTransferTransactorSession) ChangeThreshold(newThreshold *big.Int) (*types.Transaction, error) {
	return _BatchTransfer.Contract.ChangeThreshold(&_BatchTransfer.TransactOpts, newThreshold)
}

// RemoveSubAccount is a paid mutator transaction binding the contract method 0x0ba3aae3.
//
// Solidity: function removeSubAccount(address subAccount) returns()
func (_BatchTransfer *BatchTransferTransactor) RemoveSubAccount(opts *bind.TransactOpts, subAccount common.Address) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "removeSubAccount", subAccount)
}

// RemoveSubAccount is a paid mutator transaction binding the contract method 0x0ba3aae3.
//
// Solidity: function removeSubAccount(address subAccount) returns()
func (_BatchTransfer *BatchTransferSession) RemoveSubAccount(subAccount common.Address) (*types.Transaction, error) {
	return _BatchTransfer.Contract.RemoveSubAccount(&_BatchTransfer.TransactOpts, subAccount)
}

// RemoveSubAccount is a paid mutator transaction binding the contract method 0x0ba3aae3.
//
// Solidity: function removeSubAccount(address subAccount) returns()
func (_BatchTransfer *BatchTransferTransactorSession) RemoveSubAccount(subAccount common.Address) (*types.Transaction, error) {
	return _BatchTransfer.Contract.RemoveSubAccount(&_BatchTransfer.TransactOpts, subAccount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BatchTransfer *BatchTransferTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BatchTransfer *BatchTransferSession) RenounceOwnership() (*types.Transaction, error) {
	return _BatchTransfer.Contract.RenounceOwnership(&_BatchTransfer.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BatchTransfer *BatchTransferTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BatchTransfer.Contract.RenounceOwnership(&_BatchTransfer.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchTransfer *BatchTransferTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchTransfer *BatchTransferSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BatchTransfer.Contract.TransferOwnership(&_BatchTransfer.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchTransfer *BatchTransferTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BatchTransfer.Contract.TransferOwnership(&_BatchTransfer.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_BatchTransfer *BatchTransferTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchTransfer.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_BatchTransfer *BatchTransferSession) Withdraw() (*types.Transaction, error) {
	return _BatchTransfer.Contract.Withdraw(&_BatchTransfer.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_BatchTransfer *BatchTransferTransactorSession) Withdraw() (*types.Transaction, error) {
	return _BatchTransfer.Contract.Withdraw(&_BatchTransfer.TransactOpts)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_BatchTransfer *BatchTransferTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _BatchTransfer.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_BatchTransfer *BatchTransferSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _BatchTransfer.Contract.Fallback(&_BatchTransfer.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_BatchTransfer *BatchTransferTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _BatchTransfer.Contract.Fallback(&_BatchTransfer.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BatchTransfer *BatchTransferTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchTransfer.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BatchTransfer *BatchTransferSession) Receive() (*types.Transaction, error) {
	return _BatchTransfer.Contract.Receive(&_BatchTransfer.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_BatchTransfer *BatchTransferTransactorSession) Receive() (*types.Transaction, error) {
	return _BatchTransfer.Contract.Receive(&_BatchTransfer.TransactOpts)
}

// BatchTransferOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BatchTransfer contract.
type BatchTransferOwnershipTransferredIterator struct {
	Event *BatchTransferOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchTransferOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchTransferOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchTransferOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchTransferOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchTransferOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchTransferOwnershipTransferred represents a OwnershipTransferred event raised by the BatchTransfer contract.
type BatchTransferOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchTransfer *BatchTransferFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BatchTransferOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchTransfer.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BatchTransferOwnershipTransferredIterator{contract: _BatchTransfer.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchTransfer *BatchTransferFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BatchTransferOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchTransfer.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchTransferOwnershipTransferred)
				if err := _BatchTransfer.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchTransfer *BatchTransferFilterer) ParseOwnershipTransferred(log types.Log) (*BatchTransferOwnershipTransferred, error) {
	event := new(BatchTransferOwnershipTransferred)
	if err := _BatchTransfer.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
