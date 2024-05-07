package solclient

import (
	"net/http"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
)

func GetSolanaClient(endpoint string, timeout time.Duration) *client.Client {

	httpTransport := http.Transport{
		ReadBufferSize: 128000,
	}
	httpClient := &http.Client{Transport: &httpTransport, Timeout: timeout * time.Second}

	solClient := client.New(rpc.WithEndpoint(endpoint), rpc.WithHTTPClient(httpClient))
	return solClient
}
