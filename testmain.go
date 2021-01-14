package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/howeyc/gopass"
	"github.com/hyperorchidlab/go-miner-pool/eth/generated"
	"github.com/kprc/libeth/wallet"
	"github.com/kprc/nbsnetwork/tools"
	"log"
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

	c,_:=ethclient.Dial("https://mainnet.infura.io/v3/df97d0caa3514b3d99e94bc7764cffa0")
	toAddr := common.HexToAddress("0xc0e8fd90a8e74d79dc21583ff830be165a504e47")

	tokenAddress := common.HexToAddress("0x1999ac2b141E6d5c4e27579b30f842078bc620b3")
	instance, err := generated.NewToken(tokenAddress, c)
	if err != nil {
		log.Fatal(err)
	}

	bal, err := instance.BalanceOf(&bind.CallOpts{}, toAddr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(bal.String())

	//b,e:=c.BalanceAt(context.Background(),toAddr,nil)
	//if e== nil{
	//	fmt.Println(b.String())
	//}


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

	//fmt.Println("balance is :", bn)
	//
	//toAddr := common.HexToAddress("0x101c79204C2eDA00EB589371703fA3203cDE78aD")
	////
	//tx,err:=w.SendTo(toAddr,0.1)
	//if err!=nil{
	//	fmt.Println(err)
	//}
	//fmt.Println("tx hash",tx.Hex())
	//
	//fmt.Println("addr is :", w.AccountString(),"send to:",toAddr.String())
	//fmt.Println("beatle addr:", w.BtlAddress().String())
	//
	//bn, _ = w.BalanceOf(true)
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

	fromAddr := common.HexToAddress("0xfbc01bD1fD789032c0741aef9e25810538708C20")
	hash := common.HexToHash("0x303e307bc2db70f15826ed58c34988cc70cdab33f34ee63284d405892507a15e")

	checkv, err := w.CheckReceipt(fromAddr, hash)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("receive eth: ", checkv)

}
