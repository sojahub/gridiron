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

package bank_test

import (
	"fmt"
	"math/big"
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	libgenerated "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/lib"
	generated "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/precompile/bank"
	cosmlib "pkg.furychain.dev/gridiron/cosmos/lib"
	"pkg.furychain.dev/gridiron/cosmos/precompile"
	"pkg.furychain.dev/gridiron/cosmos/precompile/bank"
	testutil "pkg.furychain.dev/gridiron/cosmos/testing/utils"
	"pkg.furychain.dev/gridiron/cosmos/x/evm/plugins/precompile/log"
	evmtypes "pkg.furychain.dev/gridiron/cosmos/x/evm/types"
	"pkg.furychain.dev/gridiron/eth/common"
	ethprecompile "pkg.furychain.dev/gridiron/eth/core/precompile"
	"pkg.furychain.dev/gridiron/lib/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBankPrecompile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmos/precompile/bank")
}

var _ = Describe("Bank Precompile Test", func() {
	var (
		contract *bank.Contract
		addr     sdk.AccAddress
		factory  *log.Factory
		bk       bankkeeper.BaseKeeper
		ctx      sdk.Context
	)

	BeforeEach(func() {
		ctx, _, bk, _ = testutil.SetupMinimalKeepers()

		contract = utils.MustGetAs[*bank.Contract](bank.NewPrecompileContract(bankkeeper.NewMsgServerImpl(bk), bk))
		addr = sdk.AccAddress([]byte("bank"))

		// Register the events.
		factory = log.NewFactory([]ethprecompile.Registrable{contract})
	})

	It("should register the send event", func() {
		event := sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoin("stg", sdk.NewInt(100)).String()),
		)
		log, err := factory.Build(&event)
		Expect(err).ToNot(HaveOccurred())
		Expect(log.Address).To(Equal(contract.RegistryKey()))
	})

	It("should register the transfer event", func() {
		event := sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoin("stg", sdk.NewInt(100)).String()),
		)
		log, err := factory.Build(&event)
		Expect(err).ToNot(HaveOccurred())
		Expect(log.Address).To(Equal(contract.RegistryKey()))
	})

	It("should register the coin spent event", func() {
		event := sdk.NewEvent(
			banktypes.EventTypeCoinSpent,
			sdk.NewAttribute(banktypes.AttributeKeySpender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoin("stg", sdk.NewInt(100)).String()),
		)
		log, err := factory.Build(&event)
		Expect(err).ToNot(HaveOccurred())
		Expect(log.Address).To(Equal(contract.RegistryKey()))
	})

	It("should register the coin received event", func() {
		event := sdk.NewEvent(
			banktypes.EventTypeCoinReceived,
			sdk.NewAttribute(banktypes.AttributeKeyReceiver, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoin("stg", sdk.NewInt(100)).String()),
		)
		log, err := factory.Build(&event)
		Expect(err).ToNot(HaveOccurred())
		Expect(log.Address).To(Equal(contract.RegistryKey()))
	})

	It("should register the burn event", func() {
		event := sdk.NewEvent(
			banktypes.EventTypeCoinBurn,
			sdk.NewAttribute(banktypes.AttributeKeyBurner, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoin("stg", sdk.NewInt(100)).String()),
		)
		log, err := factory.Build(&event)
		Expect(err).ToNot(HaveOccurred())
		Expect(log.Address).To(Equal(contract.RegistryKey()))
	})

	When("Calling Precompile Methods", func() {
		var (
			acc    sdk.AccAddress
			caller common.Address
		)

		denom := "afury"

		When("GetBalance", func() {
			It("should fail if input address is not a common.Address", func() {
				res, err := contract.GetBalance(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"0x",
					"stake",
				)
				Expect(err).To(MatchError(precompile.ErrInvalidHexAddress))
				Expect(res).To(BeNil())
			})

			It("should fail if input denom is not a valid string", func() {
				res, err := contract.GetBalance(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					666,
				)
				Expect(err).To(MatchError(precompile.ErrInvalidString))
				Expect(res).To(BeNil())
			})

			It("should fail if input denom is not a valid denom", func() {
				res, err := contract.GetBalance(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					"_invalid_denom",
				)
				// reDnmString = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
				Expect(err).To(HaveOccurred())
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())

				acc = simtestutil.CreateRandomAccounts(1)[0]

				err := FundAccount(
					ctx,
					bk,
					acc,
					sdk.NewCoins(
						sdk.NewCoin(
							denom,
							sdk.NewIntFromBigInt(balanceAmount),
						),
					),
				)
				Expect(err).ToNot(HaveOccurred())

				res, err := contract.GetBalance(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					denom,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(res[0]).To(Equal(balanceAmount))
			})
		})

		When("GetAllBalance", func() {
			It("should fail if input address is not a common.Address", func() {
				res, err := contract.GetBalance(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"0x",
				)
				Expect(err).To(MatchError(precompile.ErrInvalidHexAddress))
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				numOfDenoms := 3
				acc = simtestutil.CreateRandomAccounts(1)[0]
				for i := 0; i < numOfDenoms; i++ {
					balanceAmountStr := fmt.Sprintf("%d000000000000000000", i+1)
					balanceAmount, ok := new(big.Int).SetString(balanceAmountStr, 10)
					Expect(ok).To(BeTrue())

					err := FundAccount(
						ctx,
						bk,
						acc,
						sdk.NewCoins(
							sdk.NewCoin(
								fmt.Sprintf("denom_%d", i+1),
								sdk.NewIntFromBigInt(balanceAmount),
							),
						),
					)
					Expect(err).ToNot(HaveOccurred())
				}

				res, err := contract.GetAllBalances(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
				)
				Expect(err).ToNot(HaveOccurred())

				coins, ok := utils.GetAs[[]libgenerated.CosmosCoin](res[0])
				Expect(ok).To(BeTrue())

				for i, coin := range coins {
					balanceAmountStr := fmt.Sprintf("%d000000000000000000", i+1)
					balanceAmount, ok2 := new(big.Int).SetString(balanceAmountStr, 10)
					Expect(ok2).To(BeTrue())

					Expect(coin.Denom).To(Equal(fmt.Sprintf("denom_%d", i+1)))
					Expect(coin.Amount).To(Equal(balanceAmount))
				}
			})
		})

		When("GetSpendableBalanceByDenom", func() {
			It("should fail if input address is not a common.Address", func() {
				res, err := contract.GetSpendableBalanceByDenom(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"0x",
					"stake",
				)
				Expect(err).To(MatchError(precompile.ErrInvalidHexAddress))
				Expect(res).To(BeNil())
			})

			It("should fail if input denom is not a valid string", func() {
				res, err := contract.GetSpendableBalanceByDenom(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					666,
				)
				Expect(err).To(MatchError(precompile.ErrInvalidString))
				Expect(res).To(BeNil())
			})

			It("should fail if input denom is not a valid denom", func() {
				res, err := contract.GetSpendableBalanceByDenom(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					"_invalid_denom",
				)
				// reDnmString = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
				Expect(err).To(HaveOccurred())
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())
				// todo: use vesting accounts, lock some tokens
				acc = simtestutil.CreateRandomAccounts(1)[0]

				err := FundAccount(
					ctx,
					bk,
					acc,
					sdk.NewCoins(
						sdk.NewCoin(
							denom,
							sdk.NewIntFromBigInt(balanceAmount),
						),
					),
				)
				Expect(err).ToNot(HaveOccurred())

				res, err := contract.GetSpendableBalanceByDenom(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					denom,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(res[0]).To(Equal(balanceAmount))
			})
		})

		When("GetSpendableBalances", func() {
			It("should fail if input address is not a common.Address", func() {
				res, err := contract.GetSpendableBalances(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"0x",
				)
				Expect(err).To(MatchError(precompile.ErrInvalidHexAddress))
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				numOfDenoms := 3
				// todo: use vesting accounts, lock some tokens
				acc = simtestutil.CreateRandomAccounts(1)[0]
				for i := 0; i < numOfDenoms; i++ {
					balanceAmountStr := fmt.Sprintf("%d000000000000000000", i+1)
					balanceAmount, ok := new(big.Int).SetString(balanceAmountStr, 10)
					Expect(ok).To(BeTrue())

					err := FundAccount(
						ctx,
						bk,
						acc,
						sdk.NewCoins(
							sdk.NewCoin(
								fmt.Sprintf("denom_%d", i+1),
								sdk.NewIntFromBigInt(balanceAmount),
							),
						),
					)
					Expect(err).ToNot(HaveOccurred())
				}

				res, err := contract.GetSpendableBalances(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
				)
				Expect(err).ToNot(HaveOccurred())

				coins, ok := utils.GetAs[[]libgenerated.CosmosCoin](res[0])
				Expect(ok).To(BeTrue())

				for i, coin := range coins {
					balanceAmountStr := fmt.Sprintf("%d000000000000000000", i+1)
					balanceAmount, ok2 := new(big.Int).SetString(balanceAmountStr, 10)
					Expect(ok2).To(BeTrue())

					Expect(coin.Denom).To(Equal(fmt.Sprintf("denom_%d", i+1)))
					Expect(coin.Amount).To(Equal(balanceAmount))
				}
			})
		})

		When("GetSupplyOf", func() {
			It("should fail if input denom is not a valid string", func() {
				res, err := contract.GetSupplyOf(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					666,
				)
				Expect(err).To(MatchError(precompile.ErrInvalidString))
				Expect(res).To(BeNil())
			})

			It("should fail if input denom is not a valid Denom", func() {
				res, err := contract.GetSupplyOf(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"_invalid_denom",
				)
				// fmt.Errorf("invalid denom: %s", denom)
				Expect(err).To(HaveOccurred())
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())
				balanceAmount3, ok := new(big.Int).SetString("66000000000000000000", 10)
				Expect(ok).To(BeTrue())

				accs := simtestutil.CreateRandomAccounts(3)

				for i := 0; i < 3; i++ {
					err := FundAccount(
						ctx,
						bk,
						accs[i],
						sdk.NewCoins(
							sdk.NewCoin(
								denom,
								sdk.NewIntFromBigInt(balanceAmount),
							),
						),
					)
					Expect(err).ToNot(HaveOccurred())
				}

				res, err := contract.GetSupplyOf(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					denom,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(res[0]).To(Equal(balanceAmount3))
			})
		})

		When("GetTotalSupply", func() {
			It("should succeed", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())
				balanceAmount3, ok := new(big.Int).SetString("66000000000000000000", 10)
				Expect(ok).To(BeTrue())

				accs := simtestutil.CreateRandomAccounts(3)
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						err := FundAccount(
							ctx,
							bk,
							accs[i],
							sdk.NewCoins(
								sdk.NewCoin(
									fmt.Sprintf("%s%d", denom, j),
									sdk.NewIntFromBigInt(balanceAmount),
								),
							),
						)
						Expect(err).ToNot(HaveOccurred())
					}
				}

				res, err := contract.GetTotalSupply(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
				)
				Expect(err).ToNot(HaveOccurred())

				coins, ok := utils.GetAs[[]libgenerated.CosmosCoin](res[0])
				Expect(ok).To(BeTrue())

				for i := 0; i < 3; i++ {
					Expect(coins[i].Denom).To(Equal(fmt.Sprintf("%s%d", denom, i)))
					Expect(coins[i].Amount).To(Equal(balanceAmount3))
				}

			})
		})

		When("GetDenomMetadata", func() {
			It("should fail if input denom is not a valid string", func() {
				res, err := contract.GetDenomMetadata(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					666,
				)
				Expect(err).To(MatchError(precompile.ErrInvalidString))
				Expect(res).To(BeNil())
			})

			It("should fail if input denom is not a valid Denom", func() {
				res, err := contract.GetDenomMetadata(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"_invalid_denom",
				)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				expectedResult := generated.IBankModuleDenomMetadata{
					Name:        "Furychain fury",
					Symbol:      "FURY",
					Description: "The Fury.",
					DenomUnits: []generated.IBankModuleDenomUnit{
						{Denom: "fury", Exponent: uint32(0), Aliases: []string{"fury"}},
						{Denom: "nfury", Exponent: uint32(9), Aliases: []string{"nanofury"}},
						{Denom: "afury", Exponent: uint32(18), Aliases: []string{"attofury"}},
					},
					Base:    "afury",
					Display: "fury",
				}

				metadata := getTestMetadata()
				bk.SetDenomMetaData(ctx, metadata[0])

				res, err := contract.GetDenomMetadata(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					metadata[0].Base,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(res[0]).To(Equal(expectedResult))
			})
		})

		When("GetSendEnabled", func() {
			It("should succeed", func() {
				enabledDenom := "enabledDenom"
				// disabledDenom := "disabledDenom"

				bk.SetSendEnabled(ctx, enabledDenom, true)

				res, err := contract.GetSendEnabled(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					enabledDenom,
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(res[0]).To(BeTrue())
			})
		})

		When("Send", func() {
			It("should fail if from address is not a common.Address", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())

				acc = simtestutil.CreateRandomAccounts(1)[0]

				res, err := contract.Send(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					"0x",
					cosmlib.AccAddressToEthAddress(acc),
					sdk.NewCoins(
						sdk.NewCoin(
							denom,
							sdk.NewIntFromBigInt(balanceAmount),
						),
					),
				)
				Expect(err).To(MatchError(precompile.ErrInvalidHexAddress))
				Expect(res).To(BeNil())
			})

			It("should fail if to address is not a common.Address", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())

				acc = simtestutil.CreateRandomAccounts(1)[0]

				res, err := contract.Send(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(acc),
					"0x",
					sdk.NewCoins(
						sdk.NewCoin(
							denom,
							sdk.NewIntFromBigInt(balanceAmount),
						),
					),
				)
				Expect(err).To(MatchError(precompile.ErrInvalidHexAddress))
				Expect(res).To(BeNil())
			})

			It("should fail if amount is not sdk.Coins", func() {
				accs := simtestutil.CreateRandomAccounts(2)
				fromAcc, toAcc := accs[0], accs[1]

				res, err := contract.Send(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(fromAcc),
					cosmlib.AccAddressToEthAddress(toAcc),
					"wrong type input",
				)
				Expect(err).To(MatchError(precompile.ErrInvalidCoin))
				Expect(res).To(BeNil())
			})

			It("should succeed", func() {
				balanceAmount, ok := new(big.Int).SetString("22000000000000000000", 10)
				Expect(ok).To(BeTrue())

				accs := simtestutil.CreateRandomAccounts(2)
				fromAcc, toAcc := accs[0], accs[1]

				coins := sdk.NewCoins(
					sdk.NewCoin(
						denom,
						sdk.NewIntFromBigInt(balanceAmount),
					),
				)

				err := FundAccount(
					ctx,
					bk,
					fromAcc,
					coins,
				)
				Expect(err).ToNot(HaveOccurred())

				bk.SetSendEnabled(ctx, denom, true)

				_, err = contract.Send(
					ctx,
					nil,
					caller,
					big.NewInt(0),
					true,
					cosmlib.AccAddressToEthAddress(fromAcc),
					cosmlib.AccAddressToEthAddress(toAcc),
					sdkCoinsToEvmCoins(coins),
				)
				Expect(err).ToNot(HaveOccurred())

				balance, err := bk.Balance(ctx, banktypes.NewQueryBalanceRequest(toAcc, denom))
				Expect(err).ToNot(HaveOccurred())

				Expect(*balance.Balance).To(Equal(coins[0]))
			})
		})
	})
})

func FundAccount(ctx sdk.Context, bk bankkeeper.BaseKeeper, account sdk.AccAddress, coins sdk.Coins) error {
	if err := bk.MintCoins(ctx, evmtypes.ModuleName, coins); err != nil {
		return err
	}
	return bk.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, account, coins)
}

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
	}
}

func sdkCoinsToEvmCoins(sdkCoins sdk.Coins) []struct {
	Amount *big.Int `json:"amount"`
	Denom  string   `json:"denom"`
} {
	evmCoins := make([]struct {
		Amount *big.Int `json:"amount"`
		Denom  string   `json:"denom"`
	}, len(sdkCoins))
	for i, coin := range sdkCoins {
		evmCoins[i] = struct {
			Amount *big.Int `json:"amount"`
			Denom  string   `json:"denom"`
		}{
			Amount: coin.Amount.BigInt(),
			Denom:  coin.Denom,
		}
	}
	return evmCoins
}
