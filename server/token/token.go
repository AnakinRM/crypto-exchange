package token

import "github.com/anakinrm/crypto-exchange/server/cryptoClient"

var tokenRegistry = map[string]Token{
	"eth": &Eth{},
}

// using interface to process all the token TX
type Token interface {
	NewToken() Token
	GetPublicKey() (string, error)
	CheckBalance(cryptoClient.Client) (float64, error)
	CheckDeposit(cryptoClient.Client) (bool, error)
	Withdraw(cryptoClient.Client, string, float64) (float64, error)
	SendToExchange(cryptoClient.Client, string) (bool, error)
}

func GenerateWallet() (map[string]Token, error) {
	wallet := make(map[string]Token)
	for name, token := range tokenRegistry {
		// 调用 NewToken 方法生成新的实例

		wallet[name] = token.NewToken()
	}
	return wallet, nil
}
