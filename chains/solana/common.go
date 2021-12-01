package solana

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	solClient "github.com/stafiprotocol/solana-go-sdk/client"
	solCommon "github.com/stafiprotocol/solana-go-sdk/common"
	"github.com/stafiprotocol/solana-go-sdk/multisigprog"
	"github.com/stafiprotocol/solana-go-sdk/sysprog"
	solTypes "github.com/stafiprotocol/solana-go-sdk/types"
	"rtoken-swap/core"
	"rtoken-swap/shared/solana"
)

const (
	BlockRetryLimit    = 100
	BlockRetryInterval = time.Second * 6
	BlockConfirmNumber = 10
	initStakeAmount    = uint64(10000)
)

func (w *writer) printContentError(m *core.Message, err error) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "reason", m.Reason, "err", err)
}

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

type MultisigTxType string

var MultisigTxStakeType = MultisigTxType("stake")
var MultisigTxUnStakeType = MultisigTxType("unstake")
var MultisigTxWithdrawType = MultisigTxType("withdraw")
var MultisigTxTransferType = MultisigTxType("transfer")

func GetMultisigTxAccountPubkey(baseAccount, programID solCommon.PublicKey, txType MultisigTxType, era uint32, stakeBaseAccountIndex int) (solCommon.PublicKey, string) {
	seed := fmt.Sprintf("multisig:%s:%d:%d", txType, era, stakeBaseAccountIndex)
	return solCommon.CreateWithSeed(baseAccount, seed, programID), seed
}

func GetMultisigTxAccountPubkeyForTransfer(baseAccount, programID solCommon.PublicKey, era uint32, batchTimes int) (solCommon.PublicKey, string) {
	seed := fmt.Sprintf("multisig:%s:%d:%d", MultisigTxTransferType, era, batchTimes)
	return solCommon.CreateWithSeed(baseAccount, seed, programID), seed
}

func GetStakeAccountPubkey(baseAccount solCommon.PublicKey, era uint32) (solCommon.PublicKey, string) {
	seed := fmt.Sprintf("stake:%d", era)
	return solCommon.CreateWithSeed(baseAccount, seed, solCommon.StakeProgramID), seed
}

func mapToString(accountsMap map[solCommon.PublicKey]solClient.GetStakeActivationResponse) string {
	ret := ""
	for account, active := range accountsMap {
		ret = ret + account.ToBase58() + fmt.Sprintf(" : %+v\n", active)
	}
	return ret
}

func (w *writer) waitingForMultisigTxExe(rpcClient *solClient.Client, poolAddress, multisigTxAddress, processName string) bool {
	retry := 0
	for {
		if retry >= BlockRetryLimit {
			w.log.Error(fmt.Sprintf("[%s] GetMultisigTxAccountInfo reach retry limit", processName),
				"pool  address", poolAddress,
				"multisig tx account address", multisigTxAddress)
			return false
		}
		multisigTxAccountInfo, err := rpcClient.GetMultisigTxAccountInfo(context.Background(), multisigTxAddress)
		if err == nil && multisigTxAccountInfo.DidExecute == 1 {
			break
		} else {
			w.log.Warn(fmt.Sprintf("[%s] multisigTxAccount not execute yet, waiting...", processName),
				"pool  address", poolAddress,
				"multisig tx Account", multisigTxAddress)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
	}
	return true
}

func (w *writer) waitingForMultisigTxCreate(rpcClient *solClient.Client, poolAddress, multisigTxAddress, processName string) bool {
	retry := 0
	for {
		if retry >= BlockRetryLimit {
			w.log.Error(fmt.Sprintf("[%s] GetMultisigTxAccountInfo reach retry limit", processName),
				"pool  address", poolAddress,
				"multisig tx account address", multisigTxAddress)
			return false
		}
		_, err := rpcClient.GetMultisigTxAccountInfo(context.Background(), multisigTxAddress)
		if err != nil {
			w.log.Warn(fmt.Sprintf("[%s] GetMultisigTxAccountInfo failed, waiting...", processName),
				"pool  address", poolAddress,
				"multisig tx account", multisigTxAddress,
				"err", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		} else {
			break
		}
	}
	return true
}

func (w *writer) createMultisigTxAccount(
	rpcClient *solClient.Client,
	poolClient *solana.PoolClient,
	poolAddress string,
	programsIds []solCommon.PublicKey,
	accountMetas [][]solTypes.AccountMeta,
	datas [][]byte,
	multisigTxAccountPubkey solCommon.PublicKey,
	multisigTxAccountSeed string,
	processName string,
) bool {
	res, err := rpcClient.GetRecentBlockhash(context.Background())
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] GetRecentBlockhash failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}
	miniMumBalanceForTx, err := rpcClient.GetMinimumBalanceForRentExemption(context.Background(), solClient.MultisigTxAccountLengthDefault)
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] GetMinimumBalanceForRentExemption failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}
	miniMumBalanceForTx += initStakeAmount
	//send from one relayers
	//create multisig tx account of this era
	rawTx, err := solTypes.CreateRawTransaction(solTypes.CreateRawTransactionParam{
		Instructions: []solTypes.Instruction{
			sysprog.CreateAccountWithSeed(
				poolClient.FeeAccount.PublicKey,
				multisigTxAccountPubkey,
				poolClient.MultisigTxBaseAccount.PublicKey,
				poolClient.MultisigProgramId,
				multisigTxAccountSeed,
				miniMumBalanceForTx,
				solClient.MultisigTxAccountLengthDefault,
			),
			multisigprog.CreateTransaction(
				poolClient.MultisigProgramId,
				programsIds,
				accountMetas,
				datas,
				poolClient.MultisigInfoPubkey,
				multisigTxAccountPubkey,
				poolClient.FeeAccount.PublicKey,
			),
		},
		Signers:         []solTypes.Account{poolClient.FeeAccount, *poolClient.MultisigTxBaseAccount},
		FeePayer:        poolClient.FeeAccount.PublicKey,
		RecentBlockHash: res.Blockhash,
	})

	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] CreateTransaction CreateRawTransaction failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}

	txHash, err := rpcClient.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] createTransaction SendRawTransaction failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}
	w.log.Info(fmt.Sprintf("[%s] create multisig tx account  has send", processName),
		"tx hash", txHash,
		"multisig tx account", multisigTxAccountPubkey.ToBase58())
	return true
}

func (w *writer) approveMultisigTx(
	rpcClient *solClient.Client,
	poolClient *solana.PoolClient,
	poolAddress string,
	multisigTxAccountPubkey solCommon.PublicKey,
	remainingAccounts []solTypes.AccountMeta,
	processName string) bool {
	res, err := rpcClient.GetRecentBlockhash(context.Background())
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] GetRecentBlockhash failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}
	rawTx, err := solTypes.CreateRawTransaction(solTypes.CreateRawTransactionParam{
		Instructions: []solTypes.Instruction{
			multisigprog.Approve(
				poolClient.MultisigProgramId,
				poolClient.MultisigInfoPubkey,
				poolClient.MultisignerPubkey,
				multisigTxAccountPubkey,
				poolClient.FeeAccount.PublicKey,
				remainingAccounts,
			),
		},
		Signers:         []solTypes.Account{poolClient.FeeAccount},
		FeePayer:        poolClient.FeeAccount.PublicKey,
		RecentBlockHash: res.Blockhash,
	})

	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] approve CreateRawTransaction failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}

	txHash, err := rpcClient.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] approve SendRawTransaction failed", processName),
			"pool address", poolAddress,
			"err", err)
		return false
	}

	w.log.Info(fmt.Sprintf("[%s] approve multisig tx account has send", processName),
		"tx hash", txHash,
		"multisig tx account", multisigTxAccountPubkey.ToBase58())

	return true
}

func (w *writer) IsMultisigTxExe(
	rpcClient *solClient.Client,
	multisigTxAccountPubkey solCommon.PublicKey) bool {
	accountInfo, err := rpcClient.GetMultisigTxAccountInfo(context.Background(), multisigTxAccountPubkey.ToBase58())
	if err == nil && accountInfo.DidExecute == 1 {
		return true
	}
	return false
}

func (w *writer) CheckMultisigTx(
	rpcClient *solClient.Client,
	multisigTxAccountPubkey solCommon.PublicKey,
	programsIds []solCommon.PublicKey,
	accountMetas [][]solTypes.AccountMeta,
	datas [][]byte) bool {

	retry := 0
	var err error
	var accountInfo *solClient.GetMultisigTxAccountInfo
	for {
		if retry >= BlockRetryLimit {
			w.log.Error("CheckMultisigTx reach retry limit",
				"multisig tx account address", multisigTxAccountPubkey.ToBase58(),
				"err", err)
			return false
		}
		accountInfo, err = rpcClient.GetMultisigTxAccountInfo(context.Background(), multisigTxAccountPubkey.ToBase58())
		if err != nil {
			w.log.Warn("CheckMultisigTx failed, waiting...",
				"multisig tx account", multisigTxAccountPubkey.ToBase58(),
				"err", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		} else {
			break
		}
	}

	thisProgramsIdsBts, err := solCommon.SerializeData(programsIds)
	if err != nil {
		w.log.Error("CheckMultisigTx serializeData err",
			"programsIds", programsIds,
			"err", err)
		return false
	}
	thisAccountMetasBts, err := solCommon.SerializeData(accountMetas)
	if err != nil {
		w.log.Error("CheckMultisigTx serializeData err",
			"accountMetas", accountMetas,
			"err", err)
		return false
	}
	thisDatasBts, err := solCommon.SerializeData(datas)
	if err != nil {
		w.log.Error("CheckMultisigTx serializeData err",
			"datas", datas,
			"err", err)
		return false
	}
	onchainProgramsIdsBts, err := solCommon.SerializeData(accountInfo.ProgramID)
	if err != nil {
		w.log.Error("CheckMultisigTx serializeData err",
			"accountInfo.ProgramID", accountInfo.ProgramID,
			"err", err)
		return false
	}
	onchainAccountMetasBts, err := solCommon.SerializeData(accountInfo.Accounts)
	if err != nil {
		w.log.Error("CheckMultisigTx serializeData err",
			"accountInfo.Accounts", accountInfo.Accounts,
			"err", err)
		return false
	}
	onchainDatasBts, err := solCommon.SerializeData(accountInfo.Data)
	if err != nil {
		w.log.Error("CheckMultisigTx serializeData err",
			"accountInfo.Data", accountInfo.Data,
			"err", err)
		return false
	}
	if bytes.Equal(thisProgramsIdsBts, onchainProgramsIdsBts) &&
		bytes.Equal(thisAccountMetasBts, onchainAccountMetasBts) &&
		bytes.Equal(thisDatasBts, onchainDatasBts) {
		return true
	}
	w.log.Error("CheckMultisigTx not equal ",
		"thisprogramsIds", hex.EncodeToString(thisProgramsIdsBts),
		"onchainProgramnsIdsBts", hex.EncodeToString(onchainProgramsIdsBts),
		"thisAccountMetasBts", hex.EncodeToString(thisAccountMetasBts),
		"onchainAccountMetasBts", hex.EncodeToString(onchainAccountMetasBts),
		"thisDatasBts", hex.EncodeToString(thisDatasBts),
		"onchainDatasBts", hex.EncodeToString(onchainDatasBts))

	return false
}
