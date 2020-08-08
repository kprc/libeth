package wallet

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kprc/libeth/account"
	"github.com/kprc/libeth/client"
	"github.com/kprc/libeth/util"
	"golang.org/x/crypto/curve25519"
	"math"
	"math/big"

	"github.com/kprc/nbsnetwork/tools"
)

type Wallet struct {
	account         account.EthAccount
	btlAccount      account.BTLAccount
	client          client.Client
	SavePath        string
	RemoteEthServer string
	balance         float64
}

type WalletIntf interface {
	BalanceOf(force bool) (float64, error)
	SendTo(to common.Address, balance float64) (*common.Hash, error)
	Address() common.Address
	AccountString() string
	BtlAddress() account.BeatleAddress
	CheckReceipt(sendMeAddr common.Address, txHash common.Hash) (float64, error)
	Save(auth string) error
	Load(auth string) error
	BtlSign(data []byte) []byte
	BtlVerifySig(data, sig []byte) bool
	BtlPeerEncrypt(peerPub ed25519.PublicKey, plainBytes []byte) (cipherBytes []byte, err error)
	BtlPeerEncrypt2(peerId account.BeatleAddress, plainBytes []byte) (cipherBytes []byte, err error)
	BtlPeerDecrypt(peerPub ed25519.PublicKey, cipherBytes []byte) (plainBytes []byte, err error)
	BtlPeerDecrypt2(peerId account.BeatleAddress, cipherBytes []byte) (plainBytes []byte, err error)
	AesKey(peerPub ed25519.PublicKey) (key []byte, err error)
	AesKey2(peerId account.BeatleAddress) (key []byte, err error)
	RecoverEthAccount(hexString, auth string) error
	ExportWallet() (string, error)
	RecoverWallet(walletString, auth string) error
	String(auth string) (string, error)
}

func CreateWallet(walletSavePath string, remoteEthServer string) WalletIntf {
	acct, err := account.NewEthAccount()
	if err != nil {
		return nil
	}

	var ec *ethclient.Client
	if remoteEthServer != "" {
		ec, err = ethclient.Dial(remoteEthServer)
		if err != nil {
			return nil
		}
	}

	btlacct, err := account.NewAccount()
	if err != nil {
		return nil
	}

	w := &Wallet{account: *acct, btlAccount: *btlacct, SavePath: walletSavePath, RemoteEthServer: remoteEthServer}
	w.client.C = ec
	w.client.ServerHttpAddr = remoteEthServer

	return w
}

func NewWallet(walletSavePath string, remoteEthServer string) WalletIntf {
	w := &Wallet{SavePath: walletSavePath, RemoteEthServer: remoteEthServer}
	return w
}

func RecoverWallet(walletSavePath, remoteEthServer, auth string) (WalletIntf, error) {
	if !tools.FileExists(walletSavePath) {
		return nil, errors.New("wallet saved file not found : " + walletSavePath)
	}

	w := NewWallet(walletSavePath, remoteEthServer)

	if err := w.Load(auth); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Wallet) BalanceOf(force bool) (float64, error) {
	if !force {
		return w.balance, nil
	}
	if w.client.C == nil {
		return 0, errors.New("no eth client")
	}

	if balance, err := w.client.C.BalanceAt(context.Background(), w.account.EAddr, nil); err != nil {
		return 0, err
	} else {
		w.balance = BalanceHuman(balance)
	}

	return w.balance, nil
}

func BalanceHuman(balance *big.Int) float64 {
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	v := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	vv, _ := v.Float64()
	return vv
}

func BalanceEth(balance float64) *big.Int {
	fbalance := new(big.Float)
	fbalance.SetFloat64(balance)
	v := new(big.Float).Mul(fbalance, big.NewFloat(math.Pow10(18)))

	vv := new(big.Int)
	v.Int(vv)

	return vv
}

func (w *Wallet) SendTo(to common.Address, balance float64) (*common.Hash, error) {
	if w.client.C == nil {
		return nil, errors.New("no eth client")
	}

	nonce, err := w.client.C.PendingNonceAt(context.Background(), w.account.EAddr)
	if err != nil {
		return nil, err
	}

	var gasPrice *big.Int
	gasPrice, err = w.client.C.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	var chainId *big.Int
	chainId, err = w.client.C.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(nonce, to, BalanceEth(balance), uint64(21000), gasPrice, nil)
	var signdTx *types.Transaction
	signdTx, err = types.SignTx(tx, types.NewEIP155Signer(chainId), w.account.PrivKey)
	if err != nil {
		return nil, err
	}

	err = w.client.C.SendTransaction(context.Background(), signdTx)
	if err != nil {
		return nil, err
	}

	hash := signdTx.Hash()
	return &hash, nil
}

func (w *Wallet) Address() common.Address {
	return w.account.EAddr
}

func (w *Wallet) BtlAddress() account.BeatleAddress {
	return w.btlAccount.ID
}

func (w *Wallet) AccountString() string {
	return w.account.SAddr
}

func (w *Wallet) CheckReceipt(sendMeAddr common.Address, txHash common.Hash) (float64, error) {

	if w.client.C == nil {
		return 0, errors.New("no eth client")
	}

	if tx, isPending, err := w.client.C.TransactionByHash(context.Background(), txHash); err != nil {
		return 0, err
	} else {
		if isPending {
			return 0, errors.New("is pending, please wait")
		}

		if tx.To().Hex() != w.account.SAddr {
			return 0, errors.New("not send for me")
		}
		var chainId *big.Int
		if chainId, err = w.client.C.NetworkID(context.Background()); err != nil {
			return 0, err
		}
		var msg types.Message
		if msg, err = tx.AsMessage(types.NewEIP155Signer(chainId)); err != nil {
			return 0, err
		}
		if msg.From() != sendMeAddr {
			return 0, errors.New("not the sender")
		}

		return BalanceHuman(tx.Value()), nil
	}
}

type WalletSaveJson struct {
	EthAcct string `json:"eth_acct"`
	BtlAcct string `json:"btl_acct"`
}

func (w *Wallet) Save(auth string) error {

	if data, err := w.String(auth); err != nil {
		return err
	} else {
		if err = tools.Save2File([]byte(data), w.SavePath); err != nil {
			return err
		}
	}

	return nil
}

func (w *Wallet) String(auth string) (string, error) {
	if w.account.PrivKey == nil || w.btlAccount.PrivKey == nil {
		return "", errors.New("account error")
	}

	wsj := &WalletSaveJson{}
	var (
		ethAcct []byte
		btlAcct []byte
		data    []byte
		err     error
	)

	if ethAcct, err = w.account.Marshal(auth); err != nil {
		return "", err
	}
	if btlAcct, err = w.btlAccount.Marshal(auth); err != nil {
		return "", err
	}

	wsj.EthAcct = string(ethAcct)
	wsj.BtlAcct = string(btlAcct)
	if data, err = json.Marshal(*wsj); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func (w *Wallet) Load(auth string) error {

	var (
		data []byte
		err  error
	)

	if data, err = tools.OpenAndReadAll(w.SavePath); err != nil {
		return err
	}

	return w.RecoverWallet(string(data), auth)
}

func (w *Wallet) BtlSign(data []byte) []byte {
	return ed25519.Sign(w.btlAccount.PrivKey, data)
}
func (w *Wallet) BtlVerifySig(data, sig []byte) bool {
	return ed25519.Verify(w.btlAccount.PubKey, data, sig)
}

func (w *Wallet) BtlPeerEncrypt(peerPub ed25519.PublicKey, plainBytes []byte) (cipherBytes []byte, err error) {
	var k []byte
	if k, err = w.AesKey(peerPub); err != nil {
		return nil, err
	}

	if cipherBytes, err = util.Encrypt(k, plainBytes); err != nil {
		return nil, err
	}

	return
}
func (w *Wallet) BtlPeerEncrypt2(peerId account.BeatleAddress, plainBytes []byte) (cipherBytes []byte, err error) {
	if !peerId.IsValid() {
		return nil, errors.New("beatles address error")
	}
	pk := peerId.DerivePubKey()

	return w.BtlPeerEncrypt(pk, plainBytes)

}
func (w *Wallet) BtlPeerDecrypt(peerPub ed25519.PublicKey, cipherBytes []byte) (plainBytes []byte, err error) {
	var k []byte
	if k, err = w.AesKey(peerPub); err != nil {
		return nil, err
	}

	if plainBytes, err = util.Decrypt(k, cipherBytes); err != nil {
		return nil, err
	}

	return
}
func (w *Wallet) BtlPeerDecrypt2(peerId account.BeatleAddress, cipherBytes []byte) (plainBytes []byte, err error) {
	if !peerId.IsValid() {
		return nil, errors.New("beatles address error")
	}
	pk := peerId.DerivePubKey()

	return w.BtlPeerDecrypt(pk, cipherBytes)
}

func (w *Wallet) AesKey(peerPub ed25519.PublicKey) (key []byte, err error) {
	var priKey [32]byte
	var privateKeyBytes [64]byte
	copy(privateKeyBytes[:], w.btlAccount.PrivKey)
	util.PrivateKeyToCurve25519(&priKey, &privateKeyBytes)

	var curvePub, pubKey [32]byte
	copy(pubKey[:], peerPub)
	if ok := util.PublicKeyToCurve25519(&curvePub, &pubKey); !ok {
		return nil, errors.New("convert public key error")
	}
	return curve25519.X25519(priKey[:], curvePub[:])
}
func (w *Wallet) AesKey2(peerId account.BeatleAddress) (key []byte, err error) {
	if !peerId.IsValid() {
		return nil, errors.New("beatles address error")
	}
	pk := peerId.DerivePubKey()

	return w.AesKey(pk)
}

func (w *Wallet) RecoverEthAccount(hexString, auth string) error {
	if err := w.account.ImportFromMetaMask(hexString); err != nil {
		return err
	}

	if err := w.Save(auth); err != nil {
		return err
	}

	return nil
}

func (w *Wallet) RecoverWallet(walletString, auth string) error {

	if w.SavePath == "" {
		return errors.New("please set wallet save path")
	}

	var err error
	wsj := &WalletSaveJson{}
	if err = json.Unmarshal([]byte(walletString), wsj); err != nil {
		return err
	}

	if wsj.EthAcct != "" {
		if err = w.account.Unmarshal([]byte(wsj.EthAcct), auth); err != nil {
			return err
		}
	}

	if w.RemoteEthServer != "" {
		if w.client.C, err = ethclient.Dial(w.RemoteEthServer); err != nil {
			return err
		}
		w.client.ServerHttpAddr = w.RemoteEthServer
	}

	if wsj.BtlAcct != "" {
		if err = w.btlAccount.Unmarshal([]byte(wsj.BtlAcct), auth); err != nil {
			return err
		}
	}

	return nil
}

func (w *Wallet) ExportWallet() (string, error) {
	if data, err := tools.OpenAndReadAll(w.SavePath); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}
