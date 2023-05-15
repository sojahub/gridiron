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

package network

import (
	"encoding/json"
	"fmt"
	"time"

	cdb "github.com/cosmos/cosmos-db"

	pruningtypes "cosmossdk.io/store/pruning/types"

	baseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ethhd "pkg.furychain.dev/gridiron/cosmos/crypto/hd"
	ethkeyring "pkg.furychain.dev/gridiron/cosmos/crypto/keyring"
	"pkg.furychain.dev/gridiron/cosmos/crypto/keys/ethsecp256k1"
	runtime "pkg.furychain.dev/gridiron/cosmos/runtime"
	config "pkg.furychain.dev/gridiron/cosmos/runtime/config"
)

type (
	Network = network.Network
	Config  = network.Config
)

const (
	two         = 2
	thousand    = 1000
	fivehundred = 500
	onehundred  = 100
	megamoney   = 1000000
	gigamoney   = 1000000000
	examoney    = 1000000000000000000
	one2345     = 12345
)

type TestingT interface {
	Fatal(args ...interface{})
	Cleanup(func())
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	TempDir() string
}

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t TestingT, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}

	var cfg network.Config
	if len(configs) == 0 {
		newKey, _ := ethsecp256k1.GenPrivKey()
		cfg = DefaultConfig(map[string]*ethsecp256k1.PrivKey{"alice": newKey})
	} else {
		cfg = configs[0]
	}

	net, err := network.New(t, t.TempDir(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig.
func DefaultConfig(keysMap map[string]*ethsecp256k1.PrivKey) network.Config {
	encoding := config.BuildGridironEncodingConfig(runtime.ModuleBasics)
	cfg := network.Config{
		Codec:             encoding.Codec,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			return runtime.NewGridironApp(
				val.GetCtx().Logger, cdb.NewMemDB(), nil, true, sims.EmptyAppOptions{},
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID("gridiron-2061"),
			)
		},
		GenesisState:    BuildGenesisState(keysMap),
		TimeoutCommit:   2 * time.Second, //nolint:gomnd // 2 seconds is the default.
		ChainID:         "gridiron-2061",
		NumValidators:   1,
		BondDenom:       "afury",
		MinGasPrices:    fmt.Sprintf("0.00006%s", "afury"),
		AccountTokens:   sdk.TokensFromConsensusPower(thousand, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(fivehundred, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(onehundred, sdk.DefaultPowerReduction),
		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      false,
		SigningAlgo:     string(ethhd.EthSecp256k1Type),
		KeyringOptions:  []keyring.Option{ethkeyring.EthSecp256k1Option()},
	}

	return cfg
}

func BuildGenesisState(keysMap map[string]*ethsecp256k1.PrivKey) map[string]json.RawMessage {
	encoding := config.BuildGridironEncodingConfig(runtime.ModuleBasics)
	genState := runtime.ModuleBasics.DefaultGenesis(encoding.Codec)

	// Auth & Bank module
	var authState authtypes.GenesisState
	var bankState banktypes.GenesisState

	encoding.Codec.MustUnmarshalJSON(genState[authtypes.ModuleName], &authState)
	encoding.Codec.MustUnmarshalJSON(genState[banktypes.ModuleName], &bankState)

	for mapKey, testKey := range keysMap {
		newAccount, err := authtypes.NewBaseAccountWithPubKey(testKey.PubKey())
		if err != nil {
			panic(err)
		}
		accounts, err := authtypes.PackAccounts([]authtypes.GenesisAccount{newAccount})
		if err != nil {
			panic(err)
		}
		authState.Accounts = append(authState.Accounts, accounts[0])
		bankState.Balances = append(bankState.Balances, banktypes.Balance{
			Address: newAccount.Address,
			Coins:   getCoinsForAccount(mapKey),
		})
	}

	bankState.DenomMetadata = getTestMetadata()
	bankState.SendEnabled = []banktypes.SendEnabled{
		{
			Denom:   "afury",
			Enabled: true,
		},
		{
			Denom:   "stake",
			Enabled: true,
		},
	}

	genState[authtypes.ModuleName] = encoding.Codec.MustMarshalJSON(&authState)
	genState[banktypes.ModuleName] = encoding.Codec.MustMarshalJSON(&bankState)

	// Staking module
	var stakingState stakingtypes.GenesisState
	encoding.Codec.MustUnmarshalJSON(genState[stakingtypes.ModuleName], &stakingState)
	stakingState.Params.BondDenom = "afury"
	genState[stakingtypes.ModuleName] = encoding.Codec.MustMarshalJSON(&stakingState)

	// Distribution Module
	var distributionState distrtypes.GenesisState
	encoding.Codec.MustUnmarshalJSON(genState[distrtypes.ModuleName], &distributionState)
	params := distrtypes.DefaultParams()
	params.WithdrawAddrEnabled = true
	distributionState.Params = params
	genState[distrtypes.ModuleName] = encoding.Codec.MustMarshalJSON(&distributionState)

	return genState
}

//nolint:gomnd // its okay.
func getTestMetadata() []banktypes.Metadata {
	return []banktypes.Metadata{
		{
			Name:        "Furychain fury",
			Symbol:      "FURY",
			Description: "The Fury.",
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: "fury", Exponent: uint32(0), Aliases: []string{"fury"}},
				{Denom: "nfury", Exponent: uint32(9), Aliases: []string{"nanofury"}},
				{Denom: "afury", Exponent: uint32(18), Aliases: []string{"attofury"}},
			},
			Base:    "afury",
			Display: "fury",
		},
		{
			Name:        "Token",
			Symbol:      "TOKEN",
			Description: "The native staking token of the Token Hub.",
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: "1token", Exponent: uint32(5), Aliases: []string{"decitoken"}},
				{Denom: "2token", Exponent: uint32(4), Aliases: []string{"centitoken"}},
				{Denom: "3token", Exponent: uint32(7), Aliases: []string{"dekatoken"}},
			},
			Base:    "utoken",
			Display: "token",
		},
		{
			Name:   "Osmosis",
			Symbol: "OSMO",
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: "osmo", Exponent: uint32(0), Aliases: []string{"osmo"}},
				{Denom: "nosmo", Exponent: uint32(9), Aliases: []string{"nanoosmo"}},
				{Denom: "aosmo", Exponent: uint32(18), Aliases: []string{"attoosmo"}},
			},
			Base:    "bOSMO",
			Display: "osmosis",
		},
		{
			Name:   "Akash",
			Symbol: "AKT",
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: "akt", Exponent: uint32(0), Aliases: []string{"akt"}},
				{Denom: "nakt", Exponent: uint32(9), Aliases: []string{"nanoakt"}},
				{Denom: "aakt", Exponent: uint32(18), Aliases: []string{"attoakt"}},
			},
			Base:    "bAKT",
			Display: "akash",
		},
	}
}

func getCoinsForAccount(name string) sdk.Coins {
	switch name {
	case "alice":
		return sdk.NewCoins(
			sdk.NewCoin("afury", sdk.NewInt(examoney)),
			sdk.NewCoin("bATOM", sdk.NewInt(examoney)),
			sdk.NewCoin("bAKT", sdk.NewInt(12345)), //nolint:gomnd // its okay.
			sdk.NewCoin("stake", sdk.NewInt(examoney)),
			sdk.NewCoin("bOSMO", sdk.NewInt(12345*2)), //nolint:gomnd // its okay.
		)
	case "bob":
		return sdk.NewCoins(
			sdk.NewCoin("afury", sdk.NewInt(onehundred)),
			sdk.NewCoin("atoken", sdk.NewInt(onehundred)),
			sdk.NewCoin("stake", sdk.NewInt(examoney)),
		)
	case "charlie":
		return sdk.NewCoins(sdk.NewCoin("afury", sdk.NewInt(gigamoney)))
	default:
		return sdk.NewCoins(sdk.NewCoin("afury", sdk.NewInt(examoney)))
	}
}
