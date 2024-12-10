package server

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"strconv"
	"sync"

	"github.com/anakinrm/crypto-exchange/orderbook"
	"github.com/anakinrm/crypto-exchange/server/token"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Exchange struct {
	mu    sync.RWMutex
	Users map[int64]*User
	//orders maps a user to his order
	Orders     map[int64][]*orderbook.Order
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[token.Market]*orderbook.Orderbook
}

func NewExchange(privateKey string) (*Exchange, error) {
	orderbooks := make(map[token.Market]*orderbook.Orderbook)
	orderbooks[token.MarketETH] = orderbook.NewOrderbook()
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	return &Exchange{
		Users:      make(map[int64]*User),
		Orders:     make(map[int64][]*orderbook.Order),
		PrivateKey: privateKeyECDSA,
		orderbooks: orderbooks,
	}, nil
}

type GetOrdersResponse struct {
	Asks []Order
	Bids []Order
}

func (ex *Exchange) registerUser(id int64, userName, passWd, email string, phone int64) {
	user := NewUser(id, userName, passWd, email, phone)
	ex.Users[id] = user

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Info("new exchange user")
}

func (ex *Exchange) handleGetTrades(c echo.Context) error {
	market := token.Market(c.Param("market"))

	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, APIError{Error: "orderbook not found"})
	}

	return c.JSON(http.StatusOK, ob.Trades)

}

func (ex *Exchange) handleGetOrders(c echo.Context) error {
	userIDstr := c.Param("userID")
	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		return err
	}

	ex.mu.RLock()
	orderbookOrders := ex.Orders[int64(userID)]
	ordersResp := &GetOrdersResponse{
		Asks: []Order{},
		Bids: []Order{},
	}

	for i := 0; i < len(orderbookOrders); i++ {
		// It could be that the order is getting filled even though its included in this
		//respnse. we must double check if the limit is not nil
		if orderbookOrders[i].Limit == nil {
			continue
		}
		order := Order{
			UserID:    orderbookOrders[i].UserID,
			ID:        orderbookOrders[i].ID,
			Price:     orderbookOrders[i].Limit.Price,
			Size:      orderbookOrders[i].Size,
			Bid:       orderbookOrders[i].Bid,
			Timestamp: orderbookOrders[i].Timestamp,
		}

		if order.Bid {
			ordersResp.Bids = append(ordersResp.Bids, order)
		} else {
			ordersResp.Asks = append(ordersResp.Asks, order)
		}
	}
	ex.mu.RUnlock()

	return c.JSON(http.StatusOK, ordersResp)
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := token.Market(c.Param("market"))
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
	var (
		market = token.Market(c.Param("market"))
		ob     = ex.orderbooks[market]
		order  = Order{}
	)

	if len(ob.Bids()) == 0 {
		return c.JSON(http.StatusOK, order)
	}
	bestLimit := ob.Bids()[0]
	bestOrder := bestLimit.Orders[0]

	order.Price = bestLimit.Price
	order.UserID = bestOrder.UserID

	return c.JSON(http.StatusOK, order)
}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	var (
		market = token.Market(c.Param("market"))
		ob     = ex.orderbooks[market]
		order  = Order{}
	)
	if len(ob.Asks()) == 0 {
		return c.JSON(http.StatusOK, order)
	}
	bestLimit := ob.Asks()[0]
	bestOrder := bestLimit.Orders[0]

	order.Price = bestLimit.Price
	order.UserID = bestOrder.UserID

	return c.JSON(http.StatusOK, order)
}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[token.MarketETH]
	order, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "can't find order ID: " + idStr})
	}
	ob.CancelOrder(order)

	log.Println("order canceled id => ", id)

	return c.JSON(http.StatusOK, map[string]any{"msg": "order deleted"})

}

func (ex *Exchange) handlePlaceMarketOrder(market token.Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrders) {
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
		limitUserID := matches[i].Bid.UserID
		id := matches[i].Bid.ID
		if isBid {
			limitUserID = matches[i].Ask.UserID
			id = matches[i].Ask.ID
		}

		matchOrders[i] = &MatchedOrders{
			UserID: limitUserID,
			Size:   matches[i].SizeFilled,
			Price:  matches[i].Price,
			ID:     id,
		}

		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price * matches[i].SizeFilled
	}

	avgPrice := sumPrice / totalSizeFilled

	logrus.WithFields(logrus.Fields{
		"type":     order.Type(),
		"size":     totalSizeFilled,
		"avgPrice": avgPrice,
	}).Info("filled market order")

	// #TODO: this approch is a shit! try modify a decent one
	newOrderMap := make(map[int64][]*orderbook.Order)
	ex.mu.Lock()
	for userID, orderbookOrders := range ex.Orders {
		// If the order is not filled we place it in the map copy
		// this means that size of the order = 0
		for i := 0; i < len(orderbookOrders); i++ {
			if !orderbookOrders[i].IsFilled() {
				newOrderMap[userID] = append(newOrderMap[userID], orderbookOrders[i])
			}
		}
	}

	ex.Orders = newOrderMap
	ex.mu.Unlock()

	return matches, matchOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market token.Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	// keep track of the user orders
	ex.mu.Lock()
	ex.Orders[order.UserID] = append(ex.Orders[order.UserID], order)
	ex.mu.Unlock()

	//og.Printf("new LIMIT order => type:[%t] | price [%2.f] | size [%.2f]", order.Bid, order.Limit.Price, order.Size)

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

	market := token.Market(placeOrderData.Market)
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

		if err := ex.handleMatches(matches); err != nil {
			return err
		}

	}

	resp := &PlaceOrderResponse{
		OrderID: order.ID,
	}
	return c.JSON(http.StatusOK, resp)

}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.UserID)
		}

		err := fromUser.Wallet[token.MarketETH].SubBalance(match.SizeFilled)
		if err != nil {
			return fmt.Errorf("From user insufficient balance")
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.UserID)
		}
		err = toUser.Wallet[token.MarketETH].AddBalance(match.SizeFilled)

	}
	return nil
}
