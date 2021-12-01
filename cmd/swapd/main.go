package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"rtoken-swap/chains/bsc"
	"rtoken-swap/chains/cosmos"
	"rtoken-swap/chains/matic"
	"rtoken-swap/chains/solana"
	"rtoken-swap/chains/substrate"

	"rtoken-swap/chains/stafix"
	"rtoken-swap/config"
	"rtoken-swap/core"

	log "github.com/ChainSafe/log15"
	"github.com/urfave/cli/v2"
)

var app = cli.NewApp()

var cliFlags = []cli.Flag{
	config.ConfigFileFlag,
	config.VerbosityFlag,
}

var generateFlags = []cli.Flag{
	config.KeystorePathFlag,
	config.NetworkFlag,
}
var generateEthFlags = []cli.Flag{
	config.KeystorePathFlag,
}

var accountCommand = cli.Command{
	Name:        "accounts",
	Usage:       "manage keystores",
	Description: "The accounts command is used to manage the keystore.\n",
	Subcommands: []*cli.Command{
		{
			Action: handleGenerateSubCmd,
			Name:   "gensub",
			Usage:  "generate subsrate keystore",
			Flags:  generateFlags,
			Description: "The generate subcommand is used to generate the substrate keystore.\n" +
				"\tkeystore path should be given.",
		}, {
			Action: handleGenerateEthCmd,
			Name:   "geneth",
			Usage:  "generate eth keystore",
			Flags:  generateEthFlags,
			Description: "The generate subcommand is used to generate the eth keystore.\n" +
				"\tkeystore path should be given.",
		},
	},
}

// init initializes CLI
func init() {
	app.Action = run
	app.Copyright = "Copyright 2021 Stafi Protocol Authors"
	app.Name = "rtoken-swap"
	app.Usage = "rtoken-swap"
	app.Authors = []*cli.Author{{Name: "Stafi Protocol 2021"}}
	app.Version = "1.0.0"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		&accountCommand,
	}

	app.Flags = append(app.Flags, cliFlags...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startLogger(ctx *cli.Context) error {
	logger := log.Root()
	var lvl log.Lvl

	if lvlToInt, err := strconv.Atoi(ctx.String(config.VerbosityFlag.Name)); err == nil {
		lvl = log.Lvl(lvlToInt)
	} else if lvl, err = log.LvlFromString(ctx.String(config.VerbosityFlag.Name)); err != nil {
		return err
	}

	logger.SetHandler(log.MultiHandler(
		log.LvlFilterHandler(
			lvl,
			log.StreamHandler(os.Stdout, log.LogfmtFormat())),
		log.Must.FileHandler("relay_log.json", log.JsonFormat()),
		log.LvlFilterHandler(
			log.LvlError,
			log.Must.FileHandler("relay_log_errors.json", log.JsonFormat()))))

	return nil
}

func run(ctx *cli.Context) error {
	err := startLogger(ctx)
	if err != nil {
		return err
	}

	cfg, err := config.GetConfig(ctx)
	if err != nil {
		return err
	}
	if len(cfg.Chains) < 2 {
		return fmt.Errorf("config err, no chains info")
	}

	// Used to signal core shutdown due to fatal error
	sysErr := make(chan error)
	c := core.NewCore(sysErr)

	for _, chain := range cfg.Chains {
		chainConfig := &core.ChainConfig{
			Name:            chain.Name,
			Symbol:          core.RSymbol(chain.Rsymbol),
			Endpoint:        chain.Endpoint,
			KeystorePath:    chain.KeystorePath,
			Care:            core.RSymbol(chain.Care),
			LatestBlockFlag: chain.LatestBlockFlag,
			Insecure:        false,
			Opts:            chain.Opts,
		}
		var newChain core.Chain
		logger := log.Root().New("chain", chainConfig.Name)

		switch chain.Name {
		case "stafix":
			newChain, err = stafix.InitializeChain(chainConfig, logger, sysErr)
			if err != nil {
				return err
			}
		case "cosmos":
			newChain, err = cosmos.InitializeChain(chainConfig, logger, sysErr)
			if err != nil {
				return err
			}
		case "polkadot", "kusama", "stafi":
			newChain, err = substrate.InitializeChain(chainConfig, logger, sysErr)
			if err != nil {
				return err
			}
		case "bnb":
			newChain, err = bsc.InitializeChain(chainConfig, logger, sysErr)
			if err != nil {
				return err
			}
		case "matic":
			newChain, err = matic.InitializeChain(chainConfig, logger, sysErr)
			if err != nil {
				return err
			}
		case "solana":
			newChain, err = solana.InitializeChain(chainConfig, logger, sysErr)
			if err != nil {
				return err
			}
		default:
			return errors.New("unrecognized Chain Type")
		}

		c.AddChain(newChain)
	}

	c.Start()
	return nil
}
