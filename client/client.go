package client

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"net"
	"net/http"
	"syscall"
	"time"
)

type Client struct {
	ProtectFD func(fd int32) bool
	DialTimeout int
	ConnTimeout int
	ServerHttpAddr string
	C              *ethclient.Client
}

func NewClient(addr string, protect func(fd int32)bool, dialTimeout, connTimeout int) *Client  {
	return &Client{
		ProtectFD: protect,
		DialTimeout: dialTimeout,
		ConnTimeout: connTimeout,
		ServerHttpAddr: addr,
	}
}

func (cl *Client)dialHttp(endpoint string) (*rpc.Client,error)  {
	if cl.ServerHttpAddr == ""{
		cl.ServerHttpAddr = endpoint
	}

	var transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error){
			d := &net.Dialer{
				Timeout: time.Second * time.Duration(cl.DialTimeout),
				Control: func(network, address string, c syscall.RawConn) error {
					if cl.ProtectFD != nil {
						p:= func(fd uintptr) {
							cl.ProtectFD(int32(fd))
						}
						return c.Control(p)
					}
					return nil
				},
			}

			conn, err := d.Dial("udp", addr)
			if err != nil {
				return nil, err
			}

			if cl.ConnTimeout > 0{
				conn.SetDeadline(time.Now().Add(time.Duration(cl.ConnTimeout)*time.Second))
			}

			return conn,nil
		},
	}

	c:=&http.Client{
		Transport: transport,
	}

	return rpc.DialHTTPWithClient(endpoint,c)
}

func (cl *Client)Dial(rawurl string) (*ethclient.Client, error)  {
	c,err:=cl.dialHttp(rawurl)
	if err!=nil{
		return nil, err
	}

	cl.C = ethclient.NewClient(c)

	return cl.C, nil
}

