package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/howeyc/gopass"
	"github.com/kprc/libeth/wallet"
	"github.com/kprc/nbsnetwork/tools"
	"os"
	"path"
)

var w wallet.WalletIntf

func main() {
	//test2()

	//test1()

	//var a,b float64
	//
	//a = 1.1
	//b = 1.0
	//
	//fmt.Println(int64(a/b))
	//
	//a = 1.51
	//b = 1.0
	//
	//fmt.Println(int64(a/b))

	//fmt.Println(common.BytesToHash(base58.Decode("DGt9G5Qw9Eyr4cVtSmsncpwnVtgin4rj8VsBPfCXFVXh")).Hex())

}

func test2() {
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
			err = w.Save(string(passwd))
			if err != nil {
				panic("save wallet failed")
			}
		}
	}

	bn, _ := w.BalanceOf(true)

	//toAddr := common.HexToAddress("0x31495a3C681Aac2310e89Fa229863E48e2bf0DeD")
	//
	//tx,err:=w.SendTo(toAddr,0.1)
	//if err!=nil{
	//	fmt.Println(err)
	//}
	//fmt.Println("tx hash",tx.Hex())

	fmt.Println("addr is :", w.AccountString())
	fmt.Println("beatle addr:", w.BtlAddress().String())
	fmt.Println("balance is :", bn)
}

func test1() {
	h, _ := tools.Home()
	savepath := path.Join(h, ".wallet1")
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
			err = w.Load(string(passwd))
			if err != nil {
				panic("load wallet failed")
			}
		} else {
			w = wallet.CreateWallet(savepath, server)
			err = w.Save(string(passwd))
			if err != nil {
				panic("save wallet failed")
			}
		}
	}

	bn, _ := w.BalanceOf(true)

	fmt.Println("addr is :", w.AccountString())
	fmt.Println("beatle addr:", w.BtlAddress().String())
	fmt.Println("balance is :", bn)

	fromAddr := common.HexToAddress("0x6148AdCeA554793f0022717F5C0AC92571b30E98")
	hash := common.HexToHash("0x76b7ab014e25398e68221a06aa97f7887ff68658db4c74be3204d5d27f6d5f9d")

	checkv, err := w.CheckReceipt(fromAddr, hash)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("receive eth: ", checkv)

}
