package cryptoClient

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	clientAddr = "http://localhost:8545"
	decimals   = 18
)

type ethClient struct {
	client *ethclient.Client
}

func StartNewEthClient() *ethClient {
	c, err := ethclient.Dial(clientAddr)
	if err != nil {
		log.Fatal(err)
	}
	client := ethClient{
		client: c,
	}
	return &client
}

func (c ethClient) TransferETH(fromPrivKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {

	ctx := context.Background()
	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000)
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	return c.client.SendTransaction(ctx, signedTx)
}

func (c ethClient) GetBalance(addr string) (float64, error) {
	account := common.HexToAddress(addr)
	balance, err := c.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return 0.0, nil
	}

	return c.BigIntToFloat(balance), nil
}

func (c ethClient) FloatToBigInt(value float64) *big.Int {
	scaledValue := value * math.Pow10(decimals)

	integerValue := big.NewInt(0)
	integerValue.SetInt64(int64(scaledValue))

	return integerValue
}

func (c ethClient) BigIntToFloat(value *big.Int) float64 {
	bigFloatValue := new(big.Float).SetInt(value)

	scale := big.NewFloat(math.Pow10(decimals))

	adjustedValue := new(big.Float).Quo(bigFloatValue, scale)

	floatValue, _ := adjustedValue.Float64()
	return floatValue
}
