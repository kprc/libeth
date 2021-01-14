package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"github.com/kprc/libeth/util/edwards25519"
	"golang.org/x/crypto/scrypt"
	"io"
)

func AesKey(salt [8]byte, passwd string) ([]byte, error) {
	return scrypt.Key([]byte(passwd), salt[:], 32768, 8, 1, 32)
}

func Encrypt(key []byte, plainTxt []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipherTxt := make([]byte, aes.BlockSize+len(plainTxt))

	iv := cipherTxt[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherTxt[aes.BlockSize:], plainTxt)

	return cipherTxt, nil

}

func NewEncStream(key []byte) (stream cipher.Stream, iv [aes.BlockSize]byte,err error)  {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil,iv, err
	}
	_, err = io.ReadFull(rand.Reader, iv[:])
	if err != nil {
		return nil, iv,err
	}

	stream = cipher.NewCFBEncrypter(block,iv[:])

	return stream,iv,nil
}

func NewEncStreamWithIv(key []byte,iv [aes.BlockSize]byte) (stream cipher.Stream,err error)  {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream = cipher.NewCFBEncrypter(block,iv[:])

	return stream,nil
}



func NewDecStream(key []byte) (stream cipher.Stream, iv [aes.BlockSize]byte,err error)  {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil,iv, err
	}
	_, err = io.ReadFull(rand.Reader, iv[:])
	if err != nil {
		return nil, iv,err
	}

	stream = cipher.NewCFBDecrypter(block,iv[:])

	return stream,iv,nil
}

func NewDecStreamWithIv(key []byte,iv [aes.BlockSize]byte) (stream cipher.Stream,err error)  {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream = cipher.NewCFBDecrypter(block,iv[:])

	return stream,nil
}


func Encrypt2(stream cipher.Stream,plainTxt []byte)(cipherTxt []byte)  {
	stream.XORKeyStream(plainTxt,plainTxt)
	return plainTxt
}

func Decrypt2(stream cipher.Stream,cipherTxt []byte) (plainTxt []byte)  {
	stream.XORKeyStream(cipherTxt,cipherTxt)
	return cipherTxt
}

func Decrypt(key []byte, cipherTxt []byte) (plainTxt []byte, err error) {

	if len(cipherTxt) < aes.BlockSize {
		return nil, errors.New("cipher text too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := cipherTxt[:aes.BlockSize]
	cipherTxt = cipherTxt[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTxt, cipherTxt)

	return cipherTxt, nil
}

func PrivateKeyToCurve25519(curve25519Private *[32]byte, privateKey *[64]byte) {
	h := sha512.New()
	h.Write(privateKey[:32])
	digest := h.Sum(nil)

	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	copy(curve25519Private[:], digest)
}

func edwardsToMontgomeryX(outX, y *edwards25519.FieldElement) {
	var oneMinusY edwards25519.FieldElement
	edwards25519.FeOne(&oneMinusY)
	edwards25519.FeSub(&oneMinusY, &oneMinusY, y)
	edwards25519.FeInvert(&oneMinusY, &oneMinusY)

	edwards25519.FeOne(outX)
	edwards25519.FeAdd(outX, outX, y)

	edwards25519.FeMul(outX, outX, &oneMinusY)
}

func PublicKeyToCurve25519(curve25519Public *[32]byte, publicKey *[32]byte) bool {
	var A edwards25519.ExtendedGroupElement
	if !A.FromBytes(publicKey) {
		return false
	}

	var x edwards25519.FieldElement
	edwardsToMontgomeryX(&x, &A.Y)
	edwards25519.FeToBytes(curve25519Public, &x)
	return true
}
