package epik

import (
	"github.com/EpiK-Protocol/go-epik/api"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/filecoin-project/go-jsonrpc"
)

//Client ...
type Client struct {
	URL                string
	Token              string
	FullNodeAPI        api.FullNode
	CommonAPI          api.Common
	StorageMinerAPI    api.StorageMiner
	WorkerAPI          api.WorkerAPI
	fullNodeCloser     jsonrpc.ClientCloser
	commonCloser       jsonrpc.ClientCloser
	storageMinerCloser jsonrpc.ClientCloser
	workerCloser       jsonrpc.ClientCloser
}

//NewClient ...
func NewClient(url string, token string) (c *Client, err error) {
	c = &Client{
		URL:   url,
		Token: token,
	}
	c.FullNodeAPI, c.fullNodeCloser, err = client.NewFullNodeRPC(url, nil)
	return c, err
}
