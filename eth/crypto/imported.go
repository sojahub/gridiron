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

package crypto

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var (
	SigToPub                = crypto.SigToPub
	Ecrecover               = crypto.Ecrecover
	CreateAddress           = crypto.CreateAddress
	UnmarshalPubkey         = crypto.UnmarshalPubkey
	CompressPubkey          = crypto.CompressPubkey
	DecompressPubkey        = crypto.DecompressPubkey
	DigestLength            = crypto.DigestLength
	EthSign                 = crypto.Sign
	EthSecp256k1Sign        = secp256k1.Sign
	FromECDSA               = crypto.FromECDSA
	GenerateEthKey          = crypto.GenerateKey
	ValidateSignatureValues = crypto.ValidateSignatureValues
	Keccak256               = crypto.Keccak256
	Keccak256Hash           = crypto.Keccak256Hash
	PubkeyToAddress         = crypto.PubkeyToAddress
	SignatureLength         = crypto.SignatureLength
	ToECDSA                 = crypto.ToECDSA
	VerifySignature         = crypto.VerifySignature
	FromECDSAPub            = crypto.FromECDSAPub
)
