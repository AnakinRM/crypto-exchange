package token

import (
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/anakinrm/crypto-exchange/server/cryptoClient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Eth represents a specific token type (child class).
// It embeds BaseToken to reuse the shared logic.
type Eth struct {
	BaseToken       // Embedding BaseToken gives Eth all the shared methods
	name            string
	privateKey      *ecdsa.PrivateKey
	lastAddrBalance float64
}

// NewToken creates a new instance of Eth with a fresh key pair and address.
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
		BaseToken:       BaseToken{PublicKey: address, Balance: 0.0},
		name:            "Ethereum",
		privateKey:      privateKey,
		lastAddrBalance: 0.0,
	}
}

// CheckDeposit checks if there has been a deposit to the ETH address.
// This logic is specific to Eth, so it's implemented here.
func (e *Eth) CheckDeposit(c cryptoClient.Client) (bool, error) {
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

// Withdraw attempts to withdraw a specified amount to a given address.
// This logic is specific to Eth, so it's implemented here.
func (e *Eth) Withdraw(c cryptoClient.Client, addr string, amount float64) (float64, error) {
	if e.Balance < amount {
		return amount, fmt.Errorf("Insufficient balance")
	}

	walletBalance, err := c.Eth.GetBalance(e.PublicKey)
	if err != nil {
		return amount, err
	}

	if walletBalance < amount {
		// This may fail due to gas fees
		err := c.Eth.TransferETH(e.privateKey, common.HexToAddress(addr), c.Eth.FloatToBigInt(walletBalance))
		if err != nil {
			return amount, err
		}
		e.lastAddrBalance = 0.0
		amount -= walletBalance
		return amount, fmt.Errorf("Insufficient wallet balance")
	}

	err = c.Eth.TransferETH(e.privateKey, common.HexToAddress(addr), c.Eth.FloatToBigInt(amount))
	if err != nil {
		return amount, err
	}
	// Update local balance after successful transfer
	e.Balance -= amount
	return 0.0, nil
}

// SendToExchange sends all available balance to the specified exchange address.
// This logic is specific to Eth, so it's implemented here.
func (e *Eth) SendToExchange(c cryptoClient.Client, addr string) (bool, error) {
	walletBalance, err := c.Eth.GetBalance(e.PublicKey)
	if err != nil {
		return false, err
	}

	err = c.Eth.TransferETH(e.privateKey, common.HexToAddress(addr), c.Eth.FloatToBigInt(walletBalance))
	if err != nil {
		return false, err
	}
	e.lastAddrBalance = 0.0
	e.Balance = 0.0 // update local balance after sending all funds
	return true, nil
}
