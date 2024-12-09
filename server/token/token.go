package token

import (
	"fmt"

	"github.com/anakinrm/crypto-exchange/server/cryptoClient"
)

const (
	MarketETH Market = "ETH"
)

type (
	Market string
)

// tokenRegistry holds the initial token definitions.
// We have Eth as an example. Others could be added similarly.
var tokenRegistry = map[Market]Token{
	MarketETH: &Eth{},
}

// Token interface acts like an abstract parent class, requiring all methods to be implemented.
// Some methods (GetPublicKey, CheckBalance, AddBalance, SubBalance) will be handled by a base embedded struct.
type Token interface {
	NewToken() Token
	GetPublicKey() (string, error)
	CheckBalance(cryptoClient.Client) (float64, error)
	CheckDeposit(cryptoClient.Client) (bool, error)
	Withdraw(cryptoClient.Client, string, float64) (float64, error)
	SendToExchange(cryptoClient.Client, string) (bool, error)
	AddBalance(float64) error
	SubBalance(float64) error
}

// BaseToken provides common fields and methods for tokens.
// This acts like the "parent" part, providing shared logic.
// All operations that are the same for all tokens are implemented here.
type BaseToken struct {
	PublicKey string
	Balance   float64
}

// GetPublicKey returns the public key of the token.
// Since all tokens share this logic, it's implemented here.
func (b *BaseToken) GetPublicKey() (string, error) {
	if b.PublicKey == "" {
		return "", fmt.Errorf("public key not found")
	}
	return b.PublicKey, nil
}

// CheckBalance returns the current balance of the token.
// Since all tokens share this logic (simply returning Balance), implement it here.
func (b *BaseToken) CheckBalance(c cryptoClient.Client) (float64, error) {
	return b.Balance, nil
}

// AddBalance increases the token's balance by the specified amount.
// Shared logic is implemented here.
func (b *BaseToken) AddBalance(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("amount must be positive")
	}
	b.Balance += amount
	return nil
}

// SubBalance decreases the token's balance by the specified amount if possible.
// Shared logic is implemented here.
func (b *BaseToken) SubBalance(amount float64) error {
	if amount > b.Balance {
		return fmt.Errorf("insufficient balance")
	}
	b.Balance -= amount
	return nil
}

// GenerateWallet creates a new wallet map with fresh token instances.
func GenerateWallet() (map[Market]Token, error) {
	wallet := make(map[Market]Token)
	for name, token := range tokenRegistry {
		// Call NewToken to generate a new instance of each token
		wallet[name] = token.NewToken()
	}
	return wallet, nil
}