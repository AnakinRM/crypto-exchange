package token

import (
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/anakinrm/crypto-exchange/server/cryptoClient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Eth struct {
	name            string
	privateKey      *ecdsa.PrivateKey
	PublicKey       string
	Balance         float64
	lastAddrBalance float64
}

func (e *Eth) NewToken() Token {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return &Eth{
		name:            "Ethereum",
		privateKey:      privateKey,
		PublicKey:       address,
		Balance:         0.0,
		lastAddrBalance: 0.0,
	}
}

func (e *Eth) GetPublicKey() (string, error) {
	if e.PublicKey == "" {
		return "", fmt.Errorf("public key not found")
	}
	return e.PublicKey, nil
}

func (e *Eth) CheckBalance(c cryptoClient.Client) (float64, error) {
	return e.Balance, nil
}

func (e *Eth) CheckDeposit(c cryptoClient.Client) (bool, error) {
	// Placeholder implementation
	walletBalance, err := c.Eth.GetBalance(e.PublicKey)
	if err != nil {
		return false, err
	}
	if walletBalance != e.lastAddrBalance {
		e.Balance += (walletBalance - e.lastAddrBalance)
		e.lastAddrBalance = walletBalance
		return true, nil
	}

	return false, nil
}

func (e *Eth) Withdraw(c cryptoClient.Client, addr string, amount float64) (float64, error) {
	// Placeholder implementation

	if e.Balance < amount {
		return amount, fmt.Errorf("Insufficient balance")
	}

	walletBalance, err := c.Eth.GetBalance(e.PublicKey)
	if err != nil {
		return amount, err
	}

	if walletBalance < amount {
		//will faile because of the gas fee
		c.Eth.TransferETH(e.privateKey, common.HexToAddress(addr), c.Eth.FloatToBigInt(walletBalance))
		e.lastAddrBalance = 0.0
		amount -= walletBalance
		return amount, fmt.Errorf("Insufficient wallet balance")
	}

	c.Eth.TransferETH(e.privateKey, common.HexToAddress(addr), c.Eth.FloatToBigInt(amount))
	return 0.0, nil
}

func (e *Eth) SendToExchange(c cryptoClient.Client, addr string) (bool, error) {
	// Placeholder implementation
	walletBalance, err := c.Eth.GetBalance(e.PublicKey)
	if err != nil {
		return false, err
	}

	err = c.Eth.TransferETH(e.privateKey, common.HexToAddress(addr), c.Eth.FloatToBigInt(walletBalance))
	if err != nil {
		return false, err
	}
	e.lastAddrBalance = 0.0
	return true, nil
}
