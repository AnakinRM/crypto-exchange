package cryptoClient

type Client struct {
	Eth *ethClient
}

func newClient() *Client {
	return &Client{
		Eth: StartNewEthClient(),
	}
}
