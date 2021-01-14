package account

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kprc/libeth/util"
)

type BTLAccount struct {
	PrivKey ed25519.PrivateKey `json:"-"`
	PubKey  ed25519.PublicKey  `json:"-"`
	ID      BeatleAddress      `json:"id"`
}

type CryptBTLJson struct {
	ID        BeatleAddress `json:"id"`
	CipherTxt string        `json:"cipher_txt"`
}

func NewAccount() (acct *BTLAccount, err error) {
	var (
		priv ed25519.PrivateKey
		pub  ed25519.PublicKey
	)

	cnt := 0
	for {
		cnt++
		pub, priv, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			if cnt > 10 {
				return nil, err
			}
			continue
		} else {
			break
		}
	}

	acct = &BTLAccount{PrivKey: priv, PubKey: pub}
	acct.ID = PubKey2ID(pub)

	return
}

func (ba *BTLAccount) Marshal(auth string) ([]byte, error) {
	var salt [8]byte
	copy(salt[:], ba.PubKey[:8])

	k, err := util.AesKey(salt, auth)
	if err != nil {
		return nil, err
	}

	var cipherTxt []byte
	cipherTxt, err = util.Encrypt(k, ba.PrivKey)
	if err != nil {
		return nil, err
	}

	cj := &CryptBTLJson{}
	cj.ID = ba.ID
	cj.CipherTxt = base58.Encode(cipherTxt)

	return json.Marshal(*cj)
}

func (ba *BTLAccount) Unmarshal(data []byte, auth string) error {

	cj := &CryptBTLJson{}
	if err := json.Unmarshal(data, cj); err != nil {
		return err
	}

	if !cj.ID.IsValid() {
		return errors.New("not a valid account")
	}

	ba.ID = cj.ID
	ba.PubKey = ba.ID.DerivePubKey()
	if ba.PubKey == nil {
		return errors.New("id can't derive public key")
	}

	var salt [8]byte
	copy(salt[:], ba.PubKey[:8])

	k, err := util.AesKey(salt, auth)
	if err != nil {
		return err
	}

	var plainBytes []byte
	plainBytes, err = util.Decrypt(k, base58.Decode(cj.CipherTxt))
	if err != nil {
		return err
	}

	ba.PrivKey = plainBytes

	return nil

}

func BeatlesUnmarshal(data []byte) (*CryptBTLJson,error)  {
	cj:=&CryptBTLJson{}
	if err:=json.Unmarshal(data,cj);err!=nil{
		return nil, err
	}

	return cj,nil
}

