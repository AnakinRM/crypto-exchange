package server

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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

const (
	MarketETH Market = "ETH"

	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"

	exchangePrivateKey = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
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
		UserID    int64
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

func StartServer() {
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

	buyerAddress := "0x22d491Bde2303f2f43325b2108D26f1eAbA1e32b"
	buyerBalance, _ := ex.Client.BalanceAt(context.Background(), common.HexToAddress(buyerAddress), nil)
	fmt.Println("Buyer Balance:", buyerBalance)

	sellerAddress := "0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0"
	sellerBalance, _ := ex.Client.BalanceAt(context.Background(), common.HexToAddress(sellerAddress), nil)
	fmt.Println("Buyer Balance:", sellerBalance)

	privKeyStr1 := "6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1"
	user1 := NewUser(privKeyStr1, 1)

	privKeyStr2 := "6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c"
	user2 := NewUser(privKeyStr2, 2)

	ex.Users[user1.ID] = user1
	ex.Users[user2.ID] = user2

	johnPrivKeyStr := "646f1ce2fdad0e6deeeb5c7e8e5543bdde65e86029e2fd9fc169899c440a7913"
	johnUser := NewUser(johnPrivKeyStr, 666)
	ex.Users[johnUser.ID] = johnUser

	johnAddress := "0xE11BA2b4D45Eaed5996Cd0823791E0C93114882d"
	johnBalance, _ := ex.Client.BalanceAt(context.Background(), common.HexToAddress(johnAddress), nil)
	fmt.Println("john Balance:", johnBalance)

	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)
	e.GET("/book/:market/asks", ex.handleGetBook)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bestbid", ex.handleGetBestBid)
	e.GET("/book/:market/bestask", ex.handleGetBestAsk)

	e.Start(":3000")

}

type User struct {
	ID         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(pk string, id int64) *User {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         id,
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
				UserID:    order.UserID,
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
				UserID:    order.UserID,
				ID:        order.ID,
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

type PriceResponse struct {
	Price float64
}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]
	if len(ob.Bids()) == 0 {
		return fmt.Errorf("the bids are empty")
	}
	bestBidPrice := ob.Bids()[0].Price
	pr := PriceResponse{
		Price: bestBidPrice,
	}

	return c.JSON(http.StatusOK, pr)
}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]
	if len(ob.Asks()) == 0 {
		return fmt.Errorf("the asks are empty")
	}
	bestAskPrice := ob.Asks()[0].Price
	pr := PriceResponse{
		Price: bestAskPrice,
	}

	return c.JSON(http.StatusOK, pr)
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

	log.Println("order canceled id => ", id)

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

	totalSizeFilled := 0.0
	sumPrice := 0.0
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

		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price * matches[i].SizeFilled
	}

	avgPrice := sumPrice / totalSizeFilled

	log.Printf("filled market order => %d | size: [%.2f] | Avg Price: [%.2f]", order.ID, totalSizeFilled, avgPrice)

	return matches, matchOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	log.Printf("new LIMIT order => type:[%t] | price [%2.f] | size [%.2f]", order.Bid, order.Limit.Price, order.Size)

	return nil

}

type PlaceOrderResponse struct {
	OrderID int64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	//limit orders
	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}

	}

	// market orders
	if placeOrderData.Type == MarketOrder {
		matches, _ := ex.handlePlaceMarketOrder(market, order)

		if err := ex.handelMatches(matches); err != nil {
			return err
		}

		//return c.JSON(http.StatusOK, map[string]any{"matches": matchOrders})
	}
	resp := &PlaceOrderResponse{
		OrderID: order.ID,
	}
	return c.JSON(http.StatusOK, resp)

}

func (ex *Exchange) handelMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.UserID)
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.UserID)
		}
		toAddress := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		//this is only use for the fees
		// exchangePubKey := ex.PrivateKey.Public()
		// publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		// }

		amount := big.NewInt(int64(match.SizeFilled))

		transferETH(ex.Client, fromUser.PrivateKey, toAddress, amount)

	}
	return nil
}

func transferETH(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {

	ctx := context.Background()

	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	fmt.Println("Tx From:", fromAddress, " To: ", to)

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)
}
