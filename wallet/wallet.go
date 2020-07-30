package wallet

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kprc/libeth/account"
	"github.com/kprc/libeth/client"
	"math"
	"math/big"

	"github.com/kprc/nbsnetwork/tools"
)

type Wallet struct {
	account         account.Account
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
	CheckReceipt(sendMeAddr common.Address, txHash common.Hash) (float64, error)
	Save(auth string) error
	Load(auth string) error
}

func CreateWallet(walletSavePath string, remoteEthServer string) WalletIntf {
	acct, err := account.NewAccount()
	if err != nil {
		return nil
	}

	var ec *ethclient.Client
	ec, err = ethclient.Dial(remoteEthServer)
	if err != nil {
		return nil
	}

	w := &Wallet{account: *acct, SavePath: walletSavePath, RemoteEthServer: remoteEthServer}
	w.client.C = ec
	w.client.ServerHttpAddr = remoteEthServer

	return w
}

func NewWallet(walletSavePath string, remoteEthServer string) WalletIntf {
	w := &Wallet{SavePath: walletSavePath, RemoteEthServer: remoteEthServer}
	return w
}

func (w *Wallet) BalanceOf(force bool) (float64, error) {
	if !force {
		return w.balance, nil
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

	hash := tx.Hash()
	return &hash, nil
}

func (w *Wallet) Address() common.Address {
	return w.account.EAddr
}

func (w *Wallet) AccountString() string {
	return w.account.SAddr
}

func (w *Wallet) CheckReceipt(sendMeAddr common.Address, txHash common.Hash) (float64, error) {
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
			return 0, errors.New("not the send")
		}

		return BalanceHuman(tx.Value()), nil
	}
}

func (w *Wallet) Save(auth string) error {
	if w.account.PrivKey == nil {
		return errors.New("account error")
	}

	if data, err := w.account.Marshal(auth); err != nil {
		return err
	} else {
		if err = tools.Save2File(data, w.SavePath); err != nil {
			return err
		}
	}

	return nil
}

func (w *Wallet) Load(auth string) error {

	if data, err := tools.OpenAndReadAll(w.SavePath); err != nil {
		return err
	} else {
		if err = w.account.Unmarshal(data, auth); err != nil {
			return err
		}

		if w.client.C, err = ethclient.Dial(w.RemoteEthServer); err != nil {
			return err
		}
		w.client.ServerHttpAddr = w.RemoteEthServer
	}
	return nil
}
