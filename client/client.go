package client

import "github.com/ethereum/go-ethereum/ethclient"

type Client struct {
	ServerHttpAddr string
	C              *ethclient.Client
}
