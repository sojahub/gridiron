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

package distribution_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bindings "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/precompile/distribution"
	sbindings "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/precompile/staking"
	tbindings "pkg.furychain.dev/gridiron/contracts/bindings/testing"
	cosmlib "pkg.furychain.dev/gridiron/cosmos/lib"
	"pkg.furychain.dev/gridiron/cosmos/testing/integration"
	"pkg.furychain.dev/gridiron/eth/common"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "pkg.furychain.dev/gridiron/cosmos/testing/integration/utils"
)

func TestDistributionPrecompile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmos/testing/integration/precompile/distribution")
}

var (
	tf                *integration.TestFixture
	precompile        *bindings.DistributionModule
	stakingPrecompile *sbindings.StakingModule
	validator         common.Address
)

var _ = SynchronizedBeforeSuite(func() []byte {
	// Setup the network and clients here.
	tf = integration.NewTestFixture(GinkgoT())
	// Setup the governance precompile.
	precompile, _ = bindings.NewDistributionModule(
		common.HexToAddress("0x69"),
		tf.EthClient,
	)
	// Setup the staking precompile.
	stakingPrecompile, _ = sbindings.NewStakingModule(
		common.HexToAddress("0xd9A998CaC66092748FfEc7cFBD155Aae1737C2fF"), tf.EthClient)
	// Set the validator address.
	validator = common.Address(tf.Network.Validators[0].Address.Bytes())
	return nil
}, func(data []byte) {})

var _ = Describe("Distribution Precompile", func() {
	It("should be able to get if withdraw address is enabled", func() {
		res, err := precompile.GetWithdrawEnabled(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(BeTrue())
	})
	It("should be able to set withdraw address with cosmos address", func() {
		addr := sdk.AccAddress("addr")
		txr := tf.GenerateTransactOpts("alice")
		tx, err := precompile.SetWithdrawAddress0(txr, addr.String())
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)
	})
	It("should be able to set withdraw address with ethereum address", func() {
		addr := sdk.AccAddress("addr")
		ethAddr := cosmlib.AccAddressToEthAddress(addr)
		txr := tf.GenerateTransactOpts("alice")
		tx, err := precompile.SetWithdrawAddress(txr, ethAddr)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)
	})
	It("should be able to get delegator reward", func() {
		// Delegate some tokens to an active validator.
		validators, err := stakingPrecompile.GetActiveValidators(nil)
		Expect(err).ToNot(HaveOccurred())
		val := validators[0]
		delegateAmt := big.NewInt(123450000000)
		txr := tf.GenerateTransactOpts("alice")
		txr.Value = delegateAmt
		tx, err := stakingPrecompile.Delegate(txr, val, delegateAmt)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)

		// Wait for the 2 block to be produced, to make sure there are rewards.
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())

		// Withdraw the rewards.
		txr = tf.GenerateTransactOpts("alice")
		tx, err = precompile.WithdrawDelegatorReward(txr, tf.Address("alice"), val)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)
	})

	It("Should be able to call the precompile via the contract", func() {
		// Deploy the contract.
		contractAddress, tx, contract, err := tbindings.DeployDistributionWrapper(
			tf.GenerateTransactOpts("alice"),
			tf.EthClient,
			common.HexToAddress("0x69"),
			common.HexToAddress("0xd9A998CaC66092748FfEc7cFBD155Aae1737C2fF"),
		)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)

		// Delegate some tokens to a validator.
		validators, err := stakingPrecompile.GetActiveValidators(nil)
		Expect(err).ToNot(HaveOccurred())
		val := validators[0]
		amt := big.NewInt(123450000000)
		txr := tf.GenerateTransactOpts("alice")
		txr.Value = amt
		tx, err = contract.Delegate(txr, val)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)

		// Wait for the 2 block to be produced, to make sure there are rewards.
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())
		err = tf.Network.WaitForNextBlock()
		Expect(err).ToNot(HaveOccurred())

		// Withdraw the rewards.
		txr = tf.GenerateTransactOpts("alice")
		tx, err = contract.WithdrawRewards(txr, contractAddress, val)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
		ExpectSuccessReceipt(tf.EthClient, tx)

		// Get withdraw address enabled.
		res, err := contract.GetWithdrawEnabled(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(BeTrue())

		// Set the withdraw address.
		addr := sdk.AccAddress("addr")
		ethAddr := cosmlib.AccAddressToEthAddress(addr)
		txr = tf.GenerateTransactOpts("alice")
		tx, err = contract.SetWithdrawAddress(txr, ethAddr)
		Expect(err).ToNot(HaveOccurred())
		ExpectMined(tf.EthClient, tx)
	})
})
