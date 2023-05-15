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

package precompile

import (
	"context"
	"math/big"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.furychain.dev/gridiron/cosmos/x/evm/plugins"
	"pkg.furychain.dev/gridiron/cosmos/x/evm/plugins/state"
	"pkg.furychain.dev/gridiron/eth/common"
	"pkg.furychain.dev/gridiron/eth/core"
	ethprecompile "pkg.furychain.dev/gridiron/eth/core/precompile"
	"pkg.furychain.dev/gridiron/eth/core/vm"
	"pkg.furychain.dev/gridiron/eth/params"
	"pkg.furychain.dev/gridiron/lib/registry"
	libtypes "pkg.furychain.dev/gridiron/lib/types"
	"pkg.furychain.dev/gridiron/lib/utils"
)

// Plugin is the interface that must be implemented by the plugin.
type Plugin interface {
	plugins.Base
	core.PrecompilePlugin

	KVGasConfig() storetypes.GasConfig
	SetKVGasConfig(storetypes.GasConfig)
	TransientKVGasConfig() storetypes.GasConfig
	SetTransientKVGasConfig(storetypes.GasConfig)
}

// plugin runs precompile containers in the Cosmos environment with the context gas configs.
type plugin struct {
	libtypes.Registry[common.Address, vm.PrecompileContainer]
	// precompiles is all supported precompile contracts.
	precompiles []ethprecompile.Registrable
	// kvGasConfig is the gas config for the KV store.
	kvGasConfig storetypes.GasConfig
	// transientKVGasConfig is the gas config for the transient KV store.
	transientKVGasConfig storetypes.GasConfig
	// sp allows resetting the context for the reentrancy into the EVM.
	sp StatePlugin
}

// NewPlugin creates and returns a plugin with the default KV store gas configs.
func NewPlugin(precompiles []ethprecompile.Registrable, sp StatePlugin) Plugin {
	return &plugin{
		Registry:    registry.NewMap[common.Address, vm.PrecompileContainer](),
		precompiles: precompiles,
		// TODO: Re-enable gas config for precompiles.
		// https://github.com/furychain/gridiron/issues/393
		kvGasConfig:          storetypes.GasConfig{},
		transientKVGasConfig: storetypes.GasConfig{},
		sp:                   sp,
	}
}

// GetPrecompiles implements core.PrecompilePlugin.
func (p *plugin) GetPrecompiles(_ *params.Rules) []ethprecompile.Registrable {
	return p.precompiles
}

// GetActive implements core.PrecompilePlugin.
func (p *plugin) GetActive(rules *params.Rules) []common.Address {
	defaults := ethprecompile.GetDefaultPrecompiles(rules)
	active := make([]common.Address, len(p.precompiles)+len(defaults))
	for i, pc := range p.precompiles {
		active[i] = pc.RegistryKey()
	}
	for i, pc := range defaults {
		active[i+len(p.precompiles)] = pc.RegistryKey()
	}
	return active
}

// KVGasConfig implements Plugin.
func (p *plugin) KVGasConfig() storetypes.GasConfig {
	return p.kvGasConfig
}

// SetKVGasConfig implements Plugin.
func (p *plugin) SetKVGasConfig(kvGasConfig storetypes.GasConfig) {
	p.kvGasConfig = kvGasConfig
}

// TransientKVGasConfig implements Plugin.
func (p *plugin) TransientKVGasConfig() storetypes.GasConfig {
	return p.transientKVGasConfig
}

// SetTransientKVGasConfig implements Plugin.
func (p *plugin) SetTransientKVGasConfig(transientKVGasConfig storetypes.GasConfig) {
	p.transientKVGasConfig = transientKVGasConfig
}

// Run runs the a precompile container and returns the remaining gas after execution by injecting
// a Cosmos SDK `GasMeter`. This function returns an error if the precompile execution returns an
// error or insufficient gas is provided.
//
// Run implements core.PrecompilePlugin.
func (p *plugin) Run(
	evm ethprecompile.EVM, pc vm.PrecompileContainer, input []byte,
	caller common.Address, value *big.Int, suppliedGas uint64, readonly bool,
) ([]byte, uint64, error) {
	// use a precompile-specific gas meter for dynamic consumption
	gm := storetypes.NewInfiniteGasMeter()
	// consume static gas from RequiredGas
	gm.ConsumeGas(pc.RequiredGas(input), "RequiredGas")

	// get native Cosmos SDK context from the Gridiron StateDB
	sdb := utils.MustGetAs[vm.GridironStateDB](evm.GetStateDB())
	ctx := sdk.UnwrapSDKContext(sdb.GetContext())

	// begin precompile execution => begin emitting Cosmos event as Eth logs
	cem := utils.MustGetAs[state.ControllableEventManager](ctx.EventManager())
	cem.BeginPrecompileExecution(sdb)

	// run precompile container
	ret, err := pc.Run(
		ctx.WithGasMeter(gm).
			WithKVGasConfig(p.kvGasConfig).
			WithTransientKVGasConfig(p.transientKVGasConfig),
		evm,
		input,
		caller,
		value,
		readonly,
	)

	// end precompile execution => stop emitting Cosmos event as Eth logs
	cem.EndPrecompileExecution()

	// handle overconsumption of gas
	if gm.GasConsumed() > suppliedGas {
		return nil, 0, vm.ErrOutOfGas
	}

	// valid precompile gas consumption => return supplied gas
	return ret, suppliedGas - gm.GasConsumed(), err
}

// EnableReentrancy sets the state so that execution can enter the EVM again.
//
// EnableReentrancy implements core.PrecompilePlugin.
func (p *plugin) EnableReentrancy(_ context.Context) {
	// We remove the KVStore gas metering from the context prior to entering the EVM state
	// transition. This is because the EVM is not aware of the Cosmos SDK's gas metering and is
	// designed to be used in a standalone manner, as each of the EVM's opcodes are priced
	// individually. By setting the gas configs to empty structs, we ensure that SLOADS and SSTORES
	// in the EVM are not being charged additional gas unknowingly.
	p.sp.SetGasConfig(storetypes.GasConfig{}, storetypes.GasConfig{})
}

// DisableReentrancy sets the state so that execution cannot enter the EVM again.
//
// DisableReentrancy implements core.PrecompilePlugin.
func (p *plugin) DisableReentrancy(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// restore ctx gas configs for continuing precompile execution
	p.sp.SetGasConfig(sdkCtx.KVGasConfig(), sdkCtx.TransientKVGasConfig())
}

func (p *plugin) IsPlugin() {}
