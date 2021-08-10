// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package core

type Chain interface {
	Start() error // Start chain
	SetRouter(*Router)
	Rsymbol() RSymbol
	Name() string
	Stop()
}

type ChainConfig struct {
	Name            string  // Human-readable chain name
	Symbol          RSymbol // symbol
	Endpoint        string  // url for rpc endpoint
	Care            RSymbol
	LatestBlockFlag bool
	Accounts        []string               // addresses of key to use
	KeystorePath    string                 // Lortoken-swapion of key files
	Insecure        bool                   // Indirtoken-swaped whether the test keyring should be used
	Opts            map[string]interface{} // Per chain options
}
