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

package gas

import (
	"context"
	"math"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.furychain.dev/gridiron/cosmos/x/evm/plugins"
	"pkg.furychain.dev/gridiron/eth/core"
	"pkg.furychain.dev/gridiron/eth/core/vm"
)

// gasMeterDescriptor is the descriptor for the gas meter used in the plugin.
const gasMeterDescriptor = `gridiron-gas-plugin`

// Plugin is the interface that must be implemented by the plugin.
type Plugin interface {
	plugins.Base
	core.GasPlugin
}

// plugin wraps a Cosmos context and utilize's the underlying `GasMeter` and `BlockGasMeter`
// to implement the core.GasPlugin interface.
type plugin struct {
	gasMeter        storetypes.GasMeter
	blockGasMeter   storetypes.GasMeter
	consensusMaxGas uint64
}

// NewPlugin creates a new instance of the gas plugin from a given context.
func NewPlugin() Plugin {
	return &plugin{}
}

// Prepare implements the core.GasPlugin interface.
func (p *plugin) Prepare(ctx context.Context) {
	p.resetMeters(sdk.UnwrapSDKContext(ctx))
}

// Reset implements the core.GasPlugin interface.
func (p *plugin) Reset(ctx context.Context) {
	p.resetMeters(sdk.UnwrapSDKContext(ctx))
}

// GasRemaining implements the core.GasPlugin interface.
func (p *plugin) GasRemaining() uint64 {
	return p.gasMeter.GasRemaining()
}

// BlockGasLimit implements the core.GasPlugin interface.
func (p *plugin) BlockGasLimit() uint64 {
	if blockGasLimit := p.blockGasMeter.Limit(); blockGasLimit != 0 {
		return blockGasLimit
	}
	return p.consensusMaxGas
}

// TxConsumeGas implements the core.GasPlugin interface.
func (p *plugin) ConsumeGas(amount uint64) error {
	// We don't want to panic if we overflow so we do some safety checks.
	// TODO: probably faster / cleaner to just wrap .ConsumeGas in a panic handler, or write our
	// own custom gas meter that doesn't panic on overflow.
	if newConsumed, overflow := addUint64Overflow(p.gasMeter.GasConsumed(), amount); overflow {
		return core.ErrGasUintOverflow
	} else if newConsumed > p.gasMeter.Limit() {
		return vm.ErrOutOfGas
	} else if p.blockGasMeter.GasConsumed()+newConsumed > p.blockGasMeter.Limit() {
		return core.ErrBlockOutOfGas
	}

	p.gasMeter.ConsumeGas(amount, gasMeterDescriptor)
	return nil
}

// GasConsumed returns the gas used during the current transaction.
//
// GasConsumed implements the core.GasPlugin interface.
func (p *plugin) GasConsumed() uint64 {
	return p.gasMeter.GasConsumed()
}

// BlockGasConsumed returns the cumulative gas used during the current block. If the cumulative
// gas used is greater than the block gas limit, we expect for Gridiron to handle it.
//
// BlockGasConsumed implements the core.GasPlugin interface.
func (p *plugin) BlockGasConsumed() uint64 {
	return p.blockGasMeter.GasConsumed()
}

// addUint64Overflow performs the addition operation on two uint64 integers and returns a boolean
// on whether or not the result overflows.
func addUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}

// if either of the gas meters on the sdk context are nil, this function will panic.
func (p *plugin) resetMeters(ctx sdk.Context) {
	if p.gasMeter = ctx.GasMeter(); p.gasMeter == nil {
		panic("gas meter is nil")
	}
	if p.blockGasMeter = ctx.BlockGasMeter(); p.blockGasMeter == nil {
		panic("block gas meter is nil")
	}
	if block := ctx.ConsensusParams().Block; block != nil {
		p.consensusMaxGas = uint64(block.MaxGas)
	}
}

func (p *plugin) IsPlugin() {}
