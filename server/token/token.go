package token

import (
	"fmt"

	"github.com/anakinrm/crypto-exchange/server/cryptoClient"
	"github.com/anakinrm/crypto-exchange/server/db"
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
	StoreTokenToDataBase(int64) error
	GetTokenFromDataBase(db.Wallet) (Token, error)
}

// BaseToken provides common fields and methods for tokens.
// This acts like the "parent" part, providing shared logic.
// All operations that are the same for all tokens are implemented here.
type BaseToken struct {
	PublicKey       string
	Balance         float64
	privateKey      string
	name            Market
	lastAddrBalance float64
}

func (b *BaseToken) StoreTokenToDataBase(userID int64) error {

	w := db.Wallet{
		UserID:          userID,
		TokenType:       string(b.name),
		PublicKey:       b.PublicKey,
		Balance:         b.Balance,
		LastAddrBalance: b.lastAddrBalance,
	}
	err := w.SetEncryptPrivateKey(b.privateKey)
	if err != nil {
		return err
	}
	err = w.InsertWallet()
	if err != nil {
		return err
	}
	return nil
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

// store the tokens in the wallet into database
func StoreWalletInDataBase(wallet map[Market]Token, userID int64) error {
	for _, v := range wallet {
		err := v.StoreTokenToDataBase(userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetWalletFromDataBase(userID int64) (map[Market]Token, error) {
	walletDB, err := db.GetWalletsByUserID(userID)
	if err != nil {
		return nil, err
	}

	wallet := make(map[Market]Token)

	for _, v := range walletDB {

		wallet[Market(v.TokenType)], err = tokenRegistry[Market(v.TokenType)].GetTokenFromDataBase(v)
		if err != nil {
			return nil, err
		}
	}

	return wallet, nil
}
