package main

import (
	"os"
	"rtoken-swap/shared/cosmos/rpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	xAuthTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

const DefaultHomeDir = "./keys/cosmos"

func main() {
	encodingConfig := rpc.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(xAuthTypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(DefaultHomeDir)

	rootCmd := &cobra.Command{
		Use:   "gencos",
		Short: "tool to manage cosmos keys",
		Long:  "Notice: The keyring supports os|file|test backends, but relay now only support the file backend",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd)
		},
	}
	rootCmd.AddCommand(keys.Commands(DefaultHomeDir))
	svrcmd.Execute(rootCmd, DefaultHomeDir)
}
