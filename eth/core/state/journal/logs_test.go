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

package journal

import (
	"pkg.furychain.dev/gridiron/eth/common"
	coretypes "pkg.furychain.dev/gridiron/eth/core/types"
	"pkg.furychain.dev/gridiron/lib/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	var l *logs
	var thash = common.BytesToHash([]byte{1})
	var ti = uint(1)
	var bnum = uint64(2)
	var bhash = common.BytesToHash([]byte{2})
	var a1 = common.BytesToAddress([]byte{3})
	var a2 = common.BytesToAddress([]byte{4})

	BeforeEach(func() {
		l = utils.MustGetAs[*logs](NewLogs())
		l.SetTxContext(thash, int(ti))
		Expect(l.Capacity()).To(Equal(32))
	})

	It("should have the correct registry key", func() {
		Expect(l.RegistryKey()).To(Equal("logs"))
	})

	When("adding logs", func() {
		BeforeEach(func() {
			l.AddLog(&coretypes.Log{Address: a1})
			Expect(l.Size()).To(Equal(1))
			Expect(l.PeekAt(0).Address).To(Equal(a1))
			Expect(l.PeekAt(0).TxHash).To(Equal(thash))
			Expect(l.PeekAt(0).TxIndex).To(Equal(ti))
		})

		It("should correctly snapshot and revert", func() {
			id := l.Snapshot()

			l.AddLog(&coretypes.Log{Address: a2})
			Expect(l.Size()).To(Equal(2))
			Expect(l.PeekAt(1).Address).To(Equal(a2))

			l.RevertToSnapshot(id)
			Expect(l.Size()).To(Equal(1))
		})

		It("should correctly get logs", func() {
			logs := l.Logs()
			Expect(logs).To(HaveLen(1))
			Expect(logs[0].TxHash).To(Equal(thash))
			Expect(logs[0].BlockHash).To(Equal(common.Hash{}))
			Expect(logs[0].BlockNumber).To(Equal(uint64(0)))

			logs = l.GetLogs(thash, bnum, bhash)
			Expect(logs).To(HaveLen(1))
			Expect(logs[0].BlockHash).To(Equal(bhash))
			Expect(logs[0].BlockNumber).To(Equal(bnum))
		})

		It("should corrctly finalize", func() {
			Expect(func() { l.Finalize() }).ToNot(Panic())
		})
	})
})
