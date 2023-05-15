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

package baseapp

import (
	"cosmossdk.io/client/v2/autocli"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	ethcryptocodec "pkg.furychain.dev/gridiron/cosmos/crypto/codec"
	"pkg.furychain.dev/gridiron/cosmos/x/erc20"
	erc20keeper "pkg.furychain.dev/gridiron/cosmos/x/erc20/keeper"
	"pkg.furychain.dev/gridiron/cosmos/x/evm"
	evmkeeper "pkg.furychain.dev/gridiron/cosmos/x/evm/keeper"
	"pkg.furychain.dev/gridiron/lib/utils"

	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
)

var (
	// ModuleBasics is in charge of setting up basic, non-dependant module elements,.
	ModuleBasics = []module.AppModuleBasic{
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		consensus.AppModuleBasic{},
		evm.AppModuleBasic{},
		erc20.AppModuleBasic{},
	}
)

// GridironBaseApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type GridironBaseApp struct {
	*runtime.App
	LegacyAminoCodec       *codec.LegacyAmino
	ApplicationCodec       codec.Codec
	TxnConfig              client.TxConfig
	CodecInterfaceRegistry codectypes.InterfaceRegistry
	AutoCliOptions         autocli.AppOptions

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper

	// gridiron keepers
	EVMKeeper   *evmkeeper.Keeper
	ERC20Keeper *erc20keeper.Keeper
}

// Name returns the name of the App.
func (app *GridironBaseApp) Name() string { return app.BaseApp.Name() }

// LegacyAmino returns GridironBaseApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *GridironBaseApp) LegacyAmino() *codec.LegacyAmino {
	return app.LegacyAminoCodec
}

// AppCodec returns GridironBaseApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *GridironBaseApp) AppCodec() codec.Codec {
	return app.ApplicationCodec
}

// InterfaceRegistry returns GridironBaseApp's InterfaceRegistry.
func (app *GridironBaseApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.CodecInterfaceRegistry
}

// TxConfig returns GridironBaseApp's TxConfig.
func (app *GridironBaseApp) TxConfig() client.TxConfig {
	return app.TxnConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *GridironBaseApp) AutoCliOpts() autocli.AppOptions {
	return app.AutoCliOptions
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *GridironBaseApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *GridironBaseApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	kvStoreKey, ok := utils.GetAs[*storetypes.KVStoreKey](app.UnsafeFindStoreKey(storeKey))
	if !ok {
		return nil
	}
	return kvStoreKey
}

// KVStoreKeys returns all the registered KVStoreKeys.
func (app *GridironBaseApp) KVStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := utils.GetAs[*storetypes.KVStoreKey](k); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *GridironBaseApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
	app.EVMKeeper.SetClientCtx(apiSvr.ClientCtx)
}

// RegisterEthSecp256k1SignatureType registers the eth_secp256k1 signature type.
func (app *GridironBaseApp) RegisterEthSecp256k1SignatureType() {
	ethcryptocodec.RegisterInterfaces(app.CodecInterfaceRegistry)
}

// MountCustomStore mounts a custom store to the baseapp.
// TODO: GET UPSTREAMED
func (app *GridironBaseApp) MountCustomStores(keys ...storetypes.StoreKey) {
	for _, key := range keys {
		// StoreTypeDB doesn't do anything upon commit, so its blessed for the offchain stuff.
		app.MountStore(key, storetypes.StoreTypeDB)
	}
}
