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

package state

import (
	"pkg.furychain.dev/gridiron/eth/common"
	"pkg.furychain.dev/gridiron/eth/core/state/journal"
	coretypes "pkg.furychain.dev/gridiron/eth/core/types"
	"pkg.furychain.dev/gridiron/eth/core/vm"
	"pkg.furychain.dev/gridiron/eth/params"
	"pkg.furychain.dev/gridiron/lib/snapshot"
	libtypes "pkg.furychain.dev/gridiron/lib/types"
)

// stateDB is a struct that holds the plugins and controller to manage Ethereum state.
type stateDB struct {
	// Plugin is injected by the chain running the Gridiron EVM.
	Plugin

	// Journals built internally and required for the stateDB.
	LogsJournal
	RefundJournal
	AccessListJournal
	SuicidesJournal
	TransientStorageJournal

	// ctrl is used to manage snapshots and reverts across plugins and journals.
	ctrl libtypes.Controller[string, libtypes.Controllable[string]]
}

// NewStateDB returns a `vm.GridironStateDB` with the given `StatePlugin`.
func NewStateDB(sp Plugin) vm.GridironStateDB {
	// Build the journals required for the stateDB
	lj := journal.NewLogs()
	rj := journal.NewRefund()
	aj := journal.NewAccesslist()
	sj := journal.NewSuicides(sp)
	tj := journal.NewTransientStorage()

	// Build the controller and register the plugins and journals
	ctrl := snapshot.NewController[string, libtypes.Controllable[string]]()
	_ = ctrl.Register(sp)
	_ = ctrl.Register(lj)
	_ = ctrl.Register(rj)
	_ = ctrl.Register(aj)
	_ = ctrl.Register(sj)
	_ = ctrl.Register(tj)

	return &stateDB{
		Plugin:                  sp,
		LogsJournal:             lj,
		RefundJournal:           rj,
		AccessListJournal:       aj,
		TransientStorageJournal: tj,
		SuicidesJournal:         sj,
		ctrl:                    ctrl,
	}
}

// =============================================================================
// Snapshot
// =============================================================================

// Snapshot implements `stateDB`.
func (sdb *stateDB) Snapshot() int {
	return sdb.ctrl.Snapshot()
}

// RevertToSnapshot implements `stateDB`.
func (sdb *stateDB) RevertToSnapshot(id int) {
	sdb.ctrl.RevertToSnapshot(id)
}

// =============================================================================
// Clean state
// =============================================================================

// Reset sets the TxContext for the current transaction and also manually clears any state from the
// previous tx in the journals, in case the previous tx reverted (Finalize was not called).
func (sdb *stateDB) Reset(txHash common.Hash, txIndex int) {
	sdb.LogsJournal.Finalize()
	sdb.RefundJournal.Finalize()
	sdb.AccessListJournal.Finalize()
	sdb.TransientStorageJournal.Finalize()
	sdb.SuicidesJournal.Finalize()

	sdb.LogsJournal.SetTxContext(txHash, txIndex)
}

// Finalize deletes the suicided accounts and finalizes all plugins.
func (sdb *stateDB) Finalize() {
	sdb.DeleteAccounts(sdb.GetSuicides())
	sdb.ctrl.Finalize()
}

// =============================================================================
// Prepare
// =============================================================================

// Implementation taken directly from the `stateDB` in Go-Ethereum.
//
// Prepare implements `stateDB`.
func (sdb *stateDB) Prepare(rules params.Rules, sender, coinbase common.Address,
	dest *common.Address, precompiles []common.Address, txAccesses coretypes.AccessList) {
	if rules.IsBerlin {
		// Clear out any leftover from previous executions
		sdb.AccessListJournal = journal.NewAccesslist()

		sdb.AddAddressToAccessList(sender)
		if dest != nil {
			sdb.AddAddressToAccessList(*dest)
			// If it's a create-tx, the destination will be added inside evm.create
		}
		for _, addr := range precompiles {
			sdb.AddAddressToAccessList(addr)
		}
		for _, el := range txAccesses {
			sdb.AddAddressToAccessList(el.Address)
			for _, key := range el.StorageKeys {
				sdb.AddSlotToAccessList(el.Address, key)
			}
		}
		if rules.IsShanghai { // EIP-3651: warm coinbase
			sdb.AddAddressToAccessList(coinbase)
		}
	}
}

// =============================================================================
// PreImage
// =============================================================================

// AddPreimage implements the the `StateDB`interface, but currently
// performs a no-op since the EnablePreimageRecording flag is disabled.
func (sdb *stateDB) AddPreimage(hash common.Hash, preimage []byte) {}

// AddPreimage implements the the `StateDB“ interface, but currently
// performs a no-op since the EnablePreimageRecording flag is disabled.
func (sdb *stateDB) Preimages() map[common.Hash][]byte {
	return nil
}

// =============================================================================
// Code Size
// =============================================================================

// GetCodeSize implements the `StateDB` interface by returning the size of the
// code associated with the given account.
func (sdb *stateDB) GetCodeSize(addr common.Address) int {
	return len(sdb.GetCode(addr))
}

// =============================================================================
// Other
// =============================================================================

func (sdb *stateDB) Finalise(_ bool) {
	sdb.Finalize()
}

func (sdb *stateDB) Commit(_ bool) (common.Hash, error) {
	sdb.Finalize()
	return common.Hash{}, nil
}

func (sdb *stateDB) Copy() StateDBI {
	return NewStateDB(sdb.Plugin)
}

func (sdb *stateDB) DumpToCollector(_ DumpCollector, _ *DumpConfig) []byte {
	return nil
}

func (sdb *stateDB) Dump(_ *DumpConfig) []byte {
	return nil
}

func (sdb *stateDB) RawDump(_ *DumpConfig) Dump {
	return Dump{}
}

func (sdb *stateDB) IteratorDump(_ *DumpConfig) IteratorDump {
	return IteratorDump{}
}

func (sdb *stateDB) Database() Database {
	return nil
}

func (sdb *stateDB) StartPrefetcher(_ string) {}

func (sdb *stateDB) StopPrefetcher() {}

func (sdb *stateDB) IntermediateRoot(_ bool) common.Hash {
	return common.Hash{}
}

func (sdb *stateDB) StorageTrie(_ common.Address) (Trie, error) {
	return nil, nil
}

func (sdb *stateDB) Error() error {
	return nil
}

func (sdb *stateDB) GetStorageProof(_ common.Address, _ common.Hash) ([][]byte, error) {
	return nil, nil
}

func (sdb *stateDB) GetProof(_ common.Address) ([][]byte, error) {
	return nil, nil
}

func (sdb *stateDB) GetOrNewStateObject(_ common.Address) *StateObject {
	return nil
}
