// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2023, Furychain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package provider

import (
	"github.com/ethereum/go-ethereum/node"

	"pkg.furychain.dev/gridiron/eth/api"
	"pkg.furychain.dev/gridiron/eth/core"
	"pkg.furychain.dev/gridiron/eth/log"
	"pkg.furychain.dev/gridiron/eth/rpc"
)

// GridironProvider is the only object that an implementing chain should use.
type GridironProvider struct {
	api.Chain
	backend rpc.GridironBackend
	Node    *node.Node
}

// NewGridironProvider creates a new `GridironEVM` instance for use on an underlying blockchain.
func NewGridironProvider(
	configPath string,
	dataDir string,
	host core.GridironHostChain,
	logHandler log.Handler,
) *GridironProvider {
	// Load the config file.
	cfg, err := LoadConfigFromFilePath(configPath)
	if err != nil {
		// todo: this is hood.
		cfg = DefaultConfig()
	}

	// set the data dir
	cfg.NodeConfig.DataDir = dataDir

	// Create the Gridiron Provider.
	return NewGridironProviderWithConfig(cfg, host, logHandler)
}

// NewGridironProvider creates a new `GridironEVM` instance for use on an underlying blockchain.
func NewGridironProviderWithConfig(
	cfg *Config,
	host core.GridironHostChain,
	logHandler log.Handler,
) *GridironProvider {
	sp := &GridironProvider{}
	// When creating a Gridiron EVM, we allow the implementing chain
	// to specify their own log handler. If logHandler is nil then we
	// we use the default geth log handler.
	if logHandler != nil {
		// Root is a global in geth that is used by the evm to emit logs.
		log.Root().SetHandler(logHandler)
	}

	// Build the chain from the host.
	sp.Chain = core.NewChain(host)

	// Build and set the RPC Backend.
	sp.backend = rpc.NewGridironBackend(sp.Chain, &cfg.RPCConfig, &cfg.NodeConfig)

	var err error
	sp.Node, err = node.New(&cfg.NodeConfig)
	if err != nil {
		panic(err)
	}

	return sp
}

// StartServices starts the standard go-ethereum node-services (i.e json-rpc).
func (sp *GridironProvider) StartServices() error {
	sp.Node.RegisterAPIs(rpc.GetAPIs(sp.backend))
	return sp.Node.Start()
}
