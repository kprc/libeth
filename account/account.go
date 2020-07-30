package account

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	PrivKey *ecdsa.PrivateKey `json:"-"`
	PubKey  *ecdsa.PublicKey  `json:"-"`
	EAddr   common.Address    `json:"-"`
	SAddr   string            `json:"s_addr"`
}

type AccountJson struct {
	Acct *Account             `json:"acct"`
	Cj   *keystore.CryptoJSON `json:"cj"`
}

func NewAccount() (acct *Account, err error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	acct = &Account{}
	acct.PrivKey = key
	acct.PubKey = (key.Public()).(*ecdsa.PublicKey)
	acct.EAddr = crypto.PubkeyToAddress(*acct.PubKey)
	acct.SAddr = acct.EAddr.String()

	return acct, nil
}

func (acct *Account) Unmarshal(data []byte, auth string) error {

	aj := &AccountJson{}
	if err := json.Unmarshal(data, aj); err != nil {
		return err
	}

	acct.SAddr = aj.Acct.SAddr

	if keyBytes, err := keystore.DecryptDataV3(*aj.Cj, auth); err != nil {
		return err
	} else {
		acct.PrivKey = crypto.ToECDSAUnsafe(keyBytes)
		acct.PubKey = (acct.PrivKey.Public()).(*ecdsa.PublicKey)
		acct.EAddr = crypto.PubkeyToAddress(*acct.PubKey)
	}

	return nil
}

func (acct *Account) Marshal(auth string) ([]byte, error) {
	keyBytes := math.PaddedBigBytes(acct.PrivKey.D, 32)
	aj := &AccountJson{}
	if cs, err := keystore.EncryptDataV3(keyBytes, []byte(auth), keystore.StandardScryptN, keystore.StandardScryptP); err != nil {
		return nil, err
	} else {
		aj.Acct = acct
		aj.Cj = &cs
	}

	return json.Marshal(*aj)
}
