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

package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/event"

	"pkg.furychain.dev/gridiron/eth/common"
	"pkg.furychain.dev/gridiron/eth/core/precompile"
	"pkg.furychain.dev/gridiron/eth/core/state"
	"pkg.furychain.dev/gridiron/eth/core/types"
	"pkg.furychain.dev/gridiron/eth/params"
	libtypes "pkg.furychain.dev/gridiron/lib/types"
)

// GridironHostChain defines the plugins that the chain running the Gridiron EVM should implement.
type GridironHostChain interface {
	// GetBlockPlugin returns the `BlockPlugin` of the Gridiron host chain.
	GetBlockPlugin() BlockPlugin
	// GetConfigurationPlugin returns the `ConfigurationPlugin` of the Gridiron host chain.
	GetConfigurationPlugin() ConfigurationPlugin
	// GetGasPlugin returns the `GasPlugin` of the Gridiron host chain.
	GetGasPlugin() GasPlugin
	// GetHistoricalPlugin returns the OPTIONAL `HistoricalPlugin` of the Gridiron host chain.
	GetHistoricalPlugin() HistoricalPlugin
	// GetPrecompilePlugin returns the OPTIONAL `PrecompilePlugin` of the Gridiron host chain.
	GetPrecompilePlugin() PrecompilePlugin
	// GetStatePlugin returns the `StatePlugin` of the Gridiron host chain.
	GetStatePlugin() StatePlugin
	// GetTxPoolPlugin returns the `TxPoolPlugin` of the Gridiron host chain.
	GetTxPoolPlugin() TxPoolPlugin
}

// =============================================================================
// Mandatory Plugins
// =============================================================================

// The following plugins should be implemented by the chain running the Gridiron EVM and exposed via
// the `GridironHostChain` interface. All plugins should be resettable with a given context.
type (
	// BlockPlugin defines the methods that the chain running the Gridiron EVM should implement to
	// support getting and setting block headers.
	BlockPlugin interface {
		// BlockPlugin implements `libtypes.Preparable`. Calling `Prepare` should reset the
		// BlockPlugin to a default state.
		libtypes.Preparable
		// GetNewBlockMetadata returns a new block metadata (coinbase, timestamp) for the given
		// block number.
		GetNewBlockMetadata(int64) (common.Address, uint64)
		// GetHeaderByNumber returns the block header at the given block number.
		GetHeaderByNumber(int64) (*types.Header, error)
		// SetHeaderByNumber sets the block header at the given block number.
		SetHeaderByNumber(int64, *types.Header) error
		// BaseFee returns the base fee of the current block.
		BaseFee() *big.Int
	}

	// ConfigurationPlugin defines the methods that the chain running Gridiron EVM should
	// implement in order to configuration the parameters of the Gridiron EVM.
	ConfigurationPlugin interface {
		// ConfigurationPlugin implements `libtypes.Preparable`. Calling `Prepare` should reset
		// the `ConfigurationPlugin` to a default state.
		libtypes.Preparable
		// ChainConfig returns the current chain configuration of the Gridiron EVM.
		ChainConfig() *params.ChainConfig
		// ExtraEips returns the extra EIPs that the Gridiron EVM supports.
		ExtraEips() []int
		// `The fee collector is utilized on chains that have a fee collector account. This was added
		// specifically to support Cosmos-SDK chains, where we want the coinbase in the block header
		// to be the operator address of the proposer, but we want the coinbase in the BlockContext
		// to be the FeeCollectorAccount.
		FeeCollector() *common.Address
	}

	// GasPlugin is an interface that allows the Gridiron EVM to consume gas on the host chain.
	GasPlugin interface {
		// GasPlugin implements `libtypes.Preparable`. Calling `Prepare` should reset the
		// GasPlugin to a default state.
		libtypes.Preparable
		// GasPlugin implements `libtypes.Resettable`. Calling `Reset` should reset the
		// GasPlugin to a default state
		libtypes.Resettable
		// ConsumeGas consumes the supplied amount of gas. It should not panic due to a
		// GasOverflow and should return `core.ErrOutOfGas` if the amount of gas remaining is
		// less than the amount requested. If the requested amount is greater than the amount of
		// gas remaining in the block, it should return core.ErrBlockOutOfGas.
		ConsumeGas(uint64) error
		// GasRemaining returns the amount of gas remaining for the current transaction.
		GasRemaining() uint64
		// GasConsumed returns the amount of gas used by the current transaction.
		GasConsumed() uint64
		// BlockGasConsumed returns the amount of gas used during the current block. The value
		// returned should NOT include any gas consumed during this transaction.
		// It should not panic.
		BlockGasConsumed() uint64
		// BlockGasLimit returns the gas limit of the current block. It should not panic.
		BlockGasLimit() uint64
	}

	// StatePlugin defines the methods that the chain running Gridiron EVM should implement.
	StatePlugin interface {
		state.Plugin
		// GetStateByNumber returns the state at the given block height.
		GetStateByNumber(int64) (StatePlugin, error)
	}

	// TxPoolPlugin defines the methods that the chain running Gridiron EVM should implement to
	// support the transaction pool.
	TxPoolPlugin interface {
		// SendTx submits the tx to the transaction pool.
		SendTx(tx *types.Transaction) error
		// Pending returns all pending transactions in the transaction pool.
		Pending(bool) map[common.Address]types.Transactions
		// Get returns the transaction from the pool with the given hash.
		Get(common.Hash) *types.Transaction
		// Nonce returns the nonce of the given address in the transaction pool.
		Nonce(common.Address) uint64
		// SubscribeNewTxsEvent returns a subscription with the new txs event channel.
		SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription
		// Stats returns the number of currently pending and queued (locally created) txs.
		Stats() (int, int)
		// Content retrieves the data content of the transaction pool, returning all the pending as
		// well as queued transactions, grouped by account and nonce.
		Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
		// ContentFrom retrieves the data content of the transaction pool, returning the pending
		// as well as queued transactions of this address, grouped by nonce.
		ContentFrom(addr common.Address) (types.Transactions, types.Transactions)
	}
)

// =============================================================================
// Optional Plugins
// =============================================================================

// `The following plugins are OPTIONAL to be implemented by the chain running Gridiron EVM.
type (
	// HistoricalPlugin defines the methods that the chain running Gridiron EVM should implement
	// in order to support storing historical blocks, receipts, and transactions. This plugin will
	// be used by the RPC backend to support certain methods on the Ethereum JSON RPC spec.
	// Implementing this plugin is optional.
	HistoricalPlugin interface {
		// HistoricalPlugin implements `libtypes.Preparable`.
		libtypes.Preparable
		// GetBlockByNumber returns the block at the given block number.
		GetBlockByNumber(int64) (*types.Block, error)
		// GetBlockByHash returns the block at the given block hash.
		GetBlockByHash(common.Hash) (*types.Block, error)
		// GetTransactionByHash returns the transaction lookup entry at the given transaction
		// hash.
		GetTransactionByHash(common.Hash) (*types.TxLookupEntry, error)
		// GetReceiptByHash returns the receipts at the given block hash.
		GetReceiptsByHash(common.Hash) (types.Receipts, error)
		// StoreBlock stores the given block.
		StoreBlock(*types.Block) error
		// StoreReceipts stores the receipts for the given block hash.
		StoreReceipts(common.Hash, types.Receipts) error
		// StoreTransactions stores the transactions for the given block hash.
		StoreTransactions(int64, common.Hash, types.Transactions) error
	}

	// PrecompilePlugin defines the methods that the chain running Gridiron EVM should implement
	// in order to support running their own stateful precompiled contracts. Implementing this
	// plugin is optional.
	PrecompilePlugin = precompile.Plugin
)
