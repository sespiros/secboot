// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2023 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package bootenv

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"errors"

	"github.com/snapcore/secboot"
	"golang.org/x/crypto/cryptobyte"
	cryptobyte_asn1 "golang.org/x/crypto/cryptobyte/asn1"
)

var (
	ComputeSnapModelHash = computeSnapModelHash
)

func MockSetModel(f func(secboot.SnapModel) bool) (restore func()) {
	origSetModel := SetModel
	SetModel = f
	return func() {
		SetModel = origSetModel
	}
}

func MockSetBootMode(f func(string) bool) (restore func()) {
	origSetBootMode := SetBootMode
	SetBootMode = f
	return func() {
		SetBootMode = origSetBootMode
	}
}

func MockLoadCurrentModel(f func() (secboot.SnapModel, error)) (restore func()) {
	origLoadCurrentModel := loadCurrentModel
	loadCurrentModel = f
	return func() {
		loadCurrentModel = origLoadCurrentModel
	}
}

func MockLoadCurrenBootMode(f func() (string, error)) (restore func()) {
	origLoadCurrentBootMode := loadCurrentBootMode
	loadCurrentBootMode = f
	return func() {
		loadCurrentBootMode = origLoadCurrentBootMode
	}
}

func (d *KeyDataScope) TestSetVersion(version int) {
	d.data.Version = version
}

func UnmarshalAdditionalData(data []byte) (*AdditionalData, error) {
	s := cryptobyte.String(data)

	if !s.ReadASN1(&s, cryptobyte_asn1.SEQUENCE) {
		return nil, errors.New("malformed input")
	}

	aad := new(AdditionalData)

	if !s.ReadASN1Integer(&aad.Version) {
		return nil, errors.New("malformed version")
	}

	if !s.ReadASN1Integer(&aad.BaseVersion) {
		return nil, errors.New("malformed base version")
	}

	kdfAlg, err := unmarshalHashAlg(&s)
	if err != nil {
		return nil, errors.New("malformed kdf")
	}
	aad.KdfAlg = kdfAlg

	var authMode int
	if !s.ReadASN1Enum(&authMode) {
		return nil, errors.New("malformed Auth mode")
	}
	aad.AuthMode = secboot.AuthMode(authMode)

	keyIdAlg, err := unmarshalHashAlg(&s)
	if err != nil {
		return nil, errors.New("malformed kdf")
	}
	aad.KeyIdentifierAlg = keyIdAlg

	if !s.ReadASN1Bytes(&aad.KeyIdentifier, cryptobyte_asn1.OCTET_STRING) {
		return nil, errors.New("malformed Key identifier")
	}

	return aad, nil
}

func (d *KeyDataScope) TestMatch(KDFAlg crypto.Hash, keyIdentifier []byte) bool {
	der, err := x509.MarshalPKIXPublicKey(d.data.PublicKey.PublicKey)
	if err != nil {
		return false
	}

	h := KDFAlg.New()
	h.Write(der)
	return bytes.Equal(h.Sum(nil), keyIdentifier)
}
