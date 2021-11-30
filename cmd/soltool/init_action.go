package main

import (
	"context"
	"fmt"
	"time"

	"rtoken-swap/shared/solana/vault"

	solClient "github.com/stafiprotocol/solana-go-sdk/client"
	solCommon "github.com/stafiprotocol/solana-go-sdk/common"
	"github.com/stafiprotocol/solana-go-sdk/multisigprog"
	"github.com/stafiprotocol/solana-go-sdk/sysprog"
	solTypes "github.com/stafiprotocol/solana-go-sdk/types"
	"github.com/urfave/cli/v2"
)

func initAction(ctx *cli.Context) error {
	path := ctx.String(configFlag.Name)
	pc := PoolAccounts{}
	err := loadConfig(path, &pc)
	if err != nil {
		return err
	}
	fmt.Printf("accounts info %+v\n\n", pc)
	v, err := vault.NewVaultFromWalletFile(pc.KeystorePath)
	if err != nil {
		return err
	}
	boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap)
	if err != nil {
		return fmt.Errorf("secret boxer: %w", err)
	}

	if err := v.Open(boxer); err != nil {
		return fmt.Errorf("opening: %w", err)
	}

	privKeyMap := make(map[string]vault.PrivateKey)
	for _, privKey := range v.KeyBag {
		privKeyMap[privKey.PublicKey().String()] = privKey
	}

	FeeAccount := solTypes.AccountFromPrivateKeyBytes(privKeyMap[pc.FeeAccount])
	MultisigInfoAccount := solTypes.AccountFromPrivateKeyBytes(privKeyMap[pc.MultisigInfoAccount])

	MultisigProgramId := solCommon.PublicKeyFromString(pc.MultisigProgramId)

	multisignerPubkey, nonce, err := solCommon.FindProgramAddress([][]byte{MultisigInfoAccount.PublicKey.Bytes()}, MultisigProgramId)
	if err != nil {
		return err
	}
	fmt.Println("multisigner:", multisignerPubkey.ToBase58())

	owners := make([]solCommon.PublicKey, 0)
	owners = append(owners, FeeAccount.PublicKey)
	for _, pubkeyStr := range pc.OtherFeePubkey {
		pubkey := solCommon.PublicKeyFromString(pubkeyStr)
		owners = append(owners, pubkey)
	}

	c := solClient.NewClient([]string{pc.Endpoint})
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		return err
	}
	multisigAccountMiniMum, err := c.GetMinimumBalanceForRentExemption(context.Background(), 1000)
	if err != nil {
		return err
	}

	res, err = c.GetRecentBlockhash(context.Background())
	if err != nil {
		return fmt.Errorf("get recent block hash error, err: %v", err)
	}

	//create multisigInfo account
	rawTx, err := solTypes.CreateRawTransaction(solTypes.CreateRawTransactionParam{
		Instructions: []solTypes.Instruction{
			sysprog.CreateAccount(
				FeeAccount.PublicKey,
				MultisigInfoAccount.PublicKey,
				MultisigProgramId,
				multisigAccountMiniMum*2,
				1000,
			),
			multisigprog.CreateMultisig(
				MultisigProgramId,
				MultisigInfoAccount.PublicKey,
				owners,
				uint64(pc.Threshold),
				uint8(nonce),
			),
		},
		Signers:         []solTypes.Account{FeeAccount, MultisigInfoAccount},
		FeePayer:        FeeAccount.PublicKey,
		RecentBlockHash: res.Blockhash,
	})
	if err != nil {
		return fmt.Errorf("generate tx error, err: %v", err)
	}
	txHash, err := c.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		return fmt.Errorf("send tx error, err: %v", err)
	}
	fmt.Println("createMultisig txHash:", txHash)

	for i := 0; i < 10; i++ {
		multiSigAccountInfo, err := c.GetMultisigInfoAccountInfo(context.Background(), MultisigInfoAccount.PublicKey.ToBase58())
		if err != nil {
			fmt.Println("GetMultisigInfoAccountInfo failed will retry ...", err)
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Println("create multisig info account success", multiSigAccountInfo)
		return nil
	}
	fmt.Println("sorry init failed")
	return nil
}
