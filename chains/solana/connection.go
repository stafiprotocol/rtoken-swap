package solana

import (
	"fmt"

	"rtoken-swap/config"
	"rtoken-swap/core"
	"rtoken-swap/shared/solana"
	"rtoken-swap/shared/solana/vault"

	"github.com/ChainSafe/log15"
	solClient "github.com/stafiprotocol/solana-go-sdk/client"
	solCommon "github.com/stafiprotocol/solana-go-sdk/common"
	solTypes "github.com/stafiprotocol/solana-go-sdk/types"
)

type Connection struct {
	endpoint   string
	symbol     core.RSymbol
	poolClient *solana.PoolClient //map[poolAddressHexStr]poolClient
	log        log15.Logger
	stop       <-chan int
}

type PoolAccounts struct {
	FeeAccount            string `json:"feeAccount"`
	MultisigTxBaseAccount string `json:"multisigTxBaseAccount"`
	MultisigInfoPubkey    string `json:"multisigInfoPubkey"`
	MultisigProgramId     string `json:"multisigProgramId"`
}

func NewConnection(cfg *core.ChainConfig, log log15.Logger, stop <-chan int) (*Connection, error) {

	v, err := vault.NewVaultFromWalletFile(cfg.KeystorePath)
	if err != nil {
		return nil, err
	}
	boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap)
	if err != nil {
		return nil, fmt.Errorf("secret boxer: %w", err)
	}

	if err := v.Open(boxer); err != nil {
		return nil, fmt.Errorf("opening: %w", err)
	}

	feeAccountStr, ok := cfg.Opts[config.FeeAccountFlagKey].(string)
	if !ok || len(feeAccountStr) == 0 {
		return nil, fmt.Errorf("no fee account")
	}

	multisigTxBaseAccountStr, ok := cfg.Opts[config.MultisigTxBaseAccountFlagKey].(string)
	if !ok || len(multisigTxBaseAccountStr) == 0 {
		return nil, fmt.Errorf("no multisig tx base account pubkey")
	}
	multisigTxBaseAccountPubkey := solCommon.PublicKeyFromString(multisigTxBaseAccountStr)

	multisigInfoPubkeyStr, ok := cfg.Opts[config.MultisigInfoPubkeyFlagKey].(string)
	if !ok || len(multisigInfoPubkeyStr) == 0 {
		return nil, fmt.Errorf("no multisig tx info account pubkey")
	}
	multisigInfoPubkey := solCommon.PublicKeyFromString(multisigInfoPubkeyStr)

	programIdStr, ok := cfg.Opts[config.MultisigProgramIdFlagKey].(string)
	if !ok || len(multisigInfoPubkey) == 0 {
		return nil, fmt.Errorf("no multisig programid")
	}
	programId := solCommon.PublicKeyFromString(programIdStr)
	multisignerPubkey, _, err := solCommon.FindProgramAddress([][]byte{multisigInfoPubkey.Bytes()}, programId)
	if err != nil {
		return nil, err
	}

	//collect privkey
	pubKeyStrToPrivKey := make(map[string]vault.PrivateKey)
	for _, privKey := range v.KeyBag {
		pubKeyStrToPrivKey[privKey.PublicKey().String()] = privKey
	}

	var multisigTxBaseAccount *solTypes.Account
	if pKey, exist := pubKeyStrToPrivKey[multisigTxBaseAccountStr]; exist {
		account := solTypes.AccountFromPrivateKeyBytes(pKey)
		multisigTxBaseAccount = &account
	}

	if _, exist := pubKeyStrToPrivKey[feeAccountStr]; !exist {
		return nil, fmt.Errorf("no fee account privatekey")
	}

	//check auth
	hasBaseAccountAuth := false
	if multisigTxBaseAccount != nil {
		hasBaseAccountAuth = true
	}

	poolAccounts := solana.PoolAccounts{
		FeeAccount:                  solTypes.AccountFromPrivateKeyBytes(pubKeyStrToPrivKey[feeAccountStr]),
		MultisigTxBaseAccount:       multisigTxBaseAccount,
		MultisigTxBaseAccountPubkey: multisigTxBaseAccountPubkey,
		MultisigInfoPubkey:          multisigInfoPubkey,
		MultisignerPubkey:           multisignerPubkey,
		MultisigProgramId:           programId,
		HasBaseAccountAuth:          hasBaseAccountAuth,
	}
	solanaPoolClient := solana.NewPoolClient(log, solClient.NewClient([]string{cfg.Endpoint}), poolAccounts)

	return &Connection{
		endpoint:   cfg.Endpoint,
		symbol:     cfg.Symbol,
		log:        log,
		stop:       stop,
		poolClient: solanaPoolClient,
	}, nil
}

func (c *Connection) GetPoolClient() (*solana.PoolClient, error) {
	return c.poolClient, nil
}
