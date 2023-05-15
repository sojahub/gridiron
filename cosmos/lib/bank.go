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

	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.furychain.dev/gridiron/eth/common"
)

// MintCoinsToAddress mints coins to a given address.
func MintCoinsToAddress(
	ctx sdk.Context,
	bk BankKeeper,
	moduleAcc string,
	recipient common.Address,
	denom string,
	amount *big.Int,
) error {
	// Mint the corresponding bank denom.
	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromBigInt(amount)))
	if err := bk.MintCoins(ctx, moduleAcc, coins); err != nil {
		return err
	}

	// Send the bank denomination to the receipient.
	if err := bk.SendCoinsFromModuleToAccount(ctx, moduleAcc, recipient.Bytes(), coins); err != nil {
		return err
	}

	return nil
}

// BurnCoinsFromAddress burns coins from a given address.
func BurnCoinsFromAddress(
	ctx sdk.Context,
	bk BankKeeper,
	moduleAcc string,
	sender common.Address,
	denom string,
	amount *big.Int,
) error {
	// Burn the corresponding bank denom.
	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromBigInt(amount)))
	if err := bk.SendCoinsFromAccountToModule(ctx, sender.Bytes(), moduleAcc, coins); err != nil {
		return err
	}

	// Burn the bank denomination.
	if err := bk.BurnCoins(ctx, moduleAcc, coins); err != nil {
		return err
	}

	return nil
}
