package main

import (
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/kprc/libeth/wallet"
	"github.com/kprc/nbsnetwork/tools"
	"os"
	"path"
)

var w wallet.WalletIntf

func main() {

	h, _ := tools.Home()
	savepath := path.Join(h, ".wallet")
	server := "https://ropsten.infura.io/v3/df97d0caa3514b3d99e94bc7764cffa0"
	passwd, err := gopass.GetPasswdPrompt("Please Enter Password: ", true, os.Stdin, os.Stdout)
	if err != nil {
		panic("input password error")
	}

	if len(passwd) < 1 {
		panic("input password error")
	}

	if w == nil {
		if tools.FileExists(savepath) {
			w = wallet.NewWallet(savepath, server)
			w.Load(string(passwd))
			if err != nil {
				panic("load wallet failed")
			}
		} else {
			w = wallet.CreateWallet(savepath, server)
			fmt.Println("passwd is", string(passwd))
			err = w.Save(string(passwd))
			if err != nil {
				panic("save wallet failed")
			}
		}
	}

	bn, _ := w.BalanceOf(true)

	fmt.Println("addr is :", w.AccountString())
	fmt.Println("balance is :", bn)

}
