package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/anakinrm/crypto-exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

const (
	MarketETH Market = "ETH"

	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"

	exchangePrivateKey = "844f82132f4d1381ccc4c196e3a6c66216c40822984cacf1325ecf9667f67014"
)

type (
	Market    string
	OrderType string

	PlaceOrderRequest struct {
		UserID int64
		Type   OrderType // limit or market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	MatchedOrders struct {
		Price float64
		Size  float64
		ID    int64
	}
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	privKeyStr := "da06ed10bc10488900bdcad9f1893880a98e24cf83d29006a1960139a8492028"
	privateKey, err := crypto.HexToECDSA(privKeyStr)
	if err != nil {
		log.Fatal(err)
	}

	user := &User{
		ID:         8888,
		PrivateKey: privateKey,
	}

	ex.Users[user.ID] = user

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	address := "0x9d13e20D0494ba52B79Cb8433e711a498a8E63E5"
	balance, _ := ex.Client.BalanceAt(context.Background(), common.HexToAddress(address), nil)

	fmt.Println(balance)

	// ctx := context.Background()

	// privateKey, err := crypto.HexToECDSA("66e5718e88d6b763e9495702d08cf5003864f98f60f841c5f452799a47ae1121")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// publicKey := privateKey.Public()

	// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	// if !ok {
	// 	log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	// }

	// fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// value := big.NewInt(100000000000000000)
	// gasLimit := uint64(21000)
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// toAddress := common.HexToAddress("0x22D5AF0D15d08E3Efe1e593E1D9811F26c50708A")
	// tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	// chainID := big.NewInt(1337)

	// fmt.Println(chainID)

	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if err := client.SendTransaction(context.Background(), signedTx); err != nil {
	// 	log.Fatal(err)
	// }

	// balace, err := client.BalanceAt(ctx, toAddress, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(balace)

	e.Start(":3000")

}

type User struct {
	ID         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(pk string) *User {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		panic(err)
	}

	return &User{
		PrivateKey: privateKey,
	}
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type Exchange struct {
	Client     *ethclient.Client
	Users      map[int64]*User
	orders     map[int64]int64
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	return &Exchange{
		Client:     client,
		Users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		PrivateKey: privateKeyECDSA,
		orderbooks: orderbooks,
	}, nil
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]

	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderbookData := OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)

}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")

	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]

	order, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "can't find order ID: " + idStr})
	}
	ob.CancelOrder(order)

	//fired coding next!!!

	return c.JSON(http.StatusOK, map[string]any{"msg": "order deleted"})

}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrders) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchOrders := make([]*MatchedOrders, len(matches))

	isBid := false
	if order.Bid {
		isBid = true
	}

	for i := 0; i < len(matchOrders); i++ {
		id := matches[i].Bid.ID
		if isBid {
			id = matches[i].Ask.ID
		}
		matchOrders[i] = &MatchedOrders{
			Size:  matches[i].SizeFilled,
			Price: matches[i].Price,
			ID:    id,
		}
	}
	return matches, matchOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	user, ok := ex.Users[order.UserID]
	if !ok {
		return fmt.Errorf("user not found: %d", user.ID)
	}

	exchangePubKey := ex.PrivateKey.Public()
	publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	toAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	amount := big.NewInt(int64(order.Size))

	return transferETH(ex.Client, user.PrivateKey, toAddress, amount)

	// transfer => user => exchange

}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]any{"msg": "limit order placed"})
	}

	if placeOrderData.Type == MarketOrder {
		matches, matchOrders := ex.handlePlaceMarketOrder(market, order)

		if err := ex.handelMatches(matches); err != nil {

		}

		return c.JSON(http.StatusOK, map[string]any{"matches": matchOrders})
	}

	return nil
}

func (ex *Exchange) handelMatches(matches []orderbook.Match) error {

	return nil
}
