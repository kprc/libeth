package wallet

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kprc/libeth/account"
	"github.com/kprc/libeth/client"
)

type Wallet struct {
	account account.Account
	client client.Client
	SavePath   string
	RemoteEthServer  string
}

type WalletIntf interface {
	BalanceOf() float64
	SendTo(to common.Address) error
	Address() common.Address
	AccountString() string
	CheckReceipt(sendMeAddr common.Address) (float64,error)
	Save() error
	Load() error
}

func NewWallet(walletSavePath string, remoteEthServer string) WalletIntf {
	return &Wallet{SavePath: walletSavePath,RemoteEthServer: remoteEthServer}
}

func (w *Wallet)BalanceOf() float64  {
	return 0
}

func (w *Wallet)SendTo(to common.Address) error  {
	return nil
}

func (w *Wallet)Address() common.Address  {
	return w.account.EAddr
}

func (w *Wallet) AccountString() string {
	return w.account.SAddr
}

func (w *Wallet)CheckReceipt(sendMeAddr common.Address) (float64,error)  {
	return 0,nil
}

func (w *Wallet)Save() error  {
	return nil
}

func (w *Wallet)Load() error  {
	return nil
}

