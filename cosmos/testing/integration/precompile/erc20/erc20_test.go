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

package erc20_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	cbindings "pkg.furychain.dev/gridiron/contracts/bindings/cosmos"
	bbindings "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/precompile/bank"
	bindings "pkg.furychain.dev/gridiron/contracts/bindings/cosmos/precompile/erc20"
	tbindings "pkg.furychain.dev/gridiron/contracts/bindings/testing"
	cosmlib "pkg.furychain.dev/gridiron/cosmos/lib"
	"pkg.furychain.dev/gridiron/cosmos/testing/integration"
	erc20types "pkg.furychain.dev/gridiron/cosmos/x/erc20/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "pkg.furychain.dev/gridiron/cosmos/testing/integration/utils"
)

func TestERC20Precompile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmos/testing/integration/precompile/erc20")
}

var (
	tf                 *integration.TestFixture
	erc20Precompile    *bindings.ERC20Module
	bankPrecompile     *bbindings.BankModule
	erc20ModuleAddress = common.HexToAddress("0x696969")
	// cosmlib.AccAddressToEthAddress(
	// 	authtypes.NewModuleAddress(erc20types.ModuleName),
	// ).
)

var _ = SynchronizedBeforeSuite(func() []byte {
	// Setup the network and clients here.
	tf = integration.NewTestFixture(GinkgoT())
	bankPrecompile, _ = bbindings.NewBankModule(
		common.HexToAddress("0x4381dC2aB14285160c808659aEe005D51255adD7"), tf.EthClient,
	)
	erc20Precompile, _ = bindings.NewERC20Module(erc20ModuleAddress, tf.EthClient)
	return nil
}, func(data []byte) {})

var _ = Describe("ERC20", func() {
	Describe("calling the erc20 precompile directly", func() {
		When("calling read-only methods", func() {
			It("should handle empty inputs", func() {
				// nonexistent address
				denom, err := erc20Precompile.CoinDenomForERC20Address(nil, common.Address{})
				Expect(err).ToNot(HaveOccurred())
				Expect(denom).To(Equal(""))

				// invalid address
				_, err = erc20Precompile.CoinDenomForERC20Address0(nil, "")
				Expect(err).To(HaveOccurred())

				// nonexistent denom
				token, err := erc20Precompile.Erc20AddressForCoinDenom(nil, "")
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal(common.Address{}))
			})

			It("should handle non-empty inputs", func() {
				token, err := erc20Precompile.Erc20AddressForCoinDenom(nil, "afury")
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal(common.Address{}))

				tokenAddr := common.BytesToAddress([]byte("afury"))
				tokenBech32 := cosmlib.AddressToAccAddress(tokenAddr).String()

				denom, err := erc20Precompile.CoinDenomForERC20Address(nil, tokenAddr)
				Expect(err).ToNot(HaveOccurred())
				Expect(denom).To(Equal(""))

				denom, err = erc20Precompile.CoinDenomForERC20Address0(nil, tokenBech32)
				Expect(err).ToNot(HaveOccurred())
				Expect(denom).To(Equal(""))
			})
		})

		When("calling write methods", func() {
			It("should error on non-existent denoms/tokens", func() {
				// user does not have balance of bOSMO
				_, err := erc20Precompile.TransferCoinToERC20(
					tf.GenerateTransactOpts("alice"),
					"bOSMO",
					big.NewInt(123456789),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("insufficient funds"))

				// token doesn't exist, user does not have balance of token
				_, err = erc20Precompile.TransferERC20ToCoin(
					tf.GenerateTransactOpts("alice"),
					common.HexToAddress("0x432423432489230"),
					big.NewInt(123456789),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ERC20 token contract does not exist"))
			})

			It("should handle a IBC-originated SDK coin", func() {
				// initial balance of bOSMO = 12345 * 2

				// denom already exists, create new token
				tx, err := erc20Precompile.TransferCoinToERC20(
					tf.GenerateTransactOpts("alice"),
					"bOSMO",
					big.NewInt(12345),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)

				// check that the new ERC20 is minted to TestAddress
				tokenAddr, err := erc20Precompile.Erc20AddressForCoinDenom(nil, "bOSMO")
				Expect(err).ToNot(HaveOccurred())
				token, err := cbindings.NewGridironERC20(tokenAddr, tf.EthClient)
				Expect(err).ToNot(HaveOccurred())
				balance, err := token.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(balance).To(Equal(big.NewInt(12345 * 2)))

				// denom already exists, token already exists
				tx, err = erc20Precompile.TransferCoinToERC20(
					tf.GenerateTransactOpts("alice"),
					"bOSMO",
					big.NewInt(12345),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)

				// check that the new ERC20 is minted to TestAddress
				balance, err = token.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(balance).To(Equal(big.NewInt(12345 * 2)))

				// convert back to SDK coin
				tx, err = erc20Precompile.TransferERC20ToCoin(
					tf.GenerateTransactOpts("alice"),
					tokenAddr,
					big.NewInt(12345),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)

				// check that the new ERC20 is burned from TestAddress
				balance, err = token.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(balance).To(Equal(big.NewInt(12345 * 2)))

				// convert illegal amount back to SDK coin
				_, err = erc20Precompile.TransferERC20ToCoin(
					tf.GenerateTransactOpts("alice"),
					tokenAddr,
					big.NewInt(12345*2+1),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("insufficient funds"))
			})

			It("should handle a ERC20 originated token", func() {
				// originate a ERC20 token
				contract, token := DeployERC20(tf.GenerateTransactOpts("alice"), tf.EthClient)

				// mint some tokens to the test address
				tx, err := contract.Mint(
					tf.GenerateTransactOpts("alice"), tf.Address("alice"), big.NewInt(123456789),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)
				// check that the new ERC20 is minted to TestAddress
				bal, err := contract.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(bal).To(Equal(big.NewInt(123456789)))

				// token already exists, create new Gridiron denom
				_, err = erc20Precompile.TransferERC20ToCoin(
					tf.GenerateTransactOpts("alice"),
					token,
					big.NewInt(6789),
				)
				Expect(err).To(HaveOccurred())
				// doesn't work because owner did not approve caller to spend tokens
				Expect(err.Error()).To(ContainSubstring("gas required exceeds allowance"))

				// verify the transfer did not work
				bal, err = contract.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(bal).To(Equal(big.NewInt(123456789)))
				bankBal, err := bankPrecompile.GetBalance(
					nil, tf.Address("alice"), erc20types.NewGridironDenomForAddress(token),
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(bankBal.Cmp(big.NewInt(0))).To(Equal(0))
				denom, err := erc20Precompile.CoinDenomForERC20Address(nil, token)
				Expect(err).ToNot(HaveOccurred())
				Expect(denom).To(BeEmpty())

				// approve caller to spend tokens
				tx, err = contract.Approve(
					tf.GenerateTransactOpts("alice"),
					tf.Address("alice"),
					big.NewInt(6789),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)

				// token already exists, create new Gridiron denom
				_, err = erc20Precompile.TransferERC20ToCoin(
					tf.GenerateTransactOpts("alice"),
					token,
					big.NewInt(6789),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)

				err = tf.Network.WaitForNextBlock()
				Expect(err).ToNot(HaveOccurred())

				// check that the new ERC20 is transferred from TestAddress to precompile (escrow)
				bal, err = contract.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(bal).To(Equal(big.NewInt(123450000)))
				bal, err = contract.BalanceOf(nil, erc20ModuleAddress)
				Expect(err).ToNot(HaveOccurred())
				Expect(bal).To(Equal(big.NewInt(6789)))

				// check that the Gridiron coin is minted to TestAddress
				denom, err = erc20Precompile.CoinDenomForERC20Address(nil, token)
				Expect(err).ToNot(HaveOccurred())
				bankBal, err = bankPrecompile.GetBalance(nil, tf.Address("alice"), denom)
				Expect(err).ToNot(HaveOccurred())
				Expect(bankBal.Cmp(big.NewInt(6789))).To(Equal(0))

				// convert back to ERC20 token
				_, err = erc20Precompile.TransferCoinToERC20(
					tf.GenerateTransactOpts("alice"),
					denom,
					big.NewInt(6790),
				)
				Expect(err).To(HaveOccurred()) // not enough funds

				// convert back to ERC20 token
				tx, err = erc20Precompile.TransferCoinToERC20(
					tf.GenerateTransactOpts("alice"),
					denom,
					big.NewInt(6789),
				)
				Expect(err).ToNot(HaveOccurred())
				ExpectSuccessReceipt(tf.EthClient, tx)

				err = tf.Network.WaitForNextBlock()
				Expect(err).ToNot(HaveOccurred())

				// check that Gridiron Coin is converted back to ERC20
				bal, err = contract.BalanceOf(nil, tf.Address("alice"))
				Expect(err).ToNot(HaveOccurred())
				Expect(bal.Cmp(big.NewInt(123456789))).To(Equal(0))
				bal, err = contract.BalanceOf(nil, erc20ModuleAddress)
				Expect(err).ToNot(HaveOccurred())
				Expect(bal.Cmp(big.NewInt(0))).To(Equal(0))
				bankBal, err = bankPrecompile.GetBalance(nil, tf.Address("alice"), denom)
				Expect(err).ToNot(HaveOccurred())
				Expect(bankBal.Cmp(big.NewInt(0))).To(Equal(0))
			})
		})
	})

	Describe("calling the erc20 precompile via the another contract", func() {
		It("should work", func() {
			_, tx, swapper, err := tbindings.DeploySwapper(tf.GenerateTransactOpts("alice"), tf.EthClient)
			Expect(err).ToNot(HaveOccurred())
			ExpectSuccessReceipt(tf.EthClient, tx)

			// check that the new ERC20 is minted to TestAddress
			tokenAddr, err := swapper.GetGridironERC20(nil, "bAKT")
			Expect(err).ToNot(HaveOccurred())
			Expect(tokenAddr.Bytes()).To(Equal(common.Address{}.Bytes()))

			err = tf.Network.WaitForNextBlock()
			Expect(err).ToNot(HaveOccurred())

			tx, err = swapper.Swap(
				tf.GenerateTransactOpts("alice"),
				"bAKT",
				big.NewInt(12345),
			)
			Expect(err).ToNot(HaveOccurred())
			ExpectSuccessReceipt(tf.EthClient, tx)

			// check that the new ERC20 is minted to TestAddress
			tokenAddr, err = swapper.GetGridironERC20(nil, "bAKT")
			Expect(err).ToNot(HaveOccurred())
			token, err := cbindings.NewGridironERC20(tokenAddr, tf.EthClient)
			Expect(err).ToNot(HaveOccurred())
			balance, err := token.BalanceOf(nil, tf.Address("alice"))
			Expect(err).ToNot(HaveOccurred())
			Expect(balance).To(Equal(big.NewInt(12345)))

			tx, err = swapper.Swap0(
				tf.GenerateTransactOpts("alice"),
				tokenAddr,
				big.NewInt(345),
			)
			Expect(err).ToNot(HaveOccurred())
			ExpectSuccessReceipt(tf.EthClient, tx)

			// check that the new ERC20 is burned from TestAddress
			balance, err = token.BalanceOf(nil, tf.Address("alice"))
			Expect(err).ToNot(HaveOccurred())
			Expect(balance).To(Equal(big.NewInt(12345)))
		})
	})
})
