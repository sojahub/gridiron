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

package lib

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	generated "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/lib"
	"pkg.furychain.dev/gridiron/cosmos/precompile"
	"pkg.furychain.dev/gridiron/lib/utils"
)

/**
 * This file contains conversions between native Cosmos SDK types and go-ethereum ABI types.
 */

// SdkCoinsToEvmCoins converts sdk.Coins into []generated.CosmosCoin.
func SdkCoinsToEvmCoins(sdkCoins sdk.Coins) []generated.CosmosCoin {
	evmCoins := make([]generated.CosmosCoin, len(sdkCoins))
	for i, coin := range sdkCoins {
		evmCoins[i] = generated.CosmosCoin{
			Amount: coin.Amount.BigInt(),
			Denom:  coin.Denom,
		}
	}
	return evmCoins
}

// ExtractCoinsFromInput converts coins from input (of type any) into sdk.Coins.
func ExtractCoinsFromInput(coins any) (sdk.Coins, error) {
	// note: we have to use unnamed struct here, otherwise the compiler cannot cast
	// the any type input into IBankModuleCoin.
	amounts, ok := utils.GetAs[[]struct {
		Amount *big.Int `json:"amount"`
		Denom  string   `json:"denom"`
	}](coins)
	if !ok {
		return nil, precompile.ErrInvalidCoin
	}

	sdkCoins := sdk.NewCoins()
	for _, evmCoin := range amounts {
		sdkCoins = append(sdkCoins, sdk.NewCoin(evmCoin.Denom, sdk.NewIntFromBigInt(evmCoin.Amount)))
	}
	return sdkCoins, nil
}

// SdkCoinsToUnnamedCoins converts sdk.Coins into an unnamed struct.
func SdkCoinsToUnnamedCoins(coins sdk.Coins) any {
	unnamedCoins := []struct {
		Amount *big.Int `json:"amount"`
		Denom  string   `json:"denom"`
	}{}
	for _, coin := range coins {
		unnamedCoins = append(unnamedCoins, struct {
			Amount *big.Int `json:"amount"`
			Denom  string   `json:"denom"`
		}{
			Amount: coin.Amount.BigInt(),
			Denom:  coin.Denom,
		})
	}
	return unnamedCoins
}

// GetGrantAsSendAuth maps a list of grants to a list of send authorizations.
func GetGrantAsSendAuth(
	grants []*authz.Grant, blocktime time.Time,
) ([]*banktypes.SendAuthorization, error) {
	var sendAuths []*banktypes.SendAuthorization
	for _, grant := range grants {
		// Check that the expiration is still valid.
		if grant.Expiration == nil || grant.Expiration.After(blocktime) {
			sendAuth, ok := utils.GetAs[*banktypes.SendAuthorization](grant.Authorization.GetCachedValue())
			if !ok {
				return nil, precompile.ErrInvalidGrantType
			}
			sendAuths = append(sendAuths, sendAuth)
		}
	}
	return sendAuths, nil
}
