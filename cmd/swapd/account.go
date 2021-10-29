// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"rtoken-swap/config"

	log "github.com/ChainSafe/log15"
	"github.com/stafiprotocol/chainbridge/utils/crypto/sr25519"
	"github.com/stafiprotocol/chainbridge/utils/keystore"
	"github.com/urfave/cli/v2"
)

func handleGenerateSubCmd(ctx *cli.Context) error {
	log.Info("Generating substrate keyfile by rawseed...")
	path := ctx.String(config.KeystorePathFlag.Name)
	network := ctx.String(config.NetworkFlag.Name)
	return generateKeyFileByRawseed(path, network)
}

// keypath example: /Homepath/chainbridge/keys
func generateKeyFileByRawseed(keypath, network string) error {
	key := keystore.GetPassword("Enter mnemonic/rawseed:")
	kp, err := sr25519.NewKeypairFromSeed(string(key), network)
	if err != nil {
		return err
	}

	fp, err := filepath.Abs(keypath + "/" + kp.Address() + ".key")
	if err != nil {
		return fmt.Errorf("invalid filepath: %s", err)
	}

	file, err := os.OpenFile(filepath.Clean(fp), os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Error("generate keypair: could not close keystore file")
		}
	}()

	password := keystore.GetPassword("password for key:")
	err = keystore.EncryptAndWriteToFile(file, kp, password)
	if err != nil {
		return fmt.Errorf("could not write key to file: %s", err)
	}

	log.Info("key generated", "address", kp.Address(), "type", "sub", "file", fp)
	return nil
}
