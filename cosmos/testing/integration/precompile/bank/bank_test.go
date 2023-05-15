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
	"math/big"
	"testing"

	bindings "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/precompile/bank"
	tbindings "pkg.furychain.dev/gridiron/contracts/bindings/testing/fundraiser"
	"pkg.furychain.dev/gridiron/cosmos/testing/integration"
	"pkg.furychain.dev/gridiron/eth/common"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "pkg.furychain.dev/gridiron/cosmos/testing/integration/utils"
)

func TestCosmosPrecompiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmos/testing/integration/precompile/bank")
}

var (
	tf             *integration.TestFixture
	bankPrecompile *bindings.BankModule
)

var _ = SynchronizedBeforeSuite(func() []byte {
	// Setup the network and clients here.
	tf = integration.NewTestFixture(GinkgoT())
	bankPrecompile, _ = bindings.NewBankModule(
		common.HexToAddress("0x4381dC2aB14285160c808659aEe005D51255adD7"), tf.EthClient)
	return nil
}, func(data []byte) {})

var _ = Describe("Bank", func() {
	denom := "afury"
	denom2 := "atoken"
	denom3 := "stake"

	It("should call functions on the precompile directly", func() {
		numberOfDenoms := 7
		coinsToBeSent := []bindings.CosmosCoin{
			{
				Denom:  denom,
				Amount: big.NewInt(1000),
			},
		}
		expectedAllBalance := []bindings.CosmosCoin{
			{
				Denom:  denom,
				Amount: big.NewInt(100),
			},
			{
				Denom:  denom2,
				Amount: big.NewInt(100),
			},
			{
				Denom:  denom3,
				Amount: big.NewInt(1000000000000000000),
			},
		}
		evmDenomMetadata := bindings.IBankModuleDenomMetadata{
			Name:        "Furychain fury",
			Symbol:      "FURY",
			Description: "The Fury.",
			DenomUnits: []bindings.IBankModuleDenomUnit{
				{Denom: "fury", Exponent: uint32(0), Aliases: []string{"fury"}},
				{Denom: "nfury", Exponent: uint32(9), Aliases: []string{"nanofury"}},
				{Denom: "afury", Exponent: uint32(18), Aliases: []string{"attofury"}},
			},
			Base:    "afury",
			Display: "fury",
		}

		// charlie initially has 1000000000 afury
		balance, err := bankPrecompile.GetBalance(nil, tf.Address("charlie"), denom)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(balance.Cmp(big.NewInt(1000000000))).To(Equal(0))

		// Send 1000 fury from alice to charlie
		_, err = bankPrecompile.Send(
			tf.GenerateTransactOpts("alice"),
			tf.Address("alice"),
			tf.Address("charlie"),
			coinsToBeSent,
		)
		Expect(err).ShouldNot(HaveOccurred())

		// Wait one block.
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())

		// charlie now has 1000001000 afury
		balance, err = bankPrecompile.GetBalance(nil, tf.Address("charlie"), denom)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(balance).To(Equal(big.NewInt(1000001000)))

		// bob has 100 afury and 100 atoken
		allBalance, err := bankPrecompile.GetAllBalances(nil, tf.Address("bob"))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(allBalance).To(Equal(expectedAllBalance))

		spendableBalanceByDenom, err := bankPrecompile.GetSpendableBalance(nil, tf.Address("bob"), denom)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(spendableBalanceByDenom).To(Equal(big.NewInt(100)))

		spendableBalances, err := bankPrecompile.GetAllSpendableBalances(nil, tf.Address("bob"))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(spendableBalances).To(Equal(expectedAllBalance))

		atokenSupply, err := bankPrecompile.GetSupply(nil, denom2)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(atokenSupply).To(Equal(big.NewInt(100)))

		totalSupply, err := bankPrecompile.GetAllSupply(nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(totalSupply).To(HaveLen(numberOfDenoms))

		denomMetadata, err := bankPrecompile.GetDenomMetadata(nil, denom)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(denomMetadata).To(Equal(evmDenomMetadata))

		sendEnabled, err := bankPrecompile.GetSendEnabled(nil, "afury")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(sendEnabled).To(BeTrue())
	})

	It("should be able to call a precompile from a smart contract", func() {
		// deploy fundraiser contract with account 0
		contractAddr, tx, contract, err := tbindings.DeployFundraiser(
			tf.GenerateTransactOpts("alice"),
			tf.EthClient,
		)
		Expect(err).ToNot(HaveOccurred())
		ExpectSuccessReceipt(tf.EthClient, tx)

		coinsToDonate := []tbindings.CosmosCoin{
			{
				Denom:  denom,
				Amount: big.NewInt(1000000),
			},
		}

		// donate 1000000 afury from account 0 to contractAddr
		_, err = contract.Donate(tf.GenerateTransactOpts("alice"), coinsToDonate)
		Expect(err).ToNot(HaveOccurred())

		// Wait one block.
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())

		// contractAddr should have 1000000 afury
		balance, err := bankPrecompile.GetBalance(nil, contractAddr, denom)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(balance).To(Equal(big.NewInt(1000000)))

		// withdraw all 1000000 afury from contractAddr to account 0
		_, err = contract.WithdrawDonations(tf.GenerateTransactOpts("alice"))
		Expect(err).ToNot(HaveOccurred())

		// Wait one block.
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())

		// contractAddr should have 0 afury
		balance, err = bankPrecompile.GetBalance(nil, contractAddr, denom)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(balance.Cmp(big.NewInt(0))).To(Equal(0))
	})
})
