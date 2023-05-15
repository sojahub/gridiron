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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessList", func() {
	var (
		al *accessList
		a1 = common.BytesToAddress([]byte{1})
		a2 = common.BytesToAddress([]byte{2})
		s1 = common.BytesToHash([]byte{1})
		s2 = common.BytesToHash([]byte{2})
	)

	BeforeEach(func() {
		al = NewAccesslist()
	})

	It("should have the correct registry key", func() {
		Expect(al.RegistryKey()).To(Equal("accessList"))
	})

	It("should support controllable access list operations", func() {
		Expect(al.AddAddress(a1)).To(BeTrue())
		Expect(al.ContainsAddress(a1)).To(BeTrue())
		Expect(al.ContainsAddress(a2)).To(BeFalse())
		al.DeleteAddress(a1)
		Expect(al.ContainsAddress(a1)).To(BeFalse())

		ac, sc := al.AddSlot(a1, s1)
		Expect(ac).To(BeTrue())
		Expect(sc).To(BeTrue())
		ac, sc = al.AddSlot(a1, s2)
		Expect(ac).To(BeFalse())
		Expect(sc).To(BeTrue())

		id := al.Snapshot()

		ac, sc = al.AddSlot(a2, s1)
		Expect(ac).To(BeTrue())
		Expect(sc).To(BeTrue())
		Expect(al.ContainsAddress(a2)).To(BeTrue())

		al.RevertToSnapshot(id)
		Expect(al.ContainsAddress(a2)).To(BeFalse())

		Expect(func() { al.Finalize() }).ToNot(Panic())
		Expect(al.journal.Size()).To(Equal(1))
	})
})
