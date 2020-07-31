package account

import (
	"crypto/ed25519"
	"errors"
	"github.com/btcsuite/btcutil/base58"
)

const (
	Ed25519AddrLen int    = 32
	PrefixLen      int    = 2
	PrefixStr      string = "tg"
	Version        int    = 1
	Version1Len    int    = 1
)

type Address [Ed25519AddrLen]byte

var emptyAddr Address

type BeatleAddress string

func (ba BeatleAddress) String() string {
	return string(ba)
}

func (ba BeatleAddress) Address() (addr Address, prefix string, ver int, err error) {
	sba := ba.String()
	if !IsValidID(sba) {
		return emptyAddr, "", 0, errors.New("id is not valid")
	}

	prefix = sba[:PrefixLen]

	data := base58.Decode(sba[PrefixLen:])

	vlen := int(data[0])

	if vlen == Version1Len {
		ver = int(data[1])
	}

	ver = int(data[1])

	copy(addr[:], data[PrefixLen:])

	return
}

func (ba BeatleAddress) DerivePubKey() ed25519.PublicKey {
	addr, _, _, err := ba.Address()
	if err != nil {
		return nil
	}

	return addr.PubKey()
}

func (ba BeatleAddress) IsValid() bool {
	return IsValidID(string(ba))
}

func PubKey2ID(pk ed25519.PublicKey) BeatleAddress {
	a := &Address{}

	a.SetPubKey(pk)

	return a.ID()
}

func (a Address) String() string {
	return base58.Encode(a[:])
}

func (a Address) ID() BeatleAddress {
	var data []byte

	data = append(data, byte(Version1Len), byte(Version))
	data = append(data, a[:]...)

	return BeatleAddress(PrefixStr + base58.Encode(data))
}

func (a Address) PubKey() ed25519.PublicKey {
	return a[:]
}

func (a *Address) SetPubKey(pk ed25519.PublicKey) {
	copy((*a)[:], pk)
}

func (a *Address) SetBytes(buf []byte) error {
	if len(buf) != Ed25519AddrLen {
		return errors.New("buffer length error")
	}
	copy((*a)[:], buf)

	return nil
}

func IsValidID(id string) bool {
	if len(id) <= 2 {
		return false
	}

	if id[:PrefixLen] != "tg" {
		return false
	}

	left := base58.Decode(id[PrefixLen:])

	if len(left) < 1 {
		return false
	}

	vlen := int(left[0])

	if len(left) < (1 + vlen + Ed25519AddrLen) {
		return false
	}

	return true
}
