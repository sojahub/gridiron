// SPDX-License-Identifier: Apache-2.0
//

package mock

import libtypes "pkg.furychain.dev/gridiron/lib/types"

// Assert that `MockRegistrable` implements `Registrable`.
var _ libtypes.Registrable[string] = &Registrable{}

type Registrable struct {
	registerKey string
	data        string
}

func NewMockRegistrable(registerKey string, data string) *Registrable {
	return &Registrable{
		registerKey: registerKey,
		data:        data,
	}
}

func (m Registrable) RegistryKey() string {
	return m.registerKey
}

func (m Registrable) Data() string {
	return m.data
}
